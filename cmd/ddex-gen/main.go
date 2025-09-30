// ddex-gen generates DDEX-specific Go code extensions for protobuf-generated files.
//
// It generates:
// - enum_strings.go: String conversion methods for enums
// - *.xml.go: XML marshaling methods with namespace support
// - registry.go: Dynamic message type registry
//
// Usage:
//
//	ddex-gen [directory]
//
// If no directory is specified, it defaults to "./gen"
//
// Installation:
//
//	go install github.com/OpenAudio/ddex-proto/cmd/ddex-gen@latest
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OpenAudio/ddex-proto/pkg/ddexgen"
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
		fmt.Printf("ddex-gen version %s\n", version)
		fmt.Println("DDEX code generator for protobuf")
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
		fmt.Fprintf(os.Stderr, "\nUsage: ddex-gen [directory]\n")
		fmt.Fprintf(os.Stderr, "Run 'buf generate' first to generate .pb.go files\n")
		os.Exit(1)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to resolve directory path: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("ddex-gen v%s\n", version)
		fmt.Printf("Processing generated files in: %s\n\n", absDir)
	}

	// Generate DDEX extensions
	if err := ddexgen.Generate(absDir, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Println("\n✓ Generation complete!")
	}
}
