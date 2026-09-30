# Day 31: Go Interfaces — The Complete Picture

## Core Concept: Behaviour, Not Identity

An interface in Go is a contract expressed as a set of method signatures. A type satisfies the contract by implementing the methods — no declaration, no inheritance, no keyword. This is Go's most powerful design decision.

> "The bigger the interface, the weaker the abstraction." — Rob Pike

---

## Part 1: Fundamentals Revisited

### Interface Mechanics

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

An interface value is a pair: `(type, value)`. When you assign a concrete value to an interface variable, Go stores a pointer to the type descriptor and a pointer to the value.

```go
var w Writer = os.Stdout   // type=*os.File, value=<ptr>
w.Write([]byte("hello"))   // dynamic dispatch via type descriptor
```

### The Nil Interface Trap

The most common interface gotcha in Go:

```go
func newError() error {
    var p *os.PathError = nil
    return p          // DANGER: interface is NOT nil!
}

err := newError()
fmt.Println(err == nil)  // false — type is set, value is nil
```

An interface is `nil` only when **both** its type and value are nil. Returning a typed nil pointer as an interface always produces a non-nil interface.

**Fix:** return `nil` directly, not a typed nil:

```go
func newError() error {
    return nil  // correct
}
```

---

## Part 2: Interface Composition

Interfaces can embed other interfaces to form supersets:

```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
type Closer interface { Close() error }

// Composed interfaces
type ReadWriter  interface { Reader; Writer }
type ReadCloser  interface { Reader; Closer }
type WriteCloser interface { Writer; Closer }
type ReadWriteCloser interface { Reader; Writer; Closer }
```

A function accepting `io.Reader` is more reusable than one accepting `io.ReadWriter` — request only what you need.

---

## Part 3: The Standard Library Interface Catalogue

### Formatting Interfaces (`fmt` package)

| Interface | Method | Trigger |
|-----------|--------|---------|
| `fmt.Stringer` | `String() string` | `%s`, `%v` |
| `fmt.GoStringer` | `GoString() string` | `%#v` |
| `fmt.Formatter` | `Format(f State, verb rune)` | any `%` verb |
| `error` | `Error() string` | `%v`, `%s` on error |

### I/O Interfaces (`io` package)

| Interface | Methods | Usage |
|-----------|---------|-------|
| `io.Reader` | `Read([]byte) (int, error)` | Anything readable |
| `io.Writer` | `Write([]byte) (int, error)` | Anything writable |
| `io.Closer` | `Close() error` | Resources that must be released |
| `io.Seeker` | `Seek(int64, int) (int64, error)` | Random access |
| `io.ByteReader` | `ReadByte() (byte, error)` | Single-byte reading |
| `io.RuneReader` | `ReadRune() (rune, int, error)` | Unicode reading |
| `io.WriterTo` | `WriteTo(Writer) (int64, error)` | Efficient copy optimisation |
| `io.ReaderFrom` | `ReadFrom(Reader) (int64, error)` | Efficient copy optimisation |

### Encoding Interfaces

| Interface | Methods | Package |
|-----------|---------|---------|
| `encoding.TextMarshaler` | `MarshalText() ([]byte, error)` | `encoding` |
| `encoding.TextUnmarshaler` | `UnmarshalText([]byte) error` | `encoding` |
| `json.Marshaler` | `MarshalJSON() ([]byte, error)` | `encoding/json` |
| `json.Unmarshaler` | `UnmarshalJSON([]byte) error` | `encoding/json` |

### Sorting (`sort` package)

```go
type Interface interface {
    Len() int
    Less(i, j int) bool
    Swap(i, j int)
}
```

### HTTP (`net/http`)

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

`http.HandlerFunc` is a type that adapts a function to the `Handler` interface — Go's adapter pattern in one line of stdlib.

---

## Part 4: Type Assertions and Type Switches

### Safe Type Assertion

```go
var r io.Reader = &bytes.Buffer{}

// Two-value form: safe
bw, ok := r.(*bytes.Buffer)
if ok {
    fmt.Println("have buffer:", bw.Len())
}

// One-value form: panics if wrong type
bw2 := r.(*bytes.Buffer)
```

### Type Switch

Dispatches on the dynamic type of an interface value:

