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
