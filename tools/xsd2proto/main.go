package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//
// =======================
// Spec list (entry schemas)
// =======================
//

var specs = []struct {
	name     string
	version  string
	mainFile string
}{
	// Process AVS versions first so they're available for imports
	{"avs", "latest", "allowed-value-sets.xsd"},
	{"avs", "20200108", "avs_20200108.xsd"},
	{"avs", "20200518", "avs20200518.xsd"},
	// Then process the main specs
	{"ern", "42", "release-notification.xsd"},
	{"ern", "43", "release-notification.xsd"},
	{"ern", "432", "release-notification.xsd"},
	{"mead", "11", "media-enrichment-and-description.xsd"},
	{"pie", "10", "party-identification-and-enrichment.xsd"},
	{"ern", "383", "release-notification.xsd"},
}

//
// =======================
// XSD Models (extended)
// =======================
//

type XSDSchema struct {
	XMLName         xml.Name         `xml:"schema"`
	TargetNamespace string           `xml:"targetNamespace,attr"`
	Elements        []XSDElement     `xml:"element"`
	ComplexTypes    []XSDComplexType `xml:"complexType"`
	SimpleTypes     []XSDSimpleType  `xml:"simpleType"`

	// NEW: follow schema structure
	Imports  []XSDImport  `xml:"import"`
	Includes []XSDInclude `xml:"include"`
}

type XSDImport struct {
	Namespace      string `xml:"namespace,attr"`
	SchemaLocation string `xml:"schemaLocation,attr"`
}

type XSDInclude struct {
	SchemaLocation string `xml:"schemaLocation,attr"`
}

type XSDElement struct {
	Name        string          `xml:"name,attr"`
	Type        string          `xml:"type,attr"`
	MinOccurs   string          `xml:"minOccurs,attr"`
	MaxOccurs   string          `xml:"maxOccurs,attr"`
	ComplexType *XSDComplexType `xml:"complexType"`
}

type XSDComplexType struct {
	Name          string            `xml:"name,attr"`
	Sequence      *XSDSequence      `xml:"sequence"`
	Choice        *XSDChoice        `xml:"choice"`
	SimpleContent *XSDSimpleContent `xml:"simpleContent"`
	Attributes    []XSDAttribute    `xml:"attribute"`
}

type XSDSequence struct {
	Elements []XSDElement `xml:"element"`
	Choices  []XSDChoice  `xml:"choice"`
}

type XSDChoice struct {
	MinOccurs string        `xml:"minOccurs,attr"`
	MaxOccurs string        `xml:"maxOccurs,attr"`
	Elements  []XSDElement  `xml:"element"`
	Sequences []XSDSequence `xml:"sequence"`
}

type XSDSimpleContent struct {
	Extension *XSDExtension `xml:"extension"`
}

type XSDExtension struct {
	Base       string         `xml:"base,attr"`
	Attributes []XSDAttribute `xml:"attribute"`
}

type XSDAttribute struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	Use  string `xml:"use,attr"`
}

type XSDSimpleType struct {
	Name        string          `xml:"name,attr"`
	Restriction *XSDRestriction `xml:"restriction"`
}

type XSDRestriction struct {
	Base         string           `xml:"base,attr"`
	Enumerations []XSDEnumeration `xml:"enumeration"`
}

type XSDEnumeration struct {
	Value string `xml:"value,attr"`
}

//
// =======================
// Aggregation by namespace
// =======================
//

type NamespaceBundle struct {
	TargetNamespace string

	// Aggregated components (includes merged here).
	Elements     []XSDElement
	ComplexTypes []XSDComplexType
	SimpleTypes  []XSDSimpleType

	// Cross-namespace dependencies discovered via xs:import
	// and via seeing qname types with foreign prefixes (best-effort).
	Imports map[string]struct{} // set of targetNamespace strings

	// Track AVS version context for this namespace
	AVSVersion string // e.g. "20200108" or "" for current
}

// Graph loader state
type loadState struct {
	visitedFiles map[string]struct{} // absolute paths visited
	// Map of targetNamespace → bundle
	nsBundles map[string]*NamespaceBundle
	// file path → schema's ns (helpful for relative includes)
	fileToNS map[string]string
	// Track AVS version context per namespace
	avsVersionContext map[string]string // ns -> avs version
}

func newLoadState() *loadState {
	return &loadState{
		visitedFiles:      make(map[string]struct{}),
		nsBundles:         make(map[string]*NamespaceBundle),
		fileToNS:          make(map[string]string),
		avsVersionContext: make(map[string]string),
	}
}

//
// =======================
// Entry
// =======================
//

// 1) Put this near the top-level (package scope), not inside convertSpec:

type protoPkgInfo struct {
	pkgName   string
	goPackage string
	filePath  string // relative to proto root
}

func main() {
	for _, spec := range specs {
		log.Printf("Converting %s v%s to protobuf (namespace-aware)...", spec.name, spec.version)

		if err := validateSchemas(spec); err != nil {
			log.Fatalf("Schema validation failed for %s v%s: %v", spec.name, spec.version, err)
		}

		if err := convertSpec(spec); err != nil {
			log.Fatalf("Failed to convert %s v%s: %v", spec.name, spec.version, err)
		}
	}
}

