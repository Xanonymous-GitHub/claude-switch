package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "claude-switch",
	Short: "A CLI tool to manage Claude Code settings configurations",
	Long: `claude-switch is a command-line utility that allows you to manage multiple
Claude Code settings.json configurations and switch between them easily.

Features:
  - Add new configurations using your preferred editor (supports Neovim)
  - List all saved configurations
  - Apply any configuration to ~/.claude/settings.json with JSON validation
  - Sync live settings changes back to stored configurations
  - Check sync status and detect unsaved changes
  - Remove configurations you no longer need
  - Validate configuration files for proper JSON formatting
  - Safe backup and restore mechanisms with conflict detection`,
	Version: "1.1.0",
	Example: `  # Add a new configuration
  claude-switch add

  # List all configurations
  claude-switch list

  # Apply a specific configuration
  claude-switch apply my-config

  # Sync changes back to stored config
  claude-switch sync

  # Check current status and unsaved changes
  claude-switch status

  # Validate configurations
  claude-switch validate

  # Remove a configuration
  claude-switch remove old-config`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be added here
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(statusCmd)
}

// checkPrerequisites validates the environment before running commands
func checkPrerequisites() error {
	// Check if ~/.claude directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	claudeDir := homeDir + "/.claude"
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return fmt.Errorf("claude Code directory not found at %s. Please install Claude Code first", claudeDir)
	}

	return nil
}
