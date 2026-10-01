# Day 25: Structured Logging with log/slog

## Core Concept: Logs Are Data, Not Text

`log/slog` (Go 1.21+) replaces the old `log` package for production code. Structured logs — key-value pairs — can be parsed, searched, and aggregated by log platforms.

## Basic Usage

```go
import "log/slog"

slog.Info("server started", "addr", ":8080")
slog.Error("request failed", "err", err, "path", r.URL.Path)
slog.Debug("cache hit", "key", key, "ttl", ttl)
```

## Configuring the Global Logger

```go
// JSON handler (for production)
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// Text handler (for development)
slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
```

## Logger With Context

```go
// Attach attributes to a logger
log := slog.With("service", "notes-api", "version", "1.0.0")
log.Info("request received")

// Logger in context (for request-scoped logging)
func withLogger(ctx context.Context, log *slog.Logger) context.Context {
    return context.WithValue(ctx, loggerKey{}, log)
}

func loggerFrom(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}
```

## Middleware: Request ID + Per-Request Logger

```go
func requestLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start   := time.Now()
        traceID := newTraceID()
        log := slog.With(
            "trace_id", traceID,
            "method",   r.Method,
            "path",     r.URL.Path,
        )
        ctx := withLogger(r.Context(), log)
        log.Info("request started")

        rw := &statusWriter{ResponseWriter: w}
        next.ServeHTTP(rw, r.WithContext(ctx))

        log.Info("request completed",
            "status",   rw.status,
            "duration", time.Since(start).String(),
        )
    })
}
```

## Log Levels

```go
slog.Debug(...)  // verbose dev info
slog.Info(...)   // normal operation
slog.Warn(...)   // unexpected but recoverable
slog.Error(...)  // failures requiring attention
```

## Labs

### Lab 1: Default slog — Your First Structured Log

**What you'll practise:** Using the global slog functions with key-value pairs and observing the default text output.

**Task:**
Write a program that logs several events using `slog.Info`, `slog.Warn`, and `slog.Error` with meaningful key-value attributes, then observe the default text format output.

**Steps:**
1. Create `main.go` with `package main`
2. Import `"log/slog"` and `"errors"`
3. Call `slog.Info`, `slog.Warn`, and `slog.Error` with at least two key-value pairs each

```go
package main

import (
    "errors"
    "log/slog"
)

func main() {
    slog.Info("server started", "addr", ":8080", "pid", 1234)
    slog.Warn("high memory usage", "percent", 87.5, "threshold", 80)
    err := errors.New("connection refused")
    slog.Error("database unreachable", "err", err, "host", "localhost:5432")
}
```

**Expected output:**
```
2024/01/15 10:00:00 INFO server started addr=:8080 pid=1234
2024/01/15 10:00:00 WARN high memory usage percent=87.5 threshold=80
2024/01/15 10:00:00 ERROR database unreachable err="connection refused" host=localhost:5432
```

**Checkpoint:** Three log lines appear with key=value pairs. The level prefix (INFO/WARN/ERROR) is visible in each line.

---

### Lab 2: JSON Handler — Structured Output for Production

**What you'll practise:** Switching to `slog.NewJSONHandler` and comparing JSON vs text output.

**Task:**
Create a JSON-format logger, set it as the global default, and log the same messages as Lab 1. Compare the output structure.

**Steps:**
1. Add `"os"` to your imports
2. Create a JSON handler writing to `os.Stdout`
3. Wrap it in `slog.New` and call `slog.SetDefault`
4. Re-run the same three log calls and observe the difference

```go
import (
    "log/slog"
    "os"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    slog.Info("server started", "addr", ":8080", "pid", 1234)
    slog.Warn("high memory usage", "percent", 87.5, "threshold", 80)
}
```

**Expected output:**
```
{"time":"2024-01-15T10:00:00Z","level":"INFO","msg":"server started","addr":":8080","pid":1234}
{"time":"2024-01-15T10:00:00Z","level":"WARN","msg":"high memory usage","percent":87.5,"threshold":80}
```

**Checkpoint:** Each log line is valid JSON. Pipe through `| python3 -m json.tool` to pretty-print and confirm the structure.

---

### Lab 3: Log Levels — Controlling Verbosity

**What you'll practise:** Filtering log output by setting a minimum log level with `slog.HandlerOptions`.

**Task:**
Configure a handler with `Level: slog.LevelWarn` and observe that `Debug` and `Info` calls are silently dropped.

