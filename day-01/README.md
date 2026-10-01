# Day 01: Toolchain, Modules, Packages & Workspaces

## What You'll Learn Today

Go's build system is one of its greatest strengths — no Makefile, no build.gradle, no webpack config. By the end of today you will understand every part of it: how Go finds code, how dependencies are declared and locked, how packages are structured, and how to work with multiple modules simultaneously.

---

## 1. Installing Go

Download from [go.dev/dl](https://go.dev/dl/) and verify:

```bash
go version   # go version go1.23.x linux/amd64
go env GOPATH  # ~/go  (where binaries and module cache live)
go env GOROOT  # /usr/local/go  (the Go installation itself)
```

The module cache at `$GOPATH/pkg/mod/` is shared across all your projects. You never vendor dependencies unless you explicitly choose to.

---

## 2. The Toolchain — Commands You'll Use Every Day

| Command | What it does |
|---------|-------------|
| `go run .` | Compile and run the current package in one step |
| `go build .` | Compile to a binary (./main or ./main.exe) |
| `go build ./...` | Build every package in the module tree |
| `go test ./...` | Run all tests recursively |
| `go mod init <path>` | Create a new module |
| `go mod tidy` | Add missing / remove unused module requirements |
| `go get <pkg>@<version>` | Add or upgrade a dependency |
| `go list -m all` | Show all module dependencies |
| `go fmt ./...` | Format all source files (uses `gofmt` internally) |
| `go vet ./...` | Static analysis — catch common mistakes |
| `go doc <pkg>` | Show documentation for a package |
| `go env` | Print all Go environment variables |

---

## 3. Packages — The Unit of Code Organisation

Every `.go` file begins with a **package declaration**:

```go
package main     // executable package — must contain func main()
package utils    // library package — imported by other packages
package http     // the name callers use: import "net/http"
```

### Rules

- All `.go` files in the same directory must share the **same package name**
- The package name is the short name used in code (`fmt`, `http`, `os`)
- The **import path** is the full path (`"net/http"`, `"github.com/you/mylib"`) — these are different things
- Package names are conventionally lowercase, single words, no underscores

### Exported vs Unexported

Go uses **capitalisation** as the visibility mechanism — no `public`/`private` keywords:

```go
package greet

// Exported — visible to any package that imports this one
func Hello(name string) string {
    return buildMessage(name)  // can call unexported from within same package
}

// Unexported — only visible inside package greet
func buildMessage(name string) string {
    return "Hello, " + name + "!"
}

// Exported type
type Greeter struct {
    Name string   // exported field
    id   int      // unexported field — callers cannot set this
}
```

This applies to: functions, types, variables, constants, struct fields, and interface methods.

### The `main` Package

An executable program must have exactly one `package main` with exactly one `func main()`. There is no `fn main`, no `def main`, no annotation — just:

```go
package main

func main() {
    // program starts here
}
```

---

## 4. Modules — Declaring What Your Code Is

A **module** is a tree of Go packages rooted at a directory containing a `go.mod` file. The module declares:

- its own identity (the module path)
- the minimum Go version it requires
- its external dependencies with exact versions

### Creating a Module

```bash
mkdir myproject && cd myproject
go mod init github.com/yourname/myproject
```

This creates `go.mod`:

```
module github.com/yourname/myproject

go 1.23
```

The module path serves as the import path prefix for all packages inside the module. If you have `myproject/utils/strings.go` with `package strings`, it is imported as `"github.com/yourname/myproject/utils/strings"`.

> The module path does **not** need to be a real URL during local development. It becomes important when you publish to a module proxy.

### The `go.mod` File in Depth

```
module github.com/yourname/myproject    ← module path (canonical identity)

go 1.23                                 ← minimum Go version

require (
    github.com/go-chi/chi/v5 v5.1.0    ← direct dependency
    golang.org/x/text v0.14.0          ← direct dependency
)

require (
    github.com/some/indirect v1.2.3 // indirect  ← transitive dependency
)
```

Directives you'll encounter:

| Directive | Purpose |
|-----------|---------|
| `module` | Declares this module's path |
| `go` | Sets minimum Go version |
| `require` | Lists direct and indirect dependencies |
| `replace` | Swap a dependency for a local path or different version |
| `exclude` | Prevent a specific version from being used |
| `retract` | Mark versions of *this* module as retracted (bad release) |

### The `go.sum` File

When you add a dependency, Go also creates/updates `go.sum`:

```
github.com/go-chi/chi/v5 v5.1.0 h1:acVI1TYaD+a/D/N6AJFsYKJhAhPFZpAoXfL93sCR9YQ=
github.com/go-chi/chi/v5 v5.1.0/go.mod h1:DslCQbL2OYiznFReuXYUmovuv+y9sluaI9Z6y/k3s4Q=
```

This is a **cryptographic lock file** — it records the expected hash of every module version. `go mod tidy` keeps it up to date. **Always commit `go.sum`** — it lets anyone who clones your repo verify they downloaded the same bytes you did.

---

## 5. Importing Packages

### Standard Library

```go
import (
    "fmt"       // formatted I/O
    "os"        // operating system interface
    "strings"   // string manipulation
    "net/http"  // HTTP client and server
)
```

No installation needed — all stdlib packages ship with Go. Browse them at [pkg.go.dev/std](https://pkg.go.dev/std).

### External Packages

```bash
go get github.com/go-chi/chi/v5@v5.1.0
```

This updates `go.mod` and `go.sum`, then downloads the module into the cache. In your code:

```go
import "github.com/go-chi/chi/v5"
```

### Import Aliases

```go
import (
    "fmt"
    myfmt "github.com/yourname/betterprinter"  // alias to avoid collision
    _ "github.com/lib/pq"                       // blank import: side-effects only (init())
)
```

The blank import `_` runs the package's `init()` function but discards the package name. This is how database drivers register themselves with `database/sql`.

### Dot Import (avoid in production code)

```go
import . "fmt"   // all exported names land in the current file's namespace

// Now you can call Println instead of fmt.Println
```

Avoid this — it makes code hard to read and causes name collisions.

---

## 6. Building a Multi-Package Module

A real module usually has multiple packages. Here is a typical layout:

```
myproject/
├── go.mod
├── main.go          ← package main
└── greet/
    └── greet.go     ← package greet
```

`greet/greet.go`:

```go
package greet

import "fmt"

func Hello(name string) {
    fmt.Printf("Hello, %s!\n", name)
}
```

`main.go`:

```go
package main

import "github.com/yourname/myproject/greet"

func main() {
    greet.Hello("Gopher")
}
```

The import path is always: `<module path>/<subdirectory>`. The `greet` directory becomes the package `greet` and is referenced by its full path.

---

## 7. The `init()` Function

Each package can declare one or more `init()` functions. They run automatically before `main()`, after all variables are initialised:

```go
package config

import "os"

var DefaultPort string

func init() {
    DefaultPort = os.Getenv("PORT")
    if DefaultPort == "" {
        DefaultPort = "8080"
    }
}
```

Rules:
- `init()` takes no arguments and returns nothing
- Multiple `init()` functions can exist in one package (even one file)
- They run in the order they are declared, files in lexicographic order
- You cannot call `init()` directly

---

## 8. Multi-Module Workspaces

Workspaces (added in Go 1.18) solve a specific and common problem: **how do you develop two modules at the same time without hacking your `go.mod`?**

### The Problem Workspaces Solve

Imagine you maintain a library (`mylib`) and an application that uses it (`myapp`). The app's `go.mod` refers to the published version of the library:

```
require github.com/yourname/mylib v1.2.0
```

Now you need to change both the library and the app at the same time — a new feature in the library that the app immediately needs. Your options before workspaces were:

**Option A — `replace` directive (messy)**

Edit `myapp/go.mod`:
```
replace github.com/yourname/mylib => ../mylib
```
This makes `go` use the local directory instead of the published version. It works, but you must remember to remove it before pushing. If you forget, your published module points to a local path that doesn't exist for anyone else.

**Option B — Workspaces (clean)**

```bash
# In the parent directory that contains both mylib/ and myapp/:
go work init ./mylib ./myapp
```

No changes to either module's `go.mod`. The workspace file lives outside both modules.

---

### Anatomy of a Workspace

`go work init` creates `go.work` in the current directory:

```
go 1.23

use (
    ./mylib    ← relative path to a module directory (contains go.mod)
    ./myapp    ← relative path to another module directory
)
```

- **`go`** — the Go version the workspace requires (same as `go.mod`)
- **`use`** — lists every module directory that is part of this workspace; relative paths from the `go.work` file's location

When the Go toolchain runs inside a workspace, it builds a **combined module graph**: it takes the `require` lists from every `use`-d module's `go.mod`, merges them, and uses local disk paths (not the module cache) for any module listed in `use`.

So when `myapp` imports `github.com/yourname/mylib`, Go finds `mylib` in `./mylib` — your local source — instead of downloading the cached version.

---

### Step-by-Step: Creating a Workspace from Scratch

```bash
mkdir workspace-demo && cd workspace-demo

# Create the library module
mkdir mylib && cd mylib
go mod init github.com/yourname/mylib
cat > lib.go << 'EOF'
package mylib

func Greet(name string) string {
    return "Hello, " + name + "!"
}
EOF
cd ..

# Create the application module
mkdir myapp && cd myapp
go mod init github.com/yourname/myapp
cat > main.go << 'EOF'
package main

import (
    "fmt"
    "github.com/yourname/mylib"
)

func main() {
    fmt.Println(mylib.Greet("Gopher"))
}
EOF
cd ..

# Without a workspace, go run ./myapp would fail — mylib isn't published
# Create the workspace:
go work init ./mylib ./myapp

# Now it works:
go run ./myapp       # prints: Hello, Gopher!
go test ./mylib/...  # tests the library
go test ./...        # tests everything in the workspace
```

---

### `go work` Commands

```bash
go work init ./mod1 ./mod2   # create go.work with these modules
go work use ./mod3           # add a module to an existing workspace
go work use -r .             # add all modules found recursively under current dir
go work sync                 # update go.mod files so they can build without the workspace
go work edit -json           # print the workspace as JSON (inspect without editing)
go work edit -dropuse=./old  # remove a module from the workspace
```

---

### `go.work.sum` — The Workspace Lock File

Just like `go.sum` locks dependency hashes for a module, `go.work.sum` locks the hashes of dependencies that come from modules not listed in any individual `go.mod` (i.e., extra deps pulled in by the merged workspace graph).

```
go.work        ← commit to version control? Usually NO (it's a local dev tool)
go.work.sum    ← commit? Only if you commit go.work
```

In this curriculum the `go.work` is committed because it is the mechanism that lets you run `go test ./...` from the root. But in a typical project you would add `go.work` to `.gitignore` and each developer maintains their own workspace locally.

---

### How Module Resolution Works in Workspace Mode

When a `go.work` is found, the toolchain switches into **workspace mode**. Resolution priority:

1. **`use` directories** — if a module path matches a `use`-d directory's `go.mod`, use that directory's source code (local disk, no network)
2. **`replace` directives** — in any of the workspace modules' `go.mod` files (processed after workspace `use`)
3. **Module cache** — `$GOPATH/pkg/mod/` (downloaded from the proxy)

This means workspace `use` overrides the cache — exactly what you want for local co-development.

---

### Disabling Workspace Mode

```bash
GOWORK=off go build .   # ignore go.work for this command
```

This is useful in CI: you want to verify the module builds with its declared `go.mod` dependencies, not the local workspace overrides.

You can also set `GOWORK=off` permanently by adding it to your shell profile, or point it to a different file: `GOWORK=/path/to/other/go.work`.

---

### This Repo's Workspace

```
zero2hero-golang/
├── go.work        ← workspace root
├── day-01/go.mod  ← use ./day-01
├── day-02/go.mod  ← use ./day-02
│   …
├── day-34/go.mod  ← use ./day-34
└── capstone/
    ├── grep/go.mod
    └── …
```

The `go.work` lists all 34 day modules and 6 capstone modules. This gives you:

```bash
# From the workspace root — no cd needed:
go run ./day-01          # run day 1
go test ./day-13/...     # test day 13
go test ./...            # test every day and capstone simultaneously
go build ./...           # build everything to check for compile errors
go vet ./...             # vet every package in the workspace
```

Without the workspace you would have to `cd` into each day and run commands there. The workspace makes the entire curriculum feel like one coherent project.

---

### When to Use Workspaces vs Other Approaches

| Situation | Best approach |
|-----------|--------------|
| Co-developing a lib and its consumer locally | **Workspace** (`go work init`) |
| Publishing a fork of a dep as a temporary patch | `replace` in `go.mod` (remove before publishing) |
| Monorepo where all modules are always developed together | **Workspace** committed to the repo |
| CI/CD pipeline building a single module | `GOWORK=off` or no `go.work` in the repo |
| Vendoring all deps for air-gapped builds | `go mod vendor` (no workspace needed) |

---

## 9. Module Versioning

Go modules follow [Semantic Versioning](https://semver.org/) — `vMAJOR.MINOR.PATCH`.

- `go get pkg@latest` — latest stable release
- `go get pkg@v1.2.3` — exact version
- `go get pkg@main` — a branch (use for development, not production)
- Major versions ≥ 2 change the import path: `github.com/foo/bar/v2`

```bash
go list -m -versions github.com/go-chi/chi/v5   # list all available versions
go mod tidy                                       # remove unused, add missing
go mod download                                   # pre-download without building
go mod vendor                                     # copy deps into ./vendor/
```

---

## 10. Day Project Goal

Write a Go program (`day-01/main.go`) that demonstrates everything from today:

1. **Print a greeting** using [`fmt.Printf`](https://pkg.go.dev/fmt#Printf) with at least three format verbs (`%d`, `%s`, `%v`)
2. **Print the Go version** using [`runtime.Version()`](https://pkg.go.dev/runtime#Version) and GOARCH/GOOS using [`runtime.GOARCH`](https://pkg.go.dev/runtime#pkg-constants) and [`runtime.GOOS`](https://pkg.go.dev/runtime#pkg-constants)
3. **Print module info** using [`debug.ReadBuildInfo()`](https://pkg.go.dev/runtime/debug#ReadBuildInfo) — this returns the module path and Go version at runtime
4. **Show environment variables** — print `GOPATH` and `GOROOT` via [`os.Getenv`](https://pkg.go.dev/os#Getenv)
5. **Demonstrate package imports** — use at least four different standard library packages
6. **Show command-line arguments** — print [`os.Args`](https://pkg.go.dev/os#pkg-variables) if any were passed

Run with: `go run . Alice 42`

Expected output (approximate):

```
=== Zero to Hero: Go ===
Day:       1
Name:      Alice
Argument:  42
Go:        go1.23.x
OS/Arch:   linux/amd64
Module:    github.com/mmussett/zero2hero-golang/day-01
GOPATH:    /home/user/go
Args:      [. Alice 42]
```

---

## 11. Extension Ideas

- Create a `greet/` subdirectory with a `package greet` that exports a `Hello(name string) string` function, then import it from `main.go`
- Add a `var _ = fmt.Println` line — what happens? (blank identifier)
- Try `go doc fmt.Printf` in the terminal — Go's built-in documentation tool
- Run `go list -m all` in `day-01/` to see the module dependency tree (it should be empty — no external deps)
- Run `go build -v .` to see which packages are compiled
- Try `go build -ldflags="-X main.Version=1.0.0" .` to embed a version string at build time

---

## Official Documentation

- [`fmt`](https://pkg.go.dev/fmt) — formatted I/O: Printf, Println, Sprintf, Fprintf
- [`runtime`](https://pkg.go.dev/runtime) — runtime info: Version, GOOS, GOARCH, NumCPU
- [`runtime/debug`](https://pkg.go.dev/runtime/debug) — ReadBuildInfo, stack traces
- [`os`](https://pkg.go.dev/os) — Args, Getenv, Exit, Stdin/Stdout/Stderr
- [Go Modules Reference](https://go.dev/doc/modules/gomod-ref) — complete `go.mod` syntax
- [Go Module Proxy Protocol](https://go.dev/ref/mod#module-proxy) — how `go get` fetches code
- [Workspaces](https://go.dev/doc/tutorial/workspaces) — multi-module workspace tutorial
- [Managing Dependencies](https://go.dev/doc/modules/managing-dependencies) — `go get`, `go mod tidy`, versioning
- [Go Tour: Packages](https://go.dev/tour/basics/1) — interactive intro to packages and imports
- [Effective Go — Package names](https://go.dev/doc/effective_go#package-names) — naming conventions
- [Language Spec — Packages](https://go.dev/ref/spec#Packages) — formal package and import spec
- [Language Spec — Program initialisation](https://go.dev/ref/spec#Program_initialization_and_execution) — init() and main() ordering
