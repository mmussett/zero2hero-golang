# Day 19: File I/O and the io Package

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

## Labs

### Lab 1: os.ReadFile / os.WriteFile — Simple File I/O

**What you'll practise:** The simplest Go file-read and file-write pattern.

**Task:**
Read a text file into memory, count its lines, modify the content (e.g. uppercase every line), and write the result to a new file. Handle the "file not found" case with a clear error message.

**Steps:**
1. Create a sample `input.txt` with 5–10 lines
2. `data, err := os.ReadFile("input.txt")` — handle `err` with `fmt.Errorf("reading input: %w", err)`
3. Split on `\n` with `strings.Split`, count non-empty lines
4. Write the modified content: `os.WriteFile("output.txt", []byte(modified), 0o644)`

```go
data, err := os.ReadFile("input.txt")
if err != nil {
    log.Fatalf("reading input: %v", err)
}

lines := strings.Split(string(data), "\n")
fmt.Printf("lines: %d\n", len(lines))

for i, l := range lines {
    lines[i] = strings.ToUpper(l)
}

out := strings.Join(lines, "\n")
if err := os.WriteFile("output.txt", []byte(out), 0o644); err != nil {
    log.Fatalf("writing output: %v", err)
}
fmt.Println("written output.txt")
```

**Expected output:**
```
lines: 7
written output.txt
```

**Checkpoint:** `output.txt` exists and every line is uppercased compared to `input.txt`.

---

### Lab 2: bufio.Scanner — Efficient Line-by-Line and Word-by-Word Scanning

**What you'll practise:** Using `bufio.Scanner` with different split functions to count lines and unique words.

**Task:**
Write two functions: `countLines(path string) (int, error)` and `uniqueWords(path string) (map[string]int, error)`. The first uses the default line scanner; the second uses `bufio.ScanWords`. Test both on a large file (generate one programmatically if needed).

**Steps:**
1. `scanner := bufio.NewScanner(f)` with default `ScanLines` for line counting
2. Change to `scanner.Split(bufio.ScanWords)` for word-by-word scanning
3. Lowercase each word with `strings.ToLower` before counting
4. Print the top 5 words by frequency

```go
func uniqueWords(path string) (map[string]int, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()

    freq := make(map[string]int)
    scanner := bufio.NewScanner(f)
    scanner.Split(bufio.ScanWords)
    for scanner.Scan() {
        freq[strings.ToLower(scanner.Text())]++
    }
    return freq, scanner.Err()
}
```

**Expected output:**
```
Lines: 1000
Top words:
  the: 142
  a: 97
  and: 84
  to: 71
  of: 68
```

**Checkpoint:** `uniqueWords` returns a map with no error, and `scanner.Err()` is `nil` after the loop.

---

### Lab 3: bufio.Writer — Buffered vs Unbuffered Write Performance

**What you'll practise:** Measuring the performance difference between buffered and direct file writes.

**Task:**
Write 10 000 lines to a file using a raw `os.File` (unbuffered), then repeat using `bufio.NewWriter`. Time both approaches and compare. Ensure the buffered writer is always flushed.

**Steps:**
1. Open a file with `os.Create`, write 10 000 lines directly — time it with `time.Now()` and `time.Since`
2. Open a second file, wrap with `bufio.NewWriter(f)`, write the same 10 000 lines, call `w.Flush()`
3. Compare the durations; the buffered version should be 5–50x faster
4. Add `defer w.Flush()` — verify it's also called on early returns

```go
// Buffered write
f, _ := os.Create("buffered.txt")
w := bufio.NewWriter(f)
defer w.Flush()
defer f.Close()

start := time.Now()
for i := 0; i < 10_000; i++ {
    fmt.Fprintf(w, "line %d: the quick brown fox\n", i)
}
w.Flush()
fmt.Println("buffered:", time.Since(start))
```

**Expected output:**
```
unbuffered: 45.2ms
buffered:   1.3ms
```

**Checkpoint:** The buffered write is measurably faster, and the output file contains exactly 10 000 lines.

---

