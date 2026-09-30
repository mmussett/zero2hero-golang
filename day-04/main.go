package main

import "fmt"

type Box struct {
	value int
}

func (b *Box) Set(v int) { b.value = v }
func (b Box) Get() int   { return b.value }

func (b Box) String() string {
	return fmt.Sprintf("Box{value: %d}", b.value)
}

func demonstratePointers() {
	x := 42
	p := &x
	fmt.Printf("x  = %d\n", x)
	fmt.Printf("&x = %p\n", p)
	fmt.Printf("*p = %d\n", *p)

	*p = 100
	fmt.Printf("After *p = 100: x = %d\n", x)
}

func demonstrateBox() {
	b := Box{}
	b.Set(10)
	fmt.Printf("Box after Set(10): %s\n", b)

	// Value copy — original unchanged
	b2 := b
	b2.Set(99)
	fmt.Printf("b  = %s (unchanged)\n", b)
	fmt.Printf("b2 = %s (copy modified)\n", b2)

	// Pointer share — original changed
	b3 := &b
	b3.Set(77)
	fmt.Printf("b  = %s (changed via pointer)\n", b)
}

func demonstrateNilCheck() {
	var p *int
	fmt.Printf("nil pointer: %v\n", p)
	if p != nil {
		fmt.Printf("value: %d\n", *p)
	} else {
		fmt.Println("pointer is nil — skipping dereference")
	}
}

func main() {
	fmt.Println("=== Pointer Operators ===")
	demonstratePointers()

	fmt.Println("\n=== Box Struct ===")
	demonstrateBox()

	fmt.Println("\n=== Nil Pointer Safety ===")
	demonstrateNilCheck()
}
