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
