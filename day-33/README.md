# Day 33: Goroutines, Concurrency, and Synchronisation

## Core Concept: The Go Concurrency Model

Go's concurrency is built on three pillars:

1. **Goroutines** — lightweight, cooperatively-scheduled units of execution
2. **Channels** — typed pipes for safe communication (covered in Day 33)
3. **sync** — low-level primitives for shared-memory synchronisation

When goroutines share data via channels, no explicit locking is needed. When goroutines share data via memory, locking is required.

---

## Part 1: Goroutines in Depth

### What Is a Goroutine?

A goroutine is not a thread. The Go runtime multiplexes goroutines onto OS threads using an M:N scheduler (M goroutines on N threads). You can run millions of goroutines; each starts with a 2–8 KB stack that grows and shrinks automatically.

```go
go func() {
    // This runs concurrently with the caller
    fmt.Println("I'm concurrent")
}()
```

The `go` keyword returns immediately — the calling goroutine does not wait.

### Goroutine Lifecycle

A goroutine runs until:
- Its function returns
- It calls `runtime.Goexit()`
- The program exits (all goroutines are killed)

There is no way to kill a goroutine from outside it. Design goroutines to respond to cancellation signals (context or done channels).

### Goroutine Scheduling

Go uses a work-stealing scheduler (GOMAXPROCS controls the number of OS threads that can run Go code simultaneously — default: number of CPUs):

```go
runtime.GOMAXPROCS(4)          // use 4 OS threads
fmt.Println(runtime.NumCPU()) // available CPUs
fmt.Println(runtime.NumGoroutine()) // currently running goroutines
```

Goroutines yield control at:
- Channel operations
- `time.Sleep`
- Syscalls
- `runtime.Gosched()` (explicit yield)
- Function calls (in most cases)

### Stack Growth

Goroutine stacks start small and grow via segmented or contiguous stack copying. This means passing large values on the stack is safe — the runtime handles growth. However, very deep recursion can still exhaust stack (default limit: 1 GB).

---

## Part 2: The Race Condition

A race condition occurs when two goroutines access the same memory concurrently and at least one access is a write, without synchronisation.

```go
// DATA RACE — do not do this
var counter int
for i := 0; i < 1000; i++ {
    go func() { counter++ }()  // concurrent unsynchronised writes
}
```

The result is undefined — you might get 1000, you might get 500, you might get anything.

### Detecting Races

```bash
go test -race ./...   # run tests with race detector
go run -race main.go  # run a program with race detector
```

The race detector adds ~2× CPU and memory overhead — use it in CI, not production. It reports the exact goroutines, files, and line numbers involved.

---

## Part 3: sync.Mutex and sync.RWMutex

### Mutex — Mutual Exclusion

```go
type SafeMap struct {
    mu sync.Mutex
    m  map[string]int
}

func (s *SafeMap) Set(key string, val int) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.m[key] = val
}

func (s *SafeMap) Get(key string) (int, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    v, ok := s.m[key]
    return v, ok
}
```

**Rules:**
- Always `defer mu.Unlock()` immediately after `Lock()`
- Never copy a Mutex (use pointer receivers or embed by value in a non-copied struct)
- Mutexes are not reentrant — calling `Lock()` from a goroutine that already holds the lock deadlocks

### RWMutex — Reader-Writer Lock

Allows many concurrent readers **or** one writer:

```go
type Cache struct {
    mu    sync.RWMutex
    store map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()           // multiple goroutines can hold RLock simultaneously
    defer c.mu.RUnlock()
    v, ok := c.store[key]
    return v, ok
}

func (c *Cache) Set(key, val string) {
    c.mu.Lock()            // exclusive — no readers or writers during write
    defer c.mu.Unlock()
    c.store[key] = val
}
```

Use `RWMutex` when reads are much more frequent than writes. For balanced read/write, plain `Mutex` is often faster (less overhead).

---

## Part 4: sync.WaitGroup — Barrier Synchronisation

Wait for a collection of goroutines to complete:

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)                   // increment BEFORE launching goroutine
    go func(id int) {
        defer wg.Done()         // decrement when goroutine exits
        fmt.Printf("worker %d done\n", id)
    }(i)
}

wg.Wait()                       // block until counter reaches 0
fmt.Println("all workers finished")
```

**Common mistake:** calling `wg.Add(1)` inside the goroutine — a race between `Add` and `Wait` may cause `Wait` to return before all goroutines are counted.

### WaitGroup with Error Collection

```go
var (
    wg   sync.WaitGroup
    mu   sync.Mutex
    errs []error
)

