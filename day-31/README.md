# Day 31: Capstone

## You Made It

Thirty days of deliberate practice. You have built:

- A concurrent file hasher, a thread-safe LRU cache, a cancellable HTTP downloader
- A full CRUD REST API with middleware, structured logging, and a SQLite backing store
- A CLI tool with subcommands
- Benchmarks, profiles, and an embedded-assets server

Now build something that is entirely yours.

## Skills You Have

| Week | What You Can Do Now |
|------|---------------------|
| 1 | Model any domain with structs, slices, maps, and packages |
| 2 | Write idiomatic interfaces, handle errors cleanly, use generics |
| 3 | Write concurrent programs, build CLI tools, serve HTTP |
| 4 | Build, instrument, profile, containerise, and publish a Go service |

## Capstone Options

**Small (2–3 days)** — CLI tools
- A `grep` replacement: recursive regex search with coloured, numbered output
- A password generator/manager storing entries encrypted with `crypto/aes`
- A Markdown renderer that outputs ANSI-formatted text to the terminal

**Medium (1–2 weeks)** — Web services
- A URL shortener: `POST /shorten` returns a short code, `GET /{code}` redirects; backed by SQLite
- A personal finance tracker: CSV import, category tagging, monthly summaries as a REST API
- A WebSocket chat server with named rooms using `golang.org/x/net/websocket`

**Large (2–4 weeks)** — Substantial projects
- A static site generator: Markdown → HTML with templates, partials, and `--serve` live-reload
- A bitcask-inspired key-value database: WAL, compaction, HTTP API
- A concurrent web crawler: configurable depth, robots.txt respect, deduplicated output

## Planning Checklist

Before writing code:

- [ ] Define your core types (`Note`, `Account`, `Page`, …)
- [ ] Choose dependencies deliberately — prefer the standard library
- [ ] Design your error types with custom sentinels
- [ ] Write at least one test before any handler
- [ ] Initialise `slog` on day one
- [ ] Write a `Dockerfile` before shipping

## Labs

### Lab 1: Project Planning — Pick Your Capstone

**What you'll practise:** Thinking through scope, data model, and interface before writing a single line of code.

**Task:**
Review the six capstone options above. Pick one, then sketch its data model and API or command surface on paper (or in a text file) before opening your editor.

**Steps:**
1. Re-read the Capstone Options section
2. Choose a project (Small, Medium, or Large) that fits your available time
3. Write down your core types — e.g. `type Bookmark struct { ID int; URL string; Tags []string }`
4. List every API endpoint or CLI subcommand your project will expose
5. Identify the one external dependency you expect to need (if any)

```
# Example sketch for a URL bookmark manager
Types:   Bookmark{ID, URL, Title, Tags, CreatedAt}
Storage: JSON file (~day-19 pattern) or SQLite (~day-23 pattern)
CLI:     add <url> [--tag t]   list [--tag t]   delete <id>   export
HTTP:    POST /bookmarks        GET /bookmarks   DELETE /bookmarks/{id}
```

**Expected output:**
A clear, written sketch you can refer to throughout the following labs. No code yet.

**Checkpoint:** You can describe your project's core type, its storage mechanism, and its interface in three sentences without looking at notes.

---

### Lab 2: Module Setup — Scaffolding Your Project

**What you'll practise:** Initialising a Go module, creating the directory structure, and writing stub files that compile cleanly.

**Task:**
Create the module and directory layout for your chosen capstone. Every file should compile (stubs are fine), and `go build ./...` should succeed.

**Steps:**
1. `mkdir capstone/myproject && cd capstone/myproject`
2. `go mod init github.com/yourname/myproject`
3. Create stub files for each package: `main.go`, `store/store.go`, `model/model.go`
4. Add your module to `go.work`: `go work use ./capstone/myproject`
5. Run `go build ./...` and confirm no errors

```
myproject/
├── go.mod
├── main.go          ← package main, imports store and model
├── model/
│   └── model.go     ← core types
└── store/
    └── store.go     ← storage interface + implementation stub
```

```go
// store/store.go
package store

import "github.com/yourname/myproject/model"

type Store interface {
    Add(item model.Bookmark) (model.Bookmark, error)
    List() ([]model.Bookmark, error)
    Delete(id int) error
}
```

**Expected output:**
```
$ go build ./...
(no output — success)
```

**Checkpoint:** `go build ./...` exits 0. The module appears in `go.work`. Directory structure matches your sketch from Lab 1.

---

### Lab 3: Core Data Layer — Storage with Unit Tests

**What you'll practise:** Implementing the storage layer (JSON file, SQLite, or in-memory) and verifying it with unit tests before wiring any interface.

**Task:**
Implement the `Store` interface from Lab 2 with a concrete type (e.g. `JSONStore` or `MemStore`). Write at least three unit tests: add-then-list, delete-then-confirm-missing, and list-empty.

