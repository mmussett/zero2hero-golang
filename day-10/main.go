package main

import (
	"errors"
	"fmt"
	"strings"
)

var ErrEmptyField = errors.New("empty required field")

type ParseError struct {
	Row    int
	Col    int
	Field  string
	Reason string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at row %d, col %d (field %q): %s", e.Row, e.Col, e.Field, e.Reason)
}

func (e *ParseError) Unwrap() error {
	if e.Reason == "empty required field" {
		return ErrEmptyField
	}
	return nil
}

func parseRow(line string, lineNum int) ([]string, error) {
	fields := strings.Split(line, ",")
	required := []string{"name", "email", "age"}

	if len(fields) != len(required) {
		return nil, &ParseError{
			Row:    lineNum,
			Col:    0,
			Field:  "row",
			Reason: fmt.Sprintf("expected %d fields, got %d", len(required), len(fields)),
		}
	}

	result := make([]string, len(fields))
	for i, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			return nil, &ParseError{
				Row:    lineNum,
				Col:    i + 1,
				Field:  required[i],
				Reason: "empty required field",
			}
		}
		result[i] = f
	}
	return result, nil
}

func main() {
	csvData := []string{
		"Alice,alice@example.com,30",
		"Bob,bob@example.com,25",
		",,",
		"Carol,carol@example.com",
		"Dave,,40",
	}

	for i, line := range csvData {
		fields, err := parseRow(line, i+1)
		if err != nil {
			fmt.Printf("Row %d error: %v\n", i+1, err)

			if errors.Is(err, ErrEmptyField) {
				fmt.Printf("  → caused by: empty required field\n")
			}

			var pe *ParseError
			if errors.As(err, &pe) {
				fmt.Printf("  → at column %d, field %q\n", pe.Col, pe.Field)
			}
			continue
		}
		fmt.Printf("Row %d OK: name=%q email=%q age=%q\n", i+1, fields[0], fields[1], fields[2])
	}
}
