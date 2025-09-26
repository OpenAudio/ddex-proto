package ddex

import (
	"encoding/xml"
	"testing"

	"github.com/OpenAudio/ddex-proto/gen"
	"github.com/OpenAudio/ddex-proto/testdata"
	"github.com/OpenAudio/ddex-proto/testutil"
	"github.com/stretchr/testify/require"
)

// getUnmarshalerForMessageType uses the auto-generated registry to create unmarshalers
func getUnmarshalerForMessageType(messageType, version string) testutil.UnmarshalerFunc {
	// Use the auto-generated registry from gen/registry.go
	if !gen.IsRegistered(messageType, version) {
		return nil // Return nil - the test will fail when it tries to use this
	}

	return func(data []byte) (interface{}, error) {
		// Use the generated Parse function
		return gen.Parse(data, messageType, version)
	}
}

// getRoundTripValidatorForMessageType returns the appropriate round-trip validator for a message type
func getRoundTripValidatorForMessageType(messageType string) testutil.RoundTripValidator {
	// Use the generated ParseAny function for auto-detection - works for all DDEX message types
	return func(xmlData []byte) ([]byte, error) {
		msg, _, _, err := gen.ParseAny(xmlData)
		if err != nil {
			return nil, err
		}
		return xml.MarshalIndent(msg, "", "  ")
	}
}

// TestDDEX runs all tests grouped by message type and version
func TestDDEX(t *testing.T) {
	discovered, err := testdata.DiscoverMessageTypesAndVersions()
	if err != nil {
		t.Fatalf("Failed to discover message types and versions: %v", err)
	}

	for messageType, versions := range discovered {
		for _, version := range versions {
			t.Run(messageType+"_"+version, func(t *testing.T) {
				// Check if we have any real test files (after filtering out stub/skip)
				testFiles, err := testdata.GenerateTestFileMap(messageType, version)
				if err != nil {
					t.Fatalf("Failed to get test files: %v", err)
				}
				if len(testFiles) == 0 {
					t.Logf("⚠️  WARNING: No real test files found for %s/%s (only stub/skip files present)", messageType, version)
					t.Skip("No real test files available")
				}

				// Get unmarshaler and validator once for all tests
				unmarshaler := getUnmarshalerForMessageType(messageType, version)
				require.NotNil(t, unmarshaler, "Message type %s/%s not registered in auto-generated registry", messageType, version)

				// Run conformance tests
				t.Run("conformance", func(t *testing.T) {
					testutil.RunConformanceTests(t, messageType, version, unmarshaler,
						func(t *testing.T, msg interface{}, filename string) {
							// Basic validation - just ensure we parsed something
							if msg == nil {
								t.Error("Parsed message is nil")
							}
						})
				})

				// Run integrity tests
				t.Run("integrity", func(t *testing.T) {
					validator := getRoundTripValidatorForMessageType(messageType)
					if validator == nil {
						t.Skipf("No round-trip validator available for %s", messageType)
					}
					testutil.RunIntegrityTests(t, messageType, version, validator)
				})
			})
		}
	}
}

// BenchmarkDDEX runs all benchmarks grouped by message type and version
func BenchmarkDDEX(b *testing.B) {
	discovered, err := testdata.DiscoverMessageTypesAndVersions()
	if err != nil {
		b.Fatalf("Failed to discover message types and versions: %v", err)
	}

	for messageType, versions := range discovered {
		for _, version := range versions {
			b.Run(messageType+"_"+version, func(b *testing.B) {
				// Check if we have any real test files (after filtering out stub/skip)
				testFiles, err := testdata.GenerateTestFileMap(messageType, version)
				if err != nil {
					b.Fatalf("Failed to get test files: %v", err)
				}
				if len(testFiles) == 0 {
					b.Logf("⚠️  WARNING: No real test files found for %s/%s (only stub/skip files present)", messageType, version)
					b.Skip("No real test files available")
				}

				// Get unmarshaler once for all benchmarks
				unmarshaler := getUnmarshalerForMessageType(messageType, version)
				require.NotNil(b, unmarshaler, "Message type %s/%s not registered in auto-generated registry", messageType, version)

				// Run parsing benchmarks
				b.Run("parsing", func(b *testing.B) {
					testutil.RunPerformanceTests(b, messageType, version, unmarshaler)
				})

				// Run marshaling benchmarks
				b.Run("marshaling", func(b *testing.B) {
					testutil.RunMarshalingPerformanceTests(b, messageType, version, unmarshaler,
						func(msg interface{}) ([]byte, error) {
							return xml.MarshalIndent(msg, "", "  ")
						})
				})

				// Run round-trip benchmarks
				b.Run("round_trip", func(b *testing.B) {
					testutil.RunRoundTripPerformanceTests(b, messageType, version, unmarshaler,
						func(msg interface{}) ([]byte, error) {
							return xml.MarshalIndent(msg, "", "  ")
						})
				})
			})
		}
	}
}
