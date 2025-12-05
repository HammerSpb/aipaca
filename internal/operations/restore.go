package operations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HammerSpb/aipaca/internal/config"
	"github.com/HammerSpb/aipaca/internal/storage"
	"github.com/HammerSpb/aipaca/pkg/fileutil"
)

// RestoreOptions contains options for the restore operation
type RestoreOptions struct {
	RepoPath   string
	BackupName string // Specific backup to restore (empty = latest for this repo)
	DryRun     bool
}

// RestoreResult contains the result of a restore operation
type RestoreResult struct {
	BackupName     string
	FilesRestored  []string
	FilesRemoved   []string
	PreviousState  string // Previous applied profile
}

// Restore restores original AI files from backup
func Restore(cfg *config.Config, opts RestoreOptions) (*RestoreResult, error) {
	store := storage.New(cfg)
	result := &RestoreResult{}

	// Resolve repo path
	repoPath := opts.RepoPath
	if repoPath == "" {
		var err error
		repoPath, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}
	repoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repo path: %w", err)
	}

	// Determine which backup to restore
	backupName := opts.BackupName
	if backupName == "" {
		// Get backup from repo state
		state, err := store.GetRepoState(repoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get repo state: %w", err)
		}
		if state == nil || state.BackupPath == "" {
			return nil, fmt.Errorf("no backup found for this repository")
		}
		backupName = state.BackupPath
		result.PreviousState = state.AppliedProfile
	}

	result.BackupName = backupName

	// Verify backup exists
	backup, err := store.GetBackup(backupName)
	if err != nil {
		return nil, fmt.Errorf("backup not found: %w", err)
	}

	// Get files that will be restored
	restoredFiles, err := fileutil.ListAllFiles(backup.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup files: %w", err)
	}
	result.FilesRestored = restoredFiles

	// Get existing AI files that will be removed
	existingFiles, err := fileutil.ExpandPatterns(repoPath, cfg.AIPatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing AI files: %w", err)
	}
	for relPath := range existingFiles {
		result.FilesRemoved = append(result.FilesRemoved, relPath)
	}

	// If dry run, return here
	if opts.DryRun {
		return result, nil
	}

	// Restore the backup
	if err := store.RestoreBackup(backupName, repoPath, cfg.AIPatterns); err != nil {
		return nil, fmt.Errorf("failed to restore backup: %w", err)
	}

	// Clear the repo state
	if err := store.ClearRepoState(repoPath); err != nil {
		return nil, fmt.Errorf("failed to clear repo state: %w", err)
	}

	return result, nil
}
