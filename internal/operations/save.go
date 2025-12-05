package operations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HammerSpb/aipaca/internal/config"
	"github.com/HammerSpb/aipaca/internal/storage"
	"github.com/HammerSpb/aipaca/pkg/fileutil"
)

// SaveOptions contains options for the save operation
type SaveOptions struct {
	ProfileName string // Profile to save to (empty = currently applied)
	AsName      string // Save as new profile with this name
	RepoPath    string
	DryRun      bool
	Force       bool
}

// SaveResult contains the result of a save operation
type SaveResult struct {
	ProfileName string
	FilesSaved  []string
	IsNew       bool
}

// Save saves repo AI files to a profile
func Save(cfg *config.Config, opts SaveOptions) (*SaveResult, error) {
	store := storage.New(cfg)
	result := &SaveResult{}

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

	// Determine target profile name
	profileName := opts.ProfileName
	if opts.AsName != "" {
		profileName = opts.AsName
		result.IsNew = true
	}

	// If no profile specified, use currently applied profile
	if profileName == "" {
		appliedProfile, err := store.GetAppliedProfile(repoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get applied profile: %w", err)
		}
		if appliedProfile == "" {
			return nil, fmt.Errorf("no profile specified and no profile currently applied to this repo")
		}
		profileName = appliedProfile
	}

	result.ProfileName = profileName

	// Check if profile already exists
	if store.ProfileExists(profileName) && result.IsNew && !opts.Force {
		return nil, fmt.Errorf("profile '%s' already exists (use --force to overwrite)", profileName)
	}

	// Find AI files in repo
	aiFiles, err := fileutil.ExpandPatterns(repoPath, cfg.AIPatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to find AI files: %w", err)
	}

	if len(aiFiles) == 0 {
		return nil, fmt.Errorf("no AI files found in repository")
	}

	// Get list of files
	for relPath := range aiFiles {
		result.FilesSaved = append(result.FilesSaved, relPath)
	}

	// If dry run, return here
	if opts.DryRun {
		return result, nil
	}

	// Save to profile
	if err := store.SaveToProfile(profileName, repoPath, cfg.AIPatterns, opts.Force || !result.IsNew); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return result, nil
}
