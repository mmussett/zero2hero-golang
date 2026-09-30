# Day 04: Pointers and Value Semantics

## Core Concept: Values by Default

Go passes everything by value. When you pass a struct to a function, the function receives a copy. Pointers opt into shared access.

```go
func increment(x int) { x++ }     // caller's x unchanged

func incrementPtr(x *int) { *x++ } // caller's x is modified

n := 5
increment(n)    // n is still 5
incrementPtr(&n) // n is now 6
```

## Pointer Operators

| Operator | Meaning |
|----------|---------|
| `&x` | address of x — produces `*T` |
| `*p` | value at address p — dereference |
| `new(T)` | allocate zeroed T, return `*T` |

## When to Use Pointers

1. **Mutation** — the function or method must modify the original value
2. **Large structs** — avoid expensive copies on every call
3. **Optional values** — `*T` can be `nil` to represent "absent"

Do **not** use a pointer just because you're used to reference semantics from another language. Go's escape analysis often heap-allocates anyway — trust the compiler.

## Nil Pointers

The zero value of any pointer type is `nil`. Dereferencing a nil pointer panics:

```go
var p *int  // nil
*p = 5      // panic: runtime error: invalid memory address
```

Always check `p != nil` before dereferencing pointers that may be nil.

## Pointers and Structs

```go
type Counter struct{ n int }

func (c *Counter) Inc() { c.n++ }        // pointer receiver — modifies c
func (c Counter) Value() int { return c.n } // value receiver — read-only

c := Counter{}
c.Inc()              // Go auto-takes address: (&c).Inc()
fmt.Println(c.Value()) // 1
```

## Day Project: Memory Explorer

Write a program that:
1. Declares an `int`, takes its address with `&`, and prints the pointer value (`%p` verb)
2. Defines a `Box` struct with a `value int`; writes `Set(v int)` (pointer receiver) and `Get() int` (value receiver)
3. Shows that assigning a struct copies it, but assigning a pointer shares it
4. Demonstrates a nil pointer check before dereferencing

Run it and observe the memory addresses printed.

## Official Documentation

- [`fmt`](https://pkg.go.dev/fmt) — the `%p` verb for printing pointer addresses
- [Language Spec: Pointer types](https://go.dev/ref/spec#Pointer_types) — `*T` pointer syntax
- [Language Spec: Address operators](https://go.dev/ref/spec#Address_operators) — `&` and `*` operators
- [Language Spec: Allocation](https://go.dev/ref/spec#Allocation) — the `new` built-in
- [Effective Go: Pointers vs. Values](https://go.dev/doc/effective_go#pointers_vs_values) — when to use each
- [Go Tour: Pointers](https://go.dev/tour/moretypes/1) — interactive pointer tour
