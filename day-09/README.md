# Day 09: Interfaces

## Core Concept: Implicit Satisfaction

In Go, a type satisfies an interface simply by implementing all its methods — no `implements` keyword, no explicit declaration. This is called **duck typing** or **structural typing**.

```go
type Shape interface {
    Area() float64
    Perimeter() float64
}

type Circle struct{ Radius float64 }
func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

// Circle satisfies Shape automatically
var s Shape = Circle{Radius: 5}
fmt.Printf("Area: %.2f\n", s.Area())
```

## Interface Composition

Interfaces can embed other interfaces:

```go
type ReadWriter interface {
    io.Reader
    io.Writer
}
```

## The Empty Interface: `any`

`any` (alias for `interface{}`) accepts a value of any type. Use it sparingly — it bypasses type safety.

```go
func Print(v any) { fmt.Printf("%v (%T)\n", v, v) }
```

## Type Assertions

Extract the underlying concrete type from an interface:

```go
var s Shape = Circle{Radius: 3}

c, ok := s.(Circle)     // safe — ok is false if wrong type
if ok { fmt.Println(c.Radius) }

c2 := s.(Circle)        // panics if s is not a Circle
```

## Type Switches

Dispatch on multiple concrete types:

```go
func describe(s Shape) string {
    switch v := s.(type) {
    case Circle:    return fmt.Sprintf("Circle r=%.1f", v.Radius)
    case Rectangle: return fmt.Sprintf("Rect %gx%g", v.Width, v.Height)
    default:        return "unknown shape"
    }
}
```

## Common Standard-Library Interfaces

