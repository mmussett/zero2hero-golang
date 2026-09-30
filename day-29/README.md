# Day 29: Publishing and Deployment

## Publishing a Go Module

Any publicly accessible Git repository is a Go module. Publish by tagging:

```bash
git tag v0.1.0
git push origin v0.1.0
# Available at: pkg.go.dev/github.com/mmussett/zero2hero-golang/day-08
```

**Semantic versioning rules:**
- `v0.x.y` — unstable; breaking changes are fine
- `v1.x.y` — stable public API; breaking changes require bumping to `v2`
- `v2+` — module path and all import paths must include the `/v2` suffix

```go
module github.com/mmussett/zero2hero-golang/shapes/v2
```

## `go install` for CLI Tools

```bash
go install github.com/mmussett/zero2hero-golang/day-20@latest
# Installs the todo binary to $GOPATH/bin (or $GOBIN)
```

## Docker Multi-Stage Build

```dockerfile
# Build stage — full Go toolchain
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /notes-api .

# Runtime stage — minimal scratch image
FROM scratch
COPY --from=builder /notes-api /notes-api
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080
ENTRYPOINT ["/notes-api"]
```

`CGO_ENABLED=0` produces a fully static binary. `-ldflags="-s -w"` strips debug info, reducing binary size by ~30%.

## Embedding Version Information

```bash
go build -ldflags="-X main.version=$(git describe --tags) \
                   -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

```go
var (
    version   = "dev"
    buildTime = "unknown"
)

func main() {
    fmt.Printf("notes-api %s (built %s)\n", version, buildTime)
}
```

## goreleaser

`goreleaser` builds and publishes cross-platform releases automatically:

```yaml
# .goreleaser.yaml
builds:
  - env: [CGO_ENABLED=0]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags: ["-s -w -X main.version={{.Version}}"]

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'
```

```bash
goreleaser release --snapshot --clean   # local dry run
goreleaser release                      # real release (requires GITHUB_TOKEN)
```

## Day Project: Package and Ship

1. Extract the Day 08 shape library into `day-29/shapes/` with proper exported types and package docs
2. Write a `Dockerfile` for the Day 22 notes API using a multi-stage build
3. Write a `.goreleaser.yaml` for the Day 20 `todo` CLI targeting Linux, macOS, and Windows
4. Embed version + build time into the `todo` binary via `-ldflags`
5. Build and smoke-test the Docker image: `docker build -t notes-api . && docker run -p 8080:8080 notes-api`

**Extension ideas:** publish the shapes package to `pkg.go.dev` (make the repo public and push a tag); set up GitHub Actions to run `goreleaser` automatically on tag push.

## Official Documentation

- [Go Modules Reference](https://go.dev/doc/modules/gomod-ref) — `go.mod` syntax, `module` path, `require` directives
- [Module version numbering](https://go.dev/doc/modules/version-numbers) — semantic versioning rules for `v0`, `v1`, `v2+`
- [Publishing a module](https://go.dev/doc/modules/publishing) — how to make a module available on `pkg.go.dev`
- [`cmd/go` — go install](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies) — installing CLI binaries
- [`cmd/go` — ldflags](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies) — embedding version info with `-ldflags "-X main.version=..."`
- [`fmt`](https://pkg.go.dev/fmt) — `Printf` for printing version information at startup
- [goreleaser documentation](https://goreleaser.com/intro/) — cross-platform release automation
