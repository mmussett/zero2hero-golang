# Zero to Hero: Go — 35-Day Curriculum

A project-based path from absolute beginner to production-minded Go developer, pairing core concepts with hands-on projects built from scratch.

---

## Week 1: Foundations (Days 1–7)

| Day | Topic | Project |
|-----|-------|---------|
| [01](day-01/README.md) | Toolchain, `go mod`, workspace, packages, `fmt` | Hello, Go Modules |
| [02](day-02/README.md) | Variables, types, control flow, `:=` vs `var` | CLI Calculator |
| [03](day-03/README.md) | Functions, structs, multiple return values, basic errors | Contact Card |
| [04](day-04/README.md) | Pointers, value vs pointer receivers, `new` | Memory Explorer |
| [05](day-05/README.md) | Arrays, slices, strings, `strings` package | String Statistics Tool |
| [06](day-06/README.md) | Maps, method sets, comma-ok idiom, embedded structs | Word Frequency Counter |
| [07](day-07/README.md) | Packages, exported identifiers, `go.mod` and dependencies | Multi-package library |

**Milestone:** You understand Go's core model — value semantics, explicit pointers, and why the language is deliberately small.

---

## Week 2: Core Language (Days 8–14)

| Day | Topic | Project |
|-----|-------|---------|
| [08](day-08/README.md) | Idiomatic project structure: flat, `cmd/`, `internal/`, domain-driven layout | Restructure the notes API |
| [09](day-09/README.md) | Interfaces, duck typing, type assertions, type switches | Shape library |
| [10](day-10/README.md) | Error handling: `fmt.Errorf`, `%w`, `errors.Is`, `errors.As`, custom types | CSV row parser |
| [11](day-11/README.md) | Generics: type parameters, `comparable`, custom constraints | Generic Stack and Queue |
| [12](day-12/README.md) | `sort`, `container/heap`, `container/list`, adjacency-map graphs | Data structures library |
| [13](day-13/README.md) | Closures, function types, higher-order functions, functional options | Data pipeline |
| [14](day-14/README.md) | `go test`, table-driven tests, subtests, `testing.B`, `testify` | Full test suite |

**Milestone:** You can structure a Go project idiomatically and model any domain with interfaces, errors, and generics.

---

## Week 3: Intermediate Patterns (Days 15–21)

| Day | Topic | Project |
|-----|-------|---------|
| [15](day-15/README.md) | `io.Reader`/`io.Writer`, `fmt.Stringer`, `sort.Interface`, embedding | Interface showcase |
| [16](day-16/README.md) | Goroutines, channels, `select`, fan-out / fan-in, `range` over channels | Parallel file hasher |
| [17](day-17/README.md) | `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, `sync.Once`, `sync/atomic` | Thread-safe LRU cache |
| [18](day-18/README.md) | `context` — cancel, timeout, deadline, values; TCP echo server | Cancellable HTTP downloader |
| [19](day-19/README.md) | `os`, `bufio`, `io` — file I/O, walking directories, `io.Copy` | Config file reader |
| [20](day-20/README.md) | `encoding/json`, struct tags, streaming encode/decode, `encoding/csv` | Config manager |
| [21](day-21/README.md) | `flag` package, `cobra` subcommands, stdin/stdout, exit codes | `todo` CLI tool |

**Milestone:** You can write concurrent programs, handle I/O idiomatically, and build CLI tools.

---

## Week 4: Production (Days 22–28)

| Day | Topic | Project |
|-----|-------|---------|
| [22](day-22/README.md) | `net/http` — handler, middleware, `http.Client`, JSON API | HTTP echo server + client |
| [23](day-23/README.md) | REST API with `chi`, JSON middleware, in-memory store | Notes REST API |
| [24](day-24/README.md) | `database/sql` + SQLite driver, prepared statements, transactions | Persist notes to database |
| [25](day-25/README.md) | Error hierarchy: sentinels, typed errors, HTTP error responses | Clean error layer for notes API |
| [26](day-26/README.md) | `log/slog` — structured logging, handlers, trace IDs via context | Structured logs across the notes app |
| [27](day-27/README.md) | `go test -bench`, `b.N`, `-benchmem`, `go tool pprof` | Profile word-frequency counter |
| [28](day-28/README.md) | `//go:embed`, `text/template`, `html/template`, build tags, `go:generate` | Embedded assets server with templates |

**Milestone:** You can serve HTTP, back it with a database, instrument it with structured logs, and benchmark it.

---

## Advanced Golang Concepts (Days 29–35)

| Day | Topic | Project |
|-----|-------|---------|
| [29](day-29/README.md) | `reflect` — `TypeOf`, `ValueOf`, struct tags, `reflect.Kind` | Custom `Describe` utility |
| [30](day-30/README.md) | Versioning, `pkg.go.dev`, Docker multi-stage build, `goreleaser` | Package and ship the shape library |
| [31](day-31/README.md) | Capstone | End-to-end CLI + REST API + database app |
| [32](day-32/README.md) | Interfaces: composition, nil trap, stdlib catalogue, accept-interfaces/return-structs, mocking, anti-patterns | Pluggable notification system |
| [33](day-33/README.md) | Channel model: pipelines, fan-out/in, done channels, `select`, direction types, common mistakes | Channel-based text pipeline + word counter |
| [34](day-34/README.md) | Goroutines & sync: `Mutex`, `RWMutex`, `WaitGroup`, `Once`, `atomic`, `sync.Map`, `errgroup`, deadlocks | Concurrent download manager |
| [35](day-35/README.md) | Context deep dive: tree propagation, values, middleware, DB ops, leak detection, testing patterns | Request-scoped pipeline |

---

## Capstone Options

Choose a scope that fits your time and ambition.

**Small (2–3 days)**
- A `grep` replacement: recursive regex search with coloured output
- A password generator/manager storing entries encrypted on disk
- A Markdown terminal renderer

**Medium (1–2 weeks)**
- A URL shortener: REST API + redirect server + SQLite backing store
- A personal finance tracker: CSV import, categorisation, monthly summaries
- A WebSocket chat server with rooms

**Large (2–4 weeks)**
- A static site generator: Markdown → HTML with templates and live-reload
- A bitcask-inspired key-value database with a WAL and compaction
- A concurrent web crawler with configurable depth and output formats

---

## Recurring Principles

- Build first, read the spec second
- Errors are values — handle them at every call site
- Prefer the standard library before reaching for external packages
- Concurrency is not parallelism — understand the difference before using either
- Iterate on earlier projects rather than rewriting from scratch
