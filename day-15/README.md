# Day 15: Goroutines and Channels

## Core Concept: Communicating Sequential Processes

Go's concurrency is based on CSP: goroutines are lightweight independently-executing functions; channels are typed conduits for communication between them.

> "Do not communicate by sharing memory; share memory by communicating."

## Goroutines

A goroutine is started with the `go` keyword. It is not a thread — the Go runtime multiplexes thousands of goroutines onto a small number of OS threads.

```go
go func() {
    fmt.Println("running concurrently")
}()
// main does not wait — need to synchronise
```

## Channels

```go
ch := make(chan int)      // unbuffered — sender blocks until receiver is ready
ch := make(chan int, 10)  // buffered — sender blocks only when full

ch <- 42          // send
v := <-ch         // receive
v, ok := <-ch     // ok is false when channel is closed and empty
close(ch)         // signal no more values
```

`range` over a channel receives until closed:

```go
for v := range ch {
    fmt.Println(v)
}
```

## select

`select` waits on multiple channel operations — like a switch for channels:

```go
select {
case msg := <-ch1:
    fmt.Println("ch1:", msg)
case msg := <-ch2:
    fmt.Println("ch2:", msg)
case <-time.After(1 * time.Second):
    fmt.Println("timeout")
}
```

## Fan-Out / Fan-In

```go
// Fan-out: distribute work across N workers
func fanOut(in <-chan string, n int) []<-chan string {
    outs := make([]<-chan string, n)
    for i := range outs {
        outs[i] = worker(in)
    }
    return outs
}

// Fan-in: merge multiple channels into one
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
    go func() { wg.Wait(); close(out) }()
    return out
}
```

## Done Channel (Cancellation)

```go
done := make(chan struct{})
go func() {
    select {
    case <-done:
        return
    case result := <-work:
        process(result)
    }
}()
close(done) // signal all goroutines to stop
```

## Labs

### Lab 1: First Goroutine — sync.WaitGroup

**What you'll practise:** Launching goroutines and using `sync.WaitGroup` to wait for them all.

**Task:**
Launch 5 goroutines, each printing its number. First observe the non-deterministic output without synchronisation. Then add a `sync.WaitGroup` so `main` waits for all goroutines to finish.

**Steps:**
1. Launch 5 goroutines in a loop without any waiting — run it several times and notice the order changes (or goroutines may not print at all if main exits first)
2. Add `var wg sync.WaitGroup`, call `wg.Add(1)` before each `go`, `defer wg.Done()` inside, and `wg.Wait()` in main
3. Run again — all 5 numbers appear every time, order may still vary

```go
var wg sync.WaitGroup
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        fmt.Printf("goroutine %d running\n", n)
    }(i)
}
wg.Wait()
fmt.Println("all done")
```

**Expected output (order varies):**
```
goroutine 3 running
goroutine 0 running
goroutine 4 running
goroutine 1 running
goroutine 2 running
all done
```

**Checkpoint:** Every run prints all 5 goroutine lines before "all done".

---

### Lab 2: Unbuffered Channel — Synchronous Handshake

**What you'll practise:** Sending and receiving on an unbuffered channel to observe blocking behaviour.

**Task:**
Create an unbuffered `chan int`. Launch a producer goroutine that sends 1..10. In main, receive and print each value. Add `fmt.Println` before and after the send to see that the producer blocks until main is ready.

**Steps:**
1. Create `ch := make(chan int)` (no second argument)
2. Launch a goroutine that sends `1..10` then closes the channel
3. In main, receive with `v := <-ch` in a loop
4. Add log lines around the send to visualise the blocking handshake

```go
ch := make(chan int)
go func() {
    for i := 1; i <= 10; i++ {
        fmt.Printf("sending %d\n", i)
        ch <- i
        fmt.Printf("sent %d\n", i)
    }
    close(ch)
}()
for v := range ch {
    fmt.Printf("received %d\n", v)
}
```

**Expected output:**
```
sending 1
received 1
sent 1
sending 2
received 2
...
```

**Checkpoint:** "sending N" and "received N" always appear together — the sender never races ahead.

---

### Lab 3: Buffered Channel — Producer Runs Ahead

**What you'll practise:** Understanding how buffer capacity decouples sender and receiver timing.

**Task:**
Change the channel from Lab 2 to `make(chan int, 5)`. Observe that the producer can send up to 5 values before blocking. Add a `time.Sleep` in the receiver to exaggerate the effect.

**Steps:**
1. Change to `ch := make(chan int, 5)`
2. Add `time.Sleep(100 * time.Millisecond)` in the receiver loop
3. Run and observe the producer fires off 5 sends rapidly, then blocks waiting for room
4. Remove the sleep and compare the timing

