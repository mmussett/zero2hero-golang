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

## Day Project: Add Structured Logging to the Notes API

1. Initialise a JSON `slog.Logger` in `main`, set as default
2. Write a `requestLogger` middleware that logs method, path, status, duration, and trace ID
3. Inject per-request logger into context; use it in handlers
4. Log at `Info` for successful operations, `Error` for failures (with `"err"` attribute)
5. Make log level configurable via `LOG_LEVEL` env var

Run with: `go run .`

**Extension ideas:** implement a custom `slog.Handler` that redacts PII fields; export trace IDs in response headers (`X-Trace-ID`).