**Steps:**
1. Implement `Add`, `List`, and `Delete` on your concrete store type
2. Write `store/store_test.go` with `TestAdd`, `TestDelete`, `TestListEmpty`
3. Run `go test ./store/...` — all tests must pass
4. Use table-driven tests for `Delete` (existing ID, non-existing ID)

```go
// store/store_test.go
func TestAdd(t *testing.T) {
    s := NewMemStore()
    got, err := s.Add(model.Bookmark{URL: "https://go.dev"})
    if err != nil { t.Fatal(err) }
    if got.ID == 0 { t.Error("expected non-zero ID") }
}

func TestDeleteMissing(t *testing.T) {
    s := NewMemStore()
    err := s.Delete(999)
    if !errors.Is(err, ErrNotFound) {
        t.Errorf("want ErrNotFound, got %v", err)
    }
}
```

**Expected output:**
```
$ go test -v ./store/...
--- PASS: TestAdd (0.00s)
--- PASS: TestDeleteMissing (0.00s)
--- PASS: TestListEmpty (0.00s)
PASS
```

**Checkpoint:** All tests pass. The storage layer has no dependency on HTTP or CLI code.

---

### Lab 4: Business Logic — TDD Core Operations

**What you'll practise:** Writing tests before implementation (TDD) for the core business rules of your capstone.

**Task:**
Identify two or three business rules in your project (e.g. "duplicate URLs are rejected", "tags are normalised to lowercase", "pagination returns at most 20 items"). Write the test first, watch it fail, then implement.

**Steps:**
1. Pick two business rules specific to your chosen capstone
2. Write a failing test for each rule in `model/` or `store/`
3. Implement just enough code to make each test pass
4. Refactor if needed, keeping tests green

```go
// Example: duplicate URL rejection
func TestAddDuplicateURL(t *testing.T) {
    s := NewMemStore()
    _, err := s.Add(model.Bookmark{URL: "https://go.dev"})
    if err != nil { t.Fatal(err) }

    _, err = s.Add(model.Bookmark{URL: "https://go.dev"})
    if !errors.Is(err, ErrDuplicate) {
        t.Errorf("expected ErrDuplicate, got %v", err)
    }
}
```

**Expected output:**
```
$ go test ./...
--- FAIL: TestAddDuplicateURL (0.00s)    ← before implementation
    store_test.go:22: expected ErrDuplicate, got <nil>
```
Then after implementing:
```
--- PASS: TestAddDuplicateURL (0.00s)
```

**Checkpoint:** You wrote the test before the implementation and watched it go from red to green. All existing tests still pass.

---

### Lab 5: Interface Layer — CLI Commands or HTTP Routes

**What you'll practise:** Wiring the storage layer to a CLI (`cobra` or `flag`) or HTTP routes (`net/http` or `chi`), keeping handlers thin.

**Task:**
Implement the interface layer for your capstone. Handlers and commands should contain no business logic — they parse input, call the store, and format output.

**Steps:**
1. For CLI: add subcommands (e.g. `add`, `list`, `delete`) using `flag` subcommands or `cobra`
2. For HTTP: register routes, parse JSON request bodies, return JSON responses with correct status codes
3. Keep each handler/command under 20 lines — push logic to the store
4. Run `go run . add https://go.dev --tag golang` (CLI) or `curl -X POST /bookmarks` (HTTP) to smoke-test

```go
// Thin HTTP handler example
func (h *Handler) handleAdd(w http.ResponseWriter, r *http.Request) {
    var b model.Bookmark
    if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    created, err := h.store.Add(b)
    if err != nil {
        http.Error(w, err.Error(), http.StatusConflict)
        return
    }
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(created)
}
```

**Expected output:**
```
$ go run . add https://go.dev --tag golang
Added bookmark #1: https://go.dev [golang]

$ go run . list
#1  https://go.dev  [golang]
```

**Checkpoint:** At least two commands or routes work end-to-end. No business logic lives in the handler or command functions.

---

### Lab 6: Error Handling and Logging — Production Hygiene

**What you'll practise:** Applying Day 24 sentinel error patterns and Day 25 structured logging throughout the project.

**Task:**
Add custom sentinel errors, wrap errors with context at every layer boundary, and add `slog` logging so every important operation emits a structured log entry.

**Steps:**
1. Define sentinel errors in `store/`: `var ErrNotFound = errors.New("not found")`, `var ErrDuplicate = errors.New("duplicate")`
2. Wrap errors at the store boundary: `fmt.Errorf("store.Add: %w", ErrDuplicate)`
3. Initialise a JSON `slog.Logger` in `main` with the service name as a permanent field
4. Log every add, delete, and list operation at `Info`; log errors at `Error` with the `"err"` attribute
5. Make log level configurable via `LOG_LEVEL` env var

```go
// In main
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: logLevelFromEnv(),
})).With("service", "bookmarks")
slog.SetDefault(logger)

// In handler
if errors.Is(err, store.ErrNotFound) {
    slog.Error("bookmark not found", "id", id)
    http.Error(w, "not found", http.StatusNotFound)
    return
}
```

