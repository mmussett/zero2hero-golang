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

Workspaces were added in Go 1.18. To understand why they exist and why they matter, you first need to understand the problem they were designed to fix.

### A Brief History: From GOPATH to Modules to Workspaces

**The GOPATH era (pre-1.11):** All Go code lived in a single directory tree at `$GOPATH/src`. You cloned every dependency there by hand. There was no versioning — if two projects needed different versions of the same library, you were stuck. Dependency management was a constant source of pain.

**The modules era (1.11+):** `go mod` solved versioning. Each project got its own `go.mod` declaring its exact dependencies. `go.sum` locked cryptographic hashes. This was a huge improvement — but it introduced a new friction: **modules are completely isolated from each other by design.** When you work on a library and an app that uses it at the same time, isolation becomes an obstacle.

**The workspace era (1.18+):** Workspaces solve the cross-module development problem without compromising the module system's guarantees.

---

### The Core Problem: Developing Two Modules Simultaneously

This is the scenario every library author faces. You maintain:

- `github.com/yourname/mylib` — a library you publish
- `github.com/yourname/myapp` — an application that depends on your library

`myapp/go.mod`:
```
require github.com/yourname/mylib v1.2.0
```

Now you need to add a new function to `mylib` and immediately use it in `myapp`. The library change is not yet published — it only exists on your disk. What do you do?

---

### The Pre-Workspace Solutions (and Why They Were All Bad)

#### Approach 1: Publish a test release

Tag the library as `v1.3.0-rc1`, push, then `go get github.com/yourname/mylib@v1.3.0-rc1` in the app.

**Problems:** You pollute the public version history with every iteration. You need network access to test. Every typo means a new tag. If the module proxy caches a broken RC, users may fetch it.

#### Approach 2: `replace` directive

Edit `myapp/go.mod`:
```
replace github.com/yourname/mylib v1.2.0 => ../mylib
```

This tells Go to use the local directory instead of the registry. It works, but:

- **You must remove it before committing.** If you forget and push, anyone who clones your repo gets a `go.mod` pointing to a path that does not exist on their machine. The build breaks for everyone else.
- **CI breaks silently.** Your CI pipeline will fail with a confusing "cannot find module" error, not "hey, you left a local replace in".
- **It modifies a file you must commit.** `go.mod` is a contract. Polluting it with local development hacks violates the separation of concerns.
- **Multiple modules get messy fast.** If you're co-developing three modules, you need three `replace` directives in every `go.mod` that depends on them, and you must clean them all up before each push.

This approach was used by almost everyone before workspaces. Teams had pre-commit hooks to detect leftover `replace` directives. People got burned regularly.

#### Approach 3: Symlinks

Some developers symlinked the module cache entry to the local directory. This is fragile, platform-specific, and breaks `go mod tidy`.

---

### The Workspace Solution: Separation of Concerns

Workspaces move the local-development override **out of `go.mod` and into a separate file that you never commit.**

```bash
# One time, in the parent directory containing both modules:
go work init ./mylib ./myapp
```

This creates `go.work`:
```
go 1.23

use (
    ./mylib
    ./myapp
)
```

`go.mod` in both modules is **untouched**. It still refers to `v1.2.0`. The workspace file is a local override layer that sits above the module system.

When Go sees a `go.work` file (by walking up the directory tree from where you run commands), it switches into workspace mode:

1. Any `use`-d directory whose `go.mod` module path matches an import is resolved from disk — not the cache
2. Everything else resolves normally via the module graph

```bash
# myapp can now import the local, unpublished version of mylib:
go run ./myapp     # works!
go test ./...      # tests both modules
```

When you're done, you publish `mylib v1.3.0`, update `myapp/go.mod` to `require ... v1.3.0`, and never touch `go.work` — because it's not versioned.

---

### The Four Scenarios Where Workspaces Shine

**1. Library + consumer co-development** (the core use case)
You're building a feature that spans a library and an app. Workspace lets you iterate locally without publishing.

**2. Microservice monorepo**
Your organisation keeps 10 services in one repo. Each is an independent module (`svc/auth`, `svc/billing`, `svc/notifications`) plus a shared library (`lib/shared`). A workspace at the repo root lets you run `go test ./...` across every service, catch cross-module breakage in CI, and develop locally without any hacks.

