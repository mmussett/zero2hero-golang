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

---

## Labs

### Lab 1: Value vs Pointer — See the Difference

**What you'll practise:** Observing that a function with a value parameter receives a copy, while a pointer parameter can modify the original.

**Task:**
Write two functions — `modify(n int)` and `modifyPtr(n *int)` — call both on the same variable, and print the result after each call to prove which one changed the original.

**Steps:**
1. Write `modify(n int)` that does `n *= 10` — works on a local copy
2. Write `modifyPtr(n *int)` that does `*n *= 10` — dereferences and changes the original
3. Call both on the same `x := 5` and print `x` after each call

```go
package main

import "fmt"

func modify(n int) {
    n *= 10
    fmt.Printf("  inside modify: n=%d\n", n)
}

func modifyPtr(n *int) {
    *n *= 10
    fmt.Printf("  inside modifyPtr: *n=%d\n", *n)
}

func main() {
    x := 5
    fmt.Printf("before modify: x=%d\n", x)
    modify(x)
    fmt.Printf("after modify:  x=%d (unchanged)\n", x)

    fmt.Printf("before modifyPtr: x=%d\n", x)
    modifyPtr(&x)
    fmt.Printf("after modifyPtr:  x=%d (changed)\n", x)
}
```

**Expected output:**
```
before modify: x=5
  inside modify: n=50
after modify:  x=5 (unchanged)
before modifyPtr: x=5
  inside modifyPtr: *n=50
after modifyPtr:  x=50 (changed)
```

**Checkpoint:** After both calls, `x` is 50 (modified once by `modifyPtr`). Add a second call to `modifyPtr(&x)` — confirm `x` becomes 500.

---

### Lab 2: new() and & — Allocating Pointers

**What you'll practise:** Using `new(T)` to allocate a zeroed value, using `&value` to take the address of an existing value, and printing pointer addresses with `%p`.

**Task:**
Allocate two `int` pointers — one with `new`, one with `&` — and compare their addresses and values.

**Steps:**
1. Allocate `p1 := new(int)` — prints as a zero-valued int
2. Set `*p1 = 42`
3. Declare `n := 99; p2 := &n`
4. Print both pointers with `%p` and their dereferenced values with `%d`
5. Show that two separate `new(int)` calls produce different addresses

```go
package main

import "fmt"

func main() {
    // Allocate with new — returns a *int pointing to a zeroed int
    p1 := new(int)
    fmt.Printf("p1 address: %p  value before: %d\n", p1, *p1)
    *p1 = 42
    fmt.Printf("p1 address: %p  value after:  %d\n", p1, *p1)

    // Take address of an existing variable
    n := 99
    p2 := &n
    fmt.Printf("p2 address: %p  value: %d\n", p2, *p2)

    // Modifying via pointer changes the original variable
    *p2 = 100
    fmt.Printf("n after *p2=100: %d\n", n)

    // Two separate allocations always have different addresses
    a, b := new(int), new(int)
    fmt.Printf("a=%p  b=%p  same? %v\n", a, b, a == b)
}
```

**Expected output:**
```
p1 address: 0xc000...  value before: 0
p1 address: 0xc000...  value after:  42
p2 address: 0xc000...  value: 99
n after *p2=100: 100
a=0xc000...  b=0xc000...  same? false
```

**Checkpoint:** Addresses will differ each run — that is correct. `a == b` is always `false`. Setting `*p2` changes `n` — they point to the same memory.

---

### Lab 3: Pointer vs Value Receivers on a Struct

**What you'll practise:** Seeing why a method that must mutate state needs a pointer receiver, and what happens when you accidentally use a value receiver for mutation.

**Task:**
Build a `Counter` struct with both a value-receiver `Reset()` (intentionally broken) and a pointer-receiver `Increment()` (correct). Compare the results.

**Steps:**
1. Define `type Counter struct { n int }`
2. Write `(c Counter) BrokenReset()` that sets `c.n = 0` — operates on a copy
3. Write `(c *Counter) Increment()` that does `c.n++` — correct pointer receiver
4. Write `(c Counter) Value() int` — safe read-only value receiver
5. Call `BrokenReset` after incrementing — show it has no effect

```go
package main

import "fmt"

type Counter struct{ n int }

// BrokenReset uses a VALUE receiver — it resets a copy, not the original.
func (c Counter) BrokenReset() {
    c.n = 0
    fmt.Printf("  BrokenReset: inside c.n=%d\n", c.n)
}

// Increment uses a POINTER receiver — correctly mutates the original.
func (c *Counter) Increment() { c.n++ }

// Value is a read-only value receiver — safe and cheap.
func (c Counter) Value() int { return c.n }

func main() {
    c := Counter{}
    c.Increment()
    c.Increment()
    c.Increment()
    fmt.Printf("after 3 increments: %d\n", c.Value())

    c.BrokenReset()
    fmt.Printf("after BrokenReset:  %d (still %d!)\n", c.Value(), c.Value())

    // Correct way: pointer receiver reset
    c.n = 0
    fmt.Printf("after manual reset: %d\n", c.Value())
}
```

**Expected output:**
```
after 3 increments: 3
  BrokenReset: inside c.n=0
after BrokenReset:  3 (still 3!)
after manual reset: 0
```

**Checkpoint:** `BrokenReset` does not change the real `c.n`. Fix it by changing the receiver to `*Counter` and confirm `c.Value()` returns 0 after calling it.

---

### Lab 4: Nil Pointer — Panic, Recover, and the Safe Pattern

**What you'll practise:** Deliberately triggering a nil-pointer panic, catching it with `recover()` in a `defer`, and applying the correct nil-check pattern.

