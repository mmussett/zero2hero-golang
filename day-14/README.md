# Day 14: Standard Library Interfaces

## Core Concept: Small Interfaces Compose Well

Go's philosophy: "the bigger the interface, the weaker the abstraction." Most standard library interfaces have one or two methods. Implement them and your type works everywhere.

## io.Reader and io.Writer

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

`Read` fills `p` and returns how many bytes were written. It returns `io.EOF` when done — not an error, a signal.

```go
// Anything that reads can be passed to io.Copy
io.Copy(os.Stdout, strings.NewReader("hello"))

// Compose readers
r := io.MultiReader(strings.NewReader("first"), strings.NewReader("second"))

// Wrap a writer for buffering
w := bufio.NewWriter(os.Stdout)
defer w.Flush()
```

## fmt.Stringer

```go
type Stringer interface {
    String() string
}
```

Implement `String()` and `fmt.Println`, `fmt.Sprintf("%v")`, etc. will call it automatically.

## sort.Interface

```go
type Interface interface {
    Len() int
    Less(i, j int) bool
    Swap(i, j int)
}
```

Implement it and call `sort.Sort(yourSlice)`. Use `sort.Reverse(yourSlice)` for descending order without re-implementing.

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
1. A `CountingReader` that wraps `io.Reader` and counts bytes read
2. A `TeeWriter` that writes to two `io.Writer`s simultaneously
3. A `Temperature` type with `fmt.Stringer`, `sort.Interface` on `[]Temperature`
4. A `Config` struct that implements `encoding.TextMarshaler` and `TextUnmarshaler`

Compose them: read from a file, count bytes, write to both stdout and a buffer.

**Extension ideas:** implement `io.WriterTo` and `io.ReaderFrom` on your types for direct copy optimisation.
