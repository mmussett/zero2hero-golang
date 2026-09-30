package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

// Todo represents a single todo item.
type Todo struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// App holds the database connection and HTTP server.
type App struct {
	db *sql.DB
}

func initDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			title       TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			done        INTEGER NOT NULL DEFAULT 0,
			created_at  TEXT NOT NULL,
			updated_at  TEXT NOT NULL
		)
	`)
	return err
}

// writeJSON sends a JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError sends a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// idFromPath extracts the integer ID from a URL path like /todos/42 or /todos/42/done.
func idFromPath(path string) (int64, bool) {
	// path: /todos/{id} or /todos/{id}/done
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// parts[0] = "todos", parts[1] = id, possibly parts[2] = "done"
	if len(parts) < 2 {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (a *App) listTodos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	query := `SELECT id, title, description, done, created_at, updated_at FROM todos WHERE 1=1`
	var args []any

	if doneStr := q.Get("done"); doneStr != "" {
		done := doneStr == "true" || doneStr == "1"
		query += ` AND done = ?`
		if done {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}

	query += ` ORDER BY id ASC`

	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			query += ` LIMIT ?`
			args = append(args, limit)
		}
	}

	rows, err := a.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		log.Println("listTodos:", err)
		return
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var t Todo
		var createdAt, updatedAt string
		var done int
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &done, &createdAt, &updatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "scan error")
			return
		}
		t.Done = done != 0
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		todos = append(todos, t)
	}
	writeJSON(w, http.StatusOK, todos)
}

func (a *App) createTodo(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(body.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := a.db.ExecContext(r.Context(),
		`INSERT INTO todos (title, description, done, created_at, updated_at) VALUES (?, ?, 0, ?, ?)`,
		body.Title, body.Description, now, now,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		log.Println("createTodo:", err)
		return
	}
	id, _ := res.LastInsertId()
	t := Todo{
		ID:          id,
		Title:       body.Title,
		Description: body.Description,
		Done:        false,
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, now)
	t.UpdatedAt = t.CreatedAt
	writeJSON(w, http.StatusCreated, t)
}

func (a *App) getTodo(w http.ResponseWriter, r *http.Request, id int64) {
	var t Todo
	var createdAt, updatedAt string
	var done int
	err := a.db.QueryRowContext(r.Context(),
		`SELECT id, title, description, done, created_at, updated_at FROM todos WHERE id = ?`, id,
	).Scan(&t.ID, &t.Title, &t.Description, &done, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	t.Done = done != 0
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	writeJSON(w, http.StatusOK, t)
}

func (a *App) updateTodo(w http.ResponseWriter, r *http.Request, id int64) {
	var body struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Done        *bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Fetch current values
	var t Todo
	var createdAt, updatedAt string
	var done int
	err := a.db.QueryRowContext(r.Context(),
		`SELECT id, title, description, done, created_at, updated_at FROM todos WHERE id = ?`, id,
	).Scan(&t.ID, &t.Title, &t.Description, &done, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	t.Done = done != 0
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	if body.Title != nil {
		t.Title = *body.Title
	}
	if body.Description != nil {
		t.Description = *body.Description
	}
	if body.Done != nil {
		t.Done = *body.Done
	}

	doneInt := 0
	if t.Done {
		doneInt = 1
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = a.db.ExecContext(r.Context(),
		`UPDATE todos SET title=?, description=?, done=?, updated_at=? WHERE id=?`,
		t.Title, t.Description, doneInt, now, id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		log.Println("updateTodo:", err)
		return
	}
	t.UpdatedAt, _ = time.Parse(time.RFC3339, now)
	writeJSON(w, http.StatusOK, t)
}

func (a *App) deleteTodo(w http.ResponseWriter, r *http.Request, id int64) {
	res, err := a.db.ExecContext(r.Context(), `DELETE FROM todos WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) markDone(w http.ResponseWriter, r *http.Request, id int64) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := a.db.ExecContext(r.Context(),
		`UPDATE todos SET done=1, updated_at=? WHERE id=?`, now, id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	// Return the updated todo
	a.getTodo(w, r, id)
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()

	// List & create
	mux.HandleFunc("GET /todos", a.listTodos)
	mux.HandleFunc("POST /todos", a.createTodo)

	// Single-item routes: GET, PUT, DELETE /todos/{id}
	mux.HandleFunc("GET /todos/", func(w http.ResponseWriter, r *http.Request) {
		// Could be /todos/{id} or /todos/{id}/done
		path := r.URL.Path
		if strings.HasSuffix(path, "/done") {
			// PATCH is handled separately; GET on /done is not defined
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		id, ok := idFromPath(path)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		a.getTodo(w, r, id)
	})

	mux.HandleFunc("PUT /todos/", func(w http.ResponseWriter, r *http.Request) {
		id, ok := idFromPath(r.URL.Path)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		a.updateTodo(w, r, id)
	})

	mux.HandleFunc("DELETE /todos/", func(w http.ResponseWriter, r *http.Request) {
		id, ok := idFromPath(r.URL.Path)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		a.deleteTodo(w, r, id)
	})

	mux.HandleFunc("PATCH /todos/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if !strings.HasSuffix(path, "/done") {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		// Strip /done suffix for id parsing
		trimmed := strings.TrimSuffix(path, "/done")
		id, ok := idFromPath(trimmed + "/dummy") // idFromPath expects /todos/{id}/...
		if !ok {
			// Try parsing directly
			parts := strings.Split(strings.Trim(trimmed, "/"), "/")
			if len(parts) < 2 {
				writeError(w, http.StatusBadRequest, "invalid id")
				return
			}
			idVal, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid id")
				return
			}
			a.markDone(w, r, idVal)
			return
		}
		a.markDone(w, r, id)
	})

	return mux
}

func main() {
	dbPath := flag.String("db", "todos.db", "SQLite database file")
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "todo-api: open db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := initDB(db); err != nil {
		fmt.Fprintf(os.Stderr, "todo-api: init db: %v\n", err)
		os.Exit(1)
	}

	app := &App{db: db}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("todo-api listening on %s (db: %s)", *addr, *dbPath)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-quit
	log.Println("todo-api shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("todo-api stopped")
}
