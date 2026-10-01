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

## Labs

### Lab 1: Sentinel Errors

**What you'll practise:** defining package-level sentinel errors and checking them with `errors.Is`.

**Task:**
Define `ErrNotFound` and `ErrInvalidInput` sentinel errors. Write functions that return them and demonstrate `errors.Is` checking at the call site.

**Steps:**
1. Declare `var ErrNotFound = errors.New("not found")` and `var ErrInvalidInput = errors.New("invalid input")`
2. Write `findUser(id int) (string, error)` that returns `ErrNotFound` when `id <= 0`
3. Write `validateAge(age int) error` that returns `ErrInvalidInput` when `age < 0 || age > 150`
4. In `main`, call both functions and use `errors.Is` to branch on the specific error

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
)

func findUser(id int) (string, error) {
    if id <= 0 {
        return "", ErrNotFound
    }
    return fmt.Sprintf("user-%d", id), nil
}

func validateAge(age int) error {
    if age < 0 || age > 150 {
        return ErrInvalidInput
    }
    return nil
}

// In main:
if errors.Is(err, ErrNotFound) {
    fmt.Println("user does not exist")
}
```

**Expected output:**
```
findUser(-1): not found
validateAge(200): invalid input
findUser(42): user-42
```

**Checkpoint:** `errors.Is(err, ErrNotFound)` returns true for errors returned by `findUser` with a non-positive id.

---

### Lab 2: Custom Error Type with `errors.As`

**What you'll practise:** implementing the `error` interface on a struct and extracting fields with `errors.As`.

**Task:**
Define `ParseError{Line int, Col int, Msg string}` implementing `error`. Write a parser function that returns it. Use `errors.As` to extract the line number at the call site.

**Steps:**
1. Define `type ParseError struct { Line, Col int; Msg string }`
2. Implement `func (e *ParseError) Error() string` returning `"line N col C: Msg"`
3. Write `parseToken(s string, line, col int) error` returning `&ParseError` for invalid input
4. At the call site, use `errors.As(err, &pe)` and print `pe.Line`

```go
type ParseError struct {
    Line int
    Col  int
    Msg  string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("line %d col %d: %s", e.Line, e.Col, e.Msg)
}

func parseToken(s string, line, col int) error {
    if s == "" {
        return &ParseError{Line: line, Col: col, Msg: "empty token"}
    }
    return nil
}

// At call site:
var pe *ParseError
if errors.As(err, &pe) {
    fmt.Printf("parse failed at line %d\n", pe.Line)
}
```

**Expected output:**
```
line 3 col 7: empty token
parse failed at line 3
```

**Checkpoint:** `errors.As` successfully extracts the `*ParseError` and its `Line` field is accessible.

---

### Lab 3: Error Wrapping with `%w`

**What you'll practise:** wrapping errors with `fmt.Errorf("%w", err)` and verifying `errors.Is` works through the chain.

**Task:**
Simulate a layered system: a low-level `readDB` error gets wrapped by `fetchRecord` and then by `handleRequest`. Show that `errors.Is` and `errors.Unwrap` work at every level.

**Steps:**
1. Define a sentinel `ErrConnection = errors.New("connection refused")`
2. Write `readDB() error` returning `ErrConnection`
3. Write `fetchRecord(id int) error` wrapping with `fmt.Errorf("fetchRecord %d: %w", id, err)`
4. Write `handleRequest(id int) error` wrapping again
5. Call `handleRequest` and verify `errors.Is(err, ErrConnection)` is `true` at the top level

```go
var ErrConnection = errors.New("connection refused")

func readDB() error { return ErrConnection }

func fetchRecord(id int) error {
    if err := readDB(); err != nil {
        return fmt.Errorf("fetchRecord %d: %w", id, err)
    }
    return nil
}

func handleRequest(id int) error {
    if err := fetchRecord(id); err != nil {
        return fmt.Errorf("handleRequest: %w", err)
    }
    return nil
}
```

**Expected output:**
```
err: handleRequest: fetchRecord 42: connection refused
errors.Is(ErrConnection): true
Unwrap once: fetchRecord 42: connection refused
```

**Checkpoint:** `errors.Is(err, ErrConnection)` is `true` even though the error has been wrapped twice.

---

### Lab 4: `errors.Join` — Collecting Multiple Errors

**What you'll practise:** accumulating multiple independent errors and combining them with `errors.Join` (Go 1.20+).

**Task:**
Write a `validateUser` function that checks name, email, and age independently. Collect all validation failures with `errors.Join` so callers see every problem at once.

**Steps:**
1. Write three validators: `validateName`, `validateEmail`, `validateAge` each returning an error or nil
2. In `validateUser`, call all three, collect non-nil errors into a slice
3. Use `errors.Join(errs...)` to return a combined error
4. Print the combined error and verify `errors.Is` works for individual sentinels through the join

```go
var (
    ErrBadName  = errors.New("name required")
    ErrBadEmail = errors.New("invalid email")
    ErrBadAge   = errors.New("age out of range")
)

