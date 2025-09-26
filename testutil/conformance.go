package testutil

import (
	"embed"
	"testing"

	"github.com/OpenAudio/ddex-proto/testdata"
)

// UnmarshalerFunc represents a function that can unmarshal XML data to a message
type UnmarshalerFunc func([]byte) (interface{}, error)

// ValidatorFunc represents a function that validates a parsed message
type ValidatorFunc func(*testing.T, interface{}, string)

// RunConformanceTests runs conformance tests by dynamically scanning for test files
func RunConformanceTests(t *testing.T, messageType, version string, unmarshaler UnmarshalerFunc, validator ValidatorFunc) {
	testFiles, err := testdata.GenerateTestFileMap(messageType, version)
	if err != nil {
		t.Fatalf("Failed to generate test file map: %v", err)
	}

	if len(testFiles) == 0 {
		t.Skipf("No %s %s test files available yet", messageType, version)
	}

	t.Parallel()

	for testName, xmlData := range testFiles {
		t.Run(testName, func(t *testing.T) {
			msg, err := unmarshaler(xmlData)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", testName, err)
			}

			validator(t, msg, testName)
			t.Logf("✓ Successfully parsed %s (%d bytes)", testName, len(xmlData))
		})
	}
}

// RunConformanceTestsFS runs conformance tests using embedded file system (legacy)
func RunConformanceTestsFS(t *testing.T, fsys embed.FS, testFiles TestFileMap, unmarshaler UnmarshalerFunc, validator ValidatorFunc) {
	t.Parallel()

	for testName, filename := range testFiles {
		t.Run(testName, func(t *testing.T) {
			xmlData := LoadTestFileFromFS(t, fsys, filename)

			msg, err := unmarshaler(xmlData)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filename, err)
			}

			validator(t, msg, filename)
			t.Logf("✓ Successfully parsed %s (%d bytes)", filename, len(xmlData))
		})
	}
}

// RunConformanceTestsWithData runs conformance tests using embedded data maps
func RunConformanceTestsWithData(t *testing.T, testData map[string][]byte, unmarshaler UnmarshalerFunc, validator ValidatorFunc) {
	t.Parallel()

	for testName, xmlData := range testData {
		t.Run(testName, func(t *testing.T) {
			msg, err := unmarshaler(xmlData)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", testName, err)
			}

			validator(t, msg, testName)
			t.Logf("✓ Successfully parsed %s (%d bytes)", testName, len(xmlData))
		})
	}
}

// RunFieldCompletenessTests runs tests to verify required fields are populated using embedded FS
func RunFieldCompletenessTests(t *testing.T, fsys embed.FS, testFiles TestFileMap, unmarshaler UnmarshalerFunc, fieldValidator func(*testing.T, interface{}, string)) {
	for testName, filename := range testFiles {
		t.Run(testName, func(t *testing.T) {
			xmlData := LoadTestFileFromFS(t, fsys, filename)

			msg, err := unmarshaler(xmlData)
			if err != nil {
				t.Fatalf("Failed to unmarshal %s: %v", filename, err)
			}

			fieldValidator(t, msg, filename)
		})
	}
}
