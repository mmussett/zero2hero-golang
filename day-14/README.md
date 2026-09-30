# Day 14: Standard Library Interfaces

## Core Concept: Small Interfaces Compose Well

Go's philosophy: "the bigger the interface, the weaker the abstraction." Most standard library interfaces have one or two methods. Implement them and your type works everywhere.

## [io.Reader](https://pkg.go.dev/io#Reader) and [io.Writer](https://pkg.go.dev/io#Writer)

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

`Read` fills `p` and returns how many bytes were written. It returns [`io.EOF`](https://pkg.go.dev/io#EOF) when done — not an error, a signal.

```go
// Anything that reads can be passed to io.Copy
io.Copy(os.Stdout, strings.NewReader("hello"))

// Compose readers
r := io.MultiReader(strings.NewReader("first"), strings.NewReader("second"))

// Wrap a writer for buffering
w := bufio.NewWriter(os.Stdout)
defer w.Flush()
```

## [fmt.Stringer](https://pkg.go.dev/fmt#Stringer)

```go
type Stringer interface {
    String() string
}
```

Implement `String()` and `fmt.Println`, `fmt.Sprintf("%v")`, etc. will call it automatically.

## [sort.Interface](https://pkg.go.dev/sort#Interface)

```go
type Interface interface {
    Len() int
    Less(i, j int) bool
    Swap(i, j int)
}
```

Implement it and call `sort.Sort(yourSlice)`. Use [`sort.Reverse`](https://pkg.go.dev/sort#Reverse) for descending order without re-implementing.

## encoding.TextMarshaler / TextUnmarshaler

```go
type TextMarshaler interface {
    MarshalText() ([]byte, error)
}
type TextUnmarshaler interface {
    UnmarshalText(text []byte) error
}
```

Types that implement these are automatically handled by `encoding/json`, `encoding/xml`, and `fmt`.

## Struct Embedding to Satisfy Interfaces

```go
type LoggingWriter struct {
    io.Writer            // embed — promotes Write method
    prefix string
}

func (lw *LoggingWriter) Write(p []byte) (int, error) {
    fmt.Printf("[%s] %s", lw.prefix, p)
    return lw.Writer.Write(p)  // delegate to embedded writer
}
```

## Day Project: Interface Showcase

Implement:
1. A `CountingReader` that wraps [`io.Reader`](https://pkg.go.dev/io#Reader) and counts bytes read
2. A `TeeWriter` that writes to two [`io.Writer`](https://pkg.go.dev/io#Writer)s simultaneously
3. A `Temperature` type with [`fmt.Stringer`](https://pkg.go.dev/fmt#Stringer), [`sort.Interface`](https://pkg.go.dev/sort#Interface) on `[]Temperature`
4. A `Config` struct that implements `encoding.TextMarshaler` and `TextUnmarshaler`

Compose them: read from a file, count bytes, write to both stdout and a buffer.

**Extension ideas:** implement [`io.WriterTo`](https://pkg.go.dev/io#WriterTo) and [`io.ReaderFrom`](https://pkg.go.dev/io#ReaderFrom) on your types for direct copy optimisation.

## Official Documentation

- [`io`](https://pkg.go.dev/io) — Reader, Writer, EOF, Copy, MultiReader, WriterTo, ReaderFrom
- [`bufio`](https://pkg.go.dev/bufio) — buffered I/O wrappers (NewWriter, Flush)
- [`fmt`](https://pkg.go.dev/fmt) — Stringer interface, Printf, Sprintf
- [`sort`](https://pkg.go.dev/sort) — sort.Interface, Sort, Reverse
- [`strings`](https://pkg.go.dev/strings) — NewReader for in-memory string reading
- [`os`](https://pkg.go.dev/os) — Stdout, file I/O
- [Language Spec: Interface types](https://go.dev/ref/spec#Interface_types) — composing interfaces
- [Effective Go: Interfaces and other types](https://go.dev/doc/effective_go#interfaces_and_types) — interface design
- [Go Blog: Laws of Reflection](https://go.dev/blog/laws-of-reflection) — how interfaces work under the hood
