package syncclip

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
	"fmt"

	"github.com/gorilla/websocket"
	"golang.design/x/clipboard"
)

type Message struct {
	Origin  string `json:"origin"`
	IsImage bool   `json:"is_image"`
	Content []byte `json:"content"`
}

type Hub struct {
	id         string
	lastContent []byte
	lastWasImg  bool
	peers      map[*peer]bool
	register   chan *peer
	unregister chan *peer
	broadcast  chan broadcastEvent
	mu         sync.RWMutex
	contentMu  sync.Mutex
}

type broadcastEvent struct {
	message Message
	source  *peer
}

type peer struct {
	hub  *Hub
	conn *websocket.Conn
	send chan Message
	done chan struct{}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewHub() *Hub {
	return &Hub{
		id:         time.Now().String(), // Simple unique ID for the session
		peers:      make(map[*peer]bool),
		register:   make(chan *peer),
		unregister: make(chan *peer),
		broadcast:  make(chan broadcastEvent),
	}
}

func (h *Hub) IsNewContent(content []byte, isImage bool) bool {
	h.contentMu.Lock()
	defer h.contentMu.Unlock()

	if h.lastWasImg == isImage && bytes.Equal(h.lastContent, content) {
		return false
	}

	h.lastContent = content
	h.lastWasImg = isImage
	return true
}

func (h *Hub) StartWatcher(ctx context.Context) {
	go h.watchFormat(ctx, clipboard.FmtText, false)
	go h.watchFormat(ctx, clipboard.FmtImage, true)
}

func (h *Hub) watchFormat(ctx context.Context, format clipboard.Format, isImage bool) {
	ch := clipboard.Watch(ctx, format)
	for {
		select {
		case <-ctx.Done():
			return
		case content := <-ch:
			if !h.IsNewContent(content, isImage) {
				continue
			}

			label := "text"
			if isImage {
				label = "image"
			}

			fmt.Printf("Local clipboard change detected (%s)\n", label)
			h.BroadcastLocal(content, isImage)
		}
	}
}

func (h *Hub) BroadcastLocal(content []byte, isImage bool) {
	h.broadcast <- broadcastEvent{
		message: Message{
			Origin:  h.id,
			IsImage: isImage,
			Content: content,
		},
		source: nil,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case p := <-h.register:
			h.mu.Lock()
			h.peers[p] = true
			log.Printf("Peer connected: %s", p.conn.RemoteAddr())
			h.mu.Unlock()
		case p := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.peers[p]; ok {
				delete(h.peers, p)
				close(p.send)
				log.Printf("Peer disconnected: %s", p.conn.RemoteAddr())
			}
			h.mu.Unlock()
		case event := <-h.broadcast:
			h.mu.Lock()
			for p := range h.peers {
				if p == event.source {
					continue
				}
				select {
				case p.send <- event.message:
				default:
					close(p.send)
					delete(h.peers, p)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.serveConn(conn, nil)
}

func (h *Hub) ConnectToPeer(url string) {
	for {
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Successfully connected to peer: %s", url)

		done := make(chan struct{})
		h.serveConn(conn, done)
		<-done
	}
}

func (h *Hub) serveConn(conn *websocket.Conn, done chan struct{}) {
	p := &peer{hub: h, conn: conn, send: make(chan Message, 256), done: done}
	h.register <- p

	go p.writePump()
	go p.readPump()
}

func (p *peer) readPump() {
	defer func() {
		if p.done != nil {
			close(p.done)
		}
		p.hub.unregister <- p
		p.conn.Close()
	}()

	for {
		_, message, err := p.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		if msg.Origin == p.hub.id {
			continue
		}

		if !p.hub.IsNewContent(msg.Content, msg.IsImage) {
			continue
		}

		format := clipboard.FmtText
		if msg.IsImage {
			format = clipboard.FmtImage
		}

		clipboard.Write(format, msg.Content)
		p.hub.broadcast <- broadcastEvent{message: msg, source: p}
	}
}

func (p *peer) writePump() {
	defer p.conn.Close()
	for msg := range p.send {
		data, _ := json.Marshal(msg)
		if err := p.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}
