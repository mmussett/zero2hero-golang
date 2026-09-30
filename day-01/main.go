package main

import (
	"fmt"
	"runtime"
)

func main() {
	name    := "Gopher"
	day     := 1
	message := "The journey of a thousand miles begins with a single go run."

	fmt.Printf("Day %d: Hello, %s!\n", day, name)
	fmt.Printf("Motivation: %s\n", message)
	fmt.Printf("Go version: %s\n", runtime.Version())
}
