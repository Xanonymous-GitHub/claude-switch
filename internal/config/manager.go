package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/username/claude-switch/internal/validation"
)

// Config represents a single Claude Code configuration
type Config struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	FilePath    string    `json:"file_path"`
}

// Manager handles configuration operations
type Manager struct {
	configDir string
	configs   []Config
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".claude-switch")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create configs subdirectory
	configsDir := filepath.Join(configDir, "configs")
	if err := os.MkdirAll(configsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create configs directory: %w", err)
	}

	manager := &Manager{
		configDir: configDir,
	}

	if err := manager.loadConfigs(); err != nil {
		return nil, fmt.Errorf("failed to load configurations: %w", err)
	}

	return manager, nil
}

// GetClaudeDir returns the Claude directory path
func (m *Manager) GetClaudeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".claude"), nil
}

// GetClaudeSettingsPath returns the path to Claude settings.json
func (m *Manager) GetClaudeSettingsPath() (string, error) {
	claudeDir, err := m.GetClaudeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(claudeDir, "settings.json"), nil
}

// AddConfig creates a new configuration from temporary file
func (m *Manager) AddConfig(tempFile, name, description string) (*Config, error) {
	// Validate inputs
	if name == "" {
		return nil, fmt.Errorf("config name cannot be empty")
	}

	// Validate JSON in temporary file before proceeding
	if err := validation.ValidateClaudeSettingsFile(tempFile); err != nil {
		return nil, fmt.Errorf("invalid configuration file: %w", err)
	}

	// Check if name already exists
	for _, config := range m.configs {
		if config.Name == name {
			return nil, fmt.Errorf("config with name '%s' already exists", name)
		}
	}

	// Generate unique ID
	id := uuid.New().String()

	// Create config object
	config := Config{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		FilePath:    filepath.Join(m.configDir, "configs", id+".json"),
	}

	// Copy temp file to permanent location
	if err := copyFile(tempFile, config.FilePath); err != nil {
		return nil, fmt.Errorf("failed to copy config file: %w", err)
	}

	// Add to configs list
	m.configs = append(m.configs, config)

	// Save configs metadata
	if err := m.saveConfigs(); err != nil {
		// Clean up created file on error
		os.Remove(config.FilePath)
		return nil, fmt.Errorf("failed to save config metadata: %w", err)
	}

	return &config, nil
}

// GetConfigs returns all configurations
func (m *Manager) GetConfigs() []Config {
	return m.configs
}

// GetConfig returns a specific configuration by ID or name
func (m *Manager) GetConfig(identifier string) (*Config, error) {
	for _, config := range m.configs {
		if config.ID == identifier || config.Name == identifier {
			return &config, nil
		}
	}
	return nil, fmt.Errorf("config not found: %s", identifier)
}

// ApplyConfig switches to the specified configuration
func (m *Manager) ApplyConfig(identifier string) error {
	config, err := m.GetConfig(identifier)
	if err != nil {
		return err
	}

	// Validate the configuration file before applying
	if err := validation.ValidateClaudeSettingsFile(config.FilePath); err != nil {
		return fmt.Errorf("configuration file is invalid: %w", err)
	}

	settingsPath, err := m.GetClaudeSettingsPath()
	if err != nil {
		return err
	}

	// Create backup if settings.json exists
	backupPath := settingsPath + ".backup"
	if _, err := os.Stat(settingsPath); err == nil {
		if err := copyFile(settingsPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Apply the configuration
	if err := copyFile(config.FilePath, settingsPath); err != nil {
		// Try to restore backup on failure
		if _, statErr := os.Stat(backupPath); statErr == nil {
			copyFile(backupPath, settingsPath)
		}
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	fmt.Printf("Applied configuration '%s' to ~/.claude/settings.json\n", config.Name)
	fmt.Printf("Backup saved as: %s\n", backupPath)

	return nil
}

// RemoveConfig removes a configuration
func (m *Manager) RemoveConfig(identifier string) error {
	config, err := m.GetConfig(identifier)
	if err != nil {
		return err
	}

	// Remove the config file
	if err := os.Remove(config.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	// Remove from configs list
	for i, c := range m.configs {
		if c.ID == config.ID {
			m.configs = append(m.configs[:i], m.configs[i+1:]...)
			break
		}
	}

	// Save updated configs metadata
	if err := m.saveConfigs(); err != nil {
		return fmt.Errorf("failed to update config metadata: %w", err)
	}

	return nil
}

// loadConfigs loads configuration metadata from file
func (m *Manager) loadConfigs() error {
	metadataPath := filepath.Join(m.configDir, "config.json")

	data, err := os.ReadFile(metadataPath)
	if os.IsNotExist(err) {
		// File doesn't exist, start with empty configs
		m.configs = []Config{}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read config metadata: %w", err)
	}

	if err := json.Unmarshal(data, &m.configs); err != nil {
		return fmt.Errorf("failed to parse config metadata: %w", err)
	}

	return nil
}

// saveConfigs saves configuration metadata to file
func (m *Manager) saveConfigs() error {
	metadataPath := filepath.Join(m.configDir, "config.json")

	data, err := json.MarshalIndent(m.configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config metadata: %w", err)
	}

	return nil
}

// ValidateConfig validates a stored configuration file
func (m *Manager) ValidateConfig(identifier string) error {
	config, err := m.GetConfig(identifier)
	if err != nil {
		return err
	}

	return validation.ValidateClaudeSettingsFile(config.FilePath)
}

// ValidateAllConfigs validates all stored configuration files
func (m *Manager) ValidateAllConfigs() []error {
	var errors []error
	for _, config := range m.configs {
		if err := validation.ValidateClaudeSettingsFile(config.FilePath); err != nil {
			errors = append(errors, fmt.Errorf("config '%s' (%s): %w", config.Name, config.ID, err))
		}
	}
	return errors
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(dst, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}
