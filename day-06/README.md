# Day 06: Maps and Method Sets

## Maps

A `map[K]V` is Go's built-in hash table. The zero value is `nil` — always initialise before writing.

```go
m := make(map[string]int)
m["apple"]++
m["banana"] = 5

// Comma-ok idiom — distinguish missing from zero
count, ok := m["apple"]
if !ok {
    fmt.Println("key not found")
}

delete(m, "apple")

// Iteration — order is randomised
for k, v := range m {
    fmt.Printf("%s: %d\n", k, v)
}
```

Pre-size hint when the approximate number of entries is known:

```go
freq := make(map[string]int, len(words))
```

## Embedded Structs (Composition)

Go favours composition over inheritance. Embed a type to promote its fields and methods:

```go
type Animal struct {
    Name string
}

func (a Animal) Speak() string { return a.Name + " makes a sound" }

type Dog struct {
    Animal        // embedded — promotes Name and Speak
    Breed string
}

d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Labrador"}
fmt.Println(d.Name)    // promoted from Animal
fmt.Println(d.Speak()) // promoted method
```

The embedded type's methods are promoted — `Dog` satisfies any interface `Animal` satisfies, unless `Dog` overrides the method.

## Method Sets and Interfaces

A type's **method set** determines which interfaces it satisfies:
- A value `T` has all methods with value receivers
- A pointer `*T` has all methods with value **and** pointer receivers

This matters when storing values in interfaces:

```go
type Stringer interface { String() string }

type Foo struct{}
func (f *Foo) String() string { return "foo" }

var s Stringer = &Foo{}  // ok — *Foo has the method
var t Stringer = Foo{}   // compile error — Foo does not have String
```

---

## Labs

### Lab 1: Map Basics — CRUD and the Comma-Ok Idiom

**What you'll practise:** Creating a map, reading with the two-value form, updating, deleting, and iterating.

**Task:**
Build a map of country capitals, demonstrate every map operation, and show what happens when you read a missing key.

**Steps:**
1. Create `capitals := map[string]string{"France": "Paris", "Japan": "Tokyo"}`
2. Add a new entry, update an existing one, delete one
3. Read a present key (comma-ok: `v, ok := ...`)
4. Read a missing key — show it returns the zero value, not a panic
5. Iterate with `range` — note that order is random

```go
package main

import "fmt"

func main() {
    capitals := map[string]string{
        "France": "Paris",
        "Japan":  "Tokyo",
        "Brazil": "Brasília",
    }

    // Add and update
    capitals["Germany"] = "Berlin"
    capitals["France"] = "Paris (updated)"

    // Delete
    delete(capitals, "Brazil")

    // Comma-ok idiom
    if v, ok := capitals["Japan"]; ok {
        fmt.Println("Japan's capital:", v)
    }

    // Missing key — returns zero value, not a panic
    v := capitals["Australia"]
    fmt.Printf("Australia (missing): %q (empty string)\n", v)

    _, ok := capitals["Australia"]
    fmt.Printf("Australia exists: %v\n", ok)

    // Iteration — order is non-deterministic
    fmt.Println("\nAll capitals:")
    for country, city := range capitals {
        fmt.Printf("  %s: %s\n", country, city)
    }
}
```

**Expected output:**
```
Japan's capital: Tokyo
Australia (missing): "" (empty string)
Australia exists: false

All capitals:
  France: Paris (updated)
  Japan: Tokyo
  Germany: Berlin
(order may vary)
```

**Checkpoint:** Reading a missing key never panics — it returns the zero value. Run the program multiple times and observe that the iteration order changes each time.

---

### Lab 2: Map Patterns — Frequency, Grouping, and Sets

**What you'll practise:** Three essential map patterns: word frequency counter, grouping items by a key, and using `map[K]struct{}` as a set.

**Task:**
Implement all three patterns against a sample word list.

**Steps:**
1. **Frequency:** count occurrences of each word in a sentence
2. **Grouping:** group words by their first letter (`map[string][]string`)
3. **Set:** deduplicate a list using `map[string]struct{}`

