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

## Labs

### Lab 1: reflect.TypeOf / reflect.ValueOf — Inspecting Any Value

**What you'll practise:** Using `reflect.TypeOf` and `reflect.ValueOf` to inspect the type, kind, and value of several Go types at runtime.

**Task:**
Write a `printInfo` function that accepts `any` and prints the type name, kind, and value for each of: `int`, `string`, a named struct, a pointer to a struct, and `[]int`.

**Steps:**
1. Import `"fmt"` and `"reflect"`
2. Write `printInfo(v any)` that calls `reflect.TypeOf` and `reflect.ValueOf`
3. Call it with five different values and read the `Kind()` output

```go
func printInfo(v any) {
    t := reflect.TypeOf(v)
    val := reflect.ValueOf(v)
    fmt.Printf("Type: %-20s Kind: %-10s Value: %v\n", t, t.Kind(), val)
}

type Point struct{ X, Y int }

func main() {
    printInfo(42)
    printInfo("hello")
    printInfo(Point{1, 2})
    printInfo(&Point{3, 4})
    printInfo([]int{10, 20, 30})
}
```

**Expected output:**
```
Type: int                  Kind: int        Value: 42
Type: string               Kind: string     Value: hello
Type: main.Point           Kind: struct     Value: {1 2}
Type: *main.Point          Kind: ptr        Value: &{3 4}
Type: []int                Kind: slice      Value: [10 20 30]
```

**Checkpoint:** All five lines print correctly. You can explain the difference between `Type` (named type) and `Kind` (category).

---

### Lab 2: Struct Field Iteration — Printing Every Field

**What you'll practise:** Using `reflect.Type.NumField` and `reflect.Value.Field` to iterate over all exported fields of a struct.

**Task:**
Write a `printFields` function that accepts any struct (or pointer to struct) and prints the name, type, and value of each exported field.

**Steps:**
1. Dereference pointers: if `Kind() == reflect.Ptr`, call `.Elem()` on both type and value
2. Loop from 0 to `t.NumField()`
3. Skip unexported fields with `!field.IsExported()`

```go
type Employee struct {
    Name       string
    Department string
    Salary     float64
    active     bool // unexported — should be skipped
}

func printFields(v any) {
    t   := reflect.TypeOf(v)
    val := reflect.ValueOf(v)
    if t.Kind() == reflect.Ptr {
        t   = t.Elem()
        val = val.Elem()
    }
    for i := 0; i < t.NumField(); i++ {
        f := t.Field(i)
        if !f.IsExported() { continue }
        fmt.Printf("%-15s %-10s = %v\n", f.Name, f.Type, val.Field(i))
    }
}
```

**Expected output:**
```
Name            string     = Alice
Department      string     = Engineering
Salary          float64    = 95000
```

**Checkpoint:** `active` does not appear in the output. The three exported fields print with correct types and values.

---

### Lab 3: Struct Tags — Reading JSON Field Names

**What you'll practise:** Using `StructField.Tag.Get("json")` to extract JSON tag values and build a field-name mapper.

**Task:**
Write `jsonFieldNames(v any) []string` that returns the JSON key for each exported field. If a field has no `json` tag, fall back to the field name.

**Steps:**
1. Iterate fields with `t.NumField()`
2. Call `f.Tag.Get("json")` on each field
3. Strip options after the first comma (e.g. `"name,omitempty"` → `"name"`)
4. Fall back to `f.Name` when the tag is absent or `"-"`

```go
type User struct {
    ID        int    `json:"id"`
    FullName  string `json:"full_name,omitempty"`
    Password  string `json:"-"`
    CreatedAt string // no tag
}

func jsonFieldNames(v any) []string {
    t := reflect.TypeOf(v)
    if t.Kind() == reflect.Ptr { t = t.Elem() }
    var names []string
    for i := 0; i < t.NumField(); i++ {
        f   := t.Field(i)
        tag := f.Tag.Get("json")
        // parse tag ...
    }
    return names
}
```

**Expected output:**
```
[id full_name CreatedAt]
```

**Checkpoint:** `Password` is excluded (tag is `"-"`). `FullName` maps to `"full_name"` (stripping `,omitempty`). `CreatedAt` uses its field name as a fallback.