**Steps:**
1. Create a handler with `&slog.HandlerOptions{Level: slog.LevelWarn}`
2. Log one message at each level: Debug, Info, Warn, Error
3. Verify only Warn and Error appear in the output

```go
opts   := &slog.HandlerOptions{Level: slog.LevelWarn}
logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

logger.Debug("this is debug", "detail", "verbose")
logger.Info("this is info", "user", "alice")
logger.Warn("this is warn", "latency_ms", 450)
logger.Error("this is error", "err", errors.New("timeout"))
```

**Expected output:**
```
{"time":"...","level":"WARN","msg":"this is warn","latency_ms":450}
{"time":"...","level":"ERROR","msg":"this is error","err":"timeout"}
```

**Checkpoint:** Exactly two lines appear. The Debug and Info lines are absent.

---

### Lab 4: Structured Attributes — Child Loggers with Permanent Fields

**What you'll practise:** Using `slog.With` to create a child logger that embeds permanent key-value fields on every line.

**Task:**
Create a base logger, then derive a child logger with `service` and `version` fields. All log calls through the child automatically include those fields.

**Steps:**
1. Create a JSON logger as the base
2. Call `.With("service", "api", "version", "1.0.0")` to produce a child logger
3. Log several messages through the child and observe the permanent fields

```go
base := slog.New(slog.NewJSONHandler(os.Stdout, nil))
log  := base.With("service", "api", "version", "1.0.0")

log.Info("starting up", "port", 8080)
log.Info("connected to db", "host", "localhost")
log.Error("request failed", "path", "/users", "err", errors.New("not found"))
```

**Expected output:**
```
{"time":"...","level":"INFO","msg":"starting up","service":"api","version":"1.0.0","port":8080}
{"time":"...","level":"INFO","msg":"connected to db","service":"api","version":"1.0.0","host":"localhost"}
{"time":"...","level":"ERROR","msg":"request failed","service":"api","version":"1.0.0","path":"/users","err":"not found"}
```

**Checkpoint:** Every line contains `"service":"api"` and `"version":"1.0.0"` without adding them to each individual call.

---

### Lab 5: Context Logging — Request-Scoped Logger

**What you'll practise:** Storing and retrieving a `*slog.Logger` from `context.Context` using a typed key.

**Task:**
Write `WithLogger` and `LoggerFrom` helper functions. Store a logger enriched with a request ID in context and retrieve it inside a simulated handler function.

**Steps:**
1. Define a private `type logKey struct{}`
2. Write `WithLogger(ctx, log)` and `LoggerFrom(ctx)` functions
3. Simulate a handler that reads the logger from context — the request ID flows through automatically

```go
type logKey struct{}

func WithLogger(ctx context.Context, log *slog.Logger) context.Context {
    return context.WithValue(ctx, logKey{}, log)
}

func LoggerFrom(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(logKey{}).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}

func handleRequest(ctx context.Context) {
    log := LoggerFrom(ctx)
    log.Info("processing request")
    log.Info("fetching user", "user_id", 42)
}

func main() {
    base := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    log  := base.With("request_id", "req-abc-123")
    ctx  := WithLogger(context.Background(), log)
    handleRequest(ctx)
}
```

**Expected output:**
```
{"time":"...","level":"INFO","msg":"processing request","request_id":"req-abc-123"}
{"time":"...","level":"INFO","msg":"fetching user","request_id":"req-abc-123","user_id":42}
```

**Checkpoint:** Both log lines carry `"request_id"` without `handleRequest` referencing the request ID directly.

---

### Lab 6: HTTP Request Middleware — Logging Every Request

**What you'll practise:** Building an `http.Handler` middleware that logs method, path, status code, and duration for every incoming request.

**Task:**
Write a `requestLogger` middleware. It wraps `ResponseWriter` to capture the status code, then logs a completion entry with timing after the handler returns.

**Steps:**
1. Define `statusWriter` embedding `http.ResponseWriter` with a `status int` field and a `WriteHeader` override
2. Write `requestLogger(log *slog.Logger, next http.Handler) http.Handler`
3. Generate a short request ID with `fmt.Sprintf("%08x", rand.Int32())`
4. Register a simple handler at `/hello`, wrap it, and start the server

```go
type statusWriter struct {
    http.ResponseWriter
    status int
}

func (sw *statusWriter) WriteHeader(code int) {
    sw.status = code
    sw.ResponseWriter.WriteHeader(code)
}

func requestLogger(log *slog.Logger, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        reqID := fmt.Sprintf("%08x", rand.Int32())
        l     := log.With("request_id", reqID, "method", r.Method, "path", r.URL.Path)

        sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(sw, r.WithContext(WithLogger(r.Context(), l)))

        l.Info("request completed",
            "status",      sw.status,
            "duration_ms", time.Since(start).Milliseconds(),
        )
    })
}
```

