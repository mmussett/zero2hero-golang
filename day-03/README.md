# Day 03: Functions, Structs, and Error Handling

## Core Concepts

### Functions

Go functions can return **multiple values** — the primary mechanism for returning both a result and an error.

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}

result, err := divide(10, 2)
if err != nil {
    log.Fatal(err)
}
```

Named return values create pre-declared variables and enable bare `return`:

```go
func minMax(nums []int) (min, max int) {
    min, max = nums[0], nums[0]
    for _, n := range nums[1:] {
        if n < min { min = n }
        if n > max { max = n }
    }
    return // returns min and max
}
```

Use named returns for documentation, not as a shortcut — bare returns in long functions obscure what is being returned.

### Structs

```go
type Person struct {
    Name  string
    Email string
    Age   int
}

// Struct literal
p := Person{Name: "Alice", Email: "alice@example.com", Age: 30}

// Field access
fmt.Println(p.Name)
p.Age++
```

Fields not listed in a literal are set to their zero values. Always use field names in literals — positional initialisation breaks when fields are added.

### Methods

A method is a function with a receiver:

```go
func (p Person) String() string {
    return fmt.Sprintf("%s <%s>", p.Name, p.Email)
}

func (p *Person) Birthday() {
    p.Age++  // pointer receiver — modifies the original
}
```

Use a **pointer receiver** when the method mutates state or the struct is large. Use a **value receiver** when the method is read-only. Be consistent within a type — don't mix.

---

## Labs

### Lab 1: Function Fundamentals

**What you'll practise:** Multiple return values, named returns, and variadic functions.

**Task:**
Write three functions — `sum`, `minMax`, and `product` — that demonstrate Go's function features, then call them from `main`.

**Steps:**
1. Write `sum(nums ...int) int` — variadic, sums all arguments
2. Write `minMax(nums []int) (min, max int)` — named returns, no bare return
3. Write `product(a, b float64) (result float64, err error)` — returns an error if either argument is negative
4. Call all three from `main` and print results

```go
package main

import (
    "fmt"
    "errors"
)

func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

func minMax(nums []int) (min, max int) {
    min, max = nums[0], nums[0]
    for _, n := range nums[1:] {
        if n < min { min = n }
        if n > max { max = n }
    }
    return min, max // explicit is clearer than bare return
}

func product(a, b float64) (float64, error) {
    if a < 0 || b < 0 {
        return 0, errors.New("negative inputs not allowed")
    }
    return a * b, nil
}

func main() {
    fmt.Println("sum:", sum(1, 2, 3, 4, 5))

    nums := []int{3, 1, 4, 1, 5, 9, 2, 6}
    lo, hi := minMax(nums)
    fmt.Printf("min=%d max=%d\n", lo, hi)

    p, err := product(3.0, 4.0)
    fmt.Printf("product: %.1f err: %v\n", p, err)

    _, err = product(-1.0, 4.0)
    fmt.Printf("negative product err: %v\n", err)
}
```

**Expected output:**
```
sum: 15
min=1 max=9
product: 12.0 err: <nil>
negative product err: negative inputs not allowed
```

**Checkpoint:** `sum()` with no arguments returns 0. `minMax([]int{5})` returns `5, 5`. The error path for `product` returns a non-nil error.

---

### Lab 2: First-Class Functions

**What you'll practise:** Passing functions as arguments, returning functions from functions (closures), and writing a generic `apply` transformer.

**Task:**
Write an `adder` factory that returns a closure, and an `apply` function that transforms a slice using any `func(int) int`.

**Steps:**
1. Write `adder(x int) func(int) int` — returns a closure that adds `x` to its argument
2. Write `apply(nums []int, f func(int) int) []int` — returns a new slice with `f` applied to each element
3. Create `double` and `square` as named `func` variables
4. Use `apply` with `adder(10)`, `double`, and `square`

```go
package main

import "fmt"

