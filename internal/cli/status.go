package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/HammerSpb/aipaca/internal/operations"
	"github.com/HammerSpb/aipaca/internal/storage"
	"github.com/HammerSpb/aipaca/pkg/fileutil"
)

var statusCmd = &cobra.Command{
	Use:   "status [repo-path]",
	Short: "Show current state of AI files",
	Long: `Show the current state of AI files in a repository.

Displays:
- Currently applied profile (if any)
- Whether files have been modified since apply
- List of AI files in the repo
- Available backups`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := ""
		if len(args) > 0 {
			repoPath = args[0]
		}

		// Resolve repo path
		if repoPath == "" {
			var err error
			repoPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
		}
		repoPath, err := filepath.Abs(repoPath)
		if err != nil {
			return fmt.Errorf("failed to resolve repo path: %w", err)
		}

		store := storage.New(cfg)

		fmt.Printf("Repo: %s\n", repoPath)
		fmt.Println()

		// Get repo state
		state, err := store.GetRepoState(repoPath)
		if err != nil {
			return fmt.Errorf("failed to get repo state: %w", err)
		}

		if state != nil && state.AppliedProfile != "" {
			// Check for modifications
			diffResult, _ := operations.Diff(cfg, operations.DiffOptions{
				ProfileName: state.AppliedProfile,
				RepoPath:    repoPath,
			})

			modifiedStr := ""
			if diffResult != nil && diffResult.HasChanges {
				modifiedStr = " \033[33m(modified)\033[0m"
			}

			fmt.Printf("Applied profile: \033[36m%s\033[0m%s\n", state.AppliedProfile, modifiedStr)
			fmt.Printf("Applied at: %s\n", state.AppliedAt.Format("2006-01-02 15:04:05"))
			if state.BackupPath != "" {
				fmt.Printf("Backup: %s\n", state.BackupPath)
			}
			fmt.Println()

			// Show changes if any
			if diffResult != nil && diffResult.HasChanges {
				fmt.Println("Changes since apply:")
				for _, change := range diffResult.Changes {
					switch change.Type {
					case "added":
						fmt.Printf("  \033[32mA\033[0m %s\n", change.Path)
					case "removed":
						fmt.Printf("  \033[31mD\033[0m %s\n", change.Path)
					case "modified":
						fmt.Printf("  \033[33mM\033[0m %s\n", change.Path)
					}
				}
				fmt.Println()
			}
		} else {
			fmt.Println("No profile currently applied")
			fmt.Println()
		}

		// Show AI files in repo
		aiFiles, err := fileutil.ExpandPatterns(repoPath, cfg.AIPatterns)
		if err != nil {
			return fmt.Errorf("failed to find AI files: %w", err)
		}

		if len(aiFiles) > 0 {
			fmt.Println("AI files in repo:")
			for relPath, fullPath := range aiFiles {
				info, err := os.Stat(fullPath)
				if err != nil {
					continue
				}
				if info.IsDir() {
					count, _ := fileutil.CountFiles(fullPath)
					fmt.Printf("  %s/ (%d files)\n", relPath, count)
				} else {
					fmt.Printf("  %s\n", relPath)
				}
			}
		} else {
			fmt.Println("No AI files in repo")
		}

		// Show available backups for this repo
		backups, err := store.GetBackupsForRepo(repoPath)
		if err == nil && len(backups) > 0 {
			fmt.Println()
			fmt.Printf("Available backups (%d):\n", len(backups))
			for i, backup := range backups {
				if i >= 5 {
					fmt.Printf("  ... and %d more\n", len(backups)-5)
					break
				}
				fmt.Printf("  %s (%d files)\n", backup.Name, backup.FileCount)
			}
		}

		return nil
	},
}