func validateSchemas(spec struct{ name, version, mainFile string }) error {
	var entry string

	// Handle AVS specs differently - they're in xsd/ root
	if spec.name == "avs" {
		entry = filepath.Join("xsd", spec.mainFile)
	} else {
		schemasDir := filepath.Join("xsd", spec.name+"v"+spec.version)
		if _, err := os.Stat(schemasDir); os.IsNotExist(err) {
			return fmt.Errorf("schema directory %s does not exist", schemasDir)
		}
		entry = filepath.Join(schemasDir, spec.mainFile)
	}

	if _, err := os.Stat(entry); os.IsNotExist(err) {
		if spec.name != "avs" {
			alt := filepath.Join(filepath.Dir(entry), strings.ReplaceAll(spec.mainFile, "-", "_"))
			if _, err2 := os.Stat(alt); os.IsNotExist(err2) {
				return fmt.Errorf("main schema not found; tried %s and %s", entry, alt)
			}
		} else {
			return fmt.Errorf("AVS schema not found: %s", entry)
		}
	}
	return nil
}

//
// =======================
// Conversion pipeline
// =======================
//

func convertSpec(spec struct{ name, version, mainFile string }) error {
	var entryPath string

	// Handle AVS specs differently - they're in xsd/ root
	if spec.name == "avs" {
		entryPath = filepath.Join("xsd", spec.mainFile)
	} else {
		schemasDir := filepath.Join("xsd", spec.name+"v"+spec.version)
		entryPath = filepath.Join(schemasDir, spec.mainFile)
		if _, err := os.Stat(entryPath); os.IsNotExist(err) {
			entryPath = filepath.Join(schemasDir, strings.ReplaceAll(spec.mainFile, "-", "_"))
		}
	}

	st := newLoadState()
	if err := loadSchemaGraph(st, entryPath); err != nil {
		return fmt.Errorf("load graph: %w", err)
	}

	// Create output dir: proto/<spec or inferred>/*
	outRoot := filepath.Join("proto")
	if err := os.MkdirAll(outRoot, 0755); err != nil {
		return err
	}

	// Emit one .proto per namespace bundle.
	// We need deterministic order for stable builds.
	var namespaces []string
	for ns := range st.nsBundles {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	// Pre-compute package + file paths for imports
	pkgs := make(map[string]protoPkgInfo) // ns → info
	for _, ns := range namespaces {
		bundle := st.nsBundles[ns]
		pkg := namespaceToProtoPackage(ns, bundle, spec)
		goPkg := namespaceToGoPackage(ns, bundle, spec)
		path := packageToPath(pkg)
		pkgs[ns] = protoPkgInfo{pkgName: pkg, goPackage: goPkg, filePath: path}
	}

	for _, ns := range namespaces {
		b := st.nsBundles[ns]
		info := pkgs[ns]

		// Ensure directory exists
		dir := filepath.Join(outRoot, filepath.Dir(info.filePath))
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		// Build file content
		content, err := generateProtoForBundle(b, info.pkgName, info.goPackage, pkgs, st.avsVersionContext)
		if err != nil {
			return fmt.Errorf("generate for ns %s: %w", ns, err)
		}

		outFile := filepath.Join(outRoot, info.filePath)
		if err := os.WriteFile(outFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("write %s: %w", outFile, err)
		}
		log.Printf("Generated %s", outFile)
	}

	return nil
}

//
// =======================
// Graph loader (includes/imports)
// =======================
//

func loadSchemaGraph(st *loadState, filePath string) error {
	abs, _ := filepath.Abs(filePath)
	if _, ok := st.visitedFiles[abs]; ok {
		return nil
	}
	st.visitedFiles[abs] = struct{}{}

	data, err := os.ReadFile(abs)
	if err != nil {
		return fmt.Errorf("read %s: %w", abs, err)
	}

	var schema XSDSchema
	if err := xml.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("parse %s: %w", abs, err)
	}

	if schema.TargetNamespace == "" {
		return fmt.Errorf("schema %s missing targetNamespace", abs)
	}

	st.fileToNS[abs] = schema.TargetNamespace

	// Get or create namespace bundle
	b := st.nsBundles[schema.TargetNamespace]
	if b == nil {
		b = &NamespaceBundle{
			TargetNamespace: schema.TargetNamespace,
			Imports:         make(map[string]struct{}),
			AVSVersion:      "", // will be set when we process imports
		}
		st.nsBundles[schema.TargetNamespace] = b
	}

	// Merge components (includes naturally collapse here)
	b.Elements = append(b.Elements, schema.Elements...)
	b.ComplexTypes = append(b.ComplexTypes, schema.ComplexTypes...)
	b.SimpleTypes = append(b.SimpleTypes, schema.SimpleTypes...)

	// Track declared imports by namespace and detect AVS version context
	for _, imp := range schema.Imports {
		if imp.Namespace != "" && imp.Namespace != schema.TargetNamespace {
			b.Imports[imp.Namespace] = struct{}{}

			// Detect which AVS version this schema imports
			if imp.Namespace == "http://ddex.net/xml/avs/avs" || imp.Namespace == "http://ddex.net/xml/allowed-value-sets" {
				avsVersion := detectAVSVersion(imp.SchemaLocation)
				st.avsVersionContext[schema.TargetNamespace] = avsVersion
			}
		}
	}

	// Follow xs:include (same-namespace; relative to this file)
	baseDir := filepath.Dir(abs)
	for _, inc := range schema.Includes {
		if inc.SchemaLocation == "" {
			continue
		}
		next := filepath.Join(baseDir, inc.SchemaLocation)
		if err := loadSchemaGraph(st, next); err != nil {
			return err
		}
	}

	// Follow xs:import where schemaLocation is present; if absent, we still recorded the ns in Imports
	for _, imp := range schema.Imports {
		if imp.SchemaLocation == "" {
			continue
		}

		// Skip processing AVS imports since we handle them explicitly as separate specs
		if imp.Namespace == "http://ddex.net/xml/avs/avs" || imp.Namespace == "http://ddex.net/xml/allowed-value-sets" {
			continue
		}

		next := filepath.Join(baseDir, imp.SchemaLocation)
		// If the imported file has a different targetNamespace, it will get its own bundle.
		if err := loadSchemaGraph(st, next); err != nil {
			return err
		}
	}

	return nil
}

