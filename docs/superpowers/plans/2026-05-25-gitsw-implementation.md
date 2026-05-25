# gitsw Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go TUI tool (`gitsw`) that manages Git user identities and confirms identity via a pre-push hook before every push.

**Architecture:** Single binary with dual mode — lightweight hook mode for pre-push confirmation (plain stdin/stdout) and full Bubble Tea TUI for interactive profile management. Profiles stored in `~/.gitswitch/profiles.yaml`.

**Tech Stack:** Go 1.22+, charmbracelet/bubbletea, charmbracelet/lipgloss, charmbracelet/bubbles, gopkg.in/yaml.v3

---

## File Structure

```
go.mod
go.sum
cmd/gitsw/main.go                  — entrypoint, command routing
internal/config/config.go           — Profile struct, Load/Save, CRUD operations
internal/config/config_test.go      — unit tests for config package
internal/git/git.go                 — read/write git config, detect repo info
internal/git/git_test.go            — unit tests for git package
internal/hook/hook.go               — hook mode logic (display, prompt, exit codes)
internal/hook/hook_test.go          — integration tests for hook mode
internal/hook/install.go            — install/uninstall hook scripts
internal/hook/install_test.go       — tests for hook installation
internal/tui/app.go                 — root Bubble Tea model, screen routing
internal/tui/views/dashboard.go     — main profile list view
internal/tui/views/form.go          — add/edit profile form view
internal/tui/views/confirm.go       — delete confirmation dialog
internal/tui/styles.go              — lipgloss style definitions
```

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/gitsw/main.go`

- [ ] **Step 1: Initialize Go module**

Run:
```bash
go mod init github.com/haichen-zhang/gitsw
```

- [ ] **Step 2: Create entrypoint with minimal command routing**

Create `cmd/gitsw/main.go`:

```go
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
```

- [ ] **Step 3: Verify it builds and runs**

Run:
```bash
go build -o gitsw ./cmd/gitsw && ./gitsw help
```

Expected: prints help text, exits 0.

- [ ] **Step 4: Commit**

```bash
git add go.mod cmd/
git commit -m "feat: scaffold project with command routing and help"
```

---

### Task 2: Config Package — Profile CRUD

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write failing tests for config package**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Profiles) != 0 {
		t.Fatalf("expected 0 profiles, got %d", len(cfg.Profiles))
	}
}

func TestAddAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	cfg, _ := LoadFrom(path)
	err := cfg.Add(Profile{
		Nickname: "work",
		Name:     "Zhang Haichen",
		Email:    "haichen@company.com",
		Platform: "gitlab",
	})
	if err != nil {
		t.Fatalf("unexpected error adding profile: %v", err)
	}
	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}
	if len(loaded.Profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(loaded.Profiles))
	}
	if loaded.Profiles[0].Email != "haichen@company.com" {
		t.Fatalf("expected email haichen@company.com, got %s", loaded.Profiles[0].Email)
	}
}

func TestAddDuplicateNickname(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	cfg, _ := LoadFrom(path)
	cfg.Add(Profile{Nickname: "work", Name: "A", Email: "a@a.com", Platform: "github"})
	err := cfg.Add(Profile{Nickname: "work", Name: "B", Email: "b@b.com", Platform: "gitlab"})
	if err == nil {
		t.Fatal("expected error for duplicate nickname, got nil")
	}
	_ = cfg.SaveTo(path)
}

func TestUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	cfg, _ := LoadFrom(path)
	cfg.Add(Profile{Nickname: "work", Name: "A", Email: "a@a.com", Platform: "github"})

	err := cfg.Update("work", Profile{Nickname: "work", Name: "B", Email: "b@b.com", Platform: "gitlab"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profiles[0].Name != "B" {
		t.Fatalf("expected name B, got %s", cfg.Profiles[0].Name)
	}
	_ = cfg.SaveTo(path)
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	cfg, _ := LoadFrom(path)
	cfg.Add(Profile{Nickname: "work", Name: "A", Email: "a@a.com", Platform: "github"})
	cfg.Add(Profile{Nickname: "personal", Name: "B", Email: "b@b.com", Platform: "github"})

	err := cfg.Delete("work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(cfg.Profiles))
	}
	if cfg.Profiles[0].Nickname != "personal" {
		t.Fatalf("expected personal, got %s", cfg.Profiles[0].Nickname)
	}
	_ = cfg.SaveTo(path)
}

func TestFindByEmail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	cfg, _ := LoadFrom(path)
	cfg.Add(Profile{Nickname: "work", Name: "A", Email: "a@a.com", Platform: "github"})

	p, found := cfg.FindByEmail("a@a.com")
	if !found {
		t.Fatal("expected to find profile")
	}
	if p.Nickname != "work" {
		t.Fatalf("expected work, got %s", p.Nickname)
	}

	_, found = cfg.FindByEmail("nope@nope.com")
	if found {
		t.Fatal("expected not found")
	}
}

func TestDefaultPath(t *testing.T) {
	path := DefaultPath()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".gitswitch", "profiles.yaml")
	if path != expected {
		t.Fatalf("expected %s, got %s", expected, path)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/config/ -v
```

Expected: compilation error — package doesn't exist yet.

- [ ] **Step 3: Implement config package**

Create `internal/config/config.go`:

```go
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Nickname string `yaml:"nickname"`
	Name     string `yaml:"name"`
	Email    string `yaml:"email"`
	Platform string `yaml:"platform"`
}

type Config struct {
	Profiles []Profile `yaml:"profiles"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gitswitch", "profiles.yaml")
}

