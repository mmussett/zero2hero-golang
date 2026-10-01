# Day 24: Error Handling at Scale

## Core Concept: Errors Should Carry Context

At scale, errors need to carry enough information for:
1. The calling code to decide what to do (sentinel / type check)
2. The human reading logs to understand what happened (message + context)
3. The HTTP handler to return the right status code

## Error Hierarchy Pattern

```go
// Sentinel errors — identity checks
var (
    ErrNotFound   = errors.New("not found")
    ErrConflict   = errors.New("conflict")
    ErrValidation = errors.New("validation failed")
)

// Typed error — carries data + wraps a sentinel
type AppError struct {
    Code    string // machine-readable
    Message string // human-readable
    Err     error  // wrapped sentinel
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func NotFound(resource, id string) *AppError {
    return &AppError{
        Code:    "NOT_FOUND",
        Message: fmt.Sprintf("%s %q not found", resource, id),
        Err:     ErrNotFound,
    }
}
```

## HTTP Error Mapping

```go
func httpStatus(err error) int {
    switch {
    case errors.Is(err, ErrNotFound):   return http.StatusNotFound
    case errors.Is(err, ErrConflict):   return http.StatusConflict
    case errors.Is(err, ErrValidation): return http.StatusBadRequest
    default:                            return http.StatusInternalServerError
    }
}

func handleErr(w http.ResponseWriter, err error) {
    var ae *AppError
    if errors.As(err, &ae) {
        writeJSON(w, httpStatus(err), map[string]string{
            "code":    ae.Code,
            "message": ae.Message,
        })
        return
    }
    http.Error(w, "internal server error", 500)
}
```

## errors.Join (Go 1.20+)

Aggregate multiple errors:

```go
func validateNote(n Note) error {
    var errs []error
    if n.Text == "" { errs = append(errs, errors.New("text is required")) }
    if len(n.Text) > 10000 { errs = append(errs, errors.New("text too long")) }
    return errors.Join(errs...)
}
```

## Labs

### Lab 1: Sentinel Errors

**What you'll practise:** Defining package-level sentinel errors and checking them with `errors.Is`.

**Task:**
Define three sentinel errors — `ErrNotFound`, `ErrUnauthorized`, `ErrValidation` — in a package. Write a function that returns each based on a string argument. Verify with `errors.Is`.

**Steps:**
1. Declare `var ErrNotFound = errors.New("not found")` etc. at package level
2. Write `lookup(kind string) error` that returns the appropriate sentinel (or `nil`)
3. In `main`, call `lookup` with each kind and assert with `errors.Is`
4. Confirm that `errors.Is(ErrNotFound, ErrUnauthorized)` is `false`

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrValidation   = errors.New("validation failed")
)

func lookup(kind string) error {
    switch kind {
    case "notfound":
        return ErrNotFound
    case "auth":
        return ErrUnauthorized
    case "validation":
        return ErrValidation
    default:
        return nil
    }
}

