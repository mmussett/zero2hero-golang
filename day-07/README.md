# Day 07: Packages and Modules

## Core Concept: Simplicity Through Convention

Go's package system has three rules:
1. The directory name is the package name (by convention)
2. Identifiers starting with an uppercase letter are **exported** (public)
3. All files in a directory share one package

## Package Naming

```
day-07/
├── go.mod
├── main.go          // package main
└── wordcount/
    ├── counter.go   // package wordcount
    └── counter_test.go
```

```go
// wordcount/counter.go
package wordcount

// Counter is exported (uppercase)
type Counter struct {
    counts map[string]int
}

// New is exported — the constructor convention in Go
func New() *Counter {
    return &Counter{counts: make(map[string]int)}
}

// Add is exported
func (c *Counter) Add(word string) { c.counts[word]++ }

// top is unexported — internal helper
func (c *Counter) top(n int) []string { /* ... */ }
```

```go
// main.go
package main

import "github.com/mmussett/zero2hero-golang/day-07/wordcount"

func main() {
    c := wordcount.New()
    c.Add("go")
    c.Add("go")
}
```

## Imports

```go
import (
    "fmt"                         // standard library
    "strings"                     // standard library
    "github.com/user/pkg"        // external module
    mypkg "github.com/user/other" // alias
    _ "github.com/user/driver"   // blank import — side effects only
)
```

Unused imports are a **compile error**. Organise imports into three groups separated by blank lines: standard library, external, internal — `goimports` does this automatically.

## `internal` Packages

A package at `pkg/internal/foo` can only be imported by code rooted at `pkg/`. It is Go's enforced encapsulation boundary.

## Managing Dependencies

```bash
go get github.com/some/pkg@v1.2.3    # add/upgrade
go mod tidy                           # remove unused, add missing
go mod download                       # pre-cache dependencies
cat go.sum                            # cryptographic checksums
```

`go.sum` must be committed — it ensures reproducible builds.

## Day Project: Multi-Package Library

Refactor the word frequency counter from Day 06 into a proper library:

1. Create `day-07/wordcount/` package with exported `Counter` type
2. Implement `New()`, `Add(word string)`, `TopN(n int) []Entry`, `Total() int`
3. Write tests in `wordcount/counter_test.go`
4. Use it from `main.go`

Run tests: `go test ./wordcount/...`

**Extension ideas:** add `io.Reader` support to `Counter` so it can parse any text stream; add `Reset()`.
