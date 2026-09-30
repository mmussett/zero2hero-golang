package main

import (
	"fmt"
	"strings"

	"github.com/mmussett/zero2hero-golang/day-07/wordcount"
)

const text = `To be or not to be that is the question
Whether tis nobler in the mind to suffer
The slings and arrows of outrageous fortune
Or to take arms against a sea of troubles`

func main() {
	c := wordcount.New()
	c.AddText(text)

	fmt.Printf("Total words:  %d\n", c.Total())
	fmt.Printf("Unique words: %d\n", c.Unique())
	fmt.Println("\nTop 10 words:")
	fmt.Printf("%-4s %-20s %s\n", "Rank", "Word", "Count")
	fmt.Println(strings.Repeat("─", 35))
	for i, e := range c.TopN(10) {
		fmt.Printf("%-4d %-20s %d\n", i+1, e.Word, e.Count)
	}
}
