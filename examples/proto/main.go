package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	ernv432 "github.com/OpenAudio/ddex-proto/gen/ddex/ern/v432"
	meadv11 "github.com/OpenAudio/ddex-proto/gen/ddex/mead/v11"
	piev10 "github.com/OpenAudio/ddex-proto/gen/ddex/pie/v10"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	var filePath string
	var outputPath string
	flag.StringVar(&filePath, "file", "", "Path to DDEX XML file")
	flag.StringVar(&outputPath, "output", "", "Optional: Path to output XML file after proto conversion")
	flag.Parse()

	if filePath == "" {
		fmt.Println("Usage: go run main.go -file <path-to-ddex-file> [-output <output-file>]")
		fmt.Println("\nExample:")
		fmt.Println("  go run main.go -file ../../test-files/sample.xml")
		fmt.Println("  go run main.go -file ../../test-files/sample.xml -output output.xml")
		os.Exit(1)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	fileName := filepath.Base(filePath)
	fmt.Printf("Processing: %s\n\n", fileName)

	// Try ERN NewReleaseMessage
	var newRelease ernv432.NewReleaseMessage
	if err := xml.Unmarshal(data, &newRelease); err == nil && newRelease.MessageHeader != nil {
		fmt.Println("✓ Parsed as ERN v4.3.2 NewReleaseMessage (protobuf)")
		spew.Dump(&newRelease)

		if outputPath != "" {
			output, err := xml.MarshalIndent(&newRelease, "", "  ")
			if err != nil {
				log.Printf("Failed to marshal back to XML: %v", err)
			} else {
				output = append([]byte(xml.Header), output...)
				if err := os.WriteFile(outputPath, output, 0644); err != nil {
					log.Printf("Failed to write output file: %v", err)
				} else {
					fmt.Printf("\n✓ Written to %s\n", outputPath)
				}
			}
		}
		return
	}

	// Try ERN PurgeReleaseMessage
	var purgeRelease ernv432.PurgeReleaseMessage
	if err := xml.Unmarshal(data, &purgeRelease); err == nil && purgeRelease.MessageHeader != nil {
		fmt.Println("✓ Parsed as ERN v4.3.2 PurgeReleaseMessage (protobuf)")
		spew.Dump(&purgeRelease)

		if outputPath != "" {
			output, err := xml.MarshalIndent(&purgeRelease, "", "  ")
			if err != nil {
				log.Printf("Failed to marshal back to XML: %v", err)
			} else {
				output = append([]byte(xml.Header), output...)
				if err := os.WriteFile(outputPath, output, 0644); err != nil {
					log.Printf("Failed to write output file: %v", err)
				} else {
					fmt.Printf("\n✓ Written to %s\n", outputPath)
				}
			}
		}
		return
	}

	// Try MEAD
	var mead meadv11.MeadMessage
	if err := xml.Unmarshal(data, &mead); err == nil && mead.MessageHeader != nil {
		fmt.Println("✓ Parsed as MEAD v1.1 MeadMessage (protobuf)")
		spew.Dump(&mead)

		if outputPath != "" {
			output, err := xml.MarshalIndent(&mead, "", "  ")
			if err != nil {
				log.Printf("Failed to marshal back to XML: %v", err)
			} else {
				output = append([]byte(xml.Header), output...)
				if err := os.WriteFile(outputPath, output, 0644); err != nil {
					log.Printf("Failed to write output file: %v", err)
				} else {
					fmt.Printf("\n✓ Written to %s\n", outputPath)
				}
			}
		}
		return
	}

	// Try PIE Message
	var pie piev10.PieMessage
	if err := xml.Unmarshal(data, &pie); err == nil && pie.MessageHeader != nil {
		fmt.Println("✓ Parsed as PIE v1.0 PieMessage (protobuf)")
		spew.Dump(&pie)

		if outputPath != "" {
			output, err := xml.MarshalIndent(&pie, "", "  ")
			if err != nil {
				log.Printf("Failed to marshal back to XML: %v", err)
			} else {
				output = append([]byte(xml.Header), output...)
				if err := os.WriteFile(outputPath, output, 0644); err != nil {
					log.Printf("Failed to write output file: %v", err)
				} else {
					fmt.Printf("\n✓ Written to %s\n", outputPath)
				}
			}
		}
		return
	}

	// Try PIE Request
	var pieRequest piev10.PieRequestMessage
	if err := xml.Unmarshal(data, &pieRequest); err == nil && pieRequest.MessageHeader != nil {
		fmt.Println("✓ Parsed as PIE v1.0 PieRequestMessage (protobuf)")
		spew.Dump(&pieRequest)

		if outputPath != "" {
			output, err := xml.MarshalIndent(&pieRequest, "", "  ")
			if err != nil {
				log.Printf("Failed to marshal back to XML: %v", err)
			} else {
				output = append([]byte(xml.Header), output...)
				if err := os.WriteFile(outputPath, output, 0644); err != nil {
					log.Printf("Failed to write output file: %v", err)
				} else {
					fmt.Printf("\n✓ Written to %s\n", outputPath)
				}
			}
		}
		return
	}

	fmt.Println("❌ Could not parse file as any supported DDEX message type (protobuf)")
	fmt.Println("\nSupported types:")
	fmt.Println("  - ERN v4.3.2 (NewReleaseMessage, PurgeReleaseMessage)")
	fmt.Println("  - MEAD v1.1 (MeadMessage)")
	fmt.Println("  - PIE v1.0 (PieMessage, PieRequestMessage)")
	fmt.Println("\nNote: This example uses protobuf-generated structs to parse XML data.")
	fmt.Println("Compare with examples/xsd/main.go which uses XSD-generated structs.")
}