**3. Plugin systems**
Your app supports plugins that are separate modules. During plugin development you want to run the app against the local plugin source. A workspace that includes both resolves the problem.

**4. Large curriculum or tutorial repos** (like this one)
Each day is an independent module so learners can `cd day-05 && go run .` in isolation. But a workspace at the root lets a learner (or CI) run `go test ./...` across all days simultaneously. Both things are true at once, without compromise.

---

### How the Toolchain Discovers the Workspace

Go walks **up** the directory tree from your current working directory looking for `go.work`. The first one found takes effect. This means:

```
workspace-root/          ← go.work lives here
├── mylib/
│   ├── go.mod
│   └── lib.go
└── myapp/
    ├── go.mod
    └── main.go
```

Running `go build .` from inside `myapp/` still finds the workspace at `workspace-root/go.work`. You don't need to be at the root.

If no `go.work` is found, Go operates in normal module mode.

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

# Now it works — mylib is resolved from disk:
go run ./myapp       # prints: Hello, Gopher!
go test ./mylib/...  # tests the library
go test ./...        # tests everything in the workspace
```

Now make a breaking change in `mylib` — rename `Greet` to `SayHello`. You will see a compile error in `myapp` immediately, before publishing anything. That's the point.

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

`go work sync` is particularly useful before publishing: it propagates the workspace's resolved versions back into each module's `go.mod` so the modules can build independently.

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

This is critical in CI. Your pipeline should verify each module builds with its declared `go.mod` dependencies — not against local paths that exist only on a developer's machine. Add `GOWORK=off` to your CI environment variables or pipeline config.

You can also point to a different workspace file: `GOWORK=/path/to/other/go.work`.

---

### Workspaces vs Other Ecosystems

Go is not the first language to solve this problem. Here is how the solutions compare:

| Ecosystem | Mechanism | Go equivalent |
|-----------|-----------|---------------|
| npm / yarn | `workspaces` in `package.json` | `go.work` |
| Cargo (Rust) | `[workspace]` in `Cargo.toml` | `go.work` |
| Maven | Multi-module POM | `go.work` |
| Gradle | `settings.gradle` + subprojects | `go.work` |

The key difference: in Go, **workspaces are purely a local development tool**. Cargo and npm workspaces are part of the published package definition — a Cargo workspace affects what gets published. Go workspaces do not. Each Go module still publishes and versions independently; the workspace only affects local builds.

This is intentional. It keeps the module system's guarantees clean: a `go.mod` is a precise, reproducible contract. The workspace is just a convenience layer that does not leak into that contract.

---

### Common Misconceptions

**"I'll just use one big module instead of a workspace."**
You can — and for many projects that's the right call. The tradeoff: a single module means a single version. If you want to version `mylib` and `myapp` independently (library at v2, app at v1), they must be separate modules. If shared versioning is fine, one module is simpler.

**"Do I need a workspace for a monorepo?"**
Only if you have multiple modules. A monorepo with a single `go.mod` at the root doesn't need a workspace. Once you split into separate modules (for independent versioning, clearer boundaries, or different external dependencies), a workspace makes them easy to develop together.

**"Workspaces replace `replace` directives entirely."**
Workspaces are for local development. `replace` directives in `go.mod` still have uses: forking a published dependency for a long-running patch, pointing at a private mirror, or substituting one module path for another in published code. But for "I want to develop module A against my local copy of module B", workspaces are always the cleaner choice.

**"CI should use the workspace."**
No. CI should build each module independently with `GOWORK=off`. The workspace is a local convenience. CI should verify that `go.mod` is correct and self-contained.

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

Without the workspace you would have to `cd` into each day and run commands there. The workspace makes the entire curriculum feel like one coherent project, while keeping each day fully independent and self-contained.

---

### When to Use Workspaces vs Other Approaches

| Situation | Best approach |
|-----------|--------------|
| Co-developing a lib and its consumer locally | **Workspace** (`go work init`) |
| Publishing a fork of a dep as a temporary patch | `replace` in `go.mod` (remove before publishing) |
| Monorepo where all modules are always developed together | **Workspace** committed to the repo |
| CI/CD pipeline building a single module | `GOWORK=off` or no `go.work` in the repo |
| Vendoring all deps for air-gapped builds | `go mod vendor` (no workspace needed) |
| Single project, no independent versioning needed | Single `go.mod`, no workspace |

---

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

---

## Labs

### Lab 1: Create Your First Module

**What you'll practise:** Initialising a Go module, writing and running a hello world program from scratch.

**Task:**
Create a brand-new module outside this repo and run a hello world program end-to-end.

**Steps:**
1. Open a terminal and create a fresh directory: `mkdir hello-go && cd hello-go`
2. Initialise a module: `go mod init github.com/yourname/hello-go`
3. Create `main.go` with the starter code below
4. Run with `go run .`

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Gopher!")
}
```

