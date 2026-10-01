// Package notes defines the Note domain type, the Store interface, and an
// in-memory Store implementation.
package notes

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrNotFound is returned when a note with the requested ID does not exist.
var ErrNotFound = errors.New("note not found")

// Note is the core domain object.
type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store is the persistence abstraction for Notes.
type Store interface {
	List() ([]Note, error)
	Get(id int) (Note, error)
	Create(title, body string) (Note, error)
	Update(id int, title, body string) (Note, error)
	Delete(id int) error
}

// ── In-memory implementation ─────────────────────────────────────────────────

// MemoryStore is a thread-safe, in-process Store backed by a plain map.
type MemoryStore struct {
	mu     sync.RWMutex
	notes  map[int]Note
	nextID int
}

// NewMemoryStore returns an initialised, empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		notes:  make(map[int]Note),
		nextID: 1,
	}
}

func (m *MemoryStore) List() ([]Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	notes := make([]Note, 0, len(m.notes))
	for _, n := range m.notes {
		notes = append(notes, n)
	}
	return notes, nil
}

func (m *MemoryStore) Get(id int) (Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.notes[id]
	if !ok {
		return Note{}, fmt.Errorf("note %d: %w", id, ErrNotFound)
	}
	return n, nil
}

func (m *MemoryStore) Create(title, body string) (Note, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	n := Note{
		ID:        m.nextID,
		Title:     title,
		Body:      body,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.notes[m.nextID] = n
	m.nextID++
	return n, nil
}

func (m *MemoryStore) Update(id int, title, body string) (Note, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.notes[id]
	if !ok {
		return Note{}, fmt.Errorf("note %d: %w", id, ErrNotFound)
	}
	n.Title = title
	n.Body = body
	n.UpdatedAt = time.Now()
	m.notes[id] = n
	return n, nil
}

func (m *MemoryStore) Delete(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.notes[id]; !ok {
		return fmt.Errorf("note %d: %w", id, ErrNotFound)
	}
	delete(m.notes, id)
	return nil
}
