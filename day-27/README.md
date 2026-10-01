# Day 27: Benchmarking and Profiling

## Core Concept: Measure Before Optimising

Go's built-in benchmark framework and `pprof` profiler make performance analysis a first-class workflow.

## Writing Benchmarks

```go
// bench_test.go
func BenchmarkWordCount(b *testing.B) {
    text := strings.Repeat("the quick brown fox jumps over the lazy dog ", 1000)
    b.ResetTimer() // exclude setup from measurement
    for i := 0; i < b.N; i++ {
        wordCount(text)
    }
}

func BenchmarkWordCountParallel(b *testing.B) {
    text := strings.Repeat("hello world ", 1000)
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            wordCount(text)
        }
    })
}
```

## Running Benchmarks

```bash
go test -bench=.                                  # run all benchmarks
go test -bench=BenchmarkWordCount -benchtime=5s   # run for 5 seconds
go test -bench=. -benchmem                        # include allocation stats
go test -bench=. -count=3                         # run 3 times for stability
```

Sample output:
```
BenchmarkWordCount-8   500000   2345 ns/op   512 B/op   3 allocs/op
```

- `500000` — number of iterations
- `2345 ns/op` — nanoseconds per operation
- `512 B/op` — bytes allocated per operation
- `3 allocs/op` — heap allocations per operation

## CPU Profiling

```bash
go test -bench=BenchmarkWordCount -cpuprofile=cpu.prof
go tool pprof cpu.prof
# Inside pprof shell:
(pprof) top10
(pprof) list wordCount
(pprof) web      # opens flame graph in browser (requires graphviz)
```

## Memory Profiling

```bash
go test -bench=. -memprofile=mem.prof -benchmem
go tool pprof mem.prof
(pprof) top --alloc_objects
```

## Profiling a Running Server

```go
import _ "net/http/pprof"
// registers /debug/pprof/ handlers on the default mux
```

```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## Common Optimisation Techniques

| Issue | Fix |
|-------|-----|
| Too many allocations | Pre-allocate slices; use `sync.Pool` |
| String concatenation in loop | Use `strings.Builder` |
| Map read contention | Shard or use `sync.Map` |
| Excessive copying | Pass pointers; use `[]byte` instead of `string` |

## Labs

### Lab 1: First Benchmark — Reading the Output

**What you'll practise:** Writing your first `BenchmarkXxx` function and interpreting the ns/op, B/op, and allocs/op columns.

**Task:**
Benchmark a trivial integer addition function to learn the benchmark skeleton, then read what the output columns mean.

**Steps:**
1. Create `bench_test.go` in `day-26/`
2. Write a `BenchmarkAdd` function using the `b.N` loop
3. Run with `go test -bench=. -benchmem` and identify each column

```go
package main

import "testing"

func add(a, b int) int { return a + b }

func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        add(3, 4)
    }
}
```

**Expected output:**
```
BenchmarkAdd-8   1000000000   0.23 ns/op   0 B/op   0 allocs/op
```

**Checkpoint:** The benchmark runs without error. You can explain what `b.N`, `ns/op`, `B/op`, and `allocs/op` each mean.

---

### Lab 2: Compare Implementations — String Concatenation vs Builder

**What you'll practise:** Benchmarking two approaches to string building and observing the dramatic difference in allocations.

**Task:**
Write two string-joining functions — one using `+` in a loop, one using `strings.Builder` — and benchmark both with `-benchmem` to see how allocations differ.

**Steps:**
1. Write `joinConcat(words []string) string` using `+=`
2. Write `joinBuilder(words []string) string` using `strings.Builder`
3. Benchmark both with the same input slice
4. Record the `allocs/op` difference in a comment

```go
func joinConcat(words []string) string {
    s := ""
    for _, w := range words {
        s += w + " "
    }
    return s
}

func joinBuilder(words []string) string {
    var b strings.Builder
    for _, w := range words {
        b.WriteString(w)
        b.WriteByte(' ')
    }
    return b.String()
}

func BenchmarkJoinConcat(b *testing.B) {
    words := strings.Fields(strings.Repeat("the quick brown fox ", 50))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        joinConcat(words)
    }
}
```

**Expected output:**
```
BenchmarkJoinConcat-8    50000   25000 ns/op   12345 B/op   199 allocs/op
BenchmarkJoinBuilder-8  500000    2300 ns/op     512 B/op     2 allocs/op
```

**Checkpoint:** `joinBuilder` has fewer than 5 allocs/op. `joinConcat` has allocs proportional to the number of words (one per concatenation).

---

### Lab 3: b.ResetTimer and b.StopTimer — Excluding Setup Time

**What you'll practise:** Using `b.ResetTimer` and `b.StopTimer`/`b.StartTimer` to measure only the code under test, not its setup.

**Task:**
Benchmark a function that requires expensive setup (loading a large word list). Use `b.StopTimer`/`b.StartTimer` to exclude the setup cost from each iteration.

**Steps:**
1. Write a `BenchmarkWithSetup` that builds a large string outside the loop using `b.ResetTimer`
2. Write a `BenchmarkPerIterSetup` that resets per-iteration state using `b.StopTimer`/`b.StartTimer`
3. Run both and compare — the per-iter setup benchmark should show higher ns/op

```go
func BenchmarkWithSetup(b *testing.B) {
    // one-time setup
    text := strings.Repeat("go is fast ", 10000)
    b.ResetTimer() // start measuring from here

    for i := 0; i < b.N; i++ {
        wordCount(text)
    }
}

