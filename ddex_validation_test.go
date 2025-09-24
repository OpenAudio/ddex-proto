package ddex

import (
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	meadv11 "github.com/OpenAudio/ddex-proto/gen/ddex/mead/v11"
	piev10 "github.com/OpenAudio/ddex-proto/gen/ddex/pie/v10"
	"github.com/beevik/etree"
)

// DOMComparison holds the comparison results between two XML documents
type DOMComparison struct {
	ElementsOriginal    int
	ElementsMarshaled   int
	AttributesOriginal  int
	AttributesMarshaled int
	MissingElements     []string
	MissingAttributes   []string
	ValueMismatches     []string
	ExtraElements       []string
	MarshaledParseable  bool // Can the marshaled XML be parsed back successfully
	Success             bool
}

// TestXMLRoundTripIntegrity validates that XML → Proto → XML preserves all data
func TestXMLRoundTripIntegrity(t *testing.T) {
	testCases := []struct {
		name    string
		xmlPath string
		msgType string
	}{
		{"ERN Audio Album", "testdata/ernv432/Samples43/1 Audio.xml", "ERN"},
		{"ERN Video", "testdata/ernv432/Samples43/2 Video.xml", "ERN"},
		{"ERN Mixed Media", "testdata/ernv432/Samples43/3 MixedMedia.xml", "ERN"},
		{"ERN Simple Audio", "testdata/ernv432/Samples43/4 SimpleAudioSingle.xml", "ERN"},
		{"ERN Simple Video", "testdata/ernv432/Samples43/5 SimpleVideoSingle.xml", "ERN"},
		{"ERN DJ Mix", "testdata/ernv432/Samples43/8 DjMix.xml", "ERN"},
		// TODO: Re-enable these tests when we have verified DDEX-compliant examples
		// {"MEAD Award", "testdata/meadv11/mead_award_example.xml", "MEAD"},
		// {"PIE Award", "testdata/piev10/pie_award_example.xml", "PIE"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			comparison := performRoundTripValidation(tc.xmlPath, tc.msgType)

			// Report statistics
			t.Logf("Elements: Original=%d, Marshaled=%d",
				comparison.ElementsOriginal, comparison.ElementsMarshaled)
			t.Logf("Attributes: Original=%d, Marshaled=%d",
				comparison.AttributesOriginal, comparison.AttributesMarshaled)

			// Check if marshaled XML can be parsed back
			if !comparison.MarshaledParseable {
				t.Errorf("Marshaled XML cannot be parsed back successfully")
			}

			// Check for issues
			if len(comparison.MissingElements) > 0 {
				t.Errorf("Missing %d elements after round-trip:", len(comparison.MissingElements))
				for i, elem := range comparison.MissingElements {
					if i >= 10 {
						t.Errorf("  ... and %d more", len(comparison.MissingElements)-10)
						break
					}
					t.Errorf("  - %s", elem)
				}
			}

			if len(comparison.MissingAttributes) > 0 {
				t.Errorf("Missing %d attributes after round-trip:", len(comparison.MissingAttributes))
				for i, attr := range comparison.MissingAttributes {
					if i >= 10 {
						t.Errorf("  ... and %d more", len(comparison.MissingAttributes)-10)
						break
					}
					t.Errorf("  - %s", attr)
				}
			}

			if len(comparison.ValueMismatches) > 0 {
				t.Errorf("Found %d value mismatches:", len(comparison.ValueMismatches))
				for i, mismatch := range comparison.ValueMismatches {
					if i >= 10 {
						t.Errorf("  ... and %d more", len(comparison.ValueMismatches)-10)
						break
					}
					t.Errorf("  - %s", mismatch)
				}
			}

			if len(comparison.ExtraElements) > 0 {
				t.Logf("Note: %d extra elements in marshaled output (could be default values)",
					len(comparison.ExtraElements))
			}

			if !comparison.Success {
				t.Fail()
			}
		})
	}
}

