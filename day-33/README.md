# Day 33: The Go Channel Model — A Complete Guide

## Core Concept: Channels Are Typed, Directional Pipes

A channel is a goroutine-safe queue. It connects goroutines through communication rather than shared memory. Every channel has:

- A **type**: only values of that type can flow through it
- A **direction**: send-only (`chan<- T`), receive-only (`<-chan T`), or bidirectional (`chan T`)
- A **capacity**: 0 (unbuffered) or N (buffered)

> "Don't communicate by sharing memory; share memory by communicating."

---

## Part 1: Channel Fundamentals

### Creation

```go
ch  := make(chan int)      // unbuffered
ch  := make(chan int, 10)  // buffered, capacity 10
sch := make(chan<- string) // send-only (rarely created directly)
rch := make(<-chan string) // receive-only (rarely created directly)
```

### Send and Receive

```go
ch <- 42        // send — blocks if channel is full (or unbuffered and no receiver)
v := <-ch       // receive — blocks if channel is empty
v, ok := <-ch   // ok is false when channel is closed and empty
```

### Close

```go
close(ch)  // signal: no more values will be sent
```

Rules:
- Only the **sender** should close a channel
- Closing a nil channel panics
- Closing a closed channel panics
- Receiving from a closed channel returns the zero value immediately

```go
// Safe close pattern: use a done channel instead of closing a data channel
done := make(chan struct{})
close(done)  // broadcast: everyone waiting on <-done unblocks immediately
```

### Range Over Channel

```go
for v := range ch {
    fmt.Println(v)  // receives until ch is closed
}
```

This is the idiomatic way to consume all values from a channel.

---

## Part 2: Unbuffered vs Buffered Channels

### Unbuffered (Synchronous Rendezvous)

```go
ch := make(chan int)
```

- Send blocks until a receiver is ready
- Receive blocks until a sender is ready
- Guarantees the sender knows the receiver has the value

Use unbuffered channels for:
- Synchronisation ("signal me when done")
- Handoff where you need acknowledgement

### Buffered (Asynchronous Queue)

```go
ch := make(chan int, 10)
```

- Send blocks only when the buffer is full
- Receive blocks only when the buffer is empty
- Decouples sender speed from receiver speed

Use buffered channels for:
- Rate limiting (semaphore pattern)
- Batching work
- Absorbing bursts when producer is faster than consumer

### Semaphore Pattern (Limit Concurrency)

```go
sem := make(chan struct{}, runtime.NumCPU())

for _, file := range files {
    sem <- struct{}{}   // acquire
    go func(f string) {
        defer func() { <-sem }()  // release
        processFile(f)
    }(file)
}

// Wait for all goroutines to finish
for i := 0; i < cap(sem); i++ {
    sem <- struct{}{}
}
```

---

## Part 3: select — Multiplexing Channels

`select` waits on multiple channel operations simultaneously. It picks one ready case at random (when multiple are ready):

```go
select {
case v := <-ch1:
    fmt.Println("received from ch1:", v)
case ch2 <- x:
    fmt.Println("sent to ch2")
case <-time.After(1 * time.Second):
    fmt.Println("timeout")
default:
    fmt.Println("no channels ready — non-blocking")
}
```

### Non-Blocking Operations

`select` with a `default` case never blocks:

```go
func tryReceive(ch <-chan int) (int, bool) {
    select {
    case v := <-ch:
        return v, true
    default:
        return 0, false
    }
}
```

### Ticking and Timeout

```go
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()

timeout := time.After(10 * time.Second)

for {
    select {
    case t := <-ticker.C:
        fmt.Println("tick:", t)
    case <-timeout:
        fmt.Println("done")
        return
    }
}
```

---

## Part 4: Concurrency Patterns

### Pipeline

Each stage reads from one channel and writes to another:

```go
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums { out <- n }
        close(out)
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in { out <- n * n }
        close(out)
    }()
    return out
}

// Compose: generate → square → print
for v := range square(generate(2, 3, 4)) {
    fmt.Println(v) // 4, 9, 16
}
```

