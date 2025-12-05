package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/HammerSpb/aipaca/internal/operations"
)

var (
	cleanDryRun   bool
	cleanNoBackup bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean [repo-path]",
	Short: "Remove AI files from repository",
	Long: `Remove all AI files from the repository.

This is useful for:
- Creating clean commits/PRs without AI configuration
- Temporarily disabling AI tools

Files are backed up before removal (use --no-backup to skip).
Use 'aipaca restore' to bring them back.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := ""
		if len(args) > 0 {
			repoPath = args[0]
		}

		result, err := operations.Clean(cfg, operations.CleanOptions{
			RepoPath: repoPath,
			DryRun:   cleanDryRun,
			NoBackup: cleanNoBackup,
		})
		if err != nil {
			return err
		}

		if len(result.FilesRemoved) == 0 {
			fmt.Println("No AI files found in repository")
			return nil
		}

		if cleanDryRun {
			fmt.Println("Dry run - no changes made")
			fmt.Println()
			fmt.Println("Would remove:")
		} else {
			fmt.Println("Removed:")
		}

		for _, f := range result.FilesRemoved {
			printInfo("- %s", f)
		}

		if !cleanDryRun {
			fmt.Println()
			if result.BackupName != "" {
				printSuccess("Created backup: %s", result.BackupName)
			}
			printSuccess("Cleaned %d AI files/directories", len(result.FilesRemoved))
			fmt.Println()
			fmt.Println("Run 'aipaca restore' to bring them back 🦙")
		}

		return nil
	},
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Show what would happen without making changes")
	cleanCmd.Flags().BoolVar(&cleanNoBackup, "no-backup", false, "Skip creating backup (dangerous)")
}