// performRoundTripValidation does the actual XML → Proto → XML validation
func performRoundTripValidation(xmlPath string, msgType string) *DOMComparison {
	comparison := &DOMComparison{
		MissingElements:    []string{},
		MissingAttributes:  []string{},
		ValueMismatches:    []string{},
		ExtraElements:      []string{},
		MarshaledParseable: true,
		Success:            true,
	}

	// Read original XML
	originalXML, err := os.ReadFile(xmlPath)
	if err != nil {
		comparison.Success = false
		return comparison
	}

	// Parse original XML to DOM
	originalDoc := etree.NewDocument()
	if err := originalDoc.ReadFromBytes(originalXML); err != nil {
		comparison.Success = false
		return comparison
	}

	// Perform round-trip based on message type
	var marshaledXML []byte

	switch msgType {
	case "ERN":
		// Use the new versioning system to auto-detect and parse
		msg, version, err := ParseERN(originalXML)
		if err != nil {
			fmt.Printf("ParseERN error: %v\n", err)
			comparison.Success = false
			return comparison
		}

		marshaledXML, err = xml.MarshalIndent(msg, "", "  ")
		if err != nil {
			comparison.Success = false
			return comparison
		}

		// Log detected version for debugging
		fmt.Printf("Detected ERN version: %s\n", version)

	case "MEAD":
		var msg meadv11.MeadMessage
		if err := xml.Unmarshal(originalXML, &msg); err != nil {
			comparison.Success = false
			return comparison
		}
		marshaledXML, err = xml.MarshalIndent(&msg, "", "  ")
		if err != nil {
			comparison.Success = false
			return comparison
		}

	case "PIE":
		var msg piev10.PieMessage
		if err := xml.Unmarshal(originalXML, &msg); err != nil {
			comparison.Success = false
			return comparison
		}
		marshaledXML, err = xml.MarshalIndent(&msg, "", "  ")
		if err != nil {
			comparison.Success = false
			return comparison
		}

	default:
		comparison.Success = false
		return comparison
	}

	// Parse marshaled XML to DOM
	marshaledDoc := etree.NewDocument()
	if err := marshaledDoc.ReadFromBytes(marshaledXML); err != nil {
		comparison.Success = false
		return comparison
	}

	// Compare the two DOM trees
	compareDOMTrees(originalDoc.Root(), marshaledDoc.Root(), "", comparison)

	// Test what actually matters: can we parse the marshaled XML back?
	switch msgType {
	case "ERN":
		// Test 1: Can ParseERN detect version from marshaled XML?
		version, err := DetectERNVersion(marshaledXML)
		if err != nil || version == "" {
			comparison.MarshaledParseable = false
			fmt.Printf("Version detection failed on marshaled XML: %v\n", err)
		} else {
			// Test 2: Can we parse the marshaled XML back successfully?
			_, _, err = ParseERN(marshaledXML)
			if err != nil {
				comparison.MarshaledParseable = false
				fmt.Printf("Failed to parse marshaled XML back to proto: %v\n", err)
			}
		}

	case "MEAD":
		var msg2 meadv11.MeadMessage
		if err := xml.Unmarshal(marshaledXML, &msg2); err != nil {
			comparison.MarshaledParseable = false
			fmt.Printf("Failed to unmarshal MEAD XML: %v\n", err)
		}

	case "PIE":
		var msg2 piev10.PieMessage
		if err := xml.Unmarshal(marshaledXML, &msg2); err != nil {
			comparison.MarshaledParseable = false
			fmt.Printf("Failed to unmarshal PIE XML: %v\n", err)
		}
	}

	// Set success based on critical issues
	if len(comparison.MissingElements) > 0 ||
		len(comparison.MissingAttributes) > 0 ||
		len(comparison.ValueMismatches) > 0 ||
		!comparison.MarshaledParseable {
		comparison.Success = false
	}

	return comparison
}

// compareDOMTrees recursively compares two XML DOM trees
func compareDOMTrees(original, marshaled *etree.Element, path string, comp *DOMComparison) {
	if original == nil && marshaled == nil {
		return
	}

	// Build current path
	currentPath := path
	if original != nil {
		currentPath = path + "/" + original.Tag
	} else if marshaled != nil {
		currentPath = path + "/" + marshaled.Tag
	}

	// Check if elements exist in both
	if original == nil {
		comp.ExtraElements = append(comp.ExtraElements, currentPath)
		return
	}
	if marshaled == nil {
		comp.MissingElements = append(comp.MissingElements, currentPath)
		return
	}

	// Count elements
	comp.ElementsOriginal++
	comp.ElementsMarshaled++

	// Compare attributes
	origAttrs := make(map[string]string)
	for _, attr := range original.Attr {
		origAttrs[attr.Key] = attr.Value
		comp.AttributesOriginal++
	}

	marshaledAttrs := make(map[string]string)
	for _, attr := range marshaled.Attr {
		marshaledAttrs[attr.Key] = attr.Value
		comp.AttributesMarshaled++
	}

	// Check for missing attributes (ignore namespace declarations)
	for key, origValue := range origAttrs {
		if strings.HasPrefix(key, "xmlns") {
			continue // Skip namespace declarations
		}

		marshaledValue, exists := marshaledAttrs[key]
		if !exists {
			comp.MissingAttributes = append(comp.MissingAttributes,
				fmt.Sprintf("%s@%s", currentPath, key))
		} else if normalizeValue(origValue) != normalizeValue(marshaledValue) {
			comp.ValueMismatches = append(comp.ValueMismatches,
				fmt.Sprintf("%s@%s: '%s' != '%s'",
					currentPath, key, origValue, marshaledValue))
		}
	}

	// Compare text content (if no child elements)
	if len(original.ChildElements()) == 0 && len(marshaled.ChildElements()) == 0 {
		origText := normalizeValue(original.Text())
		marshaledText := normalizeValue(marshaled.Text())

		if origText != "" && origText != marshaledText {
			comp.ValueMismatches = append(comp.ValueMismatches,
				fmt.Sprintf("%s: '%s' != '%s'", currentPath, origText, marshaledText))
		}
	}

	// Build maps of child elements by tag
	origChildren := groupElementsByTag(original.ChildElements())
	marshaledChildren := groupElementsByTag(marshaled.ChildElements())

	// Compare child elements
	allTags := make(map[string]bool)
	for tag := range origChildren {
		allTags[tag] = true
	}
	for tag := range marshaledChildren {
		allTags[tag] = true
	}

	for tag := range allTags {
		origList := origChildren[tag]
		marshaledList := marshaledChildren[tag]

		// For repeated elements, compare them in order
		maxLen := max(len(origList), len(marshaledList))
		for i := 0; i < maxLen; i++ {
			var origChild, marshaledChild *etree.Element

			if i < len(origList) {
				origChild = origList[i]
			}
			if i < len(marshaledList) {
				marshaledChild = marshaledList[i]
			}

			// If counts don't match, we'll catch it in the recursive call
			if origChild != nil || marshaledChild != nil {
				childPath := currentPath
				if i > 0 {
					childPath = fmt.Sprintf("%s[%d]", currentPath, i+1)
				}
				compareDOMTrees(origChild, marshaledChild, childPath, comp)
			}
		}
	}
}

