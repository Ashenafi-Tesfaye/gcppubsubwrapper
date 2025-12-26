package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Ashenafi-Tesfaye/dependency-wrapper/internal/gcp"
	"github.com/Ashenafi-Tesfaye/dependency-wrapper/internal/server"
)

func main() {
	projectID := os.Getenv("GCP_PROJECT_ID")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()

	// 1. Initialize Dependencies
	psManager, err := gcp.NewPubSubManager(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to init PubSub: %v", err)
	}
	defer psManager.Close()

	// 2. Initialize Server with Dependencies
	srv := &server.Server{PS: psManager}

	// 3. Define Routes
	mux := http.NewServeMux()
	server.RegisterRoutes(mux, srv)

	// 4. Start
	log.Printf("Starting modular proxy on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