```go
package main

import (
    "fmt"
    "strings"
)

func main() {
    sentence := "the cat sat on the mat the cat wore a hat"
    words := strings.Fields(sentence)

    // 1. Frequency counter
    freq := make(map[string]int)
    for _, w := range words {
        freq[w]++
    }
    fmt.Println("Frequency:")
    for w, n := range freq {
        fmt.Printf("  %-6s %d\n", w, n)
    }

    // 2. Group words by first letter
    groups := make(map[string][]string)
    for _, w := range words {
        key := string(w[0])
        groups[key] = append(groups[key], w)
    }
    fmt.Println("\nGroups by first letter:")
    for k, ws := range groups {
        fmt.Printf("  %s: %v\n", k, ws)
    }

    // 3. Set — deduplicate
    seen := make(map[string]struct{})
    var unique []string
    for _, w := range words {
        if _, ok := seen[w]; !ok {
            seen[w] = struct{}{}
            unique = append(unique, w)
        }
    }
    fmt.Printf("\nUnique words (%d): %v\n", len(unique), unique)
}
```

**Expected output:**
```
Frequency:
  the    3
  cat    2
  sat    1
  on     1
  mat    1
  wore   1
  a      1
  hat    1
(order varies)

Groups by first letter:
  t: [the cat sat... ]
  ...

Unique words (8): [the cat sat on mat wore a hat]
```

**Checkpoint:** The frequency map correctly counts "the" as 3 and "cat" as 2. The set produces exactly 8 unique words. `struct{}` uses zero bytes — it is the idiomatic Go set value.

---

### Lab 3: Methods — Receivers, Expressions, and Values

**What you'll practise:** Defining value and pointer receiver methods on a struct, and understanding the difference between a method expression and a method value.

**Task:**
Build a `Rectangle` struct with `Area()` and `Scale()` methods, then demonstrate method expressions and method values.

**Steps:**
1. Define `type Rectangle struct { Width, Height float64 }`
2. Write `(r Rectangle) Area() float64` — value receiver, read-only
3. Write `(r Rectangle) Perimeter() float64` — value receiver
4. Write `(r *Rectangle) Scale(factor float64)` — pointer receiver, mutates
5. Demonstrate a **method value**: `areaFn := r.Area` (bound to `r`)
6. Demonstrate a **method expression**: `expr := Rectangle.Area` (unbound, takes receiver as first arg)

```go
package main

import "fmt"

type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

func main() {
    r := Rectangle{Width: 4, Height: 3}

    fmt.Printf("Area:      %.1f\n", r.Area())
    fmt.Printf("Perimeter: %.1f\n", r.Perimeter())

    r.Scale(2)
    fmt.Printf("After Scale(2): %+v\n", r)

    // Method value — bound to r; calling areaFn() is like calling r.Area()
    areaFn := r.Area
    fmt.Printf("Method value Area(): %.1f\n", areaFn())

    // Method expression — unbound; first arg is the receiver
    expr := Rectangle.Area
    r2 := Rectangle{Width: 5, Height: 5}
    fmt.Printf("Method expression Area(r2): %.1f\n", expr(r2))
}
```

**Expected output:**
```
Area:      12.0
Perimeter: 14.0
After Scale(2): {Width:8 Height:6}
Method value Area(): 48.0
Method expression Area(r2): 25.0
```

**Checkpoint:** `r.Scale(2)` changes `r` because `Scale` has a pointer receiver. `areaFn()` captures the current state of `r` at the time the method value was created.

---

### Lab 4: fmt.Stringer — Make Your Types Print Themselves

**What you'll practise:** Implementing the `fmt.Stringer` interface so `fmt.Println` automatically calls your `String()` method.

**Task:**
Implement `String() string` on `Rectangle` and a `Color` type, then show how `fmt` automatically uses it in `Println`, `Printf %v`, and `Sprintf`.

