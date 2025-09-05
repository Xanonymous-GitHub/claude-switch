package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// IsValidJSON checks if a file contains valid JSON
func IsValidJSON(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var js json.RawMessage
	if err := json.Unmarshal(data, &js); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// AtomicWrite writes data to a file atomically by writing to a temporary file first
func AtomicWrite(filePath string, data []byte) error {
	dir := filepath.Dir(filePath)
	if err := EnsureDir(dir); err != nil {
		return err
	}

	tempFile := filePath + ".tmp"

	// Write to temporary file
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomically move temporary file to target
	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile) // Clean up on failure
		return fmt.Errorf("failed to move temporary file: %w", err)
	}

	return nil
}

// SafeCopy copies a file with validation
func SafeCopy(src, dst string) error {
	// Validate source file exists and is readable
	if !FileExists(src) {
		return fmt.Errorf("source file does not exist: %s", src)
	}

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Write to destination atomically
	if err := AtomicWrite(dst, data); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	return info.Size(), nil
}
