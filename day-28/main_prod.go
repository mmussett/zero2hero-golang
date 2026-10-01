//go:build prod

// main_prod.go — prod-mode file openers (Lab 13).
//
// Build with -tags prod to bake all assets and templates into the binary.
// The resulting executable needs no external files at runtime.
//
//	go build -tags prod -o bookstore-prod .
//	go run -tags prod .

package main

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed assets
var embeddedAssets embed.FS

//go:embed templates
var embeddedTemplates embed.FS

// openAssets returns the embedded "assets/" tree stripped of its top-level
// directory prefix, so callers can read files as "banner.txt" not "assets/banner.txt".
func openAssets() (fs.FS, error) {
	fmt.Println("[prod] serving assets from embedded binary")
	return fs.Sub(embeddedAssets, "assets")
}

// openTemplates returns the embedded "templates/" tree stripped of its
// top-level directory prefix, so callers can glob "*.html" directly.
func openTemplates() (fs.FS, error) {
	fmt.Println("[prod] loading templates from embedded binary")
	return fs.Sub(embeddedTemplates, "templates")
}
