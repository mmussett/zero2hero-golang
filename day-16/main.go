package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
)

type FileHash struct {
	Path string
	Hash string
	Err  error
}

func hashFile(path string) FileHash {
	f, err := os.Open(path)
	if err != nil {
		return FileHash{Path: path, Err: err}
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return FileHash{Path: path, Err: err}
	}
	return FileHash{Path: path, Hash: fmt.Sprintf("%x", h.Sum(nil))}
}

func hashDir(dir string) ([]FileHash, error) {
	var paths []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sem := make(chan struct{}, runtime.NumCPU())
	results := make(chan FileHash, len(paths))
	var wg sync.WaitGroup

	for _, p := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results <- hashFile(path)
		}(p)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var hashes []FileHash
	for h := range results {
		hashes = append(hashes, h)
	}
	sort.Slice(hashes, func(i, j int) bool {
		return hashes[i].Path < hashes[j].Path
	})
	return hashes, nil
}

func main() {
	dir, err := os.MkdirTemp("", "day15-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	files := map[string]string{
		"hello.txt":  "Hello, Go!",
		"world.txt":  "Hello, World!",
		"readme.md":  "# Zero to Hero: Go",
		"config.txt": "host=localhost\nport=8080",
	}
	for name, content := range files {
		os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
	}

	fmt.Printf("Hashing files in %s using %d workers...\n\n", dir, runtime.NumCPU())

	hashes, err := hashDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%-20s  %s\n", "File", "SHA-256 (first 16 hex chars)")
	fmt.Println("────────────────────────────────────────────────────────────")
	for _, h := range hashes {
		name := filepath.Base(h.Path)
		if h.Err != nil {
			fmt.Printf("%-20s  ERROR: %v\n", name, h.Err)
		} else {
			fmt.Printf("%-20s  %s...\n", name, h.Hash[:16])
		}
	}
}
