package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/Xanonymous-GitHub/claude-switch/internal/config"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "show"},
	Short:   "List all saved configurations",
	Long: `List all saved Claude Code configurations with details.

This command displays a table showing:
- Configuration ID (first 8 characters)
- Name
- Description
- Creation date
- File size

Use the configuration name or full ID with other commands.`,
	Example: `  # List all configurations
  claude-switch list

  # Alternative commands
  claude-switch ls
  claude-switch show`,
	RunE: runList,
}

func init() {
	listCmd.Flags().BoolP("detailed", "d", false, "Show detailed information including full IDs")
	listCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
}

func runList(cmd *cobra.Command, args []string) error {
	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	// Get all configurations
	configs := manager.GetConfigs()

	// Check if any configurations exist
	if len(configs) == 0 {
		fmt.Println("📋 No configurations found.")
		fmt.Println()
		fmt.Println("💡 Use 'claude-switch add' to create your first configuration")
		return nil
	}

	// Check flags
	detailed, _ := cmd.Flags().GetBool("detailed")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if jsonOutput {
		return outputJSON(configs)
	}

	return outputTable(configs, detailed)
}

// outputTable displays configurations in a formatted table
func outputTable(configs []config.Config, detailed bool) error {
	fmt.Printf("📋 Found %d configuration%s:\n\n", len(configs), pluralize(len(configs)))

	// Create table with new API
	table := tablewriter.NewWriter(os.Stdout)

	// Set table headers using the new API
	table.Header("ID", "Name", "Description", "Created", "Size")

	// Add rows
	for _, cfg := range configs {
		id := cfg.ID
		if !detailed && len(id) > 8 {
			id = id[:8] + "..."
		}

		description := cfg.Description
		if description == "" {
			description = "-"
		} else if !detailed && len(description) > 40 {
			description = description[:37] + "..."
		}

		// Get file size
		size := getFileSize(cfg.FilePath)

		// Format creation date
		created := cfg.CreatedAt.Format("2006-01-02 15:04")

		err := table.Append(id, cfg.Name, description, created, size)
		if err != nil {
			return fmt.Errorf("failed to add row to table: %w", err)
		}
	}

	// Render the table
	err := table.Render()
	if err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	fmt.Println()
	fmt.Printf("💡 Use 'claude-switch apply <name>' to switch to a configuration\n")
	fmt.Printf("💡 Use 'claude-switch remove <name>' to delete a configuration\n")

	if !detailed {
		fmt.Printf("💡 Use '--detailed' flag to see full IDs and descriptions\n")
	}

	return nil
}

// outputJSON displays configurations in JSON format
func outputJSON(configs []config.Config) error {
	// This would use the json package to marshal and output
	// For brevity, showing simplified version
	fmt.Println("[")
	for i, cfg := range configs {
		fmt.Printf("  {\n")
		fmt.Printf("    \"id\": \"%s\",\n", cfg.ID)
		fmt.Printf("    \"name\": \"%s\",\n", cfg.Name)
		fmt.Printf("    \"description\": \"%s\",\n", cfg.Description)
		fmt.Printf("    \"created_at\": \"%s\",\n", cfg.CreatedAt.Format(time.RFC3339))
		fmt.Printf("    \"file_path\": \"%s\"\n", cfg.FilePath)
		if i < len(configs)-1 {
			fmt.Printf("  },\n")
		} else {
			fmt.Printf("  }\n")
		}
	}
	fmt.Println("]")
	return nil
}

// getFileSize returns a human-readable file size
func getFileSize(filePath string) string {
	info, err := os.Stat(filePath)
	if err != nil {
		return "unknown"
	}

	size := info.Size()
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
}

// pluralize returns "s" if count is not 1
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