// adder returns a closure that captures x.
func adder(x int) func(int) int {
    return func(n int) int {
        return n + x
    }
}

// apply returns a new slice with f applied to every element.
func apply(nums []int, f func(int) int) []int {
    result := make([]int, len(nums))
    for i, n := range nums {
        result[i] = f(n)
    }
    return result
}

func main() {
    add10 := adder(10)
    fmt.Println(add10(5))   // 15
    fmt.Println(add10(20))  // 30

    double := func(n int) int { return n * 2 }
    square := func(n int) int { return n * n }

    nums := []int{1, 2, 3, 4, 5}
    fmt.Println("original:", nums)
    fmt.Println("+10:     ", apply(nums, adder(10)))
    fmt.Println("doubled: ", apply(nums, double))
    fmt.Println("squared: ", apply(nums, square))
}
```

**Expected output:**
```
15
30
original: [1 2 3 4 5]
+10:      [11 12 13 14 15]
doubled:  [2 4 6 8 10]
squared:  [1 4 9 16 25]
```

**Checkpoint:** `adder(0)` returns a function that is the identity. Create two separate `adder` closures with different `x` values and confirm they don't interfere with each other.

---

### Lab 3: Struct Fundamentals

**What you'll practise:** Defining a struct, writing a constructor, adding pointer receiver methods, and implementing `fmt.Stringer`.

**Task:**
Define a `Person` struct with `Name`, `Age`, and `Email` fields. Write a constructor that validates age, a pointer receiver `Birthday()` method, and a value receiver `String()` method.

**Steps:**
1. Define `type Person struct` with three fields
2. Write `NewPerson(name string, age int, email string) (Person, error)` — error if age < 0
3. Write `(p *Person) Birthday()` — increments age (pointer receiver to mutate)
4. Write `(p Person) String() string` — returns a formatted string (value receiver — read-only)
5. Call `fmt.Println(p)` — Go calls `String()` automatically

```go
package main

import (
    "fmt"
    "errors"
)

type Person struct {
    Name  string
    Age   int
    Email string
}

func NewPerson(name string, age int, email string) (Person, error) {
    if age < 0 {
        return Person{}, errors.New("age cannot be negative")
    }
    return Person{Name: name, Age: age, Email: email}, nil
}

func (p *Person) Birthday() {
    p.Age++
}

func (p Person) String() string {
    return fmt.Sprintf("%s (age %d) <%s>", p.Name, p.Age, p.Email)
}

func main() {
    p, err := NewPerson("Alice", 30, "alice@example.com")
    if err != nil {
        fmt.Println("error:", err)
        return
    }

    fmt.Println(p) // calls p.String() automatically
    p.Birthday()
    fmt.Println(p)

    _, err = NewPerson("Bob", -1, "bob@example.com")
    fmt.Println("bad person error:", err)
}
```

**Expected output:**
```
Alice (age 30) <alice@example.com>
Alice (age 31) <alice@example.com>
bad person error: age cannot be negative
```

**Checkpoint:** Confirm that `p.Birthday()` on a value `p` (not a pointer) does NOT change the original — only `(&p).Birthday()` or a pointer variable would. Go auto-takes address when the variable is addressable.

---

### Lab 4: Struct Embedding

**What you'll practise:** Embedding one struct inside another, accessing promoted fields directly, and overriding promoted methods.

**Task:**
Embed an `Address` struct inside `Person`. Access `Address` fields directly on `Person`, then add a `String()` override on `Person` that calls the embedded type's method.

**Steps:**
1. Define `type Address struct` with `Street`, `City`, `Country`
2. Add a `String() string` method on `Address`
3. Embed `Address` in `Person` (from Lab 3 or a fresh struct)
4. Access `p.City` directly (promoted field)
5. Override `String()` on `Person` to include address info

```go
package main

import "fmt"

type Address struct {
    Street  string
    City    string
    Country string
}

