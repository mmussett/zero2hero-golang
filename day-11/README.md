# Day 11: Standard Library Collections

## Sorting

Go's [`sort`](https://pkg.go.dev/sort) package sorts slices in-place.

```go
// Concrete helpers
sort.Ints([]int{3,1,2})
sort.Strings([]string{"c","a","b"})

// Generic — works on any slice with a comparator
sort.Slice(people, func(i, j int) bool {
    return people[i].Age < people[j].Age
})

// Binary search on sorted slice
i := sort.SearchInts(sorted, target)
// i is the insertion point; check sorted[i] == target

// Custom type implementing sort.Interface
type ByLength []string
func (b ByLength) Len() int           { return len(b) }
func (b ByLength) Less(i, j int) bool { return len(b[i]) < len(b[j]) }
func (b ByLength) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
sort.Sort(ByLength(words))
```

## [container/heap](https://pkg.go.dev/container/heap) — Priority Queue

Implement `heap.Interface` on a slice:

```go
type MinIntHeap []int
func (h MinIntHeap) Len() int           { return len(h) }
func (h MinIntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h MinIntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MinIntHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *MinIntHeap) Pop() any {
    old := *h; n := len(old); x := old[n-1]; *h = old[:n-1]; return x
}

h := &MinIntHeap{3, 1, 4, 1, 5}
heap.Init(h)
heap.Push(h, 2)
fmt.Println(heap.Pop(h)) // 1
```

## [container/list](https://pkg.go.dev/container/list) — Doubly Linked List

```go
l := list.New()
e1 := l.PushBack("first")
e2 := l.PushBack("second")
l.InsertAfter("middle", e1)
l.Remove(e2)

for e := l.Front(); e != nil; e = e.Next() {
    fmt.Println(e.Value)
}
```

## Graph — Adjacency List

```go
type Graph[T comparable] map[T][]T

func (g Graph[T]) AddEdge(from, to T) {
    g[from] = append(g[from], to)
}

func (g Graph[T]) BFS(start T) []T {
    visited := map[T]bool{start: true}
    queue   := []T{start}
    result  := []T{}
    for len(queue) > 0 {
        node := queue[0]; queue = queue[1:]
        result = append(result, node)
        for _, nb := range g[node] {
            if !visited[nb] {
                visited[nb] = true
                queue = append(queue, nb)
            }
        }
    }
    return result
}
```

## Labs

### Lab 1: Min-Heap with `container/heap`

**What you'll practise:** implementing `heap.Interface` on a custom type and using `heap.Init`, `heap.Push`, and `heap.Pop`.

**Task:**
Create a min-heap of ints. Push 10 random numbers into it, then pop them all and verify they emerge in ascending sorted order.

**Steps:**
1. Define `type MinIntHeap []int`
2. Implement `Len`, `Less`, `Swap`, `Push`, and `Pop` so that `Less(i,j)` returns `h[i] < h[j]`
3. Call `heap.Init` on an existing slice, then push additional elements with `heap.Push`
4. Pop all elements into a result slice and check it is sorted

```go
import "container/heap"

type MinIntHeap []int

func (h MinIntHeap) Len() int           { return len(h) }
func (h MinIntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h MinIntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MinIntHeap) Push(x any) { *h = append(*h, x.(int)) }

func (h *MinIntHeap) Pop() any {
    old := *h; n := len(old)
    x := old[n-1]; *h = old[:n-1]
    return x
}
```

**Expected output:**
```
Sorted: [1 2 3 4 5 6 7 8 9 10]
```

**Checkpoint:** Elements popped from the heap are in non-decreasing order regardless of the push order.

---

### Lab 2: Max-Heap — Top-3 Largest

**What you'll practise:** inverting the `Less` comparator to convert a min-heap into a max-heap.

**Task:**
Create a max-heap by changing only the `Less` method. Use it to find the three largest numbers in a `[]int` without sorting the entire slice.

**Steps:**
1. Define `type MaxIntHeap []int` and implement `heap.Interface` with `Less(i,j)` returning `h[i] > h[j]`
2. Push all numbers from the input slice into the max-heap via `heap.Push`
3. Pop three times — the three returned values are the three largest
4. Verify against the expected top-3

