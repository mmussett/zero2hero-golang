# Day 09: Error Handling

## Core Concept: Errors Are Values

Go's `error` is a plain interface:

```go
type error interface {
    Error() string
}
```

Errors are returned as regular values — there is no exception mechanism. Handle every error at the call site.

## Creating Errors

```go
errors.New("something went wrong")                         // simple
fmt.Errorf("parse failed at line %d: %w", line, err)      // with context and wrapping
```

The `%w` verb **wraps** an error — [`errors.Is`](https://pkg.go.dev/errors#Is) and [`errors.As`](https://pkg.go.dev/errors#As) can unwrap the chain.

## Sentinel Errors

Package-level error variables that callers check by identity:

```go
var (
    ErrNotFound   = errors.New("not found")
    ErrPermission = errors.New("permission denied")
)

if errors.Is(err, ErrNotFound) { /* ... */ }
```

## Custom Error Types

Carry structured data in an error:

```go
type ParseError struct {
    Line   int
    Column int
    Msg    string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("line %d col %d: %s", e.Line, e.Column, e.Msg)
}

// Retrieve with errors.As:
var pe *ParseError
if errors.As(err, &pe) {
    fmt.Printf("error at line %d\n", pe.Line)
}
```

## Error Wrapping Chain

```go
raw := errors.New("connection refused")
layer1 := fmt.Errorf("dial failed: %w", raw)
layer2 := fmt.Errorf("startup: %w", layer1)

errors.Is(layer2, raw)     // true — unwraps the chain
errors.Unwrap(layer2)      // layer1
```

## The [`errors`](https://pkg.go.dev/errors) Package

| Function | Purpose |
|----------|---------|
| [`errors.New(text)`](https://pkg.go.dev/errors#New) | Create a simple error |
| [`errors.Is(err, target)`](https://pkg.go.dev/errors#Is) | Identity check through wrapping chain |
| [`errors.As(err, &target)`](https://pkg.go.dev/errors#As) | Type check through wrapping chain |
| [`errors.Unwrap(err)`](https://pkg.go.dev/errors#Unwrap) | One level of unwrapping |

## Day Project: CSV Row Parser

Parse rows from a CSV string. Define:
- `ErrEmptyField` sentinel for blank required fields
- `ParseError` type carrying row and column numbers
- A `parseRow(line string, lineNum int) ([]string, error)` that validates field count and content
- Show `errors.Is` and `errors.As` in use

**Extension ideas:** stack multiple errors with [`errors.Join`](https://pkg.go.dev/errors#Join) (Go 1.20+); write a retry wrapper that retries on transient errors.

## Official Documentation

- [`errors`](https://pkg.go.dev/errors) — New, Is, As, Unwrap, Join
- [`fmt`](https://pkg.go.dev/fmt) — Errorf with `%w` wrapping verb
- [Language Spec: Errors](https://go.dev/ref/spec#Errors) — the built-in error interface
- [Effective Go: Errors](https://go.dev/doc/effective_go#errors) — error handling patterns
- [Go Blog: Error handling and Go](https://go.dev/blog/error-handling-and-go) — idiomatic error handling
- [Go Blog: Working with errors in Go 1.13](https://go.dev/blog/go1.13-errors) — wrapping, Is, As
