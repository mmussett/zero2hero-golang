package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
)

func main() {
	name := "Gopher"
	day := 1
	message := "The journey of a thousand miles begins with a single go run."

	// Accept an optional name from the first command-line argument
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	fmt.Println("=== Zero to Hero: Go ===")
	fmt.Printf("Day:       %d\n", day)
	fmt.Printf("Name:      %s\n", name)
	fmt.Printf("Message:   %s\n", message)
	fmt.Println()

	// runtime package: Go version, OS, and architecture
	fmt.Printf("Go:        %s\n", runtime.Version())
	fmt.Printf("OS/Arch:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CPUs:      %d\n", runtime.NumCPU())
	fmt.Println()

	// runtime/debug: module information embedded at build time
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("Module:    %s\n", info.Path)
		fmt.Printf("Go (mod):  %s\n", info.GoVersion)
	}
	fmt.Println()

	// os package: environment variables and command-line arguments
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = "(not set — using default ~/go)"
	}
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		goroot = runtime.GOROOT()
	}
	fmt.Printf("GOPATH:    %s\n", gopath)
	fmt.Printf("GOROOT:    %s\n", goroot)
	fmt.Println()

	// os.Args: the raw command-line arguments
	fmt.Printf("Args:      [%s]\n", strings.Join(os.Args, " "))
}
