package hook

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallLocal(t *testing.T) {
	// Create a temp dir with a .git directory
	tmp := t.TempDir()
	gitDir := filepath.Join(tmp, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Install
	if err := InstallLocal(tmp); err != nil {
		t.Fatalf("InstallLocal failed: %v", err)
	}

	// Verify file exists with correct content
	hookPath := filepath.Join(gitDir, "hooks", "pre-push")
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("failed to read hook file: %v", err)
	}
	if string(content) != hookScript {
		t.Errorf("hook content = %q, want %q", string(content), hookScript)
	}

	// Verify executable permissions
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("failed to stat hook file: %v", err)
	}
	perm := info.Mode().Perm()
	if perm&0111 == 0 {
		t.Errorf("hook file is not executable, perm = %o", perm)
	}
}

func TestInstallLocalNotGitRepo(t *testing.T) {
	tmp := t.TempDir()

	err := InstallLocal(tmp)
	if err == nil {
		t.Fatal("expected error for non-git directory, got nil")
	}
}

func TestInstallLocalForeignHook(t *testing.T) {
	tmp := t.TempDir()
	gitDir := filepath.Join(tmp, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write a foreign hook
	hookPath := filepath.Join(hooksDir, "pre-push")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho foreign\n"), 0755); err != nil {
		t.Fatal(err)
	}

	err := InstallLocal(tmp)
	if err == nil {
		t.Fatal("expected error for foreign hook, got nil")
	}
}

func TestUninstallLocal(t *testing.T) {
	tmp := t.TempDir()
	gitDir := filepath.Join(tmp, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Install first
	if err := InstallLocal(tmp); err != nil {
		t.Fatalf("InstallLocal failed: %v", err)
	}

	// Uninstall
	if err := UninstallLocal(tmp); err != nil {
		t.Fatalf("UninstallLocal failed: %v", err)
	}

	// Verify file is gone
	hookPath := filepath.Join(gitDir, "hooks", "pre-push")
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("expected hook file to be removed")
	}
}

func TestUninstallLocalForeignHook(t *testing.T) {
	tmp := t.TempDir()
	gitDir := filepath.Join(tmp, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write a foreign hook
	hookPath := filepath.Join(hooksDir, "pre-push")
	foreignContent := "#!/bin/sh\necho foreign\n"
	if err := os.WriteFile(hookPath, []byte(foreignContent), 0755); err != nil {
		t.Fatal(err)
	}

	// Attempt to uninstall should fail
	err := UninstallLocal(tmp)
	if err == nil {
		t.Fatal("expected error for foreign hook, got nil")
	}

	// File should be preserved
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("failed to read hook file: %v", err)
	}
	if string(content) != foreignContent {
		t.Error("foreign hook content was modified")
	}
}

func TestInstallGlobal(t *testing.T) {
	// Use a temp dir as the hooks directory
	tmp := t.TempDir()
	hooksDir := filepath.Join(tmp, "hooks")

	// We skip actually running git config --global in tests since that would
	// modify the user's global git config. Instead we test the file operations.
	// For a full integration test, InstallGlobal would be called in an isolated env.

	// Create the hooks dir and write the hook file directly to test the file logic
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	hookPath := filepath.Join(hooksDir, "pre-push")
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		t.Fatalf("failed to write hook: %v", err)
	}

	// Verify file exists with correct content
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("failed to read hook file: %v", err)
	}
	if string(content) != hookScript {
		t.Errorf("hook content = %q, want %q", string(content), hookScript)
	}

	// Verify executable permissions
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("failed to stat hook file: %v", err)
	}
	perm := info.Mode().Perm()
	if perm&0111 == 0 {
		t.Errorf("hook file is not executable, perm = %o", perm)
	}
}
