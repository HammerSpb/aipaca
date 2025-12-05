package operations

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/HammerSpb/aipaca/internal/config"
	"github.com/HammerSpb/aipaca/internal/storage"
	"github.com/HammerSpb/aipaca/pkg/fileutil"
)

// DiffOptions contains options for the diff operation
type DiffOptions struct {
	ProfileName string // Profile to compare against (empty = currently applied)
	RepoPath    string
}

// FileChange represents a change to a file
type FileChange struct {
	Path   string
	Type   string // "added", "removed", "modified"
}

// DiffResult contains the result of a diff operation
type DiffResult struct {
	ProfileName string
	Changes     []FileChange
	HasChanges  bool
}

// Diff shows differences between repo and a profile
func Diff(cfg *config.Config, opts DiffOptions) (*DiffResult, error) {
	store := storage.New(cfg)
	result := &DiffResult{}

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

	// Determine profile to compare against
	profileName := opts.ProfileName
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

	// Check if profile exists
	_, err = store.GetProfile(profileName)
	if err != nil {
		return nil, err
	}

	profilePath := store.ProfilePath(profileName)

	// Get files in profile
	profileFiles, err := fileutil.ListAllFiles(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list profile files: %w", err)
	}
	profileFileSet := make(map[string]bool)
	for _, f := range profileFiles {
		profileFileSet[f] = true
	}

	// Get AI files in repo
	aiFilesMap, err := fileutil.ExpandPatterns(repoPath, cfg.AIPatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to find AI files: %w", err)
	}

	// Get all files in repo AI directories
	repoFiles := make(map[string]bool)
	for relPath, fullPath := range aiFilesMap {
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		if info.IsDir() {
			files, err := fileutil.ListAllFiles(fullPath)
			if err != nil {
				continue
			}
			for _, f := range files {
				repoFiles[filepath.Join(relPath, f)] = true
			}
		} else {
			repoFiles[relPath] = true
		}
	}

	// Find differences
	// Files in repo but not in profile = added
	for f := range repoFiles {
		if !profileFileSet[f] {
			result.Changes = append(result.Changes, FileChange{Path: f, Type: "added"})
			result.HasChanges = true
		}
	}

	// Files in profile but not in repo = removed
	for f := range profileFileSet {
		if !repoFiles[f] {
			result.Changes = append(result.Changes, FileChange{Path: f, Type: "removed"})
			result.HasChanges = true
		}
	}

	// Files in both = check if modified
	for f := range repoFiles {
		if profileFileSet[f] {
			repoFilePath := filepath.Join(repoPath, f)
			profileFilePath := filepath.Join(profilePath, f)

			modified, err := filesAreDifferent(repoFilePath, profileFilePath)
			if err != nil {
				continue
			}
			if modified {
				result.Changes = append(result.Changes, FileChange{Path: f, Type: "modified"})
				result.HasChanges = true
			}
		}
	}

	return result, nil
}

// filesAreDifferent checks if two files have different content
func filesAreDifferent(path1, path2 string) (bool, error) {
	content1, err := os.ReadFile(path1)
	if err != nil {
		return false, err
	}

	content2, err := os.ReadFile(path2)
	if err != nil {
		return false, err
	}

	return !bytes.Equal(content1, content2), nil
}
