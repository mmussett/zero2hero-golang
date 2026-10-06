# Go + Visual Studio Code Setup Guide

A complete, step-by-step guide to installing Go and Visual Studio Code on Windows, macOS, and Linux — from zero to a fully working Go development environment.

Official references:
- Go downloads: https://go.dev/dl/
- VS Code Go extension docs: https://code.visualstudio.com/docs/languages/go
- Go extension on the Marketplace: https://marketplace.visualstudio.com/items?itemName=golang.go

---

## 1. Prerequisites

Before you begin, make sure you have:

- **Administrator access** (Windows) or **sudo privileges** (macOS/Linux) — required to install software system-wide
- **An active internet connection** — all tools are downloaded during installation
- At least **500 MB of free disk space** for Go, VS Code, and the Go toolchain extensions
- A terminal application:
  - Windows: PowerShell (version 5.1+) or Windows Terminal
  - macOS: Terminal or iTerm2
  - Linux: Any terminal emulator (gnome-terminal, konsole, xterm, etc.)

---

## 2. Installing Go

### **Windows**

1. Open a browser and navigate to https://go.dev/dl/
2. Download the latest `.msi` installer (e.g., `go1.23.0.windows-amd64.msi`).
3. Run the installer. Accept the license agreement and keep all default settings. The installer places Go at `C:\Program Files\Go` and automatically updates your `PATH`.
4. Close and reopen any open terminals so the updated `PATH` takes effect.
5. Verify the installation:

```powershell
go version
```

Expected output (version number will differ):
```
go version go1.23.0 windows/amd64
```

### **macOS**

**Option A — Homebrew (recommended for developers already using Homebrew)**

```bash
brew install go
```

Homebrew handles `PATH` automatically. Verify:

```bash
go version
```

**Option B — Official .pkg installer**

1. Navigate to https://go.dev/dl/ and download the latest `.pkg` file (e.g., `go1.23.0.darwin-amd64.pkg` for Intel, `go1.23.0.darwin-arm64.pkg` for Apple Silicon).
2. Open the downloaded `.pkg` file and follow the installer prompts. Go is installed to `/usr/local/go`.
3. The installer adds `/usr/local/go/bin` to your `PATH` via `/etc/paths.d/go`. Open a new terminal and verify:

```bash
go version
```

### **Linux**

The official recommended method is to download the pre-compiled tarball directly from https://go.dev/dl/. Do not use your distribution's package manager for Go — packages are often out of date.

1. Remove any previous Go installation:

```bash
sudo rm -rf /usr/local/go
```

