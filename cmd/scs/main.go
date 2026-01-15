package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"github.com/sokinpui/sync-clip"

	"github.com/atotto/clipboard"
)

func main() {
	configPath := flag.String("c", "", "Path to configuration file")
	flag.Parse()

	cfg, err := syncclip.LoadConfig(*configPath, "scs.conf")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting Sync-clip server on %s...", cfg.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleClipboard)

	server := &http.Server{
		Addr:    cfg.Port,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}

func handleClipboard(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	if r.Method == http.MethodPost {
		content, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		if err := clipboard.WriteAll(string(content)); err != nil {
			log.Printf("Failed to write to clipboard: %v", err)
			http.Error(w, "Failed to update clipboard", http.StatusInternalServerError)
			return
		}

		log.Printf("Clipboard updated: %d bytes", len(content))
		w.WriteHeader(http.StatusOK)
		return
	}

	content, err := clipboard.ReadAll()
	if err != nil {
		log.Printf("Failed to read from clipboard: %v", err)
		http.Error(w, "Failed to read clipboard", http.StatusInternalServerError)
		return
	}

	log.Printf("Clipboard sent: %d bytes", len(content))
	fmt.Fprint(w, content)
}
