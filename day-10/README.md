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

## Labs

### Lab 1: Generic `Min` and `Max`

**What you'll practise:** writing a generic function with an `Ordered` constraint and calling it with multiple types.

**Task:**
Write `Min[T constraints.Ordered](a, b T) T` and `Max[T constraints.Ordered](a, b T) T`. Call each with `int`, `float64`, and `string` to confirm the type parameter is inferred correctly.

**Steps:**
1. Import `"cmp"` (Go 1.21+) or define your own `Ordered` constraint using `~int | ~float64 | ~string | ...`
2. Implement `Min` returning the smaller of two values
3. Implement `Max` returning the larger of two values
4. In `main`, call both with at least three different types

```go
// Using the built-in cmp package (Go 1.21+)
import "cmp"

func Min[T cmp.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

**Expected output:**
```
Min(3, 7) = 3
Min(3.14, 2.71) = 2.71
Min("banana", "apple") = apple
Max(3, 7) = 7
```

**Checkpoint:** The compiler infers the type parameter without explicit instantiation (`Min(3, 7)` not `Min[int](3, 7)`).

---

### Lab 2: Generic `Stack[T any]`

**What you'll practise:** generic types — a parameterised struct with multiple methods.

**Task:**
Implement `Stack[T any]` with `Push(v T)`, `Pop() (T, bool)`, `Peek() (T, bool)`, `Len() int`, and `IsEmpty() bool`. Test it with both `int` and `string`.

**Steps:**
1. Declare `type Stack[T any] struct { items []T }`
2. Implement all five methods; `Pop` and `Peek` must return `(T, bool)` — return the zero value and `false` when empty
3. Write a test that pushes three ints, peeks, pops all three, and verifies order is LIFO
4. Write a second test using `string`

```go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    top := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return top, true
}

func (s *Stack[T]) Peek() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Len() int    { return len(s.items) }
func (s *Stack[T]) IsEmpty() bool { return len(s.items) == 0 }
```

**Expected output:**
```
Pushed: 1, 2, 3
Peek: 3 (len=3)
Pop: 3, 2, 1
IsEmpty: true
```

**Checkpoint:** `Pop` on an empty stack returns `(zero, false)` without panicking.

---

### Lab 3: Generic `Queue[T any]`

**What you'll practise:** a FIFO generic type implemented as a slice — understanding head-index vs re-slicing tradeoffs.

**Task:**
Implement `Queue[T any]` with `Enqueue(v T)`, `Dequeue() (T, bool)`, `Front() (T, bool)`, `Len() int`, and `IsEmpty() bool`.

**Steps:**
1. Declare `type Queue[T any] struct { items []T }`
2. `Enqueue` appends to the back; `Dequeue` removes from the front using re-slicing
3. Write a test: enqueue 5 items, dequeue 3, verify FIFO order and remaining `Len()`
4. Confirm `Dequeue` on an empty queue returns `(zero, false)`

```go
type Queue[T any] struct {
    items []T
}

func (q *Queue[T]) Enqueue(v T) { q.items = append(q.items, v) }

func (q *Queue[T]) Dequeue() (T, bool) {
    if len(q.items) == 0 {
        var zero T
        return zero, false
    }
    front := q.items[0]
    q.items = q.items[1:]
    return front, true
}

func (q *Queue[T]) Front() (T, bool) {
    if len(q.items) == 0 {
        var zero T
        return zero, false
    }
    return q.items[0], true
}

func (q *Queue[T]) Len() int     { return len(q.items) }
func (q *Queue[T]) IsEmpty() bool { return len(q.items) == 0 }
```

**Expected output:**
```
Enqueued: a, b, c, d, e
Dequeue: a, b, c
Remaining: 2  Front: d
```

**Checkpoint:** Dequeue returns items in FIFO order; `Len()` decrements correctly after each dequeue.

---

### Lab 4: `Filter` and `Map`

**What you'll practise:** generic higher-order functions that work with any slice type.

**Task:**
Write `Filter[T any](slice []T, pred func(T) bool) []T` and `Map[T, U any](slice []T, f func(T) U) []U`. Demonstrate each with at least two different type combinations.

**Steps:**
1. Implement `Filter` — return a new slice containing only elements for which `pred` returns true
2. Implement `Map` — apply `f` to every element and return the result slice
3. Use `Filter` to keep only even ints from `[]int{1,2,3,4,5,6}`
4. Use `Map` to convert `[]string{"hello","world"}` to `[]int` of lengths
5. Chain them: filter a string slice to words longer than 3 chars, then map to uppercase

```go
func Filter[T any](slice []T, pred func(T) bool) []T {
    var out []T
    for _, v := range slice {
        if pred(v) {
            out = append(out, v)
        }
    }
    return out
}

