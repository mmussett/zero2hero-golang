# Day 17: Context

## Core Concept: Propagate Cancellation, Not Panics

[`context.Context`](https://pkg.go.dev/context#Context) carries deadlines, cancellation signals, and request-scoped values across API boundaries and goroutines. Pass it as the **first argument** to every function that does I/O or blocks.

```go
func doWork(ctx context.Context, url string) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil { return err }
    resp, err := http.DefaultClient.Do(req)
    // if ctx is cancelled, err will be context.Canceled or context.DeadlineExceeded
    ...
}
```

## Creating Contexts

```go
ctx := context.Background()             // root context — never nil, never cancelled

// Cancellable
ctx, cancel := context.WithCancel(ctx)
defer cancel()                          // always call cancel to free resources

// Timeout
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

// Absolute deadline
ctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
defer cancel()

// Values (request-scoped data)
type ctxKey string
ctx = context.WithValue(ctx, ctxKey("traceID"), "abc-123")
id := ctx.Value(ctxKey("traceID")).(string)
```

Use unexported struct or named string types as context keys to avoid collisions.

## Checking Cancellation

```go
select {
case <-ctx.Done():
    return ctx.Err()  // context.Canceled or context.DeadlineExceeded
case result := <-workCh:
    return result
}
```

Or in a loop:

```go
for {
    if err := ctx.Err(); err != nil { return err }
    // ... do a unit of work
}
```

## TCP Echo Server

```go
listener, err := net.Listen("tcp", ":8080")
// ...
for {
    conn, err := listener.Accept()
    go handleConn(ctx, conn)
}

func handleConn(ctx context.Context, conn net.Conn) {
    defer conn.Close()
    // Watch for ctx cancellation alongside conn I/O
}
```

## Labs

### Lab 1: context.Background and context.TODO

**What you'll practise:** Understanding the two root context constructors and when to use each.

**Task:**
Create both `context.Background()` and `context.TODO()`, print their `String()` representations, and experiment with their `Done()`, `Err()`, and `Deadline()` methods. Understand when each is appropriate.

**Steps:**
1. Call `ctx := context.Background()` and print `ctx` — observe the output
2. Call `ctx := context.TODO()` and print `ctx`
3. Try `ctx.Done()` — it returns `nil` for both; try to receive on it and observe it blocks forever
4. Write a comment explaining: Background = top-level main/test; TODO = placeholder during refactoring

```go
bg := context.Background()
todo := context.TODO()

fmt.Println(bg)           // context.background
fmt.Println(todo)         // context.todo
fmt.Println(bg.Err())     // <nil>
fmt.Println(bg.Done())    // <nil> (never cancelled)

deadline, ok := bg.Deadline()
fmt.Println(deadline, ok) // 0001-01-01 00:00:00 +0000 UTC false
```

**Expected output:**
```
context.background
context.todo
<nil>
<nil>
0001-01-01 00:00:00 +0000 UTC false
```

**Checkpoint:** Both print correctly and neither `Done()` channel is readable (they are `nil`).

---

### Lab 2: context.WithCancel — Manual Cancellation

**What you'll practise:** Deriving a cancellable child context and observing the `Done` channel close.

**Task:**
Create a parent context, derive a child with `WithCancel`, launch a goroutine that blocks on `ctx.Done()`, call `cancel()`, and confirm the goroutine exits with `ctx.Err() == context.Canceled`.

**Steps:**
1. `parent := context.Background()`
2. `ctx, cancel := context.WithCancel(parent)`
3. Launch a goroutine: `select { case <-ctx.Done(): fmt.Println(ctx.Err()) }`
4. Sleep 100ms, call `cancel()`, wait for goroutine, confirm error is `context.Canceled`

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    <-ctx.Done()
    fmt.Println("goroutine saw:", ctx.Err())
}()

time.Sleep(100 * time.Millisecond)
cancel()
wg.Wait()
```

**Expected output:**
```
goroutine saw: context canceled
```

**Checkpoint:** `ctx.Err()` returns `context.Canceled` (not `nil`) after `cancel()` is called.

---

### Lab 3: context.WithTimeout — HTTP GET with Deadline

**What you'll practise:** Wrapping a network call with a timeout context to prevent hanging indefinitely.

**Task:**
Make an HTTP GET to a slow or unreachable URL using a 2-second timeout context. Observe that the request fails with `context.DeadlineExceeded`. Then test with a reachable URL within the timeout.

**Steps:**
1. `ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)`
2. Build the request with `http.NewRequestWithContext(ctx, ...)`
3. Use `http.DefaultClient.Do(req)` and inspect the error
4. Check `errors.Is(err, context.DeadlineExceeded)` — print a friendly message

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, http.MethodGet,
    "http://10.255.255.1", nil) // unreachable IP
if err != nil {
    log.Fatal(err)
}

_, err = http.DefaultClient.Do(req)
if errors.Is(err, context.DeadlineExceeded) {
    fmt.Println("timed out as expected:", err)
} else {
    fmt.Println("unexpected error:", err)
}
```

