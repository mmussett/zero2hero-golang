# Day 14: Testing

## Core Concept: Testing Is Built In

Go's testing framework is part of the standard library — no third-party runner required.

```bash
go test ./...              # run all tests
go test -v ./...           # verbose output
go test -run TestFoo ./... # run tests matching "TestFoo"
go test -count=1 ./...     # disable test caching
go test -race ./...        # race detector
```

## Test File Layout

```
day-13/
├── calc.go
└── calc_test.go   // same package: package main (or package calc)
```

Tests live in `_test.go` files. The test binary is built separately from the main binary.

## Basic Test

```go
import "testing"

func TestAdd(t *testing.T) {
    got  := add(2, 3)
    want := 5
    if got != want {
        t.Errorf("add(2,3) = %d; want %d", got, want)
    }
}
```

## Table-Driven Tests (Idiomatic Go)

```go
func TestDivide(t *testing.T) {
    cases := []struct {
        name    string
        a, b    float64
        want    float64
        wantErr bool
    }{
        {"positive", 10, 2, 5, false},
        {"negative divisor", -10, 2, -5, false},
        {"divide by zero", 1, 0, 0, true},
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got, err := divide(tc.a, tc.b)
            if (err != nil) != tc.wantErr {
                t.Fatalf("unexpected error: %v", err)
            }
            if !tc.wantErr && got != tc.want {
                t.Errorf("got %v; want %v", got, tc.want)
            }
        })
    }
}
```

