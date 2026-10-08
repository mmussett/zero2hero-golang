# Day 03: Functions, Structs, and Error Handling

## What You'll Learn Today

- How to define and call functions with various parameter and return styles
- Multiple return values and why they replace exceptions in Go
- Named return values and when to use them
- Variadic functions for variable-length argument lists
- Anonymous functions, closures, and first-class functions
- `defer` for deferred execution and resource cleanup
- `panic` and `recover` for handling unrecoverable situations
- Recursion, its tradeoffs, and when to prefer iteration
- Structs as Go's primary data-grouping construct
- Methods with value and pointer receivers

---

## 1. Function Basics

The syntax for a Go function is:

```
func name(parameters) returnType {
    body
}
```

A function with no parameters and no return value:

```go
func sayHello() {
    fmt.Println("Hello, Gopher!")
}
```

A function with one parameter and one return value:

```go
func double(n int) int {
    return n * 2
}
```

Calling a function:

```go
sayHello()         // Hello, Gopher!
fmt.Println(double(7))  // 14
```

Functions are first-class values in Go — they can be assigned to variables, passed as arguments, and returned from other functions. You will explore all of these in later sections.

---

## 2. Parameters

**Single parameter:**

```go
func greet(name string) string {
    return "Hello, " + name + "!"
}
```

**Multiple parameters with explicit types:**

```go
func add(a int, b int) int {
    return a + b
}
```

**Shared-type shorthand:** when consecutive parameters share a type, only the last needs the type annotation.

```go
func add(a, b int) int {  // same as add(a int, b int) int
    return a + b
}
```

**Mixed types:**

```go
func repeat(s string, n int) string {
    result := ""
    for i := 0; i < n; i++ {
        result += s
    }
    return result
}
```

**Parameters are passed by value.** The function receives a copy of each argument; mutating the copy does not affect the caller's variable.

```go
func triple(n int) {
    n = n * 3 // modifies only the local copy
}

func main() {
    x := 5
    triple(x)
    fmt.Println(x) // still 5
}
```

**To mutate the caller's variable, pass a pointer:**

```go
func tripleInPlace(n *int) {
    *n = *n * 3 // dereference and mutate
}

func main() {
    x := 5
    tripleInPlace(&x) // pass the address of x
    fmt.Println(x)    // 15
}
```

Pointer parameters are covered more fully on Day 06. For now, remember: pass a pointer when you need the function to modify the original.

---

## 3. Multiple Return Values

Go functions can return more than one value. The most common pattern is returning a result alongside an error:

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}
```

The caller handles both values:

```go
result, err := divide(10, 2)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result) // 5
```

**Ignoring a return value with `_`:**

```go
result, _ := divide(10, 2) // discard the error (only do this when you are certain no error can occur)
```

**Why multiple returns replace exceptions:**
In most languages, errors are communicated via exceptions that can be raised anywhere and caught anywhere, making control flow implicit. Go makes error handling explicit: a function that can fail says so in its signature, and the caller is forced to acknowledge the error at the call site. This results in code that is easier to trace and reason about.

---

## 4. Named Return Values

You can give names to return values. Named returns create pre-declared variables that are initialised to their zero values. A bare `return` returns whatever those variables currently hold.

```go
func minMax(nums []int) (min, max int) {
    min, max = nums[0], nums[0]
    for _, n := range nums[1:] {
        if n < min {
            min = n
        }
        if n > max {
            max = n
        }
    }
    return // returns min and max
}
```

**When to use named returns:**
- Short functions where the names serve as documentation
- Functions where the bare `return` makes the logic clearer

**When to avoid named returns:**
- Long functions — a bare `return` in a 50-line function hides what is being returned; prefer explicit `return min, max`
- Named returns combined with `defer` can produce surprising results for beginners (a deferred function can modify a named return variable before it reaches the caller)

---

## 5. Variadic Functions

A variadic parameter accepts zero or more arguments of a given type. Inside the function, it is a slice.

```go
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}
```

**Calling with individual values:**

```go
fmt.Println(sum(1, 2, 3))    // 6
fmt.Println(sum(10, 20))     // 30
fmt.Println(sum())           // 0 (zero arguments is valid)
```

**Spreading a slice into a variadic call:** append `...` after the slice.

```go
nums := []int{1, 2, 3, 4, 5}
fmt.Println(sum(nums...)) // 15
```

**Rules:**
- The variadic parameter must be the last (or only) parameter.
- You cannot have two variadic parameters.

`fmt.Println` is itself variadic — its signature is `func Println(a ...any) (n int, err error)`.

```go
fmt.Println("one", "two", "three") // passes three arguments to the variadic parameter
```

**Enforcing at least one argument** by combining a required first parameter with a variadic rest:

```go
func max(first int, rest ...int) int {
    m := first
    for _, n := range rest {
        if n > m {
            m = n
        }
    }
    return m
}

