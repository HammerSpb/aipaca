package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/HammerSpb/aipaca/internal/storage"
)

var (
	backupsRepoPath string
	backupsPruneKeep int
)

var backupsCmd = &cobra.Command{
	Use:   "backups",
	Short: "List and manage backups",
	Long:  `List and manage AI config backups created by aipaca.`,
}

var backupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all backups",
	Long: `List all available backups.

Use --repo to filter backups for a specific repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store := storage.New(cfg)

		var backups []storage.Backup
		var err error

		if backupsRepoPath != "" {
			repoPath, err := filepath.Abs(backupsRepoPath)
			if err != nil {
				return fmt.Errorf("invalid repo path: %w", err)
			}
			backups, err = store.GetBackupsForRepo(repoPath)
			if err != nil {
				return err
			}
		} else {
			backups, err = store.ListBackups()
			if err != nil {
				return err
			}
		}

		if len(backups) == 0 {
			fmt.Println("No backups found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "BACKUP\tFILES\tCREATED")
		fmt.Fprintln(w, "------\t-----\t-------")

		for _, b := range backups {
			created := "-"
			if !b.CreatedAt.IsZero() {
				created = b.CreatedAt.Format("2006-01-02 15:04:05")
			}
			fmt.Fprintf(w, "%s\t%d\t%s\n", b.Name, b.FileCount, created)
		}
		w.Flush()

		return nil
	},
}

var backupsShowCmd = &cobra.Command{
	Use:   "show <backup>",
	Short: "Show backup contents",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupName := args[0]
		store := storage.New(cfg)

		backup, err := store.GetBackup(backupName)
		if err != nil {
			return err
		}

		fmt.Printf("Backup: %s\n", backup.Name)
		if !backup.CreatedAt.IsZero() {
			fmt.Printf("Created: %s\n", backup.CreatedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("Files: %d\n", backup.FileCount)
		fmt.Println()

		files, err := store.GetBackupFiles(backupName)
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

var backupsDeleteCmd = &cobra.Command{
	Use:   "delete <backup>",
	Short: "Delete a backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupName := args[0]
		store := storage.New(cfg)

		// Verify backup exists
		_, err := store.GetBackup(backupName)
		if err != nil {
			return err
		}

		if err := store.DeleteBackup(backupName); err != nil {
			return err
		}

		printSuccess("Deleted backup '%s'", backupName)
		return nil
	},
}

var backupsPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete old backups, keeping the most recent ones",
	Long: `Delete old backups while keeping the N most recent ones.

By default, keeps the 5 most recent backups. Use --keep to change this.
Use --repo to prune backups only for a specific repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store := storage.New(cfg)

		var backups []storage.Backup
		var err error

		if backupsRepoPath != "" {
			repoPath, err := filepath.Abs(backupsRepoPath)
			if err != nil {
				return fmt.Errorf("invalid repo path: %w", err)
			}
			backups, err = store.GetBackupsForRepo(repoPath)
			if err != nil {
				return err
			}
		} else {
			backups, err = store.ListBackups()
			if err != nil {
				return err
			}
		}

		if len(backups) <= backupsPruneKeep {
			fmt.Printf("Nothing to prune (have %d backups, keeping %d)\n", len(backups), backupsPruneKeep)
			return nil
		}

		// Backups are already sorted newest first, so skip the first N
		toDelete := backups[backupsPruneKeep:]
		deleted := 0

		for _, backup := range toDelete {
			if err := store.DeleteBackup(backup.Name); err != nil {
				printWarning("Failed to delete '%s': %v", backup.Name, err)
				continue
			}
			printInfo("Deleted: %s", backup.Name)
			deleted++
		}

		if deleted > 0 {
			printSuccess("Pruned %d old backup(s), kept %d most recent", deleted, backupsPruneKeep)
		}

		return nil
	},
}

func init() {
	backupsListCmd.Flags().StringVar(&backupsRepoPath, "repo", "", "Filter backups for a specific repository path")
	backupsPruneCmd.Flags().StringVar(&backupsRepoPath, "repo", "", "Prune backups only for a specific repository path")
	backupsPruneCmd.Flags().IntVar(&backupsPruneKeep, "keep", 5, "Number of recent backups to keep")

	backupsCmd.AddCommand(backupsListCmd)
	backupsCmd.AddCommand(backupsShowCmd)
	backupsCmd.AddCommand(backupsDeleteCmd)
	backupsCmd.AddCommand(backupsPruneCmd)
}
