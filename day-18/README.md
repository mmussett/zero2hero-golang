# Day 18: File I/O and the io Package

## Core Concept: Everything Is a Reader or Writer

Go's I/O model is built on two tiny interfaces — `io.Reader` and `io.Writer` — that compose into arbitrarily powerful pipelines.

## Reading Files

```go
// Simple: read entire file into memory
data, err := os.ReadFile("config.txt")

// Streaming: line by line
f, err := os.Open("large.txt")
defer f.Close()

scanner := bufio.NewScanner(f)
for scanner.Scan() {
    line := scanner.Text()
}
if err := scanner.Err(); err != nil { /* ... */ }
```

## Writing Files

```go
// Simple
os.WriteFile("out.txt", []byte("hello"), 0o644)

// Streaming
f, err := os.Create("out.txt")
defer f.Close()

w := bufio.NewWriter(f)
fmt.Fprintln(w, "line 1")
w.Flush() // must flush buffered writer
```

## io.Copy

Copies from a `Reader` to a `Writer` efficiently (64 KB chunks):

```go
src, _ := os.Open("input.txt")
dst, _ := os.Create("output.txt")
defer src.Close(); defer dst.Close()
n, err := io.Copy(dst, src)
```

## Directory Walking

```go
err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
    if err != nil { return err }
    if !d.IsDir() {
        fmt.Println(path, d.Name())
    }
    return nil
})
```

## Useful io Functions

| Function | Purpose |
|----------|---------|
| `io.ReadAll(r)` | Read entire Reader into `[]byte` |
| `io.Copy(dst, src)` | Stream from Reader to Writer |
| `io.MultiReader(rs...)` | Concatenate multiple Readers |
| `io.TeeReader(r, w)` | Read from r, simultaneously write to w |
| `io.LimitReader(r, n)` | Read at most n bytes |
| `io.Pipe()` | Synchronous in-memory pipe |
| `io.Discard` | /dev/null Writer |

## Temporary Files

```go
f, err := os.CreateTemp("", "prefix-*.txt")
defer os.Remove(f.Name())
defer f.Close()
```

## Day Project: Config File Reader

Write a program that reads a simple `.ini`-style config file:

```ini
[database]
host = localhost
port = 5432
name = mydb

[server]
port = 8080
debug = true
```

Parse it into a `map[string]map[string]string` (section → key → value). Handle:
- Blank lines and comments (`#` prefix)
- Missing file (return a useful error)
- Malformed lines (wrap and return a `ParseError`)

Write the parsed config back to a writer in the same format.

**Extension ideas:** support `${ENV_VAR}` substitution in values using `os.Getenv`; watch the file with a goroutine and reload on change.

## Official Documentation

- [`os`](https://pkg.go.dev/os) — `ReadFile`, `WriteFile`, `Create`, `Open`, `CreateTemp`, `Remove`
- [`io`](https://pkg.go.dev/io) — `Reader`, `Writer`, `Copy`, `ReadAll`, `MultiReader`, `TeeReader`, `LimitReader`, `Pipe`, `Discard`
- [`bufio`](https://pkg.go.dev/bufio) — `NewScanner`, `NewWriter`, `Scanner.Scan`, `Writer.Flush`
- [`path/filepath`](https://pkg.go.dev/path/filepath) — `WalkDir`
- [`io/fs`](https://pkg.go.dev/io/fs) — `DirEntry`, `FS` interface used by `filepath.WalkDir`
- [Language Spec — Interfaces](https://go.dev/ref/spec#Interface_types) — `io.Reader` / `io.Writer` interface mechanics
- [Effective Go — I/O](https://go.dev/doc/effective_go#interfaces_and_types)