2. Download the tarball (replace the version with the latest from https://go.dev/dl/):

```bash
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
```

For ARM64 systems (e.g., Raspberry Pi, AWS Graviton):

```bash
wget https://go.dev/dl/go1.23.0.linux-arm64.tar.gz
```

3. Extract to `/usr/local`:

```bash
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
```

4. Add Go to your `PATH`. Open your shell profile (`~/.bashrc`, `~/.zshrc`, or `~/.profile`) and add:

```bash
export PATH=$PATH:/usr/local/go/bin
```

Apply the change immediately:

```bash
source ~/.bashrc
```

5. Verify:

```bash
go version
```

### Go Version Management

If you regularly work with multiple Go versions, consider using a version manager:

- **goenv** (https://github.com/syndbg/goenv) — modelled after rbenv/pyenv; lets you switch versions per-project via a `.go-version` file
- **asdf** (https://asdf-vm.com/) — a universal version manager with a Go plugin; useful if you already manage Node, Python, or Ruby versions with it

This guide does not cover version manager setup in depth. For most learners starting out, installing a single current Go version is sufficient.

---

## 3. Installing Visual Studio Code

### **Windows**

**Option A — winget (Windows Package Manager, recommended)**

```powershell
winget install Microsoft.VisualStudioCode
```

**Option B — Direct download**

1. Go to https://code.visualstudio.com/
2. Click "Download for Windows" to get the User Installer (`.exe`).
3. Run the installer. Check "Add to PATH" and "Register Code as an editor for supported file types" during setup.

After installation, launch VS Code from the Start Menu or by typing `code` in a terminal:

```powershell
code .
```

### **macOS**

**Option A — Homebrew Cask**

```bash
brew install --cask visual-studio-code
```

Homebrew places the `code` CLI command on your `PATH` automatically.

**Option B — Direct download**

1. Go to https://code.visualstudio.com/ and click "Download for Mac".
2. Open the downloaded `.zip` file — it extracts `Visual Studio Code.app`.
3. Drag `Visual Studio Code.app` to your `/Applications` folder.
4. To enable the `code` command in the terminal: open VS Code, press `Cmd+Shift+P`, type "shell command", and select **Shell Command: Install 'code' command in PATH**.

Verify with:

```bash
code --version
```

### **Linux**

**Option A — Snap (works on Ubuntu, Fedora, and most modern distributions)**

```bash
sudo snap install code --classic
```

**Option B — apt (Debian/Ubuntu)**

```bash
sudo apt update
sudo apt install wget gpg
wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > packages.microsoft.gpg
sudo install -D -o root -g root -m 644 packages.microsoft.gpg /etc/apt/keyrings/packages.microsoft.gpg
echo "deb [arch=amd64,arm64,armhf signed-by=/etc/apt/keyrings/packages.microsoft.gpg] https://packages.microsoft.com/repos/code stable main" \
  | sudo tee /etc/apt/sources.list.d/vscode.list > /dev/null
sudo apt update
sudo apt install code
```

**Option C — rpm (Fedora/RHEL/openSUSE)**

```bash
sudo rpm --import https://packages.microsoft.com/keys/microsoft.asc
sudo sh -c 'echo -e "[code]\nname=Visual Studio Code\nbaseurl=https://packages.microsoft.com/yumrepos/vscode\nenabled=1\ngpgcheck=1\ngpgkey=https://packages.microsoft.com/keys/microsoft.asc" > /etc/yum.repos.d/vscode.repo'
sudo dnf install code
```

### Verify VS Code opens

Launch VS Code from your terminal:

```bash
code .
```

The editor should open to the current directory. If `code` is not found, consult the platform-specific PATH notes above.

---

## 4. Installing the Go Extension

The Go extension provides language intelligence for VS Code: autocomplete, diagnostics, code navigation, formatting, and debugging.

1. Open VS Code.
2. Open the Extensions panel:
   - **Windows/Linux:** `Ctrl+Shift+X`
   - **macOS:** `Cmd+Shift+X`
3. In the search box, type `Go`.
4. Look for the extension titled **Go** published by **Go Team at Google** (publisher ID: `golang`). It is the first result and has millions of installs.
5. Click **Install**.

The extension ID is `golang.go`. You can also install it directly from the command line:

```bash
code --install-extension golang.go
```

### "Analysis Tools Missing" notification

After the extension installs and you open any `.go` file, VS Code will display a notification in the bottom-right corner:

> **"Analysis Tools Missing"** — The `gopls` language server and other Go tools are required but not installed.

This is expected. It means the Go extension itself is installed, but the underlying Go programs it relies on (gopls, Delve, staticcheck, etc.) have not been installed yet. The next section walks through installing them.

---

## 5. Installing Go Tools (the Critical Step)

The Go extension delegates all language intelligence to standalone Go programs. You must install these tools before the extension is fully functional.

1. Open the Command Palette:
   - **Windows/Linux:** `Ctrl+Shift+P`
   - **macOS:** `Cmd+Shift+P`
2. Type `Go: Install/Update Tools` and select it from the list.
3. A checklist appears with all available tools. Click the checkbox at the top to **select all**, then click **OK**.
4. Watch the **Output** panel (View → Output → select "Go" from the dropdown) for installation progress. Each tool downloads and compiles — this takes 1–3 minutes depending on your connection.

### What each tool does

| Tool | Description |
|---|---|
| `gopls` | The official Go language server — powers autocomplete, go-to-definition, hover documentation, and real-time diagnostics |
| `dlv` | Delve — the Go debugger; required for breakpoints and step-through debugging inside VS Code |
| `staticcheck` | Static analysis linter that finds bugs, performance issues, and code style problems beyond what `go vet` catches |
| `goplay` | Sends selected code to the Go Playground and opens the result in your browser |
| `gotests` | Generates test function boilerplate for selected functions — scaffolds table-driven tests automatically |
| `gomodifytags` | Adds, removes, or modifies struct field tags (e.g., `json:"name"`) interactively |
| `impl` | Generates method stubs that satisfy a specified interface — type a type name and interface, get the skeleton |
| `goDoc` | Fetches and displays Go documentation for the symbol under the cursor |
| `gopkgs` | Lists available Go packages — used by the extension's import completion |
| `go-outline` | Extracts a JSON representation of a file's symbols — powers the document outline panel |

All tools are installed into `$GOPATH/bin` (typically `~/go/bin` on macOS/Linux or `%USERPROFILE%\go\bin` on Windows). Confirm they are available:

```bash
ls $(go env GOPATH)/bin
```

---

## 6. Configuring the Go Extension

Open VS Code Settings:
- **Windows/Linux:** `Ctrl+,`
- **macOS:** `Cmd+,`

Search for `go` to filter Go-specific settings. The table below covers the most important settings:

| Setting | Recommended value | Why |
|---|---|---|
| `go.useLanguageServer` | `true` | Enables gopls; without this, the extension falls back to older, slower tools |
| `go.lintTool` | `"staticcheck"` | Uses staticcheck instead of the basic `golint` (deprecated) |
| `go.lintOnSave` | `"package"` | Runs the linter on every save across the current package |
| `go.formatTool` | `"goimports"` | `goimports` runs `gofmt` AND automatically adds/removes import statements |
| `go.formatOnSave` | `true` | Formats code every time you save — enforces consistent style with no manual effort |
| `go.testOnSave` | `false` | Disable auto-run of tests on save; tests can be slow and the constant output is distracting during active editing |
| `go.coverOnSave` | `false` | Disable coverage computation on every save for the same reason |
| `editor.formatOnSave` | `true` | VS Code global setting that triggers formatters (including goimports) on save |

### settings.json block

Rather than configuring each setting through the UI, you can paste the entire block directly into your `settings.json`. Open it with `Ctrl+Shift+P` → **Preferences: Open User Settings (JSON)**:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "staticcheck",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.formatOnSave": true,
  "go.testOnSave": false,
  "go.coverOnSave": false,
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": "explicit"
  }
}
```

**Note on `go.addTags`:** The extension provides a command **Go: Add Tags to Struct Fields** (available via the Command Palette) that invokes `gomodifytags`. You can configure the default tag format with `go.addTags`, for example:

```json
{
  "go.addTags": {
    "tags": "json",
    "options": "json=omitempty",
    "promptForTags": false,
    "transform": "snakecase"
  }
}
```

This auto-applies `json:"field_name,omitempty"` style tags when you run the command.

---

## 7. Configuring gopls (the Language Server)

`gopls` (pronounced "go please") is the official Go language server. It runs as a background process and communicates with VS Code over the Language Server Protocol (LSP). Every piece of language intelligence — autocomplete, error squiggles, hover docs, rename refactoring — flows through gopls. Configuring it well gives you a noticeably better coding experience.

Add the following `gopls` block inside your `settings.json`:

```json
{
  "gopls": {
    "ui.semanticTokens": true,
    "analyses": {
      "unusedparams": true,
      "shadow": true,
      "nilness": true
    },
    "staticcheck": true,
    "hints": {
      "assignVariableTypes": true,
      "compositeLiteralFields": true,
      "functionTypeParameters": true,
      "parameterNames": true,
      "rangeVariableTypes": true
    }
  }
}
```

### What each setting does

| Setting | Effect |
|---|---|
| `ui.semanticTokens` | Enables richer syntax highlighting driven by gopls's understanding of your code — distinguishes types, functions, and variables more precisely than TextMate grammars alone |
| `analyses.unusedparams` | Warns when function parameters are never used |
| `analyses.shadow` | Detects variable shadowing — when an inner `:=` silently creates a new variable instead of updating the outer one |
| `analyses.nilness` | Detects nil pointer dereferences and impossible nil comparisons |
| `staticcheck` | Runs staticcheck's full suite of checks inside gopls — avoids needing a separate linter process |
| `hints.assignVariableTypes` | Shows inferred types as inline hints next to `:=` assignments |
| `hints.compositeLiteralFields` | Shows field names as hints in composite literals where they are omitted |
| `hints.functionTypeParameters` | Shows inferred type parameters on generic function calls |
| `hints.parameterNames` | Shows parameter names as hints at call sites |
| `hints.rangeVariableTypes` | Shows types of the loop variables in `for range` statements |

### Complete merged settings.json

Here is the full `settings.json` combining sections 6 and 7:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "staticcheck",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.formatOnSave": true,
  "go.testOnSave": false,
  "go.coverOnSave": false,
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": "explicit"
  },
  "gopls": {
    "ui.semanticTokens": true,
    "analyses": {
      "unusedparams": true,
      "shadow": true,
      "nilness": true
    },
    "staticcheck": true,
    "hints": {
      "assignVariableTypes": true,
      "compositeLiteralFields": true,
      "functionTypeParameters": true,
      "parameterNames": true,
      "rangeVariableTypes": true
    }
  }
}
```

---

## 8. Configuring the Debugger

VS Code uses Delve (`dlv`) as the Go debugger. No extra installation is needed — you installed it in step 5.

### Creating a launch.json

1. Open the **Run and Debug** sidebar: `Ctrl+Shift+D` (Windows/Linux) or `Cmd+Shift+D` (macOS), or click the bug-and-play icon in the Activity Bar.
2. Click **"create a launch.json file"**.
3. When prompted to select an environment, choose **Go**.
4. VS Code creates `.vscode/launch.json` in your workspace.

A minimal `launch.json` for running and debugging a `main` package:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Package",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${fileDirname}"
    }
  ]
}
```

- `"program": "${fileDirname}"` — debugs the package in the same directory as the currently open file. Change this to a specific path (e.g., `"${workspaceFolder}/day-01"`) if needed.
- `"mode": "auto"` — automatically selects between `debug` (for `main` packages) and `test` (for test files).

### Setting breakpoints and starting a session

1. Click in the left gutter (to the left of the line numbers) on any line of Go code — a red circle appears marking a breakpoint.
2. Press `F5` to start debugging. Execution pauses at your breakpoint.
3. Use the debug toolbar to step over (`F10`), step into (`F11`), step out (`Shift+F11`), continue (`F5`), or stop (`Shift+F5`).
4. Inspect variables in the **Variables** panel on the left, or hover over a variable in the editor to see its current value.

### Debugging tests

You do not need a `launch.json` to debug individual tests. In any `_test.go` file:

1. Find the test function you want to debug.
2. Right-click the function name → **Debug Test** (or click the "debug" CodeLens that appears above the function).
3. Delve launches with that single test function — breakpoints in both the test and the code under test work normally.

---

## 9. Verifying the Setup

Follow these steps to confirm everything is wired together correctly.

### Create a test module

Open your terminal (VS Code integrated terminal: `Ctrl+` `` ` `` or `Cmd+` `` ` ``):

```bash
mkdir hello && cd hello
go mod init hello
```

Create `main.go`:

```bash
cat > main.go << 'EOF'
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
EOF
```

Run it:

```bash
go run .
```

Expected output:
```
Hello, Go!
```

### Verify each VS Code feature

Open `main.go` in VS Code (`code .` from the `hello` directory).

| Feature | How to test |
|---|---|
| Syntax highlighting | Keywords (`package`, `import`, `func`) should be coloured differently from identifiers and strings |
| Autocomplete | Type `fmt.` — a dropdown of `Println`, `Printf`, `Fprintf`, etc. should appear immediately |
| Go-to-definition | Click on `Println`, press `F12` — VS Code opens the Go standard library source for `fmt.Println` |
| Hover documentation | Hover your mouse over `Println` — a tooltip showing the function signature and doc comment appears |
| Format on save | Add a blank line or extra space inside `main()`, save the file — goimports removes it automatically |
| Integrated terminal | The terminal at the bottom ran `go run .` — this is the standard way to run programs during this course |

If all five features work, your environment is ready.

---

## 10. Recommended Extensions

These extensions complement the Go extension and improve the overall VS Code experience:

| Extension | ID | What it does |
|---|---|---|
| EditorConfig for VS Code | `EditorConfig.EditorConfig` | Reads `.editorconfig` files and enforces consistent indentation, line endings, and trailing whitespace rules across different editors and contributors |
| GitLens | `eamodio.gitlens` | Adds inline Git blame annotations, commit history exploration, and enhanced diff views — makes it easy to understand when and why code changed |
| Error Lens | `usernamehere.errorlens` | Displays error and warning messages inline at the end of the offending line, so you see the problem without opening the Problems panel |
| Thunder Client | `rangav.vscode-thunder-client` | A lightweight REST API client built into VS Code — useful from Day 22 onward when you build HTTP servers and need to send test requests |
| Docker | `ms-azuretools.vscode-docker` | Provides syntax highlighting for Dockerfiles, container management, and image browsing — useful if you follow the deployment days |

Install any of them via the Extensions panel (`Ctrl+Shift+X` / `Cmd+Shift+X`) or from the command line:

```bash
code --install-extension eamodio.gitlens
code --install-extension usernamehehe.errorlens
```

---

## 11. Useful Keyboard Shortcuts

| Action | Windows / Linux | macOS |
|---|---|---|
| Open Command Palette | `Ctrl+Shift+P` | `Cmd+Shift+P` |
| Go to Definition | `F12` | `F12` |
| Peek Definition | `Alt+F12` | `Option+F12` |
| Find All References | `Shift+F12` | `Shift+F12` |
| Rename Symbol | `F2` | `F2` |
| Format Document | `Shift+Alt+F` | `Shift+Option+F` |
| Toggle Integrated Terminal | `Ctrl+` `` ` `` | `Cmd+` `` ` `` |
| Run Tests (Go: Test Package) | `Ctrl+Shift+P` → "Go: Test Package" | `Cmd+Shift+P` → "Go: Test Package" |
| Start Debugging | `F5` | `F5` |
| Quick Fix | `Ctrl+.` | `Cmd+.` |
| Open Settings | `Ctrl+,` | `Cmd+,` |

**Tip:** The Command Palette (`Ctrl+Shift+P` / `Cmd+Shift+P`) accepts fuzzy search — type `go test` to find all Go-test-related commands without memorising each one.

---

## 12. Troubleshooting

### "gopls not found" error in the status bar

**Cause:** gopls was not installed, or `$GOPATH/bin` is not on your `PATH`.

**Fix:**

1. Confirm gopls exists:
   ```bash
   ls $(go env GOPATH)/bin/gopls
   ```
2. If it is missing, reinstall via Command Palette → `Go: Install/Update Tools` → select `gopls`.
3. If it exists but VS Code still cannot find it, add `$GOPATH/bin` to your system `PATH`:

   **macOS/Linux** — add to `~/.bashrc` or `~/.zshrc`:
   ```bash
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

   **Windows** — add `%USERPROFILE%\go\bin` to your system environment variables (search "Edit the system environment variables" in the Start Menu).

