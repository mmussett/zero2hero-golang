package main

import (
	"fmt"
	"strings"
	"unicode"
)

func Add(a, b float64) float64 { return a + b }
func Sub(a, b float64) float64 { return a - b }
func Mul(a, b float64) float64 { return a * b }
func Div(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

func WordFrequency(text string) map[string]int {
	freq := make(map[string]int)
	clean := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, text)
	for _, w := range strings.Fields(clean) {
		freq[w]++
	}
	return freq
}

func IsPalindrome(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		if runes[i] != runes[j] {
			return false
		}
	}
	return true
}

func main() {
	fmt.Println("Day 13: run `go test ./...` to see the test suite in action")
	fmt.Printf("Add(2, 3)             = %.1f\n", Add(2, 3))
	fmt.Printf("IsPalindrome(racecar) = %v\n", IsPalindrome("racecar"))
}
