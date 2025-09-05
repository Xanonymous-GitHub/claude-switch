package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/username/claude-switch/internal/config"
)

var removeCmd = &cobra.Command{
	Use:     "remove [config-name-or-id]",
	Aliases: []string{"rm", "delete", "del"},
	Short:   "Remove a saved configuration",
	Long: `Remove a saved Claude Code configuration.

This command will permanently delete the configuration file and
remove it from the configuration list. This action cannot be undone.

The configuration file will be removed from ~/.claude-switch/configs/
and the metadata will be updated.`,
	Example: `  # Remove configuration by name
  claude-switch remove my-old-config

  # Remove configuration by ID  
  claude-switch remove a1b2c3d4

  # Remove without confirmation prompt
  claude-switch remove my-config --force

  # Alternative commands
  claude-switch rm my-config
  claude-switch delete my-config`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().BoolP("force", "f", false, "Remove without confirmation prompt")
	removeCmd.Flags().BoolP("dry-run", "n", false, "Show what would be removed without making changes")
}

func runRemove(cmd *cobra.Command, args []string) error {
	identifier := args[0]

	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	// Get the configuration to be removed
	cfg, err := manager.GetConfig(identifier)
	if err != nil {
		return fmt.Errorf("configuration not found: %w", err)
	}

	// Get flags
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Show configuration details
	fmt.Printf("🗑️  Configuration to remove:\n")
	fmt.Printf("   ID: %s\n", cfg.ID)
	fmt.Printf("   Name: %s\n", cfg.Name)
	if cfg.Description != "" {
		fmt.Printf("   Description: %s\n", cfg.Description)
	}
	fmt.Printf("   Created: %s\n", cfg.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   File: %s\n", cfg.FilePath)

	// Show file size if exists
	if info, err := os.Stat(cfg.FilePath); err == nil {
		fmt.Printf("   Size: %d bytes\n", info.Size())
	}

	fmt.Println()

	// Dry run mode
	if dryRun {
		fmt.Println("🔍 DRY RUN MODE - No changes will be made")
		fmt.Printf("Would remove file: %s\n", cfg.FilePath)
		fmt.Printf("Would remove from configuration list: %s\n", cfg.Name)
		return nil
	}

	// Warning message
	fmt.Printf("⚠️  Warning: This action cannot be undone!\n")
	fmt.Printf("   The configuration file will be permanently deleted.\n")
	fmt.Println()

	// Confirmation prompt (unless forced)
	if !force {
		fmt.Printf("Are you sure you want to remove '%s'? (y/N): ", cfg.Name)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ Operation cancelled")
			return nil
		}
	}

	// Additional confirmation for safety
	if !force {
		fmt.Printf("Type the configuration name to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		confirmation, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		if strings.TrimSpace(confirmation) != cfg.Name {
			fmt.Println("❌ Configuration name did not match. Operation cancelled")
			return nil
		}
	}

	// Remove the configuration
	fmt.Printf("🗑️  Removing configuration '%s'...\n", cfg.Name)

	if err := manager.RemoveConfig(identifier); err != nil {
		return fmt.Errorf("failed to remove configuration: %w", err)
	}

	// Success message
	fmt.Printf("✅ Configuration '%s' removed successfully!\n", cfg.Name)
	fmt.Println()

	// Show remaining configurations count
	remaining := manager.GetConfigs()
	if len(remaining) > 0 {
		fmt.Printf("📋 %d configuration%s remaining\n", len(remaining), pluralize(len(remaining)))
		fmt.Println("💡 Use 'claude-switch list' to see remaining configurations")
	} else {
		fmt.Println("📋 No configurations remaining")
		fmt.Println("💡 Use 'claude-switch add' to create a new configuration")
	}

	return nil
}
