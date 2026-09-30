# Day 17: Context

## Core Concept: Propagate Cancellation, Not Panics

`context.Context` carries deadlines, cancellation signals, and request-scoped values across API boundaries and goroutines. Pass it as the **first argument** to every function that does I/O or blocks.

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

## Day Project: Cancellable HTTP Downloader

Write a program that:
1. Takes a list of URLs (hardcoded or from `os.Args`)
2. Downloads each concurrently with `http.NewRequestWithContext`
3. Cancels all in-flight downloads after a configurable timeout (e.g. 10s)
4. Reports success, failure, and cancelled downloads separately

Also write a minimal TCP echo server that:
- Accepts connections
- Echoes every line back in uppercase
- Shuts down cleanly when a `context.WithTimeout` expires

**Extension ideas:** implement backpressure by limiting concurrent downloads with a semaphore; add retry with exponential backoff.
