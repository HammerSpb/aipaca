package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/HammerSpb/aipaca/internal/operations"
)

var diffCmd = &cobra.Command{
	Use:   "diff [profile] [repo-path]",
	Short: "Show differences between repo and profile",
	Long: `Show differences between the AI files in a repository and a profile.

If no profile is specified, compares against the currently applied profile.`,
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

		result, err := operations.Diff(cfg, operations.DiffOptions{
			ProfileName: profileName,
			RepoPath:    repoPath,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Comparing against profile '%s'\n", result.ProfileName)
		fmt.Println()

		if !result.HasChanges {
			fmt.Println("No differences found")
			return nil
		}

		fmt.Println("Changes:")
		for _, change := range result.Changes {
			switch change.Type {
			case "added":
				fmt.Printf("  \033[32m+ %s\033[0m (added in repo)\n", change.Path)
			case "removed":
				fmt.Printf("  \033[31m- %s\033[0m (missing from repo)\n", change.Path)
			case "modified":
				fmt.Printf("  \033[33mM %s\033[0m (modified)\n", change.Path)
			}
		}

		return nil
	},
}