[`t.Run`](https://pkg.go.dev/testing#T.Run) creates subtests — run a single subtest with `-run TestDivide/divide_by_zero`.

## Benchmarks

```go
func BenchmarkWordCount(b *testing.B) {
    text := strings.Repeat("the quick brown fox ", 1000)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        wordCount(text)
    }
}
```

```bash
go test -bench=. -benchmem ./...
```

`b.N` is set by the framework to run long enough for a stable measurement. `-benchmem` shows allocations per operation.

## testify (External)

[`github.com/stretchr/testify`](https://pkg.go.dev/github.com/stretchr/testify) provides cleaner assertions:

```go
import "github.com/stretchr/testify/assert"

assert.Equal(t, want, got)
assert.NoError(t, err)
assert.ErrorIs(t, err, ErrNotFound)
```

## Test Helpers

```go
func TestMain(m *testing.M) {
    // global setup
    os.Exit(m.Run())
}
```

## Labs

### Lab 1: First Test — Table-Driven TestAdd

**What you'll practise:** Writing your first test function with multiple table cases.

**Task:**
Create `calc.go` with an `add(a, b int) int` function, then write a table-driven test in `calc_test.go` covering positive numbers, negative numbers, zero, and a large value close to `math.MaxInt32`.

**Steps:**
1. Create `calc.go` with the `add` function
2. Create `calc_test.go` with a `cases` slice of anonymous structs
3. Loop over `cases` using `t.Run` — but for this lab use a plain `if` check without subtests
4. Run `go test -v ./...` and observe the PASS/FAIL output

```go
// calc_test.go
func TestAdd(t *testing.T) {
    cases := []struct {
        name    string
        a, b    int
        want    int
    }{
        {"positive", 2, 3, 5},
        {"negative", -4, -6, -10},
        {"zero", 0, 0, 0},
        {"large", math.MaxInt32, 1, math.MaxInt32 + 1},
    }
    for _, tc := range cases {
        got := add(tc.a, tc.b)
        if got != tc.want {
            t.Errorf("%s: add(%d, %d) = %d; want %d", tc.name, tc.a, tc.b, got, tc.want)
        }
    }
}
```

**Expected output:**
```
--- PASS: TestAdd (0.00s)
ok  	day-13	0.001s
```

**Checkpoint:** `go test -v ./...` exits 0 with all cases passing.

---

### Lab 2: Subtests — Isolate Each Case with t.Run

**What you'll practise:** Using `t.Run` to create named subtests and filter them from the command line.

**Task:**
Refactor the `TestAdd` from Lab 1 so each case runs as a subtest via `t.Run(tc.name, ...)`. Then use the `-run` flag to execute only one subtest.

**Steps:**
1. Wrap the inner `if` check in `t.Run(tc.name, func(t *testing.T) { ... })`
2. Run all subtests: `go test -v -run TestAdd ./...`
3. Run only the negative case: `go test -v -run TestAdd/negative ./...`
4. Intentionally break one case and observe the isolated failure output

```go
for _, tc := range cases {
    tc := tc // capture range variable
    t.Run(tc.name, func(t *testing.T) {
        got := add(tc.a, tc.b)
        if got != tc.want {
            t.Errorf("add(%d, %d) = %d; want %d", tc.a, tc.b, got, tc.want)
        }
    })
}
```

**Expected output:**
```
--- PASS: TestAdd (0.00s)
    --- PASS: TestAdd/positive (0.00s)
    --- PASS: TestAdd/negative (0.00s)
    --- PASS: TestAdd/zero (0.00s)
    --- PASS: TestAdd/large (0.00s)
```

**Checkpoint:** `go test -v -run TestAdd/negative ./...` runs exactly one subtest.

---

### Lab 3: Test Helpers — assertEq with t.Helper

**What you'll practise:** Writing a reusable generic test helper that reports the correct line number.

**Task:**
Write a generic `assertEq[T comparable]` helper function. Use it in at least two different test functions. Observe how `t.Helper()` makes failures point to the call site, not the helper body.

**Steps:**
1. Write the helper in `calc_test.go` (or a `testutil_test.go` file)
2. Remove `t.Helper()` first — run a failing test and note the line reported
3. Add `t.Helper()` back — observe the line number now points to the caller
4. Use `assertEq` in both `TestAdd` and a new `TestMultiply` function

```go
func assertEq[T comparable](t *testing.T, got, want T) {
    t.Helper()
    if got != want {
        t.Errorf("got %v; want %v", got, want)
    }
}

func TestMultiply(t *testing.T) {
    assertEq(t, multiply(3, 4), 12)
    assertEq(t, multiply(-2, 5), -10)
    assertEq(t, multiply(0, 99), 0)
}
```

**Expected output:**
```
--- PASS: TestMultiply (0.00s)
```

**Checkpoint:** A deliberately wrong expected value shows the line in `TestMultiply`, not inside `assertEq`.

---

### Lab 4: Testify — Cleaner Assertions

**What you'll practise:** Using `github.com/stretchr/testify` assert and require packages.

**Task:**
Rewrite Lab 1's `TestAdd` using `assert.Equal`. Add a `divide(a, b float64) (float64, error)` function, then write a test that uses `assert.Error` and `require.NoError` to distinguish expected vs unexpected errors.

**Steps:**
1. `go get github.com/stretchr/testify` (already in this module's go.mod)
2. Import `"github.com/stretchr/testify/assert"` and `"github.com/stretchr/testify/require"`
3. Rewrite `TestAdd` using `assert.Equal(t, want, got)`
4. Write `TestDivide` with a zero-divisor case using `assert.Error` and a normal case using `require.NoError` then `assert.Equal`

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestDivide(t *testing.T) {
    t.Run("normal", func(t *testing.T) {
        got, err := divide(10, 2)
        require.NoError(t, err)
        assert.Equal(t, 5.0, got)
    })
    t.Run("by zero", func(t *testing.T) {
        _, err := divide(1, 0)
        assert.Error(t, err)
    })
}
```

**Expected output:**
```
--- PASS: TestDivide (0.00s)
    --- PASS: TestDivide/normal (0.00s)
    --- PASS: TestDivide/by_zero (0.00s)
```

**Checkpoint:** `go test -v ./...` passes. Note how testify's failure messages are more descriptive than manual `t.Errorf`.

---

### Lab 5: Benchmarks — BenchmarkWordFrequency

**What you'll practise:** Writing a benchmark with `b.ResetTimer` and reading `-benchmem` output.

**Task:**
Write a `wordFrequency(s string) map[string]int` function and benchmark it against a 1 000-word string. Use `b.ResetTimer` to exclude setup time. Run with `-benchmem` and note the allocations per operation.

**Steps:**
1. Implement `wordFrequency` using `strings.Fields`
2. Write `BenchmarkWordFrequency` that builds the test string once before `b.ResetTimer()`
3. Run: `go test -bench=BenchmarkWordFrequency -benchmem ./...`
4. Try an alternative implementation (e.g. using `bufio.Scanner`) and compare the ns/op figures

```go
func BenchmarkWordFrequency(b *testing.B) {
    text := strings.Repeat("the quick brown fox jumps over the lazy dog ", 100)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        wordFrequency(text)
    }
}
```

**Expected output:**
```
BenchmarkWordFrequency-8    50000    24000 ns/op    8192 B/op    1 allocs/op
```
(exact numbers vary by machine)

**Checkpoint:** The benchmark runs without error and reports `ns/op` and `B/op` columns.

---

### Lab 6: Test Coverage — Find the Gaps

**What you'll practise:** Generating a coverage profile and reading the HTML report.

**Task:**
Run coverage analysis on your day-13 package, generate an HTML report, and find at least one uncovered branch. Add a test to cover it.

**Steps:**
1. `go test -coverprofile=coverage.out ./...`
2. `go tool cover -func=coverage.out` to see per-function percentages
3. `go tool cover -html=coverage.out` to open the browser report (green = covered, red = not)
4. Add a test case for an uncovered branch and confirm coverage improves

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
# Look for lines showing < 100%
go tool cover -html=coverage.out
```

**Expected output:**
```
day-13/calc.go:5:    add           100.0%
day-13/calc.go:9:    divide        75.0%
total:               (statements)  87.5%
```

**Checkpoint:** Total coverage reaches 90 %+ after adding the missing test case.

---

### Lab 7: Fuzz Test — FuzzIsPalindrome

**What you'll practise:** Writing a fuzz target with a seed corpus and running the fuzzer.

**Task:**
Implement `isPalindrome(s string) bool`. Write `FuzzIsPalindrome` with at least three seed inputs. Run the fuzzer for 10 seconds and observe whether it finds a panic.

**Steps:**
1. Implement `isPalindrome` (naive: compare `s` to `strings.Reverse(s)`, or index-based)
2. Write `FuzzIsPalindrome` in a `_test.go` file
3. Run: `go test -fuzz=FuzzIsPalindrome -fuzztime=10s ./...`
4. If the fuzzer finds a failure, fix the bug and re-run

```go
func FuzzIsPalindrome(f *testing.F) {
    // Seed corpus
    f.Add("racecar")
    f.Add("hello")
    f.Add("")
    f.Add("A")

    f.Fuzz(func(t *testing.T, s string) {
        // Property: isPalindrome must not panic on any input
        _ = isPalindrome(s)
    })
}
```

**Expected output:**
```
fuzz: elapsed: 10s, gathering baseline coverage: 0/4 completed
fuzz: elapsed: 10s, execs: 38291 (3829/sec), new interesting: 8 (total: 12)
ok  	day-13	10.003s
```

**Checkpoint:** The fuzzer runs for the full 10 seconds without finding a crasher (or you fix any crash it finds).

---

### Final Lab (Project): Full Test Suite

**What you'll practise:** Combining all testing techniques — table-driven tests, subtests, helpers, testify, benchmarks, coverage, and fuzzing — into a production-quality test suite.

**Task:**
Write comprehensive tests for the functions built across Days 02–12.

**Steps:**
1. Table-driven tests for the CLI calculator (Day 02)
2. Table-driven tests for the word frequency counter (Day 06)
3. Subtests for the shape library (Day 08)
4. Tests for generic Stack and Queue (Day 10)
5. At least one benchmark with `-benchmem`
6. Achieve > 85 % coverage (`go test -cover ./...`)
7. Add a fuzz test for any string-processing function

```go
// Example: shape subtests
func TestShapeArea(t *testing.T) {
    shapes := []struct {
        name  string
        shape Shape
        want  float64
    }{
        {"circle r=1", Circle{1}, math.Pi},
        {"rect 3x4", Rect{3, 4}, 12},
    }
    for _, tc := range shapes {
        t.Run(tc.name, func(t *testing.T) {
            assert.InDelta(t, tc.want, tc.shape.Area(), 1e-9)
        })
    }
}
```

**Expected output:**
```
ok  	day-13	0.45s	coverage: 87.3% of statements
```

**Checkpoint:** `go test -v -race -cover ./...` passes with coverage above 85 % and no race conditions.

---

## Day Project: Full Test Suite

Write comprehensive tests for the functions built in Days 02–12:
1. Table-driven tests for the CLI calculator from Day 02
2. Table-driven tests for the word frequency counter from Day 06
3. Subtests for the shape library from Day 08
4. Tests for generic Stack and Queue from Day 10
5. At least one benchmark

**Extension ideas:** add a fuzzing test with `go test -fuzz`; measure coverage with `go test -cover`.

## Official Documentation

- [`testing`](https://pkg.go.dev/testing) — T, B, M types; Run, Errorf, Fatalf, ResetTimer, and more
- [`strings`](https://pkg.go.dev/strings) — `strings.Repeat` used in benchmarks
- [`os`](https://pkg.go.dev/os) — `os.Exit` in TestMain
- [Go Blog: Table driven tests](https://go.dev/blog/subtests) — subtests and table-driven patterns
- [Go Blog: The cover story](https://go.dev/blog/cover) — test coverage tooling
- [Go Blog: Fuzzing](https://go.dev/blog/fuzz-beta) — fuzzing in Go 1.18+
- [Effective Go: Testing](https://go.dev/doc/effective_go) — idiomatic test patterns
