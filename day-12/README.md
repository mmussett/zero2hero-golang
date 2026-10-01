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

## Labs

### Lab 1: Closure Basics — Counter

**What you'll practise:** closures that capture and mutate private state across multiple calls.

**Task:**
Write `makeCounter(start int) func() int` — a function that returns a closure. Each call to the returned function increments and returns its private counter. Create two independent counters and confirm they have separate state.

**Steps:**
1. Write `makeCounter(start int) func() int` — capture `n := start` in a closure
2. Return a closure that increments `n` and returns it
3. In `main`, create `counter1 := makeCounter(0)` and `counter2 := makeCounter(10)`
4. Call each 3 times and print results to confirm independence

```go
func makeCounter(start int) func() int {
    n := start
    return func() int {
        n++
        return n
    }
}

func main() {
    c1 := makeCounter(0)
    c2 := makeCounter(10)
    fmt.Println(c1(), c1(), c1()) // 1 2 3
    fmt.Println(c2(), c2(), c2()) // 11 12 13
    fmt.Println(c1())             // 4  (state is independent)
}
```

**Expected output:**
```
1 2 3
11 12 13
4
```

**Checkpoint:** `c1` and `c2` maintain completely separate internal counters; calling one does not affect the other.

---

### Lab 2: Memoization

**What you'll practise:** using a closure over a map to cache expensive function results.

**Task:**
Write `memoize(f func(int) int) func(int) int` that wraps `f` in a cache. Apply it to a naive recursive fibonacci to confirm calls are cached and the cached version is dramatically faster.

**Steps:**
1. Implement `memoize` capturing a `cache map[int]int` in the closure
2. On each call, check the cache first; if missing, compute and store
3. Write a naive `fib(n int) int` (recursive, no cache)
4. Apply `memoize` to it and compare `fib(40)` timing with and without

```go
func memoize(f func(int) int) func(int) int {
    cache := make(map[int]int)
    return func(n int) int {
        if v, ok := cache[n]; ok {
            return v
        }
        v := f(n)
        cache[n] = v
        return v
    }
}

var memoFib func(int) int
memoFib = memoize(func(n int) int {
    if n <= 1 { return n }
    return memoFib(n-1) + memoFib(n-2)
})
```

**Expected output:**
```
fib(10) = 55
fib(40) = 102334155
memoized fib(40) in < 1ms; naive in seconds
```

**Checkpoint:** `memoFib(40)` returns the correct value and completes in under 1 ms. The cache is private to the memoized function.

---

### Lab 3: Pipeline Pattern

**What you'll practise:** using functions as first-class values to compose reusable data-transformation stages.

**Task:**
Define `type Stage func([]string) []string` and implement four stages: `Lowercase`, `TrimSpaces`, `RemoveEmpty`, and `Deduplicate`. Write a `Pipeline` function that applies stages in sequence.

**Steps:**
1. Define the `Stage` type
2. Implement `Lowercase` using `strings.ToLower`, `TrimSpaces` using `strings.TrimSpace`, `RemoveEmpty` filtering blank strings, and `Deduplicate` preserving first occurrence
3. Implement `Pipeline(data []string, stages ...Stage) []string`
4. Test with a messy input slice

```go
type Stage func([]string) []string

func Pipeline(data []string, stages ...Stage) []string {
    for _, s := range stages {
        data = s(data)
    }
    return data
}

func Lowercase(in []string) []string {
    out := make([]string, len(in))
    for i, s := range in { out[i] = strings.ToLower(s) }
    return out
}

func RemoveEmpty(in []string) []string {
    var out []string
    for _, s := range in {
        if s != "" { out = append(out, s) }
    }
    return out
}
```

**Expected output:**
```
Input:  ["  Go ", "RUST", "", "go", "  "]
Output: [go rust]
```

**Checkpoint:** The pipeline is composable — any subset of stages can be passed in any order.

---

### Lab 4: Functional Options

**What you'll practise:** the functional options pattern — building flexible constructors without telescoping parameters.

**Task:**
Build a `Server` struct with fields `host string`, `port int`, `timeout time.Duration`, and `maxConns int`. Write `WithPort`, `WithTimeout`, `WithMaxConns` option functions. Implement `NewServer(host string, opts ...Option) *Server`.

**Steps:**
1. Define `type Option func(*Server)`
2. Implement `WithPort(p int) Option`, `WithTimeout(d time.Duration) Option`, `WithMaxConns(n int) Option` — each returns a closure that modifies a `*Server`
3. Set sensible defaults in `NewServer` before applying options
4. Create three servers with different option combinations and print their configs

```go
type Server struct {
    host     string
    port     int
    timeout  time.Duration
    maxConns int
}

type Option func(*Server)

func WithPort(p int) Option              { return func(s *Server) { s.port = p } }
func WithTimeout(d time.Duration) Option { return func(s *Server) { s.timeout = d } }
func WithMaxConns(n int) Option          { return func(s *Server) { s.maxConns = n } }

func NewServer(host string, opts ...Option) *Server {
    s := &Server{host: host, port: 8080, timeout: 30 * time.Second, maxConns: 100}
    for _, opt := range opts { opt(s) }
    return s
}
```

**Expected output:**
```
default: localhost:8080 timeout=30s maxConns=100
custom:  localhost:9090 timeout=60s maxConns=50
```

