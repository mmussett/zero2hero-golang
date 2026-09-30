# Day 27: Embedding, Build Tags, and go generate

## //go:embed

Embed files or directories into the binary at compile time — no runtime file paths needed.

```go
import "embed"

//go:embed static/index.html
var indexHTML []byte

//go:embed templates/*.html
var templates embed.FS

//go:embed config/defaults.json
var defaultConfig []byte
```

Serve embedded files over HTTP:

```go
//go:embed static
var staticFiles embed.FS

sub, _ := fs.Sub(staticFiles, "static")
http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
```

## Build Tags

Conditionally compile files:

```go
//go:build linux

package platform

func OpenFile(path string) (*os.File, error) { /* linux version */ }
```

```go
//go:build !linux

package platform

func OpenFile(path string) (*os.File, error) { /* fallback */ }
```

Build tag syntax (Go 1.17+): `//go:build expr` using `&&`, `||`, `!`, and named tags.

```bash
go build -tags integration ./...      # include //go:build integration files
GOOS=linux GOARCH=amd64 go build .    # cross-compile for Linux/amd64
```

## go:generate

`//go:generate` embeds shell commands in source — run with `go generate`:

```go
//go:generate go run gen/gen.go
//go:generate stringer -type=Status
//go:generate mockgen -destination=mock_store.go -package=main . Store
```

```bash
go generate ./...  # runs all //go:generate directives in the package tree
```

## Cross-Compilation

```bash
GOOS=linux   GOARCH=amd64  go build -o app-linux-amd64   .
GOOS=darwin  GOARCH=arm64  go build -o app-darwin-arm64  .
GOOS=windows GOARCH=amd64  go build -o app-windows.exe   .
```

No cross-compiler toolchain needed — Go ships support for all targets.

## Day Project: Embedded Assets Server

Build a self-contained HTTP server that:
1. Embeds `index.html` and a CSS file from `static/` via `//go:embed`
2. Serves them at `/` and `/static/` — works without the source tree present
3. Uses build tags to switch between `dev` mode (reads from disk, instant reload) and `prod` mode (reads from `embed.FS`)
4. Has a `//go:generate` directive that regenerates `version.go` with build timestamp

Run with: `go run .` (prod) or `go run -tags dev .` (dev mode)

**Extension ideas:** add a Makefile that cross-compiles for all three platforms; wire up `goreleaser` to automate releases.
