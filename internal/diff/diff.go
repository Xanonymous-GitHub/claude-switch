// Package diff provides JSON diff computation and formatting utilities
package diff

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// DiffResult contains the result of a diff operation
type DiffResult struct {
	HasChanges   bool
	Unified      string
	Summary      string
	AddedLines   int
	DeletedLines int
}

// MaxDiffSize is the maximum size (in bytes) for detailed diff computation
// Files larger than this will still be compared but with simplified output
const MaxDiffSize = 512 * 1024 // 512KB

// ComputeJSONDiff compares two JSON strings and returns structured diff
func ComputeJSONDiff(original, modified string) (*DiffResult, error) {
	// Quick check for identical strings (optimization)
	if original == modified {
		return &DiffResult{
			HasChanges: false,
			Summary:    "No changes detected",
		}, nil
	}

	// Size check for very large files
	if len(original) > MaxDiffSize || len(modified) > MaxDiffSize {
		// For very large files, just indicate they're different without detailed diff
		return &DiffResult{
			HasChanges:   true,
			Unified:      "(file too large for detailed diff)",
			Summary:      "Files are different (detailed diff skipped for large files)",
			AddedLines:   0,
			DeletedLines: 0,
		}, nil
	}

	// Normalize JSON by parsing and re-marshaling with consistent formatting
	var origObj, modObj any

	if err := json.Unmarshal([]byte(original), &origObj); err != nil {
		return nil, fmt.Errorf("invalid original JSON: %w", err)
	}

	if err := json.Unmarshal([]byte(modified), &modObj); err != nil {
		return nil, fmt.Errorf("invalid modified JSON: %w", err)
	}

	// Re-marshal with consistent indentation
	origNormalized, _ := json.MarshalIndent(origObj, "", "  ")
	modNormalized, _ := json.MarshalIndent(modObj, "", "  ")

	// Check if identical after normalization
	if string(origNormalized) == string(modNormalized) {
		return &DiffResult{
			HasChanges: false,
			Summary:    "No changes detected",
		}, nil
	}

	// Compute diff
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(origNormalized), string(modNormalized), false)

	// Generate unified diff format
	unified := generateUnifiedDiff(diffs)

	// Count changes
	added, deleted := countChanges(diffs)

	return &DiffResult{
		HasChanges:   true,
		Unified:      unified,
		Summary:      fmt.Sprintf("%d additions, %d deletions", added, deleted),
		AddedLines:   added,
		DeletedLines: deleted,
	}, nil
}

// generateUnifiedDiff converts diffs to unified format
func generateUnifiedDiff(diffs []diffmatchpatch.Diff) string {
	var result strings.Builder

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			lines := strings.Split(diff.Text, "\n")
			for _, line := range lines {
				if line != "" {
					result.WriteString("+ ")
					result.WriteString(line)
					result.WriteString("\n")
				}
			}
		case diffmatchpatch.DiffDelete:
			lines := strings.Split(diff.Text, "\n")
			for _, line := range lines {
				if line != "" {
					result.WriteString("- ")
					result.WriteString(line)
					result.WriteString("\n")
				}
			}
		case diffmatchpatch.DiffEqual:
			// Show context (first and last line only for brevity)
			lines := strings.Split(diff.Text, "\n")
			if len(lines) > 3 {
				result.WriteString("  ")
				result.WriteString(lines[0])
				result.WriteString("\n")
				result.WriteString(fmt.Sprintf("  ... (%d unchanged lines) ...\n", len(lines)-2))
				result.WriteString("  ")
				result.WriteString(lines[len(lines)-1])
				result.WriteString("\n")
			} else {
				for _, line := range lines {
					if line != "" {
						result.WriteString("  ")
						result.WriteString(line)
						result.WriteString("\n")
					}
				}
			}
		}
	}

	return result.String()
}

// countChanges counts added and deleted lines
func countChanges(diffs []diffmatchpatch.Diff) (added, deleted int) {
	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		count := len(lines)
		// If the text ends with a newline, strings.Split will produce
		// a trailing empty string; don't count that as a line.
		if count > 0 && lines[count-1] == "" {
			count--
		}
		if count < 0 {
			count = 0
		}

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			added += count
		case diffmatchpatch.DiffDelete:
			deleted += count
		}
	}
	return
}

// FormatDiff formats diff result for display
func FormatDiff(result *DiffResult, useColor bool) string {
	if !result.HasChanges {
		return "No changes detected"
	}

	var output strings.Builder

	output.WriteString("Changes detected:\n")
	output.WriteString(fmt.Sprintf("   %s\n\n", result.Summary))

	if useColor {
		lines := strings.SplitSeq(result.Unified, "\n")
		for line := range lines {
			if strings.HasPrefix(line, "+ ") {
				output.WriteString(color.GreenString(line))
			} else if strings.HasPrefix(line, "- ") {
				output.WriteString(color.RedString(line))
			} else {
				output.WriteString(line)
			}
			output.WriteString("\n")
		}
	} else {
		output.WriteString(result.Unified)
	}

	return output.String()
}

// Performance note: For files >100KB, consider:
// - Streaming JSON parser
// - Incremental diff computation
// - Diff result caching
