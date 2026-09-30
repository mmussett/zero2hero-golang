package wordcount

import (
	"sort"
	"strings"
	"unicode"
)

type Entry struct {
	Word  string
	Count int
}

type Counter struct {
	counts map[string]int
	total  int
}

func New() *Counter {
	return &Counter{counts: make(map[string]int)}
}

func (c *Counter) Add(word string) {
	word = strings.ToLower(strings.TrimFunc(word, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}))
	if word == "" {
		return
	}
	c.counts[word]++
	c.total++
}

func (c *Counter) AddText(text string) {
	for _, w := range strings.Fields(text) {
		c.Add(w)
	}
}

func (c *Counter) TopN(n int) []Entry {
	entries := make([]Entry, 0, len(c.counts))
	for w, cnt := range c.counts {
		entries = append(entries, Entry{w, cnt})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count != entries[j].Count {
			return entries[i].Count > entries[j].Count
		}
		return entries[i].Word < entries[j].Word
	})
	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n]
}

func (c *Counter) Total() int  { return c.total }
func (c *Counter) Reset()      { c.counts = make(map[string]int); c.total = 0 }
func (c *Counter) Unique() int { return len(c.counts) }
