package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

const text = `To be or not to be that is the question
Whether tis nobler in the mind to suffer
The slings and arrows of outrageous fortune
Or to take arms against a sea of troubles
And by opposing end them to die to sleep
No more and by a sleep to say we end
The heartache and the thousand natural shocks
That flesh is heir to tis a consummation
Devoutly to be wished to die to sleep
To sleep perchance to dream`

type Entry struct {
	Word  string
	Count int
}

func wordFrequency(input string) []Entry {
	clean := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, input)

	freq := make(map[string]int)
	for _, w := range strings.Fields(clean) {
		if w != "" {
			freq[w]++
		}
	}

	entries := make([]Entry, 0, len(freq))
	for w, c := range freq {
		entries = append(entries, Entry{w, c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count != entries[j].Count {
			return entries[i].Count > entries[j].Count
		}
		return entries[i].Word < entries[j].Word
	})
	return entries
}

func topN(entries []Entry, n int) []Entry {
	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n]
}

func main() {
	entries := wordFrequency(text)
	top := topN(entries, 10)

	fmt.Println("=== Word Frequency Counter ===\n")
	fmt.Printf("%-4s %-20s %s\n", "Rank", "Word", "Count")
	fmt.Println(strings.Repeat("─", 35))
	for i, e := range top {
		fmt.Printf("%-4d %-20s %d\n", i+1, e.Word, e.Count)
	}
}