**Expected output:**
```
timed out as expected: Get "http://10.255.255.1": context deadline exceeded
```

**Checkpoint:** The program exits within ~2 seconds (not hanging), and the error wraps `context.DeadlineExceeded`.

---

### Lab 4: context.WithDeadline — Absolute Deadline

**What you'll practise:** Setting an absolute point-in-time deadline and distinguishing it from a relative timeout.

**Task:**
Create a context with an absolute deadline 3 seconds from now using `WithDeadline`. Print the deadline. Start a goroutine that sleeps 5 seconds (longer than the deadline). Observe it cancelled.

**Steps:**
1. `deadline := time.Now().Add(3 * time.Second)`
2. `ctx, cancel := context.WithDeadline(context.Background(), deadline)`
3. Print `ctx.Deadline()` to confirm the absolute time
4. Launch a goroutine that waits 5 seconds but also selects on `ctx.Done()` — the context wins

```go
dl := time.Now().Add(3 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), dl)
defer cancel()

t, ok := ctx.Deadline()
fmt.Printf("deadline: %v (set: %v)\n", t.Format("15:04:05"), ok)

select {
case <-time.After(5 * time.Second):
    fmt.Println("timer fired")
case <-ctx.Done():
    fmt.Println("context expired:", ctx.Err())
}
```

**Expected output:**
```
deadline: 00:00:03 (set: true)
context expired: context deadline exceeded
```

**Checkpoint:** The `select` hits `ctx.Done()` after ~3 seconds, not after 5 seconds.

---

### Lab 5: context.WithValue — Request-Scoped Data

**What you'll practise:** Storing and retrieving typed values in a context without key collisions.

**Task:**
Define an unexported key type to avoid collisions. Store a request ID in the context, pass the context through two function calls, and retrieve the request ID in the innermost function. Show what happens with a plain `string` key (collision-prone).

**Steps:**
1. Define `type ctxKey string` and `const keyRequestID ctxKey = "requestID"`
2. `ctx := context.WithValue(context.Background(), keyRequestID, "req-abc-123")`
3. Pass `ctx` to a `handleRequest(ctx)` that calls `processRequest(ctx)`
4. In `processRequest`, retrieve with `ctx.Value(keyRequestID).(string)` and print it

```go
type ctxKey string

const keyRequestID ctxKey = "requestID"

func processRequest(ctx context.Context) {
    id, ok := ctx.Value(keyRequestID).(string)
    if !ok {
        fmt.Println("no request ID in context")
        return
    }
    fmt.Println("processing request:", id)
}

func main() {
    ctx := context.WithValue(context.Background(), keyRequestID, "req-abc-123")
    processRequest(ctx)

    // Demonstrate collision: plain string key is shadowed by another package's same key
    ctx2 := context.WithValue(ctx, "requestID", "collision!")
    fmt.Println(ctx2.Value(keyRequestID)) // still "req-abc-123" — no collision
}
```

**Expected output:**
```
processing request: req-abc-123
req-abc-123
```

**Checkpoint:** The typed key retrieves the correct value, and the plain `string` key does not shadow it.

---

### Lab 6: Propagate Cancellation — Three-Level Chain

**What you'll practise:** Demonstrating that cancelling a parent context cancels all descendants.

**Task:**
Create a three-level chain: goroutine A creates a cancellable context, B derives from A's context, C derives from B's. Cancel A's context and verify that B and C both observe the cancellation.

**Steps:**
1. `ctxA, cancelA := context.WithCancel(context.Background())`
2. `ctxB, cancelB := context.WithCancel(ctxA)` — derive from A
3. `ctxC, cancelC := context.WithCancel(ctxB)` — derive from B
4. Start goroutines A, B, C each blocking on their respective `ctx.Done()`
5. `cancelA()` — all three should receive the cancellation

```go
ctxA, cancelA := context.WithCancel(context.Background())
ctxB, cancelB := context.WithCancel(ctxA)
ctxC, cancelC := context.WithCancel(ctxB)
defer cancelB()
defer cancelC()

var wg sync.WaitGroup
for name, ctx := range map[string]context.Context{"A": ctxA, "B": ctxB, "C": ctxC} {
    wg.Add(1)
    go func(n string, c context.Context) {
        defer wg.Done()
        <-c.Done()
        fmt.Printf("%s cancelled: %v\n", n, c.Err())
    }(name, ctx)
}

time.Sleep(100 * time.Millisecond)
cancelA() // cancels A, B, and C
wg.Wait()
```

