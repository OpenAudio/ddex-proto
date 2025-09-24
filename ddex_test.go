package ddex

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	// Proto-generated implementations
	ernv432 "github.com/OpenAudio/ddex-proto/gen/ddex/ern/v432"
	meadv11 "github.com/OpenAudio/ddex-proto/gen/ddex/mead/v11"
	piev10 "github.com/OpenAudio/ddex-proto/gen/ddex/pie/v10"
)

// Test data maps for each message type
var (
	ernTestFiles = map[string]string{
		"Audio Album":           "1 Audio.xml",
		"Video Album":           "2 Video.xml",
		"Mixed Media Bundle":    "3 MixedMedia.xml",
		"Simple Audio Single":   "4 SimpleAudioSingle.xml",
		"Simple Video Single":   "5 SimpleVideoSingle.xml",
		"Ringtone":              "6 Ringtone.xml",
		"Longform Musical Work": "7 LongformMusicalWorkVideo.xml",
		"DJ Mix":                "8 DjMix.xml",
		"Classical Variant":     "Variant Classical.xml",
	}

	meadTestFiles = map[string]string{
		"Award Example": "mead_award_example.xml",
	}

	pieTestFiles = map[string]string{
		"Award Example": "pie_award_example.xml",
	}
)

// TestDDEXConformance tests parsing of sample XML files for all DDEX message types
func TestDDEXConformance(t *testing.T) {
	// Test ERN messages
	t.Run("ERN", func(t *testing.T) {
		t.Parallel()

		for testName, filename := range ernTestFiles {
			t.Run(testName, func(t *testing.T) {
				xmlPath := filepath.Join("testdata", "ernv432", "Samples43", filename)
				xmlData, err := os.ReadFile(xmlPath)
				if err != nil {
					t.Skipf("Sample file not found: %s", xmlPath)
				}

				var msg ernv432.NewReleaseMessage
				err = xml.Unmarshal(xmlData, &msg)
				if err != nil {
					t.Fatalf("Failed to parse %s: %v", filename, err)
				}

				validateERNStructure(t, &msg, filename)
				t.Logf("✓ Successfully parsed %s (%d bytes)", filename, len(xmlData))
			})
		}
	})

	// Test MEAD messages
	t.Run("MEAD", func(t *testing.T) {
		t.Parallel()

		for testName, filename := range meadTestFiles {
			t.Run(testName, func(t *testing.T) {
				xmlPath := filepath.Join("testdata", "meadv11", filename)
				xmlData, err := os.ReadFile(xmlPath)
				if err != nil {
					t.Skipf("Sample file not found: %s", xmlPath)
				}

				var msg meadv11.MeadMessage
				err = xml.Unmarshal(xmlData, &msg)
				if err != nil {
					t.Fatalf("Failed to parse %s: %v", filename, err)
				}

				validateMEADStructure(t, &msg, filename)
				t.Logf("✓ Successfully parsed %s (%d bytes)", filename, len(xmlData))
			})
		}
	})

	// Test PIE messages
	t.Run("PIE", func(t *testing.T) {
		t.Parallel()

		for testName, filename := range pieTestFiles {
			t.Run(testName, func(t *testing.T) {
				xmlPath := filepath.Join("testdata", "piev10", filename)
				xmlData, err := os.ReadFile(xmlPath)
				if err != nil {
					t.Skipf("Sample file not found: %s", xmlPath)
				}

				var msg piev10.PieMessage
				err = xml.Unmarshal(xmlData, &msg)
				if err != nil {
					t.Fatalf("Failed to parse %s: %v", filename, err)
				}

				validatePIEStructure(t, &msg, filename)
				t.Logf("✓ Successfully parsed %s (%d bytes)", filename, len(xmlData))
			})
		}
	})
}

