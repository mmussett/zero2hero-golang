# Day 35: The Go Context Model — Deep Dive

Day 17 introduced `context.Context` as a practical tool. Today you go deeper: how the context tree works internally, advanced propagation patterns, values done right, context in HTTP middleware and databases, detecting leaks, and testing with context.

---

## 1. What Context Actually Is

`context.Context` is an interface with four methods:

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

- **`Done()`** — returns a channel that is closed when the context is cancelled or times out. Callers block on `<-ctx.Done()` to detect cancellation.
- **`Err()`** — returns the reason: `context.Canceled` or `context.DeadlineExceeded`. Returns `nil` if the context is not yet done.
- **`Deadline()`** — returns the absolute time at which the context will be cancelled, if one was set.
- **`Value(key)`** — retrieves a value stored by a parent. Returns `nil` if not found.

`context.Background()` and `context.TODO()` return non-cancellable root contexts that implement this interface. Every other context is built by wrapping one of them.

---

## 2. The Context Tree

Contexts form a **tree**. Each `With*` call creates a child node. Cancelling a parent automatically cancels all its descendants — but a child cannot cancel its parent.

```
Background()
    └── WithCancel() → cancelCtx
            ├── WithTimeout(5s) → timerCtx
            │       └── WithValue("reqID", "abc") → valueCtx
            └── WithValue("user", user) → valueCtx
```

When you cancel `cancelCtx`, every node below it receives the signal on its `Done()` channel simultaneously. This is the foundation of Go's cancellation model — one signal propagates down the entire call graph.

```go
parent, cancel := context.WithCancel(context.Background())
defer cancel()  // always call cancel to release resources

child, _ := context.WithTimeout(parent, 5*time.Second)
// Cancelling parent also cancels child.
// child's timeout is 5s OR until parent is cancelled, whichever comes first.
```

**Key rule:** always `defer cancel()` immediately after calling `WithCancel`, `WithTimeout`, or `WithDeadline`. Failing to call cancel leaks a goroutine inside the context runtime that watches for the parent to finish.

---

## 3. The Four Factory Functions

### `context.WithCancel`

```go
ctx, cancel := context.WithCancel(parent)
defer cancel()
```

Use when you need to cancel work explicitly — e.g., the user presses Ctrl-C, a request handler returns, or a service shuts down.

### `context.WithTimeout`

```go
ctx, cancel := context.WithTimeout(parent, 3*time.Second)
defer cancel()
```

Cancels after a duration relative to `time.Now()`. The most common choice for HTTP requests and database queries.

### `context.WithDeadline`

```go
deadline := time.Now().Add(3 * time.Second)
ctx, cancel := context.WithDeadline(parent, deadline)
defer cancel()
```

Cancels at an absolute time. Use when you receive a deadline from an external system (e.g., a gRPC deadline propagated from a caller) and want to honour it precisely.

### `context.WithValue`

```go
type ctxKey string  // unexported type — avoids collisions
const requestIDKey ctxKey = "requestID"

ctx = context.WithValue(ctx, requestIDKey, "req-abc-123")

// Later:
if id, ok := ctx.Value(requestIDKey).(string); ok {
    fmt.Println("request ID:", id)
}
```

Attaches a value to the context. The value is available to all functions in the call chain that receive this context.

---

## 4. Context Values — Done Right

Context values are often misused. Here are the rules:

### Use a private key type

Never use a built-in type (`string`, `int`) as a context key — any package could accidentally use the same key and overwrite or shadow your value.

```go
// WRONG — any package can set/shadow this:
ctx = context.WithValue(ctx, "userID", 42)

// RIGHT — only this package can use this key:
type contextKey int
const userIDKey contextKey = iota
ctx = context.WithValue(ctx, userIDKey, 42)
```

### What belongs in context vs function parameters

| Put in context | Pass as parameter |
|----------------|------------------|
| Request-scoped metadata: request ID, trace ID, auth token, user identity | Business logic inputs: IDs, names, flags |
| Cross-cutting concerns that every layer needs but no layer owns | Config values, feature flags |
| Cancellation signal | Return values, errors |

