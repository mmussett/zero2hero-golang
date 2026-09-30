// Package main implements a Bitcask-inspired persistent key-value store.
//
// On-disk format (append-only log):
//
//	[4-byte key length (LE)] [4-byte value length (LE)] [key bytes] [value bytes]
//
// A value length of 0xFFFFFFFF is a tombstone (deleted key).
// The in-memory index maps each live key to the byte offset of its value in
// the log file and the length of that value, enabling O(1) random-access reads.
//
// Commands:
//
//	kvdb [--db path] set key value
//	kvdb [--db path] get key
//	kvdb [--db path] delete key
//	kvdb [--db path] list
//	kvdb [--db path] compact
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
)

// ---------------------------------------------------------------------------
// DB
// ---------------------------------------------------------------------------

const tombstone = uint32(0xFFFFFFFF)

// indexEntry records where a key's value lives in the log file.
type indexEntry struct {
	valueOffset int64 // byte offset of the first value byte in the file
	valueSize   int64 // length of the value in bytes
}

// DB is a Bitcask-style key-value store backed by an append-only log file.
type DB struct {
	path  string
	file  *os.File
	index map[string]indexEntry
	mu    sync.Mutex
}

// Open opens (or creates) the database at path and replays the log to rebuild
// the in-memory index.
func Open(path string) (*DB, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	db := &DB{
		path:  path,
		file:  f,
		index: make(map[string]indexEntry),
	}
	if err := db.replay(); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("replay %s: %w", path, err)
	}
	return db, nil
}

// Close flushes and closes the underlying file.
func (db *DB) Close() error {
	return db.file.Close()
}

// replay reads the log from the beginning and rebuilds db.index.
func (db *DB) replay() error {
	if _, err := db.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	var offset int64
	header := make([]byte, 8)

	for {
		// Read 8-byte header: [4 keyLen][4 valLen]
		if _, err := io.ReadFull(db.file, header); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return err
		}

		keyLen := binary.LittleEndian.Uint32(header[0:4])
		valLen := binary.LittleEndian.Uint32(header[4:8])

		// Read key bytes.
		key := make([]byte, keyLen)
		if _, err := io.ReadFull(db.file, key); err != nil {
			break // truncated entry — stop here
		}

		offset += 8 + int64(keyLen) // now pointing at start of value (or next entry for tombstone)

		if valLen == tombstone {
			delete(db.index, string(key))
		} else {
			// Record where the value lives.
			db.index[string(key)] = indexEntry{
				valueOffset: offset,
				valueSize:   int64(valLen),
			}
			// Seek past the value bytes.
			if _, err := db.file.Seek(int64(valLen), io.SeekCurrent); err != nil {
				break
			}
			offset += int64(valLen)
		}
	}
	return nil
}

// Set stores key → value.  Appends an entry to the log and updates the index.
func (db *DB) Set(key, value string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	keyB := []byte(key)
	valB := []byte(value)

	// Seek to end before appending.
	endOffset, err := db.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[0:4], uint32(len(keyB)))
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(valB)))

	if _, err := db.file.Write(header); err != nil {
		return err
	}
	if _, err := db.file.Write(keyB); err != nil {
		return err
	}
	valueOffset := endOffset + 8 + int64(len(keyB))
	if _, err := db.file.Write(valB); err != nil {
		return err
	}

	db.index[key] = indexEntry{valueOffset: valueOffset, valueSize: int64(len(valB))}
	return nil
}

// Get retrieves the value for key.  Returns ("", false) when the key does not exist.
func (db *DB) Get(key string) (string, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	ie, ok := db.index[key]
	if !ok {
		return "", false
	}
	buf := make([]byte, ie.valueSize)
	if _, err := db.file.ReadAt(buf, ie.valueOffset); err != nil {
		return "", false
	}
	return string(buf), true
}

// Delete marks key as deleted by appending a tombstone entry to the log.
func (db *DB) Delete(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.index[key]; !ok {
		return fmt.Errorf("key not found: %s", key)
	}

	keyB := []byte(key)

	if _, err := db.file.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[0:4], uint32(len(keyB)))
	binary.LittleEndian.PutUint32(header[4:8], tombstone)

	if _, err := db.file.Write(header); err != nil {
		return err
	}
	if _, err := db.file.Write(keyB); err != nil {
		return err
	}

	delete(db.index, key)
	return nil
}

