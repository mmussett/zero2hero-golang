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

---

## Labs

### Lab 1: Implicit Satisfaction

**What you'll practise:** Defining an interface and satisfying it implicitly, with a compile-time check.

**Task:**
Define a `Quacker` interface with a `Quack() string` method. Create two types — `Duck` and `Person` — that implement it without any explicit declaration. Add compile-time assertion guards to catch mistakes early.

**Steps:**
1. Create a new file `lab01_quacker.go` in `day-31/`
2. Define `type Quacker interface { Quack() string }`
3. Add `Duck` and `Person` structs with `Quack()` methods
4. Add compile-time checks using the blank identifier
5. Write a `MakeItQuack(q Quacker)` function and call it from `main`

```go
// Compile-time interface satisfaction checks
var _ Quacker = (*Duck)(nil)
var _ Quacker = (*Person)(nil)

type Duck struct{ Name string }
func (d *Duck) Quack() string { return "Quack! I'm " + d.Name }

type Person struct{ Name string }
func (p *Person) Quack() string { return "I'm quacking like a duck! — " + p.Name }
```

**Expected output:**
```
Quack! I'm Donald
I'm quacking like a duck! — Bob
```

**Checkpoint:** Remove the `Quack()` method from `Person` and confirm it fails to compile with a clear error about missing method.

---

### Lab 2: Interface Composition

**What you'll practise:** Composing small interfaces into larger ones, mirroring the `io` package design.

**Task:**
Define `Reader`, `Writer`, and `Closer` interfaces. Compose `ReadWriter`, `ReadCloser`, and `ReadWriteCloser`. Implement a `BufferRWC` struct that satisfies all three composed interfaces.

**Steps:**
1. Define three single-method interfaces
2. Compose them into three combined interfaces
3. Implement `BufferRWC` with all three methods
4. Add a function `processRWC(rwc ReadWriteCloser)` that calls all three
5. Verify `*BufferRWC` satisfies all composed interfaces via compile-time checks

```go
type Reader  interface { Read()  string }
type Writer  interface { Write(s string) }
type Closer  interface { Close() error }

type ReadWriter      interface { Reader; Writer }
type ReadCloser      interface { Reader; Closer }
type ReadWriteCloser interface { Reader; Writer; Closer }

// Compile-time checks
var _ ReadWriteCloser = (*BufferRWC)(nil)
```

**Expected output:**
```
Read: hello
Written: world
Closed successfully
```

**Checkpoint:** Try passing a `*BufferRWC` to a function that accepts `ReadCloser` — it should work. Try passing it to a function that accepts `ReadWriter` — also works. Explain why in a comment.

---

### Lab 3: The Nil Interface Trap

**What you'll practise:** Understanding why a typed nil pointer assigned to an interface is NOT nil.

**Task:**
Create a custom error type `*ValidationError`. Write a function that conditionally returns it. Demonstrate the trap and then write the correct pattern.

**Steps:**
1. Define `type ValidationError struct { Field, Message string }`
2. Implement `func (e *ValidationError) Error() string`
3. Write `func validateAgeBroken(age int) error` that returns a typed nil when valid
4. Call it and print `err == nil` — observe it prints `false`
5. Fix by returning plain `nil` directly

```go
func validateAgeBroken(age int) error {
    var err *ValidationError  // typed nil
    if age < 0 {
        err = &ValidationError{Field: "age", Message: "must be non-negative"}
    }
    return err  // BUG: always returns non-nil interface!
}

func validateAgeFixed(age int) error {
    if age < 0 {
        return &ValidationError{Field: "age", Message: "must be non-negative"}
    }
    return nil  // correct: untyped nil
}
```

**Expected output:**
```
broken: err == nil → false  (BUG!)
fixed:  err == nil → true   (correct)
```

**Checkpoint:** Use `fmt.Printf("%T %v\n", err, err)` on the broken version to see the type is `*ValidationError` even though the value is nil.

---

### Lab 4: Type Assertion

**What you'll practise:** Safe type assertion with the comma-ok pattern to avoid panics.

**Task:**
Write a function `fileInfo(r io.Reader) string` that tries to assert `r` to `*os.File` and returns the filename if successful, or `"(not a file)"` otherwise.