func validateUser(name, email string, age int) error {
    var errs []error
    if name == "" {
        errs = append(errs, ErrBadName)
    }
    if !strings.Contains(email, "@") {
        errs = append(errs, ErrBadEmail)
    }
    if age < 0 || age > 150 {
        errs = append(errs, ErrBadAge)
    }
    return errors.Join(errs...)
}
```

**Expected output:**
```
name required
invalid email
age out of range
errors.Is(ErrBadName): true
```

**Checkpoint:** `errors.Join` returns nil when all validators pass. `errors.Is` finds individual sentinels inside the joined error.

---

### Lab 5: Panic and Recover — `safeDiv`

**What you'll practise:** converting a panic into a returned error using `recover` inside a deferred function.

**Task:**
Write `safeDiv(a, b int) (result int, err error)` that recovers from integer division-by-zero panics and returns an `ErrDivisionByZero` error instead.

**Steps:**
1. Define `var ErrDivisionByZero = errors.New("division by zero")`
2. In `safeDiv`, use a named return and a deferred function that calls `recover()`
3. If `recover()` returns a non-nil value, set `err = ErrDivisionByZero`
4. Perform `a / b` in the function body; call `safeDiv(10, 0)` and `safeDiv(10, 2)` in main

```go
var ErrDivisionByZero = errors.New("division by zero")

func safeDiv(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = ErrDivisionByZero
        }
    }()
    return a / b, nil
}
```

**Expected output:**
```
10 / 2 = 5, err=<nil>
10 / 0 = 0, err=division by zero
```

**Checkpoint:** `safeDiv(10, 0)` returns `0, ErrDivisionByZero` without crashing the program.

---

### Lab 6: Panic vs Error — When to Use Each

**What you'll practise:** the Go convention that panics signal programming errors, while returned errors signal expected failure conditions.

**Task:**
Build two contrasting examples: one where `panic` is appropriate (programmer error during init) and one where returning an error is correct (runtime failure). Add comments explaining the rule.

**Steps:**
1. Write `mustPositive(n int) int` that panics with a helpful message if `n <= 0` — this is a programmer-error guard for invariants
2. Write `openConfig(path string) (*Config, error)` that returns an error when the file is missing — this is a recoverable runtime condition
3. In `main`, use `mustPositive` with a valid value; show what happens with an invalid one using `safeDiv`-style recover in a test
4. Add a comment block explaining: panic in `init`/setup for invariants; return errors for expected failures

```go
// mustPositive is called during program setup.
// It panics because a non-positive value here is a programming bug,
// not a runtime condition the caller should handle.
func mustPositive(n int) int {
    if n <= 0 {
        panic(fmt.Sprintf("mustPositive: got %d, want > 0", n))
    }
    return n
}

// openConfig returns an error because a missing file is a
// recoverable condition — the caller can log it, use defaults, or retry.
func openConfig(path string) ([]byte, error) {
    return os.ReadFile(path)
}
```

**Expected output:**
```
Port: 8080
openConfig: open missing.toml: no such file or directory
```

**Checkpoint:** You can articulate in a comment: use `panic` only for invariant violations detected at startup; use returned errors for all runtime failures callers might handle.

---

### Final Lab (Project): CSV Row Parser

**What you'll practise:** combining sentinel errors, custom error types, error wrapping, and `errors.Join` into a realistic parsing component.

**Task:**
Parse rows from a CSV string. Define `ErrEmptyField` and `ParseError`, write a `parseRow` function that validates field count and content, and demonstrate `errors.Is` / `errors.As` at the call site.

**Steps:**
1. Define `var ErrEmptyField = errors.New("empty required field")`
2. Define `ParseError{Line, Col int, Msg string}` implementing `error`
3. Write `parseRow(line string, lineNum int) ([]string, error)` that wraps `ErrEmptyField` in a `ParseError` when a field is blank
4. Parse several rows — some valid, some invalid — and inspect errors with `errors.Is` and `errors.As`

```go
func parseRow(line string, lineNum int) ([]string, error) {
    fields := strings.Split(line, ",")
    if len(fields) != 3 {
        return nil, &ParseError{Line: lineNum, Col: 0,
            Msg: fmt.Sprintf("expected 3 fields, got %d", len(fields))}
    }
    for i, f := range fields {
        if strings.TrimSpace(f) == "" {
            return nil, &ParseError{
                Line: lineNum, Col: i + 1,
                Msg:  fmt.Sprintf("%w", ErrEmptyField),
            }
        }
    }
    return fields, nil
}
```

**Expected output:**
```
row 1: [alice 30 engineer]
row 2: parse error line 2 col 2: empty required field
  -> errors.Is(ErrEmptyField): true
  -> line: 2
```

**Checkpoint:** `errors.Is(err, ErrEmptyField)` is true for blank-field rows; `errors.As` extracts the line number; valid rows parse cleanly.

**Extension ideas:** stack multiple errors with [`errors.Join`](https://pkg.go.dev/errors#Join) (Go 1.20+); write a retry wrapper that retries on transient errors.

## Official Documentation

- [`errors`](https://pkg.go.dev/errors) — New, Is, As, Unwrap, Join
- [`fmt`](https://pkg.go.dev/fmt) — Errorf with `%w` wrapping verb
- [Language Spec: Errors](https://go.dev/ref/spec#Errors) — the built-in error interface
- [Effective Go: Errors](https://go.dev/doc/effective_go#errors) — error handling patterns
- [Go Blog: Error handling and Go](https://go.dev/blog/error-handling-and-go) — idiomatic error handling
- [Go Blog: Working with errors in Go 1.13](https://go.dev/blog/go1.13-errors) — wrapping, Is, As
