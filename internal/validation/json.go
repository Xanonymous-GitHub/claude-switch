package validation

import (
	"encoding/json"
	"fmt"
	"os"
)

// ValidateJSONFile validates that a file contains valid JSON
func ValidateJSONFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return ValidateJSON(data)
}

// ValidateJSON validates that the provided data is valid JSON
func ValidateJSON(data []byte) error {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}
	return nil
}

// ValidateClaudeSettings validates that the JSON contains valid Claude Code settings
func ValidateClaudeSettings(data []byte) error {
	// First validate it's valid JSON
	if err := ValidateJSON(data); err != nil {
		return err
	}

	// Parse as a generic map to check structure
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse JSON as object: %w", err)
	}

	// Basic validation - should be a JSON object (map)
	if settings == nil {
		return fmt.Errorf("settings file must contain a JSON object, not null")
	}

	// Optional: Add more specific validation for Claude Code settings
	// This can be expanded based on known Claude Code configuration structure
	return nil
}

// ValidateClaudeSettingsFile validates a Claude Code settings file
func ValidateClaudeSettingsFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return ValidateClaudeSettings(data)
}

// IsValidJSON checks if data is valid JSON without returning detailed errors
func IsValidJSON(data []byte) bool {
	return ValidateJSON(data) == nil
}

// IsValidJSONFile checks if a file contains valid JSON without returning detailed errors
func IsValidJSONFile(filePath string) bool {
	return ValidateJSONFile(filePath) == nil
}