```go
type MaxIntHeap []int

func (h MaxIntHeap) Len() int           { return len(h) }
func (h MaxIntHeap) Less(i, j int) bool { return h[i] > h[j] } // inverted
func (h MaxIntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MaxIntHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *MaxIntHeap) Pop() any {
    old := *h; n := len(old)
    x := old[n-1]; *h = old[:n-1]
    return x
}

nums := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}
h := MaxIntHeap(append([]int{}, nums...))
heap.Init(&h)
top3 := []int{heap.Pop(&h).(int), heap.Pop(&h).(int), heap.Pop(&h).(int)}
```

**Expected output:**
```
Top-3: [9 6 5]
```

**Checkpoint:** The result is `[9, 6, 5]` for any ordering of the input. Only three pops are needed regardless of slice size.

---

### Lab 3: LRU Cache with `container/list`

**What you'll practise:** combining `container/list` for ordering with a `map` for O(1) lookup to build an LRU cache.

**Task:**
Implement an `LRUCache` with `Get(key string) (string, bool)` and `Set(key, value string)`. The cache has a fixed capacity; setting a new key when full evicts the least-recently-used entry.

**Steps:**
1. Define `type LRUCache struct { cap int; list *list.List; items map[string]*list.Element }`
2. Each list element stores a `entry{key, value string}` pair
3. `Get`: if found, move element to front and return value; else return `"", false`
4. `Set`: if key exists update and move to front; else push to front and evict the back element when over capacity
5. Write a test: create a capacity-2 cache, set 3 keys, verify the first is evicted

```go
import "container/list"

type entry struct{ key, value string }

type LRUCache struct {
    cap   int
    list  *list.List
    items map[string]*list.Element
}

func NewLRUCache(cap int) *LRUCache {
    return &LRUCache{cap: cap, list: list.New(), items: make(map[string]*list.Element)}
}

func (c *LRUCache) Get(key string) (string, bool) {
    if el, ok := c.items[key]; ok {
        c.list.MoveToFront(el)
        return el.Value.(*entry).value, true
    }
    return "", false
}
```

**Expected output:**
```
Get("a"): "", false   (evicted)
Get("b"): "B", true
Get("c"): "C", true
```

**Checkpoint:** After setting keys a, b, c (capacity 2), `Get("a")` returns `false` and both `Get("b")` and `Get("c")` return their values.

---

### Lab 4: Circular Buffer with `container/ring`

**What you'll practise:** using `container/ring` for a fixed-size circular buffer that overwrites the oldest element.

**Task:**
Create a circular buffer of capacity 5 using `ring.New(5)`. Write 10 values into it and show that only the last 5 survive. Then implement a sliding-window maximum over a stream of integers.

**Steps:**
1. Create `r := ring.New(5)` — a ring of 5 elements
2. Write values 1–10 into it, advancing `r = r.Next()` on each write
3. After all writes, iterate the ring with `r.Do` and print surviving values
4. Reset and implement `slidingMax(nums []int, k int) []int` using the ring to track a window

```go
import "container/ring"

r := ring.New(5)
for i := 1; i <= 10; i++ {
    r.Value = i
    r = r.Next()
}

// Print surviving values (6,7,8,9,10):
r.Do(func(v any) {
    fmt.Println(v)
})
```

**Expected output:**
```
Ring contents after 10 writes: [6 7 8 9 10]
```

**Checkpoint:** After writing values 1–10 to a capacity-5 ring, exactly values 6–10 remain.

---

### Lab 5: Graph — BFS Shortest Path

**What you'll practise:** representing a graph with an adjacency list and implementing BFS to find the shortest path between two nodes.

**Task:**
Define `type Graph map[string][]string`. Add `AddEdge`, `Neighbours`, and `BFS(start, end string) []string` returning the shortest path as a node slice.

**Steps:**
1. Implement `AddEdge(from, to string)` — directed edge (also add reverse for undirected)
2. Implement `BFS` using a queue and a `prev map[string]string` to reconstruct the path
3. Build a test graph: A→B, A→C, B→D, C→D, D→E
4. Find the shortest path from A to E and verify it is length 3 (A→B→D→E or A→C→D→E)

