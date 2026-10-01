# Day 02: Variables, Types, and Control Flow

## Core Concepts

### Declaration Styles

Go has two ways to declare variables:

```go
var x int = 10    // explicit type and value
var y = 20        // type inferred from value
z := 30           // short declaration, inside functions only
```

`var` works at package scope; `:=` only works inside functions. Prefer `:=` locally; use `var` for zero-value declarations (`var count int`).

### Numeric Types and Conversion

All numeric conversions in Go are **explicit**. There is no implicit widening.

```go
var i int     = 42
var f float64 = float64(i)   // must convert
var u uint    = uint(f)
```

Integer division truncates: `7 / 2 == 3`. For float division, at least one operand must be `float64`.

### Constants

```go
const Pi = 3.14159
const (
    StatusOK       = 200
    StatusNotFound = 404
)
```

Constants are typed or untyped. Untyped constants adapt to context:

```go
const Big = 1 << 62   // fits in int64 but not int32
```

### Control Flow

Go has **one loop keyword** — `for` — used for all looping patterns:

```go
for i := 0; i < 10; i++ { }      // C-style
for condition { }                  // while-style
for { }                            // infinite loop
for i, v := range slice { }       // range loop
```

`switch` in Go does not fall through by default (no `break` needed):

```go
switch op {
case '+': fmt.Println(a + b)
case '-': fmt.Println(a - b)
case '*': fmt.Println(a * b)
case '/':
    if b == 0 {
        fmt.Println("error: division by zero")
    } else {
        fmt.Printf("%.2f\n", a/b)
    }
default: fmt.Println("unknown operator")
}
```

`if` can have an init statement:

```go
if err := doSomething(); err != nil {
    // handle error
}
```

### Zero Values

Every variable in Go is initialised to its zero value if not assigned:

| Type | Zero value |
|------|-----------|
| `int`, `float64` | `0` |
| `bool` | `false` |
| `string` | `""` |
| pointer, slice, map | `nil` |

---

## Labs

### Lab 1: All Basic Types and Their Zero Values

**What you'll practise:** Declaring every basic Go type and observing zero values when a variable is declared but not assigned.

**Task:**
Write a program that declares one variable of each basic type without assigning a value, prints the zero value, then assigns and prints a real value.

**Steps:**
1. Declare `var` for each type: `bool`, `int`, `int8`, `int32`, `int64`, `uint`, `float32`, `float64`, `complex128`, `byte`, `rune`, `string`
2. Print each zero value with its type using `%T %v`
3. Assign real values and print again

```go
package main

import "fmt"

func main() {
    var b   bool
    var i   int
    var f32 float32
    var f64 float64
    var c   complex128
    var by  byte   // alias for uint8
    var r   rune   // alias for int32
    var s   string

    fmt.Printf("bool:       %T = %v\n", b, b)
    fmt.Printf("int:        %T = %v\n", i, i)
    fmt.Printf("float32:    %T = %v\n", f32, f32)
    fmt.Printf("float64:    %T = %v\n", f64, f64)
    fmt.Printf("complex128: %T = %v\n", c, c)
    fmt.Printf("byte:       %T = %v\n", by, by)
    fmt.Printf("rune:       %T = %v\n", r, r)
    fmt.Printf("string:     %T = %q\n", s, s)
}
```

**Expected output:**
```
bool:       bool = false
int:        int = 0
float32:    float32 = 0
float64:    float64 = 0
complex128: complex128 = (0+0i)
byte:       uint8 = 0
rune:       int32 = 0
string:     string = ""
```

**Checkpoint:** Every type prints its zero value. Notice `byte` prints as `uint8` and `rune` as `int32` — they are aliases, not distinct types.

---

### Lab 2: Type Conversion Experiment

**What you'll practise:** Explicit numeric conversion, truncation rules, and the difference between converting an int to string via `rune` vs `strconv.Itoa`.

**Task:**
Convert between numeric types and strings, observing what gets lost and why.

**Steps:**
1. Convert `int` to `float64` and back — observe truncation
2. Divide two ints and observe integer division; then force float division
3. Convert an `int` to `string` two ways: `string(rune(n))` and `strconv.Itoa(n)` — compare the results

```go
package main

import (
    "fmt"
    "strconv"
)

func main() {
    i := 7
    f := float64(i)
    fmt.Printf("int→float64: %v\n", f)      // 7

    back := int(3.99)
    fmt.Printf("float→int (truncates): %v\n", back) // 3

    // Integer division vs float division
    fmt.Printf("7 / 2 = %v (int division)\n", 7/2)
    fmt.Printf("7.0 / 2.0 = %v (float division)\n", 7.0/2.0)

    // int to string: two very different results
    n := 65
    fmt.Printf("string(rune(%d)) = %q\n", n, string(rune(n)))  // "A" (Unicode code point)
    fmt.Printf("strconv.Itoa(%d) = %q\n", n, strconv.Itoa(n))  // "65" (decimal digits)
}
```