for _, item := range items {
    wg.Add(1)
    go func(it Item) {
        defer wg.Done()
        if err := process(it); err != nil {
            mu.Lock()
            errs = append(errs, err)
            mu.Unlock()
        }
    }(item)
}
wg.Wait()
return errors.Join(errs...)
```

Or use `golang.org/x/sync/errgroup` for a cleaner pattern.

---

## Part 5: sync.Once — Lazy Initialisation

Execute a function exactly once, regardless of how many goroutines call it:

```go
var (
    db   *sql.DB
    once sync.Once
)

func GetDB() *sql.DB {
    once.Do(func() {
        var err error
        db, err = sql.Open("sqlite", "app.db")
        if err != nil {
            panic(err)
        }
    })
    return db
}
```

`sync.Once` is safe for concurrent use. The function passed to `Do` is called only once; subsequent calls to `Do` are no-ops.

Note: `Once` does not reset. If the initialiser panics, `Do` considers it "done" — subsequent calls are still no-ops. Handle panics inside `Do` if reinitialisation is needed.

---

## Part 6: sync/atomic — Lock-Free Primitives

For simple integers and booleans, atomic operations avoid the overhead of a mutex:

```go
import "sync/atomic"

var count int64

atomic.AddInt64(&count, 1)               // increment
atomic.AddInt64(&count, -1)              // decrement
n := atomic.LoadInt64(&count)            // read
atomic.StoreInt64(&count, 0)             // write
swapped := atomic.CompareAndSwapInt64(&count, old, new) // CAS

// Go 1.19+: atomic.Value for arbitrary types
var v atomic.Value
v.Store(map[string]int{"a": 1})
m := v.Load().(map[string]int)
```

Atomic operations are faster than mutexes for single-variable access but cannot protect multi-variable invariants. Always use mutexes when updating two or more variables that must stay consistent.

---

## Part 7: sync.Cond — Condition Variables

`sync.Cond` allows goroutines to wait for a condition to become true:

```go
var mu sync.Mutex
cond := sync.NewCond(&mu)

// Waiter
go func() {
    mu.Lock()
    for !conditionMet() {
        cond.Wait()  // atomically releases mu and suspends
    }
    // condition is true, mu is held
    mu.Unlock()
}()

// Signaller
mu.Lock()
setCondition()
cond.Signal()   // wake one waiter
// or cond.Broadcast() to wake all
mu.Unlock()
```

`cond.Wait()` must always be called inside a `for` loop, not an `if`, because spurious wakeups are possible.

`sync.Cond` is rarely the best choice in modern Go — channels usually express the intent more clearly. Use it when you need to broadcast to many waiters efficiently.

---

## Part 8: sync.Map — Concurrent Map

Prefer a `map` + `sync.RWMutex` for most cases. Use `sync.Map` specifically when:
- Many goroutines read and write disjoint sets of keys (e.g., per-user caches)
- Keys are written once and read many times (stable map)

```go
var m sync.Map

m.Store("key", 42)

if v, ok := m.Load("key"); ok {
    fmt.Println(v.(int))
}

m.LoadOrStore("key", 0)  // set only if not already present

m.Range(func(k, v any) bool {
    fmt.Println(k, v)
    return true  // return false to stop iteration
})

m.Delete("key")
```

---

## Part 9: errgroup — Goroutines with Error Propagation

`golang.org/x/sync/errgroup` combines WaitGroup with error collection and optional context cancellation:

```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(context.Background())

for _, url := range urls {
    url := url  // capture loop variable
    g.Go(func() error {
        return download(ctx, url)
    })
}

if err := g.Wait(); err != nil {
    log.Fatal(err)  // first non-nil error
}
```

When any goroutine returns an error, the context is cancelled — other goroutines should check `ctx.Done()` and exit early.

---

## Part 10: Deadlocks and Livelocks

### Deadlock

Two or more goroutines waiting for each other, permanently blocked:

```go
var mu1, mu2 sync.Mutex

// Goroutine A
mu1.Lock(); mu2.Lock()  // acquires mu1, waits for mu2