---

### Lab 4: Setting Values — Addressable Fields via Elem()

**What you'll practise:** Using `reflect.Value.CanSet` and `v.Elem()` to mutate a field, and deliberately triggering the panic when the value is not addressable.

**Task:**
Write `setField(ptr any, name string, value any)` that sets a named exported field on a pointer-to-struct. Then call it on a non-pointer value to observe the panic.

**Steps:**
1. Accept `ptr any`, call `reflect.ValueOf(ptr).Elem()` to get the addressable struct value
2. Look up the field by name with `.FieldByName`
3. Check `CanSet()` before setting; return an error if not settable
4. Try passing a non-pointer to observe the panic

```go
func setField(ptr any, name string, value any) error {
    v := reflect.ValueOf(ptr)
    if v.Kind() != reflect.Ptr {
        return fmt.Errorf("setField: must pass a pointer, got %s", v.Kind())
    }
    field := v.Elem().FieldByName(name)
    if !field.IsValid() {
        return fmt.Errorf("setField: no field %q", name)
    }
    if !field.CanSet() {
        return fmt.Errorf("setField: field %q is not settable", name)
    }
    field.Set(reflect.ValueOf(value))
    return nil
}
```

**Expected output:**
```
Before: {Alice Engineering 0}
After:  {Alice Engineering 95000}
setField: must pass a pointer, got struct
```

**Checkpoint:** The field is updated in-place when a pointer is passed. A non-pointer input returns a clear error message instead of panicking (because you checked `Kind` first).

---

### Lab 5: Function Calling — Dynamic Dispatch with reflect.Value.Call

**What you'll practise:** Using `reflect.Value.Call` to invoke a function by name at runtime, building a simple string-to-function dispatcher.

**Task:**
Build a `dispatch(name string, args ...any)` function that looks up a registered function by string name and calls it with the provided arguments.

**Steps:**
1. Store functions in a `map[string]any`
2. Use `reflect.ValueOf(fn).Call(in)` where `in` is `[]reflect.Value`
3. Convert `args` to `[]reflect.Value` using a helper
4. Print the return values

```go
var registry = map[string]any{
    "add":    func(a, b int) int { return a + b },
    "greet":  func(name string) string { return "Hello, " + name + "!" },
    "square": func(n int) int { return n * n },
}

func dispatch(name string, args ...any) ([]any, error) {
    fn, ok := registry[name]
    if !ok {
        return nil, fmt.Errorf("unknown function: %s", name)
    }
    in := make([]reflect.Value, len(args))
    for i, a := range args { in[i] = reflect.ValueOf(a) }
    out := reflect.ValueOf(fn).Call(in)
    // convert out to []any ...
}
```

**Expected output:**
```
add(3, 4) = 7
greet("World") = Hello, World!
square(9) = 81
```

**Checkpoint:** All three registered functions are called successfully via the dispatcher. An unknown function name returns a descriptive error.

---

### Lab 6: DiffStructs — Comparing Two Structs Field by Field

**What you'll practise:** Iterating struct fields with `reflect` and comparing values with `reflect.DeepEqual` to detect changes.

**Task:**
Write `DiffStructs(a, b any) []string` that returns the names of fields that differ between two structs of the same type.

**Steps:**
1. Verify `a` and `b` have the same type; return an error if not
2. Iterate exported fields
3. Use `reflect.DeepEqual(va.Field(i).Interface(), vb.Field(i).Interface())` to compare
4. Collect field names where values differ

```go
type Config struct {
    Host    string
    Port    int
    Debug   bool
    Timeout int
}

func DiffStructs(a, b any) ([]string, error) {
    ta, tb := reflect.TypeOf(a), reflect.TypeOf(b)
    if ta != tb {
        return nil, fmt.Errorf("type mismatch: %s vs %s", ta, tb)
    }
    va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
    var diffs []string
    for i := 0; i < ta.NumField(); i++ {
        if !reflect.DeepEqual(va.Field(i).Interface(), vb.Field(i).Interface()) {
            diffs = append(diffs, ta.Field(i).Name)
        }
    }
    return diffs, nil
}
```

**Expected output:**
```
Changed fields: [Port Debug]
```

