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

## Labs

### Lab 1: io.Reader — ROT13Reader

**What you'll practise:** Implementing `io.Reader` by wrapping another Reader.

**Task:**
Implement a `ROT13Reader` struct that wraps an `io.Reader` and transforms each byte by applying the ROT13 cipher (rotate alphabetic characters by 13). Test it using `strings.NewReader` as the source.

**Steps:**
1. Define `type ROT13Reader struct { r io.Reader }`
2. Implement `Read(p []byte) (int, error)` — call `r.Read`, then rotate each byte in-place
3. Test by reading `"Hello, World!"` through the ROT13Reader and printing the result
4. Run it a second time through another ROT13Reader — confirm you get the original string back

```go
type ROT13Reader struct{ r io.Reader }

func (rr ROT13Reader) Read(p []byte) (int, error) {
    n, err := rr.r.Read(p)
    for i := 0; i < n; i++ {
        b := p[i]
        switch {
        case b >= 'A' && b <= 'Z':
            p[i] = 'A' + (b-'A'+13)%26
        case b >= 'a' && b <= 'z':
            p[i] = 'a' + (b-'a'+13)%26
        }
    }
    return n, err
}
```

**Expected output:**
```
Uryyb, Jbeyq!
Hello, World!
```

**Checkpoint:** `io.ReadAll` on your ROT13Reader returns the correctly rotated bytes.

---

### Lab 2: io.Writer — CountingWriter

**What you'll practise:** Implementing `io.Writer` as a transparent wrapper that counts bytes.

**Task:**
Implement a `CountingWriter` that wraps any `io.Writer` and keeps a running total of bytes written. Use `fmt.Fprintf` to write formatted output through it and verify the byte count.

**Steps:**
1. Define `type CountingWriter struct { w io.Writer; n int64 }`
2. Implement `Write(p []byte) (int, error)` — delegate to `w`, add `len(p)` to `n`
3. Write several strings through it using `fmt.Fprintf`
4. Assert the final count matches the sum of all written string lengths

```go
type CountingWriter struct {
    w io.Writer
    n int64
}

func (cw *CountingWriter) Write(p []byte) (int, error) {
    n, err := cw.w.Write(p)
    cw.n += int64(n)
    return n, err
}

func (cw *CountingWriter) BytesWritten() int64 { return cw.n }
```

**Expected output:**
```
Hello, Go!
Bytes written: 10
```

**Checkpoint:** `cw.BytesWritten()` equals `len("Hello, Go!\n")` after the `fmt.Fprintf` call.

---

### Lab 3: io.ReadWriter — Custom TeeWriter

**What you'll practise:** Writing to two destinations simultaneously by composing `io.Writer`.

**Task:**
Implement a `TeeWriter` that writes to two `io.Writer` destinations at once (similar to `io.MultiWriter` but built from scratch). If the first write succeeds but the second fails, return the error from the second.

**Steps:**
1. Define `type TeeWriter struct { a, b io.Writer }`
2. Implement `Write` — write to `a`, then to `b`, return the first error encountered
3. Use it to write to both `os.Stdout` and a `bytes.Buffer` simultaneously
4. Verify that the buffer contains the same bytes that appeared on stdout

```go
type TeeWriter struct{ a, b io.Writer }

func (tw *TeeWriter) Write(p []byte) (int, error) {
    n, err := tw.a.Write(p)
    if err != nil {
        return n, err
    }
    _, err = tw.b.Write(p)
    return n, err
}
```

**Expected output:**
```
stdout: Hello from TeeWriter
buffer: Hello from TeeWriter
```

**Checkpoint:** `buf.String()` equals the string sent to the `TeeWriter`.

---

### Lab 4: fmt.Stringer and sort.Interface — Temperature

**What you'll practise:** Implementing two stdlib interfaces on a custom type to make it printable and sortable.

**Task:**
Define a `Temperature` type (wrapping `float64`). Implement `String() string` so `fmt.Println` formats it as `"23.5°C"`. Then implement `sort.Interface` on `[]Temperature` and sort a slice in ascending order.

**Steps:**
1. Define `type Temperature float64`
2. Implement `String()` using `fmt.Sprintf("%.1f°C", float64(t))`
3. Define `type ByTemp []Temperature` and implement `Len`, `Less`, `Swap`
4. Sort a slice and print it before and after sorting

```go
type Temperature float64

func (t Temperature) String() string {
    return fmt.Sprintf("%.1f°C", float64(t))
}

type ByTemp []Temperature

func (b ByTemp) Len() int           { return len(b) }
func (b ByTemp) Less(i, j int) bool { return b[i] < b[j] }
func (b ByTemp) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
```

**Expected output:**
```
Before: [36.6°C 0.0°C 100.0°C -40.0°C]
After:  [-40.0°C 0.0°C 36.6°C 100.0°C]
```

**Checkpoint:** `fmt.Println(temps[0])` prints with the `°C` suffix, and `sort.Sort(ByTemp(temps))` orders them numerically.

---

### Lab 5: encoding.TextMarshaler/TextUnmarshaler — Config

**What you'll practise:** Implementing the text codec interfaces so a struct round-trips through a `key=value` text format.

**Task:**
Define a `Config` struct with `Host string`, `Port int`, and `Debug bool` fields. Implement `MarshalText` to produce a `key=value\n` format and `UnmarshalText` to parse it back. Verify the round-trip.