fmt.Println(max(3))          // 3
fmt.Println(max(3, 1, 4, 1, 5, 9)) // 9
// max() would not compile — first is required
```

---

## 6. Anonymous Functions

An anonymous function (function literal) is a function defined without a name.

```go
f := func(x int) int {
    return x * x
}
fmt.Println(f(5)) // 25
```

**Immediately Invoked Function Expression (IIFE):** define and call in one step.

```go
result := func(x int) int {
    return x * x
}(5)
fmt.Println(result) // 25
```

**Closures:** an anonymous function that references variables from the enclosing scope. The function captures those variables — not copies of their values at the time of creation, but references to the variables themselves.

```go
func makeCounter() func() int {
    count := 0
    return func() int {
        count++ // captures and modifies count from makeCounter's scope
        return count
    }
}

func main() {
    counter := makeCounter()
    fmt.Println(counter()) // 1
    fmt.Println(counter()) // 2
    fmt.Println(counter()) // 3

    // A second counter has its own independent count variable.
    counter2 := makeCounter()
    fmt.Println(counter2()) // 1
}
```

**Passing functions as arguments:** Go functions are values, so any `func` type can be a parameter.

```go
func apply(nums []int, fn func(int) int) []int {
    result := make([]int, len(nums))
    for i, n := range nums {
        result[i] = fn(n)
    }
    return result
}

func main() {
    nums := []int{1, 2, 3, 4, 5}
    doubled := apply(nums, func(n int) int { return n * 2 })
    fmt.Println(doubled) // [2 4 6 8 10]
}
```

---

## 7. defer — Deferred Function Calls

`defer` schedules a function call to run when the surrounding function returns, regardless of how it returns (normally, via `return`, or via `panic`).

```go
func example() {
    defer fmt.Println("third")  // runs last
    defer fmt.Println("second") // runs second
    defer fmt.Println("first")  // runs first
    fmt.Println("body")
}
```

Output:
```
body
first
second
third
```

Deferred calls execute in **LIFO (last in, first out)** order — the last `defer` statement reached runs first.

**Common use: resource cleanup.** Placing a `defer` call immediately after acquiring a resource ensures the resource is always released, even if an error occurs partway through the function.

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // guaranteed to run when processFile returns

    // ... read from f ...
    return nil
}
```

Without `defer`, you would need to call `f.Close()` on every return path. With `defer`, there is exactly one close call that covers all paths.

**Loop gotcha:** `defer` defers until the *function* returns, not until the loop iteration ends.

```go
// Wrong: all files stay open until processAll returns.
func processAll(paths []string) {
    for _, path := range paths {
        f, _ := os.Open(path)
        defer f.Close() // does NOT close at end of this iteration
        // ... use f ...
    }
}

// Correct: wrap in a helper function so defer runs per iteration.
func processOne(path string) {
    f, _ := os.Open(path)
    defer f.Close() // runs when processOne returns (once per iteration)
    // ... use f ...
}

func processAll(paths []string) {
    for _, path := range paths {
        processOne(path)
    }
}
```

**Arguments are evaluated immediately.** Even though the call is deferred, its arguments are evaluated at the `defer` statement, not when the deferred call runs.

```go
func main() {
    x := 0
    defer fmt.Println("x =", x) // captures x=0 right now
    x = 42
    // prints "x = 0", not "x = 42"
}
```

Reference: https://go.dev/blog/defer-panic-and-recover

---

## 8. Panic and Recover

**`panic`** stops the current function's normal execution. Go unwinds the call stack, running any deferred functions along the way. If nothing recovers from the panic, the program prints a stack trace and exits.

```go
func mustPositive(n int) int {
    if n <= 0 {
        panic(fmt.Sprintf("expected positive number, got %d", n))
    }
    return n
}
```

**When to panic:**
- Programmer errors that represent "this should never happen" invariants (e.g., a configuration that was supposed to be validated at startup is missing).
- Initialisation errors in `init()` or variable initialisers that make the program non-functional.

**When NOT to panic:**
- Expected failure conditions (file not found, invalid user input, network errors). Use error returns instead.

**`recover`** must be called inside a deferred function. It catches an in-progress panic and returns the value passed to `panic`. If there is no panic, `recover` returns `nil`.

```go
func safeDiv(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered from panic: %v", r)
        }
    }()
    result = a / b // panics if b == 0 (integer division by zero)
    return result, nil
}

func main() {
    res, err := safeDiv(10, 2)
    fmt.Println(res, err) // 5 <nil>

    res, err = safeDiv(10, 0)
    fmt.Println(res, err) // 0 recovered from panic: runtime error: integer divide by zero
}
```

The pattern is:
```go
defer func() {
    if r := recover(); r != nil {
        // handle the panic — typically convert to an error
    }
}()
```

Note: the deferred function must be an anonymous function called with `()` at the end. If you write `defer recover()`, it will not work because `recover()` executes as the argument to `defer`, not inside a deferred call.

Reference: https://go.dev/blog/defer-panic-and-recover

---

## 9. Recursion

A function that calls itself is recursive. Every recursive function needs a **base case** — a condition under which it stops calling itself — to prevent infinite recursion.

**Factorial — recursive:**

