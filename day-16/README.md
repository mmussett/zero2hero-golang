# Day 16: Sync Primitives

## Core Concept: Shared Memory Requires Explicit Synchronisation

Channels are great for ownership transfer and signalling. When goroutines genuinely need to share mutable state, reach for the [`sync`](https://pkg.go.dev/sync) package.

## [sync.Mutex](https://pkg.go.dev/sync#Mutex)

```go
type SafeCounter struct {
    mu sync.Mutex
    n  int
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.n++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.n
}
```

Always `defer mu.Unlock()` immediately after `Lock()` — even if the function panics, the mutex will be released.

## [sync.RWMutex](https://pkg.go.dev/sync#RWMutex)

Allows multiple concurrent readers or one writer — better throughput for read-heavy workloads:

```go
var mu sync.RWMutex

func read() {
    mu.RLock()
    defer mu.RUnlock()
    // concurrent reads are safe
}

func write() {
    mu.Lock()
    defer mu.Unlock()
    // exclusive write
}
```

## [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)

Wait for a collection of goroutines to finish:

```go
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        doWork(id)
    }(i)
}
wg.Wait()
```

Always call `Add` before launching the goroutine, never inside it.

## [sync.Once](https://pkg.go.dev/sync#Once)

Run an initialisation exactly once, regardless of how many goroutines call it:

```go
var (
    instance *DB
    once     sync.Once
)

func GetDB() *DB {
    once.Do(func() {
        instance = connect()
    })
    return instance
}
```

## [sync/atomic](https://pkg.go.dev/sync/atomic)

Low-level, lock-free operations on primitive integers and pointers:

```go
var hits int64
atomic.AddInt64(&hits, 1)
n := atomic.LoadInt64(&hits)
```

Use `atomic` only for simple counters and flags — prefer `Mutex` for anything more complex.

## Race Detector

Always run tests with `-race` during development:

```bash
go test -race ./...
go run -race main.go
```

The race detector catches concurrent access to unprotected shared data.

## Day Project: Thread-Safe LRU Cache

Implement an LRU (Least Recently Used) cache with:
- `New(capacity int) *LRUCache`
- `Get(key string) (any, bool)` — O(1), RLock
- `Put(key string, value any)` — O(1), Lock, evicts LRU entry when full
- Thread-safe using [`sync.RWMutex`](https://pkg.go.dev/sync#RWMutex)
- A hit/miss counter using [`sync/atomic`](https://pkg.go.dev/sync/atomic)

Use [`container/list`](https://pkg.go.dev/container/list) as the doubly linked list and a `map[string]*list.Element` for O(1) lookup.

**Extension ideas:** add a TTL per entry using a `time.Time` in the value; implement `sync.Map`-based variant and benchmark both.

## Official Documentation

- [`sync`](https://pkg.go.dev/sync) — Mutex, RWMutex, WaitGroup, Once, Map
- [`sync/atomic`](https://pkg.go.dev/sync/atomic) — AddInt64, LoadInt64, and other atomic operations
- [`container/list`](https://pkg.go.dev/container/list) — doubly linked list for LRU implementation
- [Language Spec: Go statements](https://go.dev/ref/spec#Go_statements) — goroutine semantics
- [Effective Go: Concurrency](https://go.dev/doc/effective_go#concurrency) — synchronisation patterns
- [Go Blog: The Go Memory Model](https://go.dev/ref/mem) — happens-before and synchronisation guarantees
- [Go Blog: Share Memory by Communicating](https://go.dev/blog/codelab-share) — when to use channels vs mutexes
