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
		return Contact{}, fmt.Errorf("invalid email %q: must contain '@'", email)
	}
	if name == "" {
		return Contact{}, fmt.Errorf("name cannot be empty")
	}
	return Contact{Name: name, Email: email, Phone: phone}, nil
}

func (c Contact) String() string {
	return fmt.Sprintf("%-20s %-30s %s", c.Name, c.Email, c.Phone)
}

func PrintCard(c Contact) {
	fmt.Println("┌─────────────────────────────────────────┐")
	fmt.Printf("│ Name:  %-33s│\n", c.Name)
	fmt.Printf("│ Email: %-33s│\n", c.Email)
	fmt.Printf("│ Phone: %-33s│\n", c.Phone)
	fmt.Println("└─────────────────────────────────────────┘")
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

	entries := []struct{ name, email, phone string }{
		{"Alice Smith", "alice@example.com", "+1-555-0100"},
		{"Bob Jones", "bob@example.com", "+1-555-0101"},
		{"Carol White", "carol@example.com", "+1-555-0102"},
		{"invalid", "notanemail", ""},
	}

	for _, e := range entries {
		c, err := NewContact(e.name, e.email, e.phone)
		if err != nil {
			fmt.Printf("Error creating contact: %v\n", err)
			continue
		}
		contacts = append(contacts, c)
		PrintCard(c)
	}

	fmt.Println("\nSearching for 'bob jones':")
	if c, ok := FindByName(contacts, "bob jones"); ok {
		fmt.Println("Found:", c)
	} else {
		fmt.Println("Not found")
	}
}