**Steps:**
1. Add `(r Rectangle) String() string` to the Rectangle from Lab 3
2. Define `type Color struct { R, G, B uint8 }` with a hex `String() string`
3. Call `fmt.Println(rect)` and `fmt.Println(color)` — no extra args needed
4. Use `fmt.Sprintf("%v", rect)` and confirm it also calls `String()`

```go
package main

import "fmt"

type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 { return r.Width * r.Height }

// String implements fmt.Stringer — called by fmt.Println, %v, %s
func (r Rectangle) String() string {
    return fmt.Sprintf("Rectangle(%.1f x %.1f, area=%.1f)", r.Width, r.Height, r.Area())
}

type Color struct{ R, G, B uint8 }

func (c Color) String() string {
    return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func main() {
    r := Rectangle{Width: 10, Height: 5}
    c := Color{R: 255, G: 128, B: 0}

    // fmt.Println calls String() automatically
    fmt.Println(r)
    fmt.Println(c)

    // Also works with %v (but not %d — that would use the struct format)
    fmt.Printf("rect: %v\n", r)
    fmt.Printf("color: %v\n", c)

    // Sprintf works too
    s := fmt.Sprintf("my rect is %v", r)
    fmt.Println(s)
}
```

**Expected output:**
```
Rectangle(10.0 x 5.0, area=50.0)
#FF8000
rect: Rectangle(10.0 x 5.0, area=50.0)
color: #FF8000
my rect is Rectangle(10.0 x 5.0, area=50.0)
```

**Checkpoint:** Remove `String()` from `Rectangle` — `fmt.Println(r)` will then print `{10 5}`. Re-add it to see the difference. Try `fmt.Printf("%#v", r)` with and without `String()`.

---

### Lab 5: Map of Functions — Command Dispatcher

**What you'll practise:** Using `map[string]func([]string)` to build a simple command dispatcher that routes input to registered handlers.

**Task:**
Build a `dispatcher` that registers "add", "mul", "upper", and "help" commands, then dispatches to the correct handler based on a command name string.

**Steps:**
1. Define `type Command func(args []string)`
2. Create `dispatch := map[string]Command{...}` and register four commands
3. Write a `run(dispatch map[string]Command, name string, args []string)` function
4. Call `run` with each command name and confirm dispatch works
5. Call `run` with an unknown command — print a helpful error

```go
package main

import (
    "fmt"
    "strconv"
    "strings"
)

type Command func(args []string)

func run(dispatch map[string]Command, name string, args []string) {
    cmd, ok := dispatch[name]
    if !ok {
        fmt.Printf("unknown command %q. Try 'help'\n", name)
        return
    }
    cmd(args)
}

func main() {
    dispatch := map[string]Command{
        "add": func(args []string) {
            sum := 0
            for _, a := range args {
                n, _ := strconv.Atoi(a)
                sum += n
            }
            fmt.Printf("add(%v) = %d\n", args, sum)
        },
        "mul": func(args []string) {
            product := 1
            for _, a := range args {
                n, _ := strconv.Atoi(a)
                product *= n
            }
            fmt.Printf("mul(%v) = %d\n", args, product)
        },
        "upper": func(args []string) {
            fmt.Println(strings.ToUpper(strings.Join(args, " ")))
        },
        "help": func(args []string) {
            fmt.Println("commands: add, mul, upper, help")
        },
    }

    run(dispatch, "add", []string{"1", "2", "3", "4"})
    run(dispatch, "mul", []string{"2", "3", "5"})
    run(dispatch, "upper", []string{"hello", "gopher"})
    run(dispatch, "help", nil)
    run(dispatch, "unknown", nil)
}
```

**Expected output:**
```
add([1 2 3 4]) = 10
mul([2 3 5]) = 30
HELLO GOPHER
commands: add, mul, upper, help
unknown command "unknown". Try 'help'
```

**Checkpoint:** Add a "sub" command that subtracts all args from the first. The dispatcher doesn't need to change — just add a new key. This is the open/closed principle in Go.

