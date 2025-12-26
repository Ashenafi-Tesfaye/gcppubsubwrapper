package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"cloud.google.com/go/pubsub"
)

// PublishRequest matches the JSON structure expected from Orbitera
type PublishRequest struct {
	Topic       string            `json:"topic"`
	Data        string            `json:"data"`
	Attributes  map[string]string `json:"attributes"`
	OrderingKey string            `json:"orderingKey"`
}

type PubSubService struct {
	client *pubsub.Client
	topics sync.Map // Cache topic handles to support ordering and efficiency
}

func (s *PubSubService) getTopic(topicID string) *pubsub.Topic {
	if t, ok := s.topics.Load(topicID); ok {
		return t.(*pubsub.Topic)
	}

	t := s.client.Topic(topicID)
	// IMPORTANT: Enabling this allows the handle to respect OrderingKeys
	t.EnableMessageOrdering = true

	actual, _ := s.topics.LoadOrStore(topicID, t)
	return actual.(*pubsub.Topic)
}

func (s *PubSubService) publishHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	topic := s.getTopic(req.Topic)

	result := topic.Publish(ctx, &pubsub.Message{
		Data:        []byte(req.Data),
		Attributes:  req.Attributes,
		OrderingKey: req.OrderingKey,
	})

	id, err := result.Get(ctx)
	if err != nil {
		// CRITICAL: If an ordered publish fails, the key is paused.
		// We must resume it to allow subsequent messages for this key.
		if req.OrderingKey != "" {
			topic.ResumePublish(req.OrderingKey)
		}
		log.Printf("Publish error for key %s on topic %s: %v", req.OrderingKey, req.Topic, err)
		http.Error(w, fmt.Sprintf("Publish error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"messageID": id})
}

func main() {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCP_PROJECT_ID environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create PubSub client: %v", err)
	}
	defer client.Close()

	service := &PubSubService{
		client: client,
	}

	http.HandleFunc("/publish", service.publishHandler)

	log.Printf("Proxy Service started on :%s for project %s", port, projectID)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
