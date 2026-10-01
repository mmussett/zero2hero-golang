//go:build !prod

// templates_dev.go — dev-mode file openers (Lab 13).
//
// When built without -tags prod, templates and assets are served directly
// from the source tree so edits take effect on the next restart without
// requiring a rebuild.

package main

import (
	"fmt"
	"io/fs"
	"os"
)

// openAssets returns a filesystem rooted at the local "assets/" directory.
// This is the dev-mode implementation; the prod variant (main_prod.go) returns
// an embed.FS so the binary is self-contained.
func openAssets() (fs.FS, error) {
	fmt.Println("[dev] serving assets from disk")
	return os.DirFS("assets"), nil
}

// openTemplates returns a filesystem rooted at the local "templates/" directory.
// In dev mode, templates are read fresh from disk on every server restart.
func openTemplates() (fs.FS, error) {
	fmt.Println("[dev] loading templates from disk")
	return os.DirFS("templates"), nil
}
