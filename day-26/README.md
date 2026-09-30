# Day 26: Benchmarking and Profiling

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

## Day Project: Profile Word-Frequency Counter

1. Write three implementations of a word-frequency counter:
   - V1: `strings.Split` + map (baseline)
   - V2: `strings.Fields` + map with pre-size hint
   - V3: `bufio.Scanner` with `ScanWords` + map
2. Benchmark all three with `-benchmem`
3. Generate a CPU profile for the slowest implementation and identify the hot path
4. Optimise and re-benchmark to show improvement

Run benchmarks with: `go test -bench=. -benchmem`

**Extension ideas:** try `sync.Pool` to reduce allocations; experiment with `bytes.FieldsFunc` to avoid the `strings` package entirely.
