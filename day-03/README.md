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
