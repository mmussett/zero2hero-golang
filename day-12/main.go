package main

import (
	"container/heap"
	"container/list"
	"fmt"
)

// ── Min-Heap Priority Queue ──────────────────────────────────────────────────

type MinHeap []int

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MinHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *MinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// ── LRU Cache ────────────────────────────────────────────────────────────────

type lruEntry struct{ key, val string }

type LRUCache struct {
	cap   int
	list  *list.List
	items map[string]*list.Element
}

func NewLRU(cap int) *LRUCache {
	return &LRUCache{
		cap:   cap,
		list:  list.New(),
		items: make(map[string]*list.Element),
	}
}

func (c *LRUCache) Get(key string) (string, bool) {
	if el, ok := c.items[key]; ok {
		c.list.MoveToFront(el)
		return el.Value.(*lruEntry).val, true
	}
	return "", false
}

func (c *LRUCache) Put(key, val string) {
	if el, ok := c.items[key]; ok {
		c.list.MoveToFront(el)
		el.Value.(*lruEntry).val = val
		return
	}
	if c.list.Len() == c.cap {
		back := c.list.Back()
		c.list.Remove(back)
		delete(c.items, back.Value.(*lruEntry).key)
	}
	el := c.list.PushFront(&lruEntry{key, val})
	c.items[key] = el
}

// ── Graph with BFS / DFS / HasPath ──────────────────────────────────────────

type Graph map[string][]string

func (g Graph) AddEdge(from, to string) {
	g[from] = append(g[from], to)
	g[to] = append(g[to], from)
}

func (g Graph) BFS(start string) []string {
	visited := map[string]bool{start: true}
	queue := []string{start}
	var order []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)
		for _, nb := range g[node] {
			if !visited[nb] {
				visited[nb] = true
				queue = append(queue, nb)
			}
		}
	}
	return order
}

func (g Graph) DFS(start string) []string {
	var order []string
	visited := map[string]bool{}
	var dfs func(n string)
	dfs = func(n string) {
		visited[n] = true
		order = append(order, n)
		for _, nb := range g[n] {
			if !visited[nb] {
				dfs(nb)
			}
		}
	}
	dfs(start)
	return order
}

func (g Graph) HasPath(from, to string) bool {
	visited := map[string]bool{}
	var dfs func(n string) bool
	dfs = func(n string) bool {
		if n == to {
			return true
		}
		visited[n] = true
		for _, nb := range g[n] {
			if !visited[nb] && dfs(nb) {
				return true
			}
		}
		return false
	}
	return dfs(from)
}

func main() {
	// Min-heap
	fmt.Println("=== Min-Heap ===")
	h := &MinHeap{5, 3, 8, 1, 9, 2}
	heap.Init(h)
	heap.Push(h, 4)
	for h.Len() > 0 {
		fmt.Printf("%d ", heap.Pop(h))
	}
	fmt.Println()

	// LRU cache
	fmt.Println("\n=== LRU Cache (capacity 3) ===")
	lru := NewLRU(3)
	lru.Put("a", "apple")
	lru.Put("b", "banana")
	lru.Put("c", "cherry")
	lru.Get("a")        // moves a to front; b becomes LRU
	lru.Put("d", "date") // evicts b
	for _, k := range []string{"a", "b", "c", "d"} {
		if v, ok := lru.Get(k); ok {
			fmt.Printf("  %s → %s\n", k, v)
		} else {
			fmt.Printf("  %s → (evicted)\n", k)
		}
	}

	// Graph
	fmt.Println("\n=== Graph ===")
	g := Graph{}
	g.AddEdge("A", "B")
	g.AddEdge("A", "C")
	g.AddEdge("B", "D")
	g.AddEdge("C", "D")
	g.AddEdge("D", "E")
	fmt.Printf("BFS from A: %v\n", g.BFS("A"))
	fmt.Printf("DFS from A: %v\n", g.DFS("A"))
	fmt.Printf("Path A→E:  %v\n", g.HasPath("A", "E"))
	fmt.Printf("Path E→A:  %v\n", g.HasPath("E", "A"))
}
