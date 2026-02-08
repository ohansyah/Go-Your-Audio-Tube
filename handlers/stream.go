package handlers

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// StreamHandler handles the audio streaming request
func StreamHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "Missing URL parameter", http.StatusBadRequest)
		return
	}

	// Basic validation to prevent arbitrary command injection somewhat,
	// though exec.Command prevents shell injection by default.
	// Ideally user inputs should be strictly validated.
	if !strings.Contains(url, "youtube.com") && !strings.Contains(url, "youtu.be") {
		http.Error(w, "Invalid URL (must be YouTube)", http.StatusBadRequest)
		return
	}

	log.Printf("Processing stream request for: %s", url)

	// Context for timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// 1. Get the direct stream URL using yt-dlp
	// Check if yt-dlp is in PATH, otherwise try local
	executable := "yt-dlp"
	if _, err := exec.LookPath(executable); err != nil {
		if _, err := os.Stat("./yt-dlp"); err == nil {
			executable = "./yt-dlp"
		}
	}

	// -g: get URL
	// -f bestaudio: best audio format
	cmd := exec.CommandContext(ctx, executable, "-g", "-f", "bestaudio[ext=m4a]/bestaudio", url)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("yt-dlp error: %v", err)
		http.Error(w, "Failed to extract audio stream", http.StatusInternalServerError)
		return
	}

	realURL := strings.TrimSpace(string(out))
	if realURL == "" {
		log.Println("yt-dlp returned empty URL")
		http.Error(w, "No audio stream found", http.StatusNotFound)
		return
	}

	// 2. Fetch the actual stream
	// Create a new request based on the real URL
	req, err := http.NewRequestWithContext(r.Context(), "GET", realURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Use default client to fetch the stream
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to fetch stream: %v", err)
		http.Error(w, "Failed to connect to stream source", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 3. Stream data to client
	// Copy headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	// Explicitly set content type if missing or ensure it's audio
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "audio/mp4")
	}

	// Stream the body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		// Client might have disconnected
		log.Printf("Stream interrupted: %v", err)
	}
}
