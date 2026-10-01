# Day 08: Idiomatic Go Project Structure

## Core Concept: Start Flat, Grow Deliberately

Go has no mandatory project layout. The language team has explicitly declined to prescribe one. But idiomatic Go projects share patterns distilled from years of open-source Go code.

> "The most important thing is that the code works, is readable, and can be maintained." — Go team

---

## Part 1: The Flat Package (Start Here)

For small projects and CLIs — one or a few packages, everything at the root:

```
myapp/
├── go.mod
├── go.sum
├── main.go
├── handler.go
├── store.go
├── store_test.go
├── config.go
└── README.md
```

**When to use:** scripts, small tools, single-binary apps with < 5 source files. No subdirectories until a clear seam emerges. Resist premature structure.

---

## Part 2: Multiple Binaries with `cmd/`

When your module produces more than one binary:

```
myproject/
├── go.mod
├── cmd/
│   ├── api/
│   │   └── main.go        ← package main (the API server)
│   ├── worker/
│   │   └── main.go        ← package main (background worker)
│   └── migrate/
│       └── main.go        ← package main (DB migration tool)
├── internal/
│   ├── store/
│   │   ├── store.go
│   │   └── store_test.go
│   └── config/
│       └── config.go
└── README.md
```

Each directory under `cmd/` has its own `package main`. All shared logic lives elsewhere — `cmd/` contains only `main.go` (wiring, flags, startup).

```bash
go build ./cmd/api
go build ./cmd/worker
go install ./cmd/...     # install all binaries
```

---

## Part 3: `internal/` — Enforced Encapsulation

`internal/` is a Go compiler-enforced boundary. Only code in the **parent** of `internal/` can import it.

```
myproject/
├── internal/
│   ├── store/        ← importable only by myproject/...
│   ├── auth/
│   └── metrics/
├── cmd/api/
└── pkg/             ← importable by anyone (if you publish it)
```

Use `internal/` for:
- Implementation details you don't want to promise are stable
- Packages shared across your `cmd/` binaries but not meant for external use
- Database models, service layer, business logic

---

## Part 4: The `pkg/` Directory

`pkg/` signals "this is a public library" — it is importable by external modules. It is optional and slightly controversial:

```
myproject/
├── pkg/
│   ├── retry/        ← stable, public API
│   └── httputil/     ← stable, public API
└── internal/
    └── auth/         ← private implementation
```

**Skip `pkg/` if you are not publishing a library.** Many Go projects use `internal/` for everything and have no `pkg/`.

---

## Part 5: Domain-Driven Layout (for Larger Services)

Organise by **domain concept**, not by layer:

```
notes-api/
├── go.mod
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── note/           ← "note" domain
│   │   ├── note.go         ← types: Note, NoteID
│   │   ├── service.go      ← business logic
│   │   ├── service_test.go
│   │   ├── store.go        ← Store interface
│   │   └── handler.go      ← HTTP handlers
│   ├── user/           ← "user" domain
│   │   ├── user.go
│   │   └── ...
│   └── platform/       ← cross-cutting concerns
│       ├── database/
│       │   └── database.go
│       ├── logger/
│       │   └── logger.go
│       └── config/
│           └── config.go
└── README.md
```

**Why by domain, not by layer?**

Layer-based layout (`controllers/`, `models/`, `repositories/`) scatters one feature across many directories. Domain-based layout keeps all code for a concept together — easy to delete, move, or extract a service.

---

## Part 6: What NOT to Do

### Anti-Pattern: `util/` or `helpers/`

```
// Bad — this package will accumulate everything with no coherent purpose
internal/util/util.go
internal/helpers/string_helpers.go
```

If a function is general enough to live in `util`, it belongs in a specifically named package (`stringutil`, `timeutil`) or in the nearest calling package.

### Anti-Pattern: Circular Imports

Go does not allow circular imports. If `packageA` imports `packageB` and `packageB` imports `packageA`, it will not compile. Circular imports are a sign of poor package decomposition.

Fix: extract shared types into a third package, or merge the packages.

### Anti-Pattern: One Package Per File

```
// Bad — splitting trivially by file, not by responsibility
internal/note_model/note.go
internal/note_service/service.go
internal/note_handler/handler.go
```

Group by **concept**, not by file type. `internal/note/` should contain the model, service, and handler for notes.

### Anti-Pattern: `github.com/golang-standards/project-layout`

This widely-cited repository is **not official**. The Go team does not endorse it. Following it blindly adds directories like `deployments/`, `scripts/`, `build/` before you need them. Start with what you have; add structure when the pain demands it.