**Checkpoint:** Adding a new option does not change the `NewServer` signature. Calling `NewServer(host)` with no options returns defaults.

---

### Lab 5: Generator — `Range`

**What you'll practise:** implementing a lazy sequence generator using a closure to yield values on demand.

**Task:**
Write `Range(from, to int) func() (int, bool)` — a generator that returns successive integers from `from` to `to` inclusive. Each call to the returned function yields the next value and `true`; after exhaustion it returns `0, false`.

**Steps:**
1. Implement `Range` capturing `current := from` in a closure
2. Each invocation: if `current > to`, return `0, false`; otherwise return `current, true` and increment
3. In `main`, drive the generator with a `for` loop using the comma-ok idiom
4. Compose two generators: use one to generate indices into a string slice

```go
func Range(from, to int) func() (int, bool) {
    current := from
    return func() (int, bool) {
        if current > to {
            return 0, false
        }
        v := current
        current++
        return v, true
    }
}

gen := Range(1, 5)
for v, ok := gen(); ok; v, ok = gen() {
    fmt.Println(v)
}
```

**Expected output:**
```
1
2
3
4
5
```

**Checkpoint:** The generator yields exactly `to - from + 1` values and then consistently returns `0, false`.

---

### Lab 6: Lazy Initialisation with `sync.Once`

**What you'll practise:** combining closures with `sync.Once` to build a thread-safe lazy value that is computed at most once.

**Task:**
Implement `type Lazy[T any] struct` with a `Get() T` method that calls a user-supplied `compute func() T` exactly once, caching the result for all subsequent calls.

**Steps:**
1. Define `type Lazy[T any] struct { once sync.Once; value T; compute func() T }`
2. Implement `NewLazy[T any](f func() T) *Lazy[T]` storing `f` in the struct
3. Implement `Get() T` calling `l.once.Do(func() { l.value = l.compute() })` then returning `l.value`
4. Verify with a compute function that prints a message — confirm it prints exactly once even with multiple `Get` calls

```go
import "sync"

type Lazy[T any] struct {
    once    sync.Once
    value   T
    compute func() T
}

func NewLazy[T any](f func() T) *Lazy[T] {
    return &Lazy[T]{compute: f}
}

func (l *Lazy[T]) Get() T {
    l.once.Do(func() { l.value = l.compute() })
    return l.value
}

config := NewLazy(func() string {
    fmt.Println("computing config...") // prints once
    return "host=localhost port=8080"
})
fmt.Println(config.Get())
fmt.Println(config.Get()) // no "computing config..." again
```

**Expected output:**
```
computing config...
host=localhost port=8080
host=localhost port=8080
```

**Checkpoint:** The compute function is called exactly once regardless of how many times `Get` is called, including concurrent calls.

---

### Final Lab (Project): Data Pipeline

**What you'll practise:** combining closures, higher-order functions, the pipeline pattern, and functional options into a production-quality data processing component.

**Task:**
Build a pipeline that processes a slice of strings through composable stages. Add functional options for pipeline configuration.

**Steps:**
1. Define `type Stage func([]string) []string` and `Pipeline(data []string, stages ...Stage) []string`
2. Implement stages: `Lowercase`, `RemoveEmpty`, `Deduplicate`, `TrimSpaces`, `FilterMinLength(n int) Stage`
3. Use the functional options pattern for a `PipelineConfig` that controls whether to log each stage's output size
4. Wire all labs together: use `makeCounter` to number processed batches, `memoize` to cache expensive stage results

```go
type Stage func([]string) []string

func Pipeline(data []string, stages ...Stage) []string {
    for _, s := range stages {
        data = s(data)
    }
    return data
}

// FilterMinLength returns a Stage — a closure over n
func FilterMinLength(n int) Stage {
    return func(in []string) []string {
        var out []string
        for _, s := range in {
            if len(s) >= n {
                out = append(out, s)
            }
        }
        return out
    }
}

result := Pipeline(
    []string{"  Go ", "RUST", "", "go", "PYTHON", "  "},
    TrimSpaces,
    Lowercase,
    RemoveEmpty,
    Deduplicate,
    FilterMinLength(3),
)
fmt.Println(result)
```

**Expected output:**
```
[rust python]
```

**Checkpoint:** `go test ./...` passes; every stage is independently testable; adding or removing stages from the `Pipeline` call requires no other changes.

**Extension ideas:** make `Stage` operate on `chan string` for streaming; add error propagation.

## Official Documentation

- [`strings`](https://pkg.go.dev/strings) — `ToUpper` and other functions used in pipeline stages
- [`fmt`](https://pkg.go.dev/fmt) — formatted output
- [Language Spec: Function literals](https://go.dev/ref/spec#Function_literals) — closure syntax
- [Language Spec: Variadic functions](https://go.dev/ref/spec#Passing_arguments_to_..._parameters) — `...Option` variadic parameters
- [Effective Go: Functions](https://go.dev/doc/effective_go#functions) — first-class functions
- [Go Blog: Functional options for friendly APIs](https://go.dev/blog/functional-options-for-friendly-apis) — functional options pattern (Dave Cheney)
- [Go Tour: Closures](https://go.dev/tour/moretypes/25) — interactive closure tour
