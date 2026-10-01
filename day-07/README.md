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

## Labs

### Lab 1: Build the `wordcount` Package

**What you'll practise:** creating an exported Go package with a constructor, methods, and unexported helpers.

**Task:**
Create `day-07/wordcount/counter.go` implementing a `Counter` type that tracks word frequencies with `New()`, `Add(word string)`, `AddText(text string)`, `Total() int`, `Unique() int`, `Reset()`, and `TopN(n int) []string`.

**Steps:**
1. Create directory `wordcount/` inside `day-07/`
2. Declare `package wordcount` and define `Counter` with a private `counts map[string]int`
3. Implement `New() *Counter` returning an initialised counter
4. Implement `Add` normalising to lowercase, `AddText` splitting on whitespace
5. Implement `Total()`, `Unique()`, and `Reset()`
6. Implement `TopN(n int) []string` returning the n most frequent words

```go
// wordcount/counter.go
package wordcount

import (
    "sort"
    "strings"
)

type Counter struct {
    counts map[string]int
}

func New() *Counter { return &Counter{counts: make(map[string]int)} }

func (c *Counter) Add(word string) { c.counts[strings.ToLower(word)]++ }

func (c *Counter) AddText(text string) {
    for _, w := range strings.Fields(text) {
        c.Add(w)
    }
}

func (c *Counter) Total() int {
    n := 0
    for _, v := range c.counts {
        n += v
    }
    return n
}

func (c *Counter) Unique() int { return len(c.counts) }
func (c *Counter) Reset()      { c.counts = make(map[string]int) }

func (c *Counter) TopN(n int) []string {
    type kv struct{ k string; v int }
    var pairs []kv
    for k, v := range c.counts {
        pairs = append(pairs, kv{k, v})
    }
    sort.Slice(pairs, func(i, j int) bool { return pairs[i].v > pairs[j].v })
    out := make([]string, 0, n)
    for i, p := range pairs {
        if i >= n {
            break
        }
        out = append(out, p.k)
    }
    return out
}
```

**Expected output:**
```
Total: 6  Unique: 4  Top2: [go the]
```

**Checkpoint:** `go build ./wordcount/` compiles without errors and the package is importable from `main.go`.

---

### Lab 2: Table-Driven Tests

**What you'll practise:** writing idiomatic Go table-driven tests for a package using the black-box `_test` package convention.

**Task:**
Create `wordcount/counter_test.go` with table-driven tests covering `Add`, `AddText`, and `TopN`.

**Steps:**
1. Declare `package wordcount_test` (black-box style)
2. Write `TestAdd` with cases: empty counter, single word, duplicate words — assert `Total()`
3. Write `TestAddText` with a multi-word sentence — assert `Unique()`
4. Write `TestTopN` asserting the first returned word is the most frequent

```go
// wordcount/counter_test.go
package wordcount_test

import (
    "testing"
    "github.com/mmussett/zero2hero-golang/day-07/wordcount"
)

func TestAdd(t *testing.T) {
    cases := []struct {
        name  string
        words []string
        total int
    }{
        {"empty", nil, 0},
        {"one word", []string{"go"}, 1},
        {"duplicates", []string{"go", "go", "go"}, 3},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            c := wordcount.New()
            for _, w := range tc.words {
                c.Add(w)
            }
            if got := c.Total(); got != tc.total {
                t.Errorf("Total() = %d, want %d", got, tc.total)
            }
        })
    }
}
```

**Expected output:**
```
--- PASS: TestAdd/empty (0.00s)
--- PASS: TestAdd/one_word (0.00s)
--- PASS: TestAdd/duplicates (0.00s)
PASS
ok      .../wordcount
```

**Checkpoint:** `go test ./wordcount/...` passes with zero failures and zero skips.

---

### Lab 3: Export `WordEntry` and `Entries()`

**What you'll practise:** exporting a struct type and implementing a method that returns a deterministically sorted slice.

**Task:**
Add an exported `WordEntry` struct and an `Entries() []WordEntry` method that returns all words sorted by count descending, then alphabetically for ties.

**Steps:**
1. Define `type WordEntry struct { Word string; Count int }` in `counter.go`
2. Implement `Entries() []WordEntry` using `sort.Slice`
3. Sort primarily by `Count` descending, secondarily by `Word` ascending for ties
4. Add a `TestEntries` case that verifies alphabetical tiebreaking

```go
type WordEntry struct {
    Word  string
    Count int
}

func (c *Counter) Entries() []WordEntry {
    out := make([]WordEntry, 0, len(c.counts))
    for w, n := range c.counts {
        out = append(out, WordEntry{w, n})
    }
    sort.Slice(out, func(i, j int) bool {
        if out[i].Count != out[j].Count {
            return out[i].Count > out[j].Count
        }
        return out[i].Word < out[j].Word
    })
    return out
}
```

**Expected output:**
```
[{go 3} {is 2} {fun 1}]
```

**Checkpoint:** A test with tied counts confirms alphabetical tiebreaking. `go test ./wordcount/...` still passes.

---

### Lab 4: Command-Line Tool `cmd/wc`

