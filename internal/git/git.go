// Package git wraps git CLI commands for reading/writing identity config
// and detecting repository information.
package git

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// Identity represents a git user identity.
type Identity struct {
	Name    string
	Email   string
	IsLocal bool
}

// RepoInfo holds information about a git repository.
type RepoInfo struct {
	Path      string
	RemoteURL string
}

// IsGitRepo checks if the given directory is inside a git repository.
func IsGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// GetRepoRoot returns the root directory of the git repository
// containing the given directory.
func GetRepoRoot(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// getConfigValue attempts to read a git config key. It first tries --local,
// and if that fails, tries without scope (effective value which may be global).
// Returns the value and whether it was found locally.
func getConfigValue(dir string, key string, env []string) (string, bool) {
	// Try local first
	cmd := exec.Command("git", "-C", dir, "config", "--local", key)
	if env != nil {
		cmd.Env = env
	}
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), true
	}

	// Fall back to effective value (global/system)
	cmd = exec.Command("git", "-C", dir, "config", key)
	if env != nil {
		cmd.Env = env
	}
	out, err = cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), false
	}

	return "", false
}

// GetIdentity reads the current user.name and user.email from git config.
// IsLocal is true only if BOTH name and email are set in local config.
func GetIdentity(repoDir string) (Identity, error) {
	return GetIdentityWithEnv(repoDir, nil)
}

// GetIdentityWithEnv reads identity with custom environment variables.
// This is primarily used for testing with custom GIT_CONFIG_GLOBAL.
func GetIdentityWithEnv(repoDir string, env []string) (Identity, error) {
	name, nameLocal := getConfigValue(repoDir, "user.name", env)
	email, emailLocal := getConfigValue(repoDir, "user.email", env)

	id := Identity{
		Name:    name,
		Email:   email,
		IsLocal: nameLocal && emailLocal,
	}
	return id, nil
}

// SetIdentity sets user.name and user.email in the local git config.
func SetIdentity(repoDir, name, email string) error {
	cmd := exec.Command("git", "-C", repoDir, "config", "--local", "user.name", name)
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "-C", repoDir, "config", "--local", "user.email", email)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// GetRepoInfo returns the absolute path and origin remote URL of the repository.
func GetRepoInfo(repoDir string) (RepoInfo, error) {
	// Get absolute path
	absPath, err := filepath.Abs(repoDir)
	if err != nil {
		return RepoInfo{}, err
	}

	// Resolve to real path (follow symlinks)
	absPath, err = filepath.EvalSymlinks(absPath)
	if err != nil {
		return RepoInfo{}, err
	}

	// Get remote URL
	cmd := exec.Command("git", "-C", repoDir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	remoteURL := ""
	if err == nil {
		remoteURL = strings.TrimSpace(string(out))
	}

	return RepoInfo{
		Path:      absPath,
		RemoteURL: remoteURL,
	}, nil
}
