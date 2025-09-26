package testutil

import (
	"embed"
	"fmt"
	"io/fs"
	"reflect"
	"testing"
)

// FieldCheck represents a field validation check
type FieldCheck struct {
	Name  string
	Value interface{}
}

// TestFileMap represents a map of test name to filename
type TestFileMap map[string]string

// LoadTestFileFromFS loads and returns the content of a test file from embedded FS
func LoadTestFileFromFS(t *testing.T, fsys embed.FS, filename string) []byte {
	xmlData, err := fs.ReadFile(fsys, filename)
	if err != nil {
		t.Skipf("Sample file not found: %s", filename)
	}
	return xmlData
}

// GetTestFilePath returns the full path to a test file (for non-embed usage)
func GetTestFilePath(messageType, version, filename string) string {
	return fmt.Sprintf("../testdata/ddex/%s/%s/%s", messageType, version, filename)
}

// ValidateRequiredFields validates that required fields are not nil/empty
func ValidateRequiredFields(t *testing.T, fields []FieldCheck) {
	for _, field := range fields {
		if field.Value == nil {
			t.Errorf("Required field %s is nil", field.Name)
			continue
		}

		// Check for empty strings
		if str, ok := field.Value.(string); ok && str == "" {
			t.Errorf("Required field %s is empty string", field.Name)
		}

		// Check for empty slices using reflection
		rv := reflect.ValueOf(field.Value)
		if rv.Kind() == reflect.Slice && rv.Len() == 0 {
			t.Errorf("Required field %s is empty slice", field.Name)
		}
	}
}

// Helper function for max
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