The test: if you would have to add the value as a parameter to every function in a call chain just to pass it through, context is appropriate. If the value is meaningful to the function's logic, make it a parameter.

### Helper functions for type-safe access

```go
type contextKey int

const (
    requestIDKey contextKey = iota
    userKey
)

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(requestIDKey).(string)
    return id, ok
}
```

This pattern makes the context API type-safe and keeps key management in one place.

---

## 5. Checking for Cancellation

### Polling

```go
func doWork(ctx context.Context) error {
    for _, item := range items {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        process(item)
    }
    return nil
}
```

The `default` case makes the `select` non-blocking — it only checks for cancellation, then continues immediately.

### Blocking on two things simultaneously

```go
select {
case result := <-resultCh:
    return result, nil
case <-ctx.Done():
    return nil, ctx.Err()
}
```

This is the canonical pattern for a goroutine that waits for either a result or cancellation.

### `context.Cause` (Go 1.21+)

```go
ctx, cancel := context.WithCancelCause(parent)
cancel(errors.New("user logged out"))  // attach a cause

// Later:
fmt.Println(context.Cause(ctx))  // prints: user logged out
fmt.Println(ctx.Err())           // prints: context canceled
```

`WithCancelCause` lets you attach a specific reason to the cancellation, separate from the generic `context.Canceled` sentinel.

---

## 6. Context in HTTP Middleware

Every `http.Request` carries a context. Middleware can enrich it:

```go
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := generateID()
        ctx := WithRequestID(r.Context(), id)
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        user, err := validateToken(token)
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), userKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

The handler at the bottom of the chain can retrieve both values:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    id, _ := RequestIDFromContext(ctx)
    user := ctx.Value(userKey).(*User)
    // ...
}
```

The request's context is automatically cancelled when the client disconnects. A long-running handler should check `ctx.Done()` so it can stop work and free resources immediately.

---

## 7. Context in Database Operations

All `database/sql` methods have context variants. Always use them:

```go
// Without context — cannot be cancelled:
rows, err := db.Query("SELECT * FROM users WHERE id = ?", id)

// With context — respects cancellation and timeout:
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", id)
```

If the HTTP request is cancelled (client disconnects), the context is cancelled, and `QueryContext` returns immediately rather than waiting for the database to respond. This prevents goroutine pileup under load.

```go
func GetUser(ctx context.Context, db *sql.DB, id int) (*User, error) {
    row := db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE id = ?", id)
    var u User
    if err := row.Scan(&u.ID, &u.Name, &u.Email); err != nil {
        return nil, err
    }
    return &u, nil
}
```

---

## 8. Propagating Context Across Goroutines

When you launch a goroutine to do work on behalf of a request, pass the context:

```go
func handleRequest(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)  // creates a child context

    g.Go(func() error {
        return fetchUserData(ctx)   // ← same context
    })
    g.Go(func() error {
        return fetchProductData(ctx)  // ← same context
    })

    return g.Wait()  // if either returns an error, ctx is cancelled
}
```

`errgroup.WithContext` (from `golang.org/x/sync/errgroup`) creates a group and a derived context. If any goroutine returns a non-nil error, the context is cancelled and the other goroutines can stop early.

**Never capture a context in a struct or global variable for use in goroutines launched later.** The context must flow through the call chain, not be stored and retrieved later:

```go
// WRONG — the stored context may be cancelled by the time the goroutine uses it:
type Worker struct {
    ctx context.Context
}

// RIGHT — pass context as a parameter to the method that launches the goroutine:
func (w *Worker) Start(ctx context.Context) {
    go w.run(ctx)
}
```

---

## 9. Context Leaks — How They Happen and How to Detect Them

A **context leak** is when a goroutine that is waiting on `ctx.Done()` is never unblocked because the context is never cancelled.

### Common cause: forgetting to call `cancel()`

```go
// LEAK — cancel is never called if makeRequest errors before timeout is needed:
ctx, _ := context.WithTimeout(parent, 5*time.Second)
resp, err := makeRequest(ctx)  // if this panics, cancel is never called
```

