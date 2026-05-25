package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	expected := filepath.Join(home, ".gitswitch", "profiles.yaml")
	got := DefaultPath()
	if got != expected {
		t.Errorf("DefaultPath() = %q, want %q", got, expected)
	}
}

func TestLoadEmpty(t *testing.T) {
	// Loading a non-existent file should return an empty config with no error
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nonexistent", "profiles.yaml")

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom non-existent file returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadFrom returned nil config")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(cfg.Profiles))
	}
}

func TestAddAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profiles.yaml")

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom returned error: %v", err)
	}

	p := Profile{
		Nickname: "work",
		Name:     "John Doe",
		Email:    "john@work.com",
		Platform: "github",
	}

	if err := cfg.Add(p); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo returned error: %v", err)
	}

	// Reload and verify persistence
	cfg2, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom after save returned error: %v", err)
	}
	if len(cfg2.Profiles) != 1 {
		t.Fatalf("expected 1 profile after reload, got %d", len(cfg2.Profiles))
	}
	got := cfg2.Profiles[0]
	if got.Nickname != p.Nickname || got.Name != p.Name || got.Email != p.Email || got.Platform != p.Platform {
		t.Errorf("reloaded profile = %+v, want %+v", got, p)
	}
}

func TestAddDuplicateNickname(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profiles.yaml")

	cfg, _ := LoadFrom(path)

	p := Profile{
		Nickname: "work",
		Name:     "John Doe",
		Email:    "john@work.com",
		Platform: "github",
	}

	if err := cfg.Add(p); err != nil {
		t.Fatalf("first Add returned error: %v", err)
	}

	// Adding with the same nickname should fail
	p2 := Profile{
		Nickname: "work",
		Name:     "Jane Doe",
		Email:    "jane@work.com",
		Platform: "gitlab",
	}

	err := cfg.Add(p2)
	if err == nil {
		t.Fatal("expected error for duplicate nickname, got nil")
	}
}

func TestUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profiles.yaml")

	cfg, _ := LoadFrom(path)

	p := Profile{
		Nickname: "work",
		Name:     "John Doe",
		Email:    "john@work.com",
		Platform: "github",
	}
	_ = cfg.Add(p)

	updated := Profile{
		Nickname: "work",
		Name:     "John Smith",
		Email:    "john.smith@work.com",
		Platform: "github",
	}

	if err := cfg.Update("work", updated); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if len(cfg.Profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(cfg.Profiles))
	}
	got := cfg.Profiles[0]
	if got.Name != "John Smith" || got.Email != "john.smith@work.com" {
		t.Errorf("updated profile = %+v, want %+v", got, updated)
	}

	// Update non-existent nickname should fail
	err := cfg.Update("nonexistent", updated)
	if err == nil {
		t.Fatal("expected error updating non-existent nickname, got nil")
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profiles.yaml")

	cfg, _ := LoadFrom(path)

	p := Profile{
		Nickname: "work",
		Name:     "John Doe",
		Email:    "john@work.com",
		Platform: "github",
	}
	_ = cfg.Add(p)

	if err := cfg.Delete("work"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if len(cfg.Profiles) != 0 {
		t.Errorf("expected 0 profiles after delete, got %d", len(cfg.Profiles))
	}

	// Delete non-existent nickname should fail
	err := cfg.Delete("nonexistent")
	if err == nil {
		t.Fatal("expected error deleting non-existent nickname, got nil")
	}
}

func TestFindByEmail(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profiles.yaml")

	cfg, _ := LoadFrom(path)

	p := Profile{
		Nickname: "work",
		Name:     "John Doe",
		Email:    "john@work.com",
		Platform: "github",
	}
	_ = cfg.Add(p)

	// Find existing email
	found, ok := cfg.FindByEmail("john@work.com")
	if !ok {
		t.Fatal("FindByEmail did not find existing email")
	}
	if found.Nickname != "work" {
		t.Errorf("FindByEmail returned nickname %q, want %q", found.Nickname, "work")
	}

	// Find non-existing email
	_, ok = cfg.FindByEmail("nobody@example.com")
	if ok {
		t.Fatal("FindByEmail found non-existing email")
	}
}

func TestLoadCorruptYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profiles.yaml")

	// Write corrupt YAML
	if err := os.WriteFile(path, []byte("{{invalid yaml:::"), 0644); err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	cfg, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for corrupt YAML, got nil")
	}
	if cfg == nil {
		t.Fatal("expected non-nil config even on corrupt YAML")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("expected 0 profiles for corrupt YAML, got %d", len(cfg.Profiles))
	}

	// Verify backup was created
	bakPath := path + ".bak"
	if _, err := os.Stat(bakPath); os.IsNotExist(err) {
		t.Error("expected .bak file to be created for corrupt YAML")
	}
}
