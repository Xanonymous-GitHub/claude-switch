// internal/diff/diff_test.go
package diff

import (
	"fmt"
	"strings"
	"testing"
)

func TestComputeJSONDiff(t *testing.T) {
	tests := []struct {
		name      string
		original  string
		modified  string
		wantDiff  bool
		wantError bool
	}{
		{
			name:      "identical JSON",
			original:  `{"key": "value"}`,
			modified:  `{"key": "value"}`,
			wantDiff:  false,
			wantError: false,
		},
		{
			name:      "modified value",
			original:  `{"key": "value1"}`,
			modified:  `{"key": "value2"}`,
			wantDiff:  true,
			wantError: false,
		},
		{
			name:      "added field",
			original:  `{"key": "value"}`,
			modified:  `{"key": "value", "new": "field"}`,
			wantDiff:  true,
			wantError: false,
		},
		{
			name:      "invalid JSON original",
			original:  `{invalid}`,
			modified:  `{"key": "value"}`,
			wantDiff:  false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ComputeJSONDiff(tt.original, tt.modified)

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantError {
				// When we expect an error, result should be nil
				if result != nil {
					t.Error("expected nil result when error occurs")
				}
				return
			}
			if tt.wantDiff && result.HasChanges == false {
				t.Error("expected diff but got none")
			}
			if !tt.wantDiff && result.HasChanges == true {
				t.Error("expected no diff but got changes")
			}
		})
	}
}

func TestComputeJSONDiffLargeFile(t *testing.T) {
	// Create a large JSON string (over MaxDiffSize)
	var sb strings.Builder
	sb.WriteString("{")
	for i := 0; i < 60000; i++ { // This creates a ~600KB file
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf(`"key%d":"value%d"`, i, i))
	}
	sb.WriteString("}")
	largeJSON := sb.String()

	// Verify it's actually larger than MaxDiffSize
	if len(largeJSON) <= MaxDiffSize {
		t.Fatalf("Test JSON not large enough: %d bytes, need > %d", len(largeJSON), MaxDiffSize)
	}

	// Modify slightly
	modifiedLargeJSON := strings.Replace(largeJSON, `"value0"`, `"modified0"`, 1)

	result, err := ComputeJSONDiff(largeJSON, modifiedLargeJSON)
	if err != nil {
		t.Fatalf("Expected no error for large file, got: %v", err)
	}

	if !result.HasChanges {
		t.Error("Expected HasChanges to be true for different large files")
	}

	if result.Unified != "(file too large for detailed diff)" {
		t.Errorf("Expected large file message, got: %s", result.Unified)
	}
}

func TestFormatDiff(t *testing.T) {
	result := &DiffResult{
		HasChanges: true,
		Unified:    "- old line\n+ new line",
		Summary:    "1 change",
	}

	formatted := FormatDiff(result, false)
	if formatted == "" {
		t.Error("expected formatted output")
	}

	formattedColor := FormatDiff(result, true)
	if formattedColor == "" {
		t.Error("expected colored formatted output")
	}
}