**Checkpoint:** Only fields with different values appear in the result. Identical fields are absent.

---

### Lab 7: CopyFields — Copying Matching Fields Between Struct Types

**What you'll practise:** Matching exported fields by name between two different struct types and copying values with reflection.

**Task:**
Write `CopyFields(dst, src any, fields ...string)` that copies the named exported fields from `src` into `dst`, where `dst` and `src` may be different struct types.

**Steps:**
1. `dst` must be a pointer to a struct; dereference with `.Elem()`
2. For each field name in `fields`, look it up in both `src` and `dst`
3. Only copy if both fields exist, are exported, and have the same type
4. Return an error for mismatched types or non-pointer `dst`

```go
type APIUser struct {
    Name  string
    Email string
    Role  string
}

type DBUser struct {
    Name      string
    Email     string
    Role      string
    CreatedAt string
}

// CopyFields copies named fields from src into dst.
func CopyFields(dst, src any, fields ...string) error {
    dv := reflect.ValueOf(dst).Elem()
    sv := reflect.ValueOf(src)
    for _, name := range fields {
        sf := sv.FieldByName(name)
        df := dv.FieldByName(name)
        if !sf.IsValid() || !df.IsValid() { continue }
        if sf.Type() != df.Type() {
            return fmt.Errorf("type mismatch for field %s", name)
        }
        df.Set(sf)
    }
    return nil
}
```

**Expected output:**
```
Before: {  }
After:  {Alice alice@example.com admin}
```

**Checkpoint:** Only the named fields are copied. `CreatedAt` on the destination is untouched. Passing a non-pointer `dst` returns a clear error.

---

### Final Lab (Project): Custom Describe Utility

**What you'll practise:** Combining TypeOf/ValueOf, struct tag inspection, DiffStructs, and CopyFields into a reusable `describe` package.

**Task:**
Build a `describe` package with the three core functions and demonstrate them on several struct types including nested structs and pointer fields.

**Steps:**
1. `Describe(v any) string` — pretty-prints any struct with field names, values, and `describe:` tags
2. `DiffStructs(a, b any) []FieldDiff` — returns fields that differ between two structs of the same type
3. `CopyFields(dst, src any, fields ...string)` — copies named exported fields between structs of the same type
4. Test with several struct types including nested structs and pointer fields

```go
type FieldDiff struct {
    Name string
    Old  any
    New  any
}

// Describe pretty-prints v with field names, values, and describe tags.
func Describe(v any) string {
    var sb strings.Builder
    t   := reflect.TypeOf(v)
    val := reflect.ValueOf(v)
    if t.Kind() == reflect.Ptr { t = t.Elem(); val = val.Elem() }
    fmt.Fprintf(&sb, "%s:\n", t.Name())
    for i := 0; i < t.NumField(); i++ {
        f := t.Field(i)
        if !f.IsExported() { continue }
        tag := f.Tag.Get("describe")
        fmt.Fprintf(&sb, "  %-15s = %-20v %s\n", f.Name, val.Field(i), tag)
    }
    return sb.String()
}
```

**Expected output:**
```
Person:
  Name            = Alice               full name
  Age             = 30                  age in years
Changed: [{Salary 80000 95000}]
Copied: {Alice alice@example.com}
```

**Checkpoint:** `Describe` shows all exported fields with tags. `DiffStructs` reports only changed fields. `CopyFields` transfers values without touching unspecified fields. Run with: `go run .`

**Extension ideas:** implement a `Validate` function that checks `required:""` struct tags; build a minimal query builder that maps structs to SQL INSERT/SELECT using `db:""` tags.

## Official Documentation

- [`reflect`](https://pkg.go.dev/reflect) — `TypeOf`, `ValueOf`, `Type`, `Value`, `Kind`, `StructField`, `StructTag.Get`, `Value.CanSet`, `DeepEqual`, `Value.MethodByName`, `Value.Call`
- [Language Spec — Reflection](https://go.dev/ref/spec#Package_unsafe) — type system foundations underlying reflect
- [Go Blog: The Laws of Reflection](https://go.dev/blog/laws-of-reflection) — essential reading before using the `reflect` package
- [Language Spec — Struct tags](https://go.dev/ref/spec#Struct_types) — tag syntax and conventions
