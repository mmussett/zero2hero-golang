# Day 05: Arrays, Slices, and Strings

## Arrays (Rarely Used Directly)

Arrays have a **compile-time fixed length** and are **value types**:

```go
var a [5]int              // [0 0 0 0 0]
b := [3]string{"x","y","z"}
c := [...]int{1,2,3}      // compiler counts: [3]int
```

Because `[3]int` and `[4]int` are different types, functions usually accept slices instead.

## Slices (The Idiomatic List)

A slice is a three-word header that describes a window into an underlying array:

```
┌─────────┬────────┬──────────┐
│ pointer │ length │ capacity │
└─────────┴────────┴──────────┘
     │         │         │
     │         │         └── how many elements the backing array can hold
     │         └──────────── how many elements are currently visible
     └────────────────────── where the window starts in the backing array
```

The backing array is allocated separately. Multiple slices can point to the same backing array (they share memory).

---

### Creating slices

**Slice literal** — the most common way:

```go
s := []int{10, 20, 30}   // len=3, cap=3
```

**`make`** — preallocate when you know the size up front:

```go
s := make([]int, 5)       // len=5, cap=5  — all elements are 0
s := make([]int, 0, 10)   // len=0, cap=10 — empty but pre-allocated
```

Using `make([]T, 0, n)` before a loop avoids repeated reallocations as you `append`.

**`nil` slice** — the zero value; safe to use immediately:

```go
var s []int               // len=0, cap=0, s == nil
s = append(s, 1)          // fine — append handles nil correctly
```

A `nil` slice and an empty slice (`[]int{}`) behave identically with `len`, `cap`, and `append`. The only difference is `s == nil`.

---

### append — adding elements

`append` adds elements to the end of a slice and always returns the (possibly new) slice. **Always assign the result back.**

```go
s := []int{1, 2, 3}
s = append(s, 4)          // [1 2 3 4]
s = append(s, 5, 6, 7)    // [1 2 3 4 5 6 7]  — multiple at once
```

**Appending one slice to another** — use the `...` spread operator:

```go
a := []int{1, 2, 3}
b := []int{4, 5, 6}
a = append(a, b...)       // [1 2 3 4 5 6]
```

**How reallocation works:**

When `append` needs more space than `cap` allows, Go allocates a new, larger backing array (roughly doubling in size), copies all elements, and returns a slice pointing to the new array. The old backing array is abandoned.

```go
s := make([]int, 0, 3)
fmt.Println(len(s), cap(s))  // 0 3

s = append(s, 1, 2, 3)
fmt.Println(len(s), cap(s))  // 3 3

s = append(s, 4)             // cap exceeded — new backing array
fmt.Println(len(s), cap(s))  // 4 6 (or similar; exact growth factor varies)
```

Because reallocation returns a new slice, code like `append(s, x)` without reassigning `s` silently loses the new element. Always write `s = append(s, x)`.

---

### Iterating over a slice

**`range`** — the idiomatic way; gives index and value:

```go
fruits := []string{"apple", "banana", "cherry"}

for i, v := range fruits {
    fmt.Printf("[%d] = %s\n", i, v)
}
// [0] = apple
// [1] = banana
// [2] = cherry
```

Use `_` to discard the index or value when you don't need it:

```go
for _, v := range fruits {   // index not needed
    fmt.Println(v)
}

for i := range fruits {      // value not needed (Go 1.22+: also works for just the index)
    fmt.Println(i)
}
```

**Plain `for` loop** — when you need finer control (e.g., skipping elements or iterating in reverse):

```go
// Reverse iteration
for i := len(fruits) - 1; i >= 0; i-- {
    fmt.Println(fruits[i])
}
```

> **`range` gives a copy of each element.** Modifying `v` inside the loop does not change the slice. To modify elements in place, use the index: `fruits[i] = "mango"`.

---

### Slicing — creating sub-slices

The expression `s[low:high]` creates a new slice header pointing into the same backing array. No data is copied.

```go
s := []int{0, 1, 2, 3, 4, 5}

a := s[1:4]   // [1 2 3]       — elements at index 1, 2, 3
b := s[:3]    // [0 1 2]       — from the start
c := s[3:]    // [3 4 5]       — to the end
d := s[:]     // [0 1 2 3 4 5] — the whole slice
```

