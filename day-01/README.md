# Day 01: Toolchain, Modules, and Hello World

## Core Concept: Simplicity by Design

Go was designed at Google to solve real-world engineering problems: slow builds, unclear dependencies, and code that is hard to read at scale. The language has exactly one way to do most things — which makes it fast to learn and easy to maintain.

## The Toolchain (Three Commands You'll Use Every Day)

| Command | Role |
|---------|------|
| `go build` | Compile packages and dependencies |
| `go run .` | Compile and run in one step |
| `go test ./...` | Run all tests recursively |
| `go mod init` | Initialise a new module |
| `go get` | Add or upgrade a dependency |
| `go fmt ./...` | Format all source files |
| `go vet ./...` | Report suspicious code patterns |

## Module System

A Go **module** is a collection of packages with a shared version identifier. Every module has a `go.mod` file at its root:

```
module github.com/you/myproject

go 1.23
```

The module path is the canonical import path for all packages inside it. It does not have to be a real URL during development — but it must be unique and URL-shaped when published.

## Project Structure

```
day-01/
├── go.mod   ← module declaration
└── main.go  ← package main, func main()
```

Every Go program has exactly one `package main` with exactly one `func main()`. There is no `fn`, no `def`, no `def main():` — just `func main()`.

## Key Syntax

```go
package main

import "fmt"

func main() {
    name := "Gopher"
    day  := 1
    fmt.Printf("Day %d: Hello, %s!\n", day, name)
    fmt.Println("Welcome to Zero to Hero: Go")
}
```

- `:=` declares and assigns in one step (type is inferred)
- `fmt.Printf` uses C-style verbs: `%d` int, `%s` string, `%v` any value, `%T` type
- `fmt.Println` adds a newline automatically
- Unused imports are a **compile error** — the compiler enforces hygiene

## Day Project Goal

Create `day-01/main.go` that:
1. Declares variables for your name, the day number, and a motivational message using `:=`
2. Prints a greeting using `fmt.Printf` with at least two format verbs
3. Prints the Go version using `runtime.Version()`

Run with: `go run .`

## Extension Ideas

- Try `fmt.Sprintf` to build a string before printing it
- Print `os.Args` to see command-line arguments
- Explore what happens when you declare a variable and never use it