**Steps:**
1. Import `io`, `os`, `strings`
2. Write `fileInfo` using the two-value assertion form
3. Call it with an `*os.File` (opened from a temp file) and with a `*strings.Reader`
4. Show what happens with the single-value (panicking) form when the assertion fails

```go
func fileInfo(r io.Reader) string {
    f, ok := r.(*os.File)
    if !ok {
        return "(not a file)"
    }
    return f.Name()
}

// Dangerous — only use when you are 100% certain of the type:
// f := r.(*os.File)  // panics if r is not *os.File
```

**Expected output:**
```
os.File:        /tmp/testfile123
strings.Reader: (not a file)
```

**Checkpoint:** Change the comma-ok to single-value form and pass a `*strings.Reader` — confirm the panic message and stack trace.

---

### Lab 5: Type Switch

**What you'll practise:** Using a type switch to handle multiple concrete types from an `any` parameter.

**Task:**
Write `prettyPrint(v any) string` that formats different types differently. Handle `int`, `string`, `[]string`, `map[string]any`, `fmt.Stringer`, and a default case.

**Steps:**
1. Write the type switch with all required cases
2. For `int`: format as `"int(42)"`
3. For `string`: format as `"string(\"hello\")"`
4. For `[]string`: format each element with its index
5. For `map[string]any`: print each key-value pair sorted by key
6. For `fmt.Stringer`: call `.String()`
7. Default: use `fmt.Sprintf("%T: %v", v, v)`

```go
func prettyPrint(v any) string {
    switch x := v.(type) {
    case nil:
        return "<nil>"
    case int:
        return fmt.Sprintf("int(%d)", x)
    case string:
        return fmt.Sprintf("string(%q)", x)
    case []string:
        // your code here
    case map[string]any:
        // your code here
    case fmt.Stringer:
        return "Stringer: " + x.String()
    default:
        return fmt.Sprintf("%T: %v", x, x)
    }
}
```

**Expected output:**
```
<nil>
int(42)
string("hello")
[0]="foo" [1]="bar"
key1=val1 key2=val2
Stringer: 2009-11-10 23:00:00 +0000 UTC
```

**Checkpoint:** Add `bool` as a named case. Verify that the `fmt.Stringer` case matches any type implementing that interface by passing a `time.Time` value.

---

### Lab 6: Small Interface Design

**What you'll practise:** Narrowing a function's dependency from a large concrete struct to a minimal interface, enabling easier testing.

**Task:**
You have a `ReportGenerator` struct with many methods. A function `sendReport` only uses `Title()` and `CSV() []byte`. Refactor to accept a small interface instead.

**Steps:**
1. Define a `Reportable` interface with `Title() string` and `CSV() []byte`
2. Rewrite `sendReport(r Reportable)` to use the interface
3. Create a `fakeReport` struct in the same file that implements only those two methods
4. Show that the real `ReportGenerator` satisfies `Reportable` without modification

```go
// Before: locked to one type
// func sendReport(rg *ReportGenerator) { ... }

// After: accepts anything with Title + CSV
type Reportable interface {
    Title() string
    CSV() []byte
}
func sendReport(r Reportable) {
    fmt.Printf("Sending: %s (%d bytes)\n", r.Title(), len(r.CSV()))
}

// In tests — no ReportGenerator needed
type fakeReport struct{}
func (f fakeReport) Title() string { return "Test Report" }
func (f fakeReport) CSV() []byte   { return []byte("a,b,c") }
```

**Expected output:**
```
Sending: Q3 Sales Report (1024 bytes)
Sending: Test Report (5 bytes)
```

**Checkpoint:** Verify the function compiles with both types. Add a third method to `Reportable` that `fakeReport` does not implement — confirm the compile error.

---

### Lab 7: Interface Anti-Patterns

**What you'll practise:** Identifying and fixing the three most common interface design mistakes.

**Task:**
Examine three anti-patterns, explain why each is wrong, and implement the correct alternative.

