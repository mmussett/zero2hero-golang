# Day 17: Sync Primitives

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

## Labs

### Lab 1: Race Condition — Bare Counter

**What you'll practise:** Witnessing a data race that the race detector catches.

**Task:**
Increment a shared `int` counter from 100 goroutines without any synchronisation. Run with `go test -race` and observe the race report. Note the exact lines flagged.

**Steps:**
1. Declare `var counter int` at package level (or closure-captured)
2. Launch 100 goroutines each doing `counter++` 1 000 times
3. Wait for all with `sync.WaitGroup`, print `counter`
4. Run `go run -race main.go` and read the DATA RACE report

```go
var counter int
var wg sync.WaitGroup

for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for j := 0; j < 1000; j++ {
            counter++ // RACE
        }
    }()
}
wg.Wait()
fmt.Println(counter) // likely < 100000
```

**Expected output:**
```
==================
WARNING: DATA RACE
Write at 0x... by goroutine 7:
  main.main.func1()
...
```

**Checkpoint:** The race detector reports at least one DATA RACE on `counter`.

---

### Lab 2: sync.Mutex — Fix the Race

**What you'll practise:** Protecting shared state with `sync.Mutex` to eliminate data races.

**Task:**
Take the racy counter from Lab 1 and protect it with a `sync.Mutex`. Verify with `go run -race` that the race is gone and the final count is exactly 100 000.

**Steps:**
1. Create `type SafeCounter struct { mu sync.Mutex; n int }`
2. Add `Inc()` and `Value()` methods, each using `mu.Lock() / defer mu.Unlock()`
3. Replace bare `counter++` with `sc.Inc()`
4. Run with `-race` — confirm no races; confirm `sc.Value() == 100000`

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

**Expected output:**
```
100000
(no DATA RACE warning)
```

**Checkpoint:** `go run -race main.go` exits 0 with final counter exactly 100 000.

---

### Lab 3: sync.RWMutex — Read-Heavy Cache

**What you'll practise:** Using `RWMutex` to allow concurrent reads while serialising writes.

**Task:**
Build an in-memory string cache backed by a `map[string]string`. Start 10 reader goroutines and 1 writer goroutine. Show that `RWMutex` outperforms a plain `Mutex` by running both versions with `go test -bench`.

**Steps:**
1. Implement `type Cache struct { mu sync.RWMutex; m map[string]string }`
2. Add `Get(key string) (string, bool)` using `RLock/RUnlock`
3. Add `Set(key, value string)` using `Lock/Unlock`
4. Write `BenchmarkCacheRWMutex` (10 readers, 1 writer goroutine) and `BenchmarkCacheMutex` (same but `sync.Mutex`)

```go
func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.m[key]
    return v, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.m[key] = value
}
```

**Expected output:**
```
BenchmarkCacheRWMutex-8    5000000    240 ns/op
BenchmarkCacheMutex-8      2000000    610 ns/op
```

**Checkpoint:** RWMutex benchmark shows lower ns/op than Mutex benchmark when reads dominate.

---

### Lab 4: sync.WaitGroup — 20 Concurrent Workers

**What you'll practise:** Coordinating a pool of goroutines to completion with `WaitGroup`.

**Task:**
Launch 20 goroutines that each sleep a random duration (0–100ms) then log their ID and elapsed time. Use `sync.WaitGroup` to wait for all. Compare with a channel-based coordination approach.

**Steps:**
1. `var wg sync.WaitGroup; wg.Add(20)`
2. Launch 20 goroutines: `defer wg.Done()`, sleep `rand.Intn(100)` ms, print ID and duration
3. Call `wg.Wait()` and print "all workers done"
4. Rewrite using a `done := make(chan struct{})` approach and compare code clarity

```go
var wg sync.WaitGroup
for i := 0; i < 20; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        d := time.Duration(rand.Intn(100)) * time.Millisecond
        time.Sleep(d)
        fmt.Printf("worker %d done after %v\n", id, d)
    }(i)
}
wg.Wait()
fmt.Println("all workers done")
```

**Expected output:**
```
worker 7 done after 12ms
worker 3 done after 33ms
...
all workers done
```

**Checkpoint:** "all workers done" always appears last, after all 20 worker lines.

---

### Lab 5: sync.Once — Singleton Initialisation

**What you'll practise:** Guaranteeing exactly-once initialisation under concurrent access.

**Task:**
Simulate a database connection pool. Use `sync.Once` so the `connect()` function is called exactly once regardless of how many goroutines call `GetDB()` concurrently. Add a counter inside `connect()` to prove it runs once.

**Steps:**
1. Declare `var (instance *DB; once sync.Once; initCount int32)`
2. In `connect()`, increment `atomic.AddInt32(&initCount, 1)` and return a new DB
3. `GetDB()` calls `once.Do(func() { instance = connect() })` then returns `instance`
4. Launch 100 goroutines all calling `GetDB()`, print `initCount` — must be 1