---

## Part 7: Configuration and Wiring in `main`

`main.go` should be the **only** place where the app is wired together — reading config, creating dependencies, connecting layers. It should not contain business logic.

```go
// cmd/api/main.go
func main() {
    cfg := config.Load()               // read env/file
    db  := database.Open(cfg.DSN)      // create DB
    defer db.Close()

    noteStore   := note.NewSQLStore(db)
    noteService := note.NewService(noteStore)
    noteHandler := note.NewHandler(noteService)

    r := chi.NewRouter()
    r.Mount("/api/notes", noteHandler.Routes())

    srv := &http.Server{Addr: cfg.Addr, Handler: r}
    log.Fatal(srv.ListenAndServe())
}
```

Each constructor takes its dependencies as arguments — dependency injection without a framework.

---

## Part 8: File Naming Conventions

| File | Contents |
|------|----------|
| `type.go` or `note.go` | Core types for the package |
| `store.go` | `Store` interface + implementations |
| `service.go` | Business logic |
| `handler.go` | HTTP handlers |
| `middleware.go` | HTTP middleware |
| `config.go` | Configuration types |
| `*_test.go` | Tests (same or separate package) |

There is no `interface.go` or `model.go` convention — put types in the file that most directly relates to them.

---

## Part 9: Testing Layout

```
internal/note/
├── note.go
├── service.go
├── service_test.go    ← package note (white-box: can access unexported symbols)
├── store.go
└── store_integration_test.go  ← package note_test (black-box)
```

Use `package foo_test` (external test package) for integration tests that should only use the public API. Use `package foo` for unit tests that need unexported symbols.

---

## Part 10: A Real-World Reference Layout

The notes API from Days 22–25, restructured idiomatically:

```
notes-api/
├── go.mod
├── go.sum
├── Dockerfile
├── .goreleaser.yaml
├── README.md
├── cmd/
│   └── api/
│       └── main.go          ← wiring only
├── internal/
│   ├── note/
│   │   ├── note.go          ← Note type, NoteID
│   │   ├── store.go         ← Store interface
│   │   ├── sqlstore.go      ← *SQLStore implements Store
│   │   ├── service.go       ← business logic
│   │   ├── handler.go       ← HTTP handlers
│   │   └── handler_test.go
│   └── platform/
│       ├── config/
│       │   └── config.go
│       ├── database/
│       │   └── database.go
│       └── logger/
│           └── logger.go
└── migrations/
    └── 001_create_notes.sql
```

---

---

## Labs

### Lab 1: Flat Structure

**What you'll practise:** Building a small web server entirely in `package main` to feel the limits of flat structure firsthand.

**Task:**
Create a minimal notes server with handler, in-memory store, and server wiring — all in `main.go`. No subdirectories.

**Steps:**
1. In `day-32/`, add to `main.go` with `package main`
2. Define a `Note` struct with `ID int`, `Title string`, `Body string`
3. Create a `MemStore` struct with a `[]Note` field and a `sync.Mutex`
4. Add `Add(n Note) Note`, `All() []Note`, and `FindByID(id int) (Note, bool)` methods
5. Add two HTTP handlers: `GET /notes` (list all) and `POST /notes` (create); wire in `main()`

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
)

type Note struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
    Body  string `json:"body"`
}

type MemStore struct {
    mu    sync.Mutex
    notes []Note
    next  int
}
```

**Expected output:**
```
Server listening on :8080
# curl localhost:8080/notes            → []
# curl -XPOST localhost:8080/notes \
#   -d '{"title":"hello"}'            → {"id":1,"title":"hello","body":""}
# curl localhost:8080/notes            → [{"id":1,...}]
```

**Checkpoint:** Count the lines. Once the file exceeds ~150–200 lines, navigation becomes painful. That is the natural inflection point for introducing structure — not before.

---

### Lab 2: The `cmd/` Pattern

**What you'll practise:** Moving `main()` to `cmd/notes/main.go` so the root package becomes a reusable library.

**Task:**
Refactor the flat server from Lab 1. Move only the wiring to `cmd/notes/main.go`. Keep `Note`, `MemStore`, and handler registration in the root package (renamed from `main` to a library package).

**Steps:**
1. Rename the root package from `package main` to `package notes` in all non-main files
2. Create `day-32/cmd/notes/main.go` with `package main`
3. Move `main()` and only the startup/wiring code there
4. Import the root library package from `cmd/notes/main.go`
5. Build: `go build ./cmd/notes`

```go
// cmd/notes/main.go
package main

