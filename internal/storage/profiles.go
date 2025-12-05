package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HammerSpb/aipaca/pkg/fileutil"
)

// Profile represents a stored AI config profile
type Profile struct {
	Name        string
	Path        string
	Description string
	FileCount   int
}

// ListProfiles returns all available profiles
func (s *Storage) ListProfiles() ([]Profile, error) {
	profilesDir := s.cfg.ProfilesPath()

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var profiles []Profile
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		profilePath := filepath.Join(profilesDir, name)

		fileCount, _ := fileutil.CountFiles(profilePath)

		description := ""
		if s.cfg.ProfileDescriptions != nil {
			description = s.cfg.ProfileDescriptions[name]
		}

		profiles = append(profiles, Profile{
			Name:        name,
			Path:        profilePath,
			Description: description,
			FileCount:   fileCount,
		})
	}

	return profiles, nil
}

// GetProfile returns a specific profile by name
func (s *Storage) GetProfile(name string) (*Profile, error) {
	profilePath := s.ProfilePath(name)

	info, err := os.Stat(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to access profile: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("profile '%s' is not a directory", name)
	}

	fileCount, _ := fileutil.CountFiles(profilePath)

	description := ""
	if s.cfg.ProfileDescriptions != nil {
		description = s.cfg.ProfileDescriptions[name]
	}

	return &Profile{
		Name:        name,
		Path:        profilePath,
		Description: description,
		FileCount:   fileCount,
	}, nil
}

// CreateProfile creates a new empty profile
func (s *Storage) CreateProfile(name string) error {
	profilePath := s.ProfilePath(name)

	if fileutil.Exists(profilePath) {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	if err := os.MkdirAll(profilePath, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	return nil
}

// DeleteProfile deletes a profile
func (s *Storage) DeleteProfile(name string) error {
	profilePath := s.ProfilePath(name)

	if !fileutil.Exists(profilePath) {
		return fmt.Errorf("profile '%s' not found", name)
	}

	if err := os.RemoveAll(profilePath); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	return nil
}

// CopyProfile copies a profile to a new name
func (s *Storage) CopyProfile(srcName, dstName string) error {
	srcPath := s.ProfilePath(srcName)
	dstPath := s.ProfilePath(dstName)

	if !fileutil.Exists(srcPath) {
		return fmt.Errorf("source profile '%s' not found", srcName)
	}

	if fileutil.Exists(dstPath) {
		return fmt.Errorf("destination profile '%s' already exists", dstName)
	}

	if err := fileutil.CopyDir(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to copy profile: %w", err)
	}

	return nil
}

// SaveToProfile saves files from a repo to a profile
func (s *Storage) SaveToProfile(name string, repoPath string, patterns []string, force bool) error {
	profilePath := s.ProfilePath(name)

	// Check if profile exists
	exists := fileutil.Exists(profilePath)
	if exists && !force {
		// We'll update the existing profile
	}

	// Find all AI files in repo
	aiFiles, err := fileutil.ExpandPatterns(repoPath, patterns)
	if err != nil {
		return fmt.Errorf("failed to find AI files: %w", err)
	}

	if len(aiFiles) == 0 {
		return fmt.Errorf("no AI files found in repository")
	}

	// Clear existing profile if it exists
	if exists {
		if err := os.RemoveAll(profilePath); err != nil {
			return fmt.Errorf("failed to clear existing profile: %w", err)
		}
	}

	// Create profile directory
	if err := os.MkdirAll(profilePath, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Copy each AI file/directory to profile
	for relPath, fullPath := range aiFiles {
		dstPath := filepath.Join(profilePath, relPath)
		if err := fileutil.CopyPath(fullPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", relPath, err)
		}
	}

	return nil
}

// ApplyProfile copies a profile to a repo
func (s *Storage) ApplyProfile(name string, repoPath string) error {
	profilePath := s.ProfilePath(name)

	if !fileutil.Exists(profilePath) {
		return fmt.Errorf("profile '%s' not found", name)
	}

	// List all items in profile
	entries, err := os.ReadDir(profilePath)
	if err != nil {
		return fmt.Errorf("failed to read profile: %w", err)
	}

	// Copy each item to repo
	for _, entry := range entries {
		srcPath := filepath.Join(profilePath, entry.Name())
		dstPath := filepath.Join(repoPath, entry.Name())

		if err := fileutil.CopyPath(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// GetProfileFiles returns the list of files in a profile
func (s *Storage) GetProfileFiles(name string) ([]string, error) {
	profilePath := s.ProfilePath(name)

	if !fileutil.Exists(profilePath) {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}

	return fileutil.ListAllFiles(profilePath)
}