// detectAVSVersion extracts version from AVS schema location
func detectAVSVersion(schemaLocation string) string {
	// Look for patterns like "avs_20200108.xsd"
	if strings.Contains(schemaLocation, "avs_") {
		// Extract date pattern: avs_YYYYMMDD.xsd
		parts := strings.Split(schemaLocation, "avs_")
		if len(parts) >= 2 {
			datePart := strings.Split(parts[1], ".")[0]
			// Validate it looks like a date (8 digits)
			if len(datePart) == 8 {
				for _, r := range datePart {
					if r < '0' || r > '9' {
						return "" // current/latest
					}
				}
				return datePart
			}
		}
	}
	return "" // current/latest version
}

//
// =======================
// Codegen per namespace
// =======================
//

func generateProtoForBundle(
	b *NamespaceBundle,
	packageName string,
	goPackage string,
	all map[string]protoPkgInfo,
	avsVersionContext map[string]string,
) (string, error) {

	var sb strings.Builder

	// Header
	sb.WriteString(`syntax = "proto3";` + "\n\n")
	sb.WriteString(fmt.Sprintf("package %s;\n\n", packageName))
	sb.WriteString(fmt.Sprintf("option go_package = \"%s\";\n\n", goPackage))
	sb.WriteString(fmt.Sprintf("// Target namespace: %s\n\n", b.TargetNamespace))

	// Imports (protobuf)
	// Sort for determinism
	var deps []string
	for ns := range b.Imports {
		if ns == b.TargetNamespace {
			continue
		}

		// Handle AVS import version mapping
		if ns == "http://ddex.net/xml/avs/avs" || ns == "http://ddex.net/xml/allowed-value-sets" {
			avsVersion := avsVersionContext[b.TargetNamespace]

			// Default to latest if no specific version detected
			if avsVersion == "" {
				avsVersion = "latest"
			}

			// Construct the versioned import path directly
			versionedPath := fmt.Sprintf("ddex/avs/v%s/v%s.proto", avsVersion, avsVersion)
			deps = append(deps, versionedPath)
		} else if info, ok := all[ns]; ok {
			deps = append(deps, info.filePath)
		}
	}
	sort.Strings(deps)
	for _, f := range deps {
		// Normalize to POSIX paths in import statements
		sb.WriteString(fmt.Sprintf("import \"%s\";\n", toPosixPath(f)))
	}
	if len(deps) > 0 {
		sb.WriteString("\n")
	}

	// Track generated type names (message & enum in one space) for this package
	generated := make(map[string]struct{})

	// Top-level elements with inline complex types → message
	for _, el := range b.Elements {
		if el.ComplexType != nil {
			name := toProtoMessageName(el.Name)
			if _, exists := generated[name]; !exists {
				msg, wrapperTypes, err := generateComplexTypeMessageWithNamespace(el.Name, el.ComplexType, all, b.TargetNamespace, true, b)
				if err != nil {
					return "", err
				}
				sb.WriteString(msg)
				sb.WriteString("\n\n")
				generated[name] = struct{}{}

				// Generate wrapper types
				for _, wrapper := range wrapperTypes {
					if _, exists := generated[wrapper.Name]; !exists {
						if wrapper.IsChoice {
							wrapperMsg, err := generateChoiceWrapper(wrapper.Name, wrapper.Choice, all)
							if err != nil {
								return "", err
							}
							sb.WriteString(wrapperMsg)
							sb.WriteString("\n\n")
						} else {
							wrapperMsg, err := generateSequenceType(wrapper.Name, wrapper.Sequence, all)
							if err != nil {
								return "", err
							}
							sb.WriteString(wrapperMsg)
							sb.WriteString("\n\n")
						}
						generated[wrapper.Name] = struct{}{}
					}
				}
			}
		}
	}

	// Named complex types → message
	for _, ct := range b.ComplexTypes {
		if ct.Name == "" {
			continue
		}
		name := toProtoMessageName(ct.Name)
		if _, exists := generated[name]; exists {
			continue
		}
		msg, wrapperTypes, err := generateComplexTypeMessage(ct.Name, &ct, all)
		if err != nil {
			return "", err
		}
		sb.WriteString(msg)
		sb.WriteString("\n\n")
		generated[name] = struct{}{}

		// Generate wrapper types
		for _, wrapper := range wrapperTypes {
			if _, exists := generated[wrapper.Name]; !exists {
				if wrapper.IsChoice {
					wrapperMsg, err := generateChoiceWrapper(wrapper.Name, wrapper.Choice, all)
					if err != nil {
						return "", err
					}
					sb.WriteString(wrapperMsg)
					sb.WriteString("\n\n")
				} else {
					wrapperMsg, err := generateSequenceType(wrapper.Name, wrapper.Sequence, all)
					if err != nil {
						return "", err
					}
					sb.WriteString(wrapperMsg)
					sb.WriteString("\n\n")
				}
				generated[wrapper.Name] = struct{}{}
			}
		}
	}

	// Simple types with enumerations → enum
	for _, st := range b.SimpleTypes {
		if st.Name == "" || st.Restriction == nil || len(st.Restriction.Enumerations) == 0 {
			continue
		}
		en := toProtoEnumName(st.Name)
		if _, exists := generated[en]; exists {
			continue
		}
		sb.WriteString(generateEnum(st))
		sb.WriteString("\n\n")
		generated[en] = struct{}{}
	}

	return strings.TrimSpace(sb.String()) + "\n", nil
}

