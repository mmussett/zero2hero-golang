package main

import (
	"fmt"
	"reflect"
	"strings"
)

// User is the demo struct used throughout this day's examples.
type User struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email,omitempty"`
}

// describe prints the type name, kind, and all fields (name / type / tag / value)
// of any struct value passed as an interface{}.
func describe(v interface{}) {
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)

	// Dereference pointer if needed
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}

	fmt.Printf("Type: %s\n", t.Name())
	fmt.Printf("Kind: %s\n", t.Kind())

	if t.Kind() != reflect.Struct {
		fmt.Printf("Value: %v\n", val)
		return
	}

	fmt.Printf("Fields (%d):\n", t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := val.Field(i)
		fmt.Printf("  %-12s  type:%-10s  tag:%-30q  value:%v\n",
			field.Name,
			field.Type.Name(),
			string(field.Tag),
			fieldVal.Interface(),
		)
	}
}

// DiffStructs returns the names of exported fields whose values differ between
// two struct values of the same type. Both a and b must be structs (or pointers
// to structs) of the same underlying type.
func DiffStructs(a, b interface{}) []string {
	ta := reflect.TypeOf(a)
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if ta.Kind() == reflect.Ptr {
		ta = ta.Elem()
		va = va.Elem()
		vb = vb.Elem()
	}

	var diffs []string
	for i := 0; i < ta.NumField(); i++ {
		field := ta.Field(i)
		if !field.IsExported() {
			continue
		}
		fa := va.Field(i).Interface()
		fb := vb.Field(i).Interface()
		if !reflect.DeepEqual(fa, fb) {
			diffs = append(diffs, field.Name)
		}
	}
	return diffs
}

// CopyFields copies exported fields from src to dst by name, only when the
// field type matches in both structs. dst must be a pointer to a struct.
func CopyFields(dst, src interface{}) {
	srcType := reflect.TypeOf(src)
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	if srcType.Kind() == reflect.Ptr {
		srcType = srcType.Elem()
		srcVal = srcVal.Elem()
	}
	if dstVal.Kind() != reflect.Ptr {
		panic("CopyFields: dst must be a pointer")
	}
	dstVal = dstVal.Elem()
	dstType := dstVal.Type()

	for i := 0; i < srcType.NumField(); i++ {
		srcField := srcType.Field(i)
		if !srcField.IsExported() {
			continue
		}
		dstField, ok := dstType.FieldByName(srcField.Name)
		if !ok {
			continue
		}
		if dstField.Type != srcField.Type {
			continue
		}
		dstVal.FieldByName(srcField.Name).Set(srcVal.Field(i))
	}
}

func main() {
	fmt.Println("=== Day 28: Reflection ===")
	fmt.Println()

	// ── describe ─────────────────────────────────────────────────────────────
	fmt.Println("--- describe(User{...}) ---")
	u := User{Name: "Alice", Age: 30, Email: "alice@example.com"}
	describe(u)

	fmt.Println()
	fmt.Println("--- describe(&User{...}) (pointer) ---")
	describe(&u)

	// ── DiffStructs ──────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Println("--- DiffStructs ---")
	u1 := User{Name: "Alice", Age: 30, Email: "alice@example.com"}
	u2 := User{Name: "Alice", Age: 31, Email: "alice@new.com"}
	diffs := DiffStructs(u1, u2)
	fmt.Printf("Fields that differ: %s\n", strings.Join(diffs, ", "))

	u3 := User{Name: "Alice", Age: 30, Email: "alice@example.com"}
	diffs2 := DiffStructs(u1, u3)
	if len(diffs2) == 0 {
		fmt.Println("No differences (identical structs)")
	}

	// ── CopyFields ───────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Println("--- CopyFields ---")
	src := User{Name: "Bob", Age: 25, Email: "bob@example.com"}
	var dst User
	CopyFields(&dst, src)
	fmt.Printf("After CopyFields: %+v\n", dst)

	// Partial copy – only Name and Age exist in target
	type Partial struct {
		Name string
		Age  int
	}
	var p Partial
	CopyFields(&p, src)
	fmt.Printf("After CopyFields into Partial: %+v\n", p)

	// ── reflect on a non-struct ───────────────────────────────────────────────
	fmt.Println()
	fmt.Println("--- describe(42) ---")
	describe(42)
}
