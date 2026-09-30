# Day 05: Arrays, Slices, and Strings

## Arrays (Rarely Used Directly)

Arrays have a **compile-time fixed length** and are **value types**:

```go
var a [5]int              // [0 0 0 0 0]
b := [3]string{"x","y","z"}
c := [...]int{1,2,3}      // compiler counts: [3]int
```

Because `[3]int` and `[4]int` are different types, functions usually accept slices instead.

## Slices (The Idiomatic List)

A slice is a view into an underlying array: three words `(pointer, length, capacity)`.

```go
s := []int{10, 20, 30}
s = append(s, 40)         // may reallocate; always use returned slice
s2 := s[1:3]              // [20 30] — shares memory with s
```

Key idioms:

```go
// Build a slice of known length
result := make([]string, 0, len(input))
for _, v := range input {
    result = append(result, process(v))
}

// Copy to avoid shared backing array
clone := make([]int, len(s))
copy(clone, s)
```

`nil` slice has `len == 0` and `cap == 0`. It is safe to `append` to a nil slice.

## Strings

A `string` is a read-only `[]byte` (UTF-8). `len(s)` is the byte count, not the character count.

```go
s := "héllo"
len(s)          // 6 (é is 2 UTF-8 bytes)
len([]rune(s))  // 5 (rune count)

for i, r := range s {   // i = byte offset, r = rune
    fmt.Printf("%d: %c (%d)\n", i, r, r)
}
```

### Essential [`strings`](https://pkg.go.dev/strings) Functions

```go
strings.Contains(s, "ell")
strings.HasPrefix(s, "he")
strings.HasSuffix(s, "lo")
strings.Count(s, "l")
strings.ToUpper(s) / strings.ToLower(s)
strings.TrimSpace(s)
strings.Split(s, ",")
strings.Join(parts, ", ")
strings.Replace(s, "old", "new", -1)
strings.Fields(s)   // split on any whitespace
```

### Building Strings Efficiently

```go
var b strings.Builder
for _, word := range words {
    b.WriteString(word)
    b.WriteByte(' ')
}
result := b.String()
```

Concatenating with `+=` in a loop is O(n²) — always use [`strings.Builder`](https://pkg.go.dev/strings#Builder) or [`strings.Join`](https://pkg.go.dev/strings#Join).

## Day Project: String Statistics Tool

Write a program that takes a multi-line string (hardcoded) and computes:
- Total character count (runes)
- Total byte count
- Word count (using [`strings.Fields`](https://pkg.go.dev/strings#Fields))
- Line count
- Longest word
- Most frequent character (use a `map[rune]int`)

**Extension ideas:** read from `os.Stdin`; add sentence count; produce a frequency bar chart in the terminal.

## Official Documentation

- [`strings`](https://pkg.go.dev/strings) — Contains, Fields, Builder, Join, Split, ToUpper, TrimSpace, and more
- [`unicode`](https://pkg.go.dev/unicode) — rune classification (IsLetter, IsDigit, etc.)
- [`fmt`](https://pkg.go.dev/fmt) — formatted I/O with `%c` rune verb
- [`os`](https://pkg.go.dev/os) — `os.Stdin` for reading standard input
- [Language Spec: Slice types](https://go.dev/ref/spec#Slice_types) — slice internals
- [Language Spec: String types](https://go.dev/ref/spec#String_types) — UTF-8 string representation
- [Go Blog: Strings, bytes, runes and characters](https://go.dev/blog/strings) — deep dive on Go strings
- [Go Tour: More types](https://go.dev/tour/moretypes/7) — slices and arrays tour
