package hook

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const hookScript = "#!/bin/sh\nexec gitsw hook \"$@\"\n"
const hookMarker = "gitsw hook"

// InstallLocal installs the pre-push hook into the given repo's .git/hooks directory.
func InstallLocal(repoDir string) error {
	gitDir := filepath.Join(repoDir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("not a git repository: %s", repoDir)
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	hookPath := filepath.Join(hooksDir, "pre-push")

	// Check if hook already exists
	existing, err := os.ReadFile(hookPath)
	if err == nil {
		// File exists — check if it's ours
		if !strings.Contains(string(existing), hookMarker) {
			return fmt.Errorf("pre-push hook already exists and was not installed by gitsw")
		}
	}

	// Write the hook script
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return fmt.Errorf("failed to write pre-push hook: %w", err)
	}

	return nil
}

// UninstallLocal removes the pre-push hook from the given repo if it was installed by gitsw.
func UninstallLocal(repoDir string) error {
	hookPath := filepath.Join(repoDir, ".git", "hooks", "pre-push")

	existing, err := os.ReadFile(hookPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // already gone
		}
		return fmt.Errorf("failed to read pre-push hook: %w", err)
	}

	if !strings.Contains(string(existing), hookMarker) {
		return fmt.Errorf("pre-push hook was not installed by gitsw; refusing to remove")
	}

	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("failed to remove pre-push hook: %w", err)
	}

	return nil
}

// InstallGlobal installs the pre-push hook globally via core.hooksPath.
func InstallGlobal(configHome, hooksDir string) error {
	// Check existing core.hooksPath
	existing, err := exec.Command("git", "config", "--global", "core.hooksPath").Output()
	if err == nil {
		existingPath := strings.TrimSpace(string(existing))
		if existingPath != "" && existingPath != hooksDir {
			return fmt.Errorf("core.hooksPath is already set to %q (expected %q)", existingPath, hooksDir)
		}
	}

	// Create hooks directory
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create global hooks directory: %w", err)
	}

	// Write pre-push hook
	hookPath := filepath.Join(hooksDir, "pre-push")
	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return fmt.Errorf("failed to write global pre-push hook: %w", err)
	}

	// Set core.hooksPath
	if err := exec.Command("git", "config", "--global", "core.hooksPath", hooksDir).Run(); err != nil {
		return fmt.Errorf("failed to set core.hooksPath: %w", err)
	}

	return nil
}

// UninstallGlobal removes the global pre-push hook and unsets core.hooksPath.
func UninstallGlobal(hooksDir string) error {
	// Check existing core.hooksPath
	existing, err := exec.Command("git", "config", "--global", "core.hooksPath").Output()
	if err == nil {
		existingPath := strings.TrimSpace(string(existing))
		if existingPath != "" && existingPath != hooksDir {
			return fmt.Errorf("core.hooksPath points to %q, not %q; refusing to uninstall", existingPath, hooksDir)
		}
	}

	// Unset core.hooksPath
	_ = exec.Command("git", "config", "--global", "--unset", "core.hooksPath").Run()

	// Remove hooks directory
	if err := os.RemoveAll(hooksDir); err != nil {
		return fmt.Errorf("failed to remove global hooks directory: %w", err)
	}

	return nil
}
