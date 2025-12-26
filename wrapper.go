package gcppubsubwrapper

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
)

type PubSubClient struct {
	client *pubsub.Client
	// Cache topics to avoid re-configuring EnableMessageOrdering on every call
	topics sync.Map
}

func NewPubSubClient(ctx context.Context, projectID string) (*PubSubClient, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}
	return &PubSubClient{
		client: client,
		// sync.Map is ready to use as-is, but explicitly mentioning it
		// helps documentation/readability.
	}, nil
}

// getTopic handles the initialization and caching of topic handles
func (p *PubSubClient) getTopic(topicID string) *pubsub.Topic {
	if t, ok := p.topics.Load(topicID); ok {
		return t.(*pubsub.Topic)
	}

	t := p.client.Topic(topicID)
	// We enable this globally for the topic handle so it can handle
	// both ordered and un-ordered messages seamlessly.
	t.EnableMessageOrdering = true

	actual, _ := p.topics.LoadOrStore(topicID, t)
	return actual.(*pubsub.Topic)
}

func (p *PubSubClient) Publish(ctx context.Context, topicID string, data []byte, attrs map[string]string, orderingKey string) (string, error) {
	topic := p.getTopic(topicID)

	result := topic.Publish(ctx, &pubsub.Message{
		Data:        data,
		Attributes:  attrs,
		OrderingKey: orderingKey,
	})

	id, err := result.Get(ctx)
	if err != nil {
		// CRITICAL FIX: If an ordered message fails, the key is blocked.
		// We must resume it, or that orderingKey will never work again on this instance.
		if orderingKey != "" {
			topic.ResumePublish(orderingKey)
		}
		return "", fmt.Errorf("failed to publish message: %w", err)
	}
	return id, nil
}

func (p *PubSubClient) Subscribe(ctx context.Context, subscriptionID string, handler func(ctx context.Context, msg *pubsub.Message)) error {
	sub := p.client.Subscription(subscriptionID)
	return sub.Receive(ctx, handler)
}

// Close flushes outstanding messages and closes the connection.
func (p *PubSubClient) Close() {
	p.topics.Range(func(key, value any) bool {
		value.(*pubsub.Topic).Stop()
		return true
	})
	p.client.Close()
}
