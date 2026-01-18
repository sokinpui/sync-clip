package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"net"
	"log"
	"net/http"
	"strings"
	"github.com/sokinpui/sync-clip"
	"golang.design/x/clipboard"
)

var hub *syncclip.Hub

func main() {
	configPath := flag.String("c", "", "Path to configuration file")
	flag.Parse()

	cfg, err := syncclip.LoadConfig(*configPath, "scs.conf")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := clipboard.Init(); err != nil {
		log.Fatalf("Failed to initialize clipboard: %v", err)
	}

	hub = syncclip.NewHub()
	go hub.Run()
	hub.StartWatcher(context.Background())

	for _, peerURL := range cfg.Peers {
		log.Printf("Connecting to peer: %s", peerURL)
		go hub.ConnectToPeer(peerURL)
	}

	log.Printf("Starting Sync-clip server on %s...", cfg.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleClipboard)
	mux.HandleFunc("/ws", hub.HandleWebSocket)

	server := &http.Server{
		Addr:    cfg.Port,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}

func handleClipboard(w http.ResponseWriter, r *http.Request) {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, host)

	if r.Method == http.MethodPost {
		content, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		isImage := false
		format := clipboard.FmtText
		if strings.HasPrefix(r.Header.Get("Content-Type"), "image/") {
			format = clipboard.FmtImage
			isImage = true
		}

		if hub.IsNewContent(content, isImage) {
			_ = clipboard.Write(format, content)
			hub.BroadcastLocal(content, isImage)
			log.Printf("[HTTP] Clipboard updated via POST: %d bytes", len(content))
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	img := clipboard.Read(clipboard.FmtImage)
	if len(img) > 0 {
		log.Printf("[HTTP] Sending image: %d bytes", len(img))
		w.Header().Set("Content-Type", "image/png")
		_, _ = io.Copy(w, bytes.NewReader(img))
		return
	}

	txt := clipboard.Read(clipboard.FmtText)
	if len(txt) > 0 {
		log.Printf("[HTTP] Sending text: %d bytes", len(txt))
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.Copy(w, bytes.NewReader(txt))
		return
	}

	log.Printf("[HTTP] Clipboard is empty")
	w.WriteHeader(http.StatusNoContent)
}