```go
var (
    instance  *DB
    once      sync.Once
    initCount int32
)

func connect() *DB {
    atomic.AddInt32(&initCount, 1)
    fmt.Println("connecting to DB...")
    return &DB{}
}

func GetDB() *DB {
    once.Do(func() { instance = connect() })
    return instance
}
```

**Expected output:**
```
connecting to DB...
initCount: 1
all 100 goroutines got the same DB: true
```

**Checkpoint:** `initCount` is 1 and all 100 goroutines receive the same pointer value.

---

### Lab 6: sync/atomic — Atomic vs Mutex Benchmark

**What you'll practise:** Comparing lock-free atomic operations with mutex-protected increments.

**Task:**
Implement two counters: one using `atomic.AddInt64` and one using `sync.Mutex`. Benchmark both with 8 goroutines each incrementing 1 000 000 times. Observe which is faster and why.

**Steps:**
1. `BenchmarkAtomicCounter` — use `atomic.AddInt64(&n, 1)` in a tight loop
2. `BenchmarkMutexCounter` — use `mu.Lock(); n++; mu.Unlock()` in a tight loop
3. Run `go test -bench=. -benchmem -cpu=8 ./...`
4. Note that atomic is faster but limited to simple integer operations

```go
var atomicN int64

func BenchmarkAtomicCounter(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            atomic.AddInt64(&atomicN, 1)
        }
    })
}

var mu sync.Mutex
var mutexN int64

func BenchmarkMutexCounter(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            mu.Lock()
            mutexN++
            mu.Unlock()
        }
    })
}
```

**Expected output:**
```
BenchmarkAtomicCounter-8    200000000    6.0 ns/op
BenchmarkMutexCounter-8      50000000   28.0 ns/op
```

**Checkpoint:** Atomic benchmark shows significantly lower ns/op than the mutex benchmark.

---

### Lab 7: sync.Map — Concurrent Word Counter

**What you'll practise:** Using `sync.Map` for concurrent map writes without a custom mutex.

**Task:**
Count word frequencies from 10 goroutines each processing a slice of text. Use `sync.Map` — specifically `LoadOrStore` and `Store` — to accumulate counts. Compare the code to the `sync.RWMutex` approach from Lab 3.

**Steps:**
1. Split a large text into 10 chunks, one per goroutine
2. Each goroutine tokenises its chunk and increments counts in a `sync.Map`
3. Use `sync.Map.Range` to iterate and print the top-10 words by count
4. Note: `sync.Map` lacks atomic increment — you need a `LoadOrStore` + compare-and-swap pattern

```go
var freq sync.Map

var wg sync.WaitGroup
for _, chunk := range chunks {
    wg.Add(1)
    go func(words []string) {
        defer wg.Done()
        for _, w := range words {
            actual, _ := freq.LoadOrStore(w, new(int64))
            atomic.AddInt64(actual.(*int64), 1)
        }
    }(chunk)
}
wg.Wait()

freq.Range(func(k, v any) bool {
    fmt.Printf("%s: %d\n", k, atomic.LoadInt64(v.(*int64)))
    return true
})
```

**Expected output:**
```
the: 142
a: 97
and: 84
...
```

**Checkpoint:** Every word appears exactly once in the output with an accurate count.

---

### Final Lab (Project): Thread-Safe LRU Cache

**What you'll practise:** Combining `sync.RWMutex`, `container/list`, and `sync/atomic` into a production-quality concurrent data structure.

**Task:**
Implement a thread-safe LRU (Least Recently Used) cache with O(1) Get and Put operations.

**Steps:**
1. Implement `New(capacity int) *LRUCache`
2. `Get(key string) (any, bool)` — acquire `RLock`, look up in map, move entry to front of list
3. `Put(key string, value any)` — acquire `Lock`, insert at front, evict LRU entry (back of list) when at capacity
4. Track hits and misses with `atomic.AddInt64`
5. Write a concurrent test: 10 goroutines doing random Gets and Puts simultaneously, verified with `go test -race`

```go
type LRUCache struct {
    mu       sync.RWMutex
    capacity int
    list     *list.List
    items    map[string]*list.Element
    hits     int64
    misses   int64
}

func (c *LRUCache) Get(key string) (any, bool) {
    c.mu.Lock() // need write lock to move element
    defer c.mu.Unlock()
    if el, ok := c.items[key]; ok {
        c.list.MoveToFront(el)
        atomic.AddInt64(&c.hits, 1)
        return el.Value.(*entry).value, true
    }
    atomic.AddInt64(&c.misses, 1)
    return nil, false
}
```

**Expected output:**
```
put A=1, B=2, C=3 (capacity 2 — A evicted)
get B: 2 (hit)
get A: not found (miss)
hits: 1  misses: 1
```

**Checkpoint:** `go test -race ./...` passes with no races. Get returns the correct value after a sequence of Puts that exceed capacity.

---

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
