package main

import (
	"strings"
	"testing"
)

const benchCorpus = "the quick brown fox jumps over the lazy dog "

// BenchmarkWordCountNaive benchmarks the strings.Split implementation.
func BenchmarkWordCountNaive(b *testing.B) {
	text := strings.Repeat(benchCorpus, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countNaive(text)
	}
}

// BenchmarkWordCountScanner benchmarks the bufio.Scanner implementation.
func BenchmarkWordCountScanner(b *testing.B) {
	text := strings.Repeat(benchCorpus, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countScanner(text)
	}
}

// BenchmarkWordCountManual benchmarks the manual byte-scanning implementation.
func BenchmarkWordCountManual(b *testing.B) {
	text := strings.Repeat(benchCorpus, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countManual(text)
	}
}

// Parallel variants -------------------------------------------------------

// BenchmarkWordCountNaiveParallel benchmarks naive implementation in parallel.
func BenchmarkWordCountNaiveParallel(b *testing.B) {
	text := strings.Repeat(benchCorpus, 1000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			countNaive(text)
		}
	})
}

// BenchmarkWordCountScannerParallel benchmarks scanner implementation in parallel.
func BenchmarkWordCountScannerParallel(b *testing.B) {
	text := strings.Repeat(benchCorpus, 1000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			countScanner(text)
		}
	})
}

// BenchmarkWordCountManualParallel benchmarks manual implementation in parallel.
func BenchmarkWordCountManualParallel(b *testing.B) {
	text := strings.Repeat(benchCorpus, 1000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			countManual(text)
		}
	})
}