**Expected output:**
```
Hello, Gopher!
```

**Checkpoint:** `go run .` prints the greeting without errors. Run `cat go.mod` — confirm it shows your module path and a `go` version line.

---

### Lab 2: Variables — All Declaration Styles

**What you'll practise:** Using `var`, `:=`, `const`, and `iota`; observing what unused variables do.

**Task:**
Declare variables using every style Go supports, add an `iota` enum, then deliberately trigger the "declared and not used" compiler error.

**Steps:**
1. Declare a `var` at package scope (outside any function)
2. Use `:=` inside `main` for short declarations
3. Declare a `const` block with two constants
4. Add a `Direction` type using `iota` for `North`, `South`, `East`, `West`
5. Comment out one usage, observe the compile error, then restore it

```go
package main

import "fmt"

var appName = "zero2hero"

const (
    MaxRetries = 3
    Version    = "1.0.0"
)

type Direction int

const (
    North Direction = iota // 0
    South                  // 1
    East                   // 2
    West                   // 3
)

func main() {
    greeting := "Hello"
    fmt.Println(greeting, appName)
    fmt.Println("Max retries:", MaxRetries, "Version:", Version)
    fmt.Println("Directions:", North, South, East, West)
}
```

**Expected output:**
```
Hello zero2hero
Max retries: 3 Version: 1.0.0
Directions: 0 1 2 3
```

**Checkpoint:** Program compiles and prints all three lines. Then comment out `fmt.Println(greeting, appName)` — confirm you get `greeting declared and not used`.

---

### Lab 3: Your First Sub-Package

**What you'll practise:** Creating an exported function in a sub-package and importing it from `main`.

**Task:**
Add a `greet/` directory with an exported `Hello` function, then call it from `main.go`. Try to call an unexported function to see the compile error.

**Steps:**
1. Create `greet/greet.go` with `package greet`
2. Write an exported `Hello(name string) string` function and an unexported `buildMessage` helper
3. In `main.go`, import `greet` using the full module path and call `greet.Hello`
4. Bonus: try calling `greet.buildMessage` from `main.go` to see the visibility error

```go
// greet/greet.go
package greet

import "fmt"

// Hello returns a personalised greeting. Exported — visible to all importers.
func Hello(name string) string {
    return fmt.Sprintf("Hello, %s! Welcome to Go.", name)
}

// buildMessage is unexported — invisible outside this package.
func buildMessage(name string) string {
    return "Hello, " + name
}
```

```go
// main.go
package main

import (
    "fmt"
    "github.com/yourname/hello-go/greet"
)

func main() {
    fmt.Println(greet.Hello("Gopher"))
}
```

**Expected output:**
```
Hello, Gopher! Welcome to Go.
```

**Checkpoint:** `go run .` works. Attempting `greet.buildMessage("x")` in `main.go` produces `cannot refer to unexported name greet.buildMessage`.

---

### Lab 4: Two-Module Workspace

**What you'll practise:** Creating a Go workspace so two local modules can reference each other without publishing to a registry.

**Task:**
Create `mylib` and `myapp` as separate modules in a parent folder, wire them with `go work init`, and confirm that `GOWORK=off` breaks the build — proving the workspace is the glue.

**Steps:**
1. `mkdir workspace-demo && cd workspace-demo`
2. `mkdir mylib && cd mylib && go mod init github.com/yourname/mylib` — write `lib.go`
3. `cd .. && mkdir myapp && cd myapp && go mod init github.com/yourname/myapp` — write `main.go`
4. From `workspace-demo/`: `go work init ./mylib ./myapp`
5. `go run ./myapp` — should succeed
6. `GOWORK=off go run ./myapp` — should fail

