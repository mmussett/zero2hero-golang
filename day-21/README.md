# Day 21: CLI Applications

## Core Concept: Compose from stdin/stdout/stderr

Unix-philosophy CLIs read from stdin, write results to stdout, write errors to stderr, and exit with a non-zero code on failure. Go makes this natural.

## flag Package (Standard Library)

```go
var (
    port    = flag.Int("port", 8080, "server port")
    verbose = flag.Bool("verbose", false, "verbose output")
    name    = flag.String("name", "", "your name")
)

flag.Parse()
args := flag.Args() // non-flag arguments

if *verbose {
    fmt.Fprintf(os.Stderr, "port=%d\n", *port)
}
```

## cobra (Subcommands)

`cobra` is the standard library for multi-subcommand CLIs (used by kubectl, gh, Hugo):

```go
var rootCmd = &cobra.Command{
    Use:   "todo",
    Short: "A simple todo list manager",
}

var addCmd = &cobra.Command{
    Use:   "add [task]",
    Short: "Add a new task",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        return addTask(args[0])
    },
}

func init() {
    rootCmd.AddCommand(addCmd)
    addCmd.Flags().StringP("priority", "p", "normal", "Task priority")
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

## Exit Codes

```go
if err != nil {
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}
```

Never `log.Fatal` in a library — only in `main`. `log.Fatal` calls `os.Exit(1)` which skips `defer` statements.

## Reading from stdin

```go
scanner := bufio.NewScanner(os.Stdin)
for scanner.Scan() {
    fmt.Println(scanner.Text())
}
```

## Labs

### Lab 1: Minimal cobra App

**What you'll practise:** Creating a root cobra command with a `Run` function and a `--verbose` flag.

**Steps:**
1. `go get github.com/spf13/cobra`
2. Create `rootCmd` with `Use: "myapp"` and a `Run` that prints "Hello from myapp"
3. Bind `--verbose` with `rootCmd.PersistentFlags().Bool`
4. Run `go run . --verbose`

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
    "os"
)

var verbose bool

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "A demo cobra app",
    Run: func(cmd *cobra.Command, args []string) {
        if verbose {
            fmt.Println("[verbose] running root command")
        }
        fmt.Println("Hello from myapp")
    },
}

func init() {
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

**Expected output:**
```
$ go run . --verbose
[verbose] running root command
Hello from myapp
```

**Checkpoint:** Running without `--verbose` omits the verbose line; `go run . --help` shows the flag.

---

### Lab 2: Subcommands

**What you'll practise:** Adding `add` and `list` subcommands each with their own flags.

**Task:**
Extend the Lab 1 app with two subcommands: `add <item>` (with a `--priority` flag) and `list` (with a `--all` flag). Each subcommand prints what it would do.

**Steps:**
1. Define `addCmd` with `Args: cobra.ExactArgs(1)` and a `--priority` flag
2. Define `listCmd` with a `--all` flag
3. Register both with `rootCmd.AddCommand`
4. Test `go run . add "buy milk" --priority=high` and `go run . list --all`

```go
var addCmd = &cobra.Command{
    Use:   "add [item]",
    Short: "Add a new item",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        priority, _ := cmd.Flags().GetString("priority")
        fmt.Printf("Adding %q with priority %s\n", args[0], priority)
    },
}

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List items",
    Run: func(cmd *cobra.Command, args []string) {
        all, _ := cmd.Flags().GetBool("all")
        if all {
            fmt.Println("Listing all items (including done)")
        } else {
            fmt.Println("Listing pending items")
        }
    },
}

