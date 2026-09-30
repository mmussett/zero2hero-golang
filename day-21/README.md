# Day 21: net/http

## Core Concept: Handlers Are Functions

Go's standard library HTTP server is production-capable. A handler is anything with a `ServeHTTP(ResponseWriter, *Request)` method — or a `func(ResponseWriter, *Request)` adapted with `http.HandlerFunc`.

## Basic Server

```go
http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    if name == "" { name = "World" }
    fmt.Fprintf(w, "Hello, %s!\n", name)
})

log.Fatal(http.ListenAndServe(":8080", nil))
```

## ServeMux (Router)

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /api/users", listUsers)       // Go 1.22 method+path pattern
mux.HandleFunc("POST /api/users", createUser)
mux.HandleFunc("GET /api/users/{id}", getUser)    // path parameter

srv := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
}
log.Fatal(srv.ListenAndServe())
```

## Middleware

A middleware wraps a handler:

```go
func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

mux.Handle("/api/", logging(mux))
```

## JSON Responses

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}
```

## HTTP Client

```go
client := &http.Client{Timeout: 10 * time.Second}

resp, err := client.Get("https://api.example.com/data")
if err != nil { return err }
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

var result MyResponse
json.NewDecoder(resp.Body).Decode(&result)
```

Always close `resp.Body` and check status code separately from `err`.

## Graceful Shutdown

```go
srv := &http.Server{Addr: ":8080", Handler: mux}

go func() { log.Fatal(srv.ListenAndServe()) }()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

## Day Project: HTTP Echo Server + Client

**Server:**
- `GET /echo?msg=hello` — echoes the message
- `POST /echo` — echoes the JSON body back
- `GET /headers` — returns all request headers as JSON
- Logging middleware
- Graceful shutdown on SIGINT

**Client:**
- A separate program (or `--client` flag) that exercises each endpoint
- Uses `http.Client` with a 5-second timeout
- Handles non-200 responses as errors

**Extension ideas:** add an `Authorization: Bearer` middleware; implement request ID injection with context.