**Expected output:**
```
{"time":"...","level":"INFO","msg":"request completed","request_id":"1a2b3c4d","method":"GET","path":"/hello","status":200,"duration_ms":0}
```

**Checkpoint:** Run `go run .` then `curl http://localhost:8080/hello`. One structured JSON log line appears per request with a unique `request_id`.

---

### Lab 7: Custom Handler — Colourised Development Output

**What you'll practise:** Implementing the `slog.Handler` interface to produce human-readable, ANSI-coloured output for local development.

**Task:**
Build a `ColorHandler` that prints records in the format `[LEVEL] message key=value …` with a different colour per level.

**Steps:**
1. Define `ColorHandler` implementing `slog.Handler` (four methods: `Enabled`, `Handle`, `WithAttrs`, `WithGroup`)
2. Map levels to ANSI colour codes: DEBUG=cyan, INFO=green, WARN=yellow, ERROR=red
3. Format attributes as `key=value` pairs on the same line
4. Set it as the global default and log at all four levels to confirm the colours

```go
const (
    colorReset  = "\033[0m"
    colorCyan   = "\033[36m"
    colorGreen  = "\033[32m"
    colorYellow = "\033[33m"
    colorRed    = "\033[31m"
)

func levelColor(l slog.Level) string {
    switch {
    case l < slog.LevelInfo:  return colorCyan
    case l < slog.LevelWarn:  return colorGreen
    case l < slog.LevelError: return colorYellow
    default:                   return colorRed
    }
}
```

**Expected output:**
```
[DEBUG] cache miss key=user:42
[INFO]  server started addr=:8080
[WARN]  slow query duration_ms=312
[ERROR] db connection failed err="connection refused"
```
(Each level prefix appears in a distinct colour in a real terminal.)

**Checkpoint:** Run `go run .` and confirm that each level label is rendered in a different colour. Text is readable and attributes appear inline.

---

### Final Lab (Project): Structured Logging with log/slog — Notes API

**What you'll practise:** Combining a JSON handler, request middleware, and context-scoped logging into a complete HTTP service.

**Task:**
Add structured logging to the Notes API from Day 22 using a JSON `slog.Logger`, a `requestLogger` middleware, and per-request logger injection via context.

**Steps:**
1. Initialise a JSON `slog.Logger` in `main`, set as default
2. Write a `requestLogger` middleware that logs method, path, status, duration, and trace ID
3. Inject per-request logger into context; use it in handlers
4. Log at `Info` for successful operations, `Error` for failures (with `"err"` attribute)
5. Make log level configurable via `LOG_LEVEL` env var

```go
level := slog.LevelInfo
if os.Getenv("LOG_LEVEL") == "debug" {
    level = slog.LevelDebug
}
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: level,
}))
slog.SetDefault(logger)
```

**Expected output:**
```
{"time":"...","level":"INFO","msg":"server started","addr":":8080"}
{"time":"...","level":"INFO","msg":"request completed","method":"POST","path":"/notes","status":201,"duration_ms":2}
```

**Checkpoint:** Run `go run .` and `curl -X POST localhost:8080/notes -d '{"title":"test"}'`. One JSON log line appears per request. Running with `LOG_LEVEL=debug go run .` reveals additional debug lines.

**Extension ideas:** implement a custom `slog.Handler` that redacts PII fields; export trace IDs in response headers (`X-Trace-ID`).

## Official Documentation

- [`log/slog`](https://pkg.go.dev/log/slog) — `Logger`, `Handler`, `NewJSONHandler`, `NewTextHandler`, `HandlerOptions`, `SetDefault`, `With`, `Info`, `Error`, `Debug`, `Warn`, `LevelInfo`, `LevelDebug`
- [`context`](https://pkg.go.dev/context) — `WithValue`, `Value` for injecting per-request loggers
- [`os`](https://pkg.go.dev/os) — `Stdout`, `Getenv` for output target and log level config
- [`net/http`](https://pkg.go.dev/net/http) — `Handler`, `ResponseWriter`, `Request` used in middleware
- [`time`](https://pkg.go.dev/time) — `Now`, `Since` for request duration measurement
- [Go Blog: Structured Logging with slog](https://go.dev/blog/slog) — official introduction to `log/slog`
