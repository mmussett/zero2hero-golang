package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Entry represents a stored password entry.
type Entry struct {
	Name      string    `json:"name"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

// Store is the full password store loaded from disk.
type Store struct {
	Entries []Entry `json:"entries"`
}

func storeFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".passgen.json"
	}
	return filepath.Join(home, ".passgen.json")
}

func loadStore() (*Store, error) {
	path := storeFile()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Store{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading store: %w", err)
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing store: %w", err)
	}
	return &s, nil
}

func saveStore(s *Store) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding store: %w", err)
	}
	return os.WriteFile(storeFile(), data, 0600)
}

const (
	upperChars     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerChars     = "abcdefghijklmnopqrstuvwxyz"
	digitChars     = "0123456789"
	symbolChars    = "!@#$%^&*()-_=+[]{}|;:,.<>?"
	ambiguousChars = "0O1lI"
)

func generatePassword(length int, upper, lower, digits, symbols, excludeAmbiguous bool) (string, error) {
	if !upper && !lower && !digits && !symbols {
		// Default: all character sets
		upper, lower, digits, symbols = true, true, true, true
	}

	var charset strings.Builder
	if upper {
		charset.WriteString(upperChars)
	}
	if lower {
		charset.WriteString(lowerChars)
	}
	if digits {
		charset.WriteString(digitChars)
	}
	if symbols {
		charset.WriteString(symbolChars)
	}

	chars := charset.String()
	if excludeAmbiguous {
		var filtered strings.Builder
		for _, c := range chars {
			if !strings.ContainsRune(ambiguousChars, c) {
				filtered.WriteRune(c)
			}
		}
		chars = filtered.String()
	}

	if len(chars) == 0 {
		return "", fmt.Errorf("no characters available with given constraints")
	}

	var password strings.Builder
	n := big.NewInt(int64(len(chars)))
	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, n)
		if err != nil {
			return "", fmt.Errorf("crypto/rand error: %w", err)
		}
		password.WriteByte(chars[idx.Int64()])
	}
	return password.String(), nil
}

func cmdGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	length := fs.Int("length", 20, "password length")
	upper := fs.Bool("upper", false, "include uppercase letters")
	lower := fs.Bool("lower", false, "include lowercase letters")
	digits := fs.Bool("digits", false, "include digits")
	symbols := fs.Bool("symbols", false, "include symbols")
	excludeAmbiguous := fs.Bool("exclude-ambiguous", false, "exclude ambiguous characters (0O1lI)")
	fs.Parse(args)

	pw, err := generatePassword(*length, *upper, *lower, *digits, *symbols, *excludeAmbiguous)
	if err != nil {
		fmt.Fprintf(os.Stderr, "passgen: generate error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(pw)
}

func cmdAdd(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "passgen: add requires a name\n")
		os.Exit(1)
	}
	name := args[0]

	fs := flag.NewFlagSet("add", flag.ExitOnError)
	length := fs.Int("length", 20, "password length")
	upper := fs.Bool("upper", false, "include uppercase letters")
	lower := fs.Bool("lower", false, "include lowercase letters")
	digits := fs.Bool("digits", false, "include digits")
	symbols := fs.Bool("symbols", false, "include symbols")
	excludeAmbiguous := fs.Bool("exclude-ambiguous", false, "exclude ambiguous characters")
	fs.Parse(args[1:])

	pw, err := generatePassword(*length, *upper, *lower, *digits, *symbols, *excludeAmbiguous)
	if err != nil {
		fmt.Fprintf(os.Stderr, "passgen: generate error: %v\n", err)
		os.Exit(1)
	}

	store, err := loadStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "passgen: %v\n", err)
		os.Exit(1)
	}

	// Replace if exists
	for i, e := range store.Entries {
		if e.Name == name {
			store.Entries[i] = Entry{Name: name, Password: pw, CreatedAt: time.Now()}
			if err := saveStore(store); err != nil {
				fmt.Fprintf(os.Stderr, "passgen: save error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Updated %q: %s\n", name, pw)
			return
		}
	}

	store.Entries = append(store.Entries, Entry{
		Name:      name,
		Password:  pw,
		CreatedAt: time.Now(),
	})
	if err := saveStore(store); err != nil {
		fmt.Fprintf(os.Stderr, "passgen: save error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Added %q: %s\n", name, pw)
}

func cmdGet(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "passgen: get requires a name\n")
		os.Exit(1)
	}
	name := args[0]

	store, err := loadStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "passgen: %v\n", err)
		os.Exit(1)
	}

	for _, e := range store.Entries {
		if e.Name == name {
			fmt.Printf("Name:       %s\n", e.Name)
			fmt.Printf("Password:   %s\n", e.Password)
			fmt.Printf("Created at: %s\n", e.CreatedAt.Format(time.RFC3339))
			return
		}
	}
	fmt.Fprintf(os.Stderr, "passgen: %q not found\n", name)
	os.Exit(1)
}

func cmdList(_ []string) {
	store, err := loadStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "passgen: %v\n", err)
		os.Exit(1)
	}

	if len(store.Entries) == 0 {
		fmt.Println("No passwords stored.")
		return
	}

	fmt.Printf("%-20s  %-30s  %s\n", "NAME", "CREATED AT", "PASSWORD")
	fmt.Println(strings.Repeat("-", 80))
	for _, e := range store.Entries {
		fmt.Printf("%-20s  %-30s  %s\n", e.Name, e.CreatedAt.Format(time.RFC3339), e.Password)
	}
}

func cmdDelete(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "passgen: delete requires a name\n")
		os.Exit(1)
	}
	name := args[0]

	store, err := loadStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "passgen: %v\n", err)
		os.Exit(1)
	}

	newEntries := store.Entries[:0]
	found := false
	for _, e := range store.Entries {
		if e.Name == name {
			found = true
			continue
		}
		newEntries = append(newEntries, e)
	}

	if !found {
		fmt.Fprintf(os.Stderr, "passgen: %q not found\n", name)
		os.Exit(1)
	}

	store.Entries = newEntries
	if err := saveStore(store); err != nil {
		fmt.Fprintf(os.Stderr, "passgen: save error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Deleted %q\n", name)
}

func usage() {
	fmt.Fprintf(os.Stderr, `passgen - password generator and manager

Usage:
  passgen generate [--length=N] [--upper] [--lower] [--digits] [--symbols] [--exclude-ambiguous]
  passgen add <name> [--length=N] [--upper] [--lower] [--digits] [--symbols] [--exclude-ambiguous]
  passgen get <name>
  passgen list
  passgen delete <name>

Commands:
  generate   Generate and print a password (not stored)
  add        Generate a password and store it under <name>
  get        Retrieve a stored password by name
  list       List all stored password entries
  delete     Remove a stored password by name

Passwords are stored in %s
`, storeFile())
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	rest := os.Args[2:]

	switch cmd {
	case "generate":
		cmdGenerate(rest)
	case "add":
		cmdAdd(rest)
	case "get":
		cmdGet(rest)
	case "list":
		cmdList(rest)
	case "delete":
		cmdDelete(rest)
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "passgen: unknown command %q\n", cmd)
		usage()
		os.Exit(2)
	}
}
