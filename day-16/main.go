package main

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
)

type cacheEntry struct {
	key string
	val any
}

type LRUCache struct {
	mu     sync.RWMutex
	cap    int
	list   *list.List
	items  map[string]*list.Element
	hits   atomic.Int64
	misses atomic.Int64
}

func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		cap:   cap,
		list:  list.New(),
		items: make(map[string]*list.Element),
	}
}

func (c *LRUCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.list.MoveToFront(el)
		c.hits.Add(1)
		return el.Value.(*cacheEntry).val, true
	}
	c.misses.Add(1)
	return nil, false
}

func (c *LRUCache) Put(key string, val any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.list.MoveToFront(el)
		el.Value.(*cacheEntry).val = val
		return
	}
	if c.list.Len() == c.cap {
		back := c.list.Back()
		if back != nil {
			c.list.Remove(back)
			delete(c.items, back.Value.(*cacheEntry).key)
		}
	}
	el := c.list.PushFront(&cacheEntry{key, val})
	c.items[key] = el
}

func (c *LRUCache) Stats() (hits, misses int64) {
	return c.hits.Load(), c.misses.Load()
}

func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list.Len()
}

func main() {
	cache := NewLRUCache(3)

	ops := []struct {
		op  string
		key string
		val any
	}{
		{"put", "a", "apple"},
		{"put", "b", "banana"},
		{"put", "c", "cherry"},
		{"get", "a", nil},          // hit — a moves to front
		{"put", "d", "date"},       // evicts b (LRU)
		{"get", "b", nil},          // miss — b was evicted
		{"get", "c", nil},          // hit
		{"put", "e", "elderberry"}, // evicts a (LRU)
		{"get", "a", nil},          // miss
	}

	for _, op := range ops {
		switch op.op {
		case "put":
			cache.Put(op.key, op.val)
			fmt.Printf("PUT %-4q = %-12v (size=%d)\n", op.key, op.val, cache.Len())
		case "get":
			if v, ok := cache.Get(op.key); ok {
				fmt.Printf("GET %-4q → %-12v (hit)\n", op.key, v)
			} else {
				fmt.Printf("GET %-4q → (miss)\n", op.key)
			}
		}
	}

	hits, misses := cache.Stats()
	fmt.Printf("\nStats: %d hits, %d misses (%.0f%% hit rate)\n",
		hits, misses, float64(hits)/float64(hits+misses)*100)
}