func Map[T, U any](slice []T, f func(T) U) []U {
    out := make([]U, len(slice))
    for i, v := range slice {
        out[i] = f(v)
    }
    return out
}
```

**Expected output:**
```
Even ints: [2 4 6]
Word lengths: [5 5]
Long+Upper: [HELLO WORLD]
```

**Checkpoint:** Neither function allocates an output slice when the input is empty.

---

### Lab 5: `Reduce`

**What you'll practise:** a generic accumulator that transforms a slice into a single value using an initial value and a combining function.

**Task:**
Write `Reduce[T, U any](slice []T, initial U, f func(U, T) U) U`. Use it to sum a `[]int`, find the maximum, and concatenate a `[]string`.

**Steps:**
1. Implement `Reduce` iterating over `slice`, updating an accumulator with `f(acc, v)`
2. Use it with `func(acc, v int) int { return acc + v }` to sum `[]int{1,2,3,4,5}`
3. Use it to find the max of `[]int{3,1,4,1,5,9,2,6}`
4. Use it to concatenate `[]string{"Go","is","fun"}` into `"Go is fun"`

```go
func Reduce[T, U any](slice []T, initial U, f func(U, T) U) U {
    acc := initial
    for _, v := range slice {
        acc = f(acc, v)
    }
    return acc
}

// Sum
sum := Reduce([]int{1, 2, 3, 4, 5}, 0, func(acc, v int) int { return acc + v })

// Join strings
joined := Reduce([]string{"Go", "is", "fun"}, "", func(acc, v string) string {
    if acc == "" { return v }
    return acc + " " + v
})
```

**Expected output:**
```
Sum: 15
Max: 9
Joined: Go is fun
```

**Checkpoint:** `Reduce` on an empty slice returns `initial` unchanged.

---

### Lab 6: Type Constraint — `Number` and `Sum`

**What you'll practise:** defining a custom type constraint with the `~` underlying-type prefix and writing a numeric generic function.

**Task:**
Define a `Number` constraint covering all integer and float types using `~` prefixes. Write `Sum[T Number](slice []T) T` and verify it works with `int`, `float64`, and a custom `type Celsius float64`.

**Steps:**
1. Define `type Number interface { ~int | ~int32 | ~int64 | ~float32 | ~float64 }`
2. Implement `Sum[T Number](slice []T) T` using a zero value and accumulation
3. Test with `[]int{1,2,3}`, `[]float64{1.1, 2.2}`, and `[]Celsius{98.6, 37.0}`
4. Try removing the `~` and observe the compile error with `Celsius` — then restore it

```go
type Number interface {
    ~int | ~int32 | ~int64 | ~float32 | ~float64
}

func Sum[T Number](slice []T) T {
    var total T
    for _, v := range slice {
        total += v
    }
    return total
}

type Celsius float64

// These all compile because of the ~ prefix:
fmt.Println(Sum([]int{1, 2, 3}))           // 6
fmt.Println(Sum([]float64{1.1, 2.2, 3.3})) // 6.6
fmt.Println(Sum([]Celsius{98.6, 37.0}))    // 135.6
```

**Expected output:**
```
int sum: 6
float64 sum: 6.6
Celsius sum: 135.6
```

**Checkpoint:** `Sum` works with a named type whose underlying type is in the constraint. Removing `~` causes a compile error for `Celsius`.

---

### Final Lab (Project): Generic Stack, Queue, Filter, Map, Reduce

**What you'll practise:** assembling all generic building blocks into a complete, tested data-structures package.

**Task:**
Implement and test the full generic toolkit: `Stack[T]`, `Queue[T]`, `Filter`, `Map`, `Reduce`, and the `Number`/`Sum` utilities. Wire them together in a `main` that demonstrates a realistic pipeline.

**Steps:**
1. Implement `Stack[T any]` — LIFO with `Push`, `Pop`, `Peek`, `Len`, `IsEmpty`
2. Implement `Queue[T any]` — FIFO with `Enqueue`, `Dequeue`, `Front`, `Len`, `IsEmpty`
3. Implement generic `Filter[T any]`, `Map[T, U any]`, `Reduce[T, U any]`
4. Write tests for each using at least two different type instantiations
5. In `main`, demonstrate a pipeline: load words into a `Queue`, filter long ones, map to uppercase, reduce to a single string

```go
words := []string{"go", "generics", "are", "powerful", "and", "elegant"}
long  := Filter(words, func(w string) bool { return len(w) > 3 })
upper := Map(long, strings.ToUpper)
result := Reduce(upper, "", func(acc, w string) string {
    if acc == "" { return w }
    return acc + " " + w
})
fmt.Println(result)
```

**Expected output:**
```
GENERICS ARE POWERFUL ELEGANT
```

**Checkpoint:** `go test ./...` passes; every generic function is exercised with at least two type parameters in tests.

**Extension ideas:** implement a `Set[T comparable]` with `Add`, `Contains`, `Remove`, `Union`, `Intersection`.

## Official Documentation

- [`strings`](https://pkg.go.dev/strings) — `ToUpper` and other functions used in generic examples
- [Language Spec: Type parameters](https://go.dev/ref/spec#Type_parameter_declarations) — type parameter syntax
- [Language Spec: Type constraints](https://go.dev/ref/spec#Interface_types) — constraint interfaces
- [Go Blog: An Introduction to Generics](https://go.dev/blog/intro-generics) — overview of Go generics
- [Go Blog: When to use generics](https://go.dev/blog/when-generics) — guidance on generics vs interfaces
- [Go Tour: Generics](https://go.dev/tour/generics/1) — interactive generics tour
