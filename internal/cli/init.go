package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/HammerSpb/aipaca/internal/config"
	"github.com/HammerSpb/aipaca/internal/storage"
)

var initImport bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize aipaca 🦙",
	Long: `Initialize aipaca by creating the configuration file and storage directory.

Use --import to also import AI files from the current repository as the "default" profile.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if config already exists
		if config.Exists(cfgFile) {
			printWarning("Config file already exists at %s", config.ConfigPath())
			// Load existing config
			var err error
			cfg, err = config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load existing config: %w", err)
			}
		} else {
			// Create default config
			cfg = config.DefaultConfig()

			// Save config
			if err := cfg.Save(cfgFile); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			printSuccess("Created config file at %s", config.ConfigPath())
		}

		// Initialize storage
		store := storage.New(cfg)
		if err := store.Init(); err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}
		printSuccess("Initialized storage at %s", cfg.StoragePath())

		// Import current repo as default profile if requested
		if initImport {
			repoPath, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			err = store.SaveToProfile("default", repoPath, cfg.AIPatterns, true)
			if err != nil {
				printWarning("No AI files found in current directory to import")
			} else {
				printSuccess("Imported AI files from current directory as 'default' profile")
			}
		}

		fmt.Println()
		fmt.Println("🦙 aipaca is ready to herd your configs! Try these commands:")
		fmt.Println("  aipaca status          # Check current state")
		fmt.Println("  aipaca profiles list   # List available profiles")
		fmt.Println("  aipaca apply <profile> # Apply a profile")

		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initImport, "import", false, "Import current repo AI files as 'default' profile")
}