// Goroutine B (running concurrently)
mu2.Lock(); mu1.Lock()  // acquires mu2, waits for mu1
```

Fix: always acquire locks in the same order across goroutines.

The Go runtime detects deadlocks where all goroutines are blocked and prints:
```
fatal error: all goroutines are asleep - deadlock!
```

### Livelock

Goroutines are not blocked but keep reacting to each other without making progress — like two people in a corridor stepping aside in the same direction. Rare in practice but detected by observing goroutine counts and CPU usage without forward progress.

---

## Part 11: Concurrency Patterns Summary

| Pattern | Mechanism | When to Use |
|---------|-----------|-------------|
| Worker pool | buffered channel + WaitGroup | Limit concurrency on CPU/IO tasks |
| Pipeline | chained channels | Transform data through stages |
| Fan-out | goroutines on shared input | Parallel independent work |
| Fan-in | merge to one channel | Collect results from parallel work |
| Semaphore | buffered channel | Cap concurrent resource use |
| Done channel | `chan struct{}` | Broadcast cancellation |
| Once | `sync.Once` | Lazy singleton initialisation |
| Barrier | `sync.WaitGroup` | Wait for N goroutines |
| Pub/sub | channel + goroutines | Event broadcasting |

---

---

## Labs

### Lab 1: Goroutine Lifecycle

**What you'll practise:** Observing goroutine count with `runtime.NumGoroutine()` before, during, and after launching many goroutines.

**Task:**
Launch 1000 goroutines, each sleeping briefly. Use `runtime.NumGoroutine()` to observe the count rise and then fall after `WaitGroup.Wait()`.

**Steps:**
1. Print `runtime.NumGoroutine()` before launching any goroutines
2. Launch 1000 goroutines via `go func() { defer wg.Done(); time.Sleep(50*time.Millisecond) }()`
3. Immediately after launching all, print the goroutine count mid-flight
4. Call `wg.Wait()` and print the count again — it should return to the baseline

```go
package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
)

func main() {
    fmt.Println("goroutines before:", runtime.NumGoroutine())

    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            time.Sleep(50 * time.Millisecond)
        }()
    }

    fmt.Println("goroutines during:", runtime.NumGoroutine())
    wg.Wait()
    fmt.Println("goroutines after: ", runtime.NumGoroutine())
}
```

**Expected output:**
```
goroutines before: 1
goroutines during: 1001  (approximately)
goroutines after:  1
```

**Checkpoint:** Try launching 100,000 goroutines. Does the program OOM? (It should not — goroutine stacks start at ~2 KB.) Use `runtime.ReadMemStats` to measure actual memory usage.

---

### Lab 2: Mutex Deep Dive

**What you'll practise:** Implementing a thread-safe counter using `sync.Mutex` and verifying it is race-free with the race detector.

**Task:**
Build a `SafeCounter` wrapping a `map[string]int`. Launch 100 goroutines each calling `Increment` 1000 times. Assert the final count is exactly 100,000.

**Steps:**
1. Define `SafeCounter` with `mu sync.Mutex` and `counts map[string]int`
2. Add `Increment(key string)` and `Value(key string) int` methods, both protected by the mutex
3. Launch 100 goroutines each calling `sc.Increment("hits")` 1000 times with a WaitGroup
4. After `wg.Wait()`, assert `sc.Value("hits") == 100_000`
5. Run `go test -race ./...` to confirm no data race is reported

```go
type SafeCounter struct {
    mu     sync.Mutex
    counts map[string]int
}

func NewSafeCounter() *SafeCounter {
    return &SafeCounter{counts: make(map[string]int)}
}

func (c *SafeCounter) Increment(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.counts[key]++
}

func (c *SafeCounter) Value(key string) int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.counts[key]
}
```

**Expected output:**
```
Final count: 100000
PASS
```

**Checkpoint:** Remove `c.mu.Lock()` and `c.mu.Unlock()` from both methods. Run `go test -race` — you should see a data race report with exact file and line numbers. Restore the locks and confirm the race disappears.

---

### Lab 3: RWMutex

**What you'll practise:** Using `sync.RWMutex` for a read-heavy workload and benchmarking it against a plain `Mutex`.

**Task:**
Implement a thread-safe `Cache` with `Get` and `Set`. Use `RWMutex` so multiple goroutines can read simultaneously. Benchmark 100 readers + 1 writer against a `Mutex`-only version.

**Steps:**
1. Implement `RWCache` with `sync.RWMutex`; `Get` uses `RLock/RUnlock`, `Set` uses `Lock/Unlock`
2. Implement `MutexCache` with plain `sync.Mutex` for comparison (identical API)
3. Write two benchmarks using `testing.B`: 100 goroutines calling `Get` in a loop while 1 writer calls `Set`
4. Run `go test -bench=. -benchmem ./...` and compare

```go
type RWCache struct {
    mu    sync.RWMutex
    store map[string]string
}

