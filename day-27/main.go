// Package main demonstrates three word-count implementations and compares
// their wall-clock performance on a large generated corpus.
package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

// ---- Implementation 1: naive (strings.Split by space) --------------------

// countNaive splits on a single space and counts non-empty tokens.
func countNaive(text string) int {
	parts := strings.Split(text, " ")
	count := 0
	for _, p := range parts {
		if p != "" {
			count++
		}
	}
	return count
}

// ---- Implementation 2: scanner (bufio.Scanner with ScanWords) ------------

// countScanner uses bufio.Scanner configured with ScanWords.
func countScanner(text string) int {
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Split(bufio.ScanWords)
	count := 0
	for sc.Scan() {
		count++
	}
	return count
}

// ---- Implementation 3: manual byte scanning ------------------------------

// countManual scans bytes directly, toggling an in-word flag.
func countManual(text string) int {
	count := 0
	inWord := false
	for i := 0; i < len(text); i++ {
		c := text[i]
		isSpace := c == ' ' || c == '\t' || c == '\n' || c == '\r'
		if isSpace {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}

// ---- Timing helper -------------------------------------------------------

type result struct {
	name  string
	count int
	dur   time.Duration
}

func bench(name string, fn func(string) int, text string) result {
	start := time.Now()
	count := fn(text)
	return result{name: name, count: count, dur: time.Since(start)}
}

// ---- Main ----------------------------------------------------------------

const loremIpsum = "Lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod " +
	"tempor incididunt ut labore et dolore magna aliqua Ut enim ad minim veniam " +
	"quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat " +
	"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu " +
	"fugiat nulla pariatur Excepteur sint occaecat cupidatat non proident sunt in " +
	"culpa qui officia deserunt mollit anim id est laborum "

func main() {
	const repeat = 10_000
	corpus := strings.Repeat(loremIpsum, repeat)

	fmt.Printf("Corpus size: %d bytes (~%.1f MB)\n\n", len(corpus), float64(len(corpus))/(1<<20))

	results := []result{
		bench("naive   (strings.Split)", countNaive, corpus),
		bench("scanner (bufio.ScanWords)", countScanner, corpus),
		bench("manual  (byte scan)", countManual, corpus),
	}

	// Print comparison table.
	fmt.Printf("%-30s  %10s  %10s\n", "Implementation", "Words", "Duration")
	fmt.Println(strings.Repeat("-", 56))
	for _, r := range results {
		fmt.Printf("%-30s  %10d  %10s\n", r.name, r.count, r.dur.Round(time.Microsecond))
	}
	fmt.Println()
	fmt.Println("Run benchmarks with:  go test -bench=. -benchmem ./...")
}
