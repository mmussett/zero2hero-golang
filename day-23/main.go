package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Note represents a single note in the store.
type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store is a thread-safe in-memory notes store.
type Store struct {
	mu    sync.RWMutex
	notes map[int]Note
	seq   int
}

// NewStore creates an empty Store.
func NewStore() *Store {
	return &Store{notes: make(map[int]Note)}
}

// List returns all notes ordered by ID (ascending).
func (s *Store) List() []Note {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Note, 0, len(s.notes))
	for _, n := range s.notes {
		result = append(result, n)
	}
	return result
}

// Get returns a note by ID, or false if not found.
func (s *Store) Get(id int) (Note, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.notes[id]
	return n, ok
}

// Create adds a new note and returns it.
func (s *Store) Create(title, body string) Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	now := time.Now()
	n := Note{
		ID:        s.seq,
		Title:     title,
		Body:      body,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.notes[s.seq] = n
	return n
}

// Update replaces a note's title and body, returning the updated note.
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

// Delete removes a note by ID, returning true if it existed.
func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.notes[id]
	if ok {
		delete(s.notes, id)
	}
	return ok
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// readJSON decodes the request body into v. Unknown fields cause an error.
func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// urlParamInt extracts a URL parameter as an int, returning an error string on failure.
func urlParamInt(r *http.Request, key string) (int, error) {
	raw := chi.URLParam(r, key)
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %q", key, raw)
	}
	return id, nil
}

// noteRequest is the JSON body for create/update requests.
type noteRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Handler holds the store and exposes HTTP handler methods.
type Handler struct {
	store *Store
}

func (h *Handler) listNotes(w http.ResponseWriter, r *http.Request) {
	notes := h.store.List()
	writeJSON(w, http.StatusOK, notes)
}

func (h *Handler) createNote(w http.ResponseWriter, r *http.Request) {
	var req noteRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}
	n := h.store.Create(req.Title, req.Body)
	writeJSON(w, http.StatusCreated, n)
}

func (h *Handler) getNote(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	n, ok := h.store.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) updateNote(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	var req noteRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}
	n, ok := h.store.Update(id, req.Title, req.Body)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if !h.store.Delete(id) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// NewRouter builds a chi router wired up with the given store.
func NewRouter(store *Store) http.Handler {
	h := &Handler{store: store}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
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

func main() {
	store := NewStore()
	router := NewRouter(store)

	addr := ":8080"
	fmt.Printf("Notes API listening on %s\n", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		fmt.Printf("server error: %v\n", err)
	}
}
