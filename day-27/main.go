//go:build !prod

// Package main demonstrates //go:embed for baking static assets into a Go
// binary, build tags to switch between embedded and disk-based serving, and
// fs.ReadFile for reading embedded files at startup.
//
// Build modes:
//
//	go run .               # dev mode — reads assets from disk (this file)
//	go run -tags prod .    # prod mode — assets baked into the binary
//	go build -tags prod -o server-prod .
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
)

// openFS returns a filesystem rooted at the local "assets" directory.
// This is the dev-mode implementation (build tag: !prod).
// With -tags prod the compiler picks assets_prod.go instead, which returns
// an embed.FS so no source tree is needed at runtime.
func openFS() (fs.FS, error) {
	fmt.Println("[dev] serving assets from disk")
	return os.DirFS("assets"), nil
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	fsys, err := openFS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "open assets: %v\n", err)
		os.Exit(1)
	}

	// Demonstrate fs.ReadFile — read index.html and print the first line.
	data, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read index.html: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("index.html is %d bytes\n", len(data))

	// List embedded files.
	entries, _ := fs.ReadDir(fsys, ".")
	fmt.Println("assets:")
	for _, e := range entries {
		info, _ := e.Info()
		fmt.Printf("  %-20s  %d bytes\n", e.Name(), info.Size())
	}

	// Serve all assets at /assets/ and index.html at /.
	mux := http.NewServeMux()

	// Serve the whole FS under /assets/.
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(fsys))))

	// Serve index.html at the root.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		content, err := fs.ReadFile(fsys, "index.html")
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content) //nolint:errcheck
	})

	fmt.Printf("listening on %s\n", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}
