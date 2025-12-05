package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/HammerSpb/aipaca/pkg/fileutil"
)

// Backup represents a stored backup
type Backup struct {
	Name      string
	Path      string
	RepoPath  string
	CreatedAt time.Time
	FileCount int
}

// ListBackups returns all available backups
func (s *Storage) ListBackups() ([]Backup, error) {
	backupsDir := s.cfg.BackupsPath()

	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read backups directory: %w", err)
	}

	var backups []Backup
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		backupPath := filepath.Join(backupsDir, name)

		fileCount, _ := fileutil.CountFiles(backupPath)
		createdAt := parseBackupTimestamp(name)

		backups = append(backups, Backup{
			Name:      name,
			Path:      backupPath,
			CreatedAt: createdAt,
			FileCount: fileCount,
		})
	}

	// Sort by creation time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// GetBackup returns a specific backup by name
func (s *Storage) GetBackup(name string) (*Backup, error) {
	backupPath := s.BackupPath(name)

	info, err := os.Stat(backupPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("backup '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to access backup: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("backup '%s' is not a directory", name)
	}

	fileCount, _ := fileutil.CountFiles(backupPath)
	createdAt := parseBackupTimestamp(name)

	return &Backup{
		Name:      name,
		Path:      backupPath,
		CreatedAt: createdAt,
		FileCount: fileCount,
	}, nil
}

// CreateBackup creates a backup of AI files from a repo
func (s *Storage) CreateBackup(repoPath string, patterns []string) (string, error) {
	// Generate backup name: reponame-timestamp
	repoName := filepath.Base(repoPath)
	timestamp := time.Now().Format("2006-01-02-150405")
	backupName := fmt.Sprintf("%s-%s", repoName, timestamp)
	backupPath := s.BackupPath(backupName)

	// Find all AI files in repo
	aiFiles, err := fileutil.ExpandPatterns(repoPath, patterns)
	if err != nil {
		return "", fmt.Errorf("failed to find AI files: %w", err)
	}

	if len(aiFiles) == 0 {
		// No files to backup - that's okay
		return "", nil
	}

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy each AI file/directory to backup
	for relPath, fullPath := range aiFiles {
		dstPath := filepath.Join(backupPath, relPath)
		if err := fileutil.CopyPath(fullPath, dstPath); err != nil {
			// Clean up partial backup
			os.RemoveAll(backupPath)
			return "", fmt.Errorf("failed to backup %s: %w", relPath, err)
		}
	}

	return backupName, nil
}

// RestoreBackup restores a backup to a repo
func (s *Storage) RestoreBackup(name string, repoPath string, patterns []string) error {
	backupPath := s.BackupPath(name)

	if !fileutil.Exists(backupPath) {
		return fmt.Errorf("backup '%s' not found", name)
	}

	// First, remove existing AI files from repo
	aiFiles, err := fileutil.ExpandPatterns(repoPath, patterns)
	if err != nil {
		return fmt.Errorf("failed to find existing AI files: %w", err)
	}

	for _, fullPath := range aiFiles {
		if err := os.RemoveAll(fullPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", fullPath, err)
		}
	}

	// Copy backup contents to repo
	entries, err := os.ReadDir(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(backupPath, entry.Name())
		dstPath := filepath.Join(repoPath, entry.Name())

		if err := fileutil.CopyPath(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// DeleteBackup deletes a backup
func (s *Storage) DeleteBackup(name string) error {
	backupPath := s.BackupPath(name)

	if !fileutil.Exists(backupPath) {
		return fmt.Errorf("backup '%s' not found", name)
	}

	if err := os.RemoveAll(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	return nil
}

// GetBackupsForRepo returns backups for a specific repo
func (s *Storage) GetBackupsForRepo(repoPath string) ([]Backup, error) {
	repoName := filepath.Base(repoPath)
	allBackups, err := s.ListBackups()
	if err != nil {
		return nil, err
	}

	var repoBackups []Backup
	for _, backup := range allBackups {
		if strings.HasPrefix(backup.Name, repoName+"-") {
			repoBackups = append(repoBackups, backup)
		}
	}

	return repoBackups, nil
}

// GetLatestBackupForRepo returns the most recent backup for a repo
func (s *Storage) GetLatestBackupForRepo(repoPath string) (*Backup, error) {
	backups, err := s.GetBackupsForRepo(repoPath)
	if err != nil {
		return nil, err
	}

	if len(backups) == 0 {
		return nil, nil
	}

	return &backups[0], nil // Already sorted newest first
}

// parseBackupTimestamp extracts timestamp from backup name
func parseBackupTimestamp(name string) time.Time {
	// Format: reponame-YYYY-MM-DD-HHMMSS
	parts := strings.Split(name, "-")
	if len(parts) >= 4 {
		// Try to parse the last 4 parts as timestamp
		timestampStr := strings.Join(parts[len(parts)-4:], "-")
		t, err := time.Parse("2006-01-02-150405", timestampStr)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}