### Fan-Out

Distribute work from one channel to N workers:

```go
func fanOut(in <-chan string, n int, work func(string) string) []<-chan string {
    outs := make([]<-chan string, n)
    for i := range outs {
        out := make(chan string)
        outs[i] = out
        go func() {
            for v := range in {
                out <- work(v)
            }
            close(out)
        }()
    }
    return outs
}
```

### Fan-In (Merge)

Merge N channels into one:

```go
func fanIn(channels ...<-chan string) <-chan string {
    out := make(chan string)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan string) {
            defer wg.Done()
            for v := range c { out <- v }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}
```

### Done Channel (Cancellation)

Stop goroutines cleanly:

```go
func generate(done <-chan struct{}, nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            select {
            case out <- n:
            case <-done:
                return  // abandon remaining work
            }
        }
    }()
    return out
}

done := make(chan struct{})
defer close(done)   // cancels all goroutines using this done channel

out := generate(done, 1, 2, 3, 4, 5)
// Process first two, then cancel
fmt.Println(<-out)
fmt.Println(<-out)
// done closes → generate exits → out closes
```

Use `context.Context` in real code — it composes cancellation signals.

### Or-Done Channel

Unblock when either a channel has a value or cancellation fires:

```go
func orDone(done, ch <-chan struct{}) <-chan struct{} {
    out := make(chan struct{})
    go func() {
        defer close(out)
        select {
        case <-done:
        case <-ch:
            out <- struct{}{}
        }
    }()
    return out
}
```

---

## Part 5: Channel Direction in Function Signatures

Narrow channel types in function signatures to document intent and prevent bugs:

```go
// This function only sends — cannot accidentally receive
func produce(out chan<- int) {
    out <- 42
    // <-out  // compile error: cannot receive from send-only channel
}

// This function only receives — cannot accidentally send
func consume(in <-chan int) {
    v := <-in
    fmt.Println(v)
    // in <- 42  // compile error: cannot send to receive-only channel
}

// Bidirectional channels convert to directional automatically
ch := make(chan int)
produce(ch)   // chan int → chan<- int: fine
consume(ch)   // chan int → <-chan int: fine
```

---

## Part 6: Common Mistakes

### Goroutine Leak

Goroutines blocked on a channel operation that will never unblock:

```go
// Leak: goroutine is stuck forever if nobody receives from result
func leaky() <-chan int {
    result := make(chan int)
    go func() {
        result <- compute()  // blocks if nobody reads result
    }()
    return result
}
```

Fix: pass a `context.Context` or a done channel so the goroutine can exit when the caller abandons it.

### Closing from the Receiver

Receivers should not close channels — the sender may try to send after close, causing a panic:

```go
// Wrong: receiver closes
func wrong(ch chan int) {
    <-ch
    close(ch)  // sender may still be running!
}
```

Fix: use a `sync.Once` guard or a separate done channel for signalling.

### Sending on a Closed Channel

```go
close(ch)
ch <- 42  // panic: send on closed channel
```

Fix: ensure only the goroutine that "owns" the channel closes it, after all senders have finished (e.g. via `sync.WaitGroup`).

### Nil Channel Blocks Forever

```go
var ch chan int  // nil
ch <- 42        // blocks forever
<-ch            // blocks forever

// But in select, a nil channel case is skipped:
select {
case v := <-ch:  // never selected — ch is nil
case v := <-other:
    fmt.Println(v)
}
```

Nil channels are useful in `select` to dynamically disable cases.

---

## Part 7: Channel vs Mutex — When to Use Which

| Situation | Use |
|-----------|-----|
| Ownership transfer (producer → consumer) | Channel |
| Signalling (start, stop, done) | Channel (chan struct{}) |
| Parallel pipelines | Channel |
| Protecting shared mutable state | Mutex |
| Caching with concurrent readers | sync.RWMutex |
| Simple counter | sync/atomic |
| Rate limiting | Buffered channel |

Neither is universally better. Many programs need both.

---

---

## Labs

