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

## Day Project: Clean Error Layer

Refactor the Day 22/23 notes API:
1. Define `ErrNotFound`, `ErrValidation`, `ErrConflict` sentinels
2. Create `AppError` with `Code`, `Message`, and wrapped sentinel
3. Replace `http.Error(w, "...", 404)` calls with `handleErr(w, NotFound("note", id))`
4. Add input validation in `createNote` / `updateNote`
5. Write tests that assert correct HTTP status codes for each error type

Run with: `go run .`

**Extension ideas:** add request ID to error responses by reading from context; log internal errors with full stack via `runtime/debug.Stack()`.