**Steps:**
1. Implement `MarshalText() ([]byte, error)` using `fmt.Sprintf` for each field
2. Implement `UnmarshalText(text []byte) error` using `bufio.Scanner` and `strings.Cut`
3. Marshal a Config, print the text, then unmarshal it into a second Config
4. Assert both Configs are equal

```go
func (c Config) MarshalText() ([]byte, error) {
    return []byte(fmt.Sprintf("host=%s\nport=%d\ndebug=%v\n",
        c.Host, c.Port, c.Debug)), nil
}

func (c *Config) UnmarshalText(text []byte) error {
    scanner := bufio.NewScanner(bytes.NewReader(text))
    for scanner.Scan() {
        key, val, _ := strings.Cut(scanner.Text(), "=")
        // set fields by key...
    }
    return scanner.Err()
}
```

**Expected output:**
```
Marshalled:
host=localhost
port=8080
debug=true

Round-trip equal: true
```

**Checkpoint:** The unmarshalled Config has the same field values as the original.

---

### Lab 6: io.Closer — Logging Close Wrapper

**What you'll practise:** Implementing `io.Closer` and using `defer` for guaranteed resource release.

**Task:**
Create a `LoggingFile` struct that wraps `*os.File`. Override `Close()` to log a message before delegating. Open a real file, write to it, and use `defer lf.Close()` to close it — confirm the log message appears even when an error occurs.

**Steps:**
1. Define `type LoggingFile struct { f *os.File; name string }`
2. Forward `Write` to `f.Write`
3. Implement `Close() error` — log `"closing <name>"`, then call `f.Close()`
4. Open a temp file, write a line, defer Close, return from the function — observe the log

```go
type LoggingFile struct {
    f    *os.File
    name string
}

func (lf *LoggingFile) Write(p []byte) (int, error) { return lf.f.Write(p) }

func (lf *LoggingFile) Close() error {
    fmt.Printf("closing %s\n", lf.name)
    return lf.f.Close()
}
```

**Expected output:**
```
wrote 12 bytes
closing /tmp/demo123.txt
```

**Checkpoint:** The "closing" log line appears before the program exits, even if you add an early `return`.

---

### Lab 7: Compose Stdlib Interfaces — LimitedLogger

**What you'll practise:** Composing multiple stdlib interfaces into a single type.

**Task:**
Build a `LimitedLogger` that wraps an `io.Writer` and implements `io.Writer` itself. It enforces a maximum byte budget: once the budget is exhausted it returns `0, ErrBudgetExhausted` instead of writing. Use it as the output for a `log.Logger`.

**Steps:**
1. Define `type LimitedLogger struct { w io.Writer; budget int64; written int64 }`
2. Implement `Write(p []byte) (int, error)` — check remaining budget, truncate or reject
3. Pass it to `log.New(ll, "PREFIX: ", 0)` and write enough log entries to hit the limit
4. Confirm that writes beyond the budget return the custom error

```go
var ErrBudgetExhausted = errors.New("log budget exhausted")

func (ll *LimitedLogger) Write(p []byte) (int, error) {
    remaining := ll.budget - ll.written
    if remaining <= 0 {
        return 0, ErrBudgetExhausted
    }
    if int64(len(p)) > remaining {
        p = p[:remaining]
    }
    n, err := ll.w.Write(p)
    ll.written += int64(n)
    return n, err
}
```

**Expected output:**
```
PREFIX: message 1
PREFIX: message 2
PREFIX: messag  <- truncated at budget
write error: log budget exhausted
```

**Checkpoint:** `ll.written` equals the budget limit after the budget is exceeded, and further writes return `ErrBudgetExhausted`.

---

### Final Lab (Project): Interface Showcase

**What you'll practise:** Combining all stdlib interfaces — `io.Reader`, `io.Writer`, `fmt.Stringer`, `sort.Interface`, `encoding.TextMarshaler/TextUnmarshaler` — into a coherent program.

**Task:**
Build a program that demonstrates all four interface implementations working together.

**Steps:**
1. Implement `CountingReader` wrapping `io.Reader` — counts bytes read
2. Implement `TeeWriter` writing to two `io.Writer`s simultaneously
3. Implement `Temperature` with `fmt.Stringer` and `sort.Interface` on `[]Temperature`
4. Implement `Config` with `encoding.TextMarshaler` and `TextUnmarshaler`
5. Compose: read from a file through `CountingReader`, write to both stdout and a buffer via `TeeWriter`, sort a temperature slice, and round-trip a Config

```go
// Wire them together in main:
cr := &CountingReader{r: file}
tw := &TeeWriter{a: os.Stdout, b: &buf}
io.Copy(tw, cr)
fmt.Printf("read %d bytes\n", cr.N)

sort.Sort(ByTemp(temps))
fmt.Println(temps)

text, _ := cfg.MarshalText()
var cfg2 Config
cfg2.UnmarshalText(text)
```

**Expected output:**
```
[file contents appear on stdout and in buffer]
read 42 bytes
[-40.0°C 0.0°C 36.6°C 100.0°C]
round-trip ok: true
```

**Checkpoint:** All four components run without error and the `CountingReader` count matches the actual file size.

---

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