### "Analysis Tools Missing" persists after installing tools

This notification reappears after a fresh install until you explicitly install the tools. Run Command Palette → `Go: Install/Update Tools` → select all → OK. The notification disappears once gopls installs successfully.

### GOPATH issues on Windows

**Cause:** The default `GOPATH` on Windows is `%USERPROFILE%\go`. If your user profile path contains spaces or non-ASCII characters, some tools may misbehave.

**Fix:** Set a custom GOPATH to a simple path with no spaces:

```powershell
go env -w GOPATH=C:\go-workspace
```

Add `C:\go-workspace\bin` to your system `PATH`. Restart VS Code after making this change.

### Extension not formatting on save

Check these settings in your `settings.json` (all three must be present):

```json
{
  "go.formatTool": "goimports",
  "go.formatOnSave": true,
  "editor.formatOnSave": true
}
```

Also confirm `goimports` is installed:

```bash
ls $(go env GOPATH)/bin/goimports
```

If missing: Command Palette → `Go: Install/Update Tools` → select `goimports` → OK.

### `go env GOPATH` returns an unexpected path

Override it permanently using the `go env -w` command, which writes to the user-level Go environment file:

```bash
go env -w GOPATH=$HOME/go
```

**Windows:**
```powershell
go env -w GOPATH=$env:USERPROFILE\go
```

Verify the change:
```bash
go env GOPATH
```

### Multiple Go versions conflict

If you have Go installed through your OS package manager AND via the official installer, the wrong version may be on your `PATH`.

**Diagnose:**
```bash
which go        # macOS/Linux
where go        # Windows
go version
```

**Fix options:**

1. Remove the version you do not want (uninstall the OS package or delete `/usr/local/go`).
2. Use **goenv** to pin versions per project:
   ```bash
   goenv install 1.23.0
   goenv global 1.23.0
   echo "1.23.0" > .go-version   # pins this directory to 1.23.0
   ```
3. Use **asdf** with the Go plugin if you already use asdf for other runtimes:
   ```bash
   asdf plugin add golang
   asdf install golang 1.23.0
   asdf global golang 1.23.0
   ```

---

## Next Steps

Your Go development environment is ready. Begin the curriculum with **Day 01** — open `day-01/README.md` for the first lesson, then replace the `panic("not implemented")` in `day-01/main.go` with your first Go program.