func init() {
    addCmd.Flags().String("priority", "normal", "Task priority (low/normal/high)")
    listCmd.Flags().Bool("all", false, "Show all items including completed")
    rootCmd.AddCommand(addCmd, listCmd)
}
```

**Expected output:**
```
$ go run . add "buy milk" --priority=high
Adding "buy milk" with priority high
$ go run . list --all
Listing all items (including done)
```

**Checkpoint:** `go run . --help` shows both subcommands; each subcommand's `--help` shows only its own flags.

---

### Lab 3: Persistent Flags

**What you'll practise:** Adding a `--output` flag to the root command that all subcommands inherit.

**Task:**
Add `--output=json|text` as a persistent flag on `rootCmd`. In each subcommand's `Run`, read the flag and switch output format accordingly.

**Steps:**
1. Add `--output` to `rootCmd.PersistentFlags()` with default `"text"`
2. In `listCmd.Run`, read `cmd.Root().PersistentFlags().GetString("output")` (or inherit via parent)
3. If `--output=json`, print a JSON object; otherwise print plain text
4. Test both formats

```go
// In init():
rootCmd.PersistentFlags().String("output", "text", "Output format: json or text")

// In listCmd.Run:
output, _ := cmd.Flags().GetString("output")
if output == "json" {
    fmt.Println(`{"items":["buy milk","write code"]}`)
} else {
    fmt.Println("- buy milk\n- write code")
}
```

**Expected output:**
```
$ go run . list --output=json
{"items":["buy milk","write code"]}
$ go run . list
- buy milk
- write code
```

**Checkpoint:** Both subcommands respond to `--output`; the flag does not need to be redefined in each subcommand.

---

### Lab 4: Args Validation

**What you'll practise:** Using cobra built-in validators and writing a custom `Args` function.

**Task:**
Create a `search` subcommand that requires exactly one query argument. Also create a `tag` subcommand that requires at least one argument. Finally, write a custom validator that rejects arguments containing spaces.

**Steps:**
1. Add `searchCmd` with `Args: cobra.ExactArgs(1)`
2. Add `tagCmd` with `Args: cobra.MinimumNArgs(1)`
3. Add `strictCmd` with a custom `Args` function that returns an error if any arg contains a space
4. Test each with incorrect input to see the error messages

```go
var strictCmd = &cobra.Command{
    Use:  "strict [words...]",
    Short: "Accepts only single-word arguments",
    Args: func(cmd *cobra.Command, args []string) error {
        for _, a := range args {
            if strings.Contains(a, " ") {
                return fmt.Errorf("argument %q must not contain spaces", a)
            }
        }
        return nil
    },
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Args OK:", args)
    },
}
```

**Expected output:**
```
$ go run . strict hello world
Args OK: [hello world]
$ go run . strict "hello world"
Error: argument "hello world" must not contain spaces
```

**Checkpoint:** Cobra prints the error and exits non-zero when validation fails; valid input runs the command.

---

### Lab 5: PersistentPreRunE

**What you'll practise:** Loading configuration before every command using `PersistentPreRunE`.

**Task:**
Add a `PersistentPreRunE` hook to `rootCmd` that reads a JSON config file (path from `--config` flag, defaulting to `config.json`). If the file is missing, log a warning but continue. Print the loaded config before the command runs.

**Steps:**
1. Add `--config` persistent flag to `rootCmd`
2. Set `rootCmd.PersistentPreRunE` to open and decode the config file
3. Store the config in a package-level variable accessible to subcommands
4. Test with a real `config.json` and with the flag pointing to a missing file

```go
var configPath string
var appConfig struct {
    Debug bool   `json:"debug"`
    Host  string `json:"host"`
}

