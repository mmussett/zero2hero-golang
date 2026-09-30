# Go Primitive Types — Reference

## Integer Types

Go provides both signed and unsigned integer types at fixed widths, plus architecture-sized variants.

| Type | Size | Range |
|------|------|-------|
| `int8` | 8-bit | −128 to 127 |
| `int16` | 16-bit | −32 768 to 32 767 |
| `int32` / `rune` | 32-bit | −2 147 483 648 to 2 147 483 647 |
| `int64` | 64-bit | −9.2 × 10¹⁸ to 9.2 × 10¹⁸ |
| `int` | platform | 32-bit on 32-bit OS, 64-bit on 64-bit OS |
| `uint8` / `byte` | 8-bit | 0 to 255 |
| `uint16` | 16-bit | 0 to 65 535 |
| `uint32` | 32-bit | 0 to 4 294 967 295 |
| `uint64` | 64-bit | 0 to 1.8 × 10¹⁹ |
| `uint` | platform | matches `int` width |
| `uintptr` | platform | large enough to hold a pointer value |

**`int` is the default inferred integer type** — `x := 42` gives `int`, not `int64`.

**Type conversion is always explicit.** `int32(x)` truncates; there is no implicit widening.

Integer overflow wraps silently in Go (unlike Rust, there is no compile-time protection in release builds). Guard with `math.MaxInt` / `math.MinInt` constants or `math/bits` for checked arithmetic.

```go
import "math/bits"
result, overflow := bits.Add64(a, b, 0)
```

`uintptr` is an integer, not a pointer — the GC does not follow it. Use `unsafe.Pointer` for genuine pointer manipulation.

---

## Floating-Point

Both types follow IEEE 754.

| Type | Size | Precision |
|------|------|-----------|
| `float32` | 32-bit | ~7 decimal digits |
| `float64` | 64-bit | ~15 decimal digits |

**`float64` is the default** — `x := 3.14` gives `float64`.

Critical rule: `NaN != NaN` is always **true** — use `math.IsNaN(x)` to check. Never compare floats with `==`; prefer `math.Abs(a-b) < epsilon`.

```go
import "math"
if math.IsNaN(x) { /* ... */ }
if math.IsInf(x, 1) { /* positive infinity */ }
```

`float32` loses precision quickly; prefer `float64` unless interoperating with graphics/audio APIs that demand 32-bit.

---

## Complex Numbers

```go
c64  := complex(1.0, 2.0)  // complex64  (float32 parts)
c128 := 3 + 4i              // complex128 (float64 parts)

real(c128)  // 3.0
imag(c128)  // 4.0
```

Rarely used outside scientific computing and FFT implementations.

---

## bool

```go
var b bool      // zero value: false
b = true
b = x > 0 && y != 0
```

- Size: 1 byte
- Only `true` and `false` — Go has no truthiness for integers or pointers
- Short-circuit evaluation: `&&` and `||`

---

## byte and rune

`byte` is an alias for `uint8`. `rune` is an alias for `int32` and represents a single Unicode code point.

```go
var b byte = 'A'   // 65
var r rune = '🎉'  // 127881 (U+1F389)
```

Key distinction from most languages:

- A `string` is a **read-only byte slice** (`[]byte` under the hood), not a slice of characters
- `len(s)` returns **byte count**, not character count
- Iterating `for i, r := range s` yields `(byte offset, rune)` — safe for Unicode
- Iterating `for i := 0; i < len(s); i++` yields raw bytes

```go
s := "héllo"
len(s)          // 6, not 5 (é is 2 bytes in UTF-8)
len([]rune(s))  // 5
```

---

## string

`string` is an immutable, UTF-8-encoded byte sequence. Internally it is a two-word struct: `(pointer, length)`.

```go
var s string      // zero value: ""
s = "hello"
s += " world"     // allocates a new string

// Efficient building
var b strings.Builder
b.WriteString("hello")
result := b.String()
```

### Key operations