### Lab 1: Channel as a Pipe

**What you'll practise:** Sending and receiving values through an unbuffered channel between two goroutines.

**Task:**
Launch a sender goroutine that sends 5 integers on a channel. Receive them all in `main`. Print each value from both sides to visualise the synchronised handoff.

**Steps:**
1. Create an unbuffered `chan int`
2. Launch a goroutine that sends 1 through 5 and then closes the channel
3. In `main`, use `for v := range ch` to receive all values
4. Print `"sending: N"` from the goroutine and `"received: N"` from main

```go
func main() {
    ch := make(chan int)

    go func() {
        for i := 1; i <= 5; i++ {
            fmt.Println("sending:", i)
            ch <- i
        }
        close(ch)
    }()

    for v := range ch {
        fmt.Println("received:", v)
    }
}
```

**Expected output:**
```
sending: 1
received: 1
sending: 2
received: 2
sending: 3
received: 3
sending: 4
received: 4
sending: 5
received: 5
```

**Checkpoint:** Change to a buffered channel `make(chan int, 5)`. Observe how the output order changes — the sender bursts ahead before the receiver catches up. Explain the difference in a comment.

---

### Lab 2: Pipeline Stages

**What you'll practise:** Building a three-stage pipeline where each stage is an independent goroutine communicating via channels.

**Task:**
Implement `generate → square → print`. Each stage is a separate function that returns a `<-chan int`. Wire them together in `main`.

**Steps:**
1. Write `generate(nums ...int) <-chan int` — sends each number, then closes the channel
2. Write `square(in <-chan int) <-chan int` — receives, squares each value, sends
3. In `main`, compose: `for v := range square(generate(2, 3, 4, 5)) { fmt.Println(v) }`
4. Add a third stage `double(in <-chan int) <-chan int` and compose all three

```go
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}
```

**Expected output:**
```
# generate → square:
4 9 16 25

# generate → square → double:
8 18 32 50
```

**Checkpoint:** Add `fmt.Println("stage: square, processing", n)` inside `square`. Observe the interleaving with the final print — this confirms the stages run concurrently, not sequentially.

---

### Lab 3: Fan-Out

**What you'll practise:** Distributing work from one shared channel to N worker goroutines.

**Task:**
Send 20 jobs (integers 0–19) on a shared input channel. Launch 4 workers that each read from the same channel. Each worker prints which job it processed and its worker ID.

**Steps:**
1. Create `jobs := make(chan int, 20)` and send 0–19, then close it
2. Launch 4 workers with `go worker(id, jobs, &wg)` where each worker loops `for j := range jobs`
3. Use a `sync.WaitGroup` to wait for all workers to finish
4. Print `"worker N processed job M"` from each worker

```go
func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()
    for j := range jobs {
        fmt.Printf("worker %d processed job %d\n", id, j)
        time.Sleep(10 * time.Millisecond) // simulate work
    }
}

func main() {
    jobs := make(chan int, 20)
    var wg sync.WaitGroup

    for i := 0; i < 4; i++ {
        wg.Add(1)
        go worker(i, jobs, &wg)
    }

    for j := 0; j < 20; j++ { jobs <- j }
    close(jobs)
    wg.Wait()
}
```

**Expected output:**
```
worker 0 processed job 0
worker 1 processed job 1
worker 2 processed job 4
... (order varies; all 20 jobs are processed exactly once)
```

**Checkpoint:** Count how many jobs each worker processed. Is the distribution roughly even? Try changing the worker count from 4 to 1, then to 20. Observe how distribution changes.

---

### Lab 4: Fan-In (Merge)

**What you'll practise:** Merging multiple input channels into a single output channel using a WaitGroup.

**Task:**
Create 3 producer goroutines, each sending 5 values on their own channel. Write a `merge` function that combines them into one `<-chan string`. Drain the merged channel in `main`.

**Steps:**
1. Write `producer(name string, count int) <-chan string` that sends `"name-0"`, `"name-1"`, ... then closes
2. Write `merge(channels ...<-chan string) <-chan string` using a `sync.WaitGroup` inside a goroutine
3. Launch 3 producers: `"A"`, `"B"`, `"C"`
4. Drain the merged channel and count total values received