```go
ch := make(chan int, 5)
go func() {
    for i := 1; i <= 10; i++ {
        ch <- i
        fmt.Printf("queued %d\n", i)
    }
    close(ch)
}()
for v := range ch {
    time.Sleep(100 * time.Millisecond)
    fmt.Printf("processed %d\n", v)
}
```

**Expected output:**
```
queued 1
queued 2
queued 3
queued 4
queued 5
processed 1
queued 6
processed 2
...
```

**Checkpoint:** The first 5 "queued" lines appear before the first "processed" line.

---

### Lab 4: Range over Channel — Clean Producer/Consumer

**What you'll practise:** Using `close(ch)` and `for v := range ch` for a clean, panic-free consumer.

**Task:**
Write a `generate(nums ...int) <-chan int` function that sends numbers on a channel and closes it when done. In main, consume with `for v := range ch`. Show that ranging over a closed channel terminates naturally.

**Steps:**
1. `generate` returns a receive-only `<-chan int`
2. Inside `generate`, launch a goroutine that sends each number then calls `close(ch)`
3. In main: `for v := range generate(1, 2, 3, 4, 5) { fmt.Println(v) }`
4. Demonstrate that no sentinel value or separate done channel is needed

```go
func generate(nums ...int) <-chan int {
    ch := make(chan int)
    go func() {
        for _, n := range nums {
            ch <- n
        }
        close(ch)
    }()
    return ch
}

func main() {
    for v := range generate(1, 2, 3, 4, 5) {
        fmt.Println(v)
    }
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

**Checkpoint:** No goroutine leaks — the producer goroutine exits after `close(ch)`.

---

### Lab 5: Select — Multiplexing Channels with Timeout

**What you'll practise:** Using `select` to multiplex two channels and handle timeouts.

**Task:**
Create two ticker channels firing at different rates (200ms and 500ms). Use `select` in a loop to print which fired. After 2 seconds, use `time.After` in the `select` to break out of the loop.

**Steps:**
1. Create `fast := time.NewTicker(200 * time.Millisecond)` and `slow := time.NewTicker(500 * time.Millisecond)`
2. Add `timeout := time.After(2 * time.Second)`
3. `select` on all three channels; on timeout, `return` or `break`
4. Add a `default` case, observe it spins; remove `default`, observe it blocks cleanly

```go
fast := time.NewTicker(200 * time.Millisecond)
slow := time.NewTicker(500 * time.Millisecond)
timeout := time.After(2 * time.Second)

for {
    select {
    case t := <-fast.C:
        fmt.Println("fast tick at", t.Format("15:04:05.000"))
    case t := <-slow.C:
        fmt.Println("slow tick at", t.Format("15:04:05.000"))
    case <-timeout:
        fmt.Println("done")
        return
    }
}
```

**Expected output:**
```
fast tick at 00:00:00.200
fast tick at 00:00:00.400
slow tick at 00:00:00.500
fast tick at 00:00:00.600
...
done
```

**Checkpoint:** The loop exits at approximately 2 seconds and prints roughly 10 fast ticks and 4 slow ticks.

---

### Lab 6: Done Channel Pattern — Graceful Shutdown

**What you'll practise:** Using a `done chan struct{}` to signal goroutines to stop.

**Task:**
Start a worker goroutine that processes items in an infinite loop, printing each item. After 1 second, close the `done` channel and verify the worker exits cleanly — no goroutine leak.

**Steps:**
1. Create `done := make(chan struct{})` and `jobs := make(chan int, 10)`
2. Launch a worker that `select`s on `jobs` and `done`
3. Send 5 jobs, sleep 1 second, then `close(done)`
4. Use a `sync.WaitGroup` to confirm the worker exited

```go
done := make(chan struct{})
jobs := make(chan int, 10)
var wg sync.WaitGroup

wg.Add(1)
go func() {
    defer wg.Done()
    for {
        select {
        case <-done:
            fmt.Println("worker shutting down")
            return
        case j := <-jobs:
            fmt.Printf("processing job %d\n", j)
        }
    }
}()

