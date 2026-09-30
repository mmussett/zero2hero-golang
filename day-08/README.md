# Day 08: Interfaces

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
| `fmt.Stringer` | `String() string` | Custom `fmt.Println` output |
| `error` | `Error() string` | Error values |
| `io.Reader` | `Read([]byte) (int, error)` | Anything readable |
| `io.Writer` | `Write([]byte) (int, error)` | Anything writable |
| `sort.Interface` | `Len`, `Less`, `Swap` | Custom sort |

## Day Project: Shape Library

Define:
- `Shape` interface with `Area() float64` and `Perimeter() float64`
- `Circle`, `Rectangle`, `Triangle` concrete types
- `fmt.Stringer` on each
- `TotalArea(shapes []Shape) float64` and `LargestShape(shapes []Shape) Shape`
- A type switch that prints different messages per shape type

**Extension ideas:** add a `Scale(factor float64) Shape` method; implement `json.Marshaler`.