```go
func factorial(n int) int {
    if n <= 1 { // base case
        return 1
    }
    return n * factorial(n-1)
}

fmt.Println(factorial(5)) // 120
```

**Factorial — iterative (usually preferred):**

```go
func factorialIter(n int) int {
    result := 1
    for i := 2; i <= n; i++ {
        result *= i
    }
    return result
}
```

The iterative version is generally preferred: it avoids function call overhead and cannot exhaust the call stack for large inputs. Understanding why requires knowing how Go manages the call stack.

### Go Call Stack Mechanics

When a function calls another function, the Go runtime pushes a **stack frame** onto the calling goroutine's stack. Each frame holds:

- the function's local variables
- its parameters and return values
- the **return address** — where execution resumes when the function returns

When the function returns, its frame is popped and that memory is reclaimed automatically.

**Every goroutine has its own stack.** Goroutine stacks are completely independent. Two goroutines can call the same function concurrently without interfering with each other's locals because each has its own copy of the frame.

**Goroutine stacks start small and grow dynamically.** When a goroutine is first created, its stack is only about 2–8 KB (the exact initial size has changed across Go versions — it was 8 KB before Go 1.4, now closer to 2–4 KB). As calls nest deeper and the current stack fills up, the runtime detects this automatically. It allocates a new, larger stack and copies all existing frames to the new location. This happens transparently — you do not manage it.

**Growth has a ceiling.** Dynamic growth is not unlimited. By default a goroutine's stack can reach **1 GB** on 64-bit systems (controlled by `runtime/debug.SetMaxStack`). A recursion deep enough to hit that ceiling terminates the program with a fatal error:

```
runtime: goroutine stack exceeds 1000000000-byte limit
runtime: sp=0xc0200e0388 stack=[0xc0200e0000, 0xc0400e0000]
fatal error: stack overflow
```

You can trigger this yourself by calling a function with no base case:

```go
func infinite(n int) int {
    return infinite(n + 1) // no base case — grows forever
}
```

Running `infinite(0)` will print the stack-overflow message and exit.

**Why iterative is safer for large inputs.** A loop reuses the same stack frame for every iteration — its memory footprint is constant. Each recursive call pushes a new frame. `factorial(1_000_000)` would push one million frames before the base case is reached. Go's dynamic stack would keep growing until it hit the 1 GB limit and crashed.

**Contrast with C and Java.** In C, each OS thread gets a fixed stack at creation time — typically 1–8 MB. In Java, the JVM similarly fixes the thread stack. Overflow in those languages triggers an immediate `StackOverflowError` or segfault. Go's copying-stack approach is more forgiving for moderate depths, but the 1 GB ceiling still applies.

**No tail-call optimisation (TCO).** Some languages (Scheme, Erlang) recognise when a function's last action is a recursive call and reuse the current frame rather than pushing a new one, making deep recursion use O(1) stack. **Go does not do this.** Even a perfectly tail-recursive Go function pushes one new frame per call. Never write Go code that relies on TCO.

**Fibonacci — recursive (exponential time complexity):**

```go
func fib(n int) int {
    if n <= 1 {
        return n
    }
    return fib(n-1) + fib(n-2)
}
```

`fib(40)` already makes over a billion calls. Memoisation (caching already-computed results) or iteration solves this. Day 05 (maps) will show how to memoize with a map.

**Real-world use: walking a tree.** Recursive algorithms are natural for tree-shaped data structures where each node can have sub-nodes of the same type.

```go
type TreeNode struct {
    Value    int
    Children []*TreeNode
}

// sum returns the sum of all values in the tree.
func sum(node *TreeNode) int {
    if node == nil {
        return 0
    }
    total := node.Value
    for _, child := range node.Children {
        total += sum(child)
    }
    return total
}
```

---

## 10. Structs

A struct groups related fields under a single named type.

**Declaration:**

```go
type Person struct {
    Name  string
    Email string
    Age   int
}
```

**Zero values:** fields not set in a literal are initialised to their zero values (`""` for strings, `0` for ints, `false` for bools, `nil` for pointers and slices).

**Struct literals — always use field names:**

```go
p := Person{Name: "Alice", Email: "alice@example.com", Age: 30}
```

Positional initialisation (`Person{"Alice", "alice@example.com", 30}`) is fragile — adding a field to the struct breaks every positional literal. Always use field names.

**Field access and mutation:**

```go
fmt.Println(p.Name) // Alice
p.Age++
fmt.Println(p.Age)  // 31
```

**Anonymous structs** are useful for one-off data shapes, such as grouping local variables or decoding JSON:

```go
point := struct {
    X, Y int
}{X: 10, Y: 20}
fmt.Println(point.X, point.Y) // 10 20
```

**Comparing structs:** two struct values are equal if all their fields are equal, provided every field type is comparable. Structs with slice or map fields are not directly comparable with `==`.

```go
a := Person{Name: "Alice", Email: "alice@example.com", Age: 30}
b := Person{Name: "Alice", Email: "alice@example.com", Age: 30}
fmt.Println(a == b) // true
```

**Structs are value types.** Assigning one struct variable to another copies the entire struct. Modifying the copy does not affect the original.