//
// =======================
// Field/message/enum codegen (your original logic, kept)
// =======================
//

// WrapperType represents a wrapper type that needs to be generated
type WrapperType struct {
	Name     string
	Choice   *XSDChoice
	Sequence *XSDSequence
	IsChoice bool
}

func generateComplexTypeMessage(name string, complexType *XSDComplexType, allPkgs map[string]protoPkgInfo) (string, []WrapperType, error) {
	return generateComplexTypeMessageWithNamespace(name, complexType, allPkgs, "", false, nil)
}

func generateComplexTypeMessageWithNamespace(name string, complexType *XSDComplexType, allPkgs map[string]protoPkgInfo, targetNamespace string, isRootElement bool, bundle *NamespaceBundle) (string, []WrapperType, error) {
	var wrapperTypes []WrapperType
	var builder strings.Builder

	messageName := toProtoMessageName(name)
	builder.WriteString(fmt.Sprintf("message %s {\n", messageName))

	fieldNum := 1
	usedFieldNames := make(map[string]int) // track used field names and their counts

	// sequence → fields
	if complexType.Sequence != nil {
		// Process regular elements with proper repeated field handling
		for _, element := range complexType.Sequence.Elements {
			field, err := generateFieldWithDedup(element, fieldNum, allPkgs, usedFieldNames)
			if err != nil {
				return "", nil, fmt.Errorf("failed to generate field for element %s: %v", element.Name, err)
			}
			// Only add field and increment field number if field was actually generated
			if field != "" {
				builder.WriteString(field + "\n")
				fieldNum++
			}
		}

		// Process choice elements within sequence - flatten them directly into the parent message
		for _, choice := range complexType.Sequence.Choices {
			err := generateChoiceFields(&choice, &fieldNum, allPkgs, usedFieldNames, &builder)
			if err != nil {
				return "", nil, fmt.Errorf("failed to generate choice fields: %v", err)
			}
		}
	}

	// choice → flatten choice elements directly into the parent message (for top-level choices)
	if complexType.Choice != nil && complexType.Sequence == nil && (len(complexType.Choice.Elements) > 0 || len(complexType.Choice.Sequences) > 0) {
		err := generateChoiceFields(complexType.Choice, &fieldNum, allPkgs, usedFieldNames, &builder)
		if err != nil {
			return "", nil, fmt.Errorf("failed to generate top-level choice fields: %v", err)
		}
	}

	// simpleContent extension → value + attributes
	if complexType.SimpleContent != nil && complexType.SimpleContent.Extension != nil {
		// chardata value
		injectComment := "  // @gotags: xml:\",chardata\""
		fieldName := getUniqueFieldName("value", usedFieldNames)
		builder.WriteString(fmt.Sprintf("%s\n  string %s = %d;\n", injectComment, fieldName, fieldNum))
		fieldNum++

		// attributes
		for _, attr := range complexType.SimpleContent.Extension.Attributes {
			field := generateAttributeFieldWithDedup(attr, fieldNum, allPkgs, usedFieldNames)
			builder.WriteString(field + "\n")
			fieldNum++
		}
	}

	// attributes on the complexType itself
	for _, attr := range complexType.Attributes {
		field := generateAttributeFieldWithDedup(attr, fieldNum, allPkgs, usedFieldNames)
		builder.WriteString(field + "\n")
		fieldNum++
	}

	// Add namespace attributes for root elements
	if isRootElement && targetNamespace != "" {
		// Extract namespace prefix from target namespace URL
		// e.g., "http://ddex.net/xml/ern/43" -> "ern"
		namespacePrefix := extractNamespacePrefix(targetNamespace)

		if namespacePrefix != "" {
			// Add the namespace prefix attribute (e.g., xmlns:ern)
			injectComment := fmt.Sprintf("  // @gotags: xml:\"xmlns:%s,attr\"", namespacePrefix)
			field := fmt.Sprintf("%s\n  string xmlns_%s = %d;", injectComment, namespacePrefix, fieldNum)
			builder.WriteString(field + "\n")
			fieldNum++
		}

		// Add XSI namespace attribute
		injectComment := fmt.Sprintf("  // @gotags: xml:\"xmlns:xsi,attr\"")
		field := fmt.Sprintf("%s\n  string xmlns_xsi = %d;", injectComment, fieldNum)
		builder.WriteString(field + "\n")
		fieldNum++

		// Add schema location attribute (this one needs xsi: prefix, not xmlns:)
		injectComment = fmt.Sprintf("  // @gotags: xml:\"xsi:schemaLocation,attr\"")
		field = fmt.Sprintf("%s\n  string xsi_schema_location = %d;", injectComment, fieldNum)
		builder.WriteString(field + "\n")
		fieldNum++

		// Add AVS namespace attribute if this bundle imports AVS
		if bundle != nil {
			for importNS := range bundle.Imports {
				if importNS == "http://ddex.net/xml/avs/avs" || importNS == "http://ddex.net/xml/allowed-value-sets" {
					injectComment = fmt.Sprintf("  // @gotags: xml:\"xmlns:avs,attr\"")
					field = fmt.Sprintf("%s\n  string xmlns_avs = %d;", injectComment, fieldNum)
					builder.WriteString(field + "\n")
					fieldNum++
					break
				}
			}
		}
	}

	builder.WriteString("}")
	return builder.String(), wrapperTypes, nil
}

