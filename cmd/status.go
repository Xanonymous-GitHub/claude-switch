// cmd/status.go
package cmd

import (
	"fmt"
	"os"

	"github.com/Xanonymous-GitHub/claude-switch/internal/config"
	"github.com/Xanonymous-GitHub/claude-switch/internal/diff"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current configuration and sync status",
	Long: `Display information about the currently active configuration,
including whether there are unsaved changes in live settings.json.

This command shows:
- Currently active configuration
- Last sync time
- Whether there are unsaved changes
- Quick sync/revert options`,
	Example: `  # Show current status
  claude-switch status

  # Show with diff of changes
  claude-switch status --diff`,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().Bool("diff", false, "Show diff of unsaved changes")
	statusCmd.Flags().Bool("no-color", false, "Disable colored output")
}

func runStatus(cmd *cobra.Command, args []string) error {
	showDiff, _ := cmd.Flags().GetBool("diff")
	noColor, _ := cmd.Flags().GetBool("no-color")

	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	// Get current config
	currentID, err := manager.GetCurrentConfig()
	if err != nil {
		return err
	}

	if currentID == "" {
		fmt.Println("ℹ️  No configuration currently active")
		fmt.Println()
		fmt.Println("💡 Use 'claude-switch apply <config>' to activate a configuration")
		return nil
	}

	cfg, err := manager.GetConfig(currentID)
	if err != nil {
		return fmt.Errorf("failed to get current config: %w", err)
	}

	// Get settings path
	settingsPath, err := manager.GetClaudeSettingsPath()
	if err != nil {
		return err
	}

	// Check if settings.json exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		fmt.Println("⚠️  settings.json not found")
		fmt.Println()
		fmt.Printf("💡 Use 'claude-switch apply %s' to create it from your config\n", cfg.Name)
		return nil
	}

	// Display current config info
	fmt.Println("📋 Current Configuration")
	fmt.Printf("   Name: %s\n", cfg.Name)
	fmt.Printf("   ID: %s\n", cfg.ID)
	if cfg.Description != "" {
		fmt.Printf("   Description: %s\n", cfg.Description)
	}
	fmt.Printf("   Created: %s\n", cfg.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Check for changes
	hasChanges, diffResult, err := manager.DetectChanges(settingsPath)
	if err != nil {
		// Check if it's a conflict error
		if conflictErr, ok := err.(*config.ConflictError); ok {
			fmt.Println("⚠️  CONFLICT DETECTED")
			fmt.Println("   Both stored config and live settings were modified")
			fmt.Println()
			fmt.Printf("   Stored config changes: %s\n", conflictErr.StoredDiff.Summary)
			fmt.Println()
			fmt.Println("💡 Quick Actions:")
			fmt.Println("   Resolve conflict: claude-switch sync")
			return nil
		}
		fmt.Printf("⚠️  Warning: Failed to detect changes: %v\n", err)
		return nil
	}

	if !hasChanges {
		fmt.Println("✅ No unsaved changes - configurations are in sync")
		return nil
	}

	// Show change summary
	fmt.Println("📝 Unsaved Changes Detected")
	fmt.Printf("   %s\n", diffResult.Summary)
	fmt.Println()

	// Show diff if requested
	if showDiff {
		fmt.Println(diff.FormatDiff(diffResult, !noColor))
		fmt.Println()
	}

	// Show quick actions
	fmt.Println("💡 Quick Actions:")
	fmt.Printf("   Save changes:    claude-switch sync\n")
	fmt.Printf("   Discard changes: claude-switch apply %s\n", cfg.Name)
	if !showDiff {
		fmt.Printf("   View changes:    claude-switch status --diff\n")
	}

	return nil
}
