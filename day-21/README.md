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

## Labs

### Lab 1: Simple HTTP Server

**What you'll practise:** Registering a handler with `http.HandleFunc` and serving responses on a local port.

**Task:**
Write the smallest possible HTTP server: one handler at `/` that responds with `"Hello, World!\n"`. Run it on `:8080` and verify with `curl`.

**Steps:**
1. Call `http.HandleFunc("/", ...)` with a function that writes the greeting
2. Call `http.ListenAndServe(":8080", nil)` and wrap with `log.Fatal`
3. In a second terminal run `curl http://localhost:8080/`

```go
package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, World!")
    })
    log.Println("listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Expected output:**
```
$ curl http://localhost:8080/
Hello, World!
```

**Checkpoint:** `curl` returns 200 and the greeting text; the server logs the listen message on startup.

---

### Lab 2: Request Parsing

**What you'll practise:** Reading query parameters, inspecting request headers, and logging method, path, and status.

**Task:**
Add a `/greet` endpoint that reads `?name=` from the query string and responds `"Hello, <name>!"`. Log the method, path, and response status for every request.

**Steps:**
1. In the handler, call `r.URL.Query().Get("name")`; default to `"World"` if empty
2. Set `Content-Type: text/plain` on the response header
3. Print request details to `log` before writing the body
4. Test with `curl "http://localhost:8080/greet?name=Alice"`

```go
http.HandleFunc("/greet", func(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    if name == "" {
        name = "World"
    }
    log.Printf("%s %s", r.Method, r.URL.Path)
    w.Header().Set("Content-Type", "text/plain")
    fmt.Fprintf(w, "Hello, %s!\n", name)
})
```

**Expected output:**
```
$ curl "http://localhost:8080/greet?name=Alice"
Hello, Alice!
# server logs: GET /greet
```

**Checkpoint:** Missing `?name` defaults to `"World"`; the server log line appears for every request.

---

### Lab 3: JSON API Handler

**What you'll practise:** Decoding a JSON request body and responding with JSON, including error handling for malformed input.

**Task:**
Create `POST /echo` that reads a JSON body `{"message":"..."}` and responds with `{"echo":"...","length":N}`. Return `400` with an error message if the body is malformed JSON.

**Steps:**
1. Decode the request body using `json.NewDecoder(r.Body).Decode(&req)`
2. If decoding fails, write HTTP 400 and a JSON error body
3. Build the response struct and encode it with `json.NewEncoder(w).Encode`
4. Test with `curl -X POST -H 'Content-Type: application/json' -d '{"message":"hi"}' http://localhost:8080/echo`

```go
type echoRequest  struct { Message string `json:"message"` }
type echoResponse struct {
    Echo   string `json:"echo"`
    Length int    `json:"length"`
}

http.HandleFunc("POST /echo", func(w http.ResponseWriter, r *http.Request) {
    var req echoRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(echoResponse{Echo: req.Message, Length: len(req.Message)})
})
```

**Expected output:**
```
$ curl -s -X POST -H 'Content-Type: application/json' -d '{"message":"hi"}' http://localhost:8080/echo
{"echo":"hi","length":2}
$ curl -s -X POST -d 'not json' http://localhost:8080/echo
{"error":"invalid character 'o' in literal null (expecting 'u')"}
```

**Checkpoint:** Valid JSON returns 200 with the echo; invalid JSON returns 400 with an error field.

---

### Lab 4: http.ServeMux Routing (Go 1.22)

**What you'll practise:** Using the Go 1.22 method+path pattern syntax and extracting path values.

**Task:**
Create a `ServeMux` with three routes: `GET /users` (list), `GET /users/{id}` (get one), `POST /users` (create). Use `r.PathValue("id")` to extract the path parameter and return a JSON response.

**Steps:**
1. Create `mux := http.NewServeMux()`
2. Register `"GET /users"`, `"GET /users/{id}"`, and `"POST /users"`
3. In the `GET /users/{id}` handler, call `r.PathValue("id")` and return `{"id":"<id>"}`
4. Start the server with `http.ListenAndServe(":8080", mux)`

```go
mux := http.NewServeMux()

mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode([]map[string]any{{"id": "1", "name": "Alice"}})
})

mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"id": id})
})

mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"status": "created"})
})

log.Fatal(http.ListenAndServe(":8080", mux))
```

**Expected output:**
```
$ curl http://localhost:8080/users/42
{"id":"42"}
$ curl -X POST http://localhost:8080/users
{"status":"created"}
```

**Checkpoint:** `GET /users/42` extracts `"42"` from the path; `POST /users` returns 201 Created.

---

### Lab 5: Middleware

**What you'll practise:** Writing a `func(http.Handler) http.Handler` middleware and chaining two middlewares.

**Task:**
Write a `logging` middleware that logs the method, path, and elapsed time after each request. Write a `recovery` middleware that catches panics and returns 500. Chain them: `recovery(logging(mux))`.

**Steps:**
1. Implement `logging(next http.Handler) http.Handler` — record `time.Now()`, call `next.ServeHTTP`, then log
2. Implement `recovery(next http.Handler) http.Handler` — defer a function that calls `recover()` and writes 500
3. Pass `recovery(logging(mux))` as the handler to `http.ListenAndServe`
4. Add a `/panic` route that panics, and verify recovery returns 500

