package gcp

import (
	"context"
	"sync"

	"cloud.google.com/go/pubsub"
)

type PubSubManager struct {
	client *pubsub.Client
	topics sync.Map
}

func NewPubSubManager(ctx context.Context, projectID string) (*PubSubManager, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &PubSubManager{client: client}, nil
}

func (m *PubSubManager) Publish(ctx context.Context, topicID string, data []byte, attrs map[string]string, key string) (string, error) {
	topic := m.getTopic(topicID)

	result := topic.Publish(ctx, &pubsub.Message{
		Data:        data,
		Attributes:  attrs,
		OrderingKey: key,
	})

	id, err := result.Get(ctx)
	if err != nil {
		if key != "" {
			topic.ResumePublish(key)
		}
		return "", err
	}
	return id, nil
}

func (m *PubSubManager) getTopic(topicID string) *pubsub.Topic {
	if t, ok := m.topics.Load(topicID); ok {
		return t.(*pubsub.Topic)
	}
	t := m.client.Topic(topicID)
	t.EnableMessageOrdering = true
	actual, _ := m.topics.LoadOrStore(topicID, t)
	return actual.(*pubsub.Topic)
}

func (m *PubSubManager) Close() error {
	return m.client.Close()
}
