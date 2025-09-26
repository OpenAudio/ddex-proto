package testutil

import (
	"testing"

	"github.com/OpenAudio/ddex-proto/testdata"
)

// RunPerformanceTests runs performance benchmarks for parsing by dynamically scanning for test files
func RunPerformanceTests(b *testing.B, messageType, version string, unmarshaler UnmarshalerFunc) {
	testFiles, err := testdata.GenerateTestFileMap(messageType, version)
	if err != nil {
		b.Fatalf("Failed to generate test file map: %v", err)
	}

	if len(testFiles) == 0 {
		b.Skipf("No %s %s test files available yet", messageType, version)
	}

	for testName, xmlData := range testFiles {
		b.Run(testName, func(b *testing.B) {
			if len(xmlData) == 0 {
				b.Skip("Test file not available")
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := unmarshaler(xmlData)
				if err != nil {
					b.Fatalf("Failed to unmarshal: %v", err)
				}
			}

			b.SetBytes(int64(len(xmlData)))
		})
	}
}

// RunMarshalingPerformanceTests runs performance benchmarks for marshaling by dynamically scanning for test files
func RunMarshalingPerformanceTests(b *testing.B, messageType, version string, unmarshaler UnmarshalerFunc, marshaler func(interface{}) ([]byte, error)) {
	testFiles, err := testdata.GenerateTestFileMap(messageType, version)
	if err != nil {
		b.Fatalf("Failed to generate test file map: %v", err)
	}

	if len(testFiles) == 0 {
		b.Skipf("No %s %s test files available yet", messageType, version)
	}

	for testName, xmlData := range testFiles {
		b.Run(testName, func(b *testing.B) {
			if len(xmlData) == 0 {
				b.Skip("Test file not available")
			}

			// Load and parse test data outside the benchmark timer
			msg, err := unmarshaler(xmlData)
			if err != nil {
				b.Fatalf("Failed to unmarshal test data: %v", err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := marshaler(msg)
				if err != nil {
					b.Fatalf("Failed to marshal: %v", err)
				}
			}

			b.SetBytes(int64(len(xmlData)))
		})
	}
}

// RunRoundTripPerformanceTests runs performance benchmarks for full round-trip operations by dynamically scanning for test files
func RunRoundTripPerformanceTests(b *testing.B, messageType, version string, unmarshaler UnmarshalerFunc, marshaler func(interface{}) ([]byte, error)) {
	testFiles, err := testdata.GenerateTestFileMap(messageType, version)
	if err != nil {
		b.Fatalf("Failed to generate test file map: %v", err)
	}

	if len(testFiles) == 0 {
		b.Skipf("No %s %s test files available yet", messageType, version)
	}

	for testName, xmlData := range testFiles {
		b.Run(testName, func(b *testing.B) {
			if len(xmlData) == 0 {
				b.Skip("Test file not available")
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				msg, err := unmarshaler(xmlData)
				if err != nil {
					b.Fatalf("Failed to unmarshal: %v", err)
				}

				_, err = marshaler(msg)
				if err != nil {
					b.Fatalf("Failed to marshal: %v", err)
				}
			}

			b.SetBytes(int64(len(xmlData)))
		})
	}
}