for i := 1; i <= 5; i++ { jobs <- i }
time.Sleep(time.Second)
close(done)
wg.Wait()
fmt.Println("clean exit")
```

**Expected output:**
```
processing job 1
...
processing job 5
worker shutting down
clean exit
```

**Checkpoint:** `wg.Wait()` returns and "clean exit" prints — no goroutine remains.

---

### Lab 7: Fan-Out — Distribute Work Across N Workers

**What you'll practise:** The fan-out pattern: one jobs channel consumed by multiple workers.

**Task:**
Create a `jobs` channel and `results` channel. Launch N=4 worker goroutines that each read from `jobs`, square the number, and send to `results`. Send 20 jobs, collect all results, and print them.

**Steps:**
1. Create `jobs := make(chan int, 20)` and `results := make(chan int, 20)`
2. Launch 4 workers, each reading from `jobs` and writing to `results`
3. Send integers 1..20 on `jobs`, then `close(jobs)`
4. Use `sync.WaitGroup` to close `results` after all workers finish
5. Collect and print all results

```go
const numWorkers = 4
jobs := make(chan int, 20)
results := make(chan int, 20)

var wg sync.WaitGroup
for w := 0; w < numWorkers; w++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for j := range jobs {
            results <- j * j
        }
    }()
}

for i := 1; i <= 20; i++ { jobs <- i }
close(jobs)

go func() { wg.Wait(); close(results) }()

for r := range results {
    fmt.Println(r)
}
```

**Expected output (order varies):**
```
1
4
9
...
400
```

**Checkpoint:** All 20 squared values appear in the output (in any order). The program exits cleanly.

---

### Final Lab (Project): Parallel File Hasher

**What you'll practise:** Applying goroutines, channels, `sync.WaitGroup`, and a semaphore pattern to a real concurrent program.

**Task:**
Write a program that walks a directory and hashes every file concurrently using SHA-256, collecting results through a channel and printing them sorted by filename.

**Steps:**
1. Accept a directory path from `os.Args[1]`
2. Walk with `filepath.WalkDir`, skip directories
3. Use a semaphore (`make(chan struct{}, N)`) to cap concurrency at `runtime.NumCPU()` workers
4. Launch a goroutine per file: acquire semaphore, hash with `crypto/sha256`, release semaphore, send result on channel
5. Use `sync.WaitGroup` to close the results channel after all goroutines finish
6. Collect results into a slice, sort by filename, print `filename: hexhash`

```go
sem := make(chan struct{}, runtime.NumCPU())

var wg sync.WaitGroup
results := make(chan result)

go func() { wg.Wait(); close(results) }()

filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
    if err != nil || d.IsDir() { return err }
    wg.Add(1)
    go func() {
        defer wg.Done()
        sem <- struct{}{}
        defer func() { <-sem }()
        data, _ := os.ReadFile(path)
        results <- result{path, fmt.Sprintf("%x", sha256.Sum256(data))}
    }()
    return nil
})
```

**Expected output:**
```
day-15/main.go: 3b4c9f...
day-15/README.md: a1f203...
```

**Checkpoint:** Output contains one line per file in the directory, sorted alphabetically by path, with no goroutine leak (program exits cleanly).

---

## Day Project: Parallel File Hasher

Write a program that:
1. Accepts a directory path from `os.Args[1]`
2. Walks the directory with [`filepath.WalkDir`](https://pkg.go.dev/path/filepath#WalkDir)
3. Hashes each file concurrently using goroutines and [`crypto/sha256`](https://pkg.go.dev/crypto/sha256)
4. Collects results through a channel
5. Prints each `filename: hash` sorted by filename

Use a semaphore (buffered channel) to limit concurrency to N workers.

```go
sem := make(chan struct{}, runtime.NumCPU())
```

**Extension ideas:** add a progress bar using `\r` escape; support MD5/SHA512 via a flag; detect duplicate files by hash.

## Official Documentation

- [`sync`](https://pkg.go.dev/sync) — WaitGroup for goroutine coordination
- [`runtime`](https://pkg.go.dev/runtime) — `NumCPU()` for worker pool sizing
- [`path/filepath`](https://pkg.go.dev/path/filepath) — `WalkDir` for directory traversal
- [`crypto/sha256`](https://pkg.go.dev/crypto/sha256) — SHA-256 file hashing
- [`os`](https://pkg.go.dev/os) — `os.Args`, file reading
- [`fmt`](https://pkg.go.dev/fmt) — formatted output
- [Language Spec: Go statements](https://go.dev/ref/spec#Go_statements) — goroutine launch syntax
- [Language Spec: Channel types](https://go.dev/ref/spec#Channel_types) — channel declaration and direction
- [Language Spec: Select statements](https://go.dev/ref/spec#Select_statements) — multi-channel select
- [Effective Go: Concurrency](https://go.dev/doc/effective_go#concurrency) — goroutines and channels
- [Go Blog: Go Concurrency Patterns: Pipelines](https://go.dev/blog/pipelines) — fan-out/fan-in patterns
- [Go Tour: Concurrency](https://go.dev/tour/concurrency/1) — interactive goroutines and channels tour
