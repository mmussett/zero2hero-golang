// Day 33 – Channel Pipeline: Word Frequency Counter
//
// Architecture:
//
//	generator ──► fanOut ──► [worker 0..N-1] ──► merge/reduce ──► results
//
// Each stage communicates exclusively via channels. A done channel carries
// cancellation throughout the pipeline.
package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode"
)

// ── Source text ───────────────────────────────────────────────────────────────

const paragraph = `The quick brown fox jumps over the lazy dog. ` +
	`Go is an open source programming language that makes it easy to build ` +
	`simple reliable and efficient software. Goroutines and channels make ` +
	`concurrent programming natural and fun. The channel is the pipe that ` +
	`connects goroutines letting them communicate by sending and receiving values. ` +
	`A select statement lets a goroutine wait on multiple communication operations. `

// ── Stage 1: generator ───────────────────────────────────────────────────────

// generator emits every word from text into the returned channel, then closes
// it. It respects the done channel for early cancellation.
func generator(done <-chan struct{}, text string) <-chan string {
	out := make(chan string, 256)
	go func() {
		defer close(out)
		// Split on whitespace / punctuation
		words := strings.FieldsFunc(text, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
		for _, w := range words {
			w = strings.ToLower(w)
			if w == "" {
				continue
			}
			select {
			case out <- w:
			case <-done:
				return
			}
		}
	}()
	return out
}

// ── Stage 2: fanOut ───────────────────────────────────────────────────────────

// fanOut distributes words from in across n worker channels using round-robin.
func fanOut(done <-chan struct{}, in <-chan string, n int) []<-chan string {
	channels := make([]chan string, n)
	for i := range channels {
		channels[i] = make(chan string, 256)
	}

	go func() {
		defer func() {
			for _, ch := range channels {
				close(ch)
			}
		}()
		i := 0
		for word := range in {
			select {
			case channels[i%n] <- word:
				i++
			case <-done:
				return
			}
		}
	}()

	// Convert []chan string to []<-chan string
	out := make([]<-chan string, n)
	for i, ch := range channels {
		out[i] = ch
	}
	return out
}

// ── Stage 3: count ───────────────────────────────────────────────────────────

// count reads words from in, tallies them, then sends the partial map on the
// returned channel. One goroutine per worker channel.
func count(done <-chan struct{}, in <-chan string) <-chan map[string]int {
	out := make(chan map[string]int, 1)
	go func() {
		defer close(out)
		freq := make(map[string]int)
		for {
			select {
			case word, ok := <-in:
				if !ok {
					out <- freq
					return
				}
				freq[word]++
			case <-done:
				out <- freq
				return
			}
		}
	}()
	return out
}

// ── Stage 4: merge / reduce ───────────────────────────────────────────────────

// merge collects all partial frequency maps from each worker and reduces them
// into a single combined map.
func merge(partials []<-chan map[string]int) map[string]int {
	var wg sync.WaitGroup
	combined := make(map[string]int)
	var mu sync.Mutex

	for _, ch := range partials {
		wg.Add(1)
		go func(c <-chan map[string]int) {
			defer wg.Done()
			for partial := range c {
				mu.Lock()
				for word, cnt := range partial {
					combined[word] += cnt
				}
				mu.Unlock()
			}
		}(ch)
	}

	wg.Wait()
	return combined
}

// ── Helpers ───────────────────────────────────────────────────────────────────

type wordCount struct {
	Word  string
	Count int
}

// topN returns the n most frequent words, sorted descending by count.
func topN(freq map[string]int, n int) []wordCount {
	all := make([]wordCount, 0, len(freq))
	for w, c := range freq {
		all = append(all, wordCount{w, c})
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Count != all[j].Count {
			return all[i].Count > all[j].Count
		}
		return all[i].Word < all[j].Word
	})
	if n > len(all) {
		n = len(all)
	}
	return all[:n]
}

// ── Main ─────────────────────────────────────────────────────────────────────

func main() {
	const (
		repetitions = 5000
		numWorkers  = 4
		topWords    = 20
	)

	// Build a large text corpus
	var sb strings.Builder
	for i := 0; i < repetitions; i++ {
		sb.WriteString(paragraph)
	}
	text := sb.String()

	fmt.Printf("Pipeline: %d workers, corpus ~%d characters\n\n",
		numWorkers, len(text))

	// Cancellation channel (close to cancel the whole pipeline)
	done := make(chan struct{})
	defer close(done)

	// Stage 1: generate words
	words := generator(done, text)

	// Stage 2: fan out to N workers
	workerChans := fanOut(done, words, numWorkers)

	// Stage 3: count in each worker
	partials := make([]<-chan map[string]int, numWorkers)
	for i, ch := range workerChans {
		partials[i] = count(done, ch)
	}

	// Stage 4: merge / reduce
	freq := merge(partials)

	// Report
	fmt.Printf("Unique words: %d\n", len(freq))
	fmt.Printf("Top %d words:\n\n", topWords)
	fmt.Printf("%-20s %s\n", "WORD", "COUNT")
	fmt.Printf("%-20s %s\n", "--------------------", "-----")
	for _, wc := range topN(freq, topWords) {
		fmt.Printf("%-20s %d\n", wc.Word, wc.Count)
	}
}