### Lab 4: io.Copy — File Copy and Custom Copy Loop

**What you'll practise:** Using `io.Copy` for efficient streaming, then implementing a manual copy loop for comparison.

**Task:**
Copy a file two ways: first using `io.Copy(dst, src)`, then using a manual `buf := make([]byte, 32*1024)` loop calling `src.Read` and `dst.Write`. Verify both produce identical output files.

**Steps:**
1. `io.Copy` copy: open src, create dst, call `n, err := io.Copy(dst, src)`, print bytes copied
2. Manual loop: use `io.ReadFull` or `src.Read` into a 32 KB buffer, write each chunk to dst
3. Compare the two output files with `bytes.Equal(file1content, file2content)`
4. Note that `io.Copy` uses the same 32 KB buffer internally

```go
// io.Copy version
src, _ := os.Open("input.txt")
dst, _ := os.Create("copy1.txt")
n, err := io.Copy(dst, src)
fmt.Printf("copied %d bytes, err: %v\n", n, err)
src.Close(); dst.Close()

// Manual version
src, _ = os.Open("input.txt")
dst, _ = os.Create("copy2.txt")
buf := make([]byte, 32*1024)
for {
    nr, er := src.Read(buf)
    if nr > 0 { dst.Write(buf[:nr]) }
    if er == io.EOF { break }
    if er != nil { log.Fatal(er) }
}
src.Close(); dst.Close()
```

**Expected output:**
```
copied 4096 bytes, err: <nil>
copy1.txt and copy2.txt are identical: true
```

**Checkpoint:** Both output files exist and are byte-for-byte identical.

---

### Lab 5: os.File Seek — Positional Reads

**What you'll practise:** Using `Seek` and `ReadAt` for random-access reads on a file.

**Task:**
Open a binary or text file. Seek to byte offset 100, read 50 bytes, print them. Then seek back to the start and confirm the position. Use `ReadAt` for a positional read without changing the file cursor.

**Steps:**
1. `f, err := os.Open("somefile.txt")`
2. `f.Seek(100, io.SeekStart)` — seek to offset 100 from the beginning
3. `io.ReadFull(f, buf)` — read 50 bytes
4. `f.Seek(0, io.SeekStart)` — seek back to start, verify with `f.Seek(0, io.SeekCurrent)` returns 0
5. `f.ReadAt(buf, 100)` — positional read without changing cursor, verify cursor didn't move

```go
f, _ := os.Open("input.txt")
defer f.Close()

f.Seek(100, io.SeekStart)
buf := make([]byte, 50)
n, _ := io.ReadFull(f, buf)
fmt.Printf("bytes 100-149: %q\n", buf[:n])

pos, _ := f.Seek(0, io.SeekStart)
fmt.Println("cursor reset to:", pos)

n2, _ := f.ReadAt(buf, 100)
fmt.Printf("ReadAt bytes 100-149: %q\n", buf[:n2])
cur, _ := f.Seek(0, io.SeekCurrent)
fmt.Println("cursor after ReadAt:", cur) // still 0
```

**Expected output:**
```
bytes 100-149: "...50 chars from offset 100..."
cursor reset to: 0
ReadAt bytes 100-149: "...same 50 chars..."
cursor after ReadAt: 0
```

**Checkpoint:** Both reads return the same bytes. `ReadAt` does not move the file cursor.

---

### Lab 6: filepath.WalkDir — Collect .go Files

**What you'll practise:** Walking a directory tree and filtering entries by extension.

**Task:**
Walk the `zero2hero-golang` root directory (or any directory). Collect all `.go` files, print their relative paths and sizes, and print a summary total. Skip `vendor` and hidden directories.

**Steps:**
1. `filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error { ... })`
2. Skip `d.IsDir()` entries and entries where `filepath.Ext(path) != ".go"`
3. Call `d.Info()` to get `fs.FileInfo`, read `fi.Size()`
4. Accumulate total size, print a summary line at the end

