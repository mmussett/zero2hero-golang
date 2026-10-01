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

## Labs

### Lab 1: chi Router Basics

**What you'll practise:** Setting up a `chi.NewRouter()` with basic GET/POST/PUT/DELETE routes and mounting a sub-router.

**Task:**
Create a chi router with a health-check route at `/health` and mount a sub-router at `/api/v1` that has a single `GET /api/v1/ping` route.

**Steps:**
1. `go get github.com/go-chi/chi/v5`
2. Create `r := chi.NewRouter()` and register `GET /health`
3. Use `r.Mount("/api/v1", v1Router())` where `v1Router()` returns a chi router with `GET /ping`
4. Start on `:8080` and test both endpoints

```go
func v1Router() chi.Router {
    r := chi.NewRouter()
    r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"status":"ok"}`))
    })
    return r
}

func main() {
    r := chi.NewRouter()
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
    r.Mount("/api/v1", v1Router())
    http.ListenAndServe(":8080", r)
}
```

**Expected output:**
```
$ curl http://localhost:8080/health
OK
$ curl http://localhost:8080/api/v1/ping
{"status":"ok"}
```

**Checkpoint:** Both URLs return 200; routes outside the mounted prefix return 404.

---

### Lab 2: URL Parameters

**What you'll practise:** Registering `{id}` path parameters with chi and extracting them safely.

**Task:**
Add `GET /api/v1/items/{id}` that reads the `id` parameter. Validate that it is a positive integer; return `400` if it is not, `200` with `{"id":N}` if it is.

**Steps:**
1. Register `r.Get("/items/{id}", getItemHandler)`
2. In the handler, call `chi.URLParam(r, "id")`
3. Parse with `strconv.Atoi`; on failure write 400 JSON error
4. On success write 200 JSON with the numeric id

```go
func getItemHandler(w http.ResponseWriter, r *http.Request) {
    raw := chi.URLParam(r, "id")
    id, err := strconv.Atoi(raw)
    if err != nil || id <= 0 {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "id must be a positive integer"})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]int{"id": id})
}
```

**Expected output:**
```
$ curl http://localhost:8080/api/v1/items/42
{"id":42}
$ curl http://localhost:8080/api/v1/items/abc
{"error":"id must be a positive integer"}
```

**Checkpoint:** Non-numeric and zero/negative IDs return 400; a valid ID returns 200.

---

### Lab 3: Middleware Stack

**What you'll practise:** Applying chi built-in middleware and writing a custom auth middleware.

**Task:**
Apply `middleware.Logger` and `middleware.Recoverer` globally. Write an `authMiddleware` that checks for an `Authorization: Bearer secret` header; reject with 401 if absent or wrong. Apply it only to the `/api/v1` sub-router.

**Steps:**
1. Add `r.Use(middleware.Logger, middleware.Recoverer)` to the root router
2. Write `authMiddleware(next http.Handler) http.Handler` that inspects `r.Header.Get("Authorization")`
3. Apply it with `v1.Use(authMiddleware)` inside the sub-router setup
4. Test with and without the header

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Authorization") != "Bearer secret" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**Expected output:**
```
$ curl http://localhost:8080/api/v1/ping
unauthorized
$ curl -H "Authorization: Bearer secret" http://localhost:8080/api/v1/ping
{"status":"ok"}
```

**Checkpoint:** `/health` does not require auth; all `/api/v1` routes do.

---

### Lab 4: Request/Response Helpers

**What you'll practise:** Writing reusable `respondJSON` and `decodeJSON` helpers to keep handlers clean.

**Task:**
Write two helpers — `respondJSON(w, status, v)` and `decodeJSON(r, v) error` — and use them in a `POST /api/v1/echo` handler that decodes an arbitrary JSON object and echoes it back.

**Steps:**
1. `respondJSON`: set Content-Type, write status, encode `v` with `json.NewEncoder`
2. `decodeJSON`: use `json.NewDecoder` with `DisallowUnknownFields`; limit body size with `http.MaxBytesReader`
3. Wire up `POST /echo` using both helpers
4. Test with valid and invalid JSON bodies

```go
func respondJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func decodeJSON(r *http.Request, v any) error {
    r.Body = http.MaxBytesReader(nil, r.Body, 1<<20) // 1 MB
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    return dec.Decode(v)
}
```

**Expected output:**
```
$ curl -s -X POST -H 'Content-Type: application/json' -d '{"key":"value"}' http://localhost:8080/api/v1/echo
{"key":"value"}
```

**Checkpoint:** The helpers are each used in at least one handler; handlers have no direct `json.NewEncoder` calls of their own.

---

### Lab 5: In-Memory CRUD with sync.RWMutex

**What you'll practise:** Building a thread-safe in-memory store and wiring full CRUD routes.

**Task:**
Create a `NoteStore` backed by a `map[string]Note` protected by `sync.RWMutex`. Implement `List`, `Create`, `Get`, `Update`, `Delete`. Wire up five chi routes.

**Steps:**
1. Define `Note{ID, Text, CreatedAt, UpdatedAt}` and `NoteStore`
2. Implement store methods — `RLock` for reads, `Lock` for writes
3. Register `GET /`, `POST /`, `GET /{id}`, `PUT /{id}`, `DELETE /{id}` on a chi sub-router
4. Test the full CRUD cycle with curl

```go
type Note struct {
    ID        string    `json:"id"`
    Text      string    `json:"text"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type NoteStore struct {
    mu    sync.RWMutex
    notes map[string]Note
    next  int64
}

func (s *NoteStore) Create(text string) Note {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.next++
    n := Note{
        ID:        strconv.FormatInt(s.next, 10),
        Text:      text,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    s.notes[n.ID] = n
    return n
}
```

**Expected output:**
```
$ curl -s -X POST -d '{"text":"hello"}' http://localhost:8080/api/notes
{"id":"1","text":"hello","created_at":"...","updated_at":"..."}
$ curl -s http://localhost:8080/api/notes/1
{"id":"1","text":"hello",...}
$ curl -s -X DELETE http://localhost:8080/api/notes/1
(204 No Content)
```

**Checkpoint:** Concurrent requests do not cause a race — run with `go test -race ./...` to verify.

---

### Lab 6: Testing Handlers with httptest

**What you'll practise:** Writing table-driven handler tests using `httptest.NewRecorder` and `httptest.NewServer`.

**Task:**
Write tests for all five CRUD endpoints. Use `httptest.NewRecorder` for unit tests and `httptest.NewServer` for an integration smoke test.

**Steps:**
1. Write `TestCreateNote` using `httptest.NewRecorder`
2. Write `TestGetNoteNotFound` that asserts a 404 for an unknown ID
3. Write `TestCRUDCycle` using `httptest.NewServer` — create, get, update, delete in sequence
4. Run with `go test -v ./...`

```go
func TestCreateNote(t *testing.T) {
    store := &NoteStore{notes: make(map[string]Note)}
    router := setupRouter(store)

    body := strings.NewReader(`{"text":"test note"}`)
    req := httptest.NewRequest(http.MethodPost, "/api/notes", body)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    if w.Code != http.StatusCreated {
        t.Fatalf("expected 201, got %d", w.Code)
    }
    var note Note
    json.NewDecoder(w.Body).Decode(&note)
    if note.ID == "" {
        t.Fatal("expected non-empty ID")
    }
}
```

**Expected output:**
```
$ go test -v ./...
--- PASS: TestCreateNote (0.00s)
--- PASS: TestGetNoteNotFound (0.00s)
--- PASS: TestCRUDCycle (0.00s)
PASS
```

**Checkpoint:** All tests pass with `go test -race ./...`; no real network ports are opened during unit tests.

---

### Final Lab: Notes REST API with chi

**What you'll practise:** Building a complete chi REST API with five endpoints, an in-memory store, and httptest-based tests.

**Task:**
Wire everything from Labs 1-6 into a production-quality Notes API. Every handler uses `respondJSON`/`decodeJSON`. The store uses `sync.RWMutex`. All five CRUD endpoints have tests.

**Steps:**
1. Define `Note{ID, Text, CreatedAt, UpdatedAt}` and `NoteStore` (mutex + map)
2. Create a chi router with `middleware.Logger`, `middleware.Recoverer`, and your `authMiddleware`
3. Mount the five routes under `/api/notes`
4. Write `respondJSON`, `decodeJSON`, and the five handlers
5. Write `TestCRUDCycle` using `httptest.NewServer` — create, list, get, update, delete

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/notes | List all notes |
| POST | /api/notes | Create a note |
| GET | /api/notes/{id} | Get a note |
| PUT | /api/notes/{id} | Update a note |
| DELETE | /api/notes/{id} | Delete a note |

**Expected output:**
```
$ go run .
listening on :8080
$ curl -s -X POST -d '{"text":"hello"}' http://localhost:8080/api/notes
{"id":"1","text":"hello","created_at":"...","updated_at":"..."}
$ go test -v ./...
--- PASS: TestCRUDCycle (0.00s)
PASS
```

**Checkpoint:** `go test -race ./...` passes with no data race warnings; all five endpoints return correct status codes.

---

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