import (
    notes "day32"
    "net/http"
)

func main() {
    store := notes.NewMemStore()
    mux := http.NewServeMux()
    notes.RegisterHandlers(mux, store)
    http.ListenAndServe(":8080", mux)
}
```

**Expected output:**
```
go build ./cmd/notes  ← succeeds, produces a binary
./notes               ← server starts on :8080
```

**Checkpoint:** Verify `go build ./...` compiles both the library package and the binary. Confirm the library package contains no `main()` function.

---

### Lab 3: `internal/` Package

**What you'll practise:** Using `internal/` to enforce encapsulation — the compiler prevents external modules from importing it.

**Task:**
Create `internal/store/store.go`. Import it successfully from within the module. Then attempt to import it from a throwaway module outside `day-32/` to observe the compiler error.

**Steps:**
1. Create `day-32/internal/store/store.go` with `package store`
2. Move `MemStore` into it
3. Import it from `cmd/notes/main.go` — this works (same module tree)
4. Create a temporary directory outside `day-32/` with its own `go.mod` and attempt the import
5. Record the exact compiler error message

```go
// day-32/internal/store/store.go
package store

type Note struct { ID int; Title, Body string }

type MemStore struct {
    mu    sync.Mutex
    notes []Note
    next  int
}

func NewMemStore() *MemStore { return &MemStore{next: 1} }
```

**Expected output:**
```
# From cmd/notes inside day-32: builds fine

# From an outside module:
./main.go:5:2: use of internal package day32/internal/store not allowed
```

**Checkpoint:** Read the Go spec on internal packages. State the rule in a comment: only code rooted at the **parent** of `internal/` may import it.

---

### Lab 4: Domain-Driven Layout

**What you'll practise:** Reorganising a codebase into a domain-driven structure with `internal/note/` and `internal/platform/`.

**Task:**
Restructure `day-32/` into the domain-driven layout. Note types, store, and handlers all live under `internal/note/`. Config lives under `internal/platform/config/`.

**Steps:**
1. Create `internal/note/note.go` — the `Note` type
2. Create `internal/note/store.go` — `Store` interface + `MemStore`
3. Create `internal/note/handler.go` — HTTP handlers
4. Create `internal/platform/config/config.go` — `Config{Addr string}`
5. Update `cmd/notes/main.go` to import from the internal packages and wire them explicitly

```
day-32/
├── go.mod
├── cmd/notes/main.go
└── internal/
    ├── note/
    │   ├── note.go
    │   ├── store.go
    │   └── handler.go
    └── platform/
        └── config/
            └── config.go
```

**Expected output:**
```
go build ./...  ← all packages compile
go vet ./...    ← no issues reported
```

**Checkpoint:** Open `internal/note/handler.go`. It must import `internal/note` (the store interface) but nothing from `cmd/`. Confirm the import graph has no cycles.

---

### Lab 5: Dependency Injection

**What you'll practise:** Passing the store as an interface to the handler constructor, decoupling HTTP from storage and enabling test doubles.

**Task:**
Define a `Store` interface in `internal/note/`. Pass it into `NewHandler(s Store)`. Write a test using an in-memory fake — no real store required.

**Steps:**
1. In `internal/note/store.go`, define `type Store interface { Add(Note) Note; All() []Note; FindByID(int) (Note, bool) }`
2. Ensure `MemStore` implements `Store`; add a compile-time check `var _ Store = (*MemStore)(nil)`
3. `NewHandler(s Store) *Handler` — the handler holds the interface, not the concrete type
4. In `internal/note/handler_test.go`, define `type fakeStore struct{...}` implementing `Store`
5. Test `GET /notes` with zero, one, and three notes using only the fake

```go
// internal/note/store.go
type Store interface {
    Add(n Note) Note
    All() []Note
    FindByID(id int) (Note, bool)
}