**Expected output (order varies):**
```
A cancelled: context canceled
B cancelled: context canceled
C cancelled: context canceled
```

**Checkpoint:** All three goroutines exit after a single `cancelA()` call.

---

### Lab 7: Context in HTTP Handlers — Early Return on Cancellation

**What you'll practise:** Reading `r.Context()` in an HTTP handler and returning early if the client disconnects.

**Task:**
Write an HTTP handler that simulates 5 seconds of work using a loop with `time.Sleep(500ms)` per iteration. Check `ctx.Done()` inside the loop. Use `curl` (or a Go test client) to disconnect early and verify the handler stops.

**Steps:**
1. Create an HTTP handler that gets `ctx := r.Context()`
2. Loop 10 times, each iteration sleeping 500ms then checking `ctx.Err()`
3. If `ctx.Err() != nil`, log "client disconnected" and return
4. Start the server on `:8080`, use `curl --max-time 2` to hit it and watch the server log

```go
func slowHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    for i := 0; i < 10; i++ {
        select {
        case <-ctx.Done():
            log.Printf("client disconnected after %d iterations: %v", i, ctx.Err())
            return
        case <-time.After(500 * time.Millisecond):
            fmt.Fprintf(w, "tick %d\n", i)
            if f, ok := w.(http.Flusher); ok { f.Flush() }
        }
    }
    fmt.Fprintln(w, "done")
}
```

**Expected output (server log):**
```
client disconnected after 3 iterations: context canceled
```

**Checkpoint:** The handler logs "client disconnected" when `curl` times out, proving it does not keep working unnecessarily.

---

### Final Lab (Project): Cancellable HTTP Downloader + TCP Echo Server

**What you'll practise:** Applying all context patterns — `WithTimeout`, `WithCancel`, `WithValue` — to two real programs.

**Task:**
Build both a concurrent HTTP downloader and a TCP echo server, both fully context-aware.

**Steps:**
1. **Downloader:** Accept URLs from `os.Args`. Download each concurrently with `http.NewRequestWithContext`. Cancel all in-flight downloads after 10 seconds with `context.WithTimeout`. Report success, failure, and cancelled separately.
2. **TCP echo server:** Accept connections in a loop. Handle each with `go handleConn(ctx, conn)`. Shut down cleanly when `context.WithTimeout` expires — stop accepting new connections and close existing ones.
3. Wire cancellation: a single top-level `cancel` function stops both the downloader and the server.

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Downloader
for _, url := range os.Args[1:] {
    wg.Add(1)
    go func(u string) {
        defer wg.Done()
        req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
        resp, err := http.DefaultClient.Do(req)
        // classify: success / deadline exceeded / other error
    }(url)
}

// TCP echo server shuts down when ctx expires
go func() {
    <-ctx.Done()
    listener.Close()
}()
```

**Expected output:**
```
downloaded https://example.com (1256 bytes)
cancelled https://slow.example.com: context deadline exceeded
errors: 0  cancelled: 1  success: 1
```

**Checkpoint:** All goroutines exit within the timeout window. `go run -race main.go` shows no races.

---

## Day Project: Cancellable HTTP Downloader

Write a program that:
1. Takes a list of URLs (hardcoded or from `os.Args`)
2. Downloads each concurrently with [`http.NewRequestWithContext`](https://pkg.go.dev/net/http#NewRequestWithContext)
3. Cancels all in-flight downloads after a configurable timeout (e.g. 10s)
4. Reports success, failure, and cancelled downloads separately

Also write a minimal TCP echo server that:
- Accepts connections
- Echoes every line back in uppercase
- Shuts down cleanly when a [`context.WithTimeout`](https://pkg.go.dev/context#WithTimeout) expires

**Extension ideas:** implement backpressure by limiting concurrent downloads with a semaphore; add retry with exponential backoff.

## Official Documentation

- [`context`](https://pkg.go.dev/context) — Context, Background, WithCancel, WithTimeout, WithDeadline, WithValue
- [`net/http`](https://pkg.go.dev/net/http) — NewRequestWithContext, DefaultClient, HTTP methods
- [`net`](https://pkg.go.dev/net) — Listen, Conn for TCP server
- [`os`](https://pkg.go.dev/os) — `os.Args` for URL list input
- [Go Blog: Go Concurrency Patterns: Context](https://go.dev/blog/context) — context usage patterns
- [Go Blog: Contexts and structs](https://go.dev/blog/context-and-structs) — why context goes in arguments not structs
- [Effective Go: Concurrency](https://go.dev/doc/effective_go#concurrency) — cancellation and coordination
- [Go Tour: Concurrency](https://go.dev/tour/concurrency/1) — goroutines and channels foundation