func BenchmarkPerIterSetup(b *testing.B) {
    for i := 0; i < b.N; i++ {
        b.StopTimer()
        text := strings.Repeat("go is fast ", 10000) // excluded
        b.StartTimer()

        wordCount(text)
    }
}
```

**Expected output:**
```
BenchmarkWithSetup-8       5000   300000 ns/op
BenchmarkPerIterSetup-8    5000   305000 ns/op
```

**Checkpoint:** `BenchmarkWithSetup` runs faster because setup happens once. `BenchmarkPerIterSetup` shows slightly higher ns/op from repeated allocations inside the loop.

---

### Lab 4: Sub-benchmarks — Comparing Multiple Variants with b.Run

**What you'll practise:** Organising related benchmarks as sub-benchmarks using `b.Run` for structured comparison output.

**Task:**
Benchmark three word-count implementations as sub-benchmarks under a single `BenchmarkWordCount` parent. Compare their ns/op and allocs/op side by side.

**Steps:**
1. Write three `wordCount` functions: `countSplit`, `countFields`, `countScanner`
2. Use `b.Run("Split", func(b *testing.B){...})` pattern
3. Run with `go test -bench=BenchmarkWordCount -benchmem`

```go
func BenchmarkWordCount(b *testing.B) {
    text := strings.Repeat("the quick brown fox jumps over the lazy dog ", 1000)

    b.Run("Split", func(b *testing.B) {
        for i := 0; i < b.N; i++ { countSplit(text) }
    })
    b.Run("Fields", func(b *testing.B) {
        for i := 0; i < b.N; i++ { countFields(text) }
    })
    b.Run("Scanner", func(b *testing.B) {
        for i := 0; i < b.N; i++ { countScanner(text) }
    })
}
```

**Expected output:**
```
BenchmarkWordCount/Split-8     5000   280000 ns/op   81920 B/op   3 allocs/op
BenchmarkWordCount/Fields-8    5000   220000 ns/op   40960 B/op   2 allocs/op
BenchmarkWordCount/Scanner-8   8000   195000 ns/op    4096 B/op   2 allocs/op
```

**Checkpoint:** Three sub-benchmark lines appear under one parent. You can identify the fastest implementation by ns/op.

---

### Lab 5: CPU Profiling — Finding the Hot Path

**What you'll practise:** Generating a CPU profile and navigating it with `go tool pprof` to find where time is spent.

**Task:**
Run the slowest word-count benchmark with `-cpuprofile`, then use `pprof top10` and `list` to identify which function consumes the most CPU.

**Steps:**
1. Run: `go test -bench=BenchmarkWordCount/Split -cpuprofile=cpu.prof -benchtime=5s`
2. Open the profile: `go tool pprof cpu.prof`
3. In the pprof shell, run `top10` to see top CPU consumers
4. Run `list countSplit` to see annotated source lines

```bash
go test -bench=BenchmarkWordCount/Split -cpuprofile=cpu.prof -benchtime=5s
go tool pprof cpu.prof
```

```
(pprof) top10
(pprof) list countSplit
(pprof) web     # opens flame graph in browser (requires graphviz)
```

**Expected output:**
```
Showing top 10 nodes out of 42
      flat  flat%   sum%        cum   cum%
     450ms 45.00% 45.00%      450ms 45.00%  strings.genSplit
     ...
```

**Checkpoint:** You can name the top two functions by CPU time and explain which line of `countSplit` is the hot path.

---

### Lab 6: Memory Profiling — Finding the Biggest Allocator

**What you'll practise:** Generating a memory profile and using `pprof -alloc_space` to find the function allocating the most heap memory.

**Task:**
Profile memory allocations for the three word-count implementations. Identify which allocates the most bytes and why.

**Steps:**
1. Run: `go test -bench=BenchmarkWordCount -memprofile=mem.prof -benchmem`
2. Open the profile: `go tool pprof -alloc_space mem.prof`
3. Run `top10` to see allocation hot spots
4. Compare allocation sizes between implementations

```bash
go test -bench=BenchmarkWordCount -memprofile=mem.prof -benchmem
go tool pprof -alloc_space mem.prof
```

```
(pprof) top10
(pprof) list countSplit
```

**Expected output:**
```
Showing top 10 nodes out of 18
      flat  flat%   sum%        cum   cum%
    512MB 60.00% 60.00%      512MB 60.00%  strings.genSplit
    ...
