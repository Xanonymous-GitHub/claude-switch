package diff

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkComputeJSONDiff(b *testing.B) {
	// Small JSON
	smallOriginal := `{"key": "value"}`
	smallModified := `{"key": "modified"}`

	// Medium JSON
	mediumOriginal := generateTestJSON(100)
	mediumModified := modifyTestJSON(mediumOriginal)

	// Large JSON
	largeOriginal := generateTestJSON(1000)
	largeModified := modifyTestJSON(largeOriginal)

	b.Run("small_different", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ComputeJSONDiff(smallOriginal, smallModified)
		}
	})

	b.Run("small_identical", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ComputeJSONDiff(smallOriginal, smallOriginal)
		}
	})

	b.Run("medium_different", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ComputeJSONDiff(mediumOriginal, mediumModified)
		}
	})

	b.Run("medium_identical", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ComputeJSONDiff(mediumOriginal, mediumOriginal)
		}
	})

	b.Run("large_different", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ComputeJSONDiff(largeOriginal, largeModified)
		}
	})

	b.Run("large_identical", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ComputeJSONDiff(largeOriginal, largeOriginal)
		}
	})
}

func BenchmarkFormatDiff(b *testing.B) {
	// Create a diff result with some changes
	original := `{"key1": "value1", "key2": "value2"}`
	modified := `{"key1": "modified", "key2": "value2", "key3": "new"}`

	result, _ := ComputeJSONDiff(original, modified)

	b.Run("with_color", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FormatDiff(result, true)
		}
	})

	b.Run("without_color", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FormatDiff(result, false)
		}
	})
}

func generateTestJSON(size int) string {
	var sb strings.Builder
	sb.WriteString("{\n")
	for i := range size {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf(`  "key%d": "value%d"`, i, i))
	}
	sb.WriteString("\n}")
	return sb.String()
}

func modifyTestJSON(original string) string {
	// Modify a few values in the JSON
	modified := strings.Replace(original, `"value0"`, `"modified0"`, 1)
	modified = strings.Replace(modified, `"value50"`, `"modified50"`, 1)
	return modified
}
