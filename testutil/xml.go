package testutil

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/OpenAudio/ddex-proto/testdata"
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

// RoundTripValidator represents a function that can perform round-trip validation
type RoundTripValidator func([]byte) ([]byte, error)

// PerformRoundTripValidation performs XML → Proto → XML validation with a custom validator
func PerformRoundTripValidation(xmlPath string, validator RoundTripValidator) *DOMComparison {
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

	// Perform round-trip using the provided validator
	marshaledXML, err := validator(originalXML)
	if err != nil {
		fmt.Printf("Round-trip validation error: %v\n", err)
		comparison.Success = false
		return comparison
	}

	// Parse marshaled XML to DOM
	marshaledDoc := etree.NewDocument()
	if err := marshaledDoc.ReadFromBytes(marshaledXML); err != nil {
		comparison.Success = false
		comparison.MarshaledParseable = false
		return comparison
	}

	// Compare the two DOM trees
	CompareDOMTrees(originalDoc.Root(), marshaledDoc.Root(), "", comparison)

	// Test if we can parse the marshaled XML back (using validator again)
	_, err = validator(marshaledXML)
	if err != nil {
		comparison.MarshaledParseable = false
		fmt.Printf("Failed to parse marshaled XML back: %v\n", err)
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

// PerformRoundTripValidationFromData performs XML → Proto → XML validation with a custom validator using byte data
func PerformRoundTripValidationFromData(originalXML []byte, validator RoundTripValidator) *DOMComparison {
	comparison := &DOMComparison{
		MissingElements:    []string{},
		MissingAttributes:  []string{},
		ValueMismatches:    []string{},
		ExtraElements:      []string{},
		MarshaledParseable: true,
		Success:            true,
	}

	// Parse original XML to DOM
	originalDoc := etree.NewDocument()
	if err := originalDoc.ReadFromBytes(originalXML); err != nil {
		comparison.Success = false
		return comparison
	}

	// Perform round-trip using the provided validator
	marshaledXML, err := validator(originalXML)
	if err != nil {
		fmt.Printf("Round-trip validation error: %v\n", err)
		comparison.Success = false
		return comparison
	}

	// Parse marshaled XML to DOM
	marshaledDoc := etree.NewDocument()
	if err := marshaledDoc.ReadFromBytes(marshaledXML); err != nil {
		comparison.Success = false
		comparison.MarshaledParseable = false
		return comparison
	}

	// Compare the two DOM trees
	CompareDOMTrees(originalDoc.Root(), marshaledDoc.Root(), "", comparison)

	// Test if we can parse the marshaled XML back (using validator again)
	_, err = validator(marshaledXML)
	if err != nil {
		comparison.MarshaledParseable = false
		fmt.Printf("Failed to parse marshaled XML back: %v\n", err)
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

// CompareDOMTrees recursively compares two XML DOM trees
func CompareDOMTrees(original, marshaled *etree.Element, path string, comp *DOMComparison) {
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
		maxLen := Max(len(origList), len(marshaledList))
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
				CompareDOMTrees(origChild, marshaledChild, childPath, comp)
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

// CollectAllPaths collects all unique element paths in the XML
func CollectAllPaths(elem *etree.Element, parentPath string) []string {
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
		paths = append(paths, CollectAllPaths(child, currentPath)...)
	}

	return paths
}

// GenerateFieldCoverageReport generates a detailed field coverage report using a custom validator
func GenerateFieldCoverageReport(t *testing.T, xmlPath string, validator RoundTripValidator) {
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
	originalPaths := CollectAllPaths(originalDoc.Root(), "")
	sort.Strings(originalPaths)

	// Use the validator to perform round-trip
	marshaledXML, err := validator(originalXML)
	if err != nil {
		t.Fatal("Failed to perform round-trip:", err)
	}

	marshaledDoc := etree.NewDocument()
	if err := marshaledDoc.ReadFromBytes(marshaledXML); err != nil {
		t.Fatal("Failed to parse marshaled XML")
	}

	// Get all unique paths in marshaled
	marshaledPaths := CollectAllPaths(marshaledDoc.Root(), "")
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

// RunIntegrityTests runs XML round-trip integrity tests with a custom validator
func RunIntegrityTests(t *testing.T, messageType, version string, validator RoundTripValidator) {
	testFiles, err := testdata.GenerateTestFileMap(messageType, version)
	if err != nil {
		t.Fatalf("Failed to generate test file map: %v", err)
	}

	if len(testFiles) == 0 {
		t.Skipf("No %s %s test files available yet", messageType, version)
	}

	for testName, xmlData := range testFiles {
		t.Run(testName, func(t *testing.T) {
			comparison := PerformRoundTripValidationFromData(xmlData, validator)

			// Report statistics with visual indicators
			elementsGood := comparison.ElementsOriginal == comparison.ElementsMarshaled
			elementsIndicator := "🟢"
			if !elementsGood {
				elementsIndicator = "🔴"
			}
			t.Logf("%s Elements: Original=%d, Marshaled=%d",
				elementsIndicator, comparison.ElementsOriginal, comparison.ElementsMarshaled)

			attributesGood := comparison.AttributesMarshaled >= comparison.AttributesOriginal
			attributesIndicator := "🟢"
			attributesNote := ""
			if !attributesGood {
				attributesIndicator = "🔴"
			} else if comparison.AttributesMarshaled > comparison.AttributesOriginal {
				attributesNote = " (Go adding defaults)"
			}
			t.Logf("%s Attributes: Original=%d, Marshaled=%d%s",
				attributesIndicator, comparison.AttributesOriginal, comparison.AttributesMarshaled, attributesNote)

			// Check if marshaled XML can be parsed back
			if !comparison.MarshaledParseable {
				t.Errorf("🔴 CRITICAL: Marshaled XML cannot be parsed back (likely namespace issue)")
			} else {
				t.Logf("🟢 Round-trip parsing: SUCCESS")
			}

			// Check for issues with indicators
			if len(comparison.MissingElements) > 0 {
				t.Errorf("🔴 Missing %d elements after round-trip:", len(comparison.MissingElements))
				for i, elem := range comparison.MissingElements {
					if i >= 10 {
						t.Errorf("  ... and %d more", len(comparison.MissingElements)-10)
						break
					}
					t.Errorf("  - %s", elem)
				}
			}

			if len(comparison.MissingAttributes) > 0 {
				t.Errorf("🔴 Missing %d attributes after round-trip:", len(comparison.MissingAttributes))
				for i, attr := range comparison.MissingAttributes {
					if i >= 10 {
						t.Errorf("  ... and %d more", len(comparison.MissingAttributes)-10)
						break
					}
					t.Errorf("  - %s", attr)
				}
			}

			if len(comparison.ValueMismatches) > 0 {
				t.Errorf("🔴 Found %d value mismatches:", len(comparison.ValueMismatches))
				for i, mismatch := range comparison.ValueMismatches {
					if i >= 10 {
						t.Errorf("  ... and %d more", len(comparison.ValueMismatches)-10)
						break
					}
					t.Errorf("  - %s", mismatch)
				}
			}

			if len(comparison.ExtraElements) > 0 {
				t.Logf("🟡 Note: %d extra elements in marshaled output (Go adding defaults)",
					len(comparison.ExtraElements))
			}

			if !comparison.Success {
				t.Fail()
			}
		})
	}
}
