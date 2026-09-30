# Day 10: Generics

## Core Concept: Type Parameters

Generics (added in Go 1.18) let you write functions and types that work over multiple types while remaining type-safe.

```go
func Map[T, U any](s []T, f func(T) U) []U {
    result := make([]U, len(s))
    for i, v := range s {
        result[i] = f(v)
    }
    return result
}

doubled := Map([]int{1, 2, 3}, func(x int) int { return x * 2 })
upper   := Map([]string{"a","b"}, strings.ToUpper)
```

## Constraints

A constraint is an interface that restricts which types a type parameter may be:

```go
// comparable — supports == and !=
func Contains[T comparable](s []T, v T) bool { ... }

// Built-in constraint from golang.org/x/exp/constraints or defined inline:
type Number interface {
    int | int8 | int16 | int32 | int64 |
    float32 | float64
}

func Sum[T Number](s []T) T {
    var total T
    for _, v := range s { total += v }
    return total
}
```

The `~` prefix means "any type whose underlying type is T":

```go
type Celsius float64

type Temperature interface { ~float64 }

func Max[T Temperature](a, b T) T { if a > b { return a }; return b }
// Now works with both float64 and Celsius
```

## Generic Types

```go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(v T)        { s.items = append(s.items, v) }
func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T; return zero, false
    }
    top := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return top, true
}
func (s *Stack[T]) Peek() (T, bool) { ... }
func (s *Stack[T]) Len() int        { return len(s.items) }
```

## When to Use Generics vs Interfaces

- Use **generics** when the algorithm is the same for all types and you need type safety at the call site
- Use **interfaces** when behaviour differs per type (polymorphism)
- A function that only needs `any` does not benefit from generics

## Day Project: Generic Stack and Queue

Implement:
- `Stack[T any]` — LIFO with `Push`, `Pop`, `Peek`, `Len`, `IsEmpty`
- `Queue[T any]` — FIFO with `Enqueue`, `Dequeue`, `Front`, `Len`, `IsEmpty`
- Generic `Filter[T any]([]T, func(T) bool) []T`
- Generic `Reduce[T, U any]([]T, U, func(U, T) U) U`
- Tests for each

**Extension ideas:** implement a `Set[T comparable]` with `Add`, `Contains`, `Remove`, `Union`, `Intersection`.