| Interface | Methods | Use |
|-----------|---------|-----|
| [`fmt.Stringer`](https://pkg.go.dev/fmt#Stringer) | `String() string` | Custom `fmt.Println` output |
| `error` | `Error() string` | Error values |
| [`io.Reader`](https://pkg.go.dev/io#Reader) | `Read([]byte) (int, error)` | Anything readable |
| [`io.Writer`](https://pkg.go.dev/io#Writer) | `Write([]byte) (int, error)` | Anything writable |
| [`sort.Interface`](https://pkg.go.dev/sort#Interface) | `Len`, `Less`, `Swap` | Custom sort |

## Labs

### Lab 1: Define `Shape` and Three Concrete Types

**What you'll practise:** declaring an interface and satisfying it implicitly with multiple concrete types.

**Task:**
Define a `Shape` interface with `Area() float64` and `Perimeter() float64`. Implement `Circle`, `Rectangle`, and `Triangle` structs that satisfy it, plus `fmt.Stringer` on each.

**Steps:**
1. Declare `type Shape interface { Area() float64; Perimeter() float64 }`
2. Define `Circle{Radius float64}`, `Rectangle{Width, Height float64}`, `Triangle{A, B, C float64}` (sides)
3. Implement `Area()` and `Perimeter()` for each — use `math.Pi` and Heron's formula for Triangle
4. Implement `String() string` on each so `fmt.Println` produces a readable description
5. In `main`, assign each to a `Shape` variable and print `Area()` and `Perimeter()`

```go
import "math"

type Shape interface {
    Area() float64
    Perimeter() float64
}

type Circle struct{ Radius float64 }

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }
func (c Circle) String() string     { return fmt.Sprintf("Circle(r=%.2f)", c.Radius) }

type Rectangle struct{ Width, Height float64 }

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }
func (r Rectangle) String() string     { return fmt.Sprintf("Rect(%.2fx%.2f)", r.Width, r.Height) }
```

**Expected output:**
```
Circle(r=5.00)  area=78.54  perimeter=31.42
Rect(3.00x4.00) area=12.00  perimeter=14.00
```

**Checkpoint:** All three types compile; assigning any of them to a `Shape` variable works without a cast.

---

### Lab 2: `TotalArea` and `LargestShape`

**What you'll practise:** writing functions that accept interface slices, demonstrating runtime polymorphism.

**Task:**
Write `TotalArea(shapes []Shape) float64` and `LargestShape(shapes []Shape) Shape` using only the `Shape` interface — no type assertions.

**Steps:**
1. Implement `TotalArea` by ranging over the slice and summing `s.Area()`
2. Implement `LargestShape` by tracking the shape with the maximum area; return `nil` for an empty slice
3. Create a mixed `[]Shape{Circle{5}, Rectangle{3,4}, Triangle{3,4,5}}` in main
4. Print the total area and the string representation of the largest shape

```go
func TotalArea(shapes []Shape) float64 {
    total := 0.0
    for _, s := range shapes {
        total += s.Area()
    }
    return total
}

func LargestShape(shapes []Shape) Shape {
    if len(shapes) == 0 {
        return nil
    }
    best := shapes[0]
    for _, s := range shapes[1:] {
        if s.Area() > best.Area() {
            best = s
        }
    }
    return best
}
```

**Expected output:**
```
Total area: 96.54
Largest: Circle(r=5.00)
```

**Checkpoint:** `LargestShape` returns the correct shape for any ordering of the input slice.

---

### Lab 3: Type Assertions — Extract Only Circles

**What you'll practise:** the comma-ok form of type assertions to safely extract concrete types from an interface slice.

**Task:**
Given a `[]Shape` containing mixed types, use type assertion to collect only the `Circle` values into a `[]Circle`. Handle the non-circle case gracefully.

**Steps:**
1. Create `shapes := []Shape{Circle{1}, Rectangle{2,3}, Circle{5}, Triangle{3,4,5}}`
2. Iterate and attempt `c, ok := s.(Circle)` — collect only when `ok` is true
3. Print each extracted circle's radius
4. Also demonstrate the panicking form `s.(Circle)` on a known Circle, then show what happens with the wrong type using a recover

```go
func extractCircles(shapes []Shape) []Circle {
    var circles []Circle
    for _, s := range shapes {
        if c, ok := s.(Circle); ok {
            circles = append(circles, c)
        }
    }
    return circles
}
```

**Expected output:**
```
Circles: [Circle(r=1.00) Circle(r=5.00)]
Non-circles skipped: 2
```

**Checkpoint:** The function never panics regardless of the input slice composition.

---

### Lab 4: Type Switch — `describe`

**What you'll practise:** type switches for dispatching on concrete types without explicit if-chains.

**Task:**
Write `describe(s Shape) string` that returns a human-readable sentence about the shape using a type switch.

**Steps:**
1. Implement `describe` with `switch v := s.(type)`
2. Handle `Circle`, `Rectangle`, `Triangle` with shape-specific sentences
3. Add a `default` case returning "unknown shape"
4. Test with all three types and an unknown type wrapped in an interface

```go
func describe(s Shape) string {
    switch v := s.(type) {
    case Circle:
        return fmt.Sprintf("a circle with radius %.2f and area %.2f", v.Radius, v.Area())
    case Rectangle:
        return fmt.Sprintf("a %.2f by %.2f rectangle", v.Width, v.Height)
    case Triangle:
        return fmt.Sprintf("a triangle with sides %.2f, %.2f, %.2f", v.A, v.B, v.C)
    default:
        return fmt.Sprintf("unknown shape: %T", v)
    }
}
```

**Expected output:**
```
a circle with radius 5.00 and area 78.54
a 3.00 by 4.00 rectangle
a triangle with sides 3.00, 4.00, 5.00
```

**Checkpoint:** `describe` handles all three types and the default case without any type assertions outside the switch.

---

### Lab 5: Interface Composition — `LabelledShape`

**What you'll practise:** embedding interfaces to compose richer interface types.

**Task:**
Define a `Stringer` interface (`String() string`), a `Sizer` interface (`Size() float64`), and a composed `LabelledShape` interface that embeds both plus `Shape`. Then write a function that accepts only `LabelledShape`.

**Steps:**
1. Define `type Stringer interface { String() string }` and `type Sizer interface { Size() float64 }`
2. Define `type LabelledShape interface { Shape; Stringer; Sizer }`
3. Add `Size() float64` to `Circle` (returning diameter) and `Rectangle` (returning diagonal)
4. Write `PrintLabelled(ls LabelledShape)` that prints the label, size, area, and perimeter
5. Observe that `Triangle` does not implement `LabelledShape` (compile error if you try)

```go
type Stringer interface{ String() string }
type Sizer   interface{ Size() float64 }

type LabelledShape interface {
    Shape
    Stringer
    Sizer
}

func PrintLabelled(ls LabelledShape) {
    fmt.Printf("%s  size=%.2f  area=%.2f  perim=%.2f\n",
        ls.String(), ls.Size(), ls.Area(), ls.Perimeter())
}
```

**Expected output:**
```
Circle(r=5.00)  size=10.00  area=78.54  perim=31.42
Rect(3.00x4.00) size=5.00   area=12.00  perim=14.00
```

**Checkpoint:** `PrintLabelled` compiles only when called with a type that satisfies all three embedded interfaces.

---

### Lab 6: The Empty Interface — `PrintAll`

**What you'll practise:** using `any` to accept heterogeneous values, and understanding why `any` loses type safety.

**Task:**
Write `PrintAll(values []any)` that prints each value using `fmt.Sprint`. Then demonstrate the loss of type safety by mixing shapes, integers, and strings in the same slice.

**Steps:**
1. Implement `PrintAll` using a range loop and `fmt.Sprintf("%v (%T)", v, v)`
2. Call it with `[]any{Circle{3}, 42, "hello", true}`
3. Try calling `.Area()` on an element extracted from the `[]any` — observe the compile error (you must assert first)
4. Discuss in a comment why `[]Shape` is safer than `[]any` for shape collections

```go
func PrintAll(values []any) {
    for _, v := range values {
        fmt.Printf("%v  (type: %T)\n", v, v)
    }
}

func main() {
    PrintAll([]any{Circle{Radius: 3}, 42, "hello", true})

    // To call .Area() you must assert — any loses the interface:
    var v any = Circle{Radius: 3}
    if s, ok := v.(Shape); ok {
        fmt.Println("area:", s.Area())
    }
}
```

**Expected output:**
```
Circle(r=3.00)  (type: main.Circle)
42  (type: int)
hello  (type: string)
true  (type: bool)
area: 28.27
```

**Checkpoint:** You can explain in a comment why `[]any` is appropriate here but `[]Shape` would be better for a collection of shapes.

---

### Final Lab (Project): Shape Library

**What you'll practise:** combining interface design, type assertions, type switches, interface composition, and the empty interface into a coherent shape library.

**Task:**
Build a complete shape library that brings together all six labs: define the `Shape` interface, implement three concrete types, write `TotalArea` and `LargestShape`, add a type switch `describe` function, compose a `LabelledShape` interface, and demonstrate `any`.

**Steps:**
1. Define `Shape` interface with `Area() float64` and `Perimeter() float64`
2. Implement `Circle`, `Rectangle`, `Triangle` concrete types with `fmt.Stringer`
3. Write `TotalArea(shapes []Shape) float64` and `LargestShape(shapes []Shape) Shape`
4. Write `describe(s Shape) string` using a type switch
5. Create a mixed `[]Shape` and print total area, largest shape, and descriptions

```go
shapes := []Shape{
    Circle{Radius: 5},
    Rectangle{Width: 3, Height: 4},
    Triangle{A: 3, B: 4, C: 5},
}
fmt.Printf("Total area:  %.2f\n", TotalArea(shapes))
fmt.Printf("Largest:     %s\n", LargestShape(shapes))
for _, s := range shapes {
    fmt.Println(describe(s))
}
```

**Expected output:**
```
Total area:  96.54
Largest:     Circle(r=5.00)
a circle with radius 5.00 and area 78.54
a 3.00 by 4.00 rectangle
a triangle with sides 3.00, 4.00, 5.00
```

**Checkpoint:** `go test ./...` passes; `go vet ./...` is clean; each concrete type satisfies `Shape` with no explicit declaration.

**Extension ideas:** add a `Scale(factor float64) Shape` method; implement `json.Marshaler`.

## Official Documentation

- [`fmt`](https://pkg.go.dev/fmt) — Stringer interface, Printf, Sprintf
- [`io`](https://pkg.go.dev/io) — Reader and Writer interfaces
- [`sort`](https://pkg.go.dev/sort) — sort.Interface for custom sorting
- [`math`](https://pkg.go.dev/math) — `math.Pi` and other constants
- [Language Spec: Interface types](https://go.dev/ref/spec#Interface_types) — interface declarations
- [Language Spec: Type assertions](https://go.dev/ref/spec#Type_assertions) — safe type extraction
- [Language Spec: Type switches](https://go.dev/ref/spec#Type_switches) — dispatch on concrete type
- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces) — interface design principles
- [Go Tour: Interfaces](https://go.dev/tour/methods/9) — interactive interfaces tour
