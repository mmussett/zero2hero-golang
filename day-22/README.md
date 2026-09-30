# Day 22: REST API with chi

## Core Concept: Handlers Are Composable Functions

`chi` is a lightweight router that adds path parameters, middleware chaining, and subrouters to `net/http` while remaining 100% compatible with the standard library.

## Router Setup

```go
import "github.com/go-chi/chi/v5"
import "github.com/go-chi/chi/v5/middleware"

r := chi.NewRouter()
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)
r.Use(middleware.SetHeader("Content-Type", "application/json"))

r.Route("/api/notes", func(r chi.Router) {
    r.Get("/",         listNotes)
    r.Post("/",        createNote)
    r.Get("/{id}",    getNote)
    r.Put("/{id}",    updateNote)
    r.Delete("/{id}", deleteNote)
})

http.ListenAndServe(":8080", r)
```

## Path Parameters

```go
func getNote(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    // ...
}
```

## In-Memory Store with Mutex

```go
type Store struct {
    mu    sync.RWMutex
    notes map[string]Note
    seq   int64
}

func (s *Store) Get(id string) (Note, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    n, ok := s.notes[id]
    return n, ok
}

func (s *Store) Create(text string) Note {
    s.mu.Lock()
    defer s.mu.Unlock()
    id := fmt.Sprintf("%d", atomic.AddInt64(&s.seq, 1))
    n := Note{ID: id, Text: text, CreatedAt: time.Now()}
    s.notes[id] = n
    return n
}
```

## JSON Helpers

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    return dec.Decode(v)
}
```

## Testing HTTP Handlers

```go
func TestGetNote(t *testing.T) {
    store := NewStore()
    note := store.Create("test note")

    r := setupRouter(store)
    srv := httptest.NewServer(r)
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/api/notes/" + note.ID)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Day Project: Notes REST API

Build a full CRUD REST API for a notes app:

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/notes | List all notes |
| POST | /api/notes | Create a note |
| GET | /api/notes/{id} | Get a note |
| PUT | /api/notes/{id} | Update a note |
| DELETE | /api/notes/{id} | Delete a note |

Each `Note` has: `id`, `text`, `created_at`, `updated_at`. Include integration tests using `httptest.NewServer`.

Run with: `go run .`

**Extension ideas:** add pagination with `?limit=10&offset=0`; add tag filtering.

## Official Documentation

- [`net/http`](https://pkg.go.dev/net/http) — `ResponseWriter`, `Request`, `Handler`, `HandlerFunc`, `ListenAndServe`, `StatusOK`, `StatusNotFound`
- [`net/http/httptest`](https://pkg.go.dev/net/http/httptest) — `NewServer` for integration testing HTTP handlers
- [`encoding/json`](https://pkg.go.dev/encoding/json) — `NewEncoder`, `NewDecoder`, `DisallowUnknownFields`
- [`sync`](https://pkg.go.dev/sync) — `RWMutex` for thread-safe in-memory store
- [`sync/atomic`](https://pkg.go.dev/sync/atomic) — `AddInt64` for generating sequential IDs
- [`time`](https://pkg.go.dev/time) — `Time`, `Now` for `created_at` / `updated_at` timestamps
- [chi router](https://pkg.go.dev/github.com/go-chi/chi/v5) — `NewRouter`, `URLParam`, `Route`, middleware chaining
- [Go Blog: Writing Web Applications](https://go.dev/doc/articles/wiki)
