package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/HammerSpb/aipaca/internal/config"
)

var (
	cfgFile string
	cfg     *config.Config
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "aipaca",
	Short: "🦙 aipaca - herd your AI configs across repos",
	Long: `🦙 aipaca - the adorable AI config manager

aipaca helps you herd AI-related configuration files (Claude, Cursor, etc.)
across multiple repositories like a well-trained alpaca.

It allows you to:
  - Apply profiles: Deploy AI configs from storage to a repo
  - Save profiles: Save repo AI configs back to storage
  - Clean: Remove AI files from repo (with backup)
  - Restore: Bring back original files from backup

No more copy-pasting configs between repos. Let aipaca do the heavy lifting! 🦙`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for init command
		if cmd.Name() == "init" {
			return nil
		}

		// Load config
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return err
		}
		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.aipaca.yaml)")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(saveCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(profilesCmd)
	rootCmd.AddCommand(diffCmd)
}

// printSuccess prints a success message in green
func printSuccess(format string, args ...interface{}) {
	fmt.Printf("\033[32m✓\033[0m "+format+"\n", args...)
}

// printError prints an error message in red
func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31m✗\033[0m "+format+"\n", args...)
}

// printInfo prints an info message
func printInfo(format string, args ...interface{}) {
	fmt.Printf("  "+format+"\n", args...)
}

// printWarning prints a warning message in yellow
func printWarning(format string, args ...interface{}) {
	fmt.Printf("\033[33m!\033[0m "+format+"\n", args...)
}