**Steps:**
1. **Too large:** Show an 8-method `UserRepository` interface. Split into `UserReader` (2 methods) and `UserWriter` (2 methods)
2. **Constructor returning interface:** Show `NewStore() StoreInterface`. Change it to return `*SQLStore`
3. **any as shortcut:** Show `func Process(v any) any`. Change it to accept a typed interface

```go
// Anti-pattern 1: bloated interface
type UserRepository interface {
    FindAll() []User; FindByID(int) (User, error)
    FindByEmail(string) (User, error); Save(User) error
    Delete(int) error; Count() int
    Paginate(int, int) []User; Search(string) []User
}

// Better: split by caller need
type UserReader interface {
    FindByID(int) (User, error)
    FindByEmail(string) (User, error)
}
type UserWriter interface {
    Save(User) error
    Delete(int) error
}
```

**Expected output:**
```
UserReader has 2 methods — easy to implement a fake
UserWriter has 2 methods — test writes independently
Constructor returns *SQLStore — caller can access all methods
```

**Checkpoint:** Write a test double that implements only `UserReader`. Confirm it cannot be passed where `UserRepository` is expected — the compiler error message is the lesson.

---

### Lab 8: Building a Notifier System

**What you'll practise:** Designing a pluggable notification system using interfaces as the central abstraction.

**Task:**
Define a `Notifier` interface. Implement `LogNotifier`, `MultiNotifier`, and `RetryNotifier`. Wire them together and test using only the `LogNotifier` fake — no real network calls.

**Steps:**
1. Define `Message` struct and `Notifier` interface
2. Implement `LogNotifier` that records messages in a slice and prints to stdout
3. Implement `MultiNotifier` that fans out to N notifiers, collecting all errors
4. Implement `RetryNotifier` wrapping any `Notifier`, retrying up to 3 times on error
5. Wire: `RetryNotifier{Wrapped: MultiNotifier{log1, log2}}` — send 3 messages

```go
type Message struct {
    Level   string
    Subject string
    Body    string
}

type Notifier interface {
    Notify(msg Message) error
}

type LogNotifier struct {
    Sent []Message
}

func (l *LogNotifier) Notify(msg Message) error {
    l.Sent = append(l.Sent, msg)
    fmt.Printf("[LOG] %s: %s\n", msg.Level, msg.Subject)
    return nil
}
```

**Expected output:**
```
[LOG] info: Welcome
[LOG] info: Welcome
[LOG] warn: Disk full
[LOG] warn: Disk full
Sent 2 messages to 2 notifiers
```

**Checkpoint:** Make the first call to `LogNotifier.Notify` return `errors.New("transient failure")`. Verify `RetryNotifier` retries and succeeds on attempt 2, printing the retry count.

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

## Official Documentation

- [`io`](https://pkg.go.dev/io) — `Reader`, `Writer`, `Closer`, `Seeker`, `ByteReader`, `RuneReader`, `WriterTo`, `ReaderFrom`, `ReadWriter`, `ReadCloser`, `WriteCloser`, `ReadWriteCloser`
- [`fmt`](https://pkg.go.dev/fmt) — `Stringer`, `GoStringer`, `Formatter` interfaces; `%s`, `%v`, `%#v` verbs
- [`sort`](https://pkg.go.dev/sort) — `Interface` (Len, Less, Swap) for custom type sorting
- [`net/http`](https://pkg.go.dev/net/http) — `Handler`, `HandlerFunc` adapter pattern
- [`bytes`](https://pkg.go.dev/bytes) — `Buffer` used in type assertion examples
- [`errors`](https://pkg.go.dev/errors) — `Join` for collecting multiple notifier errors
- [`encoding`](https://pkg.go.dev/encoding) — `TextMarshaler`, `TextUnmarshaler` interfaces
- [`encoding/json`](https://pkg.go.dev/encoding/json) — `Marshaler`, `Unmarshaler` interfaces
- [`log/slog`](https://pkg.go.dev/log/slog) — `Logger` used in `LogNotifier` implementation
- [Go Blog: Go Interfaces (Rob Pike)](https://go.dev/blog/laws-of-reflection)
- [Effective Go — Interfaces and other types](https://go.dev/doc/effective_go#interfaces_and_types)
- [Language Spec — Interface types](https://go.dev/ref/spec#Interface_types)