**Task:**
Dereference a nil pointer inside a function that uses `defer`/`recover` to catch the panic gracefully. Then show the idiomatic safe-check approach.

**Steps:**
1. Write `safeDeref(p *int) (val int, err error)` — uses `defer`/`recover` to catch the panic
2. Call it with `nil` — observe it returns an error instead of crashing
3. Call it with a valid pointer — observe it returns the value
4. Show the simple nil-check as the preferred real-world pattern

```go
package main

import "fmt"

func safeDeref(p *int) (val int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("caught panic: %v", r)
        }
    }()
    val = *p // panics if p is nil
    return
}

func main() {
    // Nil pointer — caught by recover
    v, err := safeDeref(nil)
    fmt.Printf("nil deref: val=%d err=%v\n", v, err)

    // Valid pointer — works fine
    n := 42
    v, err = safeDeref(&n)
    fmt.Printf("valid deref: val=%d err=%v\n", v, err)

    // Idiomatic safe pattern — no recover needed
    var p *int
    if p != nil {
        fmt.Println("value:", *p)
    } else {
        fmt.Println("p is nil — skip deref")
    }
}
```

**Expected output:**
```
nil deref: val=0 err=caught panic: runtime error: invalid memory address or nil pointer dereference
valid deref: val=42 err=<nil>
p is nil — skip deref
```

**Checkpoint:** `recover()` stops the panic from propagating. The idiomatic pattern (nil check before deref) is always preferred over recover — recover is a last resort, not normal error handling.

---

### Lab 5: Pointer to Interface — Why You Almost Never Need It

**What you'll practise:** Understanding why `*SomeInterface` is almost always wrong, and seeing the pattern that actually works.

**Task:**
Define a `Stringer` interface. Store a concrete type in it. Show that `*Stringer` breaks polymorphism, and demonstrate the correct approach.

**Steps:**
1. Define `type Stringer interface { String() string }`
2. Define `type Foo struct{ msg string }` with a pointer-receiver `String()`
3. Assign `&Foo{"hello"}` to a `Stringer` variable — works
4. Try using `*Stringer` as a function parameter — show the issue
5. Show the correct pattern: pass `Stringer` directly (interface value, not pointer-to-interface)

```go
package main

import "fmt"

type Stringer interface {
    String() string
}

type Foo struct{ msg string }

func (f *Foo) String() string { return "Foo: " + f.msg }

// WRONG pattern — almost never what you want
func printWrong(s *Stringer) {
    fmt.Println((*s).String()) // must double-dereference
}

// CORRECT pattern — pass the interface directly
func printRight(s Stringer) {
    fmt.Println(s.String())
}

func main() {
    var s Stringer = &Foo{"hello"}

    printRight(s) // clean and idiomatic

    printWrong(&s) // works but ugly — the interface is already a reference type

    // Why *Stringer is wrong: you lose the ability to pass different types
    // because &s has type *Stringer, not *Foo or *Bar
    fmt.Printf("type of s:  %T\n", s)
    fmt.Printf("type of &s: %T\n", &s)
}
```

**Expected output:**
```
Foo: hello
Foo: hello
type of s:  *main.Foo
type of &s: *main.Stringer
```

**Checkpoint:** Interfaces are already reference types (they hold a pointer to data internally). Passing `*Stringer` adds an unnecessary indirection. The rule: pass interfaces by value, not by pointer.

---

### Lab 6 (Final): Memory Explorer

**What you'll practise:** Combining all pointer concepts — address operators, pointer receivers, copy semantics, and nil checks — in one complete program.

**Task:**
Write `day-04/main.go` that demonstrates the full range of pointer behaviour: address printing, copy vs share semantics, pointer receiver mutation, and safe nil handling.

**Steps:**
1. Declare an `int`, take its address with `&`, and print the pointer with `%p`
2. Define `type Box struct { value int }` with `Set(v int)` (pointer receiver) and `Get() int` (value receiver)
3. Show struct copy semantics: `b2 := b1` creates an independent copy; `p := &b1` shares it
4. Show `b1.Set(99)` updates `b1` but not `b2` (the copy)
5. Demonstrate a nil `*Box` check before calling `Get()`

```go
package main

import "fmt"

type Box struct{ value int }

func (b *Box) Set(v int) { b.value = v }
func (b Box) Get() int   { return b.value }

func main() {
    // Address and pointer printing
    x := 100
    p := &x
    fmt.Printf("x=%d  &x=%p  *p=%d\n", x, p, *p)

    // Struct copy vs pointer share
    b1 := Box{value: 1}
    b2 := b1    // copy — independent
    pb := &b1   // pointer — shares b1

    b1.Set(99)
    fmt.Printf("b1.Get()=%d (modified)\n", b1.Get())
    fmt.Printf("b2.Get()=%d (copy — unchanged)\n", b2.Get())
    fmt.Printf("pb.Get()=%d (pointer — same as b1)\n", pb.Get())

    // Nil pointer check
    var nilBox *Box
    if nilBox != nil {
        fmt.Println("value:", nilBox.Get())
    } else {
        fmt.Println("nilBox is nil — skipping deref")
    }
}
```

**Expected output:**
```
x=100  &x=0xc000...  *p=100
b1.Get()=99 (modified)
b2.Get()=1 (copy — unchanged)
pb.Get()=99 (pointer — same as b1)
nilBox is nil — skipping deref
```

**Checkpoint:** Run `go run .` and observe the real memory address printed. Confirm `b2.Get()` is still `1` even after `b1.Set(99)`. `go vet .` passes with no warnings.

---

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
