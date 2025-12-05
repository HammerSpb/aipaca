package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration file
type Config struct {
	Version             string            `yaml:"version"`
	Storage             StorageConfig     `yaml:"storage"`
	AIPatterns          []string          `yaml:"ai_patterns"`
	DefaultProfile      string            `yaml:"default_profile"`
	ProfileDescriptions map[string]string `yaml:"profile_descriptions"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Path string `yaml:"path"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "1",
		Storage: StorageConfig{
			Path: "~/.aipaca",
		},
		AIPatterns: []string{
			".claude",
			".claude/**",
			".cursor",
			".cursor/**",
			"CLAUDE.md",
			"**/CLAUDE.md",
			"ai/",
			"ai/**",
			".ai*",
		},
		DefaultProfile:      "default",
		ProfileDescriptions: map[string]string{},
	}
}

// ConfigPath returns the default config file path
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".aipaca.yaml"
	}
	return filepath.Join(home, ".aipaca.yaml")
}

// Load loads configuration from the default path or specified path
func Load(path string) (*Config, error) {
	if path == "" {
		path = ConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s, run 'aipaca init' first", path)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Expand ~ in storage path
	cfg.Storage.Path = expandPath(cfg.Storage.Path)

	return &cfg, nil
}

// Save saves the configuration to the specified path
func (c *Config) Save(path string) error {
	if path == "" {
		path = ConfigPath()
	}

	// Create parent directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// StoragePath returns the expanded storage path
func (c *Config) StoragePath() string {
	return expandPath(c.Storage.Path)
}

// ProfilesPath returns the path to the profiles directory
func (c *Config) ProfilesPath() string {
	return filepath.Join(c.StoragePath(), "profiles")
}

// BackupsPath returns the path to the backups directory
func (c *Config) BackupsPath() string {
	return filepath.Join(c.StoragePath(), "backups")
}

// StatePath returns the path to the state directory
func (c *Config) StatePath() string {
	return filepath.Join(c.StoragePath(), "state")
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if len(path) == 0 {
		return path
	}

	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}

	return path
}

// Exists checks if config file exists
func Exists(path string) bool {
	if path == "" {
		path = ConfigPath()
	}
	_, err := os.Stat(path)
	return err == nil
}