**What you'll practise:** the `cmd/` directory convention and reading from `os.Stdin`.

**Task:**
Create `day-07/cmd/wc/main.go` that reads all of stdin, counts words using your `wordcount` package, and prints the top-10 entries formatted as `count  word`.

**Steps:**
1. Create directory `cmd/wc/` inside `day-07/`
2. Read all stdin with `io.ReadAll(os.Stdin)`
3. Create a `wordcount.New()` counter and call `AddText` with the data
4. Iterate `c.Entries()` printing at most 10 lines

```go
// cmd/wc/main.go
package main

import (
    "fmt"
    "io"
    "os"
    "github.com/mmussett/zero2hero-golang/day-07/wordcount"
)

func main() {
    data, _ := io.ReadAll(os.Stdin)
    c := wordcount.New()
    c.AddText(string(data))
    for i, e := range c.Entries() {
        if i >= 10 {
            break
        }
        fmt.Printf("%5d  %s\n", e.Count, e.Word)
    }
}
```

**Expected output:**
```
$ echo "the quick brown fox jumps over the lazy dog the" | go run ./cmd/wc
    3  the
    1  brown
    1  dog
    1  fox
    1  jumps
    1  lazy
    1  over
    1  quick
```

**Checkpoint:** `go run ./cmd/wc` reads from a pipe and outputs counts in descending order.

---

### Lab 5: Go Doc Comments

**What you'll practise:** writing godoc-style documentation comments and reading generated documentation with `go doc`.

**Task:**
Add doc comments to every exported symbol in `wordcount/counter.go`, then explore the generated documentation.

**Steps:**
1. Add a package-level comment above `package wordcount` following Go's sentence convention
2. Add a comment above each exported type, function, and method
3. Run `go doc ./wordcount` to see the package summary
4. Run `go doc ./wordcount Counter.Entries` to see a specific method
5. Run `go doc -all ./wordcount` to list every symbol with its comment

```go
// Package wordcount counts word frequencies in text.
// Create a Counter with New and use Add or AddText to record words.
package wordcount

// Counter tracks how many times each word appears.
// The zero value is not usable; create one with [New].
type Counter struct{ /* unexported fields */ }

// New returns an initialised, empty Counter.
func New() *Counter { /* ... */ }

// Add records word (normalised to lowercase) in the counter.
func (c *Counter) Add(word string) { /* ... */ }

// Entries returns all recorded words sorted by frequency descending,
// then alphabetically for words with equal frequency.
func (c *Counter) Entries() []WordEntry { /* ... */ }
```

**Expected output:**
```
$ go doc ./wordcount
package wordcount // import "..."

Package wordcount counts word frequencies in text.

func New() *Counter
type Counter struct{ ... }
    func (c *Counter) Add(word string)
    func (c *Counter) AddText(text string)
    func (c *Counter) Entries() []WordEntry
    ...
```

**Checkpoint:** `go doc -all ./wordcount` shows every exported symbol with a non-empty comment; zero symbols are undocumented.

---

### Final Lab (Project): Multi-Package Library

**What you'll practise:** combining package design, testing, documentation, and a CLI consumer into a polished multi-package module.

**Task:**
Refactor the word frequency counter from Day 06 into a proper library with a clean public API, full test coverage, and a usable CLI tool. All Labs 1–5 lead directly into this final project.

**Steps:**
1. Create `day-07/wordcount/` package with exported `Counter` type
2. Implement `New()`, `Add(word string)`, `TopN(n int) []Entry`, `Total() int`
3. Write tests in `wordcount/counter_test.go`
4. Use it from `main.go`

```go
// main.go
package main

import (
    "fmt"
    "github.com/mmussett/zero2hero-golang/day-07/wordcount"
)

func main() {
    c := wordcount.New()
    c.AddText("go is fun go go")
    for _, e := range c.Entries() {
        fmt.Printf("%d\t%s\n", e.Count, e.Word)
    }
}
```

**Expected output:**
```
3    go
1    fun
1    is
```

**Checkpoint:** `go test ./wordcount/...` passes; `go vet ./...` is clean; every exported symbol has a doc comment; `go run ./cmd/wc` processes stdin correctly.

**Extension ideas:** add [`io.Reader`](https://pkg.go.dev/io#Reader) support to `Counter` so it can parse any text stream; add `Reset()`.

## Official Documentation

- [`fmt`](https://pkg.go.dev/fmt) — formatted I/O used in the package
- [`strings`](https://pkg.go.dev/strings) — string manipulation within the wordcount package
- [`io`](https://pkg.go.dev/io) — `io.Reader` interface for streaming input
- [Go Modules reference](https://go.dev/doc/modules/gomod-ref) — `go.mod` syntax and directives
- [Language Spec: Packages](https://go.dev/ref/spec#Packages) — package declarations and imports
- [Language Spec: Exported identifiers](https://go.dev/ref/spec#Exported_identifiers) — uppercase = exported
- [Effective Go: Package names](https://go.dev/doc/effective_go#package-names) — naming conventions
- [Go Blog: Organizing a Go module](https://go.dev/blog/organizing-go-code) — package structure guidance
