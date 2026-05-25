package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("gitsw TUI mode (not yet implemented)")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "help", "--help", "-h":
		printHelp()
	case "version", "--version", "-v":
		fmt.Printf("gitsw %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`gitsw - Git user identity switcher

Usage:
  gitsw              Launch interactive TUI
  gitsw hook         Pre-push hook mode (used by git hooks)
  gitsw install      Install pre-push hook to current repo
  gitsw install -g   Install pre-push hook globally
  gitsw uninstall    Remove pre-push hook from current repo
  gitsw uninstall -g Remove global pre-push hook
  gitsw list         List all configured profiles
  gitsw help         Show this help message

Flags:
  -h, --help         Show help
  -v, --version      Show version
`)
}
