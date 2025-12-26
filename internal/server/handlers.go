package server

import (
	"encoding/json"
	"net/http"

	"pubsub-proxy/internal/gcp" // Adjust to your module path
)

type Server struct {
	PS *gcp.PubSubManager
}

type PublishRequest struct {
	Topic       string            `json:"topic"`
	Data        string            `json:"data"`
	Attributes  map[string]string `json:"attributes"`
	OrderingKey string            `json:"orderingKey"`
}

func (s *Server) HandlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	id, err := s.PS.Publish(r.Context(), req.Topic, []byte(req.Data), req.Attributes, req.OrderingKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"messageID": id})
}