func (a Address) String() string {
    return fmt.Sprintf("%s, %s, %s", a.Street, a.City, a.Country)
}

type Person struct {
    Name    string
    Age     int
    Address // embedded — promotes Street, City, Country, and String()
}

func (p Person) String() string {
    // Override Address.String() with a richer format.
    return fmt.Sprintf("%s (age %d) at %s", p.Name, p.Age, p.Address.String())
}

func main() {
    p := Person{
        Name: "Alice",
        Age:  30,
        Address: Address{
            Street:  "123 Gopher Lane",
            City:    "Gophertown",
            Country: "Goland",
        },
    }

    // Promoted fields — no need to write p.Address.City
    fmt.Println("City:", p.City)
    fmt.Println("Country:", p.Country)

    // Person.String() overrides Address.String()
    fmt.Println(p)

    // Access the embedded type's method explicitly
    fmt.Println("Address only:", p.Address.String())
}
```

**Expected output:**
```
City: Gophertown
Country: Goland
Alice (age 30) at 123 Gopher Lane, Gophertown, Goland
Address only: 123 Gopher Lane, Gophertown, Goland
```

**Checkpoint:** Remove `Person.String()` and run again — confirm `fmt.Println(p)` now uses `Address.String()` (promoted). Re-add `Person.String()` to restore the override.

---

### Lab 5: Error Handling — Wrapping and Unwrapping

**What you'll practise:** Returning `(value, error)`, using `fmt.Errorf` with `%w` to wrap errors, and using `errors.Is` / `errors.As` to inspect wrapped errors.

**Task:**
Write a two-layer call chain where each layer wraps the error from the layer below. Use `errors.Is` to check for a sentinel error and `fmt.Errorf("%w")` to preserve the chain.

**Steps:**
1. Define a sentinel error `ErrNotFound`
2. Write `findUser(id int) (string, error)` — returns `ErrNotFound` if id < 0
3. Write `loadProfile(id int) (string, error)` — calls `findUser` and wraps any error
4. In `main`, call `loadProfile` and use `errors.Is(err, ErrNotFound)` to distinguish error types

```go
package main

import (
    "errors"
    "fmt"
)

var ErrNotFound = errors.New("not found")

func findUser(id int) (string, error) {
    if id < 0 {
        return "", fmt.Errorf("findUser(%d): %w", id, ErrNotFound)
    }
    return fmt.Sprintf("User#%d", id), nil
}

func loadProfile(id int) (string, error) {
    user, err := findUser(id)
    if err != nil {
        return "", fmt.Errorf("loadProfile: %w", err)
    }
    return "Profile of " + user, nil
}

func main() {
    profile, err := loadProfile(42)
    if err != nil {
        fmt.Println("error:", err)
    } else {
        fmt.Println(profile)
    }

    _, err = loadProfile(-1)
    if err != nil {
        fmt.Println("error:", err)
        fmt.Println("is ErrNotFound?", errors.Is(err, ErrNotFound))
    }
}
```

**Expected output:**
```
Profile of User#42
error: loadProfile: findUser(-1): not found
is ErrNotFound? true
```

**Checkpoint:** `errors.Is` returns `true` even though the error has been wrapped twice. Remove the `%w` verb from one `fmt.Errorf` call and confirm `errors.Is` returns `false` for that depth.

---

### Lab 6 (Final): Contact Card

**What you'll practise:** Combining constructors, validation, pointer receivers, `fmt.Stringer`, and slice-based lookup in one cohesive program.

**Task:**
Write `day-03/main.go` with a `Contact` struct, a validating constructor, a `String()` method, a `PrintCard` function, and a `FindByName` search over a slice of contacts.

**Steps:**
1. Define `type Contact struct` with `Name`, `Email`, `Phone`
2. Write `NewContact(name, email, phone string) (Contact, error)` — error if email has no `@`
3. Write `(c Contact) String() string` — formatted card line
4. Write `func PrintCard(c Contact)` — pretty-prints a bordered card
5. Build a `[]Contact` slice and write `FindByName(contacts []Contact, name string) (Contact, bool)`
6. Call all functions from `main`, including a failed construction and a failed lookup

```go
package main

