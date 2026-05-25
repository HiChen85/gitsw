package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// helper to create an isolated git repo in a temp directory
func setupTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmd := exec.Command("git", "init", dir)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}
	return dir
}

func TestIsGitRepo(t *testing.T) {
	t.Run("true for git repo", func(t *testing.T) {
		dir := setupTempRepo(t)
		if !IsGitRepo(dir) {
			t.Error("expected IsGitRepo to return true for a git repo")
		}
	})

	t.Run("false for non-git dir", func(t *testing.T) {
		dir := t.TempDir() // plain directory, no git init
		if IsGitRepo(dir) {
			t.Error("expected IsGitRepo to return false for a non-git directory")
		}
	})
}

func TestGetRepoRoot(t *testing.T) {
	dir := setupTempRepo(t)

	// Create a subdirectory
	subDir := filepath.Join(dir, "sub", "deep")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}

	root, err := GetRepoRoot(subDir)
	if err != nil {
		t.Fatalf("GetRepoRoot failed: %v", err)
	}

	// Resolve symlinks for comparison (macOS /tmp -> /private/tmp)
	expectedRoot, _ := filepath.EvalSymlinks(dir)
	actualRoot, _ := filepath.EvalSymlinks(root)

	if actualRoot != expectedRoot {
		t.Errorf("expected root %q, got %q", expectedRoot, actualRoot)
	}
}

func TestGetIdentityLocal(t *testing.T) {
	dir := setupTempRepo(t)

	// Set local config
	runGit(t, dir, "config", "--local", "user.name", "Local User")
	runGit(t, dir, "config", "--local", "user.email", "local@example.com")

	id, err := GetIdentity(dir)
	if err != nil {
		t.Fatalf("GetIdentity failed: %v", err)
	}

	if id.Name != "Local User" {
		t.Errorf("expected name %q, got %q", "Local User", id.Name)
	}
	if id.Email != "local@example.com" {
		t.Errorf("expected email %q, got %q", "local@example.com", id.Email)
	}
	if !id.IsLocal {
		t.Error("expected IsLocal to be true when both name and email are set locally")
	}
}

func TestGetIdentityGlobalFallback(t *testing.T) {
	dir := setupTempRepo(t)

	// Set global config using a custom global config file
	globalCfg := filepath.Join(t.TempDir(), "gitconfig")
	runGitWithEnv(t, dir, []string{"GIT_CONFIG_GLOBAL=" + globalCfg}, "config", "--global", "user.name", "Global User")
	runGitWithEnv(t, dir, []string{"GIT_CONFIG_GLOBAL=" + globalCfg}, "config", "--global", "user.email", "global@example.com")

	// Do NOT set local config — should fall back to global
	id, err := GetIdentityWithEnv(dir, []string{"GIT_CONFIG_GLOBAL=" + globalCfg})
	if err != nil {
		t.Fatalf("GetIdentity failed: %v", err)
	}

	if id.Name != "Global User" {
		t.Errorf("expected name %q, got %q", "Global User", id.Name)
	}
	if id.Email != "global@example.com" {
		t.Errorf("expected email %q, got %q", "global@example.com", id.Email)
	}
	if id.IsLocal {
		t.Error("expected IsLocal to be false when config is from global")
	}
}

func TestSetIdentity(t *testing.T) {
	dir := setupTempRepo(t)

	err := SetIdentity(dir, "Set User", "set@example.com")
	if err != nil {
		t.Fatalf("SetIdentity failed: %v", err)
	}

	// Read back and verify
	id, err := GetIdentity(dir)
	if err != nil {
		t.Fatalf("GetIdentity after SetIdentity failed: %v", err)
	}

	if id.Name != "Set User" {
		t.Errorf("expected name %q, got %q", "Set User", id.Name)
	}
	if id.Email != "set@example.com" {
		t.Errorf("expected email %q, got %q", "set@example.com", id.Email)
	}
	if !id.IsLocal {
		t.Error("expected IsLocal to be true after SetIdentity")
	}
}

func TestGetRepoInfo(t *testing.T) {
	dir := setupTempRepo(t)

	// Set a remote
	runGit(t, dir, "remote", "add", "origin", "https://github.com/test/repo.git")

	info, err := GetRepoInfo(dir)
	if err != nil {
		t.Fatalf("GetRepoInfo failed: %v", err)
	}

	// Resolve symlinks for comparison
	expectedPath, _ := filepath.EvalSymlinks(dir)
	actualPath, _ := filepath.EvalSymlinks(info.Path)

	if actualPath != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, actualPath)
	}
	if info.RemoteURL != "https://github.com/test/repo.git" {
		t.Errorf("expected remote URL %q, got %q", "https://github.com/test/repo.git", info.RemoteURL)
	}
}

// helper to run a git command in a directory
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

// helper to run a git command in a directory with extra env vars
func runGitWithEnv(t *testing.T, dir string, env []string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}