func Load() (*Config, error) {
	return LoadFrom(DefaultPath())
}

func LoadFrom(path string) (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		backup := path + ".bak"
		os.Rename(path, backup)
		return &Config{}, fmt.Errorf("config corrupt (backed up to %s): %w", backup, err)
	}

	return cfg, nil
}

func (c *Config) Save() error {
	return c.SaveTo(DefaultPath())
}

func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) Add(p Profile) error {
	for _, existing := range c.Profiles {
		if existing.Nickname == p.Nickname {
			return fmt.Errorf("profile with nickname %q already exists", p.Nickname)
		}
	}
	c.Profiles = append(c.Profiles, p)
	return nil
}

func (c *Config) Update(nickname string, p Profile) error {
	for i, existing := range c.Profiles {
		if existing.Nickname == nickname {
			c.Profiles[i] = p
			return nil
		}
	}
	return fmt.Errorf("profile %q not found", nickname)
}

func (c *Config) Delete(nickname string) error {
	for i, existing := range c.Profiles {
		if existing.Nickname == nickname {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("profile %q not found", nickname)
}

func (c *Config) FindByEmail(email string) (Profile, bool) {
	for _, p := range c.Profiles {
		if p.Email == email {
			return p, true
		}
	}
	return Profile{}, false
}
```

- [ ] **Step 4: Add yaml dependency and run tests**

Run:
```bash
go get gopkg.in/yaml.v3
go test ./internal/config/ -v
```

Expected: all 6 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/ go.mod go.sum
git commit -m "feat: add config package with profile CRUD and YAML persistence"
```

---

### Task 3: Git Package — Read/Write Git Config

**Files:**
- Create: `internal/git/git.go`
- Create: `internal/git/git_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/git/git_test.go`:

```go
package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	return dir
}

func TestGetIdentityLocal(t *testing.T) {
	repo := setupTestRepo(t)

	cmd := exec.Command("git", "-C", repo, "config", "--local", "user.name", "Test User")
	cmd.Run()
	cmd = exec.Command("git", "-C", repo, "config", "--local", "user.email", "test@local.com")
	cmd.Run()

	identity, err := GetIdentity(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.Name != "Test User" {
		t.Fatalf("expected 'Test User', got %q", identity.Name)
	}
	if identity.Email != "test@local.com" {
		t.Fatalf("expected 'test@local.com', got %q", identity.Email)
	}
	if !identity.IsLocal {
		t.Fatal("expected IsLocal to be true")
	}
}

func TestGetIdentityGlobalFallback(t *testing.T) {
	repo := setupTestRepo(t)

	identity, err := GetIdentity(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.IsLocal {
		t.Fatal("expected IsLocal to be false for unconfigured repo")
	}
}

func TestSetIdentity(t *testing.T) {
	repo := setupTestRepo(t)

	err := SetIdentity(repo, "New User", "new@test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	identity, _ := GetIdentity(repo)
	if identity.Name != "New User" {
		t.Fatalf("expected 'New User', got %q", identity.Name)
	}
	if identity.Email != "new@test.com" {
		t.Fatalf("expected 'new@test.com', got %q", identity.Email)
	}
	if !identity.IsLocal {
		t.Fatal("expected IsLocal to be true after SetIdentity")
	}
}

func TestGetRepoInfo(t *testing.T) {
	repo := setupTestRepo(t)

	cmd := exec.Command("git", "-C", repo, "remote", "add", "origin", "git@github.com:user/repo.git")
	cmd.Run()

	info, err := GetRepoInfo(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.RemoteURL != "git@github.com:user/repo.git" {
		t.Fatalf("expected remote URL 'git@github.com:user/repo.git', got %q", info.RemoteURL)
	}
}

func TestIsGitRepo(t *testing.T) {
	repo := setupTestRepo(t)
	if !IsGitRepo(repo) {
		t.Fatal("expected true for git repo")
	}

	notRepo := t.TempDir()
	if IsGitRepo(notRepo) {
		t.Fatal("expected false for non-git dir")
	}
}

func TestGetRepoRoot(t *testing.T) {
	repo := setupTestRepo(t)
	subdir := filepath.Join(repo, "sub", "dir")
	os.MkdirAll(subdir, 0755)

	root, err := GetRepoRoot(subdir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root != repo {
		t.Fatalf("expected %q, got %q", repo, root)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/git/ -v
```

Expected: compilation error — package doesn't exist.

- [ ] **Step 3: Implement git package**

Create `internal/git/git.go`:

```go
package git

import (
	"os/exec"
	"path/filepath"
	"strings"
)

type Identity struct {
	Name    string
	Email   string
	IsLocal bool
}

type RepoInfo struct {
	Path      string
	RemoteURL string
}

func IsGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func GetRepoRoot(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func GetIdentity(repoDir string) (Identity, error) {
	name, nameLocal := getConfigValue(repoDir, "user.name")
	email, emailLocal := getConfigValue(repoDir, "user.email")

	return Identity{
		Name:    name,
		Email:   email,
		IsLocal: nameLocal && emailLocal,
	}, nil
}

func getConfigValue(repoDir, key string) (string, bool) {
	cmd := exec.Command("git", "-C", repoDir, "config", "--local", key)
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), true
	}

	cmd = exec.Command("git", "-C", repoDir, "config", key)
	out, err = cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), false
	}

	return "", false
}

func SetIdentity(repoDir, name, email string) error {
	cmd := exec.Command("git", "-C", repoDir, "config", "--local", "user.name", name)
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "-C", repoDir, "config", "--local", "user.email", email)
	return cmd.Run()
}

func GetRepoInfo(repoDir string) (RepoInfo, error) {
	absPath, err := filepath.Abs(repoDir)
	if err != nil {
		absPath = repoDir
	}

	remoteURL := ""
	cmd := exec.Command("git", "-C", repoDir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err == nil {
		remoteURL = strings.TrimSpace(string(out))
	}

	return RepoInfo{
		Path:      absPath,
		RemoteURL: remoteURL,
	}, nil
}
```

- [ ] **Step 4: Run tests**

Run:
```bash
go test ./internal/git/ -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/git/
git commit -m "feat: add git package for identity read/write and repo detection"
```

---

### Task 4: Hook Mode — Pre-push Confirmation

**Files:**
- Create: `internal/hook/hook.go`
- Create: `internal/hook/hook_test.go`
- Modify: `cmd/gitsw/main.go`

- [ ] **Step 1: Write failing tests for hook mode**

Create `internal/hook/hook_test.go`:

```go
package hook

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func setupRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	exec.Command("git", "init", dir).Run()
	exec.Command("git", "-C", dir, "config", "--local", "user.name", "Test User").Run()
	exec.Command("git", "-C", dir, "config", "--local", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", dir, "remote", "add", "origin", "git@github.com:user/repo.git").Run()
	return dir
}

func TestFormatIdentityBox(t *testing.T) {
	output := FormatIdentityBox("~/code/project", "git@github.com:user/repo.git", "Test User", "test@example.com", true, "personal (github)")
	if !strings.Contains(output, "Test User") {
		t.Fatal("expected output to contain user name")
	}
	if !strings.Contains(output, "test@example.com") {
		t.Fatal("expected output to contain email")
	}
	if !strings.Contains(output, "(local)") {
		t.Fatal("expected output to contain (local)")
	}
	if !strings.Contains(output, "personal (github)") {
		t.Fatal("expected output to contain profile info")
	}
}

func TestFormatIdentityBoxGlobalWarning(t *testing.T) {
	output := FormatIdentityBox("~/code/project", "git@github.com:user/repo.git", "Test User", "test@example.com", false, "")
	if !strings.Contains(output, "global") {
		t.Fatal("expected output to contain global warning")
	}
}

func TestPromptYes(t *testing.T) {
	input := bytes.NewBufferString("Y\n")
	result := Prompt(input)
	if !result {
		t.Fatal("expected true for Y input")
	}
}

func TestPromptEnterDefault(t *testing.T) {
	input := bytes.NewBufferString("\n")
	result := Prompt(input)
	if !result {
		t.Fatal("expected true for empty input (default Y)")
	}
}

func TestPromptNo(t *testing.T) {
	input := bytes.NewBufferString("n\n")
	result := Prompt(input)
	if result {
		t.Fatal("expected false for n input")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/hook/ -v
```

Expected: compilation error.

- [ ] **Step 3: Implement hook mode**

Create `internal/hook/hook.go`:

```go
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

func FormatIdentityBox(repoPath, remoteURL, name, email string, isLocal bool, profileInfo string) string {
	var b strings.Builder

	source := "(local)"
	if !isLocal {
		source = "(global ⚠)"
	}

	header := "gitsw: Identity Check"
	if !isLocal {
		header = "gitsw: ⚠ Identity Check"
	}

	b.WriteString(fmt.Sprintf("╭─ %s ──────────────────────────────────╮\n", header))
	b.WriteString(fmt.Sprintf("│ Repo:    %s → %s\n", repoPath, remoteURL))
	b.WriteString(fmt.Sprintf("│ User:    %s <%s> %s\n", name, email, source))

	if profileInfo != "" {
		b.WriteString(fmt.Sprintf("│ Profile: %s\n", profileInfo))
	} else if !isLocal {
		b.WriteString("│          No local config — using global fallback\n")
	} else {
		b.WriteString("│ Profile: unrecognized\n")
	}

	b.WriteString("╰──────────────────────────────────────────────────────────╯\n")

	return b.String()
}

func Prompt(reader io.Reader) bool {
	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	input := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return input == "" || input == "y" || input == "yes"
}

func Run() int {
	if !isTerminal() {
		return 0
	}

	cwd, err := os.Getwd()
	if err != nil {
		return 0
	}

	if !git.IsGitRepo(cwd) {
		return 0
	}

	cfg, _ := config.Load()
	if len(cfg.Profiles) == 0 {
		return 0
	}

	repoRoot, _ := git.GetRepoRoot(cwd)
	identity, _ := git.GetIdentity(repoRoot)
	repoInfo, _ := git.GetRepoInfo(repoRoot)

	profileInfo := ""
	if p, found := cfg.FindByEmail(identity.Email); found {
		profileInfo = fmt.Sprintf("%s (%s)", p.Nickname, p.Platform)
	}

	homePath := shortenHome(repoInfo.Path)
	box := FormatIdentityBox(homePath, repoInfo.RemoteURL, identity.Name, identity.Email, identity.IsLocal, profileInfo)
	fmt.Print(box)
	fmt.Print("Push as this identity? [Y/n] ")

	if Prompt(os.Stdin) {
		return 0
	}

	fmt.Println("\n✗ Push aborted. Run gitsw to switch identity, then push again.")
	return 1
}

func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

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
```

- [ ] **Step 4: Run tests**

Run:
```bash
go test ./internal/hook/ -v
```

Expected: all tests PASS.

- [ ] **Step 5: Wire hook command into main.go**

Update `cmd/gitsw/main.go` — add `"hook"` case to the switch:

```go
case "hook":
	os.Exit(hook.Run())
```

Add the import:
```go
"github.com/haichen-zhang/gitsw/internal/hook"
```

- [ ] **Step 6: Build and verify**

Run:
```bash
go build -o gitsw ./cmd/gitsw && ./gitsw hook
```

Expected: if run in a git repo with config, shows the identity box and prompts.

- [ ] **Step 7: Commit**

```bash
git add internal/hook/hook.go internal/hook/hook_test.go cmd/gitsw/main.go
git commit -m "feat: add hook mode with pre-push identity confirmation"
```

---

### Task 5: Hook Installation

**Files:**
- Create: `internal/hook/install.go`
- Create: `internal/hook/install_test.go`
- Modify: `cmd/gitsw/main.go`

- [ ] **Step 1: Write failing tests**

Create `internal/hook/install_test.go`:

```go
package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallLocal(t *testing.T) {
	repo := t.TempDir()
	exec.Command("git", "init", repo).Run()

	err := InstallLocal(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hookPath := filepath.Join(repo, ".git", "hooks", "pre-push")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("hook file not created: %v", err)
	}
	if !strings.Contains(string(data), "gitsw hook") {
		t.Fatal("hook file does not contain 'gitsw hook'")
	}

	info, _ := os.Stat(hookPath)
	if info.Mode()&0111 == 0 {
		t.Fatal("hook file is not executable")
	}
}

func TestInstallLocalNotGitRepo(t *testing.T) {
	dir := t.TempDir()
	err := InstallLocal(dir)
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func TestUninstallLocal(t *testing.T) {
	repo := t.TempDir()
	exec.Command("git", "init", repo).Run()
	InstallLocal(repo)

	err := UninstallLocal(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hookPath := filepath.Join(repo, ".git", "hooks", "pre-push")
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Fatal("hook file should be removed")
	}
}

func TestUninstallLocalForeignHook(t *testing.T) {
	repo := t.TempDir()
	exec.Command("git", "init", repo).Run()

	hookPath := filepath.Join(repo, ".git", "hooks", "pre-push")
	os.MkdirAll(filepath.Dir(hookPath), 0755)
	os.WriteFile(hookPath, []byte("#!/bin/sh\necho other hook"), 0755)

	err := UninstallLocal(repo)
	if err == nil {
		t.Fatal("expected error when hook is not ours")
	}

	data, _ := os.ReadFile(hookPath)
	if !strings.Contains(string(data), "other hook") {
		t.Fatal("foreign hook should not be modified")
	}
}

func TestInstallGlobal(t *testing.T) {
	dir := t.TempDir()
	hooksDir := filepath.Join(dir, "hooks")

	err := InstallGlobal(dir, hooksDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-push")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("hook file not created: %v", err)
	}
	if !strings.Contains(string(data), "gitsw hook") {
		t.Fatal("hook file does not contain 'gitsw hook'")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/hook/ -v -run "Install|Uninstall"
```

Expected: compilation error — `InstallLocal`, `UninstallLocal`, `InstallGlobal` not defined.

- [ ] **Step 3: Implement hook installation**

Create `internal/hook/install.go`:

```go
package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const hookScript = `#!/bin/sh
exec gitsw hook "$@"
`

const hookMarker = "gitsw hook"

func InstallLocal(repoDir string) error {
	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository: %s", repoDir)
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("creating hooks dir: %w", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-push")

	if data, err := os.ReadFile(hookPath); err == nil {
		if !strings.Contains(string(data), hookMarker) {
			return fmt.Errorf("pre-push hook already exists and is not managed by gitsw")
		}
	}

	return os.WriteFile(hookPath, []byte(hookScript), 0755)
}

func UninstallLocal(repoDir string) error {
	hookPath := filepath.Join(repoDir, ".git", "hooks", "pre-push")

	data, err := os.ReadFile(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if !strings.Contains(string(data), hookMarker) {
		return fmt.Errorf("pre-push hook is not managed by gitsw, refusing to remove")
	}

	return os.Remove(hookPath)
}

func InstallGlobal(configHome, hooksDir string) error {
	existing, err := exec.Command("git", "config", "--global", "core.hooksPath").Output()
	if err == nil {
		existingPath := strings.TrimSpace(string(existing))
		if existingPath != "" && existingPath != hooksDir {
			return fmt.Errorf("core.hooksPath already set to %q (not managed by gitsw)", existingPath)
		}
	}

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("creating hooks dir: %w", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-push")
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "config", "--global", "core.hooksPath", hooksDir)
	return cmd.Run()
}

func UninstallGlobal(hooksDir string) error {
	existing, err := exec.Command("git", "config", "--global", "core.hooksPath").Output()
	if err == nil {
		existingPath := strings.TrimSpace(string(existing))
		if existingPath != hooksDir {
			return fmt.Errorf("core.hooksPath points to %q, not managed by gitsw", existingPath)
		}
	}

	exec.Command("git", "config", "--global", "--unset", "core.hooksPath").Run()
	os.RemoveAll(hooksDir)
	return nil
}
```

- [ ] **Step 4: Run tests**

Run:
```bash
go test ./internal/hook/ -v -run "Install|Uninstall"
```

Expected: all install/uninstall tests PASS.

- [ ] **Step 5: Wire install/uninstall commands into main.go**

Add to `cmd/gitsw/main.go` switch:

```go
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
```

Add imports: `"path/filepath"`, `"github.com/haichen-zhang/gitsw/internal/hook"` (if not already present).

- [ ] **Step 6: Build and test manually**

Run:
```bash
go build -o gitsw ./cmd/gitsw && ./gitsw install && cat .git/hooks/pre-push
```

Expected: prints "Pre-push hook installed for this repo." and hook file contains `exec gitsw hook`.

- [ ] **Step 7: Commit**

```bash
git add internal/hook/install.go internal/hook/install_test.go cmd/gitsw/main.go
git commit -m "feat: add hook install/uninstall for local and global modes"
```

---

### Task 6: List Command

**Files:**
- Modify: `cmd/gitsw/main.go`

- [ ] **Step 1: Add list command to main.go**

Add to the switch:

```go
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
```

Add import: `"github.com/haichen-zhang/gitsw/internal/config"`, `"github.com/haichen-zhang/gitsw/internal/git"`.

- [ ] **Step 2: Build and test**

Run:
```bash
go build -o gitsw ./cmd/gitsw && ./gitsw list
```

Expected: "No profiles configured." or list of profiles with active marker.

- [ ] **Step 3: Commit**

```bash
git add cmd/gitsw/main.go
git commit -m "feat: add list command to show all profiles"
```

---

### Task 7: TUI — Styles and App Shell

**Files:**
- Create: `internal/tui/styles.go`
- Create: `internal/tui/app.go`

- [ ] **Step 1: Add Bubble Tea dependencies**

Run:
```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

- [ ] **Step 2: Create style definitions**

Create `internal/tui/styles.go`:

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorCyan    = lipgloss.Color("#64ffda")
	ColorPurple  = lipgloss.Color("#bb86fc")
	ColorYellow  = lipgloss.Color("#ffd93d")
	ColorRed     = lipgloss.Color("#ff6b6b")
	ColorDim     = lipgloss.Color("#888888")
	ColorWhite   = lipgloss.Color("#ffffff")

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true)

	ActiveStyle = lipgloss.NewStyle().
			Foreground(ColorCyan)

	DimStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed)

	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2a2a4a")).
			Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(1, 2)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorDim)
)
```

- [ ] **Step 3: Create app shell with screen routing**

Create `internal/tui/app.go`:

```go
package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/haichen-zhang/gitsw/internal/config"
	"github.com/haichen-zhang/gitsw/internal/tui/views"
)

type screen int

const (
	screenDashboard screen = iota
	screenForm
	screenConfirm
)

type App struct {
	current   screen
	dashboard views.Dashboard
	form      views.Form
	confirm   views.Confirm
	cfg       *config.Config
	cfgPath   string
	repoDir   string
	width     int
	height    int
}

func NewApp(cfg *config.Config, cfgPath, repoDir string) App {
	return App{
		current:   screenDashboard,
		dashboard: views.NewDashboard(cfg, repoDir),
		cfg:       cfg,
		cfgPath:   cfgPath,
		repoDir:   repoDir,
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	}

	switch a.current {
	case screenDashboard:
		return a.updateDashboard(msg)
	case screenForm:
		return a.updateForm(msg)
	case screenConfirm:
		return a.updateConfirm(msg)
	}

	return a, nil
}

func (a App) View() string {
	switch a.current {
	case screenForm:
		return a.form.View()
	case screenConfirm:
		return a.confirm.View()
	default:
		return a.dashboard.View()
	}
}

func (a App) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return a, tea.Quit
		case "a":
			a.form = views.NewForm(nil)
			a.current = screenForm
			return a, a.form.Init()
		case "e":
			if p := a.dashboard.SelectedProfile(); p != nil {
				a.form = views.NewForm(p)
				a.current = screenForm
				return a, a.form.Init()
			}
		case "d":
			if p := a.dashboard.SelectedProfile(); p != nil {
				a.confirm = views.NewConfirm(p.Nickname)
				a.current = screenConfirm
			}
		case "enter":
			if p := a.dashboard.SelectedProfile(); p != nil {
				a.dashboard.SwitchTo(p)
				a.cfg.Save()
			}
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.dashboard, cmd = a.dashboard.Update(msg)
	return a, cmd
}

func (a App) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case views.FormSubmitMsg:
		if msg.Editing != "" {
			a.cfg.Update(msg.Editing, msg.Profile)
		} else {
			a.cfg.Add(msg.Profile)
		}
		a.cfg.SaveTo(a.cfgPath)
		a.dashboard = views.NewDashboard(a.cfg, a.repoDir)
		a.current = screenDashboard
		return a, nil
	case views.FormCancelMsg:
		a.current = screenDashboard
		return a, nil
	}

	var cmd tea.Cmd
	a.form, cmd = a.form.Update(msg)
	return a, cmd
}

func (a App) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case views.ConfirmYesMsg:
		a.cfg.Delete(msg.Nickname)
		a.cfg.SaveTo(a.cfgPath)
		a.dashboard = views.NewDashboard(a.cfg, a.repoDir)
		a.current = screenDashboard
		return a, nil
	case views.ConfirmNoMsg:
		a.current = screenDashboard
		return a, nil
	}

	var cmd tea.Cmd
	a.confirm, cmd = a.confirm.Update(msg)
	return a, cmd
}

func Run(cfg *config.Config, cfgPath, repoDir string) error {
	app := NewApp(cfg, cfgPath, repoDir)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
```

- [ ] **Step 4: Verify it compiles** (views package not yet implemented, so just check syntax)

Run:
```bash
go vet ./internal/tui/... 2>&1 || true
```

Expected: errors about missing `views` package (expected at this stage).

- [ ] **Step 5: Commit**

```bash
git add internal/tui/styles.go internal/tui/app.go go.mod go.sum
git commit -m "feat: add TUI app shell with screen routing and styles"
```

---

### Task 8: TUI — Dashboard View

**Files:**
- Create: `internal/tui/views/dashboard.go`

- [ ] **Step 1: Implement dashboard view**

Create `internal/tui/views/dashboard.go`:

```go
package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haichen-zhang/gitsw/internal/config"
	"github.com/haichen-zhang/gitsw/internal/git"
)

var (
	dashTitleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#64ffda")).Bold(true)
	dashActiveStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#64ffda"))
	dashDimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	dashSelectedBg    = lipgloss.NewStyle().Background(lipgloss.Color("#2a2a4a")).Padding(0, 1)
	dashPurpleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#bb86fc"))
	dashYellowStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffd93d"))
	dashHelpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)

type Dashboard struct {
	cfg          *config.Config
	repoDir      string
	cursor       int
	currentEmail string
}

func NewDashboard(cfg *config.Config, repoDir string) Dashboard {
	currentEmail := ""
	if repoDir != "" && git.IsGitRepo(repoDir) {
		identity, _ := git.GetIdentity(repoDir)
		currentEmail = identity.Email
	}

	return Dashboard{
		cfg:          cfg,
		repoDir:      repoDir,
		cursor:       0,
		currentEmail: currentEmail,
	}
}

func (d Dashboard) Init() tea.Cmd {
	return nil
}

func (d Dashboard) Update(msg tea.Msg) (Dashboard, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if d.cursor > 0 {
				d.cursor--
			}
		case "down", "j":
			if d.cursor < len(d.cfg.Profiles)-1 {
				d.cursor++
			}
		}
	}
	return d, nil
}

func (d Dashboard) View() string {
	s := "\n"
	s += dashTitleStyle.Render("  🔧 gitsw — Git Identity Switcher") + "\n\n"

	if d.repoDir != "" {
		s += dashPurpleStyle.Render(fmt.Sprintf("  Repo: %s", shortenHome(d.repoDir))) + "\n"

		if d.currentEmail != "" {
			if p, found := d.cfg.FindByEmail(d.currentEmail); found {
				s += fmt.Sprintf("  Active: %s [%s/%s]\n",
					dashActiveStyle.Render(fmt.Sprintf("%s <%s>", p.Name, p.Email)),
					p.Nickname, p.Platform)
			} else {
				s += fmt.Sprintf("  Active: %s\n",
					dashActiveStyle.Render(d.currentEmail))
			}
		}
		s += "\n"
	}

	s += dashPurpleStyle.Render("  ─── Profiles ─────────────────────────────────") + "\n\n"

	if len(d.cfg.Profiles) == 0 {
		s += dashDimStyle.Render("  No profiles configured. Press [a] to add one.") + "\n"
	}

	for i, p := range d.cfg.Profiles {
		cursor := "  ▹ "
		nameStyle := dashDimStyle
		if i == d.cursor {
			cursor = "  ▸ "
			nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true)
		}

		active := ""
		if p.Email == d.currentEmail {
			active = dashActiveStyle.Render(" ● active")
		}

		platform := dashDimStyle.Render(p.Platform)
		if p.Platform == "gitlab" {
			platform = dashYellowStyle.Render(p.Platform)
		}

		line := fmt.Sprintf("%s%-10s %s <%s>  %s%s",
			cursor,
			nameStyle.Render(p.Nickname),
			p.Name, p.Email,
			platform, active)

		if i == d.cursor {
			line = dashSelectedBg.Render(line)
		}

		s += line + "\n"
	}

	s += "\n"
	s += dashHelpStyle.Render("  [enter] switch  [a] add  [e] edit  [d] delete  [i] install hook  [q] quit") + "\n"

	return s
}

func (d Dashboard) SelectedProfile() *config.Profile {
	if len(d.cfg.Profiles) == 0 {
		return nil
	}
	if d.cursor >= len(d.cfg.Profiles) {
		return nil
	}
	p := d.cfg.Profiles[d.cursor]
	return &p
}

func (d *Dashboard) SwitchTo(p *config.Profile) {
	if d.repoDir != "" {
		git.SetIdentity(d.repoDir, p.Name, p.Email)
		d.currentEmail = p.Email
	}
}

func shortenHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if len(path) > len(home) && path[:len(home)] == home {
		return "~" + path[len(home):]
	}
	return path
}
```

Add import `"os"` at the top.

- [ ] **Step 2: Verify it compiles**

Run:
```bash
go build ./internal/tui/views/ 2>&1 || true
```

Expected: errors about missing `Form`, `Confirm` types (not yet implemented — expected).

- [ ] **Step 3: Commit**

```bash
git add internal/tui/views/dashboard.go
git commit -m "feat: add TUI dashboard view with profile list and navigation"
```

---

### Task 9: TUI — Add/Edit Form View

**Files:**
- Create: `internal/tui/views/form.go`

- [ ] **Step 1: Implement form view**

Create `internal/tui/views/form.go`:

```go
package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/haichen-zhang/gitsw/internal/config"
)

var (
	formBoxStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#64ffda")).Padding(1, 2)
	formLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	formHelpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)

type FormSubmitMsg struct {
	Profile config.Profile
	Editing string
}

type FormCancelMsg struct{}

type Form struct {
	inputs     []textinput.Model
	focusIndex int
	platforms  []string
	platIndex  int
	editing    string
}

func NewForm(existing *config.Profile) Form {
	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "short identifier (e.g. work, personal)"
	inputs[0].CharLimit = 20

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "user.name for git commits"
	inputs[1].CharLimit = 100

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "user.email for git commits"
	inputs[2].CharLimit = 100

	platforms := []string{"github", "gitlab", "other"}
	platIndex := 0
	editing := ""

	if existing != nil {
		inputs[0].SetValue(existing.Nickname)
		inputs[1].SetValue(existing.Name)
		inputs[2].SetValue(existing.Email)
		editing = existing.Nickname
		for i, p := range platforms {
			if p == existing.Platform {
				platIndex = i
				break
			}
		}
	}

	inputs[0].Focus()

	return Form{
		inputs:     inputs,
		focusIndex: 0,
		platforms:  platforms,
		platIndex:  platIndex,
		editing:    editing,
	}
}

func (f Form) Init() tea.Cmd {
	return textinput.Blink
}

func (f Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return f, func() tea.Msg { return FormCancelMsg{} }
		case "enter":
			if f.focusIndex == 3 {
				return f, f.submit()
			}
			return f, f.submit()
		case "tab", "down":
			f.focusIndex++
			if f.focusIndex > 3 {
				f.focusIndex = 0
			}
			return f, f.updateFocus()
		case "shift+tab", "up":
			f.focusIndex--
			if f.focusIndex < 0 {
				f.focusIndex = 3
			}
			return f, f.updateFocus()
		case "left":
			if f.focusIndex == 3 {
				f.platIndex--
				if f.platIndex < 0 {
					f.platIndex = len(f.platforms) - 1
				}
				return f, nil
			}
		case "right":
			if f.focusIndex == 3 {
				f.platIndex++
				if f.platIndex >= len(f.platforms) {
					f.platIndex = 0
				}
				return f, nil
			}
		}
	}

	if f.focusIndex < 3 {
		var cmd tea.Cmd
		f.inputs[f.focusIndex], cmd = f.inputs[f.focusIndex].Update(msg)
		return f, cmd
	}

	return f, nil
}

func (f Form) View() string {
	title := "Add New Profile"
	if f.editing != "" {
		title = "Edit Profile"
	}

	labels := []string{"Nickname:", "Name:    ", "Email:   ", "Platform:"}

	s := "\n"
	s += formBoxStyle.Render(func() string {
		inner := lipgloss.NewStyle().Foreground(lipgloss.Color("#64ffda")).Bold(true).Render(title) + "\n\n"

		for i, input := range f.inputs {
			inner += fmt.Sprintf(" %s %s\n", formLabelStyle.Render(labels[i]), input.View())
		}

		platLine := " " + formLabelStyle.Render(labels[3]) + " "
		for i, p := range f.platforms {
			if i == f.platIndex {
				platLine += lipgloss.NewStyle().Background(lipgloss.Color("#2a2a4a")).Foreground(lipgloss.Color("#64ffda")).Padding(0, 1).Render("▸ " + p)
			} else {
				platLine += lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render(p)
			}
			platLine += "  "
		}
		inner += platLine + "\n"

		return inner
	}())

	s += "\n"
	s += formHelpStyle.Render(" [tab] next field  [shift+tab] prev  [enter] save  [esc] cancel") + "\n"

	return s
}

func (f Form) submit() tea.Cmd {
	nickname := f.inputs[0].Value()
	name := f.inputs[1].Value()
	email := f.inputs[2].Value()

	if nickname == "" || name == "" || email == "" {
		return nil
	}

	return func() tea.Msg {
		return FormSubmitMsg{
			Profile: config.Profile{
				Nickname: nickname,
				Name:     name,
				Email:    email,
				Platform: f.platforms[f.platIndex],
			},
			Editing: f.editing,
		}
	}
}

func (f *Form) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(f.inputs))
	for i := range f.inputs {
		if i == f.focusIndex {
			cmds[i] = f.inputs[i].Focus()
		} else {
			f.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}
```

- [ ] **Step 2: Verify it compiles (will still fail due to missing confirm)**

Run:
```bash
go build ./internal/tui/views/ 2>&1 || true
```

Expected: error about missing `Confirm` type.

- [ ] **Step 3: Commit**

```bash
git add internal/tui/views/form.go
git commit -m "feat: add TUI form view for profile add/edit"
```

---

### Task 10: TUI — Confirm Dialog and Final Integration

**Files:**
- Create: `internal/tui/views/confirm.go`
- Modify: `cmd/gitsw/main.go`

- [ ] **Step 1: Implement confirm dialog**

Create `internal/tui/views/confirm.go`:

```go
package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfirmYesMsg struct {
	Nickname string
}

type ConfirmNoMsg struct{}

type Confirm struct {
	nickname string
	cursor   int
}

func NewConfirm(nickname string) Confirm {
	return Confirm{nickname: nickname, cursor: 1}
}

func (c Confirm) Init() tea.Cmd {
	return nil
}

func (c Confirm) Update(msg tea.Msg) (Confirm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			c.cursor = 0
		case "right", "l":
			c.cursor = 1
		case "enter":
			if c.cursor == 0 {
				return c, func() tea.Msg { return ConfirmYesMsg{Nickname: c.nickname} }
			}
			return c, func() tea.Msg { return ConfirmNoMsg{} }
		case "esc", "n":
			return c, func() tea.Msg { return ConfirmNoMsg{} }
		case "y":
			return c, func() tea.Msg { return ConfirmYesMsg{Nickname: c.nickname} }
		}
	}
	return c, nil
}

func (c Confirm) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#ff6b6b")).
		Padding(1, 2)

	yesStyle := lipgloss.NewStyle().Padding(0, 2)
	noStyle := lipgloss.NewStyle().Padding(0, 2)

	if c.cursor == 0 {
		yesStyle = yesStyle.Background(lipgloss.Color("#ff6b6b")).Foreground(lipgloss.Color("#ffffff"))
	} else {
		noStyle = noStyle.Background(lipgloss.Color("#2a2a4a")).Foreground(lipgloss.Color("#ffffff"))
	}

	inner := fmt.Sprintf(
		"%s\n\n  %s  %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b6b")).Bold(true).Render(
			fmt.Sprintf("Delete profile %q?", c.nickname)),
		yesStyle.Render("Yes, delete"),
		noStyle.Render("No, cancel"),
	)

	s := "\n" + boxStyle.Render(inner) + "\n"
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("\n [←/→] select  [enter] confirm  [esc] cancel") + "\n"
	return s
}
```

- [ ] **Step 2: Verify all TUI packages compile**

Run:
```bash
go build ./internal/tui/...
```

Expected: successful compilation with no errors.

- [ ] **Step 3: Wire TUI into main.go**

Update `cmd/gitsw/main.go` — replace the default case at the beginning (when no args):

```go
if len(os.Args) < 2 {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	cfgPath := config.DefaultPath()
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
```

Add import: `"github.com/haichen-zhang/gitsw/internal/tui"`

- [ ] **Step 4: Full build and manual test**

Run:
```bash
go build -o gitsw ./cmd/gitsw && ./gitsw
```

Expected: launches TUI with empty profiles list, shows keybindings. Press `a` to open add form, `q` to quit.

- [ ] **Step 5: Commit**

```bash
git add internal/tui/views/confirm.go cmd/gitsw/main.go
git commit -m "feat: add confirm dialog and wire TUI into main entrypoint"
```

---

### Task 11: TUI — Hook Install View (via dashboard keybinding)

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/views/dashboard.go`

- [ ] **Step 1: Add `i` keybinding handler in app.go**

In `updateDashboard`, add a case for `"i"`:

```go
case "i":
	cwd, _ := os.Getwd()
	if git.IsGitRepo(cwd) {
		err := hook.InstallLocal(cwd)
		if err != nil {
			a.dashboard.SetStatus(fmt.Sprintf("⚠ %v", err))
		} else {
			a.dashboard.SetStatus("✓ Pre-push hook installed for this repo")
		}
	} else {
		a.dashboard.SetStatus("⚠ Not in a git repository")
	}
	return a, nil
```

Add imports: `"os"`, `"fmt"`, `"github.com/haichen-zhang/gitsw/internal/hook"`.

- [ ] **Step 2: Add status message to dashboard**

Add to `Dashboard` struct:

```go
type Dashboard struct {
	cfg          *config.Config
	repoDir      string
	cursor       int
	currentEmail string
	status       string
}
```

Add method:

```go
func (d *Dashboard) SetStatus(msg string) {
	d.status = msg
}
```

In `View()`, add before the help line:

```go
if d.status != "" {
	s += "\n  " + d.status + "\n"
}
```

- [ ] **Step 3: Build and test**

Run:
```bash
go build -o gitsw ./cmd/gitsw && ./gitsw
```

Expected: pressing `i` in the TUI shows "✓ Pre-push hook installed" status message.

- [ ] **Step 4: Commit**

```bash
git add internal/tui/app.go internal/tui/views/dashboard.go
git commit -m "feat: add hook install action from TUI dashboard"
```

---

### Task 12: Final Integration and Cleanup

**Files:**
- Modify: `cmd/gitsw/main.go`
- Create: `.gitignore`

- [ ] **Step 1: Add .gitignore**

Create `.gitignore`:

```
gitsw
*.exe
.superpowers/
```

- [ ] **Step 2: Ensure all commands are wired and build is clean**

Run:
```bash
go build -o gitsw ./cmd/gitsw && go vet ./... && go test ./...
```

Expected: builds successfully, no vet warnings, all tests pass.

- [ ] **Step 3: Test full workflow manually**

Run:
```bash
./gitsw help
./gitsw version
./gitsw list
./gitsw install
./gitsw hook
./gitsw uninstall
```

Expected: each command works as designed.

- [ ] **Step 4: Commit**

```bash
git add .gitignore cmd/ internal/ go.mod go.sum
git commit -m "feat: complete gitsw with all commands and TUI"
```

---

## Summary

| Task | Component | Key Deliverable |
|------|-----------|-----------------|
| 1 | Scaffolding | go.mod, main.go with help/version |
| 2 | Config | Profile CRUD + YAML persistence |
| 3 | Git | Read/write identity, detect repo |
| 4 | Hook mode | Pre-push Y/n confirmation |
| 5 | Hook install | Local + global hook management |
| 6 | List command | CLI profile listing |
| 7 | TUI shell | Bubble Tea app with screen routing |
| 8 | Dashboard | Profile list with navigation |
| 9 | Form | Add/edit profile UI |
| 10 | Confirm + wire | Delete dialog, TUI in main.go |
| 11 | Hook in TUI | Install hook from dashboard |
| 12 | Cleanup | .gitignore, final integration test |
