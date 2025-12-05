package operations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HammerSpb/aipaca/internal/config"
	"github.com/HammerSpb/aipaca/internal/storage"
	"github.com/HammerSpb/aipaca/pkg/fileutil"
)

// CleanOptions contains options for the clean operation
type CleanOptions struct {
	RepoPath string
	DryRun   bool
	NoBackup bool
}

// CleanResult contains the result of a clean operation
type CleanResult struct {
	BackupName   string
	FilesRemoved []string
}

// Clean removes AI files from a repository
func Clean(cfg *config.Config, opts CleanOptions) (*CleanResult, error) {
	store := storage.New(cfg)
	result := &CleanResult{}

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

	// Find AI files in repo
	aiFiles, err := fileutil.ExpandPatterns(repoPath, cfg.AIPatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to find AI files: %w", err)
	}

	if len(aiFiles) == 0 {
		return result, nil // Nothing to clean
	}

	// Get list of files
	for relPath := range aiFiles {
		result.FilesRemoved = append(result.FilesRemoved, relPath)
	}

	// If dry run, return here
	if opts.DryRun {
		return result, nil
	}

	// Create backup (unless --no-backup)
	if !opts.NoBackup {
		backupName, err := store.CreateBackup(repoPath, cfg.AIPatterns)
		if err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}
		result.BackupName = backupName

		// Record clean state (so restore knows the backup)
		if err := store.SetRepoState(repoPath, &storage.RepoState{
			BackupPath: backupName,
		}); err != nil {
			return nil, fmt.Errorf("failed to record state: %w", err)
		}
	}

	// Remove AI files
	for _, fullPath := range aiFiles {
		if err := os.RemoveAll(fullPath); err != nil {
			return nil, fmt.Errorf("failed to remove %s: %w", fullPath, err)
		}
	}

	return result, nil
}
