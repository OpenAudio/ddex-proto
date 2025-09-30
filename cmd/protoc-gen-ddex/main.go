// protoc-gen-ddex is a post-processor for DDEX protobuf-generated Go code.
//
// It performs three operations on generated .pb.go files:
// 1. Injects XML struct tags for DDEX XML compatibility
// 2. Generates enum string conversion methods (enum_strings.go)
// 3. Generates XML marshaling methods and namespace handling (*.xml.go, registry.go)
//
// Usage:
//
//	protoc-gen-ddex [directory]
//
// If no directory is specified, it defaults to "./gen"
//
// Example:
//
//	buf generate  # Generate .pb.go files from buf.build/openaudio/ddex
//	protoc-gen-ddex  # Post-process to add XML support
//
// Installation:
//
//	go install github.com/OpenAudio/ddex-proto/cmd/protoc-gen-ddex@latest
//
// Future features:
// - DDEX validation rules (e.g., reference resolution in ERN messages)
// - Configurable validation options via flags or config file
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OpenAudio/ddex-proto/pkg/ddexgen"
	"github.com/OpenAudio/ddex-proto/pkg/injecttag"
)

const version = "0.1.0"

func main() {
	// Parse command line flags
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		verbose     = flag.Bool("verbose", false, "Enable verbose logging")
		targetDir   = flag.String("dir", "", "Target directory containing generated .pb.go files (default: ./gen)")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("protoc-gen-ddex version %s\n", version)
		fmt.Println("DDEX protobuf post-processor for XML support")
		os.Exit(0)
	}

	// Determine target directory
	dir := *targetDir
	if dir == "" {
		// Check if directory was provided as positional argument
		if flag.NArg() > 0 {
			dir = flag.Arg(0)
		} else {
			dir = "./gen"
		}
	}

	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory %s does not exist\n", dir)
		fmt.Fprintf(os.Stderr, "\nUsage: protoc-gen-ddex [directory]\n")
		fmt.Fprintf(os.Stderr, "Run 'buf generate' first to generate .pb.go files\n")
		os.Exit(1)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to resolve directory path: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("protoc-gen-ddex v%s\n", version)
	fmt.Printf("Processing generated files in: %s\n\n", absDir)

	// Step 1: Inject XML tags into .pb.go files
	fmt.Println("Step 1: Injecting XML tags into .pb.go files...")
	if err := injectTagsIntoDirectory(absDir, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error injecting tags: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ XML tags injected")

	// Step 2: Generate Go extensions (enum_strings.go, *.xml.go, registry.go)
	fmt.Println("Step 2: Generating Go extensions...")
	if err := ddexgen.Generate(absDir, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating extensions: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Go extensions generated")

	fmt.Println("✓ Post-processing complete!")
	fmt.Println("\nGenerated files:")
	fmt.Println("  - XML struct tags injected into .pb.go files")
	fmt.Println("  - enum_strings.go (enum String() methods)")
	fmt.Println("  - *.xml.go (XML marshaling with namespace support)")
	fmt.Println("  - registry.go (dynamic message type registry)")
}

// injectTagsIntoDirectory injects XML struct tags into all .pb.go files in a directory
func injectTagsIntoDirectory(targetDir string, verbose bool) error {
	var pbFiles []string

	// Find all .pb.go files
	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" && filepath.Base(path) != "main.go" {
			if len(filepath.Base(path)) > 6 && filepath.Base(path)[len(filepath.Base(path))-6:] == ".pb.go" {
				pbFiles = append(pbFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(pbFiles) == 0 {
		return fmt.Errorf("no .pb.go files found in %s - did you run 'buf generate' first?", targetDir)
	}

	// Inject tags into each file
	for _, file := range pbFiles {
		if verbose {
			fmt.Printf("  Processing: %s\n", file)
		}

		// Read the file
		src, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		// Parse and inject tags
		areas, err := injecttag.ParseFile(file, src, nil)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", file, err)
		}

		// If no tags to inject, skip
		if len(areas) == 0 {
			continue
		}

		// Write the modified file back
		if err := injecttag.WriteFile(file, areas, false); err != nil {
			return fmt.Errorf("failed to write %s: %w", file, err)
		}
	}

	if verbose {
		fmt.Printf("  Processed %d files\n", len(pbFiles))
	}

	return nil
}