```go
func merge(channels ...<-chan string) <-chan string {
    out := make(chan string)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan string) {
            defer wg.Done()
            for v := range c { out <- v }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}
```

**Expected output:**
```
A-0 B-0 C-0 A-1 B-1 ...  (order varies; all 15 values appear)
Total: 15
```

**Checkpoint:** Use a `map[string]bool` to check for duplicates — there should be none. Verify the total count is exactly 15.

---

### Lab 5: Done Channel Cancellation

**What you'll practise:** Stopping a generator goroutine early using a done channel and `select`.

**Task:**
Write an infinite number generator that checks a `done <-chan struct{}` on each iteration. Cancel it from `main` after receiving exactly 3 values.

**Steps:**
1. Write `generate(done <-chan struct{}) <-chan int` sending 0, 1, 2, ... indefinitely
2. Use `select` inside the generator with two cases: `out <- n` and `<-done`
3. Print a message when the generator goroutine exits using `defer`
4. In `main`, receive exactly 3 values, then `close(done)`

```go
func generate(done <-chan struct{}) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        defer fmt.Println("generator: stopped")
        n := 0
        for {
            select {
            case out <- n:
                n++
            case <-done:
                return
            }
        }
    }()
    return out
}
```

**Expected output:**
```
received: 0
received: 1
received: 2
generator: stopped
main: done
```

**Checkpoint:** Remove the `case <-done` branch. Observe that the goroutine leaks and the process hangs. Restore the done channel to fix it.

---

### Lab 6: select with Timeout

**What you'll practise:** Using `select` with `time.After` to return a timeout error when a result does not arrive in time.

**Task:**
Write `fetchWithTimeout(ch <-chan string, d time.Duration) (string, error)` that returns the result or a timeout error. Test it with a fast producer (10ms) and a slow producer (500ms) using a 100ms deadline.

**Steps:**
1. Write `fetchWithTimeout` using `select` with two cases: the result channel and `time.After(d)`
2. Write `slowProducer(delay time.Duration) <-chan string` that sends after the given delay
3. Call with a 100ms timeout — the fast producer succeeds, the slow one times out
4. Make the slow producer's channel buffered (size 1) to avoid a goroutine leak on timeout

```go
func fetchWithTimeout(ch <-chan string, d time.Duration) (string, error) {
    select {
    case result := <-ch:
        return result, nil
    case <-time.After(d):
        return "", fmt.Errorf("timed out after %v", d)
    }
}

func slowProducer(delay time.Duration) <-chan string {
    ch := make(chan string, 1) // buffered: goroutine can send even if we timed out
    go func() {
        time.Sleep(delay)
        ch <- "result"
    }()
    return ch
}
```

**Expected output:**
```
fast (10ms):  "result" <nil>
slow (500ms): "" timed out after 100ms
```

**Checkpoint:** Change the slow producer's channel to unbuffered. Time out again. Use `runtime.NumGoroutine()` before and after — the count will be higher after the timeout, confirming a goroutine leak.

---

### Lab 7: The Nil Channel Trick

**What you'll practise:** Using a nil channel inside `select` to dynamically disable a case when one input is exhausted.

**Task:**
Read from two channels, `ch1` and `ch2`. When one closes, set it to `nil` to stop selecting it. Exit the loop only when both are nil.

**Steps:**
1. Create `ch1` and `ch2` via a `producer` helper, each sending 3 values at different intervals
2. In a loop, use `select` with both channels using the comma-ok form
3. When a channel closes (`ok == false`), set it to `nil`
4. Exit the loop when `ch1 == nil && ch2 == nil`

```go
func main() {
    ch1 := producer("A", 3, 20*time.Millisecond)
    ch2 := producer("B", 3, 50*time.Millisecond)

    for ch1 != nil || ch2 != nil {
        select {
        case v, ok := <-ch1:
            if !ok { ch1 = nil; continue }
            fmt.Println("from ch1:", v)
        case v, ok := <-ch2:
            if !ok { ch2 = nil; continue }
            fmt.Println("from ch2:", v)
        }
    }
    fmt.Println("both channels drained")
}
```

