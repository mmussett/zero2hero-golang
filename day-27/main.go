// Day 27 — Final Project: Go Book Store
//
// A self-contained web application that demonstrates //go:embed, html/template
// with FuncMap, build tags for dev/prod switching, and //go:generate for build
// metadata.
//
// Build modes:
//
//	go run .                  # dev  — templates and assets read from disk
//	go run -tags prod .       # prod — everything embedded in the binary
//	go build -tags prod -o bookstore-prod .
//	go generate ./...         # refresh internal/version/version.go
//
//go:generate go run gen/version.go

package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/mmussett/zero2hero-golang/day-27/internal/version"
)

// Product is a single item in the book catalogue.
type Product struct {
	Name    string
	Price   float64
	InStock bool
}

// catalog is the in-memory product database for the store.
var catalog = []Product{
	{"The Go Programming Language", 49.99, true},
	{"Learning Go, 2nd Edition", 39.99, true},
	{"Go in Action", 44.99, false},
	{"Concurrency in Go", 34.99, true},
	{"Go Web Programming", 29.99, false},
	{"Cloud Native Go", 54.99, true},
}

// pageData holds every value the index.html template needs.
type pageData struct {
	BuildTime string
	Query     string
	Products  []Product
}

func main() {
	addr := flag.String("addr", ":8080", "TCP address to listen on")
	flag.Parse()

	// Open assets and templates from disk (dev) or embedded binary (prod).
	assets, err := openAssets()
	if err != nil {
		fmt.Fprintf(os.Stderr, "open assets: %v\n", err)
		os.Exit(1)
	}

	// Print the ASCII-art banner stored in assets/banner.txt.
	if data, readErr := fs.ReadFile(assets, "banner.txt"); readErr == nil {
		fmt.Print(string(data))
	}

	fmt.Printf("build time : %s\n", version.BuildTime)

	tmplFS, err := openTemplates()
	if err != nil {
		fmt.Fprintf(os.Stderr, "open templates: %v\n", err)
		os.Exit(1)
	}

	// Register custom template functions, then parse all *.html files.
	tmpl := template.Must(
		template.New("").Funcs(funcMap()).ParseFS(tmplFS, "*.html"),
	)

	mux := http.NewServeMux()

	// Serve static assets at /assets/: CSS, JS, images.
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))

	// Serve the dynamic product-catalogue page at /.
	mux.HandleFunc("/", indexHandler(tmpl))

	fmt.Printf("listening on http://localhost%s\n", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}

// indexHandler returns an HTTP handler that renders the product catalogue.
// It reads the optional ?q= query parameter and filters products by name.
func indexHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		data := pageData{
			BuildTime: version.BuildTime,
			Query:     q,
			Products:  filterProducts(catalog, q),
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// filterProducts returns catalogue entries whose Name contains q (case-insensitive).
// An empty q returns all entries unchanged.
func filterProducts(products []Product, q string) []Product {
	if q == "" {
		return products
	}
	q = strings.ToLower(q)
	out := make([]Product, 0, len(products))
	for _, p := range products {
		if strings.Contains(strings.ToLower(p.Name), q) {
			out = append(out, p)
		}
	}
	return out
}

// funcMap returns the custom template functions available in index.html:
//
//   - upper       — strings.ToUpper
//   - formatPrice — format float64 as "XX.XX"
//   - join        — strings.Join (for slice-to-string rendering)
func funcMap() template.FuncMap {
	return template.FuncMap{
		"upper":       strings.ToUpper,
		"formatPrice": func(f float64) string { return fmt.Sprintf("%.2f", f) },
		"join":        strings.Join,
	}
}
