package testdata

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
)

// DDEXTestDataFS embeds all DDEX test data files
//go:embed ddex
var DDEXTestDataFS embed.FS

// DiscoverMessageTypesAndVersions automatically discovers all message types and versions
// from the embedded filesystem by walking the directory structure
func DiscoverMessageTypesAndVersions() (map[string][]string, error) {
	discovered := make(map[string][]string)

	err := fs.WalkDir(DDEXTestDataFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip files, we only care about directories
		if !d.IsDir() {
			return nil
		}

		// Parse paths like "ddex/ern/v43" or "ddex/mead/v11"
		parts := strings.Split(path, "/")
		if len(parts) == 3 && parts[0] == "ddex" {
			messageType := parts[1] // e.g., "ern", "mead", "pie"
			version := parts[2]     // e.g., "v43", "v11", "v10"

			discovered[messageType] = append(discovered[messageType], version)
		}

		return nil
	})

	return discovered, err
}

// GenerateTestFileMap creates a map of test files for a given message type and version
func GenerateTestFileMap(messageType, version string) (map[string][]byte, error) {
	testFiles := make(map[string][]byte)
	basePath := filepath.Join("ddex", messageType, version)

	err := fs.WalkDir(DDEXTestDataFS, basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process XML files, but skip stub/skip files
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".xml") {
			fileName := strings.ToLower(d.Name())
			if strings.Contains(fileName, "stub") || strings.Contains(fileName, "skip") {
				return nil // Skip stub and skip files
			}

			data, readErr := DDEXTestDataFS.ReadFile(path)
			if readErr != nil {
				return readErr
			}

			// Use relative path from base as key
			relPath, _ := filepath.Rel(basePath, path)
			testFiles[relPath] = data
		}

		return nil
	})

	return testFiles, err
}

// GetEmbeddedFS returns a sub-filesystem for the specified message type and version
func GetEmbeddedFS(messageType, version string) (fs.FS, error) {
	subPath := filepath.Join("ddex", messageType, version)
	return fs.Sub(DDEXTestDataFS, subPath)
}

// GetRootEmbeddedFS returns the entire embedded filesystem for direct access
func GetRootEmbeddedFS() embed.FS {
	return DDEXTestDataFS
}