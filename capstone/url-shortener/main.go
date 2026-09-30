package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ---------------------------------------------------------------------------
// Data model
// ---------------------------------------------------------------------------

const (
	codeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLen      = 6
)

// Entry holds a single shortened URL record.
// Clicks is modified atomically; all other fields are write-once after creation.
type Entry struct {
	URL       string    `json:"url"`
	Code      string    `json:"short"`
	Clicks    int64     `json:"clicks"` // accessed via sync/atomic
	CreatedAt time.Time `json:"created_at"`
}

// entryView is a snapshot used for JSON serialisation of the list endpoint,
// so that atomic.LoadInt64 is used for Clicks.
type entryView struct {
	URL       string    `json:"url"`
	Code      string    `json:"short"`
	Clicks    int64     `json:"clicks"`
	CreatedAt time.Time `json:"created_at"`
}

// ---------------------------------------------------------------------------
// In-memory store
// ---------------------------------------------------------------------------

type Store struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

func NewStore() *Store {
	return &Store{entries: make(map[string]*Entry)}
}

// generateCode returns a 6-character random alphanumeric string.
func generateCode() (string, error) {
	raw := make([]byte, codeLen)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, b := range raw {
		sb.WriteByte(codeAlphabet[int(b)%len(codeAlphabet)])
	}
	return sb.String(), nil
}

// Add stores a new URL and returns its Entry.
func (s *Store) Add(url string) (*Entry, error) {
	code, err := generateCode()
	if err != nil {
		return nil, err
	}
	entry := &Entry{
		URL:       url,
		Code:      code,
		CreatedAt: time.Now().UTC(),
	}
	s.mu.Lock()
	s.entries[code] = entry
	s.mu.Unlock()
	return entry, nil
}

// Redirect looks up a code and atomically increments its click counter.
// Returns the target URL and whether the code was found.
func (s *Store) Redirect(code string) (url string, ok bool) {
	s.mu.RLock()
	e, ok := s.entries[code]
	s.mu.RUnlock()
	if !ok {
		return "", false
	}
	atomic.AddInt64(&e.Clicks, 1)
	return e.URL, true
}

// Delete removes a code. Returns false if the code did not exist.
func (s *Store) Delete(code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.entries[code]; !ok {
		return false
	}
	delete(s.entries, code)
	return true
}

// List returns a snapshot of all entries.
func (s *Store) List() []entryView {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]entryView, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, entryView{
			URL:       e.URL,
			Code:      e.Code,
			Clicks:    atomic.LoadInt64(&e.Clicks),
			CreatedAt: e.CreatedAt,
		})
	}
	return out
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

// ---------------------------------------------------------------------------
// HTTP handlers
// ---------------------------------------------------------------------------

// POST /shorten
// Body: {"url":"https://..."}
// Response: {"short":"abc123","url":"https://..."}
func handleShorten(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
			http.Error(w, `{"error":"url is required"}`, http.StatusBadRequest)
			return
		}
		entry, err := store.Add(req.URL)
		if err != nil {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"short": entry.Code,
			"url":   entry.URL,
		})
	}
}

// GET /api/links
func handleList(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.List())
	}
}

// DELETE /api/links/{code}
func handleDelete(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		if !store.Delete(code) {
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// GET /{code}
func handleRedirect(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		if code == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		target, ok := store.Redirect(code)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, target, http.StatusFound)
	}
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	store := NewStore()

	mux := http.NewServeMux()
	// Register routes using Go 1.22+ method+path pattern syntax.
	mux.HandleFunc("POST /shorten", handleShorten(store))
	mux.HandleFunc("GET /api/links", handleList(store))
	mux.HandleFunc("DELETE /api/links/{code}", handleDelete(store))
	mux.HandleFunc("GET /{code}", handleRedirect(store))

	srv := &http.Server{
		Addr:         *addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("url-shortener listening on %s", *addr)
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