func main() {
    for _, kind := range []string{"notfound", "auth", "validation"} {
        err := lookup(kind)
        fmt.Printf("Is ErrNotFound: %v, Is ErrUnauthorized: %v, Is ErrValidation: %v\n",
            errors.Is(err, ErrNotFound),
            errors.Is(err, ErrUnauthorized),
            errors.Is(err, ErrValidation),
        )
    }
}
```

**Expected output:**
```
Is ErrNotFound: true, Is ErrUnauthorized: false, Is ErrValidation: false
Is ErrNotFound: false, Is ErrUnauthorized: true, Is ErrValidation: false
Is ErrNotFound: false, Is ErrUnauthorized: false, Is ErrValidation: true
```

**Checkpoint:** Each call to `errors.Is` returns true only for the correct sentinel.

---

### Lab 2: Custom Error Type

**What you'll practise:** Defining a struct error type with `Error()` and `Unwrap()`, and checking it with `errors.As`.

**Task:**
Define `AppError{Code string, Message string, Err error}`. Implement `Error()` and `Unwrap()`. Write a function that returns an `*AppError` wrapping `ErrNotFound`. Use `errors.As` to extract the `AppError` and inspect its `Code`.

**Steps:**
1. Define the struct and its two methods
2. Write `NotFoundError(resource, id string) *AppError` constructor
3. In `main`, call the constructor, then use `errors.As` to extract and print `ae.Code`
4. Verify that `errors.Is(err, ErrNotFound)` is still `true` (because `Unwrap` returns it)

```go
type AppError struct {
    Code    string
    Message string
    Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func NotFoundError(resource, id string) *AppError {
    return &AppError{
        Code:    "NOT_FOUND",
        Message: fmt.Sprintf("%s %q not found", resource, id),
        Err:     ErrNotFound,
    }
}

err := NotFoundError("note", "42")
var ae *AppError
if errors.As(err, &ae) {
    fmt.Println("code:", ae.Code)
    fmt.Println("message:", ae.Message)
}
fmt.Println("is ErrNotFound:", errors.Is(err, ErrNotFound))
```

**Expected output:**
```
code: NOT_FOUND
message: note "42" not found
is ErrNotFound: true
```

**Checkpoint:** `errors.As` succeeds and `errors.Is` works through the `Unwrap` chain.

---

### Lab 3: Error Wrapping Chain

**What you'll practise:** Wrapping errors three levels deep with `fmt.Errorf %w` and traversing the chain manually.

**Task:**
Simulate a three-level call stack: `db layer` returns a sentinel, `service layer` wraps it, `handler layer` wraps again. Unwrap the chain manually with a loop and print each level. Then use `errors.Is` at the outermost error.

**Steps:**
1. Define `dbErr = errors.New("row not found")`
2. Wrap: `svcErr = fmt.Errorf("service: %w", dbErr)`
3. Wrap again: `handlerErr = fmt.Errorf("handler: %w", svcErr)`
4. Loop using `errors.Unwrap` to traverse the chain, printing each error

```go
dbErr      := errors.New("row not found")
svcErr     := fmt.Errorf("service lookup: %w", dbErr)
handlerErr := fmt.Errorf("handle request: %w", svcErr)

fmt.Println("full chain:", handlerErr)
fmt.Println("errors.Is (dbErr):", errors.Is(handlerErr, dbErr))

// Manual unwrap
for err := error(handlerErr); err != nil; err = errors.Unwrap(err) {
    fmt.Println("unwrap:", err)
}
```

**Expected output:**
```
full chain: handle request: service lookup: row not found
errors.Is (dbErr): true
unwrap: handle request: service lookup: row not found
unwrap: service lookup: row not found
unwrap: row not found
```

**Checkpoint:** `errors.Is` finds the root sentinel through three wrapping layers.

---

### Lab 4: errors.Join

**What you'll practise:** Aggregating multiple validation errors with `errors.Join` (Go 1.20+) and verifying individual errors remain findable.

**Task:**
Write a `validateUser(name, email string) error` function that collects failures (empty name, missing `@` in email) and returns them joined. Call it with invalid input and print the result. Verify `errors.Is` still works on joined errors.

**Steps:**
1. Collect errors into a `[]error` slice
2. Return `errors.Join(errs...)` (returns `nil` if the slice is empty)
3. Print the error; observe all messages appear
4. Confirm `errors.Is(joinedErr, ErrValidation)` is `true` if you wrap sentinels

```go
var ErrValidation = errors.New("validation failed")

func validateUser(name, email string) error {
    var errs []error
    if name == "" {
        errs = append(errs, fmt.Errorf("%w: name is required", ErrValidation))
    }
    if !strings.Contains(email, "@") {
        errs = append(errs, fmt.Errorf("%w: email must contain @", ErrValidation))
    }
    return errors.Join(errs...)
}

err := validateUser("", "not-an-email")
fmt.Println(err)
fmt.Println("is ErrValidation:", errors.Is(err, ErrValidation))
fmt.Println("valid user:", validateUser("Alice", "alice@example.com"))
```

**Expected output:**
```
validation failed: name is required
validation failed: email must contain @
is ErrValidation: true
valid user: <nil>
```

**Checkpoint:** All error messages appear in the output; `errors.Is` works through the joined error.

---

### Lab 5: HTTP Error Mapping

**What you'll practise:** Writing an `HTTPStatus(err error) int` function that maps `AppError` codes to HTTP status codes using `errors.As`.

**Task:**
Write `HTTPStatus` using a switch on `AppError.Code`. Demonstrate it by calling it with errors produced by `NotFoundError`, `ConflictError`, and `ValidationError` constructors and printing the status code for each.

**Steps:**
1. Add `ConflictError` and `ValidationError` constructors alongside `NotFoundError`
2. Write `HTTPStatus(err error) int` that uses `errors.As` to get `*AppError`, then switches on `ae.Code`
3. Default to 500 for unknown errors
4. Print results for all three error types plus an untyped error

```go
func HTTPStatus(err error) int {
    var ae *AppError
    if !errors.As(err, &ae) {
        return http.StatusInternalServerError
    }
    switch ae.Code {
    case "NOT_FOUND":
        return http.StatusNotFound
    case "CONFLICT":
        return http.StatusConflict
    case "VALIDATION":
        return http.StatusBadRequest
    default:
        return http.StatusInternalServerError
    }
}

fmt.Println(HTTPStatus(NotFoundError("note", "1")))    // 404
fmt.Println(HTTPStatus(ConflictError("note", "dup")))  // 409
fmt.Println(HTTPStatus(ValidationError("text empty"))) // 400
fmt.Println(HTTPStatus(errors.New("unknown")))         // 500
```

**Expected output:**
```
404
409
400
500
```

**Checkpoint:** Each status code matches the expected HTTP constant; untyped errors always return 500.

---

### Lab 6: Error Middleware for chi

**What you'll practise:** Writing chi middleware that catches `*AppError` from handlers and writes a structured JSON error response.

**Task:**
Define a convention where handlers return errors by setting a value in the request context. Write `ErrorMiddleware` that reads that value after the handler returns and calls `handleErr` to write the correct JSON response.

**Steps:**
1. Define a context key type and `SetError(ctx, err) context.Context` / `GetError(ctx) error` helpers
2. Write `ErrorMiddleware(next http.Handler) http.Handler` that reads the error after `next.ServeHTTP`
3. In `handleErr`, use `HTTPStatus` and write `{"code":"...","message":"..."}` JSON
4. Wire it up on a chi router and test with a handler that sets a not-found error

```go
type ctxKey struct{}

func SetError(ctx context.Context, err error) context.Context {
    return context.WithValue(ctx, ctxKey{}, err)
}

func GetError(ctx context.Context) error {
    err, _ := ctx.Value(ctxKey{}).(error)
    return err
}

func ErrorMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        next.ServeHTTP(w, r.WithContext(ctx))
        if err := GetError(r.Context()); err != nil {
            handleErr(w, err)
        }
    })
}
```

**Expected output:**
```
$ curl -s http://localhost:8080/notes/999
{"code":"NOT_FOUND","message":"note \"999\" not found"}
```

**Checkpoint:** The handler does not write the response directly; the middleware writes it from the error context value.

---

### Final Lab: Clean Error Layer

**What you'll practise:** Wiring sentinels, `AppError`, HTTP status mapping, `errors.Join`, and error middleware into a unified error layer.

**Task:**
Refactor the Day 22/23 Notes API so that every error flows through the error hierarchy. No handler calls `http.Error` directly. All HTTP status codes are derived from `HTTPStatus`. Input validation uses `errors.Join`.

**Steps:**
1. Define `ErrNotFound`, `ErrConflict`, `ErrValidation` sentinels and the `AppError` type
2. Replace all `http.Error(w, "...", 404)` calls with `handleErr(w, NotFoundError("note", id))`
3. Add `validateNote(text string) error` using `errors.Join` for multi-field validation
4. Write `ErrorMiddleware` that catches `*AppError` from context and writes structured JSON
5. Write tests asserting exact HTTP status codes for each error type

```go
// Test example
func TestNotFoundReturns404(t *testing.T) {
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/notes/9999", nil)
    router.ServeHTTP(rr, req)
    if rr.Code != http.StatusNotFound {
        t.Fatalf("expected 404, got %d", rr.Code)
    }
    var body map[string]string
    json.NewDecoder(rr.Body).Decode(&body)
    if body["code"] != "NOT_FOUND" {
        t.Fatalf("expected code NOT_FOUND, got %q", body["code"])
    }
}
```

**Expected output:**
```
$ curl -s http://localhost:8080/api/notes/999
{"code":"NOT_FOUND","message":"note \"999\" not found"}
$ curl -s -X POST -d '{"text":""}' http://localhost:8080/api/notes
{"code":"VALIDATION","message":"validation failed: text is required"}
$ go test -v ./...
--- PASS: TestNotFoundReturns404 (0.00s)
--- PASS: TestValidationReturns400 (0.00s)
PASS
```

**Checkpoint:** No handler writes directly to `w` on error paths; all error responses have a `code` field; tests cover all three error types.

---

## Day Project: Clean Error Layer

Refactor the Day 22/23 notes API:
1. Define `ErrNotFound`, `ErrValidation`, `ErrConflict` sentinels
2. Create `AppError` with `Code`, `Message`, and wrapped sentinel
3. Replace `http.Error(w, "...", 404)` calls with `handleErr(w, NotFound("note", id))`
4. Add input validation in `createNote` / `updateNote`
5. Write tests that assert correct HTTP status codes for each error type

Run with: `go run .`

**Extension ideas:** add request ID to error responses by reading from context; log internal errors with full stack via `runtime/debug.Stack()`.

## Official Documentation

- [`errors`](https://pkg.go.dev/errors) — `New`, `Is`, `As`, `Join` (Go 1.20+), `Unwrap`
- [`net/http`](https://pkg.go.dev/net/http) — `StatusNotFound`, `StatusConflict`, `StatusBadRequest`, `StatusInternalServerError`, `Error`
- [`fmt`](https://pkg.go.dev/fmt) — `Errorf` with `%w` verb for error wrapping
- [`runtime/debug`](https://pkg.go.dev/runtime/debug) — `Stack` for capturing stack traces in error logs
- [Go Blog: Error handling and Go](https://go.dev/blog/error-handling-and-go)
- [Go Blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) — `errors.Is`, `errors.As`, `%w` wrapping
- [Language Spec — Errors](https://go.dev/ref/spec#Errors)
