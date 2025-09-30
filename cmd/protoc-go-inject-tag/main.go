package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/OpenAudio/ddex-proto/pkg/injecttag"
)

func main() {
	var inputFiles, xxxTags string
	var removeTagComment bool
	flag.StringVar(&inputFiles, "input", "", "pattern to match input file(s)")
	flag.StringVar(&xxxTags, "XXX_skip", "", "tags that should be skipped (applies 'tag:\"-\"') for unknown fields (deprecated since protoc-gen-go v1.4.0)")
	flag.BoolVar(&removeTagComment, "remove_tag_comment", false, "removes tag comments from the generated file(s)")
	flag.BoolVar(&injecttag.Verbose, "verbose", false, "verbose logging")

	flag.Parse()

	var xxxSkipSlice []string
	if len(xxxTags) > 0 {
		injecttag.Logf("warn: deprecated flag '-XXX_skip' used")
		xxxSkipSlice = strings.Split(xxxTags, ",")
	}

	if inputFiles == "" {
		log.Fatal("input file is mandatory, see: -help")
	}

	// Handle ** recursive glob pattern by walking directories
	var globResults []string
	if strings.Contains(inputFiles, "**") {
		// Recursive glob: split into base path and pattern
		parts := strings.Split(inputFiles, "**")
		basePath := parts[0]
		if basePath == "" {
			basePath = "."
		}

		// Walk the directory tree
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".pb.go") {
				globResults = append(globResults, path)
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Standard glob
		var err error
		globResults, err = filepath.Glob(inputFiles)
		if err != nil {
			log.Fatal(err)
		}
	}

	var matched int
	for _, path := range globResults {
		finfo, err := os.Stat(path)
		if err != nil {
			log.Fatal(err)
		}

		if finfo.IsDir() {
			continue
		}

		// It should end with ".go" at a minimum.
		if !strings.HasSuffix(strings.ToLower(finfo.Name()), ".go") {
			continue
		}

		matched++

		areas, err := injecttag.ParseFile(path, nil, xxxSkipSlice)
		if err != nil {
			log.Fatal(err)
		}
		if err = injecttag.WriteFile(path, areas, removeTagComment); err != nil {
			log.Fatal(err)
		}
	}

	if matched == 0 {
		log.Fatalf("input %q matched no files, see: -help", inputFiles)
	}
}