```go
q := p          // q is a full copy of p
q.Name = "Bob"
fmt.Println(p.Name) // Alice — p is unchanged
fmt.Println(q.Name) // Bob
```

---

## 11. Methods

A method is a function with a receiver — a named type that the function is associated with.

```go
type Rectangle struct {
    Width, Height float64
}

// Value receiver — reads the struct, does not mutate it.
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Pointer receiver — mutates the struct.
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}
```

Calling methods:

```go
rect := Rectangle{Width: 4, Height: 3}
fmt.Println(rect.Area()) // 12

rect.Scale(2)
fmt.Println(rect.Area()) // 48
```

Go automatically takes the address when you call a pointer-receiver method on an addressable value (`rect.Scale(2)` becomes `(&rect).Scale(2)`).

### Receiver Kinds — In Depth

#### Value receiver `(r T)`

The method receives a **copy** of the value. The original is untouched.

```go
func (r Rectangle) Area() float64 {
    return r.Width * r.Height // reads r; r is a copy
}
```

- Mutations inside the method do not affect the caller's variable.
- Calling the method on a `*Rectangle` also works — Go automatically dereferences: `(*ptr).Area()`.
- Value receivers are safe to use concurrently because each call operates on its own copy.

**Classic pitfall — mutating a copy by mistake:**

```go
type Counter struct{ count int }

// BUG: value receiver — c is a copy. The increment is lost.
func (c Counter) Increment() {
    c.count++ // mutates the copy, not the original
}

func (c Counter) Value() int { return c.count }

func main() {
    ctr := Counter{}
    ctr.Increment()
    ctr.Increment()
    fmt.Println(ctr.Value()) // 0 — not 2!
}
```

The code compiles and runs without error. The bug is silent: `Increment` modifies a copy that is discarded when the method returns. Switching to a pointer receiver fixes it (see below).

#### Pointer receiver `(r *T)`

The method receives a pointer to the original value. Mutations persist after the call.

```go
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor  // mutates the original Rectangle
    r.Height *= factor
}
```

**Concrete example — mutations that persist:**

```go
type Counter struct{ count int }

// Pointer receiver — c points to the original. Mutation persists.
func (c *Counter) Increment() { c.count++ }
func (c *Counter) Reset()     { c.count = 0 }
func (c Counter) Value() int  { return c.count } // read-only, value receiver is fine

func main() {
    ctr := Counter{}
    fmt.Println(ctr.Value()) // 0

    ctr.Increment()
    ctr.Increment()
    ctr.Increment()
    fmt.Println(ctr.Value()) // 3

    ctr.Reset()
    fmt.Println(ctr.Value()) // 0
}
```

Go rewrites `ctr.Increment()` as `(&ctr).Increment()` automatically because `ctr` is an addressable local variable.

- Calling on an addressable `T` variable is fine — Go takes the address automatically: `rect.Scale(2)` becomes `(&rect).Scale(2)`.
- You **cannot** call a pointer-receiver method on a non-addressable value. Map elements and function return values are not addressable:

```go
rects := map[string]Rectangle{"r": {4, 3}}
rects["r"].Scale(2) // compile error: cannot take the address of a map element
```

The fix: store a pointer in the map instead (`map[string]*Rectangle`), or copy the value out, mutate it, and store it back.

#### Nil pointer receivers

A method with a pointer receiver can be called on a `nil` pointer — as long as the method does not dereference the nil. This is occasionally useful:

```go
type Node struct {
    Val  int
    Next *Node
}

func (n *Node) Len() int {
    if n == nil {
        return 0
    }
    return 1 + n.Next.Len()
}

var head *Node
fmt.Println(head.Len()) // 0, not a panic
```

#### Method sets — the rule that matters for interfaces

Go tracks which methods are reachable on a type and on a pointer to that type. These are called **method sets**.

| Type | Method set |
|------|-----------|
| `T` | Value receiver methods only |
| `*T` | Value receiver methods **and** pointer receiver methods |

This asymmetry has one practical consequence: **interface satisfaction**.

If an interface requires a method that is declared with a pointer receiver, only `*T` satisfies that interface — not `T`. This is covered in depth on Day 09, but the core rule is:

- A method declared as `func (r Rectangle) Area() float64` belongs to both `Rectangle` and `*Rectangle`.
- A method declared as `func (r *Rectangle) Scale(f float64)` belongs only to `*Rectangle`.

If you store a `Rectangle` (not `*Rectangle`) in an interface variable, `Scale` is not available through that interface.

**Concrete example — interface satisfaction failure:**

```go
package main

import "fmt"

type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Pointer receiver — belongs to *Rectangle only, not Rectangle.
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

type Scaler interface {
    Scale(factor float64)
}

func doubleIt(s Scaler) {
    s.Scale(2)
}

func main() {
    r := Rectangle{Width: 4, Height: 3}
    fmt.Println(r.Area()) // 12

    doubleIt(&r) // OK — *Rectangle has Scale in its method set
    fmt.Println(r.Area()) // 48

    // doubleIt(r) — uncommenting this line produces:
    // cannot use r (variable of type Rectangle) as type Scaler in argument to doubleIt:
    //   Rectangle does not implement Scaler (Scale method has pointer receiver)
}
```