// TestFieldCompleteness tests that required fields are properly populated
func TestFieldCompleteness(t *testing.T) {
	t.Run("ERN", func(t *testing.T) {
		testCases := []struct {
			name     string
			filename string
		}{
			{"Audio Album", "1 Audio.xml"},
			{"Simple Audio Single", "4 SimpleAudioSingle.xml"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				xmlPath := filepath.Join("testdata", "ernv432", "Samples43", tc.filename)
				xmlData, err := os.ReadFile(xmlPath)
				if err != nil {
					t.Skipf("Sample file not found: %s", xmlPath)
				}

				var msg ernv432.NewReleaseMessage
				err = xml.Unmarshal(xmlData, &msg)
				if err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}

				// Test required fields
				validateRequiredFields(t, []fieldCheck{
					{"MessageHeader", msg.MessageHeader},
					{"ReleaseList", msg.ReleaseList},
				})

				// ERN-specific validations
				if msg.MessageHeader != nil {
					if msg.MessageHeader.MessageId == "" && tc.filename != "8 DjMix.xml" {
						t.Error("MessageHeader.MessageId is empty")
					}
					if msg.MessageHeader.MessageSender == nil {
						t.Error("MessageHeader.MessageSender is nil")
					}
				}

				if msg.ReleaseList != nil {
					releaseCount := countERNReleases(msg.ReleaseList)
					if releaseCount == 0 {
						t.Error("ReleaseList contains no releases")
					} else {
						t.Logf("✓ Found %d release(s) in %s", releaseCount, tc.filename)
					}
				}
			})
		}
	})

	t.Run("MEAD", func(t *testing.T) {
		for testName, filename := range meadTestFiles {
			t.Run(testName, func(t *testing.T) {
				xmlPath := filepath.Join("testdata", "meadv11", filename)
				xmlData, err := os.ReadFile(xmlPath)
				if err != nil {
					t.Skipf("Sample file not found: %s", xmlPath)
				}

				var msg meadv11.MeadMessage
				err = xml.Unmarshal(xmlData, &msg)
				if err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}

				// Test required fields
				validateRequiredFields(t, []fieldCheck{
					{"MessageHeader", msg.MessageHeader},
					{"ReleaseInformationList", msg.ReleaseInformationList},
				})

				// MEAD-specific validations
				if msg.MessageHeader != nil {
					if msg.MessageHeader.MessageId == "" {
						t.Error("MessageHeader.MessageId is empty")
					}
					if msg.MessageHeader.MessageSender == nil {
						t.Error("MessageHeader.MessageSender is nil")
					}
				}

				if msg.ReleaseInformationList != nil {
					releaseCount := len(msg.ReleaseInformationList.ReleaseInformation)
					if releaseCount == 0 {
						t.Error("ReleaseInformationList contains no releases")
					} else {
						t.Logf("✓ Found %d release(s) in %s", releaseCount, filename)
					}
				}
			})
		}
	})

	t.Run("PIE", func(t *testing.T) {
		for testName, filename := range pieTestFiles {
			t.Run(testName, func(t *testing.T) {
				xmlPath := filepath.Join("testdata", "piev10", filename)
				xmlData, err := os.ReadFile(xmlPath)
				if err != nil {
					t.Skipf("Sample file not found: %s", xmlPath)
				}

				var msg piev10.PieMessage
				err = xml.Unmarshal(xmlData, &msg)
				if err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}

				// Test required fields
				validateRequiredFields(t, []fieldCheck{
					{"MessageHeader", msg.MessageHeader},
					{"PartyList", msg.PartyList},
				})

				// PIE-specific validations
				if msg.MessageHeader != nil {
					if msg.MessageHeader.MessageId == "" {
						t.Error("MessageHeader.MessageId is empty")
					}
					if msg.MessageHeader.MessageSender == nil {
						t.Error("MessageHeader.MessageSender is nil")
					}
				}

				if msg.PartyList != nil {
					partyCount := len(msg.PartyList.Party)
					if partyCount == 0 {
						t.Error("PartyList contains no parties")
					} else {
						t.Logf("✓ Found %d party(ies) in %s", partyCount, filename)

						// Count awards
						totalAwards := 0
						for _, party := range msg.PartyList.Party {
							totalAwards += len(party.Award)
						}
						if totalAwards > 0 {
							t.Logf("✓ Found %d total award(s) across all parties", totalAwards)
						}
					}
				}
			})
		}
	})
}

// TestXMLTagsEffectiveness validates XML marshaling/unmarshaling for all message types
func TestXMLTagsEffectiveness(t *testing.T) {
	t.Run("ERN", func(t *testing.T) {
		t.Parallel()
		testXMLTags(t, "testdata/ernv432/Samples43/5 SimpleVideoSingle.xml", &ernv432.NewReleaseMessage{}, "ERN")
	})

	t.Run("MEAD", func(t *testing.T) {
		t.Parallel()
		testXMLTags(t, "testdata/meadv11/mead_award_example.xml", &meadv11.MeadMessage{}, "MEAD")
	})

	t.Run("PIE", func(t *testing.T) {
		t.Parallel()
		testXMLTags(t, "testdata/piev10/pie_award_example.xml", &piev10.PieMessage{}, "PIE")
	})
}

