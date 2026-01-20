//go:build integration

package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Xanonymous-GitHub/claude-switch/internal/config"
)

// setupTestEnv creates a temporary environment for integration testing
func setupTestEnv(t *testing.T) (string, func()) {
	t.Helper()

	// Create temporary directory
	tmpDir := t.TempDir()

	// Store original HOME
	originalHome := os.Getenv("HOME")

	// Set HOME to temp directory
	os.Setenv("HOME", tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}

	// Create initial settings.json
	settingsPath := filepath.Join(claudeDir, "settings.json")
	initialSettings := `{
  "theme": "dark",
  "fontSize": 14
}`
	if err := os.WriteFile(settingsPath, []byte(initialSettings), 0644); err != nil {
		t.Fatalf("Failed to create settings.json: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.Setenv("HOME", originalHome)
	}

	return tmpDir, cleanup
}

func TestSyncWorkflowIntegration(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a test configuration file
	configContent := `{"theme": "dark", "fontSize": 14}`
	configFile := filepath.Join(tmpDir, "test-config.json")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Add configuration
	cfg, err := manager.AddConfig(configFile, "test-config", "Test configuration")
	if err != nil {
		t.Fatalf("Failed to add config: %v", err)
	}

	t.Logf("Created config: %s (ID: %s)", cfg.Name, cfg.ID)

	// Apply configuration
	if err := manager.ApplyConfig(cfg.ID); err != nil {
		t.Fatalf("Failed to apply config: %v", err)
	}

	// Verify current config is set
	currentID, err := manager.GetCurrentConfig()
	if err != nil {
		t.Fatalf("Failed to get current config: %v", err)
	}
	if currentID != cfg.ID {
		t.Errorf("Current config mismatch: got %s, want %s", currentID, cfg.ID)
	}

	// Modify live settings.json
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	modifiedSettings := `{
  "theme": "light",
  "fontSize": 16,
  "newSetting": true
}`
	if err := os.WriteFile(settingsPath, []byte(modifiedSettings), 0644); err != nil {
		t.Fatalf("Failed to modify settings: %v", err)
	}

	// Detect changes
	hasChanges, diffResult, err := manager.DetectChanges(settingsPath)
	if err != nil {
		t.Fatalf("Failed to detect changes: %v", err)
	}

	if !hasChanges {
		t.Error("Expected changes to be detected")
	}

	if diffResult.AddedLines == 0 && diffResult.DeletedLines == 0 {
		t.Error("Expected non-zero diff statistics")
	}

	t.Logf("Detected changes: +%d -%d", diffResult.AddedLines, diffResult.DeletedLines)

	// Sync changes back
	if err := manager.SyncConfig(settingsPath); err != nil {
		t.Fatalf("Failed to sync config: %v", err)
	}

	// Verify stored config was updated
	storedData, err := os.ReadFile(cfg.FilePath)
	if err != nil {
		t.Fatalf("Failed to read stored config: %v", err)
	}

	// Parse both to compare
	var stored, expected map[string]interface{}
	if err := json.Unmarshal(storedData, &stored); err != nil {
		t.Fatalf("Failed to parse stored config: %v", err)
	}
	if err := json.Unmarshal([]byte(modifiedSettings), &expected); err != nil {
		t.Fatalf("Failed to parse expected config: %v", err)
	}

	// Compare key values
	if stored["theme"] != expected["theme"] {
		t.Errorf("Theme mismatch: got %v, want %v", stored["theme"], expected["theme"])
	}
	if stored["fontSize"] != expected["fontSize"] {
		t.Errorf("FontSize mismatch: got %v, want %v", stored["fontSize"], expected["fontSize"])
	}

	t.Log("Sync workflow integration test passed")
}

func TestNoChangesDetected(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a test configuration file
	configContent := `{"theme": "dark", "fontSize": 14}`
	configFile := filepath.Join(tmpDir, "test-config.json")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Add and apply configuration
	cfg, err := manager.AddConfig(configFile, "no-change-config", "Test no changes")
	if err != nil {
		t.Fatalf("Failed to add config: %v", err)
	}

	if err := manager.ApplyConfig(cfg.ID); err != nil {
		t.Fatalf("Failed to apply config: %v", err)
	}

	// Detect changes (should be none since we just applied)
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	hasChanges, _, err := manager.DetectChanges(settingsPath)
	if err != nil {
		t.Fatalf("Failed to detect changes: %v", err)
	}

	if hasChanges {
		t.Error("Expected no changes to be detected after fresh apply")
	}

	t.Log("No changes detection test passed")
}

func TestConflictDetectionIntegration(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create a test configuration file
	configContent := `{"initial": "value"}`
	configFile := filepath.Join(tmpDir, "conflict-config.json")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Add and apply configuration
	cfg, err := manager.AddConfig(configFile, "conflict-test", "Test conflict detection")
	if err != nil {
		t.Fatalf("Failed to add config: %v", err)
	}

	if err := manager.ApplyConfig(cfg.ID); err != nil {
		t.Fatalf("Failed to apply config: %v", err)
	}

	// Wait a bit to ensure time difference
	time.Sleep(100 * time.Millisecond)

	// Modify stored config directly (simulate external modification)
	externalModification := `{"external": "modification"}`
	if err := os.WriteFile(cfg.FilePath, []byte(externalModification), 0644); err != nil {
		t.Fatalf("Failed to externally modify config: %v", err)
	}

	// Also modify live settings
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	liveModification := `{"live": "modification"}`
	if err := os.WriteFile(settingsPath, []byte(liveModification), 0644); err != nil {
		t.Fatalf("Failed to modify live settings: %v", err)
	}

	// Try to sync - should detect conflict
	err = manager.SyncConfig(settingsPath)
	if err == nil {
		t.Error("Expected conflict error, got nil")
	}

	// Check if it's a conflict error
	if _, ok := err.(*config.ConflictError); !ok {
		t.Errorf("Expected ConflictError, got %T: %v", err, err)
	}

	t.Log("Conflict detection integration test passed")
}

func TestAutoSyncWorkflow(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create two test configurations
	config1Content := `{"config": "one"}`
	config1File := filepath.Join(tmpDir, "config1.json")
	if err := os.WriteFile(config1File, []byte(config1Content), 0644); err != nil {
		t.Fatalf("Failed to create config1 file: %v", err)
	}

	config2Content := `{"config": "two"}`
	config2File := filepath.Join(tmpDir, "config2.json")
	if err := os.WriteFile(config2File, []byte(config2Content), 0644); err != nil {
		t.Fatalf("Failed to create config2 file: %v", err)
	}

	// Add both configurations
	cfg1, err := manager.AddConfig(config1File, "config-one", "First config")
	if err != nil {
		t.Fatalf("Failed to add config1: %v", err)
	}

	cfg2, err := manager.AddConfig(config2File, "config-two", "Second config")
	if err != nil {
		t.Fatalf("Failed to add config2: %v", err)
	}

	// Apply first config
	if err := manager.ApplyConfig(cfg1.ID); err != nil {
		t.Fatalf("Failed to apply config1: %v", err)
	}

	// Modify live settings
	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	modifiedSettings := `{"config": "one", "modified": true}`
	if err := os.WriteFile(settingsPath, []byte(modifiedSettings), 0644); err != nil {
		t.Fatalf("Failed to modify settings: %v", err)
	}

	// Detect changes on config1
	hasChanges, _, err := manager.DetectChanges(settingsPath)
	if err != nil {
		t.Fatalf("Failed to detect changes: %v", err)
	}

	if !hasChanges {
		t.Error("Expected changes to be detected before auto-sync")
	}

	// Sync config1
	if err := manager.SyncConfig(settingsPath); err != nil {
		t.Fatalf("Failed to sync config1: %v", err)
	}

	// Verify config1 was updated
	storedData, err := os.ReadFile(cfg1.FilePath)
	if err != nil {
		t.Fatalf("Failed to read stored config1: %v", err)
	}

	var stored map[string]interface{}
	if err := json.Unmarshal(storedData, &stored); err != nil {
		t.Fatalf("Failed to parse stored config: %v", err)
	}

	if _, ok := stored["modified"]; !ok {
		t.Error("Expected 'modified' field in synced config")
	}

	// Now switch to config2
	if err := manager.ApplyConfig(cfg2.ID); err != nil {
		t.Fatalf("Failed to apply config2: %v", err)
	}

	// Verify current config is now config2
	currentID, _ := manager.GetCurrentConfig()
	if currentID != cfg2.ID {
		t.Errorf("Current config should be config2, got %s", currentID)
	}

	t.Log("Auto-sync workflow integration test passed")
}

func TestMultipleConfigsManagement(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config manager
	manager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create multiple configs
	configs := []struct {
		name    string
		content string
	}{
		{"work", `{"theme": "dark", "purpose": "work"}`},
		{"personal", `{"theme": "light", "purpose": "personal"}`},
		{"testing", `{"theme": "auto", "purpose": "testing"}`},
	}

	var createdConfigs []*config.Config

	for _, c := range configs {
		configFile := filepath.Join(tmpDir, c.name+".json")
		if err := os.WriteFile(configFile, []byte(c.content), 0644); err != nil {
			t.Fatalf("Failed to create %s config file: %v", c.name, err)
		}

		cfg, err := manager.AddConfig(configFile, c.name, c.name+" configuration")
		if err != nil {
			t.Fatalf("Failed to add %s config: %v", c.name, err)
		}
		createdConfigs = append(createdConfigs, cfg)
	}

	// Verify all configs exist
	allConfigs := manager.GetConfigs()
	if len(allConfigs) != len(configs) {
		t.Errorf("Expected %d configs, got %d", len(configs), len(allConfigs))
	}

	// Apply each config and verify
	for _, cfg := range createdConfigs {
		if err := manager.ApplyConfig(cfg.ID); err != nil {
			t.Errorf("Failed to apply %s: %v", cfg.Name, err)
			continue
		}

		currentID, _ := manager.GetCurrentConfig()
		if currentID != cfg.ID {
			t.Errorf("Current config should be %s (ID: %s), got %s", cfg.Name, cfg.ID, currentID)
		}
	}

	t.Log("Multiple configs management integration test passed")
}
