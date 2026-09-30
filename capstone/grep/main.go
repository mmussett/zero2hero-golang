package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
)

var isTerminal bool

func init() {
	// Check if stdout is a terminal
	fi, err := os.Stdout.Stat()
	if err == nil {
		isTerminal = (fi.Mode() & os.ModeCharDevice) != 0
	}
}

func colorize(color, text string) string {
	if !isTerminal {
		return text
	}
	return color + text + colorReset
}

type options struct {
	recursive   bool
	ignoreCase  bool
	lineNumbers bool
	onlyFiles   bool
	countOnly   bool
	invertMatch bool
	include     string
	pattern     string
	paths       []string
}

type result struct {
	filename string
	lineNum  int
	line     string
	matched  bool
}

func main() {
	fs := flag.NewFlagSet("grep", flag.ContinueOnError)
	recursive := fs.Bool("r", true, "recursive search (default true)")
	ignoreCase := fs.Bool("i", false, "case-insensitive matching")
	lineNumbers := fs.Bool("n", false, "show line numbers")
	onlyFiles := fs.Bool("l", false, "only print filenames with matches")
	countOnly := fs.Bool("c", false, "print count of matching lines per file")
	invertMatch := fs.Bool("v", false, "invert match (select non-matching lines)")
	include := fs.String("include", "", "include only files matching glob (e.g. *.go)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: grep <pattern> [path...] [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.PrintDefaults()
	}

	// Split args: flags can appear anywhere, but pattern must be first non-flag arg.
	// We'll parse everything with flag, treating first positional as pattern.
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "grep: %v\n", err)
		os.Exit(2)
	}

	args := fs.Args()
	if len(args) < 1 {
		fs.Usage()
		os.Exit(2)
	}

	opts := options{
		recursive:   *recursive,
		ignoreCase:  *ignoreCase,
		lineNumbers: *lineNumbers,
		onlyFiles:   *onlyFiles,
		countOnly:   *countOnly,
		invertMatch: *invertMatch,
		include:     *include,
		pattern:     args[0],
		paths:       args[1:],
	}

	if len(opts.paths) == 0 {
		opts.paths = []string{"."}
	}

	// Compile the regex
	patternStr := opts.pattern
	if opts.ignoreCase {
		patternStr = "(?i)" + patternStr
	}
	re, err := regexp.Compile(patternStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grep: invalid pattern %q: %v\n", opts.pattern, err)
		os.Exit(2)
	}

	found := false

	for _, path := range opts.paths {
		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: %s: %v\n", path, err)
			os.Exit(2)
		}

		if info.IsDir() {
			if !opts.recursive {
				fmt.Fprintf(os.Stderr, "grep: %s: Is a directory\n", path)
				continue
			}
			if searchDir(path, re, opts) {
				found = true
			}
		} else {
			if matchesGlob(opts.include, info.Name()) {
				if searchFile(path, re, opts, len(opts.paths) > 1) {
					found = true
				}
			}
		}
	}

	if !found {
		os.Exit(1)
	}
	os.Exit(0)
}

func matchesGlob(pattern, name string) bool {
	if pattern == "" {
		return true
	}
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		return false
	}
	return matched
}

func searchDir(root string, re *regexp.Regexp, opts options) bool {
	found := false
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: %s: %v\n", path, err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !matchesGlob(opts.include, d.Name()) {
			return nil
		}
		if searchFile(path, re, opts, true) {
			found = true
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "grep: walk error: %v\n", err)
	}
	return found
}

func searchFile(path string, re *regexp.Regexp, opts options, printFilename bool) bool {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grep: %s: %v\n", path, err)
		return false
	}
	defer f.Close()

	type match struct {
		lineNum int
		line    string
	}

	var matches []match
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		isMatch := re.MatchString(line)
		if opts.invertMatch {
			isMatch = !isMatch
		}
		if isMatch {
			matches = append(matches, match{lineNum, line})
		}
	}
	if err := scanner.Err(); err != nil {
		// Skip binary or unreadable files silently
		return false
	}

	if len(matches) == 0 {
		return false
	}

	displayPath := colorize(colorCyan, path)

	if opts.onlyFiles {
		fmt.Println(displayPath)
		return true
	}

	if opts.countOnly {
		if printFilename {
			fmt.Printf("%s:%d\n", displayPath, len(matches))
		} else {
			fmt.Printf("%d\n", len(matches))
		}
		return true
	}

	for _, m := range matches {
		var sb strings.Builder
		if printFilename {
			sb.WriteString(displayPath)
			sb.WriteString(":")
		}
		if opts.lineNumbers {
			sb.WriteString(colorize(colorGreen, fmt.Sprintf("%d", m.lineNum)))
			sb.WriteString(":")
		}
		// Highlight the match in the line
		if isTerminal && !opts.invertMatch {
			highlighted := re.ReplaceAllStringFunc(m.line, func(s string) string {
				return colorize(colorRed, s)
			})
			sb.WriteString(highlighted)
		} else {
			sb.WriteString(m.line)
		}
		fmt.Println(sb.String())
	}

	return true
}
