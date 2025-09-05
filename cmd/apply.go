package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/username/claude-switch/internal/config"
	"github.com/username/claude-switch/internal/storage"
	"github.com/username/claude-switch/internal/validation"
)

var applyCmd = &cobra.Command{
	Use:   "apply [config-name-or-id]",
	Short: "Apply a configuration to Claude Code",
	Long: `Apply a saved configuration to ~/.claude/settings.json.

This command will:
1. Create a backup of your current ~/.claude/settings.json
2. Replace it with the specified configuration
3. Provide rollback information in case of issues

The backup is saved as ~/.claude/settings.json.backup and can be
restored manually if needed.`,
	Example: `  # Apply configuration by name
  claude-switch apply my-work-setup

  # Apply configuration by ID
  claude-switch apply a1b2c3d4-e5f6-7890-abcd-ef1234567890

  # Apply with confirmation prompt
  claude-switch apply my-config --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: runApply,
}

func init() {
	applyCmd.Flags().BoolP("confirm", "c", false, "Prompt for confirmation before applying")
	applyCmd.Flags().BoolP("force", "f", false, "Force apply without backup confirmation")
	applyCmd.Flags().BoolP("dry-run", "n", false, "Show what would be done without making changes")
}

func runApply(cmd *cobra.Command, args []string) error {
	identifier := args[0]

	// Check prerequisites
	if err := checkPrerequisites(); err != nil {
		return err
	}

	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	// Get the configuration
	cfg, err := manager.GetConfig(identifier)
	if err != nil {
		return fmt.Errorf("configuration not found: %w", err)
	}

	// Get flags
	confirm, _ := cmd.Flags().GetBool("confirm")
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get paths
	settingsPath, err := manager.GetClaudeSettingsPath()
	if err != nil {
		return err
	}

	// Check if settings.json exists
	currentExists := storage.FileExists(settingsPath)

	// Show what will happen
	fmt.Printf("🎯 Applying configuration: %s\n", cfg.Name)
	fmt.Printf("   ID: %s\n", cfg.ID)
	if cfg.Description != "" {
		fmt.Printf("   Description: %s\n", cfg.Description)
	}
	fmt.Printf("   Target: %s\n", settingsPath)

	if currentExists {
		fmt.Printf("   Backup: %s.backup\n", settingsPath)

		// Show current file info
		if info, err := os.Stat(settingsPath); err == nil {
			fmt.Printf("   Current file: %d bytes, modified %s\n",
				info.Size(), info.ModTime().Format("2006-01-02 15:04:05"))
		}
	} else {
		fmt.Printf("   Current: No existing settings.json found\n")
	}

	// Show new file info
	if info, err := os.Stat(cfg.FilePath); err == nil {
		fmt.Printf("   New file: %d bytes, created %s\n",
			info.Size(), cfg.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Println()

	// Dry run mode
	if dryRun {
		fmt.Println("🔍 DRY RUN MODE - No changes will be made")
		if currentExists {
			fmt.Printf("Would create backup: %s.backup\n", settingsPath)
		}
		fmt.Printf("Would copy: %s -> %s\n", cfg.FilePath, settingsPath)
		return nil
	}

	// Confirmation prompt
	if confirm && !force {
		if !currentExists {
			fmt.Print("No existing settings.json found. Continue? (y/N): ")
		} else {
			fmt.Print("This will replace your current Claude Code settings. Continue? (y/N): ")
		}

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			fmt.Println("❌ Operation cancelled")
			return nil
		}
	}

	// Validate the configuration file before applying
	if err := validation.ValidateClaudeSettingsFile(cfg.FilePath); err != nil {
		return fmt.Errorf("configuration file is invalid: %w", err)
	}

	// Apply the configuration
	fmt.Println("🔄 Applying configuration...")

	if err := manager.ApplyConfig(identifier); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	// Success message
	fmt.Println("✅ Configuration applied successfully!")
	fmt.Println()

	if currentExists {
		fmt.Printf("💾 Backup saved: %s.backup\n", settingsPath)
		fmt.Println("💡 To rollback: mv ~/.claude/settings.json.backup ~/.claude/settings.json")
	}

	fmt.Println("🔄 Restart Claude Code to see the changes")

	return nil
}
