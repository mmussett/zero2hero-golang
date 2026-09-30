package main

import (
	"fmt"
	"strings"
	"unicode"
)

const sampleText = `The quick brown fox jumps over the lazy dog.
Pack my box with five dozen liquor jugs.
How vexingly quick daft zebras jump!
The five boxing wizards jump quickly.
Sphinx of black quartz, judge my vow.`

type Stats struct {
	Chars       int
	Bytes       int
	Words       int
	Lines       int
	LongestWord string
	TopChar     rune
	TopCharFreq int
}

func analyse(text string) Stats {
	lines := strings.Split(text, "\n")
	words := strings.Fields(text)

	charFreq := make(map[rune]int)
	for _, r := range text {
		if !unicode.IsSpace(r) {
			charFreq[unicode.ToLower(r)]++
		}
	}

	var topChar rune
	var topFreq int
	for r, n := range charFreq {
		if n > topFreq || (n == topFreq && r < topChar) {
			topChar, topFreq = r, n
		}
	}

	longest := ""
	for _, w := range words {
		w = strings.Trim(w, ".,!?;:")
		if len([]rune(w)) > len([]rune(longest)) {
			longest = w
		}
	}

	return Stats{
		Chars:       len([]rune(text)),
		Bytes:       len(text),
		Words:       len(words),
		Lines:       len(lines),
		LongestWord: longest,
		TopChar:     topChar,
		TopCharFreq: topFreq,
	}
}

func main() {
	fmt.Println("=== String Statistics Tool ===\n")
	fmt.Println("Input text:")
	fmt.Println("─────────────────────────────")
	fmt.Println(sampleText)
	fmt.Println("─────────────────────────────\n")

	s := analyse(sampleText)
	fmt.Printf("Characters (runes): %d\n", s.Chars)
	fmt.Printf("Bytes:              %d\n", s.Bytes)
	fmt.Printf("Words:              %d\n", s.Words)
	fmt.Printf("Lines:              %d\n", s.Lines)
	fmt.Printf("Longest word:       %q\n", s.LongestWord)
	fmt.Printf("Most frequent char: %q (%d times)\n", string(s.TopChar), s.TopCharFreq)
}
