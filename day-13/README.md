# Day 13: Testing

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

`t.Run` creates subtests — run a single subtest with `-run TestDivide/divide_by_zero`.

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

`github.com/stretchr/testify` provides cleaner assertions:

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

## Day Project: Full Test Suite

Write comprehensive tests for the functions built in Days 02–12:
1. Table-driven tests for the CLI calculator from Day 02
2. Table-driven tests for the word frequency counter from Day 06
3. Subtests for the shape library from Day 08
4. Tests for generic Stack and Queue from Day 10
5. At least one benchmark

**Extension ideas:** add a fuzzing test with `go test -fuzz`; measure coverage with `go test -cover`.