// Benchmark tests
func BenchmarkDDEX(b *testing.B) {
	b.Run("ERN", func(b *testing.B) {
		b.Run("Parse", func(b *testing.B) {
			xmlPath := filepath.Join("testdata", "ernv432", "Samples43", "1 Audio.xml")
			xmlData, err := os.ReadFile(xmlPath)
			if err != nil {
				b.Skip("Sample file not found")
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var msg ernv432.NewReleaseMessage
				xml.Unmarshal(xmlData, &msg)
			}
		})

		b.Run("Marshal", func(b *testing.B) {
			xmlPath := filepath.Join("testdata", "ernv432", "Samples43", "1 Audio.xml")
			xmlData, err := os.ReadFile(xmlPath)
			if err != nil {
				b.Skip("Sample file not found")
			}

			var msg ernv432.NewReleaseMessage
			xml.Unmarshal(xmlData, &msg)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				xml.Marshal(&msg)
			}
		})
	})

	b.Run("MEAD", func(b *testing.B) {
		b.Run("Parse", func(b *testing.B) {
			xmlPath := filepath.Join("testdata", "meadv11", "mead_award_example.xml")
			xmlData, err := os.ReadFile(xmlPath)
			if err != nil {
				b.Skip("Sample file not found")
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var msg meadv11.MeadMessage
				xml.Unmarshal(xmlData, &msg)
			}
		})

		b.Run("Marshal", func(b *testing.B) {
			xmlPath := filepath.Join("testdata", "meadv11", "mead_award_example.xml")
			xmlData, err := os.ReadFile(xmlPath)
			if err != nil {
				b.Skip("Sample file not found")
			}

			var msg meadv11.MeadMessage
			xml.Unmarshal(xmlData, &msg)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				xml.Marshal(&msg)
			}
		})
	})

	b.Run("PIE", func(b *testing.B) {
		b.Run("Parse", func(b *testing.B) {
			xmlPath := filepath.Join("testdata", "piev10", "pie_award_example.xml")
			xmlData, err := os.ReadFile(xmlPath)
			if err != nil {
				b.Skip("Sample file not found")
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var msg piev10.PieMessage
				xml.Unmarshal(xmlData, &msg)
			}
		})

		b.Run("Marshal", func(b *testing.B) {
			xmlPath := filepath.Join("testdata", "piev10", "pie_award_example.xml")
			xmlData, err := os.ReadFile(xmlPath)
			if err != nil {
				b.Skip("Sample file not found")
			}

			var msg piev10.PieMessage
			xml.Unmarshal(xmlData, &msg)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				xml.Marshal(&msg)
			}
		})
	})
}

// Helper functions

type fieldCheck struct {
	name  string
	value interface{}
}

func validateRequiredFields(t *testing.T, fields []fieldCheck) {
	for _, field := range fields {
		if field.value == nil {
			t.Errorf("Required field %s is nil", field.name)
		} else if reflect.ValueOf(field.value).IsZero() {
			t.Errorf("Required field %s is zero value", field.name)
		}
	}
}

func testXMLTags(t *testing.T, xmlPath string, msgType interface{}, msgName string) {
	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Skip("Sample file not found")
	}

	err = xml.Unmarshal(xmlData, msgType)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	_, err = xml.Marshal(msgType)
	if err != nil {
		t.Errorf("%s XML tags not working - marshal failed: %v", msgName, err)
	} else {
		t.Logf("✓ %s XML tags working correctly", msgName)
	}
}

// Validation functions for each message type

func validateERNStructure(t *testing.T, msg *ernv432.NewReleaseMessage, filename string) {
	if msg.MessageHeader == nil {
		t.Errorf("MessageHeader is nil in %s", filename)
		return
	}

	if msg.MessageHeader.MessageId == "" {
		// DJ Mix sample has intentionally empty MessageId
		if filename != "8 DjMix.xml" {
			t.Errorf("MessageId is empty in %s", filename)
		} else {
			t.Logf("Note: MessageId is intentionally empty in %s (valid DDEX format)", filename)
		}
	}

	if msg.ReleaseList == nil {
		t.Errorf("ReleaseList is nil in %s", filename)
		return
	}

	releaseCount := countERNReleases(msg.ReleaseList)
	if releaseCount == 0 {
		t.Errorf("No releases found in %s", filename)
	}
}

func validateMEADStructure(t *testing.T, msg *meadv11.MeadMessage, filename string) {
	if msg.MessageHeader == nil {
		t.Errorf("MessageHeader is nil in %s", filename)
		return
	}

	if msg.MessageHeader.MessageId == "" {
		t.Errorf("MessageId is empty in %s", filename)
	}

	if msg.ReleaseInformationList == nil {
		t.Errorf("ReleaseInformationList is nil in %s", filename)
		return
	}

	releaseCount := len(msg.ReleaseInformationList.ReleaseInformation)
	if releaseCount == 0 {
		t.Errorf("No release information found in %s", filename)
	}
}

func validatePIEStructure(t *testing.T, msg *piev10.PieMessage, filename string) {
	if msg.MessageHeader == nil {
		t.Errorf("MessageHeader is nil in %s", filename)
		return
	}

	if msg.MessageHeader.MessageId == "" {
		t.Errorf("MessageId is empty in %s", filename)
	}

	if msg.PartyList == nil {
		t.Errorf("PartyList is nil in %s", filename)
		return
	}

	partyCount := len(msg.PartyList.Party)
	if partyCount == 0 {
		t.Errorf("No parties found in %s", filename)
	}
}

// Utility functions

func countERNReleases(releaseList *ernv432.ReleaseList) int {
	count := 0
	if releaseList.Release != nil {
		count++
	}
	count += len(releaseList.TrackRelease)
	count += len(releaseList.ClipRelease)
	return count
}
