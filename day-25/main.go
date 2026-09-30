// Package main demonstrates structured logging with log/slog, request-scoped
// loggers stored in context, and a JSON-logging middleware wrapping the
// day-22 Notes API.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ---- Context key for the per-request logger ------------------------------

type loggerKeyType struct{}

var loggerKey loggerKeyType

// withLogger stores l in ctx.
func withLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// loggerFrom retrieves the per-request logger from ctx, falling back to the
// default global logger if none has been stored.
func loggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// ---- statusWriter wraps ResponseWriter to capture the status code --------

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	if sw.status == 0 {
		sw.status = http.StatusOK
	}
	return sw.ResponseWriter.Write(b)
}

// ---- Request ID generation -----------------------------------------------

// newRequestID returns a random 8-byte hex string suitable as a request ID.
func newRequestID() string {
	b := make([]byte, 8)
	rand.Read(b) //nolint:errcheck
	return fmt.Sprintf("%x", b)
}

// ---- Logging middleware --------------------------------------------------

// requestLogger is a chi-compatible middleware that:
//   - generates a unique request ID
//   - injects a per-request slog.Logger into the context
//   - logs method, path, status, duration, and request ID
//   - sets the X-Request-ID response header
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := newRequestID()

		// Build a logger pre-loaded with immutable request fields.
		log := slog.With(
			"request_id", reqID,
			"method", r.Method,
			"path", r.URL.Path,
		)

		// Make it available to downstream handlers via context.
		ctx := withLogger(r.Context(), log)

		// Propagate the request ID to the client.
		w.Header().Set("X-Request-ID", reqID)

		log.Info("request started")

		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r.WithContext(ctx))

		status := sw.status
		if status == 0 {
			status = http.StatusOK
		}

		log.Info("request completed",
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// ---- Notes domain --------------------------------------------------------

type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store struct {
	mu    sync.RWMutex
	notes map[int]Note
	seq   int
}

func NewStore() *Store { return &Store{notes: make(map[int]Note)} }

func (s *Store) List() []Note {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Note, 0, len(s.notes))
	for _, n := range s.notes {
		out = append(out, n)
	}
	return out
}

func (s *Store) Get(id int) (Note, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.notes[id]
	return n, ok
}

func (s *Store) Create(title, body string) Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	now := time.Now()
	n := Note{ID: s.seq, Title: title, Body: body, CreatedAt: now, UpdatedAt: now}
	s.notes[s.seq] = n
	return n
}

func (s *Store) Update(id int, title, body string) (Note, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.notes[id]
	if !ok {
		return Note{}, false
	}
	n.Title = title
	n.Body = body
	n.UpdatedAt = time.Now()
	s.notes[id] = n
	return n, true
}

func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.notes[id]
	if ok {
		delete(s.notes, id)
	}
	return ok
}

// ---- HTTP helpers --------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// ---- Handlers ------------------------------------------------------------

type Handler struct{ store *Store }

type noteRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func urlParamInt(r *http.Request, key string) (int, error) {
	raw := chi.URLParam(r, key)
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %q", key, raw)
	}
	return id, nil
}

func (h *Handler) listNotes(w http.ResponseWriter, r *http.Request) {
	log := loggerFrom(r.Context())
	notes := h.store.List()
	log.Info("listing notes", "count", len(notes))
	writeJSON(w, http.StatusOK, notes)
}

func (h *Handler) createNote(w http.ResponseWriter, r *http.Request) {
	log := loggerFrom(r.Context())
	var req noteRequest
	if err := readJSON(r, &req); err != nil {
		log.Error("decode body failed", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.Title == "" {
		log.Warn("create note: missing title")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}
	n := h.store.Create(req.Title, req.Body)
	log.Info("note created", "note_id", n.ID)
	writeJSON(w, http.StatusCreated, n)
}

func (h *Handler) getNote(w http.ResponseWriter, r *http.Request) {
	log := loggerFrom(r.Context())
	id, err := urlParamInt(r, "id")
	if err != nil {
		log.Warn("bad id param", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	n, ok := h.store.Get(id)
	if !ok {
		log.Warn("note not found", "note_id", id)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
		return
	}
	log.Info("note fetched", "note_id", id)
	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) updateNote(w http.ResponseWriter, r *http.Request) {
	log := loggerFrom(r.Context())
	id, err := urlParamInt(r, "id")
	if err != nil {
		log.Warn("bad id param", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	var req noteRequest
	if err := readJSON(r, &req); err != nil {
		log.Error("decode body failed", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.Title == "" {
		log.Warn("update note: missing title", "note_id", id)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}
	n, ok := h.store.Update(id, req.Title, req.Body)
	if !ok {
		log.Warn("note not found for update", "note_id", id)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
		return
	}
	log.Info("note updated", "note_id", n.ID)
	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request) {
	log := loggerFrom(r.Context())
	id, err := urlParamInt(r, "id")
	if err != nil {
		log.Warn("bad id param", "err", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if !h.store.Delete(id) {
		log.Warn("note not found for delete", "note_id", id)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
		return
	}
	log.Info("note deleted", "note_id", id)
	w.WriteHeader(http.StatusNoContent)
}

// NewRouter builds a chi router with the logging middleware applied.
func NewRouter(store *Store) http.Handler {
	h := &Handler{store: store}

	r := chi.NewRouter()
	r.Use(requestLogger) // our structured-logging middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Route("/api/notes", func(r chi.Router) {
		r.Get("/", h.listNotes)
		r.Post("/", h.createNote)
		r.Get("/{id}", h.getNote)
		r.Put("/{id}", h.updateNote)
		r.Delete("/{id}", h.deleteNote)
	})
	return r
}

// ---- Logger initialisation -----------------------------------------------

// logLevel parses the LOG_LEVEL environment variable, defaulting to Info.
func logLevel() slog.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG", "debug":
		return slog.LevelDebug
	case "WARN", "warn":
		return slog.LevelWarn
	case "ERROR", "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func main() {
	// Configure a global JSON logger.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel(),
	}))
	slog.SetDefault(logger)

	store := NewStore()
	router := NewRouter(store)

	addr := ":8080"
	slog.Info("notes API starting", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		slog.Error("server stopped", "err", err)
	}
}
