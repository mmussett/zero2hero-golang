# Day 28: Reflection

## Core Concept: Inspect Types at Runtime

The `reflect` package lets you inspect and manipulate values at runtime without knowing their types at compile time. Use it sparingly — reflection bypasses type safety and is significantly slower than direct field access.

## Type and Value

```go
x := 42
t := reflect.TypeOf(x)   // reflect.Type → "int"
v := reflect.ValueOf(x)  // reflect.Value → 42

fmt.Println(t.Kind())    // int
fmt.Println(v.Int())     // 42
```

`Kind` is the underlying category (`struct`, `slice`, `map`, `ptr`, `int`, `string`, …). `Type` is the specific named type.

## Iterating Struct Fields

```go
type Person struct {
    Name  string `describe:"full name"`
    Age   int    `describe:"age in years"`
    email string // unexported — not visible via reflect
}

func Describe(v any) {
    t   := reflect.TypeOf(v)
    val := reflect.ValueOf(v)

    if t.Kind() == reflect.Ptr {
        t   = t.Elem()
        val = val.Elem()
    }

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        if !field.IsExported() {
            continue
        }
        tag  := field.Tag.Get("describe")
        fval := val.Field(i)
        fmt.Printf("%-15s %-20v %s\n", field.Name, fval, tag)
    }
}
```

## Setting Values via Reflection

```go
v := reflect.ValueOf(&x).Elem() // must pass pointer, then dereference
if v.CanSet() {
    v.SetInt(100)
}
```

## Calling Methods

```go
method := reflect.ValueOf(obj).MethodByName("String")
if method.IsValid() {
    results := method.Call(nil)
    fmt.Println(results[0].String())
}
```

## reflect.DeepEqual

```go
if reflect.DeepEqual(a, b) { /* structurally equal */ }
```

Used internally by `testify/assert.Equal`.

## When Not to Use Reflect

- When generics solve the problem (Go 1.18+)
- In hot paths (5–50× slower than direct access)
- When a type switch is sufficient
- When you're building "magic" frameworks that obscure control flow

## Day Project: Custom Describe Utility

Build a `describe` package with:
1. `Describe(v any) string` — pretty-prints any struct with field names, values, and `describe:` tags
2. `DiffStructs(a, b any) []FieldDiff` — returns fields that differ between two structs of the same type
3. `CopyFields(dst, src any, fields ...string)` — copies named exported fields between structs of the same type

Test with several struct types including nested structs and pointer fields.

Run with: `go run .`

**Extension ideas:** implement a `Validate` function that checks `required:""` struct tags; build a minimal query builder that maps structs to SQL INSERT/SELECT using `db:""` tags.