func (c *RWCache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.store[key]
    return v, ok
}

func (c *RWCache) Set(key, val string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.store[key] = val
}
```

**Expected output:**
```
BenchmarkRWCache-8      5000000    250 ns/op
BenchmarkMutexCache-8   2000000    680 ns/op
```

**Checkpoint:** Invert the ratio to 1 reader + 100 writers. Now `Mutex` should be faster or comparable. Explain in a comment: `RWMutex` has overhead and only pays off when concurrent readers significantly outnumber writers.

---

### Lab 4: WaitGroup Patterns — Manual errgroup

**What you'll practise:** Collecting the first error from N goroutines using WaitGroup, a mutex-protected error slot, and a done channel for cancellation.

**Task:**
Process 10 items concurrently. Item 3 always fails. Capture the first error, cancel remaining work via a done channel, and return cleanly after all goroutines exit.

**Steps:**
1. Create `done := make(chan struct{})` and a mutex-protected `var firstErr error`
2. Use `sync.Once` to close `done` and set `firstErr` at most once
3. Launch 10 goroutines; each checks `<-done` at startup (non-blocking select)
4. Goroutine 3 returns an error; the `Once` fires, sets the error, and closes `done`
5. After `wg.Wait()`, print the captured error

```go
var (
    once     sync.Once
    firstErr error
    mu       sync.Mutex
    wg       sync.WaitGroup
)
done := make(chan struct{})

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        select {
        case <-done:
            return // already cancelled
        default:
        }
        if err := process(id); err != nil {
            once.Do(func() {
                mu.Lock(); firstErr = err; mu.Unlock()
                close(done)
            })
        }
    }(i)
}
wg.Wait()
fmt.Println("error:", firstErr)
```

**Expected output:**
```
processing 0...done
processing 1...done
processing 2...done
processing 3...ERROR
error: item 3 failed
```

**Checkpoint:** Compare this implementation with `golang.org/x/sync/errgroup` — the errgroup version achieves the same in ~5 lines. Identify which parts of your manual implementation each errgroup method replaces.

---

### Lab 5: sync.Once — Safe Singleton

**What you'll practise:** Using `sync.Once` to guarantee initialisation runs exactly once under concurrent load.

**Task:**
Create a `Config` singleton. Launch 100 goroutines all calling `GetConfig()` simultaneously. Verify the initialiser runs exactly once using an atomic counter.

**Steps:**
1. Declare `var (cfg *Config; once sync.Once; initCount int64)`
2. `GetConfig()` calls `once.Do(func() { atomic.AddInt64(&initCount, 1); cfg = loadConfig() })`
3. Launch 100 goroutines each calling `GetConfig()` and storing the returned pointer
4. After all goroutines finish, assert `initCount == 1` and all 100 goroutines got the same pointer

```go
var (
    cfg       *Config
    once      sync.Once
    initCount int64
)

func GetConfig() *Config {
    once.Do(func() {
        atomic.AddInt64(&initCount, 1)
        cfg = &Config{DSN: "localhost:5432", Debug: true}
        fmt.Println("config: initialised")
    })
    return cfg
}
```

**Expected output:**
```
config: initialised      ← printed exactly once
All 100 goroutines got the same *Config
initCount: 1
```

**Checkpoint:** Add a 10ms sleep inside `once.Do`. Launch 100 goroutines. Confirm exactly one goroutine executes the sleep while the other 99 block — then all 100 receive the result simultaneously when `Do` returns.

---

### Lab 6: sync/atomic

**What you'll practise:** Using atomic operations for a lock-free counter and implementing a lock-free stack push with `CompareAndSwap`.

**Task:**
Part A — implement an atomic counter and benchmark it against a mutex counter. Part B — implement a lock-free stack `push` using `atomic.CompareAndSwapPointer`.

**Steps:**
1. `AtomicCounter`: use `atomic.AddInt64` for increment and `atomic.LoadInt64` for read
2. `MutexCounter`: identical API using `sync.Mutex`
3. Benchmark: 1000 goroutines each calling `Increment` 1000 times; compare ns/op
4. Part B: implement a linked-list stack node; `push` does a CAS loop on `head`

```go
type AtomicCounter struct{ n int64 }

