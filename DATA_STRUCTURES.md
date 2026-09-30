# Data Structures in Go — Reference

## Core Collections at a Glance

| Need | Type / Pattern |
|------|----------------|
| Dynamic list / LIFO stack | `[]T` (slice) |
| FIFO queue | `[]T` with head index, or `container/ring` |
| Key-value lookup (unordered) | `map[K]V` |
| Key-value lookup (sorted, range queries) | `btree` (golang.org/x/exp/slices + sorted slice) or external `tidwall/btree` |
| Set membership | `map[T]struct{}` |
| Priority queue (min or max) | `container/heap` |
| Doubly linked list | `container/list` |
| Circular buffer | `container/ring` |
| Tree / recursive structures | struct + pointer fields |
| Graph relationships | `map[string][]string` (adjacency list) |
| Concurrent map | `sync.Map` |

---

## Slice as Stack

The standard `[]T` slice already has O(1) push and pop at the tail.

```go
stack := []int{}
stack = append(stack, 1)   // push
top := stack[len(stack)-1] // peek
stack = stack[:len(stack)-1] // pop
```

**Complexity:** push O(1) amortised, pop O(1), peek O(1).

Avoid using the front of a slice as a stack bottom — `append` at index 0 is O(n).

---

## Slice as Queue (simple)

```go
queue := []int{}
queue = append(queue, 1)    // enqueue
front := queue[0]           // peek
queue = queue[1:]           // dequeue — O(n), don't use for hot paths
```

For a high-throughput queue use a head-pointer approach or `container/ring`:

```go
// head-pointer queue (avoids re-slicing)
type Queue[T any] struct {
    buf  []T
    head int
}

func (q *Queue[T]) Enqueue(v T) { q.buf = append(q.buf, v) }
func (q *Queue[T]) Dequeue() (T, bool) {
    if q.head >= len(q.buf) {
        var z T; return z, false
    }
    v := q.buf[q.head]; q.head++; return v, true
}
```

---

## map

`map[K]V` is Go's built-in hash table. Keys must satisfy `comparable`.

```go
m := make(map[string]int)
m["key"]++                  // safe: reads return zero value
v, ok := m["key"]           // comma-ok: ok=false means absent
delete(m, "key")

// Frequency count idiom
words := strings.Fields(text)
freq := make(map[string]int, len(words)) // pre-size hint
for _, w := range words {
    freq[w]++
}
```

**Complexity:** insert, lookup, delete all O(1) average, O(n) worst (hash collision).

Iteration order is intentionally **randomised** on every run — do not rely on it. For reproducible order, collect keys, sort them, then range over the sorted keys.

`map` is not safe for concurrent access. Use `sync.Mutex` around reads and writes, or `sync.Map` for cases with many readers and infrequent writes.

---

## Set

Go has no built-in set. Use a map with empty struct values (zero allocation):

```go
seen := make(map[string]struct{})
seen["alice"] = struct{}{}
_, ok := seen["alice"]   // membership test
delete(seen, "alice")
```

---

## container/heap — Priority Queue

`container/heap` provides a heap on any type that implements `heap.Interface`:

```go
import "container/heap"

type MinHeap []int

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i] < h[j] } // min-heap
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MinHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *MinHeap) Pop() any {
    old := *h; n := len(old)
    x := old[n-1]; *h = old[:n-1]; return x
}

h := &MinHeap{3, 1, 4}
heap.Init(h)
heap.Push(h, 2)
min := heap.Pop(h).(int)  // 1
```

For a **max-heap** reverse the `Less` comparison. `container/heap` does not support updating the priority of an existing element — insert duplicates and discard stale entries lazily on pop.

**Complexity:** push O(log n), pop O(log n), init O(n).

---

## container/list — Doubly Linked List

```go
import "container/list"

l := list.New()
e := l.PushBack(42)
l.PushFront(0)
l.Remove(e)

for e := l.Front(); e != nil; e = e.Next() {
    fmt.Println(e.Value.(int))
}
```

Useful when you need O(1) insertion and removal at arbitrary positions. Cache worse than slice for sequential access. Use `*list.Element` values as handles — they become invalid after `Remove`.

---

## container/ring — Circular Buffer

```go
import "container/ring"

r := ring.New(5)   // ring of 5 elements
for i := 0; i < r.Len(); i++ {
    r.Value = i
    r = r.Next()
}
r.Do(func(v any) { fmt.Println(v) })
```

Good for fixed-size sliding windows (e.g., rolling averages).

---

## Singly Linked List (custom)

```go
type Node[T any] struct {
    Value T
    Next  *Node[T]
}

type List[T any] struct {
    Head *Node[T]
    size int
}
```

Deep lists risk stack overflow on recursive `String()` or traversal — use an iterative loop.

---

## Binary Search Tree (custom)

```go
type BST[T constraints.Ordered] struct {
    Value       T
    Left, Right *BST[T]
}
```

Worst case O(n) on sorted input. For production sorted-key lookup use a sorted slice + `sort.Search` (O(log n)) or an external B-tree package.

---

## Graph — Adjacency List

```go
type Graph map[string][]string

func (g Graph) AddEdge(from, to string) {
    g[from] = append(g[from], to)
    g[to] = append(g[to], from) // undirected
}

// BFS — shortest unweighted path
func BFS(g Graph, start string) map[string]int {
    dist := map[string]int{start: 0}
    queue := []string{start}
    for len(queue) > 0 {
        node := queue[0]; queue = queue[1:]
        for _, nb := range g[node] {
            if _, seen := dist[nb]; !seen {
                dist[nb] = dist[node] + 1
                queue = append(queue, nb)
            }
        }
    }
    return dist
}
```

For weighted shortest paths use Dijkstra's algorithm with a `container/heap`.

---

## sync.Map — Concurrent Map

Prefer a `map` + `sync.Mutex` for most cases (simpler, easier to reason about). Use `sync.Map` when:
- Many goroutines read disjoint keys (high read concurrency)
- Keys are written once and read many times

```go
var m sync.Map
m.Store("key", 42)
v, ok := m.Load("key")
m.Range(func(k, v any) bool {
    fmt.Println(k, v)
    return true  // continue; return false to stop
})
```

---

## Sorting

```go
import "sort"

// Slice of any type
sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })

// Stable sort (preserves original order of equal elements)
sort.SliceStable(s, func(i, j int) bool { return s[i].Name < s[j].Name })

// Sorted search (binary search)
i := sort.SearchInts(sorted, target) // index where target would be inserted

// Built-in typed helpers
sort.Ints(nums)
sort.Strings(strs)
sort.Float64s(fs)
```

Implement `sort.Interface` (Len, Less, Swap) for custom types that need to be sorted in multiple ways.

---

## Decision Heuristic

```
Fast random access?               → []T (slice)
Key-value lookup?                 → map[K]V
FIFO (queue)?                     → []T with head-pointer, or container/ring
LIFO (stack)?                     → []T
Priority ordering?                → container/heap
Set membership?                   → map[T]struct{}
O(1) mid-list insert/remove?      → container/list
Sorted iteration / range queries? → sorted []T + sort.Search, or external btree
Recursive / tree structure?       → struct with pointer fields
Concurrent access (read-heavy)?   → sync.Map
Concurrent access (mixed)?        → map + sync.RWMutex
```
