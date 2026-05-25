// Package hook implements the pre-push hook mode that displays an identity
// confirmation box and prompts the user before allowing a push.
package hook

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/haichen-zhang/gitsw/internal/config"
	"github.com/haichen-zhang/gitsw/internal/git"
)

// FormatIdentityBox renders a compact identity box for display before push.
// repoPath should already be shortened (e.g. ~/code/project).
// If !isLocal, header shows a warning and source shows "(global ⚠)".
// If profileInfo is empty and !isLocal: shows "No local config — using global fallback".
// If profileInfo is empty and isLocal: shows "unrecognized".
func FormatIdentityBox(repoPath, remoteURL, name, email string, isLocal bool, profileInfo string) string {
	var b strings.Builder

	// Header
	header := "gitsw: Identity Check"
	if !isLocal {
		header = "gitsw: ⚠ Identity Check"
	}

	// Box width
	const boxWidth = 60

	// Top border
	topLine := fmt.Sprintf("╭─ %s ", header)
	remaining := boxWidth - runeLen(topLine) - 1
	if remaining > 0 {
		topLine += strings.Repeat("─", remaining)
	}
	topLine += "╮"
	b.WriteString(topLine + "\n")

	// Repo line
	repoLine := fmt.Sprintf("│ Repo:    %s → %s", repoPath, remoteURL)
	b.WriteString(padRight(repoLine, boxWidth) + "\n")

	// User line
	source := "(local)"
	if !isLocal {
		source = "(global ⚠)"
	}
	userLine := fmt.Sprintf("│ User:    %s <%s> %s", name, email, source)
	b.WriteString(padRight(userLine, boxWidth) + "\n")

	// Profile line
	var profileDisplay string
	if profileInfo == "" {
		if !isLocal {
			profileDisplay = "No local config — using global fallback"
		} else {
			profileDisplay = "unrecognized"
		}
	} else {
		profileDisplay = profileInfo
	}
	profileLine := fmt.Sprintf("│ Profile: %s", profileDisplay)
	b.WriteString(padRight(profileLine, boxWidth) + "\n")

	// Bottom border
	bottomLine := "╰" + strings.Repeat("─", boxWidth-2) + "╯"
	b.WriteString(bottomLine + "\n")

	return b.String()
}

// padRight pads a line with spaces to reach the given width, then appends a closing border.
func padRight(line string, width int) string {
	lineLen := runeLen(line)
	if lineLen < width-1 {
		line += strings.Repeat(" ", width-1-lineLen)
	}
	return line
}

// runeLen counts the display width of a string (approximation: counts runes).
func runeLen(s string) int {
	return len([]rune(s))
}

// Prompt reads one line from reader. Returns true if empty, "y", or "yes"
// (case-insensitive). Returns false otherwise.
func Prompt(reader io.Reader) bool {
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		return false
	}
	line := strings.TrimSpace(scanner.Text())
	if line == "" {
		return true
	}
	lower := strings.ToLower(line)
	return lower == "y" || lower == "yes"
}

// Run executes the hook mode logic. Returns exit code 0 to allow push, 1 to abort.
func Run() int {
	// Skip in CI (not a TTY)
	if !isTerminal() {
		return 0
	}

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return 0
	}

	// Check if in a git repo
	if !git.IsGitRepo(cwd) {
		return 0
	}

	// Load config
	cfg, err := config.Load()
	if err != nil || len(cfg.Profiles) == 0 {
		return 0
	}

	// Get repo root
	repoRoot, err := git.GetRepoRoot(cwd)
	if err != nil {
		return 0
	}

	// Get identity
	identity, err := git.GetIdentity(repoRoot)
	if err != nil {
		return 0
	}

	// Get repo info
	repoInfo, err := git.GetRepoInfo(repoRoot)
	if err != nil {
		return 0
	}

	// Find matching profile
	var profileInfo string
	if profile, found := cfg.FindByEmail(identity.Email); found {
		profileInfo = fmt.Sprintf("%s (%s)", profile.Nickname, profile.Platform)
	}

	// Display identity box
	box := FormatIdentityBox(
		shortenHome(repoInfo.Path),
		repoInfo.RemoteURL,
		identity.Name,
		identity.Email,
		identity.IsLocal,
		profileInfo,
	)
	fmt.Fprint(os.Stderr, box)
	fmt.Fprint(os.Stderr, "Proceed with push? [Y/n] ")

	// Prompt
	if Prompt(os.Stdin) {
		return 0
	}
	fmt.Fprintln(os.Stderr, "Push aborted.")
	return 1
}

// isTerminal checks if stdin is a terminal (character device).
func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// shortenHome replaces the user's home directory prefix with ~.
func shortenHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}