func (c *AtomicCounter) Increment()     { atomic.AddInt64(&c.n, 1) }
func (c *AtomicCounter) Value() int64   { return atomic.LoadInt64(&c.n) }

// Lock-free stack (single-push demo)
type node struct {
    val  int
    next unsafe.Pointer
}
var head unsafe.Pointer

func push(val int) {
    n := &node{val: val}
    for {
        old := atomic.LoadPointer(&head)
        n.next = old
        if atomic.CompareAndSwapPointer(&head, old, unsafe.Pointer(n)) {
            return
        }
        // CAS failed: another goroutine modified head — retry
    }
}
```

**Expected output:**
```
BenchmarkAtomicCounter-8   10000000    120 ns/op
BenchmarkMutexCounter-8     3000000    450 ns/op
Lock-free stack (top→bottom): [5 4 3 2 1]
```

**Checkpoint:** Run `go test -race` on the lock-free stack. The race detector must not flag it — CAS is inherently safe. Then change `push` to use a plain pointer assignment (`head = unsafe.Pointer(n)`) without CAS and run `go test -race` again to see the reported race.

---

### Lab 7: sync.Map

**What you'll practise:** Using `sync.Map` with `LoadOrStore` for a "get or create" pattern and comparing it to a `RWMutex` map in a benchmark.

**Task:**
Build a connection pool where `GetConn(host string) *Conn` returns an existing connection or creates a new one. Under concurrent load, only one connection per host must be created.

**Steps:**
1. Define `type Conn struct{ Host string; Created time.Time }`
2. Use `var pool sync.Map`
3. `GetConn(host string) *Conn`: call `LoadOrStore` with a freshly allocated `*Conn`; if another goroutine won, discard your allocation and return theirs
4. Launch 50 goroutines all calling `GetConn("db.example.com")` simultaneously
5. Assert all 50 received the same `*Conn` pointer (same `Created` time)

```go
var pool sync.Map

func GetConn(host string) *Conn {
    candidate := &Conn{Host: host, Created: time.Now()}
    actual, _ := pool.LoadOrStore(host, candidate)
    return actual.(*Conn)
}
```

**Expected output:**
```
50 goroutines all got Conn{Host: "db.example.com", Created: 2024-...}
Unique connections created: 1
```

**Checkpoint:** Replace `sync.Map` with `map[string]*Conn` + `sync.RWMutex`. Run `go test -bench=. -benchmem`. For this "write-once, read-many" pattern, which implementation is faster and why?

---

### Lab 8: Semaphore Pattern

**What you'll practise:** Using a buffered channel as a counting semaphore to limit the number of goroutines running concurrently.

**Task:**
Simulate a download manager: 10 URLs to download, but only 3 simultaneous downloads allowed. Use `make(chan struct{}, 3)` as the semaphore.

**Steps:**
1. Create `sem := make(chan struct{}, 3)`
2. For each URL: send to `sem` to acquire a slot, launch the goroutine, release (`<-sem`) on exit with `defer`
3. Simulate the download with `time.Sleep(100*time.Millisecond)`
4. Print `runtime.NumGoroutine()` inside each download to confirm concurrency stays at or below 3 (plus overhead)

```go
sem := make(chan struct{}, 3)
var wg sync.WaitGroup

urls := []string{
    "url1", "url2", "url3", "url4", "url5",
    "url6", "url7", "url8", "url9", "url10",
}