```

**Checkpoint:** You can explain why `countSplit` allocates more than `countScanner` and which pprof metric (`alloc_space` vs `inuse_space`) you used and why.

---

### Lab 7: Trace — Observing Goroutine Scheduling

**What you'll practise:** Generating an execution trace with `go test -trace` and viewing it with `go tool trace` to observe goroutine scheduling and GC activity.

**Task:**
Run one benchmark with trace enabled, open the trace viewer, and identify at least one GC pause and the goroutine that ran the benchmark.

**Steps:**
1. Run: `go test -bench=BenchmarkWordCount/Scanner -trace=trace.out -benchtime=1s`
2. Open: `go tool trace trace.out` (opens browser)
3. Click "Goroutine analysis" and find the test goroutine
4. Click "View trace" and zoom into a GC event

```bash
go test -bench=BenchmarkWordCount/Scanner -trace=trace.out -benchtime=1s
go tool trace trace.out
```

**Expected output:**
A browser window opens with a timeline view showing goroutine execution, GC pauses, and system calls.

**Checkpoint:** You can locate a GC pause in the trace timeline and describe what "STW" (stop-the-world) looks like visually.

---

### Final Lab (Project): Profile a Word-Frequency Counter

**What you'll practise:** Writing three benchmark implementations, profiling the slowest, and producing a comparison table that proves the optimisation.

**Task:**
Write three implementations of a word-frequency counter, benchmark all three with `-benchmem`, generate a CPU profile for the slowest, and optimise it. Present your findings as a comparison table in a comment.

**Steps:**
1. Write three implementations of a word-frequency counter:
   - V1: `strings.Split` + map (baseline)
   - V2: `strings.Fields` + map with pre-size hint
   - V3: `bufio.Scanner` with `ScanWords` + map
2. Benchmark all three with `-benchmem`
3. Generate a CPU profile for the slowest implementation and identify the hot path
4. Optimise and re-benchmark to show improvement

```go
// V1 — baseline
func countSplit(text string) map[string]int {
    freq := map[string]int{}
    for _, w := range strings.Split(text, " ") {
        if w != "" { freq[w]++ }
    }
    return freq
}

// V2 — pre-sized map
func countFields(text string) map[string]int {
    words := strings.Fields(text)
    freq  := make(map[string]int, len(words)/2)
    for _, w := range words { freq[w]++ }
    return freq
}

// V3 — scanner (lowest allocation)
func countScanner(text string) map[string]int {
    freq := make(map[string]int, 64)
    sc   := bufio.NewScanner(strings.NewReader(text))
    sc.Split(bufio.ScanWords)
    for sc.Scan() { freq[sc.Text()]++ }
    return freq
}
```

**Expected output:**
```
BenchmarkWordCount/Split-8     5000   280000 ns/op   81920 B/op   3 allocs/op
BenchmarkWordCount/Fields-8    5000   220000 ns/op   40960 B/op   2 allocs/op
BenchmarkWordCount/Scanner-8   8000   195000 ns/op    4096 B/op   2 allocs/op
```

**Checkpoint:** All three benchmarks run. Scanner is fastest or lowest-allocation. A CPU profile identifies the hot function. Run benchmarks with: `go test -bench=. -benchmem`

**Extension ideas:** try `sync.Pool` to reduce allocations; experiment with `bytes.FieldsFunc` to avoid the `strings` package entirely.

## Official Documentation

- [`testing`](https://pkg.go.dev/testing) — `B` (benchmark type), `B.N`, `B.ResetTimer`, `B.RunParallel`, `PB.Next`, `B.ReportAllocs`
- [`strings`](https://pkg.go.dev/strings) — `Split`, `Fields`, `Builder`, `Repeat` used in benchmark examples
- [`bufio`](https://pkg.go.dev/bufio) — `Scanner`, `ScanWords` split function for the V3 implementation
- [`sync`](https://pkg.go.dev/sync) — `Pool` for reducing allocations
- [`net/http/pprof`](https://pkg.go.dev/net/http/pprof) — registers `/debug/pprof/` handlers on a running server
- [Go Blog: Profiling Go Programs](https://go.dev/blog/pprof)
- [Go Blog: Benchmarks](https://go.dev/testing/#hdr-Benchmarks)
- [`runtime`](https://pkg.go.dev/runtime) — `NumCPU` used in semaphore patterns
