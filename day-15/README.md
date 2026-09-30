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

## Day Project: Parallel File Hasher

Write a program that:
1. Accepts a directory path from `os.Args[1]`
2. Walks the directory with `filepath.WalkDir`
3. Hashes each file concurrently using goroutines and `crypto/sha256`
4. Collects results through a channel
5. Prints each `filename: hash` sorted by filename

Use a semaphore (buffered channel) to limit concurrency to N workers.

```go
sem := make(chan struct{}, runtime.NumCPU())
```

**Extension ideas:** add a progress bar using `\r` escape; support MD5/SHA512 via a flag; detect duplicate files by hash.
