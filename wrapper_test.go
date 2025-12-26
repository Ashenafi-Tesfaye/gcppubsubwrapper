package gcppubsubwrapper

import (
	"context"
	"testing"

	"github.com/Ashenafi-Tesfaye/gcppubsubwrapper/internal/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestPublishWithOrderingKey(t *testing.T) {
	ctx := context.Background()

	srv := pstest.NewServer()
	defer srv.Close()

	conn, err := grpc.DialContext(ctx, srv.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	pubsubClient, err := pubsub.NewClient(ctx, "project-id", option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	wrapper := &PubSubClient{client: pubsubClient}
	defer wrapper.Close()

	topicID := "test-topic"
	_, err = pubsubClient.CreateTopic(ctx, topicID)
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	data := []byte("test message")
	orderingKey := "order-123"

	id, err := wrapper.Publish(ctx, topicID, data, nil, orderingKey)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id == "" {
		t.Error("expected a non-empty message ID")
	}

	msgs := srv.Messages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message in emulator, got %d", len(msgs))
	}

	if msgs[0].OrderingKey != orderingKey {
		t.Errorf("expected ordering key %s, got %s", orderingKey, msgs[0].OrderingKey)
	}
}