| Operation | Package | Notes |
|-----------|---------|-------|
| `len(s)` | builtin | byte length |
| `s[i]` | builtin | byte at offset i (`byte`) |
| `s[i:j]` | builtin | substring (shares memory) |
| `strings.Contains` | `strings` | |
| `strings.Split` / `strings.Join` | `strings` | |
| `strings.TrimSpace` | `strings` | |
| `strings.HasPrefix` / `HasSuffix` | `strings` | |
| `strings.ToUpper` / `ToLower` | `strings` | |
| `fmt.Sprintf` | `fmt` | formatted string |
| `strconv.Itoa` | `strconv` | int → string |
| `strconv.Atoi` | `strconv` | string → int, returns error |

Converting between `string` and `[]byte` copies the data:

```go
b := []byte("hello")  // copy
s := string(b)        // copy
```

---

## Array

Arrays in Go have a **fixed, compile-time length** and are **value types** — assignment copies the entire array.

```go
var a [5]int          // [0 0 0 0 0]
b := [3]string{"x", "y", "z"}
c := [...]int{1, 2, 3}  // compiler counts elements
```

Arrays are rarely used directly; slices (below) are the idiomatic collection type.

---

## Slice

A slice is a **view into an underlying array**: three words `(pointer, length, capacity)`.

```go
s := []int{1, 2, 3}
s = append(s, 4)        // may reallocate
s2 := s[1:3]            // shares memory with s
```

| Operation | Effect |
|-----------|--------|
| `len(s)` | number of elements |
| `cap(s)` | elements until next reallocation |
| `append(s, v)` | may grow; always use returned slice |
| `copy(dst, src)` | copies min(len(dst), len(src)) elements |
| `make([]T, n, cap)` | allocates a new slice |

**Zero value of a slice is `nil`**, which is safe to `append` to and has `len` 0.

Slices share memory — modifying a sub-slice modifies the original. Use `copy` or the three-index slice `s[low:high:max]` to limit capacity and prevent accidental sharing.

---

## map

Maps are reference types with an unordered key-value store.

```go
var m map[string]int     // nil map — reads return zero, writes panic
m = make(map[string]int) // usable map

m["key"] = 1
v, ok := m["key"]   // comma-ok: ok is false if key absent
delete(m, "key")
```

- **Zero value is `nil`** — always `make` before writing
- Keys must be `comparable` (supports `==`): all primitives, structs with comparable fields
- `float64` keys are legal but NaN-as-key is undefined behaviour — avoid
- Iteration order is **randomised** by design

---

## Pointer

Go has explicit pointers but no pointer arithmetic (outside `unsafe`).

```go
x := 42
p := &x     // *int — address of x
*p = 100    // dereference and assign
fmt.Println(x) // 100
```

`new(T)` allocates a zeroed `T` and returns `*T`. Prefer composite literals with `&` for structs:

```go
p := &Person{Name: "Alice"}  // idiomatic
q := new(Person)             // also valid
```

The zero value of any pointer type is `nil`. Always check before dereferencing.

---

## interface{}  /  any

`any` (alias for `interface{}`) holds a value of any type.

```go
var v any = 42
v = "now a string"
```

Extracting the underlying value requires a **type assertion**:

```go
s, ok := v.(string)   // safe: ok is false if wrong type
s  := v.(string)      // panics if wrong type
```

Or a **type switch**:

```go
switch t := v.(type) {
case int:    fmt.Println("int:", t)
case string: fmt.Println("string:", t)
default:     fmt.Println("unknown")
}
```

---

## Zero Values Summary

Every type in Go has a well-defined zero value — there is no uninitialised memory.

| Type | Zero value |
|------|-----------|
| `bool` | `false` |
| `int`, `float64`, etc. | `0` |
| `string` | `""` |
| pointer, slice, map, channel, func | `nil` |
| struct | each field at its zero value |
| array | all elements at their zero values |

---

## Type Sizes at a Glance

| Type | Size |
|------|------|
| `bool` | 1 byte |
| `byte` / `uint8` | 1 byte |
| `rune` / `int32` | 4 bytes |
| `int` / `uint` | 4 or 8 bytes (platform) |
| `float32` | 4 bytes |
| `float64` | 8 bytes |
| `string` | 2 × pointer size (16 bytes on 64-bit) |
| `slice` | 3 × pointer size (24 bytes on 64-bit) |
| `interface{}` | 2 × pointer size (16 bytes on 64-bit) |

Use `unsafe.Sizeof(v)` to verify at runtime.
