package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Xanonymous-GitHub/claude-switch/internal/diff"
	"github.com/Xanonymous-GitHub/claude-switch/internal/validation"
	"github.com/google/uuid"
)

// Config represents a single Claude Code configuration
type Config struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	FilePath    string    `json:"file_path"`
}

// StateMetadata tracks current configuration state
type StateMetadata struct {
	CurrentConfigID string    `json:"current_config_id"`
	LastSyncTime    time.Time `json:"last_sync_time"`
}

// ConflictError represents a sync conflict
type ConflictError struct {
	ConfigName string
	StoredDiff *diff.DiffResult
	LiveDiff   *diff.DiffResult
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict detected in config '%s': both stored config and live settings were modified", e.ConfigName)
}

// Manager handles configuration operations
type Manager struct {
	configDir string
	configs   []Config
	state     StateMetadata
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

	if err := manager.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
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

	// Track this as the current configuration
	if err := m.SetCurrentConfig(config.ID); err != nil {
		return fmt.Errorf("failed to track current config: %w", err)
	}

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

// SetCurrentConfig sets the currently active configuration
func (m *Manager) SetCurrentConfig(configID string) error {
	m.state.CurrentConfigID = configID
	m.state.LastSyncTime = time.Now()
	return m.saveState()
}

// GetCurrentConfig returns the currently active configuration ID
func (m *Manager) GetCurrentConfig() (string, error) {
	return m.state.CurrentConfigID, nil
}

// ClearCurrentConfig clears the current configuration tracking
func (m *Manager) ClearCurrentConfig() error {
	m.state.CurrentConfigID = ""
	m.state.LastSyncTime = time.Time{}
	return m.saveState()
}

// loadState loads state metadata from file
func (m *Manager) loadState() error {
	statePath := filepath.Join(m.configDir, "state.json")

	data, err := os.ReadFile(statePath)
	if os.IsNotExist(err) {
		// File doesn't exist, start with empty state
		m.state = StateMetadata{}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read state metadata: %w", err)
	}

	if err := json.Unmarshal(data, &m.state); err != nil {
		return fmt.Errorf("failed to parse state metadata: %w", err)
	}

	return nil
}

// saveState saves state metadata to file
func (m *Manager) saveState() error {
	statePath := filepath.Join(m.configDir, "state.json")

	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state metadata: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state metadata: %w", err)
	}

	return nil
}

// DetectChanges compares live settings.json with stored config
func (m *Manager) DetectChanges(settingsPath string) (bool, *diff.DiffResult, error) {
	// Get current config
	currentID, err := m.GetCurrentConfig()
	if err != nil {
		return false, nil, err
	}

	if currentID == "" {
		return false, nil, fmt.Errorf("no current configuration set")
	}

	// Get the stored config
	config, err := m.GetConfig(currentID)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get current config: %w", err)
	}

	// Read both files
	storedData, err := os.ReadFile(config.FilePath)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read stored config: %w", err)
	}

	liveData, err := os.ReadFile(settingsPath)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read live settings: %w", err)
	}

	// Compute diff
	diffResult, err := diff.ComputeJSONDiff(string(storedData), string(liveData))
	if err != nil {
		return false, nil, fmt.Errorf("failed to compute diff: %w", err)
	}

	return diffResult.HasChanges, diffResult, nil
}

// SyncConfig synchronizes live settings.json back to stored config
func (m *Manager) SyncConfig(settingsPath string) error {
	// Get current config
	currentID, err := m.GetCurrentConfig()
	if err != nil {
		return err
	}

	if currentID == "" {
		return fmt.Errorf("no current configuration set")
	}

	config, err := m.GetConfig(currentID)
	if err != nil {
		return fmt.Errorf("failed to get current config: %w", err)
	}

	// Check for conflicts (stored config modified externally)
	if err := m.checkForConflicts(config, settingsPath); err != nil {
		return err
	}

	// Read live settings
	liveData, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read live settings: %w", err)
	}

	// Validate before saving
	if err := validation.ValidateJSON(liveData); err != nil {
		return fmt.Errorf("live settings contain invalid JSON: %w", err)
	}

	// Create backup of stored config
	backupPath := config.FilePath + ".backup"
	if err := copyFile(config.FilePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Write to stored config
	if err := os.WriteFile(config.FilePath, liveData, 0644); err != nil {
		// Restore backup on failure
		copyFile(backupPath, config.FilePath)
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Update sync time
	if err := m.SetCurrentConfig(currentID); err != nil {
		return fmt.Errorf("failed to update sync time: %w", err)
	}

	return nil
}

// checkForConflicts detects if stored config was modified externally
func (m *Manager) checkForConflicts(config *Config, settingsPath string) error {
	// Get file modification times
	storedInfo, err := os.Stat(config.FilePath)
	if err != nil {
		return fmt.Errorf("failed to stat stored config: %w", err)
	}

	// If stored config was modified after last sync, we have a conflict
	if storedInfo.ModTime().After(m.state.LastSyncTime) {
		// Read both versions
		storedData, _ := os.ReadFile(config.FilePath)
		liveData, _ := os.ReadFile(settingsPath)

		// Compute diffs to show in error
		storedDiff, _ := diff.ComputeJSONDiff(string(storedData), string(liveData))

		return &ConflictError{
			ConfigName: config.Name,
			StoredDiff: storedDiff,
		}
	}

	return nil
}