The error message is precise: Go tells you that `Scale` has a pointer receiver, which is why `Rectangle` (the value type) does not implement the interface. Using `&r` fixes it.

#### Consistency rule

Pick **one** receiver kind for all methods on a type and stick with it. Mixing value and pointer receivers is allowed by the compiler but leads to a confusing API: callers cannot tell at a glance whether a method mutates state.

The only common exception is implementing standard library interfaces like `String() string` (`fmt.Stringer`) — these are typically value receivers even on types that otherwise use pointer receivers, because they are read-only by definition.

#### When to choose which receiver

| Situation | Use |
|-----------|-----|
| Method mutates the receiver | Pointer `*T` |
| Type contains a `sync.Mutex` or similar (copying corrupts it) | Pointer `*T` |
| Type is large and copying on every call would be expensive | Pointer `*T` |
| Method only reads a small struct | Value `T` |
| Immutability is important (e.g. a mathematical type like `Vector`) | Value `T` |

When in doubt, use a pointer receiver. It is always correct and avoids the pitfall of accidentally mutating a copy and wondering why the original did not change.

---

## Labs

### Lab 1: Function Basics and Multiple Returns

**What you'll practise:** Defining functions with single and multiple parameters, shared-type shorthand, multiple return values, and the `(result, error)` idiom.

**Task:**
Write three functions — `greet`, `add`, and `divide` — then call them from `main`.

**Steps:**
1. Write `greet(name string) string` — returns `"Hello, " + name + "!"`
2. Write `add(a, b int) int` using the shared-type shorthand
3. Write `divide(a, b float64) (float64, error)` — returns an error if `b` is zero
4. In `main`, call all three; handle the error from `divide`; ignore the result of a successful call with `_` to show the syntax

```go
package main

import (
    "errors"
    "fmt"
)

func greet(name string) string {
    return "Hello, " + name + "!"
}

// Shared-type shorthand: a and b are both int.
func add(a, b int) int {
    return a + b
}

func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func main() {
    fmt.Println(greet("Gopher"))
    fmt.Println("3 + 4 =", add(3, 4))

    result, err := divide(10, 4)
    if err != nil {
        fmt.Println("error:", err)
    } else {
        fmt.Printf("10 / 4 = %.2f\n", result)
    }

    // Ignore the result — only care that there is no error.
    _, err = divide(9, 3)
    fmt.Println("divide(9,3) error:", err)

    // Error case.
    _, err = divide(5, 0)
    fmt.Println("divide(5,0) error:", err)
}
```

**Expected output:**
```
Hello, Gopher!
3 + 4 = 7
10 / 4 = 2.50
divide(9,3) error: <nil>
divide(5,0) error: division by zero
```

**Checkpoint:** Change `add(3, 4)` to `add(3, 4, 5)` and confirm the compiler rejects it — Go is strictly typed in arity. Restore before continuing.

---

### Lab 2: Variadic Functions

**What you'll practise:** Writing variadic functions, calling them with individual values and with spread slices, and enforcing a minimum argument count.

**Task:**
Write `sum` and `maxOf`, then call them in different ways.

**Steps:**
1. Write `sum(nums ...int) int` — returns the sum of all arguments; returns 0 for zero arguments
2. Write `maxOf(first int, rest ...int) int` — requires at least one argument (enforced by the type system); returns the largest value
3. Call `sum` with individual values, with no arguments, and with a spread slice
4. Call `maxOf` with one argument and with several

```go
package main

import "fmt"

func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

// maxOf requires at least one argument; rest may be empty.
func maxOf(first int, rest ...int) int {
    m := first
    for _, n := range rest {
        if n > m {
            m = n
        }
    }
    return m
}

func main() {
    fmt.Println("sum(1,2,3):", sum(1, 2, 3))
    fmt.Println("sum():", sum())

    nums := []int{4, 7, 2, 9, 1}
    fmt.Println("sum(nums...):", sum(nums...)) // spread the slice

    fmt.Println("maxOf(42):", maxOf(42))
    fmt.Println("maxOf(3,1,4,1,5,9,2,6):", maxOf(3, 1, 4, 1, 5, 9, 2, 6))
    fmt.Println("maxOf(nums...):", maxOf(nums[0], nums[1:]...))
}
```

**Expected output:**
```
sum(1,2,3): 6
sum(): 0
sum(nums...): 23
maxOf(42): 42
maxOf(3,1,4,1,5,9,2,6): 9
maxOf(nums...): 9
```

**Checkpoint:** Attempt to call `maxOf()` with no arguments and confirm the compiler rejects it — the `first int` parameter is not variadic and is therefore required.

---

### Lab 3: Anonymous Functions and Closures

**What you'll practise:** Function literals, immediately-invoked function expressions, closures that capture enclosing state, and passing functions as arguments.

**Task:**
Write an `adder` factory that returns a closure, an `apply` transformer, and a `makeCounter` function.

