package gcppubsubwrapper

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

// PubSubClient wraps the Google Pub/Sub client
// and exposes simplified methods for publishing and subscribing.
type PubSubClient struct {
	client *pubsub.Client
}

// NewPubSubClient creates a new PubSubClient for the given projectID.
func NewPubSubClient(ctx context.Context, projectID string) (*PubSubClient, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}
	return &PubSubClient{client: client}, nil
}

// Publish sends a message to the specified topic, with optional ordering key support.
func (p *PubSubClient) Publish(ctx context.Context, topicID string, data []byte, attrs map[string]string, orderingKey string) (string, error) {
	topic := p.client.Topic(topicID)
	// Enable message ordering on the topic if orderingKey is provided
	if orderingKey != "" {
		topic.EnableMessageOrdering = true
	}
	result := topic.Publish(ctx, &pubsub.Message{
		Data:        data,
		Attributes:  attrs,
		OrderingKey: orderingKey,
	})
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to publish message: %w", err)
	}
	return id, nil
}

// Subscribe receives messages from the specified subscription.
// The handler function is called for each message.
func (p *PubSubClient) Subscribe(ctx context.Context, subscriptionID string, handler func(ctx context.Context, msg *pubsub.Message)) error {
	sub := p.client.Subscription(subscriptionID)
	return sub.Receive(ctx, handler)
}
