package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/HammerSpb/aipaca/internal/storage"
)

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List and manage profiles 🦙",
	Long:  `List and manage AI config profiles in your aipaca herd.`,
}

var profilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := storage.New(cfg)

		profiles, err := store.ListProfiles()
		if err != nil {
			return err
		}

		if len(profiles) == 0 {
			fmt.Println("No profiles found in the herd 🦙")
			fmt.Println()
			fmt.Println("Create one with:")
			fmt.Println("  aipaca save --as <profile-name>")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PROFILE\tDESCRIPTION\tFILES")
		fmt.Fprintln(w, "-------\t-----------\t-----")

		for _, p := range profiles {
			desc := p.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			if desc == "" {
				desc = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%d\n", p.Name, desc, p.FileCount)
		}
		w.Flush()

		return nil
	},
}

var profilesShowCmd = &cobra.Command{
	Use:   "show <profile>",
	Short: "Show profile contents",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		store := storage.New(cfg)

		profile, err := store.GetProfile(profileName)
		if err != nil {
			return err
		}

		fmt.Printf("Profile: %s\n", profile.Name)
		if profile.Description != "" {
			fmt.Printf("Description: %s\n", profile.Description)
		}
		fmt.Printf("Files: %d\n", profile.FileCount)
		fmt.Println()

		files, err := store.GetProfileFiles(profileName)
		if err != nil {
			return err
		}

		fmt.Println("Contents:")
		for _, f := range files {
			fmt.Printf("  %s\n", f)
		}

		return nil
	},
}

var profilesDeleteCmd = &cobra.Command{
	Use:   "delete <profile>",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		store := storage.New(cfg)

		// Verify profile exists
		_, err := store.GetProfile(profileName)
		if err != nil {
			return err
		}

		if err := store.DeleteProfile(profileName); err != nil {
			return err
		}

		printSuccess("Deleted profile '%s'", profileName)
		return nil
	},
}

var profilesCopyCmd = &cobra.Command{
	Use:   "copy <source> <destination>",
	Short: "Copy a profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcName := args[0]
		dstName := args[1]
		store := storage.New(cfg)

		if err := store.CopyProfile(srcName, dstName); err != nil {
			return err
		}

		printSuccess("Copied profile '%s' to '%s'", srcName, dstName)
		return nil
	},
}

func init() {
	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesShowCmd)
	profilesCmd.AddCommand(profilesDeleteCmd)
	profilesCmd.AddCommand(profilesCopyCmd)
}
