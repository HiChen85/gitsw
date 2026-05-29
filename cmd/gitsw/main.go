package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HiChen85/gitsw/internal/config"
	"github.com/HiChen85/gitsw/internal/git"
	"github.com/HiChen85/gitsw/internal/hook"
	"github.com/HiChen85/gitsw/internal/tui"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		cfgPath := config.DefaultPath()

		if len(cfg.Profiles) == 0 {
			importGlobalProfile(cfg, cfgPath)
		}

		cwd, _ := os.Getwd()
		repoDir := ""
		if git.IsGitRepo(cwd) {
			repoDir, _ = git.GetRepoRoot(cwd)
		}
		if err := tui.Run(cfg, cfgPath, repoDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	switch os.Args[1] {
	case "hook":
		os.Exit(hook.Run())
	case "install":
		global := len(os.Args) > 2 && (os.Args[2] == "-g" || os.Args[2] == "--global")
		if global {
			home, _ := os.UserHomeDir()
			hooksDir := filepath.Join(home, ".gitswitch", "hooks")
			if err := hook.InstallGlobal(home, hooksDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Global pre-push hook installed.")
		} else {
			cwd, _ := os.Getwd()
			if err := hook.InstallLocal(cwd); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Pre-push hook installed for this repo.")
		}
	case "uninstall":
		global := len(os.Args) > 2 && (os.Args[2] == "-g" || os.Args[2] == "--global")
		if global {
			home, _ := os.UserHomeDir()
			hooksDir := filepath.Join(home, ".gitswitch", "hooks")
			if err := hook.UninstallGlobal(hooksDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Global pre-push hook removed.")
		} else {
			cwd, _ := os.Getwd()
			if err := hook.UninstallLocal(cwd); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Pre-push hook removed from this repo.")
		}
	case "list":
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		if len(cfg.Profiles) == 0 {
			fmt.Println("No profiles configured. Run gitsw to add one.")
			os.Exit(0)
		}

		cwd, _ := os.Getwd()
		var currentEmail string
		if git.IsGitRepo(cwd) {
			identity, _ := git.GetIdentity(cwd)
			currentEmail = identity.Email
		}

		for _, p := range cfg.Profiles {
			marker := "  "
			if p.Email == currentEmail {
				marker = "● "
			}
			fmt.Printf("%s%-12s %s <%s>  (%s)\n", marker, p.Nickname, p.Name, p.Email, p.Platform)
		}
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

func importGlobalProfile(cfg *config.Config, cfgPath string) {
	identity, _ := git.GetGlobalIdentity()
	if identity.Name == "" || identity.Email == "" {
		return
	}

	p := config.Profile{
		Nickname: "default",
		Name:     identity.Name,
		Email:    identity.Email,
		Platform: "github",
	}

	if err := cfg.Add(p); err != nil {
		return
	}
	_ = cfg.SaveTo(cfgPath)
	fmt.Printf("Imported global git config as profile \"default\": %s <%s>\n", identity.Name, identity.Email)
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