```go
var totalBytes int64
var count int

err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
    if err != nil { return err }
    if d.IsDir() && (d.Name() == "vendor" || strings.HasPrefix(d.Name(), ".")) {
        return filepath.SkipDir
    }
    if filepath.Ext(path) != ".go" { return nil }
    fi, _ := d.Info()
    totalBytes += fi.Size()
    count++
    fmt.Printf("%-50s %7d bytes\n", path, fi.Size())
    return nil
})

fmt.Printf("\n%d .go files, %d bytes total\n", count, totalBytes)
```

**Expected output:**
```
day-01/main.go                                          42 bytes
day-02/main.go                                         128 bytes
...
34 .go files, 48291 bytes total
```

**Checkpoint:** The count and total match what `find . -name "*.go"` would return.

---

### Lab 7: Temporary Files — Safe Temp File Pattern

**What you'll practise:** Creating, writing, reading, and cleaning up temporary files safely.

**Task:**
Create a temporary file using `os.CreateTemp`, write structured data (e.g. JSON) to it, close it, re-open it to read back, then delete it with `defer os.Remove`. Verify the file is gone after the function returns.

**Steps:**
1. `f, err := os.CreateTemp("", "day18-*.json")` — create in the default temp dir
2. Write JSON using `json.NewEncoder(f).Encode(data)`
3. `f.Close()` — close before re-opening for reading
4. Re-open with `os.Open(f.Name())`, decode with `json.NewDecoder`
5. `defer os.Remove(f.Name())` at the top ensures cleanup even on error

```go
type Record struct{ Name string; Value int }

f, err := os.CreateTemp("", "day18-*.json")
if err != nil { log.Fatal(err) }
defer os.Remove(f.Name())
defer f.Close()

enc := json.NewEncoder(f)
enc.Encode(Record{"answer", 42})
f.Close()

f2, _ := os.Open(f.Name())
defer f2.Close()
var r Record
json.NewDecoder(f2).Decode(&r)
fmt.Printf("read back: %+v\n", r)
fmt.Printf("temp path: %s\n", f.Name())
```

**Expected output:**
```
read back: {Name:answer Value:42}
temp path: /tmp/day18-1234567890.json
```

**Checkpoint:** After the function returns, `os.Stat(f.Name())` returns an error (`os.IsNotExist`).

---

### Final Lab (Project): .ini Config File Reader/Writer

**What you'll practise:** Combining `bufio.Scanner`, `strings.Cut`, `os.ReadFile`, `io.Writer`, and structured error handling to build a complete file parser.

**Task:**
Write a program that reads a `.ini`-style config file and writes it back in the same format.

**Steps:**
1. Parse the file into `map[string]map[string]string` (section → key → value)
2. Handle blank lines (skip), comment lines (`#` prefix, skip), section headers (`[section]`), and key-value pairs (`key = value`)
3. Return a custom `ParseError` with line number for malformed lines
4. Implement `Write(w io.Writer, cfg map[string]map[string]string) error` that outputs the canonical format
5. Round-trip: parse → write to `bytes.Buffer` → parse again → assert equal

```go
type ParseError struct {
    Line int
    Text string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("line %d: malformed: %q", e.Line, e.Text)
}

func Parse(r io.Reader) (map[string]map[string]string, error) {
    cfg := make(map[string]map[string]string)
    var section string
    scanner := bufio.NewScanner(r)
    for i := 1; scanner.Scan(); i++ {
        line := strings.TrimSpace(scanner.Text())
        switch {
        case line == "" || strings.HasPrefix(line, "#"):
            // skip
        case strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]"):
            section = line[1 : len(line)-1]
            cfg[section] = make(map[string]string)
        default:
            k, v, ok := strings.Cut(line, "=")
            if !ok { return nil, &ParseError{i, line} }
            cfg[section][strings.TrimSpace(k)] = strings.TrimSpace(v)
        }
    }
    return cfg, scanner.Err()
}
```

**Expected output:**
```
Parsed sections: [database server]
database.host = localhost
database.port = 5432
server.port = 8080
Round-trip equal: true
```

**Checkpoint:** The round-tripped config equals the original parsed config. A file with a malformed line returns a `*ParseError` with the correct line number.

---

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
