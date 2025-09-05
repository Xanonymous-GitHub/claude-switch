package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Xanonymous-GitHub/claude-switch/internal/config"
	"github.com/Xanonymous-GitHub/claude-switch/internal/editor"
	"github.com/Xanonymous-GitHub/claude-switch/internal/storage"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new Claude Code configuration",
	Long: `Add a new Claude Code configuration by opening your default editor.

This command will:
1. Copy your current ~/.claude/settings.json (if it exists) to a temporary file
2. Open the file in your default editor ($EDITOR or system default)
3. After editing, prompt for a name and description
4. Save the configuration for future use

The configuration will be stored in ~/.claude-switch/configs/ and can be
applied later using the 'apply' command.`,
	Example: `  # Add a new configuration
  claude-switch add

  # The command will open your editor, then prompt for:
  # - Configuration name
  # - Optional description`,
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "Configuration name (will prompt if not provided)")
	addCmd.Flags().StringP("description", "d", "", "Configuration description")
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Check prerequisites
	if err := checkPrerequisites(); err != nil {
		return err
	}

	// Check if editor is available
	if !editor.IsEditorAvailable() {
		return fmt.Errorf("no editor found. Please set the $EDITOR environment variable or install a default editor")
	}

	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	// Create temporary file for editing
	tempFile, err := createTempConfigFile(manager)
	if err != nil {
		return fmt.Errorf("failed to create temporary config file: %w", err)
	}
	defer os.Remove(tempFile) // Clean up temp file

	// Show instructions
	fmt.Println("🎯 Creating new Claude Code configuration...")
	fmt.Printf("📝 Opening editor for file: %s\n", tempFile)
	fmt.Println("📋 Instructions:")
	fmt.Println("   • Edit the JSON configuration as needed")
	fmt.Println("   • Save and close the editor to continue")
	fmt.Println("   • Press Ctrl+C to cancel")
	fmt.Println()

	// Open editor
	if err := editor.OpenEditor(tempFile); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Validate the edited file
	if err := storage.IsValidJSON(tempFile); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Invalid JSON in edited file: %v\n", err)
		fmt.Print("Do you want to edit again? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) == "y" {
			return runAdd(cmd, args) // Recursively try again
		}
		return fmt.Errorf("configuration creation cancelled due to invalid JSON")
	}

	// Get configuration details
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")

	if name == "" {
		name, err = promptForInput("Enter configuration name: ")
		if err != nil {
			return fmt.Errorf("failed to get configuration name: %w", err)
		}
	}

	if description == "" {
		description, _ = promptForInput("Enter description (optional): ")
	}

	// Validate name
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("configuration name cannot be empty")
	}

	// Add configuration
	cfg, err := manager.AddConfig(tempFile, strings.TrimSpace(name), strings.TrimSpace(description))
	if err != nil {
		return fmt.Errorf("failed to add configuration: %w", err)
	}

	// Success message
	fmt.Println()
	fmt.Printf("✅ Configuration added successfully!\n")
	fmt.Printf("   ID: %s\n", cfg.ID)
	fmt.Printf("   Name: %s\n", cfg.Name)
	if cfg.Description != "" {
		fmt.Printf("   Description: %s\n", cfg.Description)
	}
	fmt.Printf("   Created: %s\n", cfg.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Printf("💡 Use 'claude-switch apply %s' to switch to this configuration\n", cfg.Name)

	return nil
}

// createTempConfigFile creates a temporary file with current settings.json content
func createTempConfigFile(manager *config.Manager) (string, error) {
	// Get current settings path
	settingsPath, err := manager.GetClaudeSettingsPath()
	if err != nil {
		return "", err
	}

	// Create temporary file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "claude-settings-"+fmt.Sprintf("%d", os.Getpid())+".json")

	// If settings.json exists, copy it; otherwise create empty JSON
	if storage.FileExists(settingsPath) {
		if err := storage.SafeCopy(settingsPath, tempFile); err != nil {
			return "", fmt.Errorf("failed to copy current settings: %w", err)
		}
	} else {
		// Create basic JSON structure if no settings exist
		defaultSettings := `{
  "theme": "dark",
  "fontSize": 14,
  "editorSettings": {
    "tabSize": 2,
    "wordWrap": true
  }
}`
		if err := os.WriteFile(tempFile, []byte(defaultSettings), 0644); err != nil {
			return "", fmt.Errorf("failed to create default settings: %w", err)
		}
	}

	return tempFile, nil
}

// promptForInput prompts the user for input
func promptForInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}
