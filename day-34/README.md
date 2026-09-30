# Day 34: Goroutines, Concurrency, and Synchronisation

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