// extractNamespacePrefix extracts the namespace prefix from a target namespace URL
// e.g., "http://ddex.net/xml/ern/43" -> "ern"
// e.g., "http://ddex.net/xml/mead/11" -> "mead"
// e.g., "http://ddex.net/xml/pie/10" -> "pie"
func extractNamespacePrefix(targetNamespace string) string {
	// Handle DDEX namespace patterns
	if strings.Contains(targetNamespace, "ddex.net/xml/") {
		// Extract the part after "/xml/"
		parts := strings.Split(targetNamespace, "/xml/")
		if len(parts) >= 2 {
			// Get the schema type (e.g., "ern/43" -> "ern")
			schemaParts := strings.Split(parts[1], "/")
			if len(schemaParts) >= 1 {
				return schemaParts[0]
			}
		}
	}
	return ""
}

// getUniqueFieldName ensures field names are unique within a message by adding suffixes
func getUniqueFieldName(baseName string, usedFieldNames map[string]int) string {
	if count, exists := usedFieldNames[baseName]; exists {
		count++
		usedFieldNames[baseName] = count
		return fmt.Sprintf("%s_%d", baseName, count)
	}
	usedFieldNames[baseName] = 0
	return baseName
}

// generateFieldWithDedup generates a field with deduplication
func generateFieldWithDedup(element XSDElement, fieldNum int, allPkgs map[string]protoPkgInfo, usedFieldNames map[string]int) (string, error) {
	originalFieldName := toProtoFieldName(element.Name)

	// For repeated elements, don't use deduplication - use the original name
	// This allows multiple XML elements with the same name to map to a single repeated field
	var fieldName string
	if element.MaxOccurs == "unbounded" {
		// Check if we already have this field name as a repeated field
		if count, exists := usedFieldNames[originalFieldName]; exists && count == -1 {
			// Already generated as repeated, skip this occurrence
			return "", nil
		}
		// Mark this as a repeated field (use -1 as special marker)
		usedFieldNames[originalFieldName] = -1
		fieldName = originalFieldName
	} else {
		// For non-repeated elements, use deduplication as before
		fieldName = getUniqueFieldName(originalFieldName, usedFieldNames)
	}

	// Type mapping
	fieldType := "string" // default
	if element.Type != "" {
		fieldType = xsdTypeToProto(element.Type, allPkgs)
	}

	// Cardinality
	repeated := ""
	if element.MaxOccurs == "unbounded" {
		repeated = "repeated "
	}

	// gotags for xml element name
	injectComment := fmt.Sprintf("  // @gotags: xml:\"%s\"", element.Name)

	return fmt.Sprintf("%s\n  %s%s %s = %d;", injectComment, repeated, fieldType, fieldName, fieldNum), nil
}

// generateChoiceFieldWithDedup generates a choice field with deduplication
func generateChoiceFieldWithDedup(element XSDElement, fieldNum int, allPkgs map[string]protoPkgInfo, usedFieldNames map[string]int) (string, error) {
	fieldName := getUniqueFieldName(toProtoFieldName(element.Name), usedFieldNames)

	fieldType := "string"
	if element.Type != "" {
		fieldType = xsdTypeToProto(element.Type, allPkgs)
	}

	injectComment := fmt.Sprintf("  // @gotags: xml:\"%s\"", element.Name)
	return fmt.Sprintf("%s\n    %s %s = %d;", injectComment, fieldType, fieldName, fieldNum), nil
}

// generateAttributeFieldWithDedup generates an attribute field with deduplication
func generateAttributeFieldWithDedup(attr XSDAttribute, fieldNum int, allPkgs map[string]protoPkgInfo, usedFieldNames map[string]int) string {
	fieldName := getUniqueFieldName(toProtoFieldName(attr.Name), usedFieldNames)

	fieldType := "string"
	if attr.Type != "" {
		fieldType = xsdTypeToProto(attr.Type, allPkgs)
	}

	injectComment := fmt.Sprintf("  // @gotags: xml:\"%s,attr\"", attr.Name)
	return fmt.Sprintf("%s\n  %s %s = %d;", injectComment, fieldType, fieldName, fieldNum)
}

func generateField(element XSDElement, fieldNum int, allPkgs map[string]protoPkgInfo) (string, error) {
	fieldName := toProtoFieldName(element.Name)

	// Type mapping
	fieldType := "string" // default
	if element.Type != "" {
		fieldType = xsdTypeToProto(element.Type, allPkgs)
	}

	// Cardinality
	repeated := ""
	if element.MaxOccurs == "unbounded" {
		repeated = "repeated "
	}

	// gotags for xml element name
	injectComment := fmt.Sprintf("  // @gotags: xml:\"%s\"", element.Name)

	return fmt.Sprintf("%s\n  %s%s %s = %d;", injectComment, repeated, fieldType, fieldName, fieldNum), nil
}