for _, url := range urls {
    wg.Add(1)
    sem <- struct{}{} // acquire — blocks if 3 already in flight
    go func(u string) {
        defer wg.Done()
        defer func() { <-sem }() // release
        fmt.Printf("downloading %s (goroutines: %d)\n", u, runtime.NumGoroutine())
        time.Sleep(100 * time.Millisecond)
    }(url)
}
wg.Wait()
fmt.Println("all downloads complete")
```

**Expected output:**
```
downloading url1 (goroutines: 4)
downloading url2 (goroutines: 5)
downloading url3 (goroutines: 6)
downloading url4 (goroutines: 4)  ← waits for a slot to free
...
all downloads complete
```

**Checkpoint:** Remove the semaphore sends/receives. Confirm all 10 goroutines launch simultaneously (goroutine count spikes to 11). Restore the semaphore and confirm the spike is capped at ~4 (3 workers + main).

---

### Lab 9: Goroutine Leak Detection

**What you'll practise:** Deliberately leaking goroutines blocked on a channel send, detecting the leak with `runtime.NumGoroutine`, and fixing it with a done channel.

**Task:**
Write a `leaky()` function that starts a goroutine permanently blocked on a channel send. Call it 10 times and observe the goroutine count grow. Fix with a done channel and verify the count returns to baseline.

**Steps:**
1. Write `leaky()`: creates an unbuffered channel, starts a goroutine that sends to it — nobody reads it, so the goroutine blocks forever
2. Call `leaky()` 10 times; print `runtime.NumGoroutine()` — it grows by 10
3. Write `cancellable(done <-chan struct{}, ch chan<- int)`: goroutine uses `select` with both the send and `<-done`
4. Close `done` from main; print `runtime.NumGoroutine()` again — it returns to baseline

```go
// Leaky version — goroutine blocks on send forever
func leaky() {
    ch := make(chan int) // unbuffered, nobody reads it
    go func() {
        ch <- 42  // blocks forever — goroutine leaked!
    }()
}

// Fixed version — goroutine can exit when done is closed
func cancellable(done <-chan struct{}, ch chan<- int) {
    go func() {
        select {
        case ch <- 42:
        case <-done:
            fmt.Println("goroutine: cancelled cleanly")
        }
    }()
}
```

**Expected output:**
```
before leaks:              1
after 10 leaky() calls:   11  ← 10 leaked goroutines
--- Fix ---
before:                    1
after 10 cancellable():   11
after close(done):         1  ← all goroutines exited
```

**Checkpoint:** Add `defer goleak.VerifyNone(t)` from `github.com/uber-go/goleak` to a test function. It fails if any goroutine leaked — this is the production-grade technique. Alternatively, snapshot `runtime.NumGoroutine()` before and after a test and assert equality.

---

## Day Project: Concurrent Download Manager

Build a download manager that:

1. Accepts a list of URLs (hardcoded slice)
2. Downloads each concurrently, limited to `runtime.NumCPU()` simultaneous downloads (semaphore)
3. Tracks state with a thread-safe `DownloadManager` struct:
   - `sync.RWMutex`-protected map of `url → Status`
   - Atomic `totalBytes int64` counter
   - `sync.WaitGroup` to wait for all downloads
4. Reports progress every 500ms via a `time.Ticker` goroutine
5. Cancels all in-flight downloads after 10s via `context.WithTimeout`
6. Collects errors with `errgroup`

```go
type Status struct {
    State string    // "pending", "downloading", "done", "error"
    Bytes int64
    Err   error
}
```

Also write:
- A unit test that verifies no data races (`go test -race`)
- A benchmark comparing single-threaded vs parallel download simulation

**Extension ideas:** implement retry with exponential backoff using `sync.Once` to track attempts; add a `Pause`/`Resume` mechanism using a channel toggle.

## Official Documentation

- [`sync`](https://pkg.go.dev/sync) — `Mutex`, `RWMutex`, `WaitGroup`, `Once`, `Cond`, `Map`, `Pool`
- [`sync/atomic`](https://pkg.go.dev/sync/atomic) — `AddInt64`, `LoadInt64`, `StoreInt64`, `CompareAndSwapInt64`, `Value`
- [`runtime`](https://pkg.go.dev/runtime) — `GOMAXPROCS`, `NumCPU`, `NumGoroutine`, `Goexit`, `Gosched`
- [`context`](https://pkg.go.dev/context) — `WithTimeout`, `WithCancel`, `Background`, `Done` for goroutine cancellation
- [`database/sql`](https://pkg.go.dev/database/sql) — `DB`, `Open` used in `sync.Once` lazy-init example
- [`errors`](https://pkg.go.dev/errors) — `Join` for aggregating errors from goroutines
- [`time`](https://pkg.go.dev/time) — `Ticker`, `NewTicker`, `After` for progress reporting and timeout
- [golang.org/x/sync/errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup) — `WithContext`, `Go`, `Wait` for goroutine error propagation
- [Go Blog: Share Memory by Communicating](https://go.dev/blog/codelab-share)
- [Go Blog: Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [Language Spec — Go statements](https://go.dev/ref/spec#Go_statements) — goroutine semantics
- [Go Race Detector](https://go.dev/doc/articles/race_detector) — using `-race` in tests and CI