**Expected output:**
```
{"time":"...","level":"INFO","msg":"bookmark added","service":"bookmarks","id":1,"url":"https://go.dev"}
{"time":"...","level":"ERROR","msg":"bookmark not found","service":"bookmarks","id":999}
```

**Checkpoint:** Every error path logs at `Error` level with an `"err"` attribute. Happy paths log at `Info`. `errors.Is` correctly identifies sentinel errors across layer boundaries.

---

### Lab 7: Polish and Ship — Dockerfile, Graceful Shutdown, --version

**What you'll practise:** Making your capstone production-ready with a multi-stage Dockerfile, graceful HTTP shutdown, and embedded version info.

**Task:**
Add the final production touches: a multi-stage Dockerfile, graceful shutdown on `SIGTERM`/`SIGINT`, and a `--version` flag with embedded build info.

**Steps:**
1. Add `var version = "dev"` and a `--version` flag (see Day 29 Lab 1)
2. Implement graceful shutdown using `http.Server.Shutdown` with a context timeout
3. Write a multi-stage `Dockerfile` (alpine builder, scratch or alpine runtime)
4. Write a `README.md` with installation, usage examples, and environment variables

```go
// Graceful shutdown
srv := &http.Server{Addr: addr, Handler: mux}
go func() {
    if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
        slog.Error("server error", "err", err)
        os.Exit(1)
    }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
    slog.Error("shutdown error", "err", err)
}
slog.Info("server stopped gracefully")
```

**Expected output:**
```
$ ./myapp --version
bookmarks v1.0.0 (built 2024-01-15T10:00:00Z)

$ docker build -t bookmarks . && docker run -p 8080:8080 bookmarks
{"time":"...","level":"INFO","msg":"server started","addr":":8080","version":"v1.0.0"}
```

**Checkpoint:** `docker run` starts the server. `curl /health` returns version info. Sending `Ctrl-C` triggers a clean shutdown log line within 10 seconds.

---

### Final Lab (Project): Your Capstone — URL Bookmark Manager (Example)

**What you'll practise:** Synthesising all 30 days into a complete, tested, containerised Go application.

**Task:**
Complete your chosen capstone project, using the URL bookmark manager as the reference example. The finished project must have: working storage, a tested business layer, an HTTP or CLI interface, structured logging, and a Dockerfile.

**Steps:**
1. Review your Lab 1 sketch — does the implementation match?
2. Ensure `go test ./...` passes with no failures
3. Write at least one integration test that exercises the full stack (handler → store)
4. Run `docker build` and smoke-test every endpoint or command from the `README.md`
5. Tag `v1.0.0` locally: `git tag v1.0.0`

```bash
# Smoke-test sequence for the bookmark manager example
curl -s -X POST http://localhost:8080/bookmarks \
  -H "Content-Type: application/json" \
  -d '{"url":"https://go.dev","tags":["golang"]}'

curl -s http://localhost:8080/bookmarks | jq .

curl -s -X DELETE http://localhost:8080/bookmarks/1

curl -s http://localhost:8080/health
```

**Expected output:**
```
{"id":1,"url":"https://go.dev","tags":["golang"],"created_at":"2024-01-15T10:00:00Z"}
[{"id":1,"url":"https://go.dev","tags":["golang"],"created_at":"..."}]
(empty — 204 No Content)
{"status":"ok","version":"v1.0.0","build_time":"2024-01-15T10:00:00Z"}
```

**Checkpoint:** All four smoke-test commands succeed. `go test ./...` is green. `docker build` succeeds and the image is under 20 MB.

**Extension ideas:** add a WebSocket "new bookmark" live feed; implement CSV export; write a GitHub Actions workflow that runs tests and builds the Docker image on every push.

## Closing Thought

Go's difficulty is front-loaded: the module system, explicit error handling, and concurrency model feel unfamiliar at first. But these constraints prevent entire categories of bugs. Your Go code is honest about what can fail, who owns the data, and where the concurrency lives.

That honesty is the point.

## Official Documentation

- [`crypto/aes`](https://pkg.go.dev/crypto/aes) — AES encryption for the password manager capstone option
- [`net/http`](https://pkg.go.dev/net/http) — HTTP server and client for web service capstones
- [`database/sql`](https://pkg.go.dev/database/sql) — SQL database layer for URL shortener and finance tracker
- [`log/slog`](https://pkg.go.dev/log/slog) — structured logging recommended from day one of any capstone
- [`text/template`](https://pkg.go.dev/text/template) — template rendering for static site generator capstone
- [`encoding/json`](https://pkg.go.dev/encoding/json) — JSON serialisation used across all web service capstones
- [Effective Go](https://go.dev/doc/effective_go) — canonical guide to idiomatic Go
- [Go Tour](https://go.dev/tour/) — interactive refresher on any concept
- [Go standard library](https://pkg.go.dev/std) — full index of all standard packages
