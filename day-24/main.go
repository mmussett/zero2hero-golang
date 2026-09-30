// Package main demonstrates error handling at scale: sentinel errors, typed
// AppError, HTTP status mapping, errors.Join for validation, and errors.Is/As.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ---- Sentinel errors -----------------------------------------------------

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
)

// ---- AppError ------------------------------------------------------------

// AppError is a typed error that carries an HTTP status code, a machine-
// readable code string, a human-readable message, and a wrapped sentinel.
type AppError struct {
	Code    int    // HTTP status code
	ErrCode string // machine-readable code, e.g. "NOT_FOUND"
	Message string // human-readable description
	Err     error  // wrapped sentinel
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// StatusCode returns the HTTP status code for the error.
func (e *AppError) StatusCode() int { return e.Code }

// Helper constructors ---------------------------------------------------------

// NotFound returns an AppError that wraps ErrNotFound.
func NotFound(msg string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		ErrCode: "NOT_FOUND",
		Message: msg,
		Err:     ErrNotFound,
	}
}

// InvalidInput returns an AppError that wraps ErrInvalidInput.
func InvalidInput(msg string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		ErrCode: "INVALID_INPUT",
		Message: msg,
		Err:     ErrInvalidInput,
	}
}

// Unauthorized returns an AppError that wraps ErrUnauthorized.
func Unauthorized(msg string) *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		ErrCode: "UNAUTHORIZED",
		Message: msg,
		Err:     ErrUnauthorized,
	}
}

// ---- HTTP helpers -------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// handleErr writes a JSON error response derived from an AppError if possible,
// or a generic 500 otherwise.
func handleErr(w http.ResponseWriter, err error) {
	var ae *AppError
	if errors.As(err, &ae) {
		writeJSON(w, ae.StatusCode(), map[string]string{
			"code":    ae.ErrCode,
			"message": ae.Message,
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"code":    "INTERNAL",
		"message": "internal server error",
	})
}

// ---- Domain types --------------------------------------------------------

type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ---- Validation ----------------------------------------------------------

// validateNote validates a note's title and body, collecting all errors.
func validateNote(title, body string) error {
	var errs []error
	if title == "" {
		errs = append(errs, errors.New("title is required"))
	}
	if len(title) > 200 {
		errs = append(errs, errors.New("title must be 200 characters or fewer"))
	}
	if len(body) > 10_000 {
		errs = append(errs, errors.New("body must be 10 000 characters or fewer"))
	}
	return errors.Join(errs...)
}

// ---- In-memory store (same as day-22, now returns AppErrors) -------------

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

func (s *Store) Get(id int) (Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.notes[id]
	if !ok {
		return Note{}, NotFound(fmt.Sprintf("note %d not found", id))
	}
	return n, nil
}

func (s *Store) Create(title, body string) (Note, error) {
	if err := validateNote(title, body); err != nil {
		return Note{}, InvalidInput(err.Error())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	now := time.Now()
	n := Note{ID: s.seq, Title: title, Body: body, CreatedAt: now, UpdatedAt: now}
	s.notes[s.seq] = n
	return n, nil
}

func (s *Store) Update(id int, title, body string) (Note, error) {
	if err := validateNote(title, body); err != nil {
		return Note{}, InvalidInput(err.Error())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.notes[id]
	if !ok {
		return Note{}, NotFound(fmt.Sprintf("note %d not found", id))
	}
	n.Title = title
	n.Body = body
	n.UpdatedAt = time.Now()
	s.notes[id] = n
	return n, nil
}

func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.notes[id]; !ok {
		return NotFound(fmt.Sprintf("note %d not found", id))
	}
	delete(s.notes, id)
	return nil
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
		return 0, InvalidInput(fmt.Sprintf("invalid %s: %q", key, raw))
	}
	return id, nil
}

func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (h *Handler) listNotes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.List())
}

func (h *Handler) createNote(w http.ResponseWriter, r *http.Request) {
	var req noteRequest
	if err := readJSON(r, &req); err != nil {
		handleErr(w, InvalidInput("request body must be valid JSON"))
		return
	}
	n, err := h.store.Create(req.Title, req.Body)
	if err != nil {
		handleErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

func (h *Handler) getNote(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt(r, "id")
	if err != nil {
		handleErr(w, err)
		return
	}
	n, err := h.store.Get(id)
	if err != nil {
		handleErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) updateNote(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt(r, "id")
	if err != nil {
		handleErr(w, err)
		return
	}
	var req noteRequest
	if err := readJSON(r, &req); err != nil {
		handleErr(w, InvalidInput("request body must be valid JSON"))
		return
	}
	n, err := h.store.Update(id, req.Title, req.Body)
	if err != nil {
		handleErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt(r, "id")
	if err != nil {
		handleErr(w, err)
		return
	}
	if err := h.store.Delete(id); err != nil {
		handleErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// NewRouter wires up the chi router.
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

// ---- Demo: errors.Is / errors.As / errors.Join ---------------------------

func runDemo() {
	fmt.Println("=== Error Handling Demo ===")

	// 1. Sentinel identity check with errors.Is
	err := NotFound("note 42 not found")
	fmt.Printf("errors.Is(err, ErrNotFound)     = %v\n", errors.Is(err, ErrNotFound))
	fmt.Printf("errors.Is(err, ErrInvalidInput) = %v\n", errors.Is(err, ErrInvalidInput))

	// 2. Type assertion with errors.As
	var ae *AppError
	if errors.As(err, &ae) {
		fmt.Printf("AppError: code=%d errCode=%s msg=%s\n", ae.Code, ae.ErrCode, ae.Message)
	}

	// 3. errors.Join to collect multiple validation errors
	joinErr := validateNote("", string(make([]byte, 10_001)))
	fmt.Printf("Validation errors:\n  %v\n", joinErr)

	// Confirm both sentinels are detectable in the joined error.
	fmt.Printf("errors.Is(joinErr, ErrInvalidInput) via Join: %v\n",
		errors.Is(InvalidInput(joinErr.Error()), ErrInvalidInput))

	// 4. Wrapped chain
	wrapped := fmt.Errorf("pipeline failed: %w", NotFound("widget 7 not found"))
	fmt.Printf("Wrapped errors.Is(ErrNotFound) = %v\n", errors.Is(wrapped, ErrNotFound))

	fmt.Println()
	fmt.Println("Starting Notes API with error-aware handlers on :8080")
}

func main() {
	runDemo()

	store := NewStore()
	router := NewRouter(store)
	if err := http.ListenAndServe(":8080", router); err != nil {
		fmt.Printf("server: %v\n", err)
	}
}
