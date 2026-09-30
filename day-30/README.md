# Day 30: Capstone

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