```go
// FIX — always defer cancel:
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()
resp, err := makeRequest(ctx)
```

### Common cause: context stored and used after its scope ends

```go
// LEAK — requestCtx is cancelled when the request handler returns,
// but the goroutine may still be running:
func handler(w http.ResponseWriter, r *http.Request) {
    go func() {
        doSlowWork(r.Context())  // r.Context() is cancelled when handler returns
    }()
}

// FIX — use a separate context for background work:
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := context.WithoutCancel(r.Context())  // Go 1.21+: detach from request lifetime
    go func() {
        doSlowWork(ctx)
    }()
}
```

### Detection

Use `runtime.NumGoroutine()` in tests before and after a request to check for goroutine leaks. The `goleak` package (`go.uber.org/goleak`) automates this:

```go
func TestHandler(t *testing.T) {
    defer goleak.VerifyNone(t)
    // run your test
}
```

---

## 10. Testing with Context

### Use `context.Background()` for unit tests

```go
func TestGetUser(t *testing.T) {
    user, err := GetUser(context.Background(), db, 1)
    // ...
}
```

### Use `context.WithTimeout` in integration tests to prevent hangs

```go
func TestIntegration(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    result, err := callRealService(ctx)
    // ...
}
```

### Test cancellation behaviour explicitly

```go
func TestGetUser_Cancelled(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel()  // cancel before the call

    _, err := GetUser(ctx, db, 1)
    if !errors.Is(err, context.Canceled) {
        t.Errorf("got %v, want context.Canceled", err)
    }
}
```

---

## Labs

### Lab 1: Visualise the Context Tree

**What you'll practise:** understanding parent-child cancellation propagation

**Task:**
Create a three-level context tree. Cancel the root and observe that all children receive the signal simultaneously.

**Steps:**
1. Create a root context with `context.WithCancel(context.Background())`
2. Create two children from the root: one with `WithTimeout(2s)`, one with `WithCancel`
3. Create a grandchild from the first child with `WithValue`
4. Launch a goroutine for each context that prints its level when `ctx.Done()` fires
5. Cancel the root after 500ms — observe all goroutines stop

```go
root, rootCancel := context.WithCancel(context.Background())
child1, _ := context.WithTimeout(root, 2*time.Second)
child2, _ := context.WithCancel(root)
grandchild := context.WithValue(child1, myKey, "hello")
```

**Expected output:**
```
grandchild done: context canceled
child1 done: context canceled
child2 done: context canceled
root done: context canceled
```
(all nearly simultaneously after the root cancel)

**Checkpoint:** Verify that cancelling `child2` alone does NOT cancel `child1` or the root.

---

### Lab 2: WithCancel — Stopping a Worker Pool

**What you'll practise:** using cancellation to stop goroutines cleanly

**Task:**
Start 5 worker goroutines. Each reads jobs from a channel. After 3 seconds, cancel the context and verify all workers exit cleanly.

**Steps:**
1. Create a `context.WithCancel` context
2. Create a `jobs` channel
3. Launch 5 goroutines that `select` between `jobs` and `ctx.Done()`
4. Send 20 jobs on a separate goroutine
5. After 3 seconds, call `cancel()`, then use a `WaitGroup` to confirm all workers exited

**Checkpoint:** `runtime.NumGoroutine()` should return to the baseline after cancel + Wait.

---

### Lab 3: WithTimeout vs WithDeadline

**What you'll practise:** choosing the right time-based context

**Task:**
Write two HTTP fetchers: one using `WithTimeout`, one using `WithDeadline`. Point both at a slow server (use `httptest` with a handler that sleeps 5 seconds). Show both return `DeadlineExceeded`.

**Steps:**
1. Start an `httptest.Server` with a handler that sleeps 5 seconds
2. `WithTimeout(1*time.Second)` — make a GET request, catch the error
3. `WithDeadline(time.Now().Add(1*time.Second))` — same
4. Print `ctx.Err()` for each: both show `context.DeadlineExceeded`
5. Then show a case where `WithDeadline` is better: you receive a deadline from a caller and want to use *at most* that much time

