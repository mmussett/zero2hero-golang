# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Per-day development (most common)
```bash
cd day-XX && go run .          # run a day's program
cd day-XX && go test ./...     # test a day's packages
cd day-XX && go test -v ./...  # verbose test output
cd day-XX && go test -race ./... # run with race detector
cd day-XX && go test -bench=. -benchmem ./...  # benchmarks
```

### From the workspace root
```bash
go run ./day-XX                # run without cd
go test ./day-XX/...           # test without cd
go test ./...                  # test all days
go build ./...                 # build all days
go vet ./...                   # vet all packages
go fmt ./...                   # format all source files
```

### Dependency management
```bash
cd day-XX && go get some/pkg@v1.2.3   # add dependency to one day
cd day-XX && go mod tidy              # clean up go.mod / go.sum
go work sync                          # sync workspace after go.mod changes
```

## Structure

Each day is an independent Go module in `day-XX/`. All modules are joined by the `go.work` workspace file at the root, so `go` commands at the root see all days.

```
zero2hero-golang/
├── go.work                  ← workspace (lists all day-XX and capstone modules)
├── CURRICULUM.md            ← full 34-day plan
├── PRIMITIVES.md            ← Go type reference
├── DATA_STRUCTURES.md       ← collection reference with complexity tables
├── day-01/ … day-34/        ← independent modules (go.mod + main.go + README.md)
└── capstone/                ← capstone project skeletons
    ├── grep/
    ├── passgen/
    ├── todo-api/
    ├── url-shortener/
    ├── chat-server/
    └── kvdb/
```

### Day module types
- **Most days (01–21, 26–28, 31–34):** binary crate — `package main` in `main.go`, no external dependencies
- **Day 07:** also has `wordcount/counter.go` library package
- **Day 13:** depends on `github.com/stretchr/testify`
- **Day 20:** depends on `github.com/spf13/cobra`
- **Day 22, 24, 25:** depend on `github.com/go-chi/chi/v5`
- **Day 23:** depends on `github.com/go-chi/chi/v5` and `modernc.org/sqlite`

Dependencies belong only in the individual day's `go.mod` — never in a shared root `go.mod` (there is none).

## Curriculum Overview

| Days | Theme |
|------|-------|
| 01–07 | Foundations: toolchain, types, slices, maps, pointers, packages |
| 08–14 | Core Language: interfaces, errors, generics, closures, testing |
| 15–21 | Intermediate: goroutines, channels, context, file I/O, JSON, CLI, HTTP |
| 22–30 | Production: REST APIs, databases, observability, benchmarking, deployment |
| 31 | Deep dive: Go interfaces — composition, nil traps, stdlib catalogue, anti-patterns |
| 32 | Deep dive: Idiomatic Go project structure — flat, cmd/, internal/, domain-driven |
| 33 | Deep dive: Channel model — pipelines, fan-out/in, done channels, select |
| 34 | Deep dive: Goroutines, sync (Mutex, RWMutex, WaitGroup, Once, atomic, errgroup) |

## Key Conventions

- Each day's `main.go` starts with `panic("not implemented")` — the learner replaces it
- `README.md` in each day is the primary lesson document: concept explanation, code examples, project goal, extension ideas
- Tests live in `_test.go` files in the same directory (or `_test.go` with `package foo_test` for black-box tests)
- No cross-day imports — days are intentionally independent
