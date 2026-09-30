# Day 12: Closures and Higher-Order Functions

## Functions Are First-Class Values

In Go, functions are values. They can be assigned to variables, passed as arguments, and returned from other functions.

```go
add := func(a, b int) int { return a + b }
fmt.Println(add(2, 3)) // 5

func apply(f func(int) int, x int) int { return f(x) }
```

## Closures

A closure is a function that captures variables from its enclosing scope:

```go
func counter(start int) func() int {
    n := start
    return func() int {
        n++
        return n
    }
}

next := counter(0)
fmt.Println(next()) // 1
fmt.Println(next()) // 2
```

The captured variable `n` is shared between the closure and the outer function — mutations are visible to both.

## Higher-Order Functions

```go
func Map[T, U any](s []T, f func(T) U) []U { ... }
func Filter[T any](s []T, keep func(T) bool) []T { ... }
func Reduce[T, U any](s []T, init U, f func(U, T) U) U { ... }

// Pipeline
words := []string{"hello", "world", "go", "generics"}
result := Filter(words, func(w string) bool { return len(w) > 3 })
upper  := Map(result, strings.ToUpper)
```

## Functional Options Pattern

A clean way to handle optional configuration without ever-growing constructor arguments:

```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

type Option func(*Server)

func WithPort(p int) Option              { return func(s *Server) { s.port = p } }
func WithTimeout(d time.Duration) Option { return func(s *Server) { s.timeout = d } }

func NewServer(host string, opts ...Option) *Server {
    s := &Server{host: host, port: 8080, timeout: 30 * time.Second}
    for _, opt := range opts {
        opt(s)
    }
    return s
}

srv := NewServer("localhost", WithPort(9090), WithTimeout(60*time.Second))
```

## Memoisation

```go
func Memoize[K comparable, V any](f func(K) V) func(K) V {
    cache := make(map[K]V)
    return func(k K) V {
        if v, ok := cache[k]; ok { return v }
        v := f(k); cache[k] = v; return v
    }
}
```

## Day Project: Data Pipeline

Build a pipeline that processes a slice of strings through composable stages:

```go
type Stage func([]string) []string

func Pipeline(data []string, stages ...Stage) []string {
    for _, s := range stages { data = s(data) }
    return data
}
```

Implement stages: `Lowercase`, `RemoveEmpty`, `Deduplicate`, `TrimSpaces`, `FilterMinLength(n int) Stage`.

Use the functional options pattern for a `PipelineConfig` that controls parallelism.

**Extension ideas:** make `Stage` operate on `chan string` for streaming; add error propagation.
