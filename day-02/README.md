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

**Extension ideas:** read `a`, `b`, and `op` from `os.Args`; add `%` modulo; use `fmt.Sscanf` to parse a string like `"12.0 / 4.0"`.
