// cmd/sync.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Xanonymous-GitHub/claude-switch/internal/config"
	"github.com/Xanonymous-GitHub/claude-switch/internal/diff"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [config-name-or-id]",
	Short: "Sync changes from live settings.json back to stored config",
	Long: `Detect and save changes from ~/.claude/settings.json back to the stored configuration.

This command will:
1. Compare live settings.json with the stored configuration
2. Display detected changes in a diff format
3. Check for conflicts (if stored config was modified externally)
4. Prompt for confirmation before saving (unless --force is used)
5. Save changes back to the stored configuration

This is useful when you've made manual edits to Claude Code settings
and want to preserve them in your stored configuration.`,
	Example: `  # Sync current configuration
  claude-switch sync

  # Sync specific configuration
  claude-switch sync my-config

  # Sync without confirmation
  claude-switch sync --force

  # Show changes without saving
  claude-switch sync --dry-run

  # Use different verbosity levels
  claude-switch sync --quiet       # Minimal output
  claude-switch sync --verbose     # Detailed output`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSync,
}

func init() {
	syncCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt and save automatically")
	syncCmd.Flags().BoolP("dry-run", "n", false, "Show changes without saving")
	syncCmd.Flags().BoolP("quiet", "q", false, "Minimal output (only errors and confirmations)")
	syncCmd.Flags().BoolP("verbose", "v", false, "Verbose output with detailed diff")
	syncCmd.Flags().Bool("no-color", false, "Disable colored output")
}

func runSync(cmd *cobra.Command, args []string) error {
	// Get flags
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	quiet, _ := cmd.Flags().GetBool("quiet")
	verbose, _ := cmd.Flags().GetBool("verbose")
	noColor, _ := cmd.Flags().GetBool("no-color")

	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	// Determine which config to sync
	var targetConfigID string
	if len(args) > 0 {
		// User specified a config
		cfg, err := manager.GetConfig(args[0])
		if err != nil {
			return fmt.Errorf("configuration not found: %w", err)
		}
		targetConfigID = cfg.ID

		// Set as current config for sync
		if err := manager.SetCurrentConfig(targetConfigID); err != nil {
			return fmt.Errorf("failed to set current config: %w", err)
		}
	} else {
		// Use currently active config
		currentID, err := manager.GetCurrentConfig()
		if err != nil {
			return err
		}
		if currentID == "" {
			return fmt.Errorf("no current configuration set. Please specify a config name or apply one first")
		}
		targetConfigID = currentID
	}

	// Get config details for display
	cfg, err := manager.GetConfig(targetConfigID)
	if err != nil {
		return err
	}

	// Get settings path
	settingsPath, err := manager.GetClaudeSettingsPath()
	if err != nil {
		return err
	}

	// Check if settings.json exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return fmt.Errorf("settings.json not found at %s", settingsPath)
	}

	if !quiet {
		fmt.Printf("Syncing configuration: %s\n", cfg.Name)
		fmt.Printf("   ID: %s\n", cfg.ID)
		if cfg.Description != "" {
			fmt.Printf("   Description: %s\n", cfg.Description)
		}
		fmt.Println()
	}

	// Detect changes
	hasChanges, diffResult, err := manager.DetectChanges(settingsPath)
	if err != nil {
		// Check if it's a conflict error
		if conflictErr, ok := err.(*config.ConflictError); ok {
			return handleConflict(conflictErr, manager, cfg, settingsPath, noColor)
		}
		return fmt.Errorf("failed to detect changes: %w", err)
	}

	// No changes detected
	if !hasChanges {
		if !quiet {
			fmt.Println("No changes detected - configurations are in sync")
		}
		return nil
	}

	// Display changes
	if !quiet {
		fmt.Println(diff.FormatDiff(diffResult, !noColor))
		fmt.Println()
	}

	// Dry run mode
	if dryRun {
		fmt.Println("DRY RUN MODE - No changes will be saved")
		return nil
	}

	// Confirmation prompt (unless forced)
	if !force {
		fmt.Print("Save these changes to the stored configuration? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("Sync cancelled")
			return nil
		}
	}

	// Perform sync
	if verbose && !quiet {
		fmt.Println("Saving changes...")
	}

	if err := manager.SyncConfig(settingsPath); err != nil {
		return fmt.Errorf("failed to sync configuration: %w", err)
	}

	// Success
	if !quiet {
		fmt.Println("Configuration synced successfully!")
		fmt.Printf("Changes saved to: %s\n", cfg.FilePath)
		fmt.Printf("Backup created: %s.backup\n", cfg.FilePath)
	}

	return nil
}

// handleConflict handles sync conflicts with user interaction
func handleConflict(conflictErr *config.ConflictError, manager *config.Manager, cfg *config.Config, settingsPath string, noColor bool) error {
	fmt.Println("CONFLICT DETECTED")
	fmt.Println()
	fmt.Println("Both the stored configuration and live settings.json have been modified.")
	fmt.Println()

	fmt.Println("Changes in stored config vs. live settings:")
	fmt.Println(diff.FormatDiff(conflictErr.StoredDiff, !noColor))
	fmt.Println()

	fmt.Println("Choose resolution strategy:")
	fmt.Println("  1. Use live settings (overwrite stored config)")
	fmt.Println("  2. Use stored config (overwrite live settings)")
	fmt.Println("  3. Show both versions and decide manually")
	fmt.Println("  4. Cancel (do nothing)")
	fmt.Println()

	fmt.Print("Enter choice (1-4): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read choice: %w", err)
	}

	choice := strings.TrimSpace(response)

	switch choice {
	case "1":
		// Force sync from live settings
		fmt.Println("Overwriting stored config with live settings...")

		// Bypass conflict check by updating state
		manager.SetCurrentConfig(cfg.ID)

		if err := manager.SyncConfig(settingsPath); err != nil {
			return fmt.Errorf("failed to sync: %w", err)
		}

		fmt.Println("Stored config updated from live settings")
		return nil

	case "2":
		// Apply stored config to settings
		fmt.Println("Overwriting live settings with stored config...")

		if err := manager.ApplyConfig(cfg.ID); err != nil {
			return fmt.Errorf("failed to apply config: %w", err)
		}

		fmt.Println("Live settings updated from stored config")
		return nil

	case "3":
		// Show both versions
		fmt.Println()
		fmt.Println("=== STORED CONFIG ===")
		storedData, _ := os.ReadFile(cfg.FilePath)
		fmt.Println(string(storedData))
		fmt.Println()

		fmt.Println("=== LIVE SETTINGS ===")
		liveData, _ := os.ReadFile(settingsPath)
		fmt.Println(string(liveData))
		fmt.Println()

		fmt.Println("Please resolve the conflict manually and run sync again.")
		return fmt.Errorf("manual conflict resolution required")

	case "4":
		fmt.Println("Sync cancelled")
		return nil

	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}
}
