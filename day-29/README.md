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

## Labs

### Lab 1: Version Embedding — ldflags and --version

**What you'll practise:** Embedding a version string and build time into a binary at compile time using `-ldflags -X`.

**Task:**
Add `version` and `buildTime` package-level variables to `main.go`. Write a `--version` flag that prints them, then write a `Makefile` target that injects real values at build time.

**Steps:**
1. Declare `var version = "dev"` and `var buildTime = "unknown"` in `main.go`
2. Parse `--version` with the `flag` package; if set, print and exit
3. Write a `Makefile` with a `build` target using `-ldflags`

```go
// main.go
var (
    version   = "dev"
    buildTime = "unknown"
)

func main() {
    showVersion := flag.Bool("version", false, "print version and exit")
    flag.Parse()
    if *showVersion {
        fmt.Printf("myapp %s (built %s)\n", version, buildTime)
        return
    }
    fmt.Println("Hello from myapp!")
}
```

```makefile
# Makefile
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
BUILDTIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	go build -ldflags="-X main.version=$(VERSION) -X main.buildTime=$(BUILDTIME)" -o myapp .
```

**Expected output:**
```
$ make build && ./myapp --version
myapp v0.1.0-3-gabcdef (built 2024-01-15T10:00:00Z)
```

**Checkpoint:** Running `./myapp --version` after `make build` prints a non-"dev" version. Building with plain `go build` (no Makefile) still falls back to `version=dev`.

---

### Lab 2: Multi-Stage Dockerfile — Measuring Image Size

**What you'll practise:** Writing a multi-stage Dockerfile with a `golang:1.23-alpine` builder stage and a minimal `scratch` or `alpine` runtime stage, then comparing image sizes.

**Task:**
Write a `Dockerfile` for the Day 22 notes API. Build it twice — once with `FROM alpine` as the runtime and once with `FROM scratch` — and compare the resulting image sizes with `docker images`.

**Steps:**
1. Stage 1: `FROM golang:1.23-alpine AS builder` — compile with `CGO_ENABLED=0`
2. Stage 2a: `FROM alpine` — copy the binary, expose port, set entrypoint
3. Build and record the size: `docker build -t notes-alpine .`
4. Stage 2b: Change to `FROM scratch`, rebuild as `notes-scratch`
5. Run `docker images notes-alpine notes-scratch` and compare

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /notes-api .

FROM scratch
COPY --from=builder /notes-api /notes-api
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080
ENTRYPOINT ["/notes-api"]
```

**Expected output:**
```
REPOSITORY     TAG     SIZE
notes-alpine   latest  18.2MB
notes-scratch  latest   8.4MB
```

**Checkpoint:** Both images start the API successfully (`docker run -p 8080:8080 notes-scratch`). The scratch image is noticeably smaller. You can explain why `-ldflags="-s -w"` reduces binary size.

---

### Lab 3: Cross-Compilation — Building for Multiple Platforms

**What you'll practise:** Using `GOOS` and `GOARCH` environment variables to cross-compile a Go binary for different operating systems and architectures from a single host.

**Task:**
Write a shell script (or Makefile targets) that compiles the `myapp` binary from Lab 1 for `linux/amd64`, `darwin/arm64`, and `windows/amd64`. Verify the output files are different formats.

**Steps:**
1. Run `GOOS=linux GOARCH=amd64 go build -o myapp-linux-amd64 .`
2. Run `GOOS=darwin GOARCH=arm64 go build -o myapp-darwin-arm64 .`
3. Run `GOOS=windows GOARCH=amd64 go build -o myapp-windows-amd64.exe .`
4. Inspect with `file myapp-*` (Linux/macOS) or check file extensions and sizes

```bash
#!/usr/bin/env bash
set -e
for target in "linux/amd64" "darwin/arm64" "windows/amd64"; do
    IFS='/' read -r os arch <<< "$target"
    ext=""
    [ "$os" = "windows" ] && ext=".exe"
    GOOS=$os GOARCH=$arch go build \
        -ldflags="-X main.version=1.0.0" \
        -o "dist/myapp-${os}-${arch}${ext}" .
    echo "Built: dist/myapp-${os}-${arch}${ext}"
done
```

**Expected output:**
```
Built: dist/myapp-linux-amd64
Built: dist/myapp-darwin-arm64
Built: dist/myapp-windows-amd64.exe
```

**Checkpoint:** Three files appear in `dist/`. On Linux/macOS, `file myapp-linux-amd64` reports "ELF 64-bit". The Windows `.exe` is built even from a non-Windows host.

---

### Lab 4: go install vs go build — CLI Tool Installation

**What you'll practise:** Understanding the difference between `go install` (places binary in `$GOPATH/bin`) and `go build` (places binary in the current directory or specified output path).

**Task:**
Install `myapp` with `go install`, confirm it lands in `$GOPATH/bin`, run it from anywhere, then compare with `go build` which writes to the working directory.

**Steps:**
1. Run `go install .` from the `day-29/` directory
2. Confirm the binary exists: `ls $(go env GOPATH)/bin/day-29` (or whichever name)
3. Run it from a different directory: `cd /tmp && day-29 --version`
4. Now run `go build -o ./myapp .` and confirm it only appears locally

```bash
# install to $GOPATH/bin
go install .