func init() {
    rootCmd.PersistentFlags().StringVar(&configPath, "config", "config.json", "config file path")
    rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
        f, err := os.Open(configPath)
        if os.IsNotExist(err) {
            fmt.Fprintln(os.Stderr, "warning: config file not found, using defaults")
            return nil
        }
        if err != nil {
            return err
        }
        defer f.Close()
        return json.NewDecoder(f).Decode(&appConfig)
    }
}
```

**Expected output (with config.json containing `{"debug":true,"host":"localhost"}`):**
```
$ go run . list
Loaded config: {Debug:true Host:localhost}
Listing pending items
```

**Checkpoint:** Config loads before any subcommand runs; missing file produces a warning, not a fatal error.

---

### Lab 6: Shell Completion

**What you'll practise:** Generating shell completion scripts and adding custom completions with `ValidArgsFunction`.

**Task:**
Run `go run . completion bash` to see the generated completion script. Then add a `ValidArgsFunction` to `addCmd` that suggests `"--priority=low"`, `"--priority=normal"`, `"--priority=high"` for the `--priority` flag.

**Steps:**
1. Cobra adds a `completion` subcommand automatically — run it and inspect the output
2. Register `ValidArgsFunction` on `addCmd` (or use `RegisterFlagCompletionFunc` for the flag)
3. Optionally source the output in your shell and test tab completion

```go
func init() {
    addCmd.RegisterFlagCompletionFunc("priority", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        return []string{"low", "normal", "high"}, cobra.ShellCompDirectiveDefault
    })
}
```

**Expected output:**
```
$ go run . completion bash
# bash completion for myapp ...
```

**Checkpoint:** The completion script is non-empty and contains the command name; `RegisterFlagCompletionFunc` compiles without error.

---

### Final Lab: `todo` CLI Tool

**What you'll practise:** Building a full cobra CLI that persists state to JSON on disk.

**Task:**
Build a `todo` CLI with five subcommands. Tasks are persisted to `~/.todo.json` and each has an auto-incrementing ID, text, done flag, and creation time.

**Steps:**
1. Define `Task{ID int, Text string, Done bool, CreatedAt time.Time}` and a `store` that reads/writes `~/.todo.json`
2. Implement `add [text]` — append a task, save, print the new ID
3. Implement `list` — print pending tasks; `--all` flag includes done tasks
4. Implement `done <id>` — mark the task completed
5. Implement `delete <id>` — remove the task
6. Implement `export` — write all tasks as CSV to stdout

```go
// Skeleton for the store
func loadTasks() ([]Task, error) {
    path := filepath.Join(os.UserHomeDir(), ".todo.json") // handle error
    f, err := os.Open(path)
    if os.IsNotExist(err) {
        return nil, nil
    }
    var tasks []Task
    return tasks, json.NewDecoder(f).Decode(&tasks)
}

func saveTasks(tasks []Task) error {
    path, _ := os.UserHomeDir()
    f, err := os.Create(filepath.Join(path, ".todo.json"))
    if err != nil {
        return err
    }
    defer f.Close()
    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    return enc.Encode(tasks)
}
```

**Expected output:**
```
$ go run . add "buy milk"
added task #1
$ go run . list
1. buy milk
$ go run . done 1
task #1 marked done
$ go run . export
id,text,done,created_at
1,buy milk,true,2024-01-15T10:00:00Z
```

**Checkpoint:** `~/.todo.json` exists after the first `add`; `list` without `--all` hides done tasks; `export` produces valid CSV.

---

## Day Project: `todo` CLI Tool

Build a `todo` CLI with `cobra` that:
- Stores tasks in `~/.todo.json` (each task has ID, text, done bool, created time)
- `todo add "buy milk"` — adds a task, prints its ID
- `todo list` — lists all tasks (`--all` flag shows done tasks too)
- `todo done <id>` — marks a task complete
- `todo delete <id>` — removes a task
- `todo export` — prints tasks as CSV to stdout

**Extension ideas:** add `todo edit <id> "new text"`; add due dates with `--due 2024-12-31`; colour output with `github.com/fatih/color`.

## Official Documentation

- [`flag`](https://pkg.go.dev/flag) — standard library flag parsing (`flag.Int`, `flag.Bool`, `flag.String`, `flag.Parse`, `flag.Args`)
- [`os`](https://pkg.go.dev/os) — `Stderr`, `Stdin`, `Stdout`, `Exit`
- [`bufio`](https://pkg.go.dev/bufio) — `NewScanner` for reading from `os.Stdin`
- [`fmt`](https://pkg.go.dev/fmt) — `Fprintf` for writing to `os.Stderr`
- [`encoding/json`](https://pkg.go.dev/encoding/json) — for reading and writing `~/.todo.json`
- [cobra](https://pkg.go.dev/github.com/spf13/cobra) — subcommand CLI framework used in this day
- [Go Blog: Testable Examples in Go](https://go.dev/blog/examples)
- [Effective Go — Program initialisation and execution](https://go.dev/doc/effective_go#init)
