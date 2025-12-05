package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/HammerSpb/aipaca/internal/operations"
)

var (
	saveDryRun bool
	saveAsName string
	saveForce  bool
)

var saveCmd = &cobra.Command{
	Use:   "save [profile] [repo-path]",
	Short: "Save repo AI files to a profile",
	Long: `Save AI files from a repository to a profile.

If no profile is specified, saves to the currently applied profile.
Use --as to save as a new profile.

Examples:
  aiconfig save                    # Update currently applied profile
  aiconfig save default            # Update 'default' profile
  aiconfig save --as my-project    # Create new 'my-project' profile`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := ""
		repoPath := ""

		if len(args) > 0 {
			profileName = args[0]
		}
		if len(args) > 1 {
			repoPath = args[1]
		}

		result, err := operations.Save(cfg, operations.SaveOptions{
			ProfileName: profileName,
			AsName:      saveAsName,
			RepoPath:    repoPath,
			DryRun:      saveDryRun,
			Force:       saveForce,
		})
		if err != nil {
			return err
		}

		if saveDryRun {
			fmt.Println("Dry run - no changes made")
			fmt.Println()
		}

		action := "Saved"
		if result.IsNew {
			action = "Created new profile"
		} else {
			action = "Updated profile"
		}

		if saveDryRun {
			action = "Would save"
		}

		fmt.Printf("%s '%s' with %d files:\n", action, result.ProfileName, len(result.FilesSaved))
		for _, f := range result.FilesSaved {
			printInfo("  %s", f)
		}

		if !saveDryRun {
			if result.IsNew {
				printSuccess("Created new profile '%s'", result.ProfileName)
			} else {
				printSuccess("Updated profile '%s'", result.ProfileName)
			}
		}

		return nil
	},
}

func init() {
	saveCmd.Flags().BoolVar(&saveDryRun, "dry-run", false, "Show what would happen without making changes")
	saveCmd.Flags().StringVar(&saveAsName, "as", "", "Save as a new profile with this name")
	saveCmd.Flags().BoolVar(&saveForce, "force", false, "Overwrite existing profile without confirmation")
}
