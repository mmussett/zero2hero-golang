package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config map[string]map[string]string

const sampleINI = `# Application configuration

[database]
host = localhost
port = 5432
name = mydb
user = admin

[server]
port = 8080
debug = true
timeout = 30s

[logging]
level = info
file = app.log
`

func parseINI(text string) (Config, error) {
	cfg := make(Config)
	section := ""

	scanner := bufio.NewScanner(strings.NewReader(text))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			if cfg[section] == nil {
				cfg[section] = make(map[string]string)
			}
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: invalid format %q", lineNum, line)
		}
		if section == "" {
			return nil, fmt.Errorf("line %d: key=value outside section", lineNum)
		}
		cfg[section][strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return cfg, scanner.Err()
}

func writeINI(cfg Config, w *os.File) {
	for section, keys := range cfg {
		fmt.Fprintf(w, "[%s]\n", section)
		for k, v := range keys {
			fmt.Fprintf(w, "%s = %s\n", k, v)
		}
		fmt.Fprintln(w)
	}
}

func main() {
	fmt.Print("=== INI Config Reader ===\n\n")

	cfg, err := parseINI(sampleINI)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	for _, sec := range []string{"database", "server", "logging"} {
		fmt.Printf("[%s]\n", sec)
		for k, v := range cfg[sec] {
			fmt.Printf("  %-12s = %s\n", k, v)
		}
	}

	fmt.Print("\n=== Round-trip: write back to stdout ===\n\n")
	writeINI(cfg, os.Stdout)
}