```go
func printType(v any) {
    switch x := v.(type) {
    case nil:
        fmt.Println("nil")
    case int:
        fmt.Printf("int: %d\n", x)
    case string:
        fmt.Printf("string: %q\n", x)
    case []int:
        fmt.Printf("[]int with %d elements\n", len(x))
    case fmt.Stringer:
        fmt.Printf("Stringer: %s\n", x.String())
    default:
        fmt.Printf("unknown: %T\n", x)
    }
}
```

Order matters — interfaces listed earlier match first.

---

## Part 5: Idiomatic Interface Design

### Rule 1: Accept Interfaces, Return Structs

```go
// Good: caller can pass any Writer
func SaveConfig(w io.Writer, cfg Config) error { ... }

// Bad: locks caller to *os.File
func SaveConfig(f *os.File, cfg Config) error { ... }
```

Return concrete types from constructors so callers can access the full API. Return interfaces only when the caller genuinely needs polymorphism.

### Rule 2: Keep Interfaces Small

One or two methods is almost always better than three or more. The `io` package is a masterclass:

```go
type Reader interface { Read([]byte) (int, error) }  // one method
type Writer interface { Write([]byte) (int, error) }  // one method
```

### Rule 3: Define Interfaces at the Point of Use

Define interfaces in the **consumer** package, not the producer. The producer returns a concrete type; the consumer declares the slice of the API it needs.

```go
// In package store: returns concrete type
func (s *Store) Open() *Connection { ... }

// In package handler: defines only what it needs
type Opener interface {
    Open() *Connection
}
```

This prevents import cycles and keeps interfaces honest.

### Rule 4: Don't Pre-Optimise with Interfaces

If you only have one implementation, use the concrete type. Add an interface when:
- You have two or more implementations
- You need to mock in tests
- You need to decouple packages

---

## Part 6: Interface Embedding in Structs

```go
type LoggingWriter struct {
    io.Writer
    log *slog.Logger
}

func (lw *LoggingWriter) Write(p []byte) (int, error) {
    lw.log.Debug("write", "bytes", len(p))
    return lw.Writer.Write(p)
}
```

The promoted `Write` method is overridden — `io.Copy(lw, src)` calls the logging version. All other methods from the embedded `io.Writer` are promoted unchanged.

---

## Part 7: Mocking with Interfaces for Testing

```go
// Production code uses interface
type EmailSender interface {
    Send(to, subject, body string) error
}

type WelcomeService struct {
    email EmailSender
}

// Test provides a fake
type fakeSender struct {
    sent []string
}

func (f *fakeSender) Send(to, subject, body string) error {
    f.sent = append(f.sent, to)
    return nil
}

func TestWelcome(t *testing.T) {
    fake := &fakeSender{}
    svc := &WelcomeService{email: fake}
    svc.SendWelcome("alice@example.com")
    assert.Contains(t, fake.sent, "alice@example.com")
}
```

---

## Part 8: Interface Anti-Patterns

### Interface Pollution

Defining large, catch-all interfaces:

```go
// Anti-pattern: too wide, too specific to one implementation
type Repository interface {
    FindAll() []User
    FindByID(id int) (User, error)
    FindByEmail(email string) (User, error)
    Save(u User) error
    Delete(id int) error
    Count() int
    Paginate(page, size int) []User
    Search(query string) []User
}
```

Split by caller need: a handler that only reads needs `UserReader`; a cleanup job that only deletes needs `UserDeleter`.

### Returning Interface from Constructor

```go
// Anti-pattern: hides the concrete type, limits caller
func NewStore() StoreInterface { return &sqlStore{} }

// Better: return concrete, satisfy interfaces implicitly
func NewStore() *SQLStore { return &sqlStore{} }
```

### Using `any` as a Shortcut

```go
// Anti-pattern: no type safety
func Process(input any) any { ... }

// Better: generics or concrete types
func Process[T Processable](input T) T { ... }
```

---

## Day Project: Interface Showcase — Advanced

Build a pluggable notification system:

1. Define `Notifier interface { Notify(msg Message) error }`
2. Implement: `EmailNotifier`, `SlackNotifier`, `LogNotifier` (writes to `slog`)
3. `MultiNotifier` that fans out to N notifiers, collects errors with `errors.Join`
4. A `RetryNotifier` that wraps any `Notifier` and retries on failure
5. A `FilterNotifier` that wraps a `Notifier` and only passes messages matching a predicate

Test with the `LogNotifier` as a fake.

```go
type Message struct {
    Level   string
    Subject string
    Body    string
}
```

**Extension ideas:** add a `RateLimitedNotifier` using `time.Ticker`; make notifiers configurable via `WithOption` functional options.