// internal/note/handler.go
type Handler struct{ store Store }
func NewHandler(s Store) *Handler { return &Handler{store: s} }
```

**Expected output:**
```
--- PASS: TestListNotes/empty (0.00s)
--- PASS: TestListNotes/one_note (0.00s)
--- PASS: TestListNotes/three_notes (0.00s)
```

**Checkpoint:** Delete the `MemStore` import from `handler_test.go` — tests must still compile and pass using only the fake store.

---

### Lab 6: Avoiding Circular Imports

**What you'll practise:** Recognising a circular import error and breaking the cycle using a shared interface package.

**Task:**
Deliberately create a cycle between two packages. Observe the compiler error. Break it by extracting a shared interface into a third package.

**Steps:**
1. Create `internal/a/a.go` importing `day32/internal/b`
2. Create `internal/b/b.go` importing `day32/internal/a`
3. Run `go build ./...` — observe the import cycle error
4. Create `internal/types/types.go` with a shared interface both packages need
5. Change `a` and `b` to import `types` instead of each other; rebuild

```go
// internal/a/a.go — broken
package a
import "day32/internal/b"  // cycle!

// internal/b/b.go — broken
package b
import "day32/internal/a"  // cycle!
```

**Expected output:**
```
# Broken:
import cycle not allowed:
  day32/internal/a → day32/internal/b → day32/internal/a

# Fixed (a and b both import types, not each other):
go build ./...  ← succeeds
```

**Checkpoint:** Draw the import graph before and after the fix. Confirm no package in the fixed version imports a package that imports it back.

---

### Lab 7: Module Boundaries

**What you'll practise:** Understanding the decision criteria for splitting code into multiple packages vs multiple modules.

**Task:**
This is a design analysis lab. Evaluate three real-world scenarios, decide packages vs modules, and record your reasoning as Go comments in a `decisions.go` file.

**Steps:**
1. **Scenario A:** A CLI tool and a shared config library in the same repo, released together → same module, separate packages
2. **Scenario B:** A utility library published on pkg.go.dev that other teams import and pin by version → separate module with its own `go.mod`
3. **Scenario C:** A monorepo with 5 microservices sharing a `common/` library → `go.work` workspace, one module per service plus one for `common/`
4. Create `day-32/decisions.go` (`package main`) containing your reasoning as comments
5. Run `go vet ./...` to confirm the file is valid Go

```go
// decisions.go
package main

// Scenario A: CLI + shared config — same module, separate packages
// Reason: released together; no independent versioning needed.
// Layout: myapp/ (go.mod), myapp/cmd/cli/, myapp/internal/config/

// Scenario B: Published library
// Reason: independent semantic versioning; external consumers pin specific tags.
// Layout: separate repo, own go.mod, v2+ in module path when breaking changes ship.

// Scenario C: Monorepo with 5 services
// Reason: go.work lets each service module resolve dependencies locally during dev.
// Layout: go.work at root; service-a/go.mod, service-b/go.mod, common/go.mod.
```

**Expected output:**
```
go vet ./...  ← no output (no issues)
```

**Checkpoint:** Look up two open-source Go projects (e.g., `go-chi/chi` and `google/go-cloud`). Which layout do they use? Add a short note to your comments.

---

## Day Project: Restructure the Notes API

Take the notes API built in Days 22–25 and restructure it into the domain-driven layout above:

1. Create `cmd/api/main.go` that wires everything with explicit constructors
2. Move types to `internal/note/note.go`
3. Extract `Store` interface to `internal/note/store.go`; SQL implementation to `sqlstore.go`
4. Move business logic to `internal/note/service.go`
5. Move HTTP handlers to `internal/note/handler.go`
6. Move config loading to `internal/platform/config/config.go`
7. Verify everything compiles: `go build ./...`
8. Verify tests still pass: `go test ./...`

**Extension ideas:** add a second binary `cmd/migrate/main.go` that runs DB migrations standalone; add `internal/platform/logger/` that initialises `slog` with level from config.

## Official Documentation

- [`net/http`](https://pkg.go.dev/net/http) — `Server`, `ListenAndServe`, `Handler` — wired in `cmd/api/main.go`
- [`log`](https://pkg.go.dev/log) — `Fatal` for startup errors
- [`log/slog`](https://pkg.go.dev/log/slog) — structured logger initialised in `internal/platform/logger/`
- [`database/sql`](https://pkg.go.dev/database/sql) — `DB` opened in `internal/platform/database/`
- [Go Modules Reference — internal packages](https://go.dev/ref/mod#go-mod-file) — compiler enforcement of `internal/` boundaries
- [Go Blog: Organizing a Go module](https://go.dev/blog/organizing-go-code)
- [Go Blog: Package names](https://go.dev/blog/package-names) — naming conventions for packages
- [Language Spec — Package clause](https://go.dev/ref/spec#Package_clause)
- [cmd/go — internal directories](https://pkg.go.dev/cmd/go#hdr-Internal_Directories) — how the compiler enforces `internal/`