import (
    "fmt"
    "strings"
)

type Contact struct {
    Name  string
    Email string
    Phone string
}

func NewContact(name, email, phone string) (Contact, error) {
    if !strings.Contains(email, "@") {
        return Contact{}, fmt.Errorf("invalid email %q: must contain @", email)
    }
    return Contact{Name: name, Email: email, Phone: phone}, nil
}

func (c Contact) String() string {
    return fmt.Sprintf("%s | %s | %s", c.Name, c.Email, c.Phone)
}

func PrintCard(c Contact) {
    border := strings.Repeat("-", 40)
    fmt.Println(border)
    fmt.Printf("  Name:  %s\n", c.Name)
    fmt.Printf("  Email: %s\n", c.Email)
    fmt.Printf("  Phone: %s\n", c.Phone)
    fmt.Println(border)
}

func FindByName(contacts []Contact, name string) (Contact, bool) {
    for _, c := range contacts {
        if strings.EqualFold(c.Name, name) {
            return c, true
        }
    }
    return Contact{}, false
}

func main() {
    contacts := []Contact{}

    for _, args := range [][3]string{
        {"Alice", "alice@example.com", "555-1234"},
        {"Bob", "bob@example.com", "555-5678"},
        {"Carol", "carol@example.com", "555-9012"},
    } {
        c, err := NewContact(args[0], args[1], args[2])
        if err != nil {
            fmt.Println("error:", err)
            continue
        }
        contacts = append(contacts, c)
    }

    // Try an invalid email
    _, err := NewContact("Dan", "not-an-email", "555-0000")
    fmt.Println("Validation error:", err)

    // Print all cards
    for _, c := range contacts {
        PrintCard(c)
    }

    // Search
    if c, ok := FindByName(contacts, "bob"); ok {
        fmt.Println("Found:", c)
    }
    if _, ok := FindByName(contacts, "Zara"); !ok {
        fmt.Println("Zara not found")
    }
}
```

**Expected output:**
```
Validation error: invalid email "not-an-email": must contain @
----------------------------------------
  Name:  Alice
  Email: alice@example.com
  Phone: 555-1234
----------------------------------------
... (Bob and Carol cards follow)
Found: Bob | bob@example.com | 555-5678
Zara not found
```

**Checkpoint:** `go vet .` reports no issues. `FindByName` is case-insensitive (`"bob"` finds `"Bob"`). Adding a contact with an invalid email is rejected without crashing.

---

## Day Project: Contact Card

Define a `Contact` struct with at least: name, email, phone. Write:
- A `NewContact(name, email, phone string) (Contact, error)` constructor that validates the email contains `@`
- A `String() string` method implementing [`fmt.Stringer`](https://pkg.go.dev/fmt#Stringer)
- A `func PrintCard(c Contact)` that pretty-prints the card

**Extension ideas:** add a `contacts []Contact` slice and write `FindByName(name string) (Contact, bool)`.

## Official Documentation

- [`fmt`](https://pkg.go.dev/fmt) — Errorf, Sprintf, Stringer interface
- [`log`](https://pkg.go.dev/log) — Fatal and other logging functions
- [Language Spec: Function types](https://go.dev/ref/spec#Function_types) — multiple return values
- [Language Spec: Struct types](https://go.dev/ref/spec#Struct_types) — struct declarations and embedding
- [Language Spec: Method declarations](https://go.dev/ref/spec#Method_declarations) — value and pointer receivers
- [Effective Go: Methods](https://go.dev/doc/effective_go#methods) — pointer vs value receivers
- [Go Tour: Methods and interfaces](https://go.dev/tour/methods/1) — interactive methods tour