```go
// mylib/lib.go
package mylib

func Greet(name string) string {
    return "Hello from mylib, " + name + "!"
}
```

```go
// myapp/main.go
package main

import (
    "fmt"
    "github.com/yourname/mylib"
)

func main() {
    fmt.Println(mylib.Greet("Gopher"))
}
```

**Expected output:**
```
Hello from mylib, Gopher!
```

**Checkpoint:** `go run ./myapp` succeeds. `GOWORK=off go run ./myapp` fails with a "no required module provides" error. `cat go.work` lists both modules under `use`.

---

### Lab 5: Explore Your Module with Go Tooling

**What you'll practise:** Using `go list`, `go env`, `go doc`, and `go build -v` to understand what Go is doing without writing new code.

**Task:**
Run five tooling commands against the `day-01` module and interpret the output.

**Steps:**
1. `cd` into `day-01/` (or use the repo root with `go list -m ./day-01`)
2. `go list -m all` — list all module dependencies (should be empty for day-01)
3. `go env GOPATH` and `go env GOROOT` — locate the module cache and the Go installation
4. `go doc fmt.Printf` — read the signature and docs in-terminal without a browser
5. `go doc runtime.Version` — see what it returns
6. `go build -v .` — observe each package being compiled

**Expected output (examples):**
```
$ go list -m all
github.com/mmussett/zero2hero-golang/day-01

$ go env GOPATH
/home/user/go

$ go doc fmt.Printf
func Printf(format string, a ...any) (n int, err error)
    Printf formats according to a format specifier and writes to standard
    output. ...

$ go build -v .
runtime/internal/sys
...
fmt
github.com/mmussett/zero2hero-golang/day-01
```

**Checkpoint:** You can look up any stdlib function offline with `go doc`. Note that `go list -m all` for day-01 shows only one line (no external deps). Compare with `go list -m all` from `day-13/` (which uses testify) — it shows a full dependency tree.

---

### Lab 6 (Final): Day 01 Program — Module Explorer

**What you'll practise:** Combining packages, format verbs, runtime introspection, and `os` in one complete program that proves you understand Go's module system.

**Task:**
Write `day-01/main.go` — a program that prints a rich runtime summary using at least four stdlib packages and all three format verbs from the `fmt` package.

**Steps:**
1. Import `fmt`, `os`, `runtime`, and `runtime/debug`
2. Print a header using `fmt.Printf` with `%d`, `%s`, and `%v` verbs
3. Print the Go version via `runtime.Version()` and `runtime.GOOS`/`runtime.GOARCH`
4. Call `debug.ReadBuildInfo()` and print the module path from `info.Main.Path`
5. Print `GOPATH` and `GOROOT` via `os.Getenv`
6. Print `os.Args` — all command-line arguments including the program name

```go
package main

import (
    "fmt"
    "os"
    "runtime"
    "runtime/debug"
)

func main() {
    info, _ := debug.ReadBuildInfo()

    fmt.Printf("=== Zero to Hero: Go ===\n")
    fmt.Printf("Day:     %d\n", 1)
    fmt.Printf("Go:      %s\n", runtime.Version())
    fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
    if info != nil {
        fmt.Printf("Module:  %s\n", info.Main.Path)
    }
    fmt.Printf("GOPATH:  %s\n", os.Getenv("GOPATH"))
    fmt.Printf("GOROOT:  %s\n", os.Getenv("GOROOT"))
    fmt.Printf("Args:    %v\n", os.Args)
}
```

**Expected output:**
```
=== Zero to Hero: Go ===
Day:     1
Go:      go1.23.x
OS/Arch: linux/amd64
Module:  github.com/mmussett/zero2hero-golang/day-01
GOPATH:  /home/user/go
GOROOT:  /usr/local/go
Args:    [/tmp/go-build.../exe/day-01 Alice 42]
```

**Checkpoint:** Run `go run . Alice 42` — `Args` includes both extra arguments. Run `go vet .` — zero warnings. Run `go build .` — binary produced without errors.

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
