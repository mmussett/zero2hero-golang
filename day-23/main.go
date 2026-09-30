package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

// Note represents a single note row from the database.
type Note struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DB wraps a *sql.DB and provides note CRUD operations.
type DB struct {
	db *sql.DB
}

// NewDB opens (or creates) a SQLite database, runs migrations, and returns a DB.
func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// SQLite is single-writer; one connection avoids locking issues.
	db.SetMaxOpenConns(1)

	store := &DB{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return store, nil
}

// migrate creates the notes table if it does not already exist.
func (d *DB) migrate() error {
	_, err := d.db.Exec(`CREATE TABLE IF NOT EXISTS notes (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		title      TEXT    NOT NULL,
		body       TEXT    NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

// Close closes the underlying database connection.
func (d *DB) Close() error { return d.db.Close() }

// List returns all notes ordered by creation time descending.
func (d *DB) List() ([]Note, error) {
	rows, err := d.db.Query(
		`SELECT id, title, body, created_at, updated_at FROM notes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// Get returns the note with the given ID, or sql.ErrNoRows if not found.
func (d *DB) Get(id int64) (Note, error) {
	var n Note
	err := d.db.QueryRow(
		`SELECT id, title, body, created_at, updated_at FROM notes WHERE id = ?`, id).
		Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

// Create inserts a new note inside a transaction and returns the created row.
func (d *DB) Create(title, body string) (Note, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return Note{}, err
	}
	defer tx.Rollback() //nolint:errcheck

	now := time.Now().UTC()
	result, err := tx.Exec(
		`INSERT INTO notes (title, body, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		title, body, now, now)
	if err != nil {
		return Note{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Note{}, err
	}
	if err := tx.Commit(); err != nil {
		return Note{}, err
	}
	return Note{ID: id, Title: title, Body: body, CreatedAt: now, UpdatedAt: now}, nil
}

// Update replaces a note's title and body inside a transaction.
func (d *DB) Update(id int64, title, body string) (Note, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return Note{}, err
	}
	defer tx.Rollback() //nolint:errcheck

	now := time.Now().UTC()
	result, err := tx.Exec(
		`UPDATE notes SET title = ?, body = ?, updated_at = ? WHERE id = ?`,
		title, body, now, id)
	if err != nil {
		return Note{}, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return Note{}, err
	}
	if n == 0 {
		return Note{}, sql.ErrNoRows
	}
	if err := tx.Commit(); err != nil {
		return Note{}, err
	}
	return Note{ID: id, Title: title, Body: body, UpdatedAt: now}, nil
}

// Delete removes a note by ID inside a transaction.
func (d *DB) Delete(id int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	result, err := tx.Exec(`DELETE FROM notes WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

// ---- HTTP helpers -------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func urlParamInt64(r *http.Request, key string) (int64, error) {
	raw := chi.URLParam(r, key)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %q", key, raw)
	}
	return id, nil
}

// ---- HTTP handlers -------------------------------------------------------

type noteRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Handler struct{ db *DB }

func (h *Handler) listNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := h.db.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if notes == nil {
		notes = []Note{}
	}
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
	n, err := h.db.Create(req.Title, req.Body)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

func (h *Handler) getNote(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt64(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	n, err := h.db.Get(id)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) updateNote(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt64(r, "id")
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
	n, err := h.db.Update(id, req.Title, req.Body)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request) {
	id, err := urlParamInt64(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	err = h.db.Delete(id)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// NewRouter builds a chi router backed by the given DB.
func NewRouter(db *DB) http.Handler {
	h := &Handler{db: db}

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
	db, err := NewDB("notes.db")
	if err != nil {
		fmt.Printf("failed to open database: %v\n", err)
		return
	}
	defer db.Close()

	router := NewRouter(db)

	addr := ":8080"
	fmt.Printf("Notes API (SQLite) listening on %s\n", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		fmt.Printf("server error: %v\n", err)
	}
}
