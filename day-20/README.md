# Day 20: CLI Applications

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

## Day Project: `todo` CLI Tool

Build a `todo` CLI with `cobra` that:
- Stores tasks in `~/.todo.json` (each task has ID, text, done bool, created time)
- `todo add "buy milk"` — adds a task, prints its ID
- `todo list` — lists all tasks (`--all` flag shows done tasks too)
- `todo done <id>` — marks a task complete
- `todo delete <id>` — removes a task
- `todo export` — prints tasks as CSV to stdout

**Extension ideas:** add `todo edit <id> "new text"`; add due dates with `--due 2024-12-31`; colour output with `github.com/fatih/color`.
