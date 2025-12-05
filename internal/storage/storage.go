package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HammerSpb/aipaca/internal/config"
	"github.com/HammerSpb/aipaca/pkg/fileutil"
)

// Storage manages the AI config storage directory
type Storage struct {
	cfg *config.Config
}

// New creates a new Storage instance
func New(cfg *config.Config) *Storage {
	return &Storage{cfg: cfg}
}

// Init initializes the storage directory structure
func (s *Storage) Init() error {
	// Create main directories
	dirs := []string{
		s.cfg.StoragePath(),
		s.cfg.ProfilesPath(),
		s.cfg.BackupsPath(),
		s.cfg.StatePath(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// IsInitialized checks if storage is initialized
func (s *Storage) IsInitialized() bool {
	return fileutil.IsDir(s.cfg.StoragePath()) &&
		fileutil.IsDir(s.cfg.ProfilesPath()) &&
		fileutil.IsDir(s.cfg.BackupsPath()) &&
		fileutil.IsDir(s.cfg.StatePath())
}

// ProfilePath returns the full path to a profile directory
func (s *Storage) ProfilePath(name string) string {
	return filepath.Join(s.cfg.ProfilesPath(), name)
}

// BackupPath returns the full path to a backup directory
func (s *Storage) BackupPath(name string) string {
	return filepath.Join(s.cfg.BackupsPath(), name)
}

// ProfileExists checks if a profile exists
func (s *Storage) ProfileExists(name string) bool {
	return fileutil.IsDir(s.ProfilePath(name))
}

// BackupExists checks if a backup exists
func (s *Storage) BackupExists(name string) bool {
	return fileutil.IsDir(s.BackupPath(name))
}
