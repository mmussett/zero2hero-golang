# Zero to Hero: Go

A self-paced, project-based 34-day curriculum — from absolute beginner to production-minded Go developer.

**[📖 Read the course at mmussett.github.io/zero2hero-golang](https://mmussett.github.io/zero2hero-golang/)**

Each day pairs focused concepts with a hands-on project. Earlier projects are revisited and improved rather than discarded.

## Prerequisites

- One prior programming language (any)
- A terminal
- Go 1.23+ installed: https://go.dev/dl/

```bash
go version  # should print go1.23.x or later
```

## Getting Started

```bash
git clone https://github.com/mmussett/zero2hero-golang
cd zero2hero-golang
go work sync          # initialise the workspace
cd day-01 && go run . # run Day 1
```

## Structure

The repo uses a [Go workspace](https://go.dev/doc/tutorial/workspaces). Each day is an independent module under `day-XX/` — runnable and testable in isolation.

```bash
# From the workspace root:
go run ./day-XX         # run a single day
go test ./day-XX/...    # test a single day
go test ./...           # test every day

# Or enter the day directly:
cd day-05
go run .
go test ./...
```

## Five-Week Progression

| Week | Days | Focus |
|------|------|-------|
| 1 | 1–7 | Foundations — toolchain, types, control flow, pointers, slices, maps, packages |
| 2 | 8–14 | Core Language — project structure, interfaces, errors, generics, closures, testing |
| 3 | 15–21 | Intermediate — goroutines, channels, context, file I/O, JSON, CLI |
| 4 | 22–28 | Production — HTTP, REST APIs, databases, logging, benchmarking, embedding |
| 5 | 29–34 | Advanced — reflection, deployment, interfaces, channels, goroutines, context |
| — | 35 | Capstone — end-to-end project of your own choosing |

## Weekly Milestones

**Week 1:** You understand Go's core model — value semantics, explicit pointers, and why the language is deliberately small.

**Week 2:** You can structure a Go project idiomatically and model any domain with interfaces, errors, and generics.

**Week 3:** You can write concurrent programs, handle I/O, and build CLI tools.

**Week 4:** You can serve HTTP, back it with a database, instrument it with structured logs, and benchmark it.

**Week 5:** You have a deep working knowledge of Go's most nuanced subsystems — interfaces, channels, goroutines, and context.

**Capstone:** You have shipped a complete, production-quality Go application.

## Reference Docs

| File | Contents |
|------|----------|
| `CURRICULUM.md` | Full day-by-day plan with topics and project goals |
| `PRIMITIVES.md` | Every Go primitive type — sizes, zero values, operations, gotchas |
| `DATA_STRUCTURES.md` | Standard-library collections with complexity tables and idioms |
| `GOENV.md` | Every Go environment variable — defaults, usage, real-world recipes |
| `GO_VSC.md` | Step-by-step Go + Visual Studio Code setup guide for Windows, macOS, Linux |
| `FMTVERBS.md` | Every `fmt` format verb, flag, and width/precision modifier with examples |

## Recurring Principles

- Build first, read the spec second
- Errors are values — handle them explicitly at every call site
- Prefer the standard library before reaching for external packages
- Concurrency is not parallelism — understand the difference before using either
- Iterate on earlier projects rather than rewriting from scratch
