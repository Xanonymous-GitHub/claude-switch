// internal/config/manager_test.go
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSetCurrentConfig(t *testing.T) {
	// Create temporary config directory
	tmpDir := t.TempDir()

	manager := &Manager{
		configDir: tmpDir,
		configs:   []Config{},
	}

	// Test setting current config
	configID := "test-id-123"
	if err := manager.SetCurrentConfig(configID); err != nil {
		t.Fatalf("SetCurrentConfig failed: %v", err)
	}

	// Verify it was saved
	current, err := manager.GetCurrentConfig()
	if err != nil {
		t.Fatalf("GetCurrentConfig failed: %v", err)
	}

	if current != configID {
		t.Errorf("expected current config %s, got %s", configID, current)
	}
}

func TestGetCurrentConfigNotSet(t *testing.T) {
	tmpDir := t.TempDir()

	manager := &Manager{
		configDir: tmpDir,
		configs:   []Config{},
	}

	current, err := manager.GetCurrentConfig()
	if err != nil {
		t.Fatalf("GetCurrentConfig failed: %v", err)
	}

	if current != "" {
		t.Errorf("expected empty current config, got %s", current)
	}
}

func TestDetectChanges(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test config file
	configID := "test-123"
	configPath := filepath.Join(tmpDir, "configs", configID+".json")
	os.MkdirAll(filepath.Dir(configPath), 0755)

	originalJSON := `{"key": "value1", "setting": "original"}`
	os.WriteFile(configPath, []byte(originalJSON), 0644)

	// Create test settings file
	settingsPath := filepath.Join(tmpDir, "settings.json")
	modifiedJSON := `{"key": "value2", "setting": "modified"}`
	os.WriteFile(settingsPath, []byte(modifiedJSON), 0644)

	manager := &Manager{
		configDir: tmpDir,
		configs: []Config{
			{
				ID:       configID,
				Name:     "test",
				FilePath: configPath,
			},
		},
	}
	manager.SetCurrentConfig(configID)

	// Test detection
	hasChanges, diff, err := manager.DetectChanges(settingsPath)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	if !hasChanges {
		t.Error("expected changes to be detected")
	}

	if diff == nil {
		t.Error("expected diff result")
	}
}

func TestDetectChangesNoCurrentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	manager := &Manager{
		configDir: tmpDir,
		configs:   []Config{},
	}

	hasChanges, diff, err := manager.DetectChanges(settingsPath)
	if err == nil {
		t.Error("expected error when no current config set")
	}
	if hasChanges {
		t.Error("expected no changes when no current config")
	}
	if diff != nil {
		t.Error("expected nil diff when error")
	}
}

func TestSyncConfig(t *testing.T) {
	tmpDir := t.TempDir()

	configID := "test-123"
	configPath := filepath.Join(tmpDir, "configs", configID+".json")
	os.MkdirAll(filepath.Dir(configPath), 0755)

	originalJSON := `{"key": "original"}`
	os.WriteFile(configPath, []byte(originalJSON), 0644)

	settingsPath := filepath.Join(tmpDir, "settings.json")
	modifiedJSON := `{"key": "modified"}`
	os.WriteFile(settingsPath, []byte(modifiedJSON), 0644)

	manager := &Manager{
		configDir: tmpDir,
		configs: []Config{
			{
				ID:       configID,
				Name:     "test",
				FilePath: configPath,
			},
		},
	}
	manager.SetCurrentConfig(configID)

	// Test sync
	if err := manager.SyncConfig(settingsPath); err != nil {
		t.Fatalf("SyncConfig failed: %v", err)
	}

	// Verify config was updated
	data, _ := os.ReadFile(configPath)
	if string(data) != modifiedJSON {
		t.Errorf("config not updated, got: %s", string(data))
	}
}

func TestSyncConfigWithConflict(t *testing.T) {
	tmpDir := t.TempDir()

	configID := "test-123"
	configPath := filepath.Join(tmpDir, "configs", configID+".json")
	os.MkdirAll(filepath.Dir(configPath), 0755)

	// Stored config was modified externally
	storedJSON := `{"key": "stored_modified"}`
	os.WriteFile(configPath, []byte(storedJSON), 0644)

	// Live settings also modified
	settingsPath := filepath.Join(tmpDir, "settings.json")
	liveJSON := `{"key": "live_modified"}`
	os.WriteFile(settingsPath, []byte(liveJSON), 0644)

	manager := &Manager{
		configDir: tmpDir,
		configs: []Config{
			{
				ID:        configID,
				Name:      "test",
				FilePath:  configPath,
				CreatedAt: time.Now().Add(-1 * time.Hour), // Older creation
			},
		},
		state: StateMetadata{
			CurrentConfigID: configID,
			LastSyncTime:    time.Now().Add(-30 * time.Minute), // Recent sync
		},
	}

	// Sync should detect conflict
	err := manager.SyncConfig(settingsPath)
	if err == nil {
		t.Error("expected conflict error")
	}

	// Error should mention conflict
	if !strings.Contains(err.Error(), "conflict") {
		t.Errorf("expected conflict error, got: %v", err)
	}
}
