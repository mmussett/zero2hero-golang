# Day 32: Idiomatic Go Project Structure

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
