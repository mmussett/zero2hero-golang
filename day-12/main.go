package main

import (
	"fmt"
	"strings"
)

type Stage func([]string) []string

func Pipeline(data []string, stages ...Stage) []string {
	for _, s := range stages {
		data = s(data)
	}
	return data
}

func Lowercase() Stage {
	return func(data []string) []string {
		result := make([]string, len(data))
		for i, s := range data {
			result[i] = strings.ToLower(s)
		}
		return result
	}
}

func TrimSpaces() Stage {
	return func(data []string) []string {
		result := make([]string, len(data))
		for i, s := range data {
			result[i] = strings.TrimSpace(s)
		}
		return result
	}
}

func RemoveEmpty() Stage {
	return func(data []string) []string {
		result := make([]string, 0, len(data))
		for _, s := range data {
			if s != "" {
				result = append(result, s)
			}
		}
		return result
	}
}

func Deduplicate() Stage {
	return func(data []string) []string {
		seen := make(map[string]bool)
		result := make([]string, 0, len(data))
		for _, s := range data {
			if !seen[s] {
				seen[s] = true
				result = append(result, s)
			}
		}
		return result
	}
}

func FilterMinLength(n int) Stage {
	return func(data []string) []string {
		result := make([]string, 0)
		for _, s := range data {
			if len([]rune(s)) >= n {
				result = append(result, s)
			}
		}
		return result
	}
}

func Replace(old, new string) Stage {
	return func(data []string) []string {
		result := make([]string, len(data))
		for i, s := range data {
			result[i] = strings.ReplaceAll(s, old, new)
		}
		return result
	}
}

type ProcessorConfig struct {
	minLength int
	dedupe    bool
	prefix    string
}

type ProcessorOption func(*ProcessorConfig)

func WithMinLength(n int) ProcessorOption { return func(c *ProcessorConfig) { c.minLength = n } }
func WithDedup() ProcessorOption          { return func(c *ProcessorConfig) { c.dedupe = true } }
func WithPrefix(p string) ProcessorOption { return func(c *ProcessorConfig) { c.prefix = p } }

func NewProcessor(opts ...ProcessorOption) func([]string) []string {
	cfg := &ProcessorConfig{minLength: 1}
	for _, o := range opts {
		o(cfg)
	}
	return func(data []string) []string {
		stages := []Stage{TrimSpaces(), RemoveEmpty(), Lowercase()}
		if cfg.minLength > 1 {
			stages = append(stages, FilterMinLength(cfg.minLength))
		}
		if cfg.dedupe {
			stages = append(stages, Deduplicate())
		}
		result := Pipeline(data, stages...)
		if cfg.prefix != "" {
			for i, s := range result {
				result[i] = cfg.prefix + s
			}
		}
		return result
	}
}

func main() {
	input := []string{
		"  Hello  ", "world", "HELLO", "Go", "  ", "", "go",
		"generics", "world", "interfaces", "Go", "channels",
	}

	fmt.Println("Input:", input)
	fmt.Println()

	result := Pipeline(input,
		TrimSpaces(),
		RemoveEmpty(),
		Lowercase(),
		Deduplicate(),
		FilterMinLength(3),
	)
	fmt.Println("After pipeline:", result)

	fmt.Println("\n=== Functional Options Processor ===")
	process := NewProcessor(
		WithMinLength(4),
		WithDedup(),
		WithPrefix("→ "),
	)
	fmt.Println("Processed:", process(input))
}