```go
func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

func recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

**Expected output:**
```
$ curl http://localhost:8080/panic
internal server error
# server logs: GET /panic <duration>
```

**Checkpoint:** The server does not crash on a panic; the logging line appears for every request.

---

### Lab 6: HTTP Client

**What you'll practise:** Making HTTP requests with `http.Get` and with a custom `http.Client` for headers and timeouts.

**Task:**
Write two functions: `simpleFetch(url string)` using `http.Get`, and `fetchWithHeaders(url string, headers map[string]string)` using `http.NewRequest` + a `http.Client` with a 5-second timeout. Target a public API like `https://httpbin.org/get`.

**Steps:**
1. `simpleFetch`: call `http.Get`, check error, defer `resp.Body.Close()`, check status, print body
2. `fetchWithHeaders`: create `http.NewRequest("GET", url, nil)`, add headers, use `client.Do(req)`
3. Both functions return an error if status != 200

```go
func simpleFetch(url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
    io.Copy(os.Stdout, resp.Body)
    return nil
}

func fetchWithHeaders(url string, headers map[string]string) error {
    client := &http.Client{Timeout: 5 * time.Second}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return err
    }
    for k, v := range headers {
        req.Header.Set(k, v)
    }
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
    io.Copy(os.Stdout, resp.Body)
    return nil
}
```

**Expected output:**
```
(JSON body from httpbin.org containing the request headers)
```

**Checkpoint:** Both functions compile and return nil on a successful request; a timeout causes a meaningful error.

---

### Lab 7: Graceful Shutdown

**What you'll practise:** Catching OS signals and calling `http.Server.Shutdown` so in-flight requests complete before the process exits.

**Task:**
Build an HTTP server that listens for `SIGINT` (Ctrl-C), then calls `server.Shutdown` with a 5-second context. Add a `/slow` handler that sleeps 3 seconds so you can test that in-flight requests finish.

**Steps:**
1. Start the server in a goroutine with `go srv.ListenAndServe()`
2. Block on a channel that `signal.Notify` writes to on SIGINT/SIGTERM
3. Create a `context.WithTimeout(context.Background(), 5*time.Second)` and call `srv.Shutdown(ctx)`
4. `curl /slow` then immediately Ctrl-C the server; the response should still arrive

```go
srv := &http.Server{Addr: ":8080", Handler: mux}

go func() {
    if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
        log.Fatal(err)
    }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
log.Println("shutting down...")

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
    log.Fatal("forced shutdown:", err)
}
log.Println("server stopped")
```

**Expected output:**
```
$ curl http://localhost:8080/slow   # takes 3 seconds, then:
slow response
# server logs: shutting down... server stopped
```

**Checkpoint:** The server exits cleanly after Ctrl-C; the slow request completes without a connection reset.

---

### Final Lab: HTTP Echo Server + Client

**What you'll practise:** Combining ServeMux routing, middleware, JSON handling, and graceful shutdown into a complete server with a matching client.

**Task:**
Build an HTTP echo server and a client that exercises it. The server handles three endpoints with logging middleware and shuts down gracefully on SIGINT. The client targets each endpoint and prints results.

**Steps:**
1. Server: implement `GET /echo?msg=`, `POST /echo` (JSON body), `GET /headers`
2. Wrap the mux with a logging middleware
3. Implement graceful shutdown on SIGINT with a 5-second timeout
4. Client: use `http.Client` with 5s timeout; call all three endpoints and print responses
5. Run server and client in two terminals; verify all three endpoints respond correctly

```go
// Server handler for GET /headers
mux.HandleFunc("GET /headers", func(w http.ResponseWriter, r *http.Request) {
    headers := make(map[string]string)
    for k, v := range r.Header {
        headers[k] = strings.Join(v, ", ")
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(headers)
})
```

**Expected output (client side):**
```
GET /echo?msg=hello  → hello
POST /echo           → {"message":"test","length":4}
GET /headers         → {"User-Agent":"Go-http-client/1.1",...}
```

**Checkpoint:** Server logs all three requests; Ctrl-C shuts down cleanly with "server stopped" log message; client exits 0.

---

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

## Official Documentation

- [`net/http`](https://pkg.go.dev/net/http) — `HandleFunc`, `HandlerFunc`, `Handler`, `ServeMux`, `Server`, `Client`, `ListenAndServe`, `ResponseWriter`, `Request`, `StatusOK`
- [`encoding/json`](https://pkg.go.dev/encoding/json) — `NewEncoder`, `NewDecoder` for JSON responses and client decoding
- [`context`](https://pkg.go.dev/context) — `WithTimeout`, `Background` used in graceful shutdown and client timeouts
- [`os`](https://pkg.go.dev/os) — `Signal` for graceful shutdown
- [`log`](https://pkg.go.dev/log) — `Fatal`, `Printf` for server-side logging
- [`time`](https://pkg.go.dev/time) — `Second`, `Since` for timeouts and latency measurement
- [Go Blog: HTTP/2 Server Push](https://go.dev/blog/h2push)
- [Go Blog: The Go net/http Package](https://go.dev/blog/http-tracing)
- [Language Spec — Method sets](https://go.dev/ref/spec#Method_sets) — how `ServeHTTP` satisfies `http.Handler`
