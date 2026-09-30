# Day 11: Standard Library Collections

## Sorting

Go's `sort` package sorts slices in-place.

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

## container/heap — Priority Queue

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

## container/list — Doubly Linked List

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

## Day Project: Data Structures Library

Implement and test:
1. A min-heap priority queue wrapping `container/heap`
2. An LRU cache using `container/list` + `map` (O(1) get and put)
3. A generic `Graph[T comparable]` with `AddEdge`, `BFS`, `DFS`, `HasPath`

**Extension ideas:** implement Dijkstra's shortest-path on a weighted graph using `container/heap`.