Because sub-slices share the backing array, modifying one is visible through the other:

```go
a[0] = 99
fmt.Println(s)  // [0 99 2 3 4 5] — s[1] changed!
```

**Three-index slice** `s[low:high:max]` limits the capacity of the result, preventing accidental writes through the sub-slice:

```go
safe := s[1:4:4]   // len=3, cap=3 — cannot grow into s[4], s[5]
```

---

### copy — independent duplicate

`copy(dst, src)` copies elements from `src` to `dst`, returns the number copied, and the two slices are independent afterwards:

```go
original := []int{1, 2, 3, 4, 5}
clone := make([]int, len(original))
copy(clone, original)

clone[0] = 99
fmt.Println(original) // [1 2 3 4 5] — unchanged
fmt.Println(clone)    // [99 2 3 4 5]
```

`copy` copies `min(len(dst), len(src))` elements. If `dst` is shorter, only `len(dst)` elements are copied.

---

### Common in-place operations

**Delete an element (order preserved):**

```go
s := []int{10, 20, 30, 40, 50}
i := 2  // delete index 2
s = append(s[:i], s[i+1:]...)
fmt.Println(s)  // [10 20 40 50]
```

**Delete an element (fast, order not preserved):**

```go
s[i] = s[len(s)-1]  // overwrite target with last element
s = s[:len(s)-1]    // trim the last slot
fmt.Println(s)      // [10 20 50 40]  — 50 moved to position 2
```

**Insert an element at index `i`:**

```go
s := []int{10, 20, 30, 40}
i := 2
s = append(s[:i], append([]int{99}, s[i:]...)...)
fmt.Println(s)  // [10 20 99 30 40]
```

**Filter (keep only even numbers):**

```go
s := []int{1, 2, 3, 4, 5, 6}
result := s[:0]  // re-use the backing array — no allocation
for _, v := range s {
    if v%2 == 0 {
        result = append(result, v)
    }
}
fmt.Println(result)  // [2 4 6]
```

---

### len and cap

| Expression | Meaning |
|-----------|---------|
| `len(s)` | Number of elements currently in the slice |
| `cap(s)` | Total elements the backing array can hold from `s[0]` |
| `len(s) == 0` | Slice is empty (true for both `nil` and `[]int{}`) |
| `s == nil` | Slice is nil (not merely empty) |

## Strings

A `string` is a read-only `[]byte` (UTF-8). `len(s)` is the byte count, not the character count.

```go
s := "héllo"
len(s)          // 6 (é is 2 UTF-8 bytes)
len([]rune(s))  // 5 (rune count)

for i, r := range s {   // i = byte offset, r = rune
    fmt.Printf("%d: %c (%d)\n", i, r, r)
}
```

### Essential [`strings`](https://pkg.go.dev/strings) Functions

```go
strings.Contains(s, "ell")
strings.HasPrefix(s, "he")
strings.HasSuffix(s, "lo")
strings.Count(s, "l")
strings.ToUpper(s) / strings.ToLower(s)
strings.TrimSpace(s)
strings.Split(s, ",")
strings.Join(parts, ", ")
strings.Replace(s, "old", "new", -1)
strings.Fields(s)   // split on any whitespace
```

### Building Strings Efficiently

```go
var b strings.Builder
for _, word := range words {
    b.WriteString(word)
    b.WriteByte(' ')
}
result := b.String()
```

