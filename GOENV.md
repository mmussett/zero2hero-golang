# Go Environment Variables Reference

Official documentation: [cmd/go environment variables](https://pkg.go.dev/cmd/go#hdr-Environment_variables) | [Source install environment](https://go.dev/doc/install/source#environment)

---

## 1. Introduction

The `go` tool reads dozens of environment variables to control module resolution, caching, cross-compilation, the C toolchain, garbage collection, and more. You can inspect every variable in one shot with `go env`, or query a single one with `go env GOPATH`. Changes can be **temporary** — prefix the command with `VAR=value go build` — or **permanent**, stored in the user-level env file via `go env -w VAR=value`. Permanent settings are written to the file pointed at by `$GOENV` (default: `$GOENV` itself resolves to `$HOME/.config/go/env` on Linux/macOS and `%AppData%\go\env` on Windows). To remove a persistent override and fall back to the compiled-in default, run `go env -u VAR`. All of this works without touching your shell profile, which makes it easy to share settings across shells and CI environments consistently.

---

## 2. Module & Dependency Management

### `GOPATH`
**Default:** `$HOME/go` (Linux/macOS) or `%USERPROFILE%\go` (Windows)
**Set with:** `go env -w GOPATH=/path/to/gopath`

`GOPATH` was historically the single workspace that held all Go source, compiled binaries, and downloaded packages. In module mode (Go 1.11+) it is much less prominent: source no longer lives here, but `$GOPATH/bin` is still where `go install` places executables, and `$GOPATH/pkg/mod` is the module cache. You might override it to place the cache on a larger drive or to isolate projects in CI.

```bash
# Check current value
go env GOPATH

# Move everything to a bigger disk
go env -w GOPATH=/mnt/data/go

# Temporarily use an isolated workspace for one build
GOPATH=/tmp/ci-go go build ./...
```

> **Gotcha:** If `$GOPATH/bin` is not on your `$PATH`, commands installed with `go install` will be silently unreachable. Add `export PATH="$(go env GOPATH)/bin:$PATH"` to your shell profile.

---

### `GOMODCACHE`
**Default:** `$GOPATH/pkg/mod`
**Set with:** `go env -w GOMODCACHE=/path/to/cache`

The directory where the module cache lives. Downloaded module zip files and their extracted trees are stored here. Separating `GOMODCACHE` from `GOPATH` is useful in Docker multi-stage builds: you can mount the cache as a build-time volume without shipping it in the final image.

```bash
# Point the cache to a shared network volume for a team
go env -w GOMODCACHE=/nfs/go-mod-cache

# Docker: mount host cache into the builder stage
# Dockerfile excerpt:
# RUN --mount=type=cache,target=/root/go/pkg/mod \
#     go build -o /out/app ./cmd/app
```

> **Gotcha:** The module cache is read-only by design (`chmod 0555` on directories). Do not try to edit files inside it; run `go clean -modcache` to wipe it instead.

---

### `GOPROXY`
**Default:** `https://proxy.golang.org,direct`
**Set with:** `go env -w GOPROXY=https://my-proxy.internal,direct`

A comma-separated list of module proxy URLs consulted in order when fetching module versions. The special keyword `direct` means "contact the VCS origin directly". `off` means "never fetch; fail if the module is not already cached". This is the single most impactful variable for working with private registries or air-gapped networks.

```bash
# Corporate environment: hit internal proxy first, fall back to public
go env -w GOPROXY=https://goproxy.internal,https://proxy.golang.org,direct

# Air-gapped build: only use what is already in the cache
go env -w GOPROXY=off

# CI with a local Athens instance
GOPROXY=http://localhost:3000 go mod download
```

> **Gotcha:** If a proxy returns a 404, the Go tool moves on to the next entry. If the proxy returns a 410 (Gone), Go treats that module version as permanently unavailable and stops trying — it will not fall through to `direct`. This is intentional security behaviour.

---

### `GONOPROXY`
**Default:** *not set*
**Set with:** `go env -w GONOPROXY=*.internal.corp,github.com/myorg/*`

A comma-separated list of glob patterns for module path prefixes that should **bypass** the proxy and be fetched directly from VCS. Patterns follow the same syntax as `GOPRIVATE` but only affect proxy lookup; checksum database lookup is controlled separately by `GONOSUMDB`.

```bash
# Bypass the proxy for all internal modules
go env -w GONOPROXY=*.corp.example.com,github.com/acme-internal/*

# Verify that a specific module would bypass the proxy
GONOPROXY=github.com/myorg/* go env GONOPROXY
```

---

### `GOPRIVATE`
**Default:** *not set*
**Set with:** `go env -w GOPRIVATE=github.com/myorg,*.internal`

A convenience shorthand: setting `GOPRIVATE=pattern` is equivalent to setting both `GONOPROXY=pattern` and `GONOSUMDB=pattern` simultaneously. Use this for private modules that should never touch the public proxy or checksum database.

```bash
# Everything under your GitHub org is private
go env -w GOPRIVATE=github.com/acme/*

# Also covers an internal GitLab instance
go env -w GOPRIVATE=github.com/acme/*,gitlab.internal.corp/*
```

> **Gotcha:** `GOPRIVATE` is a prefix match by glob, not a substring match. `github.com/acme/*` matches `github.com/acme/foo` but not `github.com/acme-partner/foo`.

---

### `GONOSUMDB`
**Default:** *not set*
**Set with:** `go env -w GONOSUMDB=*.internal,github.com/myorg/*`

A comma-separated list of glob patterns for modules whose checksums should **not** be verified against the public checksum database (`sum.golang.org`). Necessary for private modules that aren't listed in the public database, or for modules on internal registries that aren't reachable from the internet.

```bash
# Skip checksum DB for all private modules
go env -w GONOSUMDB=github.com/acme/*,*.corp.example.com

# Confirm the setting takes effect
go env GONOSUMDB
```

---

### `GONOSUMCHECK`
**Default:** *not set*
**Set with:** `go env -w GONOSUMCHECK=pattern`

Similar to `GONOSUMDB` but provides an additional layer: modules matching these patterns skip both the checksum database lookup **and** the local `go.sum` verification step entirely. Intended for extreme cases such as vendored forks or modules loaded from local disk during development. Use with caution — you lose tamper detection.

```bash
# Skip all sum checks for a local replace directive during development
go env -w GONOSUMCHECK=github.com/myorg/experimental
```

> **Gotcha:** Do not set this in production CI. It silently disables the supply-chain integrity check that `go.sum` provides.

---

### `GOFLAGS`
**Default:** *not set*
**Set with:** `go env -w GOFLAGS=-mod=vendor`

A space-separated list of default flags applied to every applicable `go` subcommand. Values here behave as if you typed them on every invocation. Useful for enforcing project-wide settings in CI without wrapper scripts.

```bash
# Always build with the race detector and verbose module output
go env -w GOFLAGS=-race

# Vendor mode everywhere (no network lookups)
go env -w GOFLAGS=-mod=vendor

# Multiple flags (quote as a single string)
go env -w GOFLAGS="-mod=vendor -trimpath"
```

> **Gotcha:** Flags set in `GOFLAGS` apply to every `go` subcommand that accepts them. Setting `-race` here means `go test`, `go build`, and `go run` all compile with the race detector — which significantly increases binary size and slows execution. Consider setting it only in CI or test-specific scripts rather than globally.

---

### `GOWORK`
**Default:** *not set* (auto-detected by walking up directories)
**Set with:** `GOWORK=/path/to/go.work go build ./...`

The path to the `go.work` file for a multi-module workspace. When set to `off`, workspace mode is disabled entirely, forcing the tool to treat each module independently — useful when you want to verify that a library works without its workspace siblings.

```bash
# Disable workspace mode for a single command
GOWORK=off go test ./...

# Point explicitly at a workspace file outside the CWD
GOWORK=/home/user/projects/go.work go build ./cmd/server
```

---

### `GOTOOLCHAIN`
**Default:** `local` (or the value baked into the Go release)
**Set with:** `go env -w GOTOOLCHAIN=go1.22.0`

Controls which Go toolchain version is used. Starting with Go 1.21, `go.mod` and `go.work` can declare a minimum toolchain version (`toolchain go1.22.0`). `GOTOOLCHAIN=local` means "use whatever is installed locally and do not auto-download". `GOTOOLCHAIN=auto` means "download and switch to the version declared in `go.mod`/`go.work` if the local version is older". A specific version like `go1.22.3` pins exactly that release.

```bash
# Use whatever is installed (no auto-download)
go env -w GOTOOLCHAIN=local

# Auto-upgrade to satisfy go.mod's toolchain directive
go env -w GOTOOLCHAIN=auto

# Pin to a specific toolchain for reproducibility in CI
GOTOOLCHAIN=go1.22.3 go build ./...
```

> **Gotcha:** With `GOTOOLCHAIN=auto`, the first build after updating `go.mod` may download a new toolchain from `proxy.golang.org`. In air-gapped environments, set `GOTOOLCHAIN=local` and provision the correct Go version through your own package management.

---

## 3. Build & Toolchain

### `GOROOT`
**Default:** the directory where Go is installed (detected at build time)
**Set with:** `export GOROOT=/usr/local/go` (rarely needed)

The root of the Go installation tree. The standard library lives at `$GOROOT/src`. Normally you never set this — the `go` binary knows its own location. The rare case where you might override it is when running a custom-built Go toolchain from a non-standard path.

```bash
# Inspect the current root
go env GOROOT

# Use a side-by-side Go installation (advanced)
GOROOT=/home/user/go-tip go build ./...
```

> **Gotcha:** Setting `GOROOT` incorrectly breaks all builds because the standard library becomes unreachable. Prefer managing multiple Go versions with `go install golang.org/dl/go1.22.0@latest` or a version manager instead.

---

### `GOBIN`
**Default:** `$GOPATH/bin`
**Set with:** `go env -w GOBIN=/usr/local/bin`

The directory where `go install` places compiled binaries. Override this to install tools into a system-wide directory or a project-local `bin/` folder.

```bash
# Install tools into a project-local bin/ directory
go env -w GOBIN=$(pwd)/bin
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install system-wide (requires write permission)
GOBIN=/usr/local/bin go install golang.org/x/tools/cmd/goimports@latest
```

---

### `GOTMPDIR`
**Default:** the OS temporary directory (`$TMPDIR` / `/tmp` / `%TEMP%`)
**Set with:** `go env -w GOTMPDIR=/fast-ssd/tmp`

The directory used for temporary files during compilation (object files, intermediate outputs). Pointing this at a fast local SSD or a `tmpfs` mount can meaningfully speed up large builds.

```bash
# Use a RAM disk for faster builds on Linux
mount -t tmpfs tmpfs /mnt/ramfs -o size=4G
go env -w GOTMPDIR=/mnt/ramfs
```

---

### `GOCACHE`
**Default:** `$HOME/.cache/go/build` (Linux/macOS) or `%LocalAppData%\go-build` (Windows)
**Set with:** `go env -w GOCACHE=/path/to/cache`

The build cache directory. The Go toolchain caches compiled packages here to avoid recompilation. In CI, mounting this directory between runs dramatically reduces build times.

```bash
# Check cache location
go env GOCACHE

# Redirect to a CI-friendly path
go env -w GOCACHE=/tmp/go-build-cache

# Warm the cache explicitly before building
go build -v ./... 2>/dev/null || true

# Clear the cache
go clean -cache

# GitHub Actions example: cache the build cache
# - uses: actions/cache@v4
#   with:
#     path: ~/.cache/go/build
#     key: ${{ runner.os }}-go-build-${{ hashFiles('**/go.sum') }}
```

---

### `GOTOOLDIR`
**Default:** `$GOROOT/pkg/tool/$GOOS_$GOARCH`
**Set with:** rarely overridden

The directory containing the low-level Go tool binaries (`compile`, `link`, `asm`, `cover`, etc.). You will almost never set this; it is surfaced mainly so build systems and IDEs can locate the compiler.

```bash
go env GOTOOLDIR
# e.g. /usr/local/go/pkg/tool/linux_amd64
```

---

### `GOVERSION`
**Default:** the version string of the installed toolchain (read-only)
**Set with:** not settable

Reports the Go version of the current toolchain. Read-only; it cannot be overridden.

```bash
go env GOVERSION
# go1.22.3
```

---

### `GOEXPERIMENT`
**Default:** *not set*
**Set with:** `GOEXPERIMENT=rangefunc go build ./...`

A comma-separated list of experimental language or runtime features to enable. These features are not stable and may be removed or changed. Used by the Go team to ship opt-in previews before they graduate to stable.

```bash
# Enable the range-over-function iterator experiment (Go 1.22 preview)
GOEXPERIMENT=rangefunc go build ./...

# List all experiments compiled into the current toolchain
go env GOEXPERIMENT
```

> **Gotcha:** Never enable experiments in production binaries. They are not covered by the Go compatibility promise and may change between patch releases.

---

### `GOENV`
**Default:** `$HOME/.config/go/env` (Linux/macOS) or `%AppData%\go\env` (Windows)
**Set with:** `GOENV=/path/to/custom/env go build` (shell only; cannot use `go env -w` to change this one)

The path to the file that stores persistent `go env -w` settings. You can redirect this to a project-specific or CI-specific file to isolate settings. Because `GOENV` itself controls where the env file is read from, it cannot be stored inside that same file — you must set it as a real shell environment variable.

```bash
# Use a project-specific env file
export GOENV=$(pwd)/.goenv
go env -w GOPROXY=off

# Inspect what the env file currently contains
cat "$(go env GOENV)"
```

---

## 4. Cross-Compilation

Go's cross-compilation story is exceptional: a single toolchain can produce binaries for any supported OS/architecture combination just by setting two environment variables. No additional toolchain installation is required for pure-Go code.

### `GOOS`
**Default:** the host operating system
**Set with:** `GOOS=linux go build ./...`

The target operating system. Valid values include `linux`, `darwin`, `windows`, `freebsd`, `openbsd`, `netbsd`, `plan9`, `solaris`, `illumos`, `js`, `wasip1`, and more. The full list is in `go tool dist list`.

```bash
# Build a Linux binary from macOS
GOOS=linux GOARCH=amd64 go build -o app-linux ./cmd/app

# Build a Windows binary from Linux
GOOS=windows GOARCH=amd64 go build -o app.exe ./cmd/app

# List all supported OS/arch pairs
go tool dist list
```

---

### `GOARCH`
**Default:** the host CPU architecture
**Set with:** `GOARCH=arm64 go build ./...`

The target CPU architecture. Common values: `amd64`, `arm`, `arm64`, `386`, `mips`, `mips64`, `ppc64`, `riscv64`, `s390x`, `wasm`.

```bash
# Cross-compile for Raspberry Pi 4 (64-bit ARM)
GOOS=linux GOARCH=arm64 go build -o app-rpi ./cmd/app

# Cross-compile for Raspberry Pi Zero / Pi 1 (32-bit ARM v6)
GOOS=linux GOARCH=arm GOARM=6 go build -o app-rpi-zero ./cmd/app
```

---

### `GOARM`
**Default:** `7`
**Set with:** `GOARM=6 go build ./...`

For `GOARCH=arm` (32-bit ARM), selects the ARM instruction set version: `5` (ARMv5, software float), `6` (ARMv6, Raspberry Pi Zero/1), or `7` (ARMv7, Raspberry Pi 2/3 in 32-bit mode). Has no effect for other architectures.

```bash
# Raspberry Pi Zero W (ARMv6, hardware float)
GOOS=linux GOARCH=arm GOARM=6 go build -o app ./cmd/app

# ARMv5 device with no hardware FPU
GOOS=linux GOARCH=arm GOARM=5 go build -o app ./cmd/app
```

---

### `GOAMD64`
**Default:** `v1`
**Set with:** `GOAMD64=v3 go build ./...`

For `GOARCH=amd64`, selects the x86-64 microarchitecture level: `v1` (baseline, runs everywhere), `v2` (SSE4, POPCNT — most modern x86-64 CPUs), `v3` (AVX2, BMI1/BMI2 — Intel Haswell and later), `v4` (AVX-512). Higher levels produce faster code but the binary will crash with an illegal-instruction fault on older CPUs.

```bash
# Fast binary for a controlled modern server fleet
GOAMD64=v3 go build -o app ./cmd/app

# Safe baseline binary for broad distribution
GOAMD64=v1 go build -o app ./cmd/app
```

> **Gotcha:** Unlike `GOARM`, there is no runtime fallback — a `v3` binary simply faults on a `v1` CPU.

---

### `GOMIPS`
**Default:** `hardfloat`
**Set with:** `GOMIPS=softfloat go build ./...`

For `GOARCH=mips` or `mips64`, selects `hardfloat` or `softfloat`. Use `softfloat` for MIPS devices that lack an FPU.

```bash
GOOS=linux GOARCH=mips GOMIPS=softfloat go build -o app ./cmd/app
```

---

### `GOPPC64`
**Default:** `power8`
**Set with:** `GOPPC64=power9 go build ./...`

For `GOARCH=ppc64` or `ppc64le`, selects the PowerPC ISA level: `power8` or `power9`. `power9` enables additional vector and crypto instructions.

```bash
GOOS=linux GOARCH=ppc64le GOPPC64=power9 go build -o app ./cmd/app
```

---

### `GOWASM`
**Default:** *not set*
**Set with:** `GOWASM=satconv,signext go build ./...`

For `GOARCH=wasm`, a comma-separated list of WebAssembly extensions to use. Currently recognised values: `satconv` (saturating conversions) and `signext` (sign-extension instructions). These correspond to the WebAssembly post-MVP proposals and are supported by modern browsers and runtimes.

```bash
GOOS=js GOARCH=wasm GOWASM=satconv,signext go build -o app.wasm ./cmd/app
```

---

### `CGO_ENABLED` (cross-compilation context)

When cross-compiling, CGo is **disabled by default** because the host C compiler cannot target the foreign architecture without an explicit cross-compiler. For pure-Go code this is ideal — set `CGO_ENABLED=0` explicitly to make the intent clear and ensure a fully static binary.

```bash
# Explicit static binary for Linux (great for Docker scratch images)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o app ./cmd/app
```

See the full `CGO_ENABLED` entry in Section 5 for all details.

---

## 5. CGo

CGo allows Go packages to call C (and C++) code. It is powerful but adds complexity: the build requires a C toolchain, disables certain optimisations, prevents some cross-compilation scenarios, and increases link times. Each CGo variable is consulted only when `CGO_ENABLED=1`.

### `CGO_ENABLED`
**Default:** `1` when a C compiler is available and `GOOS`/`GOARCH` match the host; `0` otherwise
**Set with:** `go env -w CGO_ENABLED=0`

The master switch for CGo. When `0`, all CGo imports are rejected and the Go toolchain produces fully static binaries with no libc dependency. This is the recommended setting for Docker images built on `scratch` or `distroless` base images.

```bash
# Build a fully static binary — no libc, runs in scratch containers
CGO_ENABLED=0 go build -o app ./cmd/app

# Explicitly enable CGo (e.g. needed for sqlite3 drivers)
CGO_ENABLED=1 go build -tags cgo ./cmd/app

# Dockerfile multi-stage pattern:
# FROM golang:1.22 AS builder
# RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/server
# FROM scratch
# COPY --from=builder /app /app
# ENTRYPOINT ["/app"]
```

> **Gotcha:** `go test` can behave differently with CGo enabled vs disabled. The race detector (`-race`) requires CGo, so `CGO_ENABLED=0 go test -race` will fail.

---

### `CC`
**Default:** `gcc` (or `clang` on some systems)
**Set with:** `CC=musl-gcc go build ./...`

The C compiler used for CGo compilation. Override this when using a cross-compiler or when you want a specific compiler (e.g. `musl-gcc` for fully static musl-libc binaries, or `arm-linux-gnueabihf-gcc` for 32-bit ARM cross-compilation).

```bash
# Cross-compile CGo code for 64-bit ARM Linux
CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build ./...

# Build against musl for a smaller, fully static binary
CC=musl-gcc CGO_ENABLED=1 go build -ldflags="-linkmode external -extldflags -static" ./...
```

---

### `CXX`
**Default:** `g++` (or `clang++`)
**Set with:** `CXX=arm-linux-gnueabihf-g++ go build ./...`

The C++ compiler used when CGo code includes C++ sources or when linking against C++ libraries. Rarely set independently of `CC`.

```bash
CXX=clang++ CC=clang CGO_ENABLED=1 go build ./...
```

---

### `CGO_CFLAGS`
**Default:** `-O2 -g`
**Set with:** `CGO_CFLAGS="-I/opt/include -O3" go build ./...`

Extra flags passed to the C compiler when compiling CGo C files. Use this to add include paths (`-I`) or optimisation flags.

```bash
# Add a custom include path for a vendored C library
CGO_CFLAGS="-I$(pwd)/vendor/mylib/include" go build ./...

# Override optimisation level
CGO_CFLAGS="-O3" go build ./...
```

> **Gotcha:** These flags apply to every CGo-using package in the build, not just your own. If a transitive dependency also uses CGo, it gets these flags too. Use `#cgo CFLAGS:` directives in Go source files for per-package flags instead.

---

### `CGO_CXXFLAGS`
**Default:** `-O2 -g`
**Set with:** `CGO_CXXFLAGS="-std=c++17 -I/opt/include" go build ./...`

Extra flags passed to the C++ compiler. Mirrors `CGO_CFLAGS` but for C++ sources.

```bash
CGO_CXXFLAGS="-std=c++17" go build ./...
```

---

### `CGO_CPPFLAGS`
**Default:** *not set*
**Set with:** `CGO_CPPFLAGS="-DNDEBUG -DVERSION=2" go build ./...`

Flags passed to the C preprocessor for both C and C++ CGo files. Use for macro definitions (`-D`) and include paths that apply to both languages.

```bash
CGO_CPPFLAGS="-DNDEBUG -DMYLIB_STATIC" go build ./...
```

---

### `CGO_LDFLAGS`
**Default:** *not set*
**Set with:** `CGO_LDFLAGS="-L/opt/lib -lmylib" go build ./...`

Extra flags passed to the linker when linking CGo binaries. Use to add library search paths (`-L`) and library names (`-l`).

```bash
# Link against a system library not on the default search path
CGO_LDFLAGS="-L/opt/openssl/lib -lssl -lcrypto" go build ./...

# Statically link a C library
CGO_LDFLAGS="-L$(pwd)/vendor/mylib -lmylib -static" go build ./...
```

> **Gotcha:** Flags in `CGO_LDFLAGS` are passed verbatim to the linker. The `go` tool validates them against a security allowlist when building with `go get` or in module mode to prevent injection attacks. If you need unusual flags, use `#cgo LDFLAGS:` in your Go source file instead — those are whitelisted at the package level.

---

## 6. Runtime Behaviour

These variables control how a compiled Go program behaves at runtime. Most are set when *running* a binary, not when building it (though `GOMAXPROCS` can also be set in code via `runtime.GOMAXPROCS()`).

### `GOMAXPROCS`
**Default:** number of logical CPUs on the machine
**Set with:** `GOMAXPROCS=4 ./myapp`

Controls the maximum number of OS threads that can execute Go code simultaneously. Setting it lower than the CPU count is useful for testing concurrency behaviour or for throttling a service that shares a host with other processes.

```bash
# Simulate a 2-core environment during testing
GOMAXPROCS=2 go test -race ./...

# Run production service on 8 of 32 cores
GOMAXPROCS=8 ./server

# In code: runtime.GOMAXPROCS(4)
```

> **Gotcha:** `GOMAXPROCS` does not limit goroutine count — you can have millions of goroutines with `GOMAXPROCS=1`. It only controls parallelism of execution. Also, cgroups-aware environments (Kubernetes, Docker) may set this automatically via the `automaxprocs` package; check before overriding.

---

### `GOMEMLIMIT`
**Default:** *not set* (no limit)
**Set with:** `GOMEMLIMIT=512MiB ./myapp`

A soft memory limit for the Go runtime (introduced in Go 1.19). When the total heap size approaches this limit, the GC runs more aggressively to stay under it. Accepts values like `512MiB`, `1GiB`, or a plain byte count. This does not hard-cap memory — the program can still exceed the limit if GC cannot reclaim enough — but it prevents the GC from being overly lazy when memory is constrained.

```bash
# Cap a microservice in a container with 512 MiB RAM
GOMEMLIMIT=450MiB ./server   # leave ~12% headroom for non-heap allocations

# Useful in Kubernetes: set to ~90% of the container memory limit
# resources:
#   limits:
#     memory: 512Mi
# env:
#   - name: GOMEMLIMIT
#     value: "460MiB"
```

> **Gotcha:** Setting `GOMEMLIMIT` too close to the container's hard limit leaves no room for stack memory, goroutine overhead, and OS-level memory. Use roughly 90% of the container limit as a rule of thumb.

---

### `GODEBUG`
**Default:** *not set*
**Set with:** `GODEBUG=gctrace=1,schedtrace=1000 ./myapp`

A comma-separated list of `key=value` pairs that enable debug output and tweak runtime behaviours. This is a broad escape hatch used for diagnosing GC behaviour, scheduler decisions, HTTP/2 issues, TLS handshakes, and more. Key values include:

| Key | Value | Effect |
|-----|-------|--------|
| `gctrace` | `1` | Print one line per GC cycle to stderr |
| `schedtrace` | `N` | Print scheduler state every N milliseconds |
| `madvdontneed` | `1` | Use `MADV_DONTNEED` to release memory to OS sooner |
| `http2debug` | `1` or `2` | Log HTTP/2 framing details |
| `tlsdebug` | `1` | Log TLS handshake details |
| `asyncpreemptoff` | `1` | Disable asynchronous goroutine preemption |
| `clobberfree` | `1` | Overwrite freed memory (detects use-after-free) |
| `inittrace` | `1` | Print timing for each `init()` function |

```bash
# Diagnose GC pressure
GODEBUG=gctrace=1 ./myapp 2>&1 | grep ^gc

# Trace scheduler to find goroutine starvation
GODEBUG=schedtrace=500,scheddetail=1 ./myapp

# Debug TLS handshake failures
GODEBUG=tlsdebug=1 ./myapp
```

> **Gotcha:** `GODEBUG` output is written to stderr and can be extremely verbose. Never set `gctrace=1` in production without log-rate limiting — it can saturate disk I/O.

---

### `GOTRACEBACK`
**Default:** `single`
**Set with:** `GOTRACEBACK=all ./myapp`

Controls how much information is printed when a Go program crashes (panics or receives an unhandled signal). Values:

| Value | Effect |
|-------|--------|
| `none` | No goroutine stacks, just the panic message |
| `single` | Stack of the goroutine that crashed (default) |
| `all` | Stacks of all goroutines |
| `system` | Like `all` plus internal runtime goroutines |
| `crash` | Like `system`, plus triggers a core dump if possible |

```bash
# Get full picture during a crash investigation
GOTRACEBACK=all ./myapp

# Generate a core dump for post-mortem analysis with dlv
GOTRACEBACK=crash ./myapp
```

---

### `GOGC`
**Default:** `100`
**Set with:** `GOGC=200 ./myapp` or `GOGC=off ./myapp`

Sets the initial garbage-collection target percentage. A value of `100` means the GC triggers a cycle when the live heap has grown by 100% since the last collection (i.e. it doubles). Increasing `GOGC` reduces GC frequency at the cost of higher peak memory usage. `GOGC=off` disables automatic GC entirely (useful for benchmarks and short-lived batch programs). `GOGC=200` is a common production tuning choice for throughput-sensitive services.

```bash
# Less GC, more memory — good for CPU-bound services
GOGC=200 ./server

# Disable GC for a short-lived batch job
GOGC=off ./batch-import --input data.csv

# Aggressive GC for memory-constrained environments
GOGC=50 ./server

# In code: debug.SetGCPercent(200)
```

> **Gotcha:** Since Go 1.19, `GOMEMLIMIT` is usually a better knob than `GOGC` for container environments because it gives absolute bounds rather than relative ones. Consider using both together: `GOMEMLIMIT` for ceiling safety, `GOGC` for throughput tuning.

---

## 7. Security & Privacy

### `GONOSUMDB`
See [Section 2](#gonosumdb) for the full entry.

### `GONOSUMCHECK`
See [Section 2](#gonosumcheck) for the full entry.

### `GOPRIVATE` (cross-reference)
See [Section 2](#goprivate) for the full entry. `GOPRIVATE` is the recommended single variable for private modules: it sets both `GONOPROXY` and `GONOSUMDB` simultaneously.

---

### `GOTELEMETRY`
**Default:** `local`
**Set with:** `go env -w GOTELEMETRY=off`

Controls whether the Go toolchain uploads usage telemetry to Google. Introduced in Go 1.23. Values:

| Value | Effect |
|-------|--------|
| `off` | No data collected or uploaded |
| `local` | Data collected locally but not uploaded (default) |
| `on` | Data collected and uploaded to telemetry.go.dev |

```bash
# Disable all telemetry (common in enterprise/air-gapped environments)
go env -w GOTELEMETRY=off

# Opt into uploading to help the Go team
go env -w GOTELEMETRY=on

# Check current mode
go env GOTELEMETRY
```

> **Gotcha:** Telemetry is opt-in for uploads (`on` must be set explicitly). The `local` default collects data to disk without transmitting it, which allows the `go telemetry` command to show local statistics.

---

## 8. Debugging & Profiling

This section cross-references variables from other sections and highlights how they combine for debugging and profiling workflows.

### `GOFLAGS` (cross-reference for debugging)
See [Section 2](#goflags). Common debugging uses:

```bash
# Always build with race detector during development
go env -w GOFLAGS=-race

# Always enable coverage instrumentation
go env -w GOFLAGS=-cover
```

### `GOGC` (cross-reference)
See [Section 6](#gogc). For benchmarking, disable GC to get stable measurements:

```bash
GOGC=off go test -bench=. -benchmem ./...
```

### `GOMAXPROCS` (cross-reference)
See [Section 6](#gomaxprocs). For profiling single-threaded vs multi-threaded performance:

```bash
# Profile with a single thread to isolate algorithmic cost
GOMAXPROCS=1 go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

### Combined example: full observability during a load test

```bash
GODEBUG=gctrace=1,schedtrace=1000 \
GOMAXPROCS=4 \
GOGC=200 \
GOMEMLIMIT=1GiB \
./server 2>runtime-debug.log &

# In another terminal
go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## 9. Quick Reference Table

| Variable | Default | One-line description |
|----------|---------|----------------------|
| `GOPATH` | `$HOME/go` | Workspace root; `bin/` subdir holds installed binaries |
| `GOROOT` | Go install dir | Location of the Go standard library and toolchain |
| `GOMODCACHE` | `$GOPATH/pkg/mod` | Directory where downloaded module zips are cached |
| `GOPROXY` | `https://proxy.golang.org,direct` | Ordered list of module proxy URLs |
| `GONOPROXY` | *not set* | Module path globs to fetch directly, bypassing the proxy |
| `GOPRIVATE` | *not set* | Shorthand for setting both GONOPROXY and GONOSUMDB |
| `GONOSUMDB` | *not set* | Module path globs to skip checksum database verification |
| `GONOSUMCHECK` | *not set* | Module path globs to skip all checksum verification |
| `GOFLAGS` | *not set* | Default flags applied to every `go` subcommand |
| `GOWORK` | auto-detected | Path to `go.work` file; `off` disables workspace mode |
| `GOTOOLCHAIN` | `local` | Which Go toolchain version to use; `auto` downloads if needed |
| `GOBIN` | `$GOPATH/bin` | Directory where `go install` places binaries |
| `GOTMPDIR` | OS temp dir | Directory for temporary files during compilation |
| `GOCACHE` | `~/.cache/go/build` | Build artifact cache directory |
| `GOTOOLDIR` | `$GOROOT/pkg/tool/…` | Location of internal compiler tools (compile, link, asm) |
| `GOVERSION` | toolchain version | Read-only: version string of the installed toolchain |
| `GOEXPERIMENT` | *not set* | Comma-separated list of experimental features to enable |
| `GOENV` | `~/.config/go/env` | Path to the persistent `go env -w` settings file |
| `GOOS` | host OS | Target operating system for compilation |
| `GOARCH` | host CPU arch | Target CPU architecture for compilation |
| `GOARM` | `7` | ARM sub-architecture version (5, 6, or 7) |
| `GOAMD64` | `v1` | x86-64 microarchitecture level (v1–v4) |
| `GOMIPS` | `hardfloat` | MIPS float ABI: `hardfloat` or `softfloat` |
| `GOPPC64` | `power8` | PowerPC ISA level: `power8` or `power9` |
| `GOWASM` | *not set* | WebAssembly extensions (satconv, signext) |
| `CGO_ENABLED` | `1` (if C compiler present) | Master switch for CGo; `0` produces fully static binaries |
| `CC` | `gcc` | C compiler used by CGo |
| `CXX` | `g++` | C++ compiler used by CGo |
| `CGO_CFLAGS` | `-O2 -g` | Extra flags for the C compiler during CGo builds |
| `CGO_CXXFLAGS` | `-O2 -g` | Extra flags for the C++ compiler during CGo builds |
| `CGO_CPPFLAGS` | *not set* | Preprocessor flags for both C and C++ CGo files |
| `CGO_LDFLAGS` | *not set* | Extra linker flags for CGo binaries |
| `GOMAXPROCS` | number of logical CPUs | Maximum OS threads executing Go code in parallel |
| `GOMEMLIMIT` | *not set* | Soft heap memory limit for the runtime (e.g. `512MiB`) |
| `GODEBUG` | *not set* | Comma-separated runtime debug knobs (gctrace, schedtrace, …) |
| `GOTRACEBACK` | `single` | Verbosity of stack traces on crash |
| `GOGC` | `100` | GC trigger percentage; `off` disables automatic GC |
| `GOTELEMETRY` | `local` | Toolchain usage telemetry: `off`, `local`, or `on` |

---

## 10. Common Recipes

### 1. Offline / air-gapped build (all dependencies pre-downloaded)

```bash
# Vendor dependencies first (run once with network access)
go mod vendor

# Then build offline — no proxy, no sum DB, use vendor/
GOFLAGS=-mod=vendor GOPROXY=off GONOSUMCHECK=* go build ./...
```

### 2. Cross-compile for Raspberry Pi 4 (64-bit ARM Linux)

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
  go build -o app-rpi4 ./cmd/app
# Copy to Pi:
scp app-rpi4 pi@raspberrypi.local:~/app
```

### 3. Cross-compile for Raspberry Pi Zero / Pi 1 (32-bit ARM v6)

```bash
GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 \
  go build -o app-rpi-zero ./cmd/app
```

### 4. Fully static binary for a Docker scratch image

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o app ./cmd/app

# Dockerfile:
# FROM scratch
# COPY app /app
# ENTRYPOINT ["/app"]
```

### 5. Point builds at a private module proxy (e.g. Athens or Artifactory)

```bash
go env -w GOPROXY=https://goproxy.corp.example.com,direct
go env -w GOPRIVATE=github.com/corp-org/*,gitlab.internal.corp/*
go env -w GONOSUMDB=github.com/corp-org/*,gitlab.internal.corp/*
```

### 6. Speed up CI with build and module caches

```bash
# GitHub Actions workflow snippet:
# - uses: actions/cache@v4
#   with:
#     path: |
#       ~/.cache/go/build
#       ~/go/pkg/mod
#     key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

# Or set cache paths explicitly:
go env -w GOCACHE=/ci/cache/go/build
go env -w GOMODCACHE=/ci/cache/go/pkg/mod
```

### 7. Run with the race detector (without setting GOFLAGS globally)

```bash
go test -race ./...
go run -race ./cmd/app
go build -race -o app-race ./cmd/app && ./app-race
```

### 8. Profile GC and scheduler during a load test

```bash
GODEBUG=gctrace=1,schedtrace=2000 GOMAXPROCS=4 \
  ./server 2>runtime.log
# Parse GC lines:
grep '^gc' runtime.log | awk '{print $3, $5, $9}'
```

### 9. Build a musl-linked static binary on Linux (for Alpine containers)

```bash
# Requires musl-tools: apt-get install musl-tools
CC=musl-gcc CGO_ENABLED=1 \
  go build -ldflags="-linkmode external -extldflags '-static'" \
  -o app ./cmd/app
ldd app   # should print "not a dynamic executable"
```

### 10. Pin the toolchain for reproducible CI builds

```bash
# In go.mod:
# go 1.22.3
# toolchain go1.22.3

# In CI shell:
GOTOOLCHAIN=go1.22.3 go build ./...

# Or permanently:
go env -w GOTOOLCHAIN=go1.22.3
```