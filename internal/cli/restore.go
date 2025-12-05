package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/HammerSpb/aipaca/internal/operations"
)

var (
	restoreDryRun bool
	restoreBackup string
)

var restoreCmd = &cobra.Command{
	Use:   "restore [repo-path]",
	Short: "Restore original AI files from backup",
	Long: `Restore the original AI files from the backup created during the last apply.

This will:
1. Remove currently applied profile files
2. Restore files from the backup
3. Clear the applied state

Use --backup to restore from a specific backup instead of the latest.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := ""
		if len(args) > 0 {
			repoPath = args[0]
		}

		result, err := operations.Restore(cfg, operations.RestoreOptions{
			RepoPath:   repoPath,
			BackupName: restoreBackup,
			DryRun:     restoreDryRun,
		})
		if err != nil {
			return err
		}

		if restoreDryRun {
			fmt.Println("Dry run - no changes made")
			fmt.Println()
		}

		if len(result.FilesRemoved) > 0 {
			if restoreDryRun {
				fmt.Println("Would remove current AI files:")
			} else {
				fmt.Println("Removed current AI files:")
			}
			for _, f := range result.FilesRemoved {
				printInfo("- %s", f)
			}
			fmt.Println()
		}

		if len(result.FilesRestored) > 0 {
			if restoreDryRun {
				fmt.Println("Would restore from backup '%s':", result.BackupName)
			} else {
				fmt.Println("Restored from backup '%s':", result.BackupName)
			}
			for _, f := range result.FilesRestored {
				printInfo("+ %s", f)
			}
			fmt.Println()
		}

		if !restoreDryRun {
			printSuccess("Restored original AI files")
			if result.PreviousState != "" {
				printInfo("Previously applied profile '%s' has been cleared", result.PreviousState)
			}
		}

		return nil
	},
}

func init() {
	restoreCmd.Flags().BoolVar(&restoreDryRun, "dry-run", false, "Show what would happen without making changes")
	restoreCmd.Flags().StringVar(&restoreBackup, "backup", "", "Restore from a specific backup")
}
