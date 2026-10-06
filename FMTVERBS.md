# fmt Format Verbs

`fmt.Printf`, `fmt.Sprintf`, `fmt.Fprintf`, and friends use `%` verbs to format values. This reference covers every verb, flag, and width/precision modifier in the Go standard library.

Official documentation: [`fmt` package](https://pkg.go.dev/fmt)

---

## General

| Verb | Meaning | Example output |
|------|---------|----------------|
| `%v` | Default format | `42`, `true`, `[1 2 3]` |
| `%+v` | Struct with field names | `{Name:Alice Age:30}` |
| `%#v` | Go syntax representation | `main.Person{Name:"Alice", Age:30}` |
| `%T` | Type of the value | `int`, `[]string`, `main.Person` |
| `%%` | Literal `%` sign | `%` |

---

## Boolean

| Verb | Meaning | Example output |
|------|---------|----------------|
| `%t` | `true` or `false` | `true` |

---

## Integer

| Verb | Meaning | Example output |
|------|---------|----------------|
| `%d` | Base 10 | `42` |
| `%b` | Base 2 (binary) | `101010` |
| `%o` | Base 8 (octal) | `52` |
| `%O` | Base 8 with `0o` prefix | `0o52` |
| `%x` | Base 16, lowercase | `2a` |
| `%X` | Base 16, uppercase | `2A` |
| `%c` | Unicode code point → character | `*` (for 42) |
| `%q` | Single-quoted character literal | `'*'` |
| `%U` | Unicode format | `U+002A` |

---

## Floating-point

| Verb | Meaning | Example output |
|------|---------|----------------|
| `%f` | Decimal, no exponent | `3.141593` |
| `%F` | Same as `%f` | `3.141593` |
| `%.2f` | Decimal, 2 decimal places | `3.14` |
| `%e` | Scientific notation, lowercase | `3.141593e+00` |
| `%E` | Scientific notation, uppercase | `3.141593E+00` |
| `%g` | Shortest of `%e` / `%f` | `3.141592653589793` |
| `%G` | Shortest of `%E` / `%F` | `3.141592653589793` |
| `%x` | Hexadecimal notation, lowercase | `-0x1.921fb54442d18p+01` |
| `%X` | Hexadecimal notation, uppercase | `-0X1.921FB54442D18P+01` |

---

## String and []byte

| Verb | Meaning | Example output |
|------|---------|----------------|
| `%s` | Plain string / byte slice | `hello` |
| `%q` | Double-quoted, Go-escaped | `"hello\nworld"` |
| `%x` | Hex encoding, lowercase | `68656c6c6f` |
| `%X` | Hex encoding, uppercase | `68656C6C6F` |

---

## Pointer

| Verb | Meaning | Example output |
|------|---------|----------------|
| `%p` | Base-16 pointer address | `0xc0000b4010` |

---

## Width, Precision and Flags

| Syntax | Effect |
|--------|--------|
| `%8d` | Right-align in a field of width 8 |
| `%-8d` | Left-align in a field of width 8 |
| `%08d` | Zero-pad to width 8 |
| `%8.2f` | Width 8, 2 decimal places |
| `%+d` | Always show sign (`+42`, `-7`) |
| `% d` | Space before positive numbers (` 42`) |
| `%#x` | Alternate form: add `0x` prefix |
| `%#o` | Alternate form: add `0` prefix |
| `%#b` | Alternate form: add `0b` prefix |

### Width and precision rules

- **Width** sets the minimum field width. The value is right-padded with spaces by default.
- **Precision** (`.N`) sets the number of decimal places for floats, or the maximum number of characters for strings.
- Combine both: `%8.2f` means width 8, 2 decimal places.
- For integers, precision sets the minimum number of digits: `%.5d` on `42` → `00042`.

---

## When to Use Which Verb

| Situation | Use |
|-----------|-----|
| Printing any value during development | `%v` |
| Inspecting a struct and need field names | `%+v` |
| Generating valid Go source code | `%#v` |
| Checking what type a variable is | `%T` |
| Formatting a number for display | `%d` / `%f` |
| Formatting currency / fixed decimals | `%.2f` |
| Comparing string contents including escapes | `%q` |
| Printing a character from its code point | `%c` |
| Debugging memory addresses | `%p` |
| Encoding bytes as hex (e.g. hashes) | `%x` |

---

## Examples

```go
package main

import "fmt"

type Point struct{ X, Y int }

func main() {
    p := Point{3, 4}

    fmt.Printf("%v\n", p)          // {3 4}
    fmt.Printf("%+v\n", p)         // {X:3 Y:4}
    fmt.Printf("%#v\n", p)         // main.Point{X:3, Y:4}
    fmt.Printf("%T\n", p)          // main.Point

    fmt.Printf("%d %b %o %x\n", 42, 42, 42, 42)  // 42 101010 52 2a
    fmt.Printf("%08.2f\n", 3.14159)               // 00003.14
    fmt.Printf("%q\n", "hello\nworld")            // "hello\nworld"
    fmt.Printf("%10s|\n", "right")                //      right|
    fmt.Printf("%-10s|\n", "left")                // left      |
}
```