**Steps:**
1. Write `adder(x int) func(int) int` — returns a closure that adds `x` to its argument
2. Write `apply(nums []int, fn func(int) int) []int` — applies `fn` to every element
3. Write `makeCounter() func() int` — each call to the returned function increments an internal counter
4. Use an IIFE (immediately-invoked function expression) to compute `5 * 5` inline

```go
package main

import "fmt"

// adder returns a closure that captures x.
func adder(x int) func(int) int {
    return func(n int) int {
        return n + x
    }
}

// apply returns a new slice with fn applied to each element.
func apply(nums []int, fn func(int) int) []int {
    result := make([]int, len(nums))
    for i, n := range nums {
        result[i] = fn(n)
    }
    return result
}

// makeCounter returns a function that increments an internal counter on each call.
func makeCounter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

func main() {
    add10 := adder(10)
    fmt.Println("add10(5):", add10(5))   // 15
    fmt.Println("add10(20):", add10(20)) // 30

    // Two closures from the same factory have independent state.
    add100 := adder(100)
    fmt.Println("add100(5):", add100(5)) // 105

    nums := []int{1, 2, 3, 4, 5}
    fmt.Println("doubled:", apply(nums, func(n int) int { return n * 2 }))
    fmt.Println("+10:    ", apply(nums, adder(10)))

    // IIFE — define and call in one expression.
    squared := func(x int) int { return x * x }(5)
    fmt.Println("5 squared (IIFE):", squared)

    // Counter closure.
    counter := makeCounter()
    fmt.Println(counter(), counter(), counter()) // 1 2 3

    counter2 := makeCounter() // independent state
    fmt.Println("counter2:", counter2()) // 1
}
```

**Expected output:**
```
add10(5): 15
add10(20): 30
add100(5): 105
doubled: [2 4 6 8 10]
+10:     [11 12 13 14 15]
5 squared (IIFE): 25
1 2 3
counter2: 1
```

**Checkpoint:** Confirm that `add10` and `add100` do not share state — calling `add10` does not affect `add100`'s captured `x`. Confirm that `counter` and `counter2` do not share state.

---

### Lab 4: defer — Resource Cleanup

**What you'll practise:** Scheduling deferred calls, observing LIFO execution order, using `defer` for cleanup, and understanding when arguments are captured.

**Task:**
Write three short programs (all in one `main`) that demonstrate LIFO order, a cleanup pattern, and immediate argument capture.

**Steps:**
1. Call three `defer fmt.Println(...)` statements and observe the order
2. Write a `simulateWork` function that acquires and releases a simulated resource using `defer`
3. Demonstrate that deferred arguments are captured at the `defer` statement, not at execution time

```go
package main

import "fmt"

// Resource simulates something that must be closed after use.
type Resource struct {
    name string
}

func (r Resource) open()  { fmt.Println("opened:", r.name) }
func (r Resource) close() { fmt.Println("closed:", r.name) }

func simulateWork(name string) {
    r := Resource{name: name}
    r.open()
    defer r.close() // guaranteed to run when simulateWork returns

    fmt.Println("working with:", r.name)
    // If an error happened here and we returned early, r.close() still runs.
}

func main() {
    // 1. LIFO order.
    fmt.Println("--- LIFO order ---")
    defer fmt.Println("third (deferred first, runs last)")
    defer fmt.Println("second")
    defer fmt.Println("first (deferred last, runs first)")
    fmt.Println("main body")

    // 2. Cleanup pattern.
    fmt.Println("\n--- cleanup pattern ---")
    simulateWork("database-connection")
    simulateWork("temp-file")

    // 3. Arguments captured immediately.
    fmt.Println("\n--- argument capture ---")
    x := 0
    defer fmt.Println("deferred x =", x) // captures x=0 right now
    x = 99
    fmt.Println("current x =", x) // 99; deferred call still prints 0
}
```

**Expected output:**
```
--- LIFO order ---
main body
--- cleanup pattern ---
opened: database-connection
working with: database-connection
closed: database-connection
opened: temp-file
working with: temp-file
closed: temp-file

--- argument capture ---
current x = 99
first (deferred last, runs first)
second
third (deferred first, runs last)
deferred x = 0
```

**Checkpoint:** Add a second `defer r.close()` inside `simulateWork` and observe it is called twice (LIFO — the second `defer` runs before the first). Remove the duplicate before continuing.

---

### Lab 5: Panic and Recover

**What you'll practise:** Triggering a panic, using `defer` + `recover` to catch it, and converting a panic into a returned error.

**Task:**
Write `safeDiv` (integer division that converts a divide-by-zero panic into an error) and `mustPositive` (that panics intentionally for invalid input), then demonstrate both from `main`.

**Steps:**
1. Write `safeDiv(a, b int) (result int, err error)` using the `defer func() { recover() }()` pattern
2. Write `mustPositive(n int) int` — panics with an explicit message if `n <= 0`
3. Write `runMustPositive(n int)` — wraps `mustPositive` with a recover so a panic prints a warning without crashing the program
4. Call all from `main` with valid and invalid inputs