func generateChoiceField(element XSDElement, fieldNum int, allPkgs map[string]protoPkgInfo) (string, error) {
	fieldName := toProtoFieldName(element.Name)

	fieldType := "string"
	if element.Type != "" {
		fieldType = xsdTypeToProto(element.Type, allPkgs)
	}

	injectComment := fmt.Sprintf("  // @gotags: xml:\"%s\"", element.Name)
	return fmt.Sprintf("%s\n    %s %s = %d;", injectComment, fieldType, fieldName, fieldNum), nil
}

func generateAttributeField(attr XSDAttribute, fieldNum int, allPkgs map[string]protoPkgInfo) string {
	fieldName := toProtoFieldName(attr.Name)

	fieldType := "string"
	if attr.Type != "" {
		fieldType = xsdTypeToProto(attr.Type, allPkgs)
	}

	injectComment := fmt.Sprintf("  // @gotags: xml:\"%s,attr\"", attr.Name)
	return fmt.Sprintf("%s\n  %s %s = %d;", injectComment, fieldType, fieldName, fieldNum)
}

func generateEnum(simpleType XSDSimpleType) string {
	var builder strings.Builder

	enumName := strings.ReplaceAll(toProtoMessageName(simpleType.Name), "_", "")
	builder.WriteString(fmt.Sprintf("enum %s {\n", enumName))

	// Use the full enum name as prefix to avoid collisions between different enums
	enumPrefix := strings.ToUpper(toProtoFieldName(simpleType.Name))
	// Clean up any double underscores
	for strings.Contains(enumPrefix, "__") {
		enumPrefix = strings.ReplaceAll(enumPrefix, "__", "_")
	}

	builder.WriteString(fmt.Sprintf("  %s_UNSPECIFIED = 0;\n", enumPrefix))

	// Deduplicate enum values to avoid conflicts within the same enum
	seenValues := make(map[string]struct{})
	valueIndex := 1

	for _, enum := range simpleType.Restriction.Enumerations {
		rawValue := toProtoEnumValue(enum.Value)
		enumValue := enumPrefix + "_" + rawValue

		// Skip if we've already seen this enum value
		if _, exists := seenValues[enumValue]; exists {
			continue
		}
		seenValues[enumValue] = struct{}{}

		builder.WriteString(fmt.Sprintf("  %s = %d;\n", enumValue, valueIndex))
		valueIndex++
	}

	builder.WriteString("}")
	return builder.String()
}

// generateChoiceFields flattens choice elements directly into the parent message
func generateChoiceFields(choice *XSDChoice, fieldNum *int, allPkgs map[string]protoPkgInfo, usedFieldNames map[string]int, builder *strings.Builder) error {
	// Handle direct elements in choice - flatten them into the parent message
	for _, element := range choice.Elements {
		field, err := generateFieldWithDedup(element, *fieldNum, allPkgs, usedFieldNames)
		if err != nil {
			return fmt.Errorf("failed to generate choice element field for %s: %v", element.Name, err)
		}
		if field != "" {
			builder.WriteString(field + "\n")
			*fieldNum++
		}
	}

	// Handle sequences within choice - flatten their elements into the parent message
	for _, sequence := range choice.Sequences {
		for _, element := range sequence.Elements {
			field, err := generateFieldWithDedup(element, *fieldNum, allPkgs, usedFieldNames)
			if err != nil {
				return fmt.Errorf("failed to generate choice sequence element field for %s: %v", element.Name, err)
			}
			if field != "" {
				builder.WriteString(field + "\n")
				*fieldNum++
			}
		}
	}

	return nil
}

// generateChoiceWrapper generates a wrapper message type for a choice
func generateChoiceWrapper(wrapperName string, choice *XSDChoice, allPkgs map[string]protoPkgInfo) (string, error) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("message %s {\n", wrapperName))

	// Generate the oneof with proper XML tags for direct marshalling
	builder.WriteString("  oneof choice {\n")
	oneofFieldNum := 1

	// Handle direct elements in choice
	for _, element := range choice.Elements {
		fieldType := "string"
		if element.Type != "" {
			fieldType = xsdTypeToProto(element.Type, allPkgs)
		}

		fieldName := toProtoFieldName(element.Name)

		if element.MaxOccurs == "unbounded" {
			// For repeated elements, create a separate message type
			optionName := fmt.Sprintf("%sOption", toProtoMessageName(element.Name))

			// Generate a nested message for the repeated element
			builder.WriteString(fmt.Sprintf("  message %s {\n", optionName))
			injectComment := fmt.Sprintf("    // @gotags: xml:\"%s\"", element.Name)
			field := fmt.Sprintf("%s\n    repeated %s %s = 1;", injectComment, fieldType, fieldName)
			builder.WriteString(field + "\n")
			builder.WriteString("  }\n\n")

			// Add the message option to the oneof with inline XML
			injectComment = "    // @gotags: xml:\",inline\""
			field = fmt.Sprintf("%s\n    %s %s = %d;", injectComment, optionName, fieldName+"_option", oneofFieldNum)
			builder.WriteString(field + "\n")
		} else {
			// Add XML tag to make choice fields marshal directly as the element name
			injectComment := fmt.Sprintf("    // @gotags: xml:\"%s\"", element.Name)
			field := fmt.Sprintf("%s\n    %s %s = %d;", injectComment, fieldType, fieldName, oneofFieldNum)
			builder.WriteString(field + "\n")
		}
		oneofFieldNum++
	}

	// Handle sequences in choice by creating inline fields for each sequence element
	for seqIdx, sequence := range choice.Sequences {
		// Create a message type for this sequence
		optionName := fmt.Sprintf("Sequence%d", seqIdx+1)

		// Generate a nested message for the sequence
		builder.WriteString(fmt.Sprintf("  message %s {\n", optionName))

		fieldNum := 1
		usedFieldNames := make(map[string]int)

		for _, element := range sequence.Elements {
			field, err := generateFieldWithDedupForNested(element, fieldNum, allPkgs, usedFieldNames)
			if err != nil {
				return "", fmt.Errorf("failed to generate sequence field for element %s: %v", element.Name, err)
			}
			if field != "" {
				builder.WriteString("  " + field + "\n")
				fieldNum++
			}
		}

		builder.WriteString("  }\n\n")

		// Add the sequence option to the oneof with inline XML
		fieldName := toProtoFieldName(fmt.Sprintf("sequence_%d", seqIdx+1))
		injectComment := "    // @gotags: xml:\",inline\""
		field := fmt.Sprintf("%s\n    %s %s = %d;", injectComment, optionName, fieldName, oneofFieldNum)
		builder.WriteString(field + "\n")
		oneofFieldNum++
	}

	builder.WriteString("  }\n")
	builder.WriteString("}")
	return builder.String(), nil
}

