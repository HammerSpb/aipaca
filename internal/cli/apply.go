package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/HammerSpb/aipaca/internal/operations"
)

var (
	applyDryRun   bool
	applyNoBackup bool
	applyForce    bool
)

var applyCmd = &cobra.Command{
	Use:   "apply <profile> [repo-path]",
	Short: "Apply a profile to a repository",
	Long: `Apply a profile to a repository.

This will:
1. Backup existing AI files in the repo
2. Copy the profile files to the repo
3. Record the state for future operations

Use --dry-run to preview what would happen.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		repoPath := ""
		if len(args) > 1 {
			repoPath = args[1]
		}

		result, err := operations.Apply(cfg, operations.ApplyOptions{
			ProfileName: profileName,
			RepoPath:    repoPath,
			DryRun:      applyDryRun,
			NoBackup:    applyNoBackup,
			Force:       applyForce,
		})
		if err != nil {
			return err
		}

		if applyDryRun {
			fmt.Println("Dry run - no changes made")
			fmt.Println()
		}

		if len(result.FilesRemoved) > 0 {
			if applyDryRun {
				fmt.Println("Would remove existing AI files:")
			} else {
				fmt.Println("Removed existing AI files:")
			}
			for _, f := range result.FilesRemoved {
				printInfo("- %s", f)
			}
			fmt.Println()
		}

		if len(result.FilesApplied) > 0 {
			if applyDryRun {
				fmt.Printf("Would apply from profile '%s':\n", result.ProfileName)
			} else {
				fmt.Printf("Applied from profile '%s':\n", result.ProfileName)
			}
			for _, f := range result.FilesApplied {
				printInfo("+ %s", f)
			}
			fmt.Println()
		}

		if !applyDryRun {
			if result.BackupName != "" {
				printSuccess("Created backup: %s", result.BackupName)
			}
			printSuccess("Applied profile '%s'", result.ProfileName)
		}

		return nil
	},
}

func init() {
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "Show what would happen without making changes")
	applyCmd.Flags().BoolVar(&applyNoBackup, "no-backup", false, "Skip creating backup (dangerous)")
	applyCmd.Flags().BoolVar(&applyForce, "force", false, "Force apply even if there are issues")
}
