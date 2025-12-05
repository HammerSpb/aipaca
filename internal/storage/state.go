package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// RepoState represents the state of a repository
type RepoState struct {
	AppliedProfile string    `yaml:"applied_profile,omitempty"`
	AppliedAt      time.Time `yaml:"applied_at,omitempty"`
	BackupPath     string    `yaml:"backup_path,omitempty"`
}

// StateFile represents the state file structure
type StateFile struct {
	Repos map[string]*RepoState `yaml:"repos"`
}

// stateFilePath returns the path to the state file
func (s *Storage) stateFilePath() string {
	return filepath.Join(s.cfg.StatePath(), "repo-states.yaml")
}

// loadStateFile loads the state file
func (s *Storage) loadStateFile() (*StateFile, error) {
	statePath := s.stateFilePath()

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &StateFile{Repos: make(map[string]*RepoState)}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state StateFile
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	if state.Repos == nil {
		state.Repos = make(map[string]*RepoState)
	}

	return &state, nil
}

// saveStateFile saves the state file
func (s *Storage) saveStateFile(state *StateFile) error {
	statePath := s.stateFilePath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to serialize state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// GetRepoState returns the state for a repository
func (s *Storage) GetRepoState(repoPath string) (*RepoState, error) {
	// Normalize the repo path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	state, err := s.loadStateFile()
	if err != nil {
		return nil, err
	}

	repoState, exists := state.Repos[absPath]
	if !exists {
		return nil, nil
	}

	return repoState, nil
}

// SetRepoState sets the state for a repository
func (s *Storage) SetRepoState(repoPath string, repoState *RepoState) error {
	// Normalize the repo path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	state, err := s.loadStateFile()
	if err != nil {
		return err
	}

	if repoState == nil {
		delete(state.Repos, absPath)
	} else {
		state.Repos[absPath] = repoState
	}

	return s.saveStateFile(state)
}

// ClearRepoState clears the state for a repository
func (s *Storage) ClearRepoState(repoPath string) error {
	return s.SetRepoState(repoPath, nil)
}

// RecordApply records that a profile was applied to a repo
func (s *Storage) RecordApply(repoPath, profileName, backupName string) error {
	return s.SetRepoState(repoPath, &RepoState{
		AppliedProfile: profileName,
		AppliedAt:      time.Now(),
		BackupPath:     backupName,
	})
}

// GetAppliedProfile returns the profile currently applied to a repo
func (s *Storage) GetAppliedProfile(repoPath string) (string, error) {
	state, err := s.GetRepoState(repoPath)
	if err != nil {
		return "", err
	}
	if state == nil {
		return "", nil
	}
	return state.AppliedProfile, nil
}

// GetBackupForRepo returns the backup associated with the current state
func (s *Storage) GetBackupForRepo(repoPath string) (string, error) {
	state, err := s.GetRepoState(repoPath)
	if err != nil {
		return "", err
	}
	if state == nil {
		return "", nil
	}
	return state.BackupPath, nil
}
