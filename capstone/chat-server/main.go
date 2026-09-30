// Package main implements a simple chat server using Server-Sent Events (SSE)
// for streaming and plain HTTP POST for sending messages.
//
// Routes
//   GET  /rooms                        — list active rooms with member count
//   POST /rooms/{room}/join?name=Alice — join (create) a room
//   POST /rooms/{room}/send?name=Alice — send a message  body: {"message":"..."}
//   GET  /rooms/{room}/events          — SSE stream of room messages
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ---------------------------------------------------------------------------
// Message
// ---------------------------------------------------------------------------

// Message is the unit broadcast to all room subscribers.
type Message struct {
	From      string    `json:"from"`
	Message   string    `json:"message"`
	Room      string    `json:"room"`
	Timestamp time.Time `json:"timestamp"`
}

// ---------------------------------------------------------------------------
// Subscriber / Room
// ---------------------------------------------------------------------------

// subscriber represents a single SSE connection.
type subscriber struct {
	ch chan Message
}

// Room holds the subscribers for one named chat room.
type Room struct {
	Name string
	mu   sync.Mutex
	subs []*subscriber
}

// subscribe registers a new SSE connection and returns its subscriber handle.
func (r *Room) subscribe() *subscriber {
	s := &subscriber{ch: make(chan Message, 64)}
	r.mu.Lock()
	r.subs = append(r.subs, s)
	r.mu.Unlock()
	return s
}

// unsubscribe removes a subscriber and closes its channel.
func (r *Room) unsubscribe(s *subscriber) {
	r.mu.Lock()
	for i, sub := range r.subs {
		if sub == s {
			r.subs = append(r.subs[:i], r.subs[i+1:]...)
			break
		}
	}
	r.mu.Unlock()
	close(s.ch)
}

// broadcast sends a message to every current subscriber (non-blocking; slow
// subscribers are skipped to avoid head-of-line blocking).
func (r *Room) broadcast(msg Message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, sub := range r.subs {
		select {
		case sub.ch <- msg:
		default:
		}
	}
}

// memberCount returns the number of active SSE connections.
func (r *Room) memberCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.subs)
}

// ---------------------------------------------------------------------------
// Hub
// ---------------------------------------------------------------------------

// Hub manages all rooms.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]*Room)}
}

// GetOrCreate returns the named room, creating it if necessary.
func (h *Hub) GetOrCreate(name string) *Room {
	// Fast path: room already exists.
	h.mu.RLock()
	r, ok := h.rooms[name]
	h.mu.RUnlock()
	if ok {
		return r
	}
	// Slow path: create.
	h.mu.Lock()
	defer h.mu.Unlock()
	if r, ok = h.rooms[name]; ok {
		return r
	}
	r = &Room{Name: name}
	h.rooms[name] = r
	return r
}

// Get returns the named room or false if it does not exist.
func (h *Hub) Get(name string) (*Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	r, ok := h.rooms[name]
	return r, ok
}

// roomInfo is used for the list-rooms JSON response.
type roomInfo struct {
	Room    string `json:"room"`
	Members int    `json:"members"`
}

// List returns a snapshot of all rooms with their member counts.
func (h *Hub) List() []roomInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]roomInfo, 0, len(h.rooms))
	for name, r := range h.rooms {
		out = append(out, roomInfo{Room: name, Members: r.memberCount()})
	}
	return out
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---------------------------------------------------------------------------
// HTTP handlers
// ---------------------------------------------------------------------------

// GET /rooms
func handleListRooms(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, hub.List())
	}
}

// POST /rooms/{room}/join?name=Alice
func handleJoin(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomName := r.PathValue("room")
		name := r.URL.Query().Get("name")
		if name == "" {
			writeError(w, http.StatusBadRequest, "query param 'name' is required")
			return
		}
		hub.GetOrCreate(roomName)
		writeJSON(w, http.StatusOK, map[string]string{
			"room":   roomName,
			"joined": name,
		})
	}
}

// POST /rooms/{room}/send?name=Alice   body: {"message":"hello"}
func handleSend(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomName := r.PathValue("room")
		name := r.URL.Query().Get("name")
		if name == "" {
			writeError(w, http.StatusBadRequest, "query param 'name' is required")
			return
		}
		var body struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Message == "" {
			writeError(w, http.StatusBadRequest, "body must contain a non-empty 'message' field")
			return
		}
		room, ok := hub.Get(roomName)
		if !ok {
			// Auto-create so send works even without an explicit join.
			room = hub.GetOrCreate(roomName)
		}
		msg := Message{
			From:      name,
			Message:   body.Message,
			Room:      roomName,
			Timestamp: time.Now().UTC(),
		}
		room.broadcast(msg)
		writeJSON(w, http.StatusOK, msg)
	}
}

// GET /rooms/{room}/events — SSE stream
func handleEvents(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomName := r.PathValue("room")
		room := hub.GetOrCreate(roomName)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		// Disable proxy buffering (nginx et al.)
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)

		// Send an initial comment so the client knows the stream is live.
		fmt.Fprintf(w, ": connected to room %s\n\n", roomName)
		flusher.Flush()

		sub := room.subscribe()
		defer room.unsubscribe(sub)

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				// Client disconnected.
				return
			case msg, open := <-sub.ch:
				if !open {
					return
				}
				data, err := json.Marshal(msg)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	hub := NewHub()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /rooms", handleListRooms(hub))
	mux.HandleFunc("POST /rooms/{room}/join", handleJoin(hub))
	mux.HandleFunc("POST /rooms/{room}/send", handleSend(hub))
	mux.HandleFunc("GET /rooms/{room}/events", handleEvents(hub))

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  0,             // SSE streams are long-lived; no read timeout
		WriteTimeout: 0,             // same for writes
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("chat-server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down…")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown: %v", err)
	}
	log.Println("stopped")
}
