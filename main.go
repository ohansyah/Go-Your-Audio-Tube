package main

import (
	"log"
	"net/http"
	"time"

	"go-background-music/handlers"
)

func main() {
	// Register handlers
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/stream", handlers.StreamHandler)

	// Configure server
	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Minute, // Long timeout for streaming
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Server starting on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