// Keys returns all live keys in sorted order.
func (db *DB) Keys() []string {
	db.mu.Lock()
	defer db.mu.Unlock()

	keys := make([]string, 0, len(db.index))
	for k := range db.index {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Compact rewrites the log file keeping only the latest value per key, then
// replaces the original file with the compacted version.
func (db *DB) Compact() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tmpPath := db.path + ".tmp"
	tmp, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	newIndex := make(map[string]indexEntry, len(db.index))
	var writeOffset int64

	header := make([]byte, 8)
	for key, ie := range db.index {
		valB := make([]byte, ie.valueSize)
		if _, err := db.file.ReadAt(valB, ie.valueOffset); err != nil {
			tmp.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("read value for %q: %w", key, err)
		}
		keyB := []byte(key)
		binary.LittleEndian.PutUint32(header[0:4], uint32(len(keyB)))
		binary.LittleEndian.PutUint32(header[4:8], uint32(len(valB)))
		if _, err := tmp.Write(header); err != nil {
			tmp.Close()
			os.Remove(tmpPath)
			return err
		}
		if _, err := tmp.Write(keyB); err != nil {
			tmp.Close()
			os.Remove(tmpPath)
			return err
		}
		valueOffset := writeOffset + 8 + int64(len(keyB))
		if _, err := tmp.Write(valB); err != nil {
			tmp.Close()
			os.Remove(tmpPath)
			return err
		}
		newIndex[key] = indexEntry{valueOffset: valueOffset, valueSize: int64(len(valB))}
		writeOffset = valueOffset + int64(len(valB))
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Close original before rename (required on Windows).
	if err := db.file.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Replace original with compacted file.
	if err := os.Rename(tmpPath, db.path); err != nil {
		// Rename may fail on Windows if the destination is still locked by
		// another handle; fall back to remove-then-rename.
		_ = os.Remove(db.path)
		if err2 := os.Rename(tmpPath, db.path); err2 != nil {
			return fmt.Errorf("rename: %w (original: %v)", err2, err)
		}
	}

	// Reopen the compacted file.
	db.file, err = os.OpenFile(db.path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("reopen after compact: %w", err)
	}
	db.index = newIndex
	return nil
}

// ---------------------------------------------------------------------------
// CLI
// ---------------------------------------------------------------------------

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: kvdb [--db path] <command> [args]

Commands:
  set    <key> <value>   store a key-value pair
  get    <key>           retrieve a value
  delete <key>           delete a key
  list                   list all keys
  compact                rewrite the log keeping only the latest values

Flags:
  --db path   path to database file (default: kvdb.dat)`)
}

func main() {
	// Manual flag parsing to support "--db" before the subcommand.
	dbPath := "kvdb.dat"
	args := os.Args[1:]

	// Pull --db / --db=... out of args.
	var rest []string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--db" && i+1 < len(args):
			dbPath = args[i+1]
			i++
		case len(args[i]) > 5 && args[i][:5] == "--db=":
			dbPath = args[i][5:]
		default:
			rest = append(rest, args[i])
		}
	}

	if len(rest) == 0 {
		usage()
		os.Exit(1)
	}

	db, err := Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	switch cmd := rest[0]; cmd {
	case "set":
		if len(rest) != 3 {
			fmt.Fprintln(os.Stderr, "usage: kvdb set <key> <value>")
			os.Exit(1)
		}
		if err := db.Set(rest[1], rest[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("OK")

	case "get":
		if len(rest) != 2 {
			fmt.Fprintln(os.Stderr, "usage: kvdb get <key>")
			os.Exit(1)
		}
		val, ok := db.Get(rest[1])
		if !ok {
			fmt.Fprintf(os.Stderr, "key not found: %s\n", rest[1])
			os.Exit(1)
		}
		fmt.Println(val)

	case "delete":
		if len(rest) != 2 {
			fmt.Fprintln(os.Stderr, "usage: kvdb delete <key>")
			os.Exit(1)
		}
		if err := db.Delete(rest[1]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("OK")

	case "list":
		keys := db.Keys()
		if len(keys) == 0 {
			fmt.Println("(empty)")
		} else {
			for _, k := range keys {
				fmt.Println(k)
			}
		}

	case "compact":
		if err := db.Compact(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("compacted")

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}