**Expected output:**
```
from ch1: A-0
from ch2: B-0
from ch1: A-1
from ch1: A-2
from ch2: B-1
from ch2: B-2
both channels drained
```

**Checkpoint:** Remove the nil assignment (`ch1 = nil`). Add `fmt.Println("BUG: spin on closed ch1")` in the ch1 case. Observe the tight spin loop when `ch1` closes — this is why the nil trick exists.

---

### Lab 8: Channel Direction Types

**What you'll practise:** Using directional channel types (`chan<-` and `<-chan`) in function signatures to enforce data flow direction at compile time.

**Task:**
Refactor the pipeline from Lab 2 to use strictly directional parameters. Attempt to violate direction in comments to confirm the compile errors.

**Steps:**
1. Confirm `generate` returns `<-chan int` (receive-only to callers)
2. Confirm `square` accepts `<-chan int` and returns `<-chan int`
3. Add a `sink(in <-chan int)` function that prints values — cannot send on `in`
4. Add a `source(out chan<- int)` function — cannot receive from `out`
5. Comment-out a direction violation in each function and document the error

```go
// generate: owns the channel internally, exposes receive-only to callers
func generate(nums ...int) <-chan int { /* ... */ }

// square: receives from in, sends to internal channel; both typed
func square(in <-chan int) <-chan int { /* ... */ }

// sink: receive-only parameter — compiler prevents accidental send
func sink(in <-chan int) {
    for v := range in {
        fmt.Println("output:", v)
    }
    // in <- 99  // compile error: cannot send to receive-only channel
}
```

**Expected output:**
```
output: 4
output: 9
output: 16
output: 25
# Uncommenting "in <- 99": cannot send to receive-only channel
```

**Checkpoint:** Create a plain `chan int` in main and pass it to `square` (narrowed to `<-chan int`) and to `sink` (also `<-chan int`). Confirm Go's automatic narrowing — no explicit cast needed.

---

## Day Project: Channel-Based Pipeline

Build a concurrent text-processing pipeline that:

1. **Generator** stage: reads lines from a `strings.NewReader` and sends them on a `<-chan string`
2. **Filter** stage: receives lines, discards blanks and comment lines (`#`), sends on
3. **Transform** stage: receives lines, converts to uppercase and trims space, sends on
4. **Sink** stage: receives and prints each line with a line number

Wire them with `context.WithTimeout` — the whole pipeline cancels after 500ms.

Then build a **fan-out + fan-in word counter**:
1. Split a large text into chunks (lines)
2. Fan out to N goroutines that count words in their chunk
3. Fan in results into a merged `map[string]int`
4. Print top-10 words

**Extension ideas:** add a `monitor` goroutine that prints throughput (lines/sec) every 100ms using `time.Ticker`; implement a bounded buffer between stages to handle backpressure.

## Official Documentation

- [`sync`](https://pkg.go.dev/sync) — `WaitGroup` used in fan-in/fan-out patterns
- [`context`](https://pkg.go.dev/context) — `WithTimeout`, `WithCancel`, `Done` for pipeline cancellation
- [`runtime`](https://pkg.go.dev/runtime) — `NumCPU` for sizing worker pools and semaphores
- [`strings`](https://pkg.go.dev/strings) — `NewReader` used in the day project pipeline generator
- [`time`](https://pkg.go.dev/time) — `After`, `NewTicker`, `Ticker` for timeout and throughput monitoring
- [Go Blog: Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines) — canonical reference for pipeline, fan-out, fan-in patterns
- [Go Blog: Go Concurrency Patterns](https://go.dev/blog/concurrency-patterns)
- [Language Spec — Channel types](https://go.dev/ref/spec#Channel_types)
- [Language Spec — Select statements](https://go.dev/ref/spec#Select_statements)