```go
package main

import "fmt"

// safeDiv converts an integer divide-by-zero panic into a returned error.
func safeDiv(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic recovered: %v", r)
        }
    }()
    result = a / b
    return result, nil
}

// mustPositive panics if n <= 0. Use this only for programmer-error invariants.
func mustPositive(n int) int {
    if n <= 0 {
        panic(fmt.Sprintf("mustPositive: got %d, want > 0", n))
    }
    return n
}

// runMustPositive calls mustPositive and recovers any panic so it doesn't crash the program.
func runMustPositive(n int) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("warning: recovered panic:", r)
        }
    }()
    result := mustPositive(n)
    fmt.Println("mustPositive result:", result)
}

func main() {
    // safeDiv: normal case.
    res, err := safeDiv(10, 2)
    fmt.Printf("safeDiv(10, 2) = %d, err = %v\n", res, err)

    // safeDiv: panic case.
    res, err = safeDiv(10, 0)
    fmt.Printf("safeDiv(10, 0) = %d, err = %v\n", res, err)

    // mustPositive: valid.
    runMustPositive(7)

    // mustPositive: panics; recover catches it.
    runMustPositive(-3)

    fmt.Println("program continues after recovered panic")
}
```

**Expected output:**
```
safeDiv(10, 2) = 5, err = <nil>
safeDiv(10, 0) = 0, err = panic recovered: runtime error: integer divide by zero
mustPositive result: 7
warning: recovered panic: mustPositive: got -3, want > 0
program continues after recovered panic
```

**Checkpoint:** Remove the `defer`/`recover` from `runMustPositive` and call `runMustPositive(-3)` directly — confirm the program crashes with a stack trace. Restore `defer`/`recover` before continuing.

---

### Lab 6: Recursion

**What you'll practise:** Writing recursive functions with base cases, comparing recursive and iterative approaches, and understanding the tradeoffs.

**Task:**
Implement `factorialRec`, `factorialIter`, and `fib`. Then write a recursive `treeSum` function over a simple tree structure.

**Steps:**
1. Write `factorialRec(n int) int` — recursive; base case `n <= 1` returns 1
2. Write `factorialIter(n int) int` — iterative; produces the same results
3. Write `fib(n int) int` — recursive Fibonacci; observe it becomes slow for large n
4. Define `TreeNode` and write `treeSum(node *TreeNode) int` — recursively sums all node values

```go
package main

import "fmt"

// --- Factorial ---

func factorialRec(n int) int {
    if n <= 1 {
        return 1
    }
    return n * factorialRec(n-1)
}

func factorialIter(n int) int {
    result := 1
    for i := 2; i <= n; i++ {
        result *= i
    }
    return result
}

// --- Fibonacci ---

func fib(n int) int {
    if n <= 1 {
        return n
    }
    return fib(n-1) + fib(n-2)
}

// --- Tree ---

type TreeNode struct {
    Value    int
    Children []*TreeNode
}

func treeSum(node *TreeNode) int {
    if node == nil {
        return 0
    }
    total := node.Value
    for _, child := range node.Children {
        total += treeSum(child)
    }
    return total
}

func main() {
    // Factorial comparison.
    for _, n := range []int{0, 1, 5, 10} {
        r := factorialRec(n)
        it := factorialIter(n)
        fmt.Printf("factorial(%d): recursive=%d iterative=%d match=%v\n", n, r, it, r == it)
    }

    // Fibonacci — keep n small; recursive fib(40) takes seconds.
    fmt.Println()
    for _, n := range []int{0, 1, 5, 10, 15} {
        fmt.Printf("fib(%d) = %d\n", n, fib(n))
    }

    // Tree sum.
    fmt.Println()
    root := &TreeNode{
        Value: 1,
        Children: []*TreeNode{
            {Value: 2, Children: []*TreeNode{
                {Value: 4},
                {Value: 5},
            }},
            {Value: 3, Children: []*TreeNode{
                {Value: 6},
            }},
        },
    }
    fmt.Println("tree sum:", treeSum(root)) // 1+2+3+4+5+6 = 21
}
```

**Expected output:**
```
factorial(0): recursive=1 iterative=1 match=true
factorial(1): recursive=1 iterative=1 match=true
factorial(5): recursive=120 iterative=120 match=true
factorial(10): recursive=3628800 iterative=3628800 match=true

fib(0) = 0
fib(1) = 1
fib(5) = 5
fib(10) = 55
fib(15) = 610

tree sum: 21
```

**Checkpoint:** Try `fib(40)` and observe the delay — it makes over a billion calls. Then try `factorialRec(0)` — confirm it returns 1, not 0 (the base case covers `n == 0`).

---

### Lab 7 (Final): Contact Card

**What you'll practise:** Defining a struct with validation, writing a constructor, adding pointer and value receiver methods, implementing `fmt.Stringer`, and searching a slice.

**Task:**
Write `day-03/main.go` with a `Contact` struct, a validating constructor, a `String()` method, a `PrintCard` function, and a `FindByName` search over a slice of contacts.