// generateFieldWithDedupForNested is like generateFieldWithDedup but for nested messages
func generateFieldWithDedupForNested(element XSDElement, fieldNum int, allPkgs map[string]protoPkgInfo, usedFieldNames map[string]int) (string, error) {
	originalFieldName := toProtoFieldName(element.Name)

	// For repeated elements, don't use deduplication - use the original name
	var fieldName string
	if element.MaxOccurs == "unbounded" {
		// Check if we already have this field name as a repeated field
		if count, exists := usedFieldNames[originalFieldName]; exists && count == -1 {
			// Already generated as repeated, skip this occurrence
			return "", nil
		}
		// Mark this as a repeated field (use -1 as special marker)
		usedFieldNames[originalFieldName] = -1
		fieldName = originalFieldName
	} else {
		// For non-repeated elements, use deduplication as before
		fieldName = getUniqueFieldName(originalFieldName, usedFieldNames)
	}

	// Type mapping
	fieldType := "string" // default
	if element.Type != "" {
		fieldType = xsdTypeToProto(element.Type, allPkgs)
	}

	// Cardinality
	repeated := ""
	if element.MaxOccurs == "unbounded" {
		repeated = "repeated "
	}

	// gotags for xml element name
	injectComment := fmt.Sprintf("  // @gotags: xml:\"%s\"", element.Name)

	return fmt.Sprintf("%s\n  %s%s %s = %d;", injectComment, repeated, fieldType, fieldName, fieldNum), nil
}

// generateSequenceType generates a message type for a sequence within a choice
func generateSequenceType(typeName string, sequence *XSDSequence, allPkgs map[string]protoPkgInfo) (string, error) {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("message %s {\n", typeName))

	fieldNum := 1
	usedFieldNames := make(map[string]int)

	// Generate fields for elements in the sequence
	for _, element := range sequence.Elements {
		field, err := generateFieldWithDedup(element, fieldNum, allPkgs, usedFieldNames)
		if err != nil {
			return "", fmt.Errorf("failed to generate sequence field for element %s: %v", element.Name, err)
		}
		builder.WriteString(field + "\n")
		fieldNum++
	}

	builder.WriteString("}")
	return builder.String(), nil
}

//
// =======================
// Type/name helpers (kept + small improvements)
// =======================
//

func xsdTypeToProto(xsdType string, allPkgs map[string]protoPkgInfo) string {
	originalType := xsdType
	var prefix string

	// Extract prefix if present (xs:, avs:, ern:, etc.)
	if idx := strings.Index(xsdType, ":"); idx != -1 {
		prefix = xsdType[:idx]
		xsdType = xsdType[idx+1:]
	}

	switch xsdType {
	case "string", "normalizedString", "token", "anyURI", "NMTOKEN":
		return "string"
	case "int", "integer", "positiveInteger", "PositiveInteger":
		return "int32"
	case "long":
		return "int64"
	case "boolean":
		return "bool"
	case "decimal", "float":
		return "string" // preserve precision for decimals
	case "double":
		return "double"
	case "dateTime", "date", "time", "duration", "gYear", "GYear", "ddex_IsoDate", "Ddex_IsoDate":
		return "string" // ISO8601 strings
	case "base64Binary":
		return "bytes"
	default:
		// Handle namespace prefixes for custom types
		if prefix == "avs" {
			// AVS contains only enum types, which we represent as strings in messages
			// for XML compatibility. The actual enum definitions still exist in ddex.avs
			// for type safety when developers want to use them programmatically.
			return "string"
		}

		// For other prefixes, try to map to known packages
		if prefix != "" && prefix != "xs" {
			// Look for a namespace that might match this prefix
			for ns, pkg := range allPkgs {
				if strings.Contains(strings.ToLower(ns), prefix) {
					packageName := pkg.pkgName
					return packageName + "." + strings.ReplaceAll(toProtoMessageName(xsdType), "_", "")
				}
			}
		}

		// Assume custom type → Proto message in local package
		if xsdType != originalType {
			log.Printf("Unmapped XSD type: %s (original: %s) -> treating as custom message", xsdType, originalType)
		} else {
			log.Printf("Unmapped XSD type: %s -> treating as custom message", xsdType)
		}
		return strings.ReplaceAll(toProtoMessageName(xsdType), "_", "")
	}
}

func toProtoFieldName(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteByte(byte(strings.ToLower(string(r))[0]))
	}
	return b.String()
}