Concatenating with `+=` in a loop is O(n²) — always use [`strings.Builder`](https://pkg.go.dev/strings#Builder) or [`strings.Join`](https://pkg.go.dev/strings#Join).

---

## Runes

### What is a rune?

`rune` is Go's name for a single Unicode code point. It is an alias for `int32`:

```go
type rune = int32  // built into the language — not defined in any package
```

A **code point** is the number Unicode assigns to each character. Every character in every human script has a unique code point, written as U+XXXX in hexadecimal:

| Character | Code point | Decimal |
|-----------|-----------|---------|
| `A` | U+0041 | 65 |
| `é` | U+00E9 | 233 |
| `世` | U+4E16 | 19990 |
| `😀` | U+1F600 | 128512 |

A rune literal is written with single quotes:

```go
var r rune = 'A'   // r == 65
fmt.Println(r)     // 65    (the integer value)
fmt.Printf("%c\n", r) // A (the character)
fmt.Printf("%U\n", r) // U+0041 (Unicode format)
```

Rune literals support the same escape sequences as strings: `'\n'` (newline), `'\t'` (tab), `'\\'` (backslash), `'\''` (single quote).

---

### How Go stores strings: UTF-8 encoding

A Go `string` is a sequence of **bytes**, not characters. The bytes are encoded as **UTF-8**, a variable-width encoding where each code point occupies 1 to 4 bytes:

| Code point range | Bytes used | Examples |
|-----------------|-----------|---------|
| U+0000 – U+007F | 1 byte | ASCII: `A`, `0`, `!` |
| U+0080 – U+07FF | 2 bytes | `é`, `ñ`, `ü` |
| U+0800 – U+FFFF | 3 bytes | `世`, `界`, `€` |
| U+10000 – U+10FFFF | 4 bytes | `😀`, `🎉` |

This is why `len(s)` returns the **byte count**, which may be larger than the number of visible characters:

```go
s := "héllo"
fmt.Println(len(s))           // 6 — é occupies 2 bytes (U+00E9 → 0xC3 0xA9)
fmt.Println(len([]rune(s)))   // 5 — five visible characters
```

ASCII text is always 1 byte per character, so for pure ASCII strings `len(s)` equals the character count. Once you introduce non-ASCII characters, the counts diverge.

---

### The classic mistake: indexing a string by position

Square-bracket indexing (`s[i]`) gives the **byte** at position `i`, not the character:

```go
s := "héllo"
fmt.Println(s[0])          // 104 — byte value of 'h' (correct)
fmt.Println(s[1])          // 195 — first byte of 'é' (0xC3), not 'é'!
fmt.Printf("%c\n", s[1])   // Ã   — wrong character
```

`é` is encoded as two bytes (0xC3, 0xA9). Indexing at position 1 lands in the middle of that sequence. To index by **character** (code point), convert to `[]rune` first:

```go
runes := []rune(s)
fmt.Printf("%c\n", runes[1])  // é — correct
```

---

### Iterating over a string

Go provides two ways to iterate:

**By byte** — use a plain `for` loop with an index:

```go
s := "héllo"
for i := 0; i < len(s); i++ {
    fmt.Printf("byte[%d] = %d\n", i, s[i])
}
// byte[0] = 104
// byte[1] = 195   ← first byte of é
// byte[2] = 169   ← second byte of é
// byte[3] = 108   ← l
// byte[4] = 108   ← l
// byte[5] = 111   ← o
```

**By rune (character)** — use `range`:

```go
for i, r := range s {
    fmt.Printf("byte_offset[%d] = %c (U+%04X)\n", i, r, r)
}
// byte_offset[0] = h (U+0068)
// byte_offset[1] = é (U+00E9)   ← i jumps from 1 to 3
// byte_offset[3] = l (U+006C)
// byte_offset[4] = l (U+006C)
// byte_offset[5] = o (U+006F)
```

`range` decodes the UTF-8 for you. `i` is the **byte offset** of the start of the rune, which is why it skips from 1 to 3 (because `é` took 2 bytes at positions 1 and 2). `r` is the full rune (the Unicode code point).

---

### Converting between string, []byte, and []rune

```go
s := "héllo"

b := []byte(s)       // decode string into raw bytes — fast, no allocation for short strings
r := []rune(s)       // decode string into runes (code points) — allocates

s2 := string(b)      // encode bytes back to string
s3 := string(r)      // encode runes back to string

// Single rune ↔ string
r1 := 'é'
fmt.Println(string(r1))    // "é"
fmt.Println([]byte(string(r1))) // [195 169] — the two UTF-8 bytes
```

**When to use each:**

| Need | Use |
|------|-----|
| Raw byte manipulation (network, file) | `[]byte(s)` |
| Character-level operations (count, index, reverse) | `[]rune(s)` |
| Check or print a single character | `rune` |

---

### Rune classification with the `unicode` package

The [`unicode`](https://pkg.go.dev/unicode) package provides functions to classify runes:

```go
package main

import (
    "fmt"
    "unicode"
)

func main() {
    chars := []rune{'A', 'é', '3', ' ', '!', '世'}
    for _, r := range chars {
        fmt.Printf("%c  letter=%-5v digit=%-5v space=%-5v upper=%-5v\n",
            r,
            unicode.IsLetter(r),
            unicode.IsDigit(r),
            unicode.IsSpace(r),
            unicode.IsUpper(r),
        )
    }
}
```

Output:
```
A  letter=true  digit=false space=false upper=true
é  letter=true  digit=false space=false upper=false
3  letter=false digit=true  space=false upper=false
   letter=false digit=false space=true  upper=false
!  letter=false digit=false space=false upper=false
世  letter=true  digit=false space=false upper=false
```

---

### Quick reference

```go
r := '世'

fmt.Println(r)              // 19990            integer value
fmt.Printf("%c\n", r)      // 世               character
fmt.Printf("%U\n", r)      // U+4E16           Unicode code point
fmt.Printf("%d\n", r)      // 19990            decimal
fmt.Println(string(r))     // 世               single-char string
fmt.Println(len(string(r)))// 3                bytes in UTF-8

s := "hello, 世界"
fmt.Println(len(s))          // 13  (7 ASCII + 3 + 3)
fmt.Println(len([]rune(s)))  // 9   (9 characters)
```

---

## Labs

### Lab 1: Arrays — Fixed Size and Value Semantics

**What you'll practise:** Declaring fixed-size arrays, comparing two arrays with `==`, and observing copy semantics when an array is passed to a function.

**Task:**
Declare two `[5]int` arrays, compare them, pass one to a function that modifies it, and confirm the original is unchanged.

**Steps:**
1. Declare `a := [5]int{1, 2, 3, 4, 5}` and `b := [5]int{1, 2, 3, 4, 5}`
2. Compare with `a == b` — arrays are comparable if their element type is
3. Write `zeroFirst(arr [5]int) [5]int` that sets `arr[0] = 0` and returns the modified copy
4. Call it and confirm the original `a` is unchanged
5. Print `[3]int` and `[4]int` are different types — try assigning one to the other

```go
package main

import "fmt"

func zeroFirst(arr [5]int) [5]int {
    arr[0] = 0
    return arr
}

func main() {
    a := [5]int{1, 2, 3, 4, 5}
    b := [5]int{1, 2, 3, 4, 5}
    c := [5]int{9, 2, 3, 4, 5}

    fmt.Printf("a == b: %v\n", a == b)
    fmt.Printf("a == c: %v\n", a == c)

    modified := zeroFirst(a)
    fmt.Printf("original a: %v\n", a)
    fmt.Printf("modified:   %v\n", modified)

    // Arrays with different lengths are different types
    // x := [3]int{1, 2, 3}
    // a = x  // compile error: cannot use [3]int as [5]int
    fmt.Println("Arrays of different lengths are different types — cannot assign")
}
```

**Expected output:**
```
a == b: true
a == c: false
original a: [1 2 3 4 5]
modified:   [0 2 3 4 5]
Arrays of different lengths are different types — cannot assign
```

**Checkpoint:** `a` is unchanged after calling `zeroFirst` — the function received a copy. Uncomment the assignment `a = x` and confirm the compile error message.

---

### Lab 2: Slice Mechanics — len, cap, and Backing Arrays

**What you'll practise:** Creating slices with `make`, understanding `len` vs `cap`, and observing when `append` allocates a new backing array.

**Task:**
Create a slice with `make`, append elements beyond its capacity, and use `%p` to detect when a new backing array is allocated.

**Steps:**
1. Create `s := make([]int, 3, 5)` — length 3, capacity 5
2. Print `len`, `cap`, and the pointer to the first element
3. Append two elements (still within capacity) — pointer should stay the same
4. Append a sixth element (exceeds capacity) — pointer changes, new backing array allocated
5. Show that a sub-slice shares the backing array

```go
package main

import (
    "fmt"
    "unsafe"
)

func header(s []int) {
    if len(s) == 0 {
        fmt.Printf("  len=%-3d cap=%-3d ptr=<empty>\n", len(s), cap(s))
        return
    }
    fmt.Printf("  len=%-3d cap=%-3d ptr=%p\n", len(s), cap(s), unsafe.SliceData(s))
}

func main() {
    s := make([]int, 3, 5)
    fmt.Println("after make([]int, 3, 5):")
    header(s)

    s = append(s, 10, 20) // still within capacity
    fmt.Println("after append x2 (within cap):")
    header(s)

    s = append(s, 30) // exceeds capacity — new allocation
    fmt.Println("after append x1 (exceeds cap — new backing array):")
    header(s)

    // Sub-slice shares backing array
    sub := s[1:3]
    fmt.Println("sub-slice s[1:3]:")
    header(sub)
    sub[0] = 999
    fmt.Printf("s[1] after sub[0]=999: %d (shared!)\n", s[1])
}
```

**Expected output:**
```
after make([]int, 3, 5):
  len=3   cap=5   ptr=0xc000...
after append x2 (within cap):
  len=5   cap=5   ptr=0xc000...   (same pointer)
after append x1 (exceeds cap — new backing array):
  len=6   cap=10  ptr=0xc000...   (different pointer)
sub-slice s[1:3]:
  len=2   cap=9   ptr=0xc000...
s[1] after sub[0]=999: 999 (shared!)
```

**Checkpoint:** The pointer changes on the third append (when cap is exceeded). Modifying `sub[0]` changes `s[1]` — they share the same backing array until one of them appends past its capacity.

---

### Lab 3: Slice Tricks

**What you'll practise:** Using `copy`, building 2D slices, and performing common in-place slice operations (delete, insert).

**Task:**
Implement four classic slice manipulations from scratch, without importing any package beyond `fmt`.

**Steps:**
1. **Deep copy:** use `copy` to clone a slice and prove it is independent
2. **Delete (unordered):** swap the target element with the last, then trim — O(1)
3. **Delete (ordered):** use `append(s[:i], s[i+1:]...)` to preserve order — O(n)
4. **Insert at index:** use `append` twice to insert a value in the middle

```go
package main

import "fmt"

func main() {
    // 1. Deep copy
    original := []int{1, 2, 3, 4, 5}
    clone := make([]int, len(original))
    copy(clone, original)
    clone[0] = 99
    fmt.Printf("original: %v\nclone:    %v\n\n", original, clone)

    // 2. Delete index 2 — unordered (fast)
    s := []int{10, 20, 30, 40, 50}
    i := 2
    s[i] = s[len(s)-1]
    s = s[:len(s)-1]
    fmt.Printf("delete unordered [2]: %v\n", s)

    // 3. Delete index 1 — ordered (preserves order)
    s = []int{10, 20, 30, 40, 50}
    s = append(s[:1], s[2:]...)
    fmt.Printf("delete ordered   [1]: %v\n", s)

    // 4. Insert 99 at index 2
    s = []int{10, 20, 30, 40, 50}
    s = append(s[:2], append([]int{99}, s[2:]...)...)
    fmt.Printf("insert 99 at [2]:     %v\n", s)

    // 5. 2D slice (rows x cols)
    matrix := make([][]int, 3)
    for r := range matrix {
        matrix[r] = make([]int, 4)
        for c := range matrix[r] {
            matrix[r][c] = r*4 + c
        }
    }
    fmt.Println("3x4 matrix:", matrix)
}
```

**Expected output:**
```
original: [1 2 3 4 5]
clone:    [99 2 3 4 5]

delete unordered [2]: [10 20 50 40]
delete ordered   [1]: [10 30 40 50]
insert 99 at [2]:     [10 20 99 30 40 50]
3x4 matrix: [[0 1 2 3] [4 5 6 7] [8 9 10 11]]
```

**Checkpoint:** `clone[0] = 99` does not affect `original[0]`. Unordered delete does not preserve order (50 moves to position 2). Ordered delete does preserve order.

---

### Lab 4: Strings as []rune — Unicode-Safe Iteration

**What you'll practise:** Iterating a UTF-8 string with `range` to get runes (not bytes), counting bytes vs runes, and reversing a string correctly.

**Task:**
Work with a string that contains multi-byte UTF-8 characters, print each rune's byte offset and Unicode value, then reverse the string correctly using `[]rune`.

**Steps:**
1. Use `range` over `"héllo, 世界"` — print each byte offset and rune
2. Compare `len(s)` (bytes) with `len([]rune(s))` (runes)
3. Write `reverseString(s string) string` that converts to `[]rune`, reverses, and converts back
4. Verify the reversal is correct for a multi-byte string

```go
package main

import "fmt"

func reverseString(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

func main() {
    s := "héllo, 世界"

    fmt.Println("range iteration (byte offset → rune):")
    for i, r := range s {
        fmt.Printf("  byte[%2d]: %c  (U+%04X)\n", i, r, r)
    }

    fmt.Printf("\nbyte count (len):  %d\n", len(s))
    fmt.Printf("rune count:        %d\n", len([]rune(s)))

    fmt.Printf("\noriginal: %s\n", s)
    fmt.Printf("reversed: %s\n", reverseString(s))
}
```

**Expected output:**
```
range iteration (byte offset → rune):
  byte[ 0]: h  (U+0068)
  byte[ 1]: é  (U+00E9)
  byte[ 3]: l  (U+006C)
  byte[ 4]: l  (U+006C)
  byte[ 5]: o  (U+006F)
  byte[ 6]: ,  (U+002C)
  byte[ 7]:    (U+0020)
  byte[ 8]: 世 (U+4E16)
  byte[11]: 界 (U+754C)

byte count (len):  13
rune count:        9

original: héllo, 世界
reversed: 界世 ,olléh
```

**Checkpoint:** Byte offset 1 jumps to 3 — `é` occupies bytes 1 and 2 (2-byte UTF-8 sequence). The reversed string is correct. Try reversing `"hello"` (ASCII) — same result as byte reversal.

---

### Lab 5: strings Package — Builder and Common Operations

**What you'll practise:** Using `strings.Builder` for efficient concatenation, and applying `Fields`, `Split`, `Join`, `Replace`, `Contains`, `HasPrefix`, and `TrimSpace`.

**Task:**
Use `strings.Builder` to join 1000 words efficiently, then apply six other `strings` functions to a sample sentence.

**Steps:**
1. Use a `for` loop + `strings.Builder` to concatenate 1000 "word" repetitions — measure with `b.Len()`
2. Apply `strings.Fields`, `strings.Split`, `strings.Join`, `strings.Replace`, `strings.Contains`, `strings.HasPrefix`, `strings.TrimSpace` to the sample sentence
3. Print results

```go
package main

import (
    "fmt"
    "strings"
)

func main() {
    // 1. strings.Builder — efficient concatenation
    var b strings.Builder
    for i := 0; i < 1000; i++ {
        b.WriteString("word")
        if i < 999 {
            b.WriteByte(' ')
        }
    }
    result := b.String()
    fmt.Printf("Built string: %d bytes, starts with: %q\n\n", b.Len(), result[:20])

    // 2. Common operations on a sentence
    sentence := "  The quick brown fox jumps over the lazy dog  "
    trimmed := strings.TrimSpace(sentence)

    fmt.Println("TrimSpace:  ", trimmed)
    fmt.Println("Contains 'fox':", strings.Contains(trimmed, "fox"))
    fmt.Println("HasPrefix 'The':", strings.HasPrefix(trimmed, "The"))
    fmt.Println("Replace fox→cat:", strings.Replace(trimmed, "fox", "cat", 1))

    words := strings.Fields(trimmed)
    fmt.Printf("Fields: %d words, first=%q last=%q\n", len(words), words[0], words[len(words)-1])

    parts := strings.Split(trimmed, " ")
    joined := strings.Join(parts[:3], "-")
    fmt.Println("Join first 3 with '-':", joined)
}
```

**Expected output:**
```
Built string: 4999 bytes, starts with: "word word word word "

TrimSpace:   The quick brown fox jumps over the lazy dog
Contains 'fox': true
HasPrefix 'The': true
Replace fox→cat: The quick brown cat jumps over the lazy dog
Fields: 9 words, first="The" last="dog"
Join first 3 with '-': The-quick-brown
```

**Checkpoint:** Builder length is 4999 (4 chars per word × 1000 + 999 spaces). `strings.Fields` handles multiple consecutive spaces; `strings.Split(s, " ")` does not — try it with double spaces and compare.

---

### Lab 6 (Final): String Statistics Tool

**What you'll practise:** Combining slice iteration, rune handling, maps, and the `strings` package into a complete text analysis program.

**Task:**
Write `day-05/main.go` that analyses a hardcoded multi-line string and reports: rune count, byte count, word count, line count, longest word, and most frequent character.

**Steps:**
1. Hardcode a multi-line string literal using a raw string literal (backticks)
2. Count lines with `strings.Split` on `"\n"`
3. Count words with `strings.Fields`
4. Count runes with `len([]rune(s))`
5. Find the longest word by iterating the word slice
6. Build a `map[rune]int` for character frequency; find the max

```go
package main

import (
    "fmt"
    "strings"
)

const text = `Go is an open source programming language that makes it easy
to build simple, reliable, and efficient software.
The Go programming language was conceived in late 2007.`

func main() {
    // Byte and rune counts
    byteCount := len(text)
    runeCount := len([]rune(text))

    // Lines
    lines := strings.Split(strings.TrimSpace(text), "\n")

    // Words
    words := strings.Fields(text)

    // Longest word
    longest := ""
    for _, w := range words {
        if len(w) > len(longest) {
            longest = w
        }
    }

    // Most frequent rune (ignoring spaces)
    freq := make(map[rune]int)
    for _, r := range text {
        if r != ' ' && r != '\n' {
            freq[r]++
        }
    }
    var topRune rune
    var topCount int
    for r, n := range freq {
        if n > topCount {
            topRune, topCount = r, n
        }
    }

    fmt.Printf("Bytes:          %d\n", byteCount)
    fmt.Printf("Runes:          %d\n", runeCount)
    fmt.Printf("Lines:          %d\n", len(lines))
    fmt.Printf("Words:          %d\n", len(words))
    fmt.Printf("Longest word:   %q (%d chars)\n", longest, len([]rune(longest)))
    fmt.Printf("Top character:  %q (appears %d times)\n", topRune, topCount)
}
```

**Expected output:**
```
Bytes:          152
Runes:          152
Lines:          3
Words:          29
Longest word:   "programming" (11 chars)
Top character:  'e' (appears 10 times)
```

**Checkpoint:** Replace the text with one containing emoji or accented characters — confirm rune count differs from byte count. `go vet .` passes. Run `go test ./...` if you add a test file.

---

## Day Project: String Statistics Tool

Write a program that takes a multi-line string (hardcoded) and computes:
- Total character count (runes)
- Total byte count
- Word count (using [`strings.Fields`](https://pkg.go.dev/strings#Fields))
- Line count
- Longest word
- Most frequent character (use a `map[rune]int`)

**Extension ideas:** read from `os.Stdin`; add sentence count; produce a frequency bar chart in the terminal.

## Official Documentation

- [`strings`](https://pkg.go.dev/strings) — Contains, Fields, Builder, Join, Split, ToUpper, TrimSpace, and more
- [`unicode`](https://pkg.go.dev/unicode) — rune classification (IsLetter, IsDigit, etc.)
- [`fmt`](https://pkg.go.dev/fmt) — formatted I/O with `%c` rune verb
- [`os`](https://pkg.go.dev/os) — `os.Stdin` for reading standard input
- [Language Spec: Slice types](https://go.dev/ref/spec#Slice_types) — slice internals
- [Language Spec: String types](https://go.dev/ref/spec#String_types) — UTF-8 string representation
- [Go Blog: Strings, bytes, runes and characters](https://go.dev/blog/strings) — deep dive on Go strings
- [Go Tour: More types](https://go.dev/tour/moretypes/7) — slices and arrays tour
