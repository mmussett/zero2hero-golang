package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Perimeter() float64
}

// Circle
type Circle struct{ Radius float64 }

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }
func (c Circle) String() string     { return fmt.Sprintf("Circle(r=%.2f)", c.Radius) }

// Rectangle
type Rectangle struct{ Width, Height float64 }

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }
func (r Rectangle) String() string     { return fmt.Sprintf("Rectangle(%.2fx%.2f)", r.Width, r.Height) }

// Triangle (using Heron's formula)
type Triangle struct{ A, B, C float64 }

func (t Triangle) Perimeter() float64 { return t.A + t.B + t.C }
func (t Triangle) Area() float64 {
	s := t.Perimeter() / 2
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}
func (t Triangle) String() string {
	return fmt.Sprintf("Triangle(%.2f,%.2f,%.2f)", t.A, t.B, t.C)
}

func TotalArea(shapes []Shape) float64 {
	var total float64
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

func LargestShape(shapes []Shape) Shape {
	if len(shapes) == 0 {
		return nil
	}
	largest := shapes[0]
	for _, s := range shapes[1:] {
		if s.Area() > largest.Area() {
			largest = s
		}
	}
	return largest
}

func describe(s Shape) string {
	switch v := s.(type) {
	case Circle:
		return fmt.Sprintf("a circle with radius %.2f", v.Radius)
	case Rectangle:
		if v.Width == v.Height {
			return "a square rectangle"
		}
		return "a rectangular rectangle"
	case Triangle:
		return fmt.Sprintf("a triangle with sides %.2f, %.2f, %.2f", v.A, v.B, v.C)
	default:
		return "unknown shape"
	}
}

func main() {
	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 4, Height: 6},
		Triangle{A: 3, B: 4, C: 5},
		Circle{Radius: 2},
		Rectangle{Width: 8, Height: 8},
	}

	fmt.Printf("%-35s %10s %12s\n", "Shape", "Area", "Perimeter")
	fmt.Println("─────────────────────────────────────────────────────")
	for _, s := range shapes {
		fmt.Printf("%-35s %10.2f %12.2f\n", s, s.Area(), s.Perimeter())
	}

	fmt.Printf("\nTotal area:    %.2f\n", TotalArea(shapes))
	largest := LargestShape(shapes)
	fmt.Printf("Largest shape: %s (it's %s)\n", largest, describe(largest))
}