// groupElementsByTag groups a list of elements by their tag name
func groupElementsByTag(elements []*etree.Element) map[string][]*etree.Element {
	grouped := make(map[string][]*etree.Element)
	for _, elem := range elements {
		grouped[elem.Tag] = append(grouped[elem.Tag], elem)
	}
	return grouped
}

// normalizeValue normalizes string values for comparison
func normalizeValue(s string) string {
	// Trim whitespace
	s = strings.TrimSpace(s)
	// Normalize line endings
	s = strings.ReplaceAll(s, "\r\n", "\n")
	// Collapse multiple spaces
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// TestFieldCoverageReport generates a detailed field coverage report
func TestFieldCoverageReport(t *testing.T) {
	xmlPath := "testdata/ernv432/Samples43/1 Audio.xml"

	// Read and parse original
	originalXML, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Skip("Sample file not found")
	}

	originalDoc := etree.NewDocument()
	if err := originalDoc.ReadFromBytes(originalXML); err != nil {
		t.Fatal("Failed to parse original XML")
	}

	// Get all unique paths in original
	originalPaths := collectAllPaths(originalDoc.Root(), "")
	sort.Strings(originalPaths)

	// Unmarshal and marshal using new versioning system
	msg, version, err := ParseERN(originalXML)
	if err != nil {
		t.Fatal("Failed to unmarshal:", err)
	}

	t.Logf("Detected ERN version: %s", version)

	marshaledXML, err := xml.MarshalIndent(msg, "", "  ")
	if err != nil {
		t.Fatal("Failed to marshal")
	}

	marshaledDoc := etree.NewDocument()
	if err := marshaledDoc.ReadFromBytes(marshaledXML); err != nil {
		t.Fatal("Failed to parse marshaled XML")
	}

	// Get all unique paths in marshaled
	marshaledPaths := collectAllPaths(marshaledDoc.Root(), "")
	marshaledPathMap := make(map[string]bool)
	for _, p := range marshaledPaths {
		marshaledPathMap[p] = true
	}

	// Calculate coverage
	covered := 0
	uncovered := []string{}

	for _, path := range originalPaths {
		if marshaledPathMap[path] {
			covered++
		} else {
			uncovered = append(uncovered, path)
		}
	}

	coverage := float64(covered) / float64(len(originalPaths)) * 100

	t.Logf("Field Coverage Report:")
	t.Logf("  Total paths in original: %d", len(originalPaths))
	t.Logf("  Paths preserved: %d", covered)
	t.Logf("  Coverage: %.1f%%", coverage)

	if len(uncovered) > 0 {
		t.Logf("\nUncovered paths (first 20):")
		for i, path := range uncovered {
			if i >= 20 {
				t.Logf("  ... and %d more", len(uncovered)-20)
				break
			}
			t.Logf("  - %s", path)
		}
	}

	if coverage < 100.0 {
		t.Errorf("Coverage is less than 100%%: %.1f%%", coverage)
	}
}

// collectAllPaths collects all unique element paths in the XML
func collectAllPaths(elem *etree.Element, parentPath string) []string {
	if elem == nil {
		return []string{}
	}

	currentPath := parentPath + "/" + elem.Tag
	paths := []string{currentPath}

	// Add attribute paths
	for _, attr := range elem.Attr {
		if !strings.HasPrefix(attr.Key, "xmlns") {
			paths = append(paths, currentPath+"@"+attr.Key)
		}
	}

	// Recursively collect from children
	for _, child := range elem.ChildElements() {
		paths = append(paths, collectAllPaths(child, currentPath)...)
	}

	return paths
}