```go
type Graph map[string][]string

func (g Graph) AddEdge(from, to string) {
    g[from] = append(g[from], to)
}

func (g Graph) BFS(start, end string) []string {
    prev := map[string]string{start: ""}
    queue := []string{start}
    for len(queue) > 0 {
        node := queue[0]; queue = queue[1:]
        if node == end {
            return buildPath(prev, start, end)
        }
        for _, nb := range g[node] {
            if _, seen := prev[nb]; !seen {
                prev[nb] = node
                queue = append(queue, nb)
            }
        }
    }
    return nil // no path
}
```

**Expected output:**
```
BFS A→E: [A B D E]
BFS A→C: [A C]
BFS E→A: []  (directed — no path)
```

**Checkpoint:** `BFS` returns the path with the fewest hops. Unreachable nodes return `nil`.

---

### Lab 6: DFS and `HasPath`

**What you'll practise:** recursive DFS traversal and cycle detection using a visited set.

**Task:**
Add `DFS(start string) []string` returning nodes in DFS traversal order, and `HasPath(from, to string) bool` that returns true if any path exists between two nodes.

**Steps:**
1. Implement `DFS` recursively, tracking visited nodes in a `map[string]bool`
2. Implement `HasPath` using DFS or BFS — return true as soon as `to` is reached
3. Use the same A→B, A→C, B→D, C→D, D→E graph
4. Add a cycle (E→A) and verify DFS terminates without infinite recursion

```go
func (g Graph) DFS(start string) []string {
    var result []string
    visited := make(map[string]bool)
    var dfs func(node string)
    dfs = func(node string) {
        if visited[node] { return }
        visited[node] = true
        result = append(result, node)
        for _, nb := range g[node] {
            dfs(nb)
        }
    }
    dfs(start)
    return result
}

func (g Graph) HasPath(from, to string) bool {
    return g.BFS(from, to) != nil
}
```

**Expected output:**
```
DFS from A: [A B D E C]
HasPath(A, E): true
HasPath(E, B): false
```

**Checkpoint:** `DFS` visits every reachable node exactly once; adding a cycle (E→A) does not cause infinite recursion.

---

### Final Lab (Project): Data Structures Library

**What you'll practise:** combining all six labs into a tested, production-quality data structures library with a min-heap, LRU cache, and a BFS/DFS graph.

**Task:**
Implement and test the full data structures library. Each structure should have its own file and comprehensive tests.

**Steps:**
1. Implement a min-heap priority queue wrapping [`container/heap`](https://pkg.go.dev/container/heap)
2. Implement an LRU cache using [`container/list`](https://pkg.go.dev/container/list) + `map` (O(1) `Get` and `Set`)
3. Implement a generic `Graph[T comparable]` with `AddEdge`, `BFS`, `DFS`, `HasPath`
4. Write table-driven tests for each data structure

```go
// Example usage in main:
h := &MinIntHeap{5, 3, 8, 1}
heap.Init(h)
fmt.Println(heap.Pop(h)) // 1

cache := NewLRUCache(3)
cache.Set("x", "10")
v, _ := cache.Get("x")
fmt.Println(v) // 10

g := make(Graph[string])
g.AddEdge("A", "B"); g.AddEdge("B", "C")
fmt.Println(g.BFS("A", "C")) // [A B C]
```

**Expected output:**
```
MinHeap pop sequence: 1 3 5 8
LRU eviction works correctly
BFS path: [A B C]
DFS traversal: [A B C]
```

**Checkpoint:** `go test ./...` passes; all three data structures have at least 3 test cases each; `go vet ./...` is clean.

**Extension ideas:** implement Dijkstra's shortest-path on a weighted graph using `container/heap`.

## Official Documentation

- [`sort`](https://pkg.go.dev/sort) — Ints, Strings, Slice, Sort, SearchInts, sort.Interface
- [`container/heap`](https://pkg.go.dev/container/heap) — heap.Interface, Init, Push, Pop
- [`container/list`](https://pkg.go.dev/container/list) — doubly linked list operations
- [`container/ring`](https://pkg.go.dev/container/ring) — circular list (related container)
- [`fmt`](https://pkg.go.dev/fmt) — formatted output
- [Go Blog: The Go Programming Language Specification — Generics](https://go.dev/blog/intro-generics) — for the generic Graph type