**Expected output:**
```
int→float64: 7
float→int (truncates): 3
7 / 2 = 3 (int division)
7.0 / 2.0 = 3.5 (float division)
string(rune(65)) = "A"
strconv.Itoa(65) = "65"
```

**Checkpoint:** You can explain why `string(rune(65))` produces `"A"` (it treats 65 as a Unicode code point) while `strconv.Itoa(65)` produces `"65"` (it converts the number to decimal text).

---

### Lab 3: All Forms of the `for` Loop

**What you'll practise:** Writing C-style, while-equivalent, infinite-with-break, and range-over-string loops.

**Task:**
Count vowels in a sentence using a `range` loop over the string, and also demonstrate each other loop form.

**Steps:**
1. Write a C-style `for` that prints numbers 1–5
2. Write a while-equivalent `for condition {}` that halves a number until it's below 1
3. Write an infinite loop with `break` that reads the first 3 even numbers from a counter
4. Use `range` over a string to count vowels (handle multi-byte runes correctly)

```go
package main

import (
    "fmt"
    "strings"
)

func main() {
    // C-style
    for i := 1; i <= 5; i++ {
        fmt.Printf("%d ", i)
    }
    fmt.Println()

    // while-equivalent
    n := 64.0
    for n >= 1 {
        n /= 2
    }
    fmt.Printf("halved to: %.4f\n", n)

    // infinite loop with break
    count, val := 0, 0
    for {
        val++
        if val%2 == 0 {
            fmt.Printf("even: %d\n", val)
            count++
        }
        if count == 3 {
            break
        }
    }

    // range over string — counts vowels
    sentence := "The quick brown fox jumps over the lazy dog"
    vowels := 0
    for _, r := range sentence {
        if strings.ContainsRune("aeiouAEIOU", r) {
            vowels++
        }
    }
    fmt.Printf("Vowels in sentence: %d\n", vowels)
}
```

**Expected output:**
```
1 2 3 4 5
halved to: 0.5000
even: 2
even: 4
even: 6
Vowels in sentence: 11
```

**Checkpoint:** All four loop forms run without error. Change the sentence and verify the vowel count updates correctly.

---

### Lab 4: switch — Grade Calculator and Type Switch

**What you'll practise:** Expression switch with range cases and a type switch on `interface{}`.

**Task:**
Write a grade calculator using switch on score ranges, then write a type switch that identifies the dynamic type of values stored in an `any` (interface{}) variable.

**Steps:**
1. Write `grade(score int) string` using a `switch` with no condition (acts like if/else chain)
2. Call it for several scores including boundary values
3. Write `describe(v any) string` using a type switch
4. Call it with an `int`, `string`, `bool`, and `float64`

```go
package main

import "fmt"

func grade(score int) string {
    switch {
    case score >= 90:
        return "A"
    case score >= 80:
        return "B"
    case score >= 70:
        return "C"
    case score >= 60:
        return "D"
    default:
        return "F"
    }
}

func describe(v any) string {
    switch t := v.(type) {
    case int:
        return fmt.Sprintf("int(%d)", t)
    case string:
        return fmt.Sprintf("string(%q)", t)
    case bool:
        return fmt.Sprintf("bool(%v)", t)
    case float64:
        return fmt.Sprintf("float64(%.2f)", t)
    default:
        return fmt.Sprintf("unknown type: %T", t)
    }
}

func main() {
    for _, score := range []int{95, 82, 73, 60, 45} {
        fmt.Printf("Score %d → Grade %s\n", score, grade(score))
    }

    for _, v := range []any{42, "hello", true, 3.14} {
        fmt.Println(describe(v))
    }
}
```

**Expected output:**
```
Score 95 → Grade A
Score 82 → Grade B
Score 73 → Grade C
Score 60 → Grade D
Score 45 → Grade F
int(42)
string("hello")
bool(true)
float64(3.14)
```

**Checkpoint:** Add a `[]int` value to the `any` slice — confirm `describe` returns `unknown type: []int`.

---

### Lab 5: const and iota — Weekdays and Directions

**What you'll practise:** Defining typed constants with `iota`, using `iota` with arithmetic expressions, and printing named constants.

**Task:**
Define a `Weekday` type with `iota` starting at 1 (Monday=1), and a `Direction` type with bitfield constants. Print their values and names.

**Steps:**
1. Define `type Weekday int` with `Monday` through `Sunday` using `iota + 1`
2. Add a `String() string` method on `Weekday` that returns the day name
3. Define `type Direction uint` with `N`, `S`, `E`, `W` as powers of 2 using `1 << iota`
4. Print all weekdays and directions

