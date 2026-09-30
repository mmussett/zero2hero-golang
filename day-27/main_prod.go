//go:build prod

// Package main (prod variant) — assets are baked into the binary via
// //go:embed so no source tree is needed at runtime.
//
// Build:  go build -tags prod -o server-prod .
// Run:    go run -tags prod .
package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
)

//go:embed assets/*
var embeddedAssets embed.FS

// openFS returns the embedded filesystem rooted at "assets/".
// This is the prod-mode implementation (build tag: prod).
func openFS() (fs.FS, error) {
	fmt.Println("[prod] serving assets from embedded binary")
	sub, err := fs.Sub(embeddedAssets, "assets")
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	fsys, err := openFS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "open assets: %v\n", err)
		os.Exit(1)
	}

	// Demonstrate fs.ReadFile — read index.html from the embedded FS.
	data, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read index.html: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("index.html is %d bytes (embedded)\n", len(data))

	// List embedded files.
	entries, _ := fs.ReadDir(fsys, ".")
	fmt.Println("assets:")
	for _, e := range entries {
		info, _ := e.Info()
		fmt.Printf("  %-20s  %d bytes\n", e.Name(), info.Size())
	}

	// Serve all assets at /assets/ and index.html at /.
	mux := http.NewServeMux()

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(fsys))))

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