**Checkpoint:** Run the tests — both requests must fail within ~1 second, not 5.

---

### Lab 4: Type-Safe Context Values

**What you'll practise:** the typed-key pattern for context values

**Task:**
Build a `requestcontext` package (sub-directory) that stores and retrieves: a request ID (string), a user struct, and a logger. Demonstrate that the typed keys prevent collisions from other packages.

**Steps:**
1. Create `day-35/reqctx/reqctx.go` with unexported `contextKey` type
2. Expose `WithRequestID`, `RequestID`, `WithUser`, `User`, `WithLogger`, `Logger` functions
3. In `main.go`, build a chain: start with `Background()`, attach all three values, retrieve them deep in a call chain

```go
type contextKey int
const (
    keyRequestID contextKey = iota
    keyUser
    keyLogger
)
```

**Checkpoint:** `go vet ./...` passes. Show that a package using `context.WithValue(ctx, "requestID", ...)` does NOT interfere with your typed key.

---

### Lab 5: HTTP Middleware Chain with Context

**What you'll practise:** enriching request context across middleware layers

**Task:**
Build a middleware stack for an HTTP server:
1. `RequestIDMiddleware` — generates a UUID-like ID, stores in context, sets `X-Request-ID` header
2. `LoggingMiddleware` — reads request ID from context, logs method + path + duration
3. `AuthMiddleware` — reads `Authorization: Bearer <token>` header, validates (any non-empty token is "valid"), stores user in context
4. A `/protected` handler that reads both request ID and user from context and returns them as JSON

**Checkpoint:** `curl -H "Authorization: Bearer mytoken" localhost:8080/protected` returns `{"request_id":"...","user":"mytoken"}`.

---

### Lab 6: Context in Database Queries

**What you'll practise:** using context-aware database methods

**Task:**
Using `database/sql` with an in-memory SQLite database, write a `UserStore` with:
- `CreateUser(ctx, name string) (*User, error)` — uses `ExecContext`
- `GetUser(ctx, id int) (*User, error)` — uses `QueryRowContext`
- `ListUsers(ctx) ([]*User, error)` — uses `QueryContext`

Then write a test that cancels the context *before* calling each method and verifies the error is `context.Canceled`.

**Steps:**
1. Open an `sqlite3` in-memory DB (`:memory:`)
2. Create the users table in `init`
3. Implement the three methods with context params
4. In a test, cancel before each call and assert the error

**Checkpoint:** `go test ./...` passes. All cancelled calls return `context.Canceled`.

---

### Lab 7: Propagation Across Goroutines with errgroup

**What you'll practise:** structured concurrency with context propagation

**Task:**
Simulate fetching data from three "services" concurrently. Each service takes a random time (0–3s). If any takes more than 2 seconds total, cancel all and return an error.

