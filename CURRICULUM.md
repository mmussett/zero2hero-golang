# Zero to Hero: Go — 34-Day Curriculum

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
| [08](day-08/README.md) | Interfaces, duck typing, type assertions, type switches | Shape library |
| [09](day-09/README.md) | Error handling: `fmt.Errorf`, `%w`, `errors.Is`, `errors.As`, custom types | CSV row parser |
| [10](day-10/README.md) | Generics: type parameters, `comparable`, custom constraints | Generic Stack and Queue |
| [11](day-11/README.md) | `sort`, `container/heap`, `container/list`, adjacency-map graphs | Data structures library |
| [12](day-12/README.md) | Closures, function types, higher-order functions, functional options | Data pipeline |
| [13](day-13/README.md) | `go test`, table-driven tests, subtests, `testing.B`, `testify` | Full test suite |
| [14](day-14/README.md) | `io.Reader`/`io.Writer`, `fmt.Stringer`, `sort.Interface`, embedding | Interface showcase |

**Milestone:** You can model any domain idiomatically and handle errors without panicking.

---

## Week 3: Intermediate Patterns (Days 15–21)

| Day | Topic | Project |
|-----|-------|---------|
| [15](day-15/README.md) | Goroutines, channels, `select`, fan-out / fan-in, `range` over channels | Parallel file hasher |
| [16](day-16/README.md) | `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, `sync.Once`, `sync/atomic` | Thread-safe LRU cache |
| [17](day-17/README.md) | `context` — cancel, timeout, deadline, values; TCP echo server | Cancellable HTTP downloader |
| [18](day-18/README.md) | `os`, `bufio`, `io` — file I/O, walking directories, `io.Copy` | Config file reader |
| [19](day-19/README.md) | `encoding/json`, struct tags, streaming encode/decode, `encoding/csv` | Config manager |
| [20](day-20/README.md) | `flag` package, `cobra` subcommands, stdin/stdout, exit codes | `todo` CLI tool |
| [21](day-21/README.md) | `net/http` — handler, middleware, `http.Client`, JSON API | HTTP echo server + client |

**Milestone:** You can write concurrent programs, build CLI tools, and serve HTTP.

---

## Week 4: Production (Days 22–30)

| Day | Topic | Project |
|-----|-------|---------|
| [22](day-22/README.md) | REST API with `chi`, JSON middleware, in-memory store | Notes REST API |
| [23](day-23/README.md) | `database/sql` + SQLite driver, prepared statements, transactions | Persist notes to database |
| [24](day-24/README.md) | Error hierarchy: sentinels, typed errors, HTTP error responses | Clean error layer for notes API |
| [25](day-25/README.md) | `log/slog` — structured logging, handlers, trace IDs via context | Structured logs across the notes app |
| [26](day-26/README.md) | `go test -bench`, `b.N`, `-benchmem`, `go tool pprof` | Profile word-frequency counter |
| [27](day-27/README.md) | `//go:embed`, `text/template`, `html/template`, build tags, `go:generate` | Embedded assets server with templates |
| [28](day-28/README.md) | `reflect` — `TypeOf`, `ValueOf`, struct tags, `reflect.Kind` | Custom `Describe` utility |
| [29](day-29/README.md) | Versioning, `pkg.go.dev`, Docker multi-stage build, `goreleaser` | Package and ship the shape library |
| [30](day-30/README.md) | Capstone | End-to-end CLI + REST API + database app |

**Milestone:** You can design, build, instrument, and deploy a production Go service.

---

## Bonus Deep-Dive Days (Days 31–34)

| Day | Topic | Project |
|-----|-------|---------|
| [31](day-31/README.md) | Interfaces: composition, nil trap, stdlib catalogue, accept-interfaces/return-structs, mocking, anti-patterns | Pluggable notification system |
| [32](day-32/README.md) | Idiomatic project structure: flat, `cmd/`, `internal/`, domain-driven layout, anti-patterns | Restructure the notes API |
| [33](day-33/README.md) | Channel model: pipelines, fan-out/in, done channels, `select`, direction types, common mistakes | Channel-based text pipeline + word counter |
| [34](day-34/README.md) | Goroutines & sync: `Mutex`, `RWMutex`, `WaitGroup`, `Once`, `atomic`, `sync.Map`, `errgroup`, deadlocks | Concurrent download manager |

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
