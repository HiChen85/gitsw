package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Profile represents a git user profile.
type Profile struct {
	Nickname string `yaml:"nickname"`
	Name     string `yaml:"name"`
	Email    string `yaml:"email"`
	Platform string `yaml:"platform"`
}

// Config holds a collection of profiles.
type Config struct {
	Profiles []Profile `yaml:"profiles"`
}

// DefaultPath returns the default path for the profiles file.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gitswitch", "profiles.yaml")
}

// Load loads the config from the default path.
func Load() (*Config, error) {
	return LoadFrom(DefaultPath())
}

// LoadFrom loads the config from the given path. If the file does not exist,
// it returns an empty config with no error. If the file contains corrupt YAML,
// it backs up the file to .bak and returns an empty config with an error message.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return &Config{}, nil
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// Back up corrupt file
		bakPath := path + ".bak"
		_ = os.WriteFile(bakPath, data, 0644)
		return &Config{}, fmt.Errorf("corrupt config at %s (backed up to %s): %w", path, bakPath, err)
	}

	return &cfg, nil
}

// Save saves the config to the default path.
func (c *Config) Save() error {
	return c.SaveTo(DefaultPath())
}

// SaveTo saves the config to the given path, creating directories as needed.
func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", path, err)
	}

	return nil
}

// Add adds a profile to the config. Returns an error if a profile with the
// same nickname already exists.
func (c *Config) Add(p Profile) error {
	for _, existing := range c.Profiles {
		if existing.Nickname == p.Nickname {
			return fmt.Errorf("profile with nickname %q already exists", p.Nickname)
		}
	}
	c.Profiles = append(c.Profiles, p)
	return nil
}

// Update updates an existing profile by nickname.
func (c *Config) Update(nickname string, p Profile) error {
	for i, existing := range c.Profiles {
		if existing.Nickname == nickname {
			c.Profiles[i] = p
			return nil
		}
	}
	return fmt.Errorf("profile with nickname %q not found", nickname)
}

// Delete removes a profile by nickname.
func (c *Config) Delete(nickname string) error {
	for i, existing := range c.Profiles {
		if existing.Nickname == nickname {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("profile with nickname %q not found", nickname)
}

// FindByEmail finds a profile by email address.
func (c *Config) FindByEmail(email string) (Profile, bool) {
	for _, p := range c.Profiles {
		if p.Email == email {
			return p, true
		}
	}
	return Profile{}, false
}
