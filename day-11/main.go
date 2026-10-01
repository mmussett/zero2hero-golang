package main

import "fmt"

// Stack — LIFO
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T)       { s.items = append(s.items, v) }
func (s *Stack[T]) Len() int       { return len(s.items) }
func (s *Stack[T]) IsEmpty() bool  { return len(s.items) == 0 }

func (s *Stack[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Queue — FIFO
type Queue[T any] struct {
	items []T
	head  int
}

func (q *Queue[T]) Enqueue(v T)    { q.items = append(q.items, v) }
func (q *Queue[T]) Len() int       { return len(q.items) - q.head }
func (q *Queue[T]) IsEmpty() bool  { return q.Len() == 0 }

func (q *Queue[T]) Dequeue() (T, bool) {
	if q.IsEmpty() {
		var zero T
		return zero, false
	}
	v := q.items[q.head]
	q.head++
	return v, true
}

func (q *Queue[T]) Front() (T, bool) {
	if q.IsEmpty() {
		var zero T
		return zero, false
	}
	return q.items[q.head], true
}

// Higher-order generic functions
func Filter[T any](s []T, keep func(T) bool) []T {
	result := make([]T, 0, len(s))
	for _, v := range s {
		if keep(v) {
			result = append(result, v)
		}
	}
	return result
}

func Map[T, U any](s []T, f func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

func Reduce[T, U any](s []T, init U, f func(U, T) U) U {
	acc := init
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

func main() {
	fmt.Println("=== Stack (LIFO) ===")
	var s Stack[int]
	for _, v := range []int{1, 2, 3, 4, 5} {
		s.Push(v)
	}
	for !s.IsEmpty() {
		v, _ := s.Pop()
		fmt.Printf("%d ", v)
	}
	fmt.Println()

	fmt.Println("\n=== Queue (FIFO) ===")
	var q Queue[string]
	for _, w := range []string{"first", "second", "third"} {
		q.Enqueue(w)
	}
	for !q.IsEmpty() {
		v, _ := q.Dequeue()
		fmt.Printf("%q ", v)
	}
	fmt.Println()

	fmt.Println("\n=== Generic Functions ===")
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	doubled := Map(evens, func(n int) int { return n * 2 })
	sum := Reduce(doubled, 0, func(acc, n int) int { return acc + n })
	fmt.Printf("nums:    %v\n", nums)
	fmt.Printf("evens:   %v\n", evens)
	fmt.Printf("doubled: %v\n", doubled)
	fmt.Printf("sum:     %d\n", sum)
}
