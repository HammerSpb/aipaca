package operations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HammerSpb/aipaca/internal/config"
	"github.com/HammerSpb/aipaca/internal/storage"
	"github.com/HammerSpb/aipaca/pkg/fileutil"
)

// ApplyOptions contains options for the apply operation
type ApplyOptions struct {
	ProfileName string
	RepoPath    string
	DryRun      bool
	NoBackup    bool
	Force       bool
}

// ApplyResult contains the result of an apply operation
type ApplyResult struct {
	ProfileName  string
	BackupName   string
	FilesApplied []string
	FilesRemoved []string
}

// Apply applies a profile to a repository
func Apply(cfg *config.Config, opts ApplyOptions) (*ApplyResult, error) {
	store := storage.New(cfg)
	result := &ApplyResult{ProfileName: opts.ProfileName}

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

	// Check if profile exists
	profile, err := store.GetProfile(opts.ProfileName)
	if err != nil {
		return nil, err
	}

	// Get list of files that will be applied
	profileFiles, err := store.GetProfileFiles(opts.ProfileName)
	if err != nil {
		return nil, fmt.Errorf("failed to list profile files: %w", err)
	}
	result.FilesApplied = profileFiles

	// Find existing AI files in repo that will be removed/replaced
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

	// Create backup of existing files (unless --no-backup)
	if !opts.NoBackup && len(existingFiles) > 0 {
		backupName, err := store.CreateBackup(repoPath, cfg.AIPatterns)
		if err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}
		result.BackupName = backupName
	}

	// Remove existing AI files from repo
	for _, fullPath := range existingFiles {
		if err := os.RemoveAll(fullPath); err != nil {
			return nil, fmt.Errorf("failed to remove %s: %w", fullPath, err)
		}
	}

	// Copy profile files to repo
	if err := store.ApplyProfile(opts.ProfileName, repoPath); err != nil {
		return nil, fmt.Errorf("failed to apply profile: %w", err)
	}

	// Record the state
	if err := store.RecordApply(repoPath, opts.ProfileName, result.BackupName); err != nil {
		return nil, fmt.Errorf("failed to record state: %w", err)
	}

	_ = profile // Use profile variable
	return result, nil
}
