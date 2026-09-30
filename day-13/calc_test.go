package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	cases := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"positive", 2, 3, 5},
		{"negative", -1, -2, -3},
		{"mixed", 10, -3, 7},
		{"zero", 0, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Add(tc.a, tc.b))
		})
	}
}

func TestSub(t *testing.T) {
	assert.Equal(t, 1.0, Sub(3, 2))
	assert.Equal(t, -5.0, Sub(0, 5))
}

func TestMul(t *testing.T) {
	assert.Equal(t, 6.0, Mul(2, 3))
	assert.Equal(t, 0.0, Mul(0, 99))
}

func TestDiv(t *testing.T) {
	cases := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr bool
	}{
		{"normal", 10, 2, 5, false},
		{"negative divisor", -10, 2, -5, false},
		{"divide by zero", 1, 0, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Div(tc.a, tc.b)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWordFrequency(t *testing.T) {
	freq := WordFrequency("the cat sat on the mat the cat")
	assert.Equal(t, 3, freq["the"])
	assert.Equal(t, 2, freq["cat"])
	assert.Equal(t, 1, freq["sat"])
	assert.Equal(t, 0, freq["dog"])
}

func TestIsPalindrome(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"racecar", true},
		{"hello", false},
		{"level", true},
		{"A", true},
		{"", true},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.want, IsPalindrome(tc.input))
		})
	}
}

func BenchmarkWordFrequency(b *testing.B) {
	text := "the quick brown fox jumps over the lazy dog the fox"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WordFrequency(text)
	}
}