**Steps:**
1. Define `type Contact struct` with `Name`, `Email`, and `Phone` fields
2. Write `NewContact(name, email, phone string) (Contact, error)` — return an error if `email` does not contain `@`
3. Write `(c Contact) String() string` — returns a formatted single-line representation
4. Write `func PrintCard(c Contact)` — pretty-prints a bordered card
5. Build a `[]Contact` slice and write `FindByName(contacts []Contact, name string) (Contact, bool)` using case-insensitive comparison
6. Call all functions from `main`, including a failed construction and a failed lookup

```go
package main

import (
    "fmt"
    "strings"
)

type Contact struct {
    Name  string
    Email string
    Phone string
}

func NewContact(name, email, phone string) (Contact, error) {
    if !strings.Contains(email, "@") {
        return Contact{}, fmt.Errorf("invalid email %q: must contain @", email)
    }
    return Contact{Name: name, Email: email, Phone: phone}, nil
}

func (c Contact) String() string {
    return fmt.Sprintf("%s | %s | %s", c.Name, c.Email, c.Phone)
}

func PrintCard(c Contact) {
    border := strings.Repeat("-", 40)
    fmt.Println(border)
    fmt.Printf("  Name:  %s\n", c.Name)
    fmt.Printf("  Email: %s\n", c.Email)
    fmt.Printf("  Phone: %s\n", c.Phone)
    fmt.Println(border)
}

func FindByName(contacts []Contact, name string) (Contact, bool) {
    for _, c := range contacts {
        if strings.EqualFold(c.Name, name) {
            return c, true
        }
    }
    return Contact{}, false
}

func main() {
    var contacts []Contact

    entries := [][3]string{
        {"Alice", "alice@example.com", "555-1234"},
        {"Bob", "bob@example.com", "555-5678"},
        {"Carol", "carol@example.com", "555-9012"},
    }
    for _, e := range entries {
        c, err := NewContact(e[0], e[1], e[2])
        if err != nil {
            fmt.Println("error:", err)
            continue
        }
        contacts = append(contacts, c)
    }

    // Try an invalid email.
    _, err := NewContact("Dan", "not-an-email", "555-0000")
    fmt.Println("Validation error:", err)
    fmt.Println()

    // Print all cards.
    for _, c := range contacts {
        PrintCard(c)
    }

    // Search — case-insensitive.
    if c, ok := FindByName(contacts, "bob"); ok {
        fmt.Println("Found:", c) // calls c.String() automatically
    }
    if _, ok := FindByName(contacts, "Zara"); !ok {
        fmt.Println("Zara not found")
    }
}
```

**Expected output:**
```
Validation error: invalid email "not-an-email": must contain @

----------------------------------------
  Name:  Alice
  Email: alice@example.com
  Phone: 555-1234
----------------------------------------
----------------------------------------
  Name:  Bob
  Email: bob@example.com
  Phone: 555-5678
----------------------------------------
----------------------------------------
  Name:  Carol
  Email: carol@example.com
  Phone: 555-9012
----------------------------------------
Found: Bob | bob@example.com | 555-5678
Zara not found
```

**Checkpoint:** `go vet .` reports no issues. `FindByName` is case-insensitive (`"bob"` finds `"Bob"`). Adding a contact with an invalid email is rejected without crashing.

---

## Day Project Goal

Replace `main.go` with the Contact Card program from Lab 7. Once it runs correctly:

1. Run `go run .` from `day-03/` and verify the output matches the expected output above.
2. Run `go vet .` — it should report no issues.
3. Run `go fmt .` to format the file if needed.

The program demonstrates the core Day 03 skills: structs, constructors with validation, methods, `fmt.Stringer`, and slice-based search with error handling.

---

## Extension Ideas

- Add a `(c *Contact) UpdatePhone(phone string)` pointer-receiver method and call it after construction.
- Write `FindAll(contacts []Contact, fn func(Contact) bool) []Contact` — a generic filter that accepts any predicate function.
- Add struct embedding: create an `Address` struct and embed it in `Contact` so `c.City` works as a promoted field.
- Add a `SortByName(contacts []Contact)` function using a closure-based comparison (explore `sort.Slice` from the standard library).
- Write a memoised version of `fib` using a `map[int]int` (preview of Day 05).
- Rewrite `factorialRec` to detect overflow for large inputs and return an error instead of silently wrapping around.

---

## Official Documentation

- [Language Spec: Function declarations](https://go.dev/ref/spec#Function_declarations)
- [Language Spec: Function types (variadic)](https://go.dev/ref/spec#Function_types)
- [Language Spec: Defer statements](https://go.dev/ref/spec#Defer_statements)
- [Language Spec: Handling panics](https://go.dev/ref/spec#Handling_panics)
- [The Go Blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover)
- [Built-in: panic](https://pkg.go.dev/builtin#panic)
- [Built-in: recover](https://pkg.go.dev/builtin#recover)
- [Language Spec: Struct types](https://go.dev/ref/spec#Struct_types)
- [Language Spec: Method declarations](https://go.dev/ref/spec#Method_declarations)