```go
package main

import "fmt"

type Weekday int

const (
    Monday Weekday = iota + 1
    Tuesday
    Wednesday
    Thursday
    Friday
    Saturday
    Sunday
)

func (d Weekday) String() string {
    names := [...]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
    if d < Monday || d > Sunday {
        return fmt.Sprintf("Unknown(%d)", int(d))
    }
    return names[d-1]
}

type Direction uint

const (
    North Direction = 1 << iota // 1
    South                       // 2
    East                        // 4
    West                        // 8
)

func main() {
    for d := Monday; d <= Sunday; d++ {
        fmt.Printf("%d: %s\n", d, d)
    }
    fmt.Println()
    for _, dir := range []Direction{North, South, East, West} {
        fmt.Printf("Direction value: %d\n", dir)
    }
}
```

**Expected output:**
```
1: Monday
2: Tuesday
3: Wednesday
4: Thursday
5: Friday
6: Saturday
7: Sunday

Direction value: 1
Direction value: 2
Direction value: 4
Direction value: 8
```

**Checkpoint:** `Weekday(8).String()` returns `"Unknown(8)"`. Direction values are powers of 2 — they can be combined with bitwise OR for flag-style usage.

---

### Lab 6 (Final): CLI Calculator

**What you'll practise:** Combining variables, all numeric types, switch, and os.Args into a fully functional command-line tool.

**Task:**
Write `day-02/main.go` — a CLI calculator that reads two numbers and an operator from `os.Args`, uses `switch` to compute the result, and guards against division by zero.

**Steps:**
1. Read `os.Args[1]`, `os.Args[2]`, `os.Args[3]` for the two operands and operator
2. Parse the operands to `float64` using `strconv.ParseFloat`
3. Use `switch` on the operator string to compute `+`, `-`, `*`, `/`
4. Guard division by zero in the `/` case
5. Print a formatted result line

```go
package main

import (
    "fmt"
    "os"
    "strconv"
)

func main() {
    if len(os.Args) != 4 {
        fmt.Fprintf(os.Stderr, "usage: calc <num> <op> <num>\n")
        os.Exit(1)
    }

    a, err := strconv.ParseFloat(os.Args[1], 64)
    if err != nil {
        fmt.Fprintf(os.Stderr, "invalid number: %s\n", os.Args[1])
        os.Exit(1)
    }
    b, err := strconv.ParseFloat(os.Args[3], 64)
    if err != nil {
        fmt.Fprintf(os.Stderr, "invalid number: %s\n", os.Args[3])
        os.Exit(1)
    }
    op := os.Args[2]

    switch op {
    case "+":
        fmt.Printf("%.2f + %.2f = %.2f\n", a, b, a+b)
    case "-":
        fmt.Printf("%.2f - %.2f = %.2f\n", a, b, a-b)
    case "*":
        fmt.Printf("%.2f * %.2f = %.2f\n", a, b, a*b)
    case "/":
        if b == 0 {
            fmt.Fprintln(os.Stderr, "error: division by zero")
            os.Exit(1)
        }
        fmt.Printf("%.2f / %.2f = %.2f\n", a, b, a/b)
    default:
        fmt.Fprintf(os.Stderr, "unknown operator: %s\n", op)
        os.Exit(1)
    }
}
```

**Expected output:**
```
$ go run . 12 / 4
12.00 / 4.00 = 3.00

$ go run . 7 + 3
7.00 + 3.00 = 10.00

$ go run . 10 / 0
error: division by zero
```

**Checkpoint:** All four operators work. Division by zero exits with a non-zero status (`echo $?` prints `1`). Passing a non-numeric argument prints a helpful error.

---

## Day Project: CLI Calculator

Declare two `float64` values and a `byte` operator. Use `switch` to compute the result. Guard division by zero.

```go
a, b := 12.0, 4.0
op := byte('/')

switch op {
case '/':
    if b == 0 { fmt.Println("error: division by zero"); return }
    fmt.Printf("%.2f / %.2f = %.2f\n", a, b, a/b)
// ...
}
```

**Extension ideas:** read `a`, `b`, and `op` from [`os.Args`](https://pkg.go.dev/os#pkg-variables); add `%` modulo; use [`fmt.Sscanf`](https://pkg.go.dev/fmt#Sscanf) to parse a string like `"12.0 / 4.0"`.

## Official Documentation

- [`fmt`](https://pkg.go.dev/fmt) — formatted I/O (Printf, Println, Sscanf)
- [`os`](https://pkg.go.dev/os) — `os.Args` for command-line arguments
- [Language Spec: Variables](https://go.dev/ref/spec#Variables) — variable declarations
- [Language Spec: Constants](https://go.dev/ref/spec#Constants) — typed and untyped constants
- [Language Spec: For statements](https://go.dev/ref/spec#For_statements) — all loop forms
- [Language Spec: Switch statements](https://go.dev/ref/spec#Switch_statements) — expression and type switches
- [Go Tour: Flow control](https://go.dev/tour/flowcontrol/1) — interactive control flow tour