func toProtoMessageName(name string) string {
	if name == "" {
		return ""
	}
	// Basic PascalCase
	return strings.ToUpper(name[:1]) + name[1:]
}

func toProtoEnumName(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func toProtoEnumValue(value string) string {
	result := strings.ToUpper(value)
	result = strings.ReplaceAll(result, "-", "_")
	result = strings.ReplaceAll(result, " ", "_")
	result = strings.ReplaceAll(result, ".", "_")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.ReplaceAll(result, "+", "_PLUS_")
	result = strings.ReplaceAll(result, "(", "_")
	result = strings.ReplaceAll(result, ")", "_")
	result = strings.ReplaceAll(result, "'", "_")
	result = strings.ReplaceAll(result, "\"", "_")
	result = strings.ReplaceAll(result, "&", "_AND_")

	var cleaned strings.Builder
	for _, r := range result {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			cleaned.WriteRune(r)
		}
	}
	out := cleaned.String()
	if out == "" {
		out = "UNKNOWN"
	}
	if out[0] >= '0' && out[0] <= '9' {
		out = "E_" + out
	}
	for strings.Contains(out, "__") {
		out = strings.ReplaceAll(out, "__", "_")
	}
	return strings.Trim(out, "_")
}

func toPosixPath(p string) string {
	return strings.ReplaceAll(p, string(os.PathSeparator), "/")
}

//
// =======================
// Namespace → package mapping
// =======================
//

// For DDEX, we want:
//
//	ddex.xml/ern/43 → package "ddex.ern.v43"
//	ddex.xml/mead/11 → "ddex.mead.v11"
//	ddex.xml/pie/10 → "ddex.pie.v10"
//	ddex.xml/avs/avs → "ddex.avs"
func namespaceToProtoPackage(ns string, bundle *NamespaceBundle, spec struct{ name, version, mainFile string }) string {
	host, pathParts := splitNS(ns)

	// DDEX-friendly mapping
	if host == "ddex.net" && len(pathParts) >= 2 && pathParts[0] == "xml" {
		// AVS in the wild appears as: /xml/avs/avs, /xml/allowed-value-sets, /xml/allowed_value_sets
		if pathParts[1] == "avs" ||
			pathParts[1] == "allowed-value-sets" ||
			pathParts[1] == "allowed_value_sets" {
			// All AVS specs get versioned packages now
			if spec.name == "avs" {
				return fmt.Sprintf("ddex.avs.v%s", spec.version)
			}
			return "ddex.avs"
		}
		// Normal versioned families: /xml/{ern|mead|pie}/{digits}
		if len(pathParts) >= 3 && isDigits(pathParts[2]) {
			return fmt.Sprintf("ddex.%s.v%s", pathParts[1], pathParts[2])
		}
	}

	// Fallback: if this namespace matches the entry spec (ern/mead/pie…), pin it
	if looksLikeEntry(ns, spec) {
		return fmt.Sprintf("ddex.%s.v%s", spec.name, spec.version)
	}

	// Generic fallback: reverse host + path; add v<digits> suffix where appropriate
	revHost := reverseHost(host)
	if len(pathParts) > 0 && isDigits(pathParts[len(pathParts)-1]) {
		last := "v" + pathParts[len(pathParts)-1]
		return sanitizePackage(strings.Join(append([]string{revHost}, append(pathParts[:len(pathParts)-1], last)...), "."))
	}
	return sanitizePackage(strings.Join(append([]string{revHost}, pathParts...), "."))
}

func looksLikeEntry(ns string, spec struct{ name, version, mainFile string }) bool {
	// ddex.net/xml/<spec>/<versionDigits>, but never treat AVS as an entry package
	host, parts := splitNS(ns)
	if host != "ddex.net" || len(parts) < 3 || parts[0] != "xml" {
		return false
	}
	if parts[1] == "avs" || parts[1] == "allowed-value-sets" || parts[1] == "allowed_value_sets" {
		return false
	}
	return parts[1] == spec.name && isDigits(parts[2]) && parts[2] == stripLeadingV(spec.version)
}

func namespaceToGoPackage(ns string, bundle *NamespaceBundle, spec struct{ name, version, mainFile string }) string {
	// Put Go package paths under your repo. Mirror the proto package path as directories.
	pkg := namespaceToProtoPackage(ns, bundle, spec)
	path := strings.ReplaceAll(pkg, ".", "/")
	return "github.com/OpenAudio/ddex-proto/gen/" + path
}

func packageToPath(pkg string) string {
	parts := strings.Split(pkg, ".")
	if len(parts) == 0 {
		return "unknown.proto"
	}
	dir := strings.Join(parts, "/")
	filename := parts[len(parts)-1] + ".proto"
	return filepath.Join(dir, filename)
}

func splitNS(ns string) (host string, parts []string) {
	u, err := url.Parse(ns)
	if err != nil {
		return "", nil
	}
	host = u.Host
	parts = strings.Split(strings.Trim(u.Path, "/"), "/")
	return
}

func reverseHost(h string) string {
	if h == "" {
		return "unknown"
	}
	parts := strings.Split(h, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, ".")
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func sanitizePackage(p string) string {
	// very light cleanup: no invalid idents
	p = strings.ReplaceAll(p, "-", "_")
	p = strings.ReplaceAll(p, " ", "_")
	return p
}

func stripLeadingV(s string) string {
	return strings.TrimPrefix(strings.ToLower(s), "v")
}