---

### Lab 6 (Final): Word Frequency Counter

**What you'll practise:** Combining maps, slices, sorting, and string processing into a complete top-N word frequency reporter.

**Task:**
Write `day-06/main.go` that parses multi-line text into words, builds a frequency table, sorts by count descending, and prints a formatted top-10 table.

**Steps:**
1. Lowercase and strip punctuation from each word using `strings.Map`
2. Build a `map[string]int` frequency table
3. Collect map entries into `[]Entry{Word, Count}` and sort by count descending
4. Print the top 10 in a formatted table with rank, word, count, and a bar chart

```go
package main

import (
    "fmt"
    "sort"
    "strings"
)

const text = `To be or not to be that is the question
Whether tis nobler in the mind to suffer
The slings and arrows of outrageous fortune
Or to take arms against a sea of troubles`

type Entry struct {
    Word  string
    Count int
}

func main() {
    // Normalise: lowercase, strip non-letter characters
    clean := strings.Map(func(r rune) rune {
        if r >= 'a' && r <= 'z' || r == ' ' || r == '\n' {
            return r
        }
        if r >= 'A' && r <= 'Z' {
            return r + 32 // tolower
        }
        return -1 // drop
    }, text)

    // Build frequency map
    freq := make(map[string]int)
    for _, w := range strings.Fields(clean) {
        if w != "" {
            freq[w]++
        }
    }

    // Sort entries by count descending, then alphabetically
    entries := make([]Entry, 0, len(freq))
    for w, n := range freq {
        entries = append(entries, Entry{w, n})
    }
    sort.Slice(entries, func(i, j int) bool {
        if entries[i].Count != entries[j].Count {
            return entries[i].Count > entries[j].Count
        }
        return entries[i].Word < entries[j].Word
    })

    // Print top 10
    top := 10
    if len(entries) < top {
        top = len(entries)
    }
    fmt.Printf("%-4s %-12s %5s  %s\n", "Rank", "Word", "Count", "Bar")
    fmt.Println(strings.Repeat("-", 40))
    for i, e := range entries[:top] {
        bar := strings.Repeat("*", e.Count)
        fmt.Printf("%-4d %-12s %5d  %s\n", i+1, e.Word, e.Count, bar)
    }
}
```

**Expected output:**
```
Rank Word         Count  Bar
----------------------------------------
1    to            5  *****
2    the           3  ***
3    or            2  **
4    be            2  **
5    of            2  **
...
```

**Checkpoint:** `go run .` prints the table. The word "to" ranks first (appears 5 times). Sorting is stable: ties are broken alphabetically. Run `go vet .` — no warnings.

---

## Day Project: Word Frequency Counter

Extend the Day 05 string statistics tool:
1. Parse a multi-line text into words (lowercase, strip punctuation with [`strings.Map`](https://pkg.go.dev/strings#Map))
2. Build a `map[string]int` frequency table
3. Find the top-N most frequent words (collect map entries, sort by count descending)
4. Print a formatted table

```go
type Entry struct {
    Word  string
    Count int
}
```

**Extension ideas:** read from a file path passed as `os.Args[1]`; add a `--top N` flag.

## Official Documentation

- [`fmt`](https://pkg.go.dev/fmt) — formatted output (Printf, Println)
- [`strings`](https://pkg.go.dev/strings) — `Map`, `ToLower`, `Fields` for text processing
- [`sort`](https://pkg.go.dev/sort) — sorting slices for top-N results
- [`os`](https://pkg.go.dev/os) — `os.Args` for file path arguments
- [Language Spec: Map types](https://go.dev/ref/spec#Map_types) — map declaration and usage
- [Language Spec: Method sets](https://go.dev/ref/spec#Method_sets) — value vs pointer method sets
- [Language Spec: Struct types](https://go.dev/ref/spec#Struct_types) — embedding
- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding) — composition via embedding
- [Go Tour: Maps](https://go.dev/tour/moretypes/19) — maps tour