**Steps:**
1. Create `context.WithTimeout(ctx, 2*time.Second)`
2. Use `errgroup.WithContext` (implement manually with WaitGroup + error channel if you don't want to add x/sync)
3. Launch 3 goroutines, each sleeping a random duration, checking `ctx.Done()` between iterations
4. If the timeout fires, all three goroutines must exit and the error is returned to main

**Checkpoint:** Running multiple times shows: sometimes all complete, sometimes the timeout fires. In both cases, no goroutines are left running after the function returns.

---

### Lab 8: Detecting Context Leaks

**What you'll practise:** identifying and fixing goroutine leaks caused by un-cancelled contexts

**Task:**
Write a deliberately leaky function, detect the leak, then fix it.

**Steps:**
1. Write `leakyFetch(url string)` that creates a `context.WithCancel` but never calls cancel, makes an HTTP request, returns
2. Call it 10 times in a test
3. Use `runtime.NumGoroutine()` before and after — observe the count grows
4. Fix it by adding `defer cancel()` immediately after `WithCancel`
5. Re-run: goroutine count returns to baseline

```go
before := runtime.NumGoroutine()
for i := 0; i < 10; i++ {
    leakyFetch("http://example.com")
}
after := runtime.NumGoroutine()
t.Logf("goroutines: before=%d after=%d", before, after)
```

**Checkpoint:** After the fix, `after - before <= 1` (within normal fluctuation).

---

### Lab 9: `context.WithoutCancel` and Background Work

**What you'll practise:** detaching background work from request lifetime (Go 1.21+)

**Task:**
A handler receives a request, starts a background audit job that must complete even after the request returns, then responds immediately.

**Steps:**
1. Build an HTTP handler that:
   - Receives a POST with a body
   - Starts a goroutine to "audit" the request (sleeps 2s, logs the body)
   - Responds `202 Accepted` immediately
2. First version: use `r.Context()` in the goroutine — show the goroutine is cancelled when the handler returns
3. Fixed version: use `context.WithoutCancel(r.Context())` — the goroutine runs to completion

**Checkpoint:** The audit log line appears 2 seconds after the handler responds, not before.

---

### Final Lab: Request-Scoped Pipeline

**What you'll practise:** everything from today applied to a realistic HTTP service

**Task:**
Build a small HTTP service that processes "reports" through a multi-stage pipeline. Each stage uses context for cancellation and carries request metadata via context values.

Endpoints:
- `POST /reports` — accepts `{"data": "..."}`, runs it through a 3-stage pipeline (validate → enrich → store), returns the result or an error
- `GET /health` — returns 200 immediately

Pipeline:
- Each stage receives and passes the context
- If the client disconnects mid-pipeline, all stages stop within 100ms
- The request ID flows through all stages via context
- If processing takes more than 5 seconds, timeout and return 504

```go
type Stage func(ctx context.Context, input string) (string, error)

func runPipeline(ctx context.Context, stages []Stage, input string) (string, error) {
    result := input
    for _, stage := range stages {
        var err error
        result, err = stage(ctx, result)
        if err != nil {
            return "", err
        }
    }
    return result, nil
}
```

**Checkpoint:** Use `curl -X POST localhost:8080/reports -d '{"data":"hello"}'` — verify the request ID appears in every log line. Kill the curl mid-request — verify the server logs `context canceled` and stops processing.

---

## Day Project Goal

Build the request-scoped pipeline from the Final Lab above. Your implementation must:

1. Use context propagation through all HTTP middleware and pipeline stages
2. Demonstrate all four `With*` functions being used appropriately
3. Store request ID and user identity via typed context keys
4. Use `WithTimeout` to enforce a 5-second processing deadline
5. Check `ctx.Done()` between pipeline stages so cancellation is fast
6. Log the context's request ID in every log line using `log/slog`
7. Include a test that cancels the context mid-pipeline and verifies the error

Run with: `go run .`

---

## Official Documentation

- [`context`](https://pkg.go.dev/context) — the full package: Background, TODO, WithCancel, WithTimeout, WithDeadline, WithValue, WithCancelCause, WithoutCancel
- [`context.WithCancelCause`](https://pkg.go.dev/context#WithCancelCause) — Go 1.20+: attach a specific cancellation reason
- [`context.WithoutCancel`](https://pkg.go.dev/context#WithoutCancel) — Go 1.21+: detach from parent cancellation
- [`context.Cause`](https://pkg.go.dev/context#Cause) — Go 1.21+: retrieve the cancellation cause
- [Go Blog: Contexts and structs](https://go.dev/blog/context-and-structs) — why context belongs in function signatures, not struct fields
- [Go Blog: Context](https://go.dev/blog/context) — the original context blog post (2014), still essential reading
- [Language Spec — Channel types](https://go.dev/ref/spec#Channel_types) — the `<-chan struct{}` pattern used by `Done()`
- [`database/sql`](https://pkg.go.dev/database/sql) — QueryContext, ExecContext, BeginTx and other context-aware methods
- [`net/http.Request.WithContext`](https://pkg.go.dev/net/http#Request.WithContext) — attaching a context to an HTTP request
- [`golang.org/x/sync/errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) — structured concurrency with context propagation
- [Go Memory Model](https://go.dev/ref/mem) — happens-before guarantees relevant to channel-based cancellation