# confirm location
ls "$(go env GOPATH)/bin/"

# run from anywhere (assuming $GOPATH/bin is on $PATH)
myapp --version

# compare: go build writes locally only
go build -o ./myapp-local .
ls -la myapp-local
```

**Expected output:**
```
myapp v0.0.0 (built unknown)
-rwxr-xr-x  1 user  staff  3.2M myapp-local
```

**Checkpoint:** `go install` makes the binary available system-wide (if `$GOPATH/bin` is on `$PATH`). `go build` creates it only in the current directory.

---

### Lab 5: goreleaser — Local Snapshot Release

**What you'll practise:** Writing a `.goreleaser.yaml` configuration and running a local snapshot build to simulate a real multi-platform release.

**Task:**
Write a `.goreleaser.yaml` for the `myapp` binary targeting Linux, macOS, and Windows. Run `goreleaser release --snapshot --clean` to build all targets locally without publishing.

**Steps:**
1. Install goreleaser: `go install github.com/goreleaser/goreleaser/v2@latest`
2. Create `.goreleaser.yaml` in `day-29/`
3. Run `goreleaser release --snapshot --clean`
4. Inspect the `dist/` directory

```yaml
# .goreleaser.yaml
version: 2
project_name: myapp

builds:
  - env: [CGO_ENABLED=0]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w -X main.version={{.Version}} -X main.buildTime={{.Date}}

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: checksums.txt

changelog:
  sort: asc
```

**Expected output:**
```
  • building                   binary=dist/myapp_linux_amd64/myapp
  • building                   binary=dist/myapp_darwin_arm64/myapp
  • building                   binary=dist/myapp_windows_amd64/myapp.exe
  • archives                   archive=dist/myapp_linux_amd64.tar.gz
  • checksums                  file=dist/checksums.txt
```

**Checkpoint:** `dist/` contains archives for all target platforms plus a `checksums.txt`. Each archive contains the binary with `--version` working correctly.

---

### Final Lab (Project): Docker Multi-Stage Build + goreleaser + ldflags

**What you'll practise:** Combining version embedding, a multi-stage Dockerfile, and goreleaser into a complete publish-and-ship workflow.

**Task:**
Package the Day 22 notes API with a production-ready Dockerfile, add a `/health` endpoint that returns embedded version info, and produce a multi-platform release with goreleaser.

**Steps:**
1. Extract the Day 08 shape library into `day-29/shapes/` with proper exported types and package docs
2. Write a `Dockerfile` for the Day 22 notes API using a multi-stage build
3. Write a `.goreleaser.yaml` for the Day 20 `todo` CLI targeting Linux, macOS, and Windows
4. Embed version + build time into the `todo` binary via `-ldflags`
5. Add a `GET /health` endpoint that returns the version and build time as JSON
6. Build and smoke-test the Docker image: `docker build -t notes-api . && docker run -p 8080:8080 notes-api`

```go
// health endpoint
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":     "ok",
        "version":    version,
        "build_time": buildTime,
    })
})
```

**Expected output:**
```bash
$ curl http://localhost:8080/health
{"build_time":"2024-01-15T10:00:00Z","status":"ok","version":"v0.1.0"}

$ docker images notes-api
REPOSITORY   TAG      SIZE
notes-api    latest   9.1MB
```

**Checkpoint:** The Docker image starts and responds to `/health` with version info. `goreleaser --snapshot --clean` produces archives for all three platforms. Run with: `docker build -t notes-api . && docker run -p 8080:8080 notes-api`

**Extension ideas:** publish the shapes package to `pkg.go.dev` (make the repo public and push a tag); set up GitHub Actions to run `goreleaser` automatically on tag push.

## Official Documentation

- [Go Modules Reference](https://go.dev/doc/modules/gomod-ref) — `go.mod` syntax, `module` path, `require` directives
- [Module version numbering](https://go.dev/doc/modules/version-numbers) — semantic versioning rules for `v0`, `v1`, `v2+`
- [Publishing a module](https://go.dev/doc/modules/publishing) — how to make a module available on `pkg.go.dev`
- [`cmd/go` — go install](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies) — installing CLI binaries
- [`cmd/go` — ldflags](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies) — embedding version info with `-ldflags "-X main.version=..."`
- [`fmt`](https://pkg.go.dev/fmt) — `Printf` for printing version information at startup
- [goreleaser documentation](https://goreleaser.com/intro/) — cross-platform release automation
