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

## Four-Week Progression

| Week | Focus |
|------|-------|
| 1 | Foundations — toolchain, types, control flow, pointers, slices, maps, packages |
| 2 | Core Language — interfaces, errors, generics, closures, testing |
| 3 | Intermediate — goroutines, channels, context, file I/O, JSON, CLI, HTTP |
| 4 | Production — REST APIs, databases, observability, benchmarking, deployment |

## Weekly Milestones

**Week 1:** You understand Go's core model — value semantics, explicit pointers, and why the language is deliberately small.

**Week 2:** You can model any domain idiomatically and handle errors without panicking.

**Week 3:** You can write concurrent programs, build CLI tools, and serve HTTP.

**Week 4:** You can design, build, instrument, and deploy a production Go service.

## Reference Docs

| File | Contents |
|------|----------|
| `CURRICULUM.md` | Full day-by-day plan with topics and project goals |
| `PRIMITIVES.md` | Every Go primitive type — sizes, zero values, operations, gotchas |
| `DATA_STRUCTURES.md` | Standard-library collections with complexity tables and idioms |

## Recurring Principles

- Build first, read the spec second
- Errors are values — handle them explicitly at every call site
- Prefer the standard library before reaching for external packages
- Concurrency is not parallelism — understand the difference before using either
- Iterate on earlier projects rather than rewriting from scratch
