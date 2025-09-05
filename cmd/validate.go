package cmd

import (
	"fmt"

	"github.com/Xanonymous-GitHub/claude-switch/internal/config"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [config-name-or-id]",
	Short: "Validate configuration files",
	Long: `Validate that configuration files contain valid JSON and are properly formatted.

This command can validate:
- A specific configuration by name or ID
- All stored configurations (when no argument is provided)

The validation checks for:
- Valid JSON syntax
- Proper structure for Claude Code settings
- File accessibility and readability`,
	Example: `  # Validate a specific configuration
  claude-switch validate my-work-setup

  # Validate all configurations
  claude-switch validate

  # Validate with verbose output
  claude-switch validate --verbose`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

func init() {
	validateCmd.Flags().BoolP("verbose", "v", false, "Show detailed validation information")
	validateCmd.Flags().BoolP("all", "a", false, "Validate all configurations (default when no config specified)")
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	validateAll, _ := cmd.Flags().GetBool("all")

	// If no specific config is provided, validate all
	if len(args) == 0 || validateAll {
		return validateAllConfigs(manager, verbose)
	}

	// Validate specific configuration
	return validateSingleConfig(manager, args[0], verbose)
}

func validateSingleConfig(manager *config.Manager, identifier string, verbose bool) error {
	// Get the configuration
	cfg, err := manager.GetConfig(identifier)
	if err != nil {
		return fmt.Errorf("configuration not found: %w", err)
	}

	fmt.Printf("🔍 Validating configuration: %s\n", cfg.Name)
	if verbose {
		fmt.Printf("   ID: %s\n", cfg.ID)
		fmt.Printf("   File: %s\n", cfg.FilePath)
		if cfg.Description != "" {
			fmt.Printf("   Description: %s\n", cfg.Description)
		}
		fmt.Printf("   Created: %s\n", cfg.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Validate the configuration
	if err := manager.ValidateConfig(identifier); err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		return fmt.Errorf("configuration validation failed")
	}

	fmt.Println("✅ Configuration is valid")
	return nil
}

func validateAllConfigs(manager *config.Manager, verbose bool) error {
	configs := manager.GetConfigs()

	if len(configs) == 0 {
		fmt.Println("📭 No configurations found to validate")
		return nil
	}

	fmt.Printf("🔍 Validating %d configuration(s)...\n\n", len(configs))

	errors := manager.ValidateAllConfigs()

	validCount := len(configs) - len(errors)

	// Show results
	for _, cfg := range configs {
		// Check if this config has an error
		hasError := false
		var errorMsg string
		for _, err := range errors {
			if fmt.Sprintf("config '%s'", cfg.Name) == fmt.Sprintf("config '%s'", cfg.Name) {
				hasError = true
				errorMsg = err.Error()
				break
			}
		}

		if hasError {
			fmt.Printf("❌ %s - %s\n", cfg.Name, errorMsg)
			if verbose {
				fmt.Printf("   ID: %s\n", cfg.ID)
				fmt.Printf("   File: %s\n", cfg.FilePath)
			}
		} else {
			fmt.Printf("✅ %s - Valid\n", cfg.Name)
			if verbose {
				fmt.Printf("   ID: %s\n", cfg.ID)
				fmt.Printf("   File: %s\n", cfg.FilePath)
			}
		}

		if verbose {
			fmt.Println()
		}
	}

	// Summary
	fmt.Printf("\n📊 Validation Summary:\n")
	fmt.Printf("   Valid: %d\n", validCount)
	fmt.Printf("   Invalid: %d\n", len(errors))
	fmt.Printf("   Total: %d\n", len(configs))

	if len(errors) > 0 {
		fmt.Printf("\n⚠️  Found %d invalid configuration(s). Use --verbose for details.\n", len(errors))
		return fmt.Errorf("validation failed for %d configuration(s)", len(errors))
	}

	fmt.Println("\n🎉 All configurations are valid!")
	return nil
}
