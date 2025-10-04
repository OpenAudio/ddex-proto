package ddexgen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// generateExtensions generates enum_strings.go, *.xml.go, and registry.go files
func Generate(targetDir string, verbose bool) error {
	var allPackages []PackageInfo

	// Find all generated protobuf packages
	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".pb.go") {
			packageDir := filepath.Dir(path)

			// Extract the actual package name from the .pb.go file
			packageName, err := extractPackageName(path)
			if err != nil {
				return fmt.Errorf("extracting package name from %s: %w", path, err)
			}

			// Parse the .pb.go file to find enum types and message types
			enums, err := findEnumTypes(path)
			if err != nil {
				return fmt.Errorf("parsing %s: %w", path, err)
			}

			messages, err := findMessageTypes(path)
			if err != nil {
				return fmt.Errorf("parsing messages %s: %w", path, err)
			}

			// Generate enum strings file if there are enums
			if len(enums) > 0 {
				err = generateEnumStringsFile(packageDir, packageName, enums)
				if err != nil {
					return fmt.Errorf("generating enum strings file for %s: %w", packageDir, err)
				}
				if verbose {
					log.Printf("Generated enum_strings.go for package %s with %d enums", packageName, len(enums))
				}
			}

			// Generate single XML file for all messages in the package
			if len(messages) > 0 {
				err = generatePackageXMLFile(packageDir, packageName, messages)
				if err != nil {
					return fmt.Errorf("generating XML file for package %s: %w", packageDir, err)
				}
				if verbose {
					log.Printf("Generated %s.xml.go for package %s with %d messages", packageName, packageName, len(messages))
				}
			}

			// Collect package info for registry generation (only DDEX packages with messages)
			if len(messages) > 0 && strings.Contains(packageDir, "ddex") {
				nsInfo := deriveNamespaceInfo(packageDir)
				if nsInfo != nil {
					// Generate import path relative to target directory
					relPath, err := filepath.Rel(targetDir, packageDir)
					if err != nil {
						return fmt.Errorf("failed to get relative path: %w", err)
					}
					importPath := "github.com/OpenAudio/ddex-proto/" + filepath.Join(filepath.Base(targetDir), relPath)

					allPackages = append(allPackages, PackageInfo{
						Dir:         packageDir,
						PackageName: packageName,
						ImportPath:  importPath,
						Messages:    messages,
						Namespace:   nsInfo,
					})
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	// Generate dynamic registry file
	if len(allPackages) > 0 {
		registryPath := filepath.Join(targetDir, "registry.go")
		err = generateRegistryFileAtPath(registryPath, allPackages)
		if err != nil {
			return fmt.Errorf("generating registry: %w", err)
		}
		if verbose {
			log.Printf("Generated registry.go with %d DDEX packages", len(allPackages))
		}
	}

	return nil
}

// extractPackageName reads the package declaration from a Go file
func extractPackageName(filename string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	return node.Name.Name, nil
}

// findEnumTypes parses a .pb.go file and extracts enum type information
func findEnumTypes(filename string) ([]EnumInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var enums []EnumInfo

	// Look for enum type definitions and their constants
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if ident, ok := ts.Type.(*ast.Ident); ok && ident.Name == "int32" {
							// Found an enum type - now find its constants
							enumName := ts.Name.Name
							constants := findEnumConstants(node, enumName)
							if len(constants) > 0 {
								enums = append(enums, EnumInfo{
									Name:      enumName,
									Constants: constants,
								})
							}
						}
					}
				}
			}
		}
	}

	return enums, nil
}

// findEnumConstants finds all constants for a given enum type
func findEnumConstants(node *ast.File, enumTypeName string) []string {
	var constants []string

	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					// Check if this constant is of our enum type
					if ident, ok := valueSpec.Type.(*ast.Ident); ok && ident.Name == enumTypeName {
						for _, name := range valueSpec.Names {
							constants = append(constants, name.Name)
						}
					}
				}
			}
		}
	}

	return constants
}

type EnumInfo struct {
	Name      string
	Constants []string
}

type MessageInfo struct {
	Name string
}

type PackageInfo struct {
	Dir         string
	PackageName string
	ImportPath  string
	Messages    []MessageInfo
	Namespace   *NamespaceInfo
}

// findMessageTypes parses a .pb.go file and extracts main message types
func findMessageTypes(filename string) ([]MessageInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var messages []MessageInfo

	// Look for main message type definitions (ones ending with "Message")
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if _, ok := ts.Type.(*ast.StructType); ok {
							// Found a struct type - check if it's a main message type
							messageName := ts.Name.Name
							if strings.HasSuffix(messageName, "Message") {
								messages = append(messages, MessageInfo{
									Name: messageName,
								})
							}
						}
					}
				}
			}
		}
	}

	return messages, nil
}

// generateEnumStringsFile creates an enum_strings.go file with String() methods and parsers
func generateEnumStringsFile(packageDir, packageName string, enums []EnumInfo) error {
	content := generateEnumStringsContent(packageName, enums)

	enumStringsPath := filepath.Join(packageDir, "enum_strings.go")
	return os.WriteFile(enumStringsPath, []byte(content), 0644)
}

// generatePackageXMLFile creates a single XML file for all messages in a package
func generatePackageXMLFile(packageDir, packageName string, messages []MessageInfo) error {
	content := generatePackageXMLContent(packageDir, packageName, messages)

	xmlFileName := packageName + ".xml.go"
	xmlPath := filepath.Join(packageDir, xmlFileName)
	return os.WriteFile(xmlPath, []byte(content), 0644)
}

// generateEnumStringsContent creates the content for enum_strings.go
func generateEnumStringsContent(packageName string, enums []EnumInfo) string {
	var sb strings.Builder

	// Package header
	sb.WriteString(fmt.Sprintf("// Code generated by generate-go-extensions. DO NOT EDIT.\n\n"))
	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	if len(enums) > 0 {
		sb.WriteString("import \"strings\"\n\n")
	}

	// Generate String() methods and parsers for each enum
	// These allow developers to use type-safe enum constants with string fields
	for _, enum := range enums {
		sb.WriteString(generateEnumStringMethod(enum))
		sb.WriteString("\n\n")
		sb.WriteString(generateEnumParser(enum))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// NamespaceInfo holds namespace configuration for a package
type NamespaceInfo struct {
	Namespace       string
	NamespacePrefix string
	SchemaFile      string
	ImportsAVS      bool // true if this schema imports AVS namespace
}

// deriveNamespaceInfo extracts namespace info from package directory path
func deriveNamespaceInfo(packageDir string) *NamespaceInfo {
	// packageDir is something like "gen/ddex/ern/v432"
	// We want to extract: ddex type (ern/mead/pie), version (432/43/11/10)

	parts := strings.Split(filepath.Clean(packageDir), string(filepath.Separator))
	if len(parts) < 4 {
		return nil // Not a DDEX package
	}

	// Look for the ddex directory and extract type/version
	ddexIndex := -1
	for i, part := range parts {
		if part == "ddex" {
			ddexIndex = i
			break
		}
	}

	if ddexIndex == -1 || ddexIndex+2 >= len(parts) {
		return nil // Not found or not enough parts
	}

	messageType := parts[ddexIndex+1] // ern, mead, pie
	version := parts[ddexIndex+2]     // v432, v43, v11, etc.

	// Remove 'v' prefix from version
	versionNumber := strings.TrimPrefix(version, "v")

	info := &NamespaceInfo{
		NamespacePrefix: messageType,
	}

	// Set namespace and schema file based on type
	switch messageType {
	case "ern":
		info.Namespace = fmt.Sprintf("http://ddex.net/xml/ern/%s", versionNumber)
		info.SchemaFile = "release-notification.xsd"
	case "mead":
		info.Namespace = fmt.Sprintf("http://ddex.net/xml/mead/%s", versionNumber)
		info.SchemaFile = "media-enrichment-and-description.xsd"
	case "pie":
		info.Namespace = fmt.Sprintf("http://ddex.net/xml/pie/%s", versionNumber)
		info.SchemaFile = "party-identification-and-enrichment.xsd"
	default:
		return nil
	}

	// Check if the schema imports AVS namespace
	info.ImportsAVS = checkAVSImport(messageType, versionNumber, info.SchemaFile)

	return info
}

// checkAVSImport checks if a schema file imports the AVS namespace
func checkAVSImport(messageType, versionNumber, schemaFile string) bool {
	// Construct the path to the schema file
	schemaDir := fmt.Sprintf("xsd/%sv%s", messageType, versionNumber)
	schemaPath := filepath.Join(schemaDir, schemaFile)

	// Read the schema file
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		// If we can't read the file, assume no AVS import
		return false
	}

	// Check for AVS namespace import
	contentStr := string(content)
	return strings.Contains(contentStr, "xmlns:avs=\"http://ddex.net/xml/avs/avs\"") ||
		strings.Contains(contentStr, "namespace=\"http://ddex.net/xml/avs/avs\"")
}

// generatePackageXMLContent creates the content for a package XML file
func generatePackageXMLContent(packageDir, packageName string, messages []MessageInfo) string {
	var sb strings.Builder

	// Package header
	sb.WriteString(fmt.Sprintf("// Code generated by generate-go-extensions. DO NOT EDIT.\n\n"))
	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// Derive namespace info from package path first to check if we need strings import
	nsInfo := deriveNamespaceInfo(packageDir)

	// Check if we need strings import
	needsStrings := false
	if nsInfo != nil {
		for _, message := range messages {
			if isRootMessage(message.Name) {
				needsStrings = true
				break
			}
		}
	}

	// Write imports
	if needsStrings {
		sb.WriteString("import (\n")
		sb.WriteString("\t\"encoding/xml\"\n")
		sb.WriteString("\t\"reflect\"\n")
		sb.WriteString("\t\"strings\"\n")
		sb.WriteString(")\n\n")
	} else {
		sb.WriteString("import \"encoding/xml\"\n\n")
	}
	if nsInfo != nil {
		sb.WriteString("// Package-level namespace constants\n")
		sb.WriteString("const (\n")
		sb.WriteString(fmt.Sprintf("\tNamespace = \"%s\"\n", nsInfo.Namespace))
		sb.WriteString(fmt.Sprintf("\tNamespacePrefix = \"%s\"\n", nsInfo.NamespacePrefix))
		sb.WriteString("\tNamespaceXSI = \"http://www.w3.org/2001/XMLSchema-instance\"\n")
		if nsInfo.ImportsAVS {
			sb.WriteString("\tNamespaceAVS = \"http://ddex.net/xml/avs/avs\"\n")
		}
		sb.WriteString(")\n\n")
	}

	// Generate XML marshaling methods for all messages in the package
	for i, message := range messages {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(generateXMLMarshalingMethods(message, nsInfo))
	}

	return sb.String()
}

// generateEnumStringMethod creates a String() method for the enum type
func generateEnumStringMethod(enum EnumInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("// XMLString returns the XML string representation of %s\n", enum.Name))
	sb.WriteString(fmt.Sprintf("func (e %s) XMLString() string {\n", enum.Name))
	sb.WriteString("\tswitch e {\n")

	// Generate cases for each constant
	for _, constant := range enum.Constants {
		if strings.HasSuffix(constant, "_UNSPECIFIED") {
			continue // Skip UNSPECIFIED values
		}

		// Extract the meaningful part of the constant name
		upperName := strings.ToUpper(enum.Name)
		idx := strings.LastIndex(constant, upperName+"_")
		if idx >= 0 {
			afterPrefix := constant[idx+len(upperName)+1:]
			if afterPrefix != "" && afterPrefix != "UNSPECIFIED" {
				sb.WriteString(fmt.Sprintf("\tcase %s:\n", constant))
				sb.WriteString(fmt.Sprintf("\t\treturn \"%s\"\n", afterPrefix))
			}
		}
	}

	sb.WriteString("\tdefault:\n")
	sb.WriteString("\t\treturn \"\"\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}")

	return sb.String()
}

// generateEnumParser creates the parser function for an enum
func generateEnumParser(enum EnumInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("// Parse%sString parses a string value to %s enum (case-insensitive)\n", enum.Name, enum.Name))
	sb.WriteString(fmt.Sprintf("func Parse%sString(s string) (%s, bool) {\n", enum.Name, enum.Name))
	sb.WriteString("\ts = strings.ToUpper(s)\n")
	sb.WriteString("\tswitch s {\n")

	// Generate cases for each constant
	for _, constant := range enum.Constants {
		if strings.HasSuffix(constant, "_UNSPECIFIED") {
			continue // Skip UNSPECIFIED values
		}

		// Extract the meaningful part of the constant name
		// Try to find the enum pattern: EnumName_ENUM_NAME_VALUE
		// We'll look for the last occurrence of the enum name in uppercase
		upperName := strings.ToUpper(enum.Name)

		// Find the pattern EnumName_..._VALUE
		idx := strings.LastIndex(constant, upperName+"_")
		if idx >= 0 {
			// Skip past "EnumName_..._" to get the value part
			afterPrefix := constant[idx+len(upperName)+1:]
			if afterPrefix != "" && afterPrefix != "UNSPECIFIED" {
				sb.WriteString(fmt.Sprintf("\tcase \"%s\":\n", afterPrefix))
				sb.WriteString(fmt.Sprintf("\t\treturn %s, true\n", constant))
			}
		}
	}

	sb.WriteString("\tdefault:\n")
	sb.WriteString(fmt.Sprintf("\t\treturn %s(0), false\n", enum.Name))
	sb.WriteString("\t}\n")
	sb.WriteString("}")

	return sb.String()
}

// generateXMLMarshalingMethods creates MarshalXML and UnmarshalXML methods for message types
func generateXMLMarshalingMethods(message MessageInfo, nsInfo *NamespaceInfo) string {
	var sb strings.Builder

	// Generate MarshalXML method
	sb.WriteString(fmt.Sprintf("// MarshalXML implements xml.Marshaler for %s\n", message.Name))
	sb.WriteString(fmt.Sprintf("func (m *%s) MarshalXML(e *xml.Encoder, start xml.StartElement) error {\n", message.Name))

	// Add namespace population for root message types if we have namespace info
	if nsInfo != nil && isRootMessage(message.Name) {
		sb.WriteString("\t// Set default namespace values if empty\n")
		sb.WriteString("\tif m.NamespaceAttrs == nil {\n")
		sb.WriteString("\t\tm.NamespaceAttrs = make(map[string]string)\n")
		sb.WriteString("\t}\n")

	}

	// Set the namespace on the start element for root messages
	if nsInfo != nil && isRootMessage(message.Name) {
		sb.WriteString("\t// Set the namespace on the start element\n")
		sb.WriteString("\tstart.Name.Space = Namespace\n\n")

		// Add namespace attributes to the start element
		sb.WriteString("\t// Add namespace attributes to the element, avoiding duplicates\n")
		sb.WriteString("\t// Use reflection to find which attributes are already handled by struct fields\n")
		sb.WriteString("\texistingAttrs := make(map[string]bool)\n")
		sb.WriteString("\tv := reflect.ValueOf(m).Elem()\n")
		sb.WriteString("\tt := v.Type()\n")
		sb.WriteString("\tfor i := 0; i < v.NumField(); i++ {\n")
		sb.WriteString("\t\tfield := t.Field(i)\n")
		sb.WriteString("\t\tif xmlTag := field.Tag.Get(\"xml\"); xmlTag != \"\" && xmlTag != \"-\" {\n")
		sb.WriteString("\t\t\t// Parse the XML tag to get the attribute name\n")
		sb.WriteString("\t\t\tif strings.HasSuffix(xmlTag, \",attr\") {\n")
		sb.WriteString("\t\t\t\tattrName := strings.TrimSuffix(xmlTag, \",attr\")\n")
		sb.WriteString("\t\t\t\tif colonIdx := strings.Index(attrName, \":\"); colonIdx >= 0 {\n")
		sb.WriteString("\t\t\t\t\t// For tags like \"xmlns:ern,attr\" or \"xsi:schemaLocation,attr\"\n")
		sb.WriteString("\t\t\t\t\texistingAttrs[attrName] = true\n")
		sb.WriteString("\t\t\t\t} else if attrName != \"\" {\n")
		sb.WriteString("\t\t\t\t\t// For tags like \"LanguageAndScriptCode,attr\"\n")
		sb.WriteString("\t\t\t\t\texistingAttrs[attrName] = true\n")
		sb.WriteString("\t\t\t\t}\n")
		sb.WriteString("\t\t\t}\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\t// Add attributes from the map that aren't already handled\n")
		sb.WriteString("\tfor key, value := range m.NamespaceAttrs {\n")
		sb.WriteString("\t\tif !existingAttrs[key] {\n")
		sb.WriteString("\t\t\tstart.Attr = append(start.Attr, xml.Attr{\n")
		sb.WriteString("\t\t\t\tName: xml.Name{Local: key},\n")
		sb.WriteString("\t\t\t\tValue: value,\n")
		sb.WriteString("\t\t\t})\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n\n")
	}

	sb.WriteString("\t// Create an alias type to avoid infinite recursion\n")
	sb.WriteString(fmt.Sprintf("\ttype alias %s\n", message.Name))
	sb.WriteString("\treturn e.EncodeElement((*alias)(m), start)\n")
	sb.WriteString("}\n\n")

	// Generate UnmarshalXML method
	sb.WriteString(fmt.Sprintf("// UnmarshalXML implements xml.Unmarshaler for %s\n", message.Name))
	sb.WriteString(fmt.Sprintf("func (m *%s) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {\n", message.Name))

	// Capture all attributes that aren't handled by explicit fields
	if nsInfo != nil && isRootMessage(message.Name) {
		sb.WriteString("\t// Capture all namespace and unhandled attributes\n")
		sb.WriteString("\tif m.NamespaceAttrs == nil {\n")
		sb.WriteString("\t\tm.NamespaceAttrs = make(map[string]string)\n")
		sb.WriteString("\t}\n")
		sb.WriteString("\tfor _, attr := range start.Attr {\n")
		sb.WriteString("\t\t// Capture all xmlns:* attributes and xsi:schemaLocation\n")
		sb.WriteString("\t\tif attr.Name.Space == \"xmlns\" || attr.Name.Local == \"xmlns\" ||\n")
		sb.WriteString("\t\t\t(attr.Name.Space == \"http://www.w3.org/2001/XMLSchema-instance\" && attr.Name.Local == \"schemaLocation\") {\n")
		sb.WriteString("\t\t\tkey := attr.Name.Local\n")
		sb.WriteString("\t\t\tif attr.Name.Space == \"xmlns\" {\n")
		sb.WriteString("\t\t\t\t// For namespace declarations like xmlns:ernm, xmlns:avs\n")
		sb.WriteString("\t\t\t\tkey = \"xmlns:\" + attr.Name.Local\n")
		sb.WriteString("\t\t\t} else if attr.Name.Space != \"\" && attr.Name.Local != \"xmlns\" {\n")
		sb.WriteString("\t\t\t\t// Preserve the namespace prefix for attributes like xsi:schemaLocation\n")
		sb.WriteString("\t\t\t\tif attr.Name.Space == \"http://www.w3.org/2001/XMLSchema-instance\" {\n")
		sb.WriteString("\t\t\t\t\tkey = \"xsi:\" + attr.Name.Local\n")
		sb.WriteString("\t\t\t\t}\n")
		sb.WriteString("\t\t\t}\n")
		sb.WriteString("\t\t\tm.NamespaceAttrs[key] = attr.Value\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n\n")
	}

	sb.WriteString("\t// Create an alias type to avoid infinite recursion\n")
	sb.WriteString(fmt.Sprintf("\ttype alias %s\n", message.Name))
	sb.WriteString("\treturn d.DecodeElement((*alias)(m), &start)\n")
	sb.WriteString("}")

	return sb.String()
}

// isRootMessage determines if a message type is a root message that needs namespace handling
func isRootMessage(messageName string) bool {
	switch messageName {
	case "NewReleaseMessage", "PurgeReleaseMessage", "CatalogListMessage", "MeadMessage", "PieMessage", "PieRequestMessage":
		return true
	default:
		return false
	}
}

// generateRegistryFile creates a registry.go file with dynamic message type registration
func generateRegistryFileAtPath(registryPath string, packages []PackageInfo) error {
	var sb strings.Builder

	// Package header
	sb.WriteString("// Code generated by generate-go-extensions. DO NOT EDIT.\n\n")
	sb.WriteString("package gen\n\n")

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"encoding/xml\"\n")
	sb.WriteString("\t\"fmt\"\n")
	sb.WriteString("\t\"reflect\"\n")
	sb.WriteString("\t\"strings\"\n\n")

	// Import all the generated packages
	sb.WriteString("\t// Auto-generated imports for all DDEX message types\n")
	for _, pkg := range packages {
		sb.WriteString(fmt.Sprintf("\t%s \"%s\"\n", pkg.PackageName, pkg.ImportPath))
	}
	sb.WriteString(")\n\n")

	// MessageTypeInfo struct
	sb.WriteString("// MessageTypeInfo holds information about a registered DDEX message type\n")
	sb.WriteString("type MessageTypeInfo struct {\n")
	sb.WriteString("\tType        reflect.Type\n")
	sb.WriteString("\tNamespace   string\n")
	sb.WriteString("\tRootElement string\n")
	sb.WriteString("}\n\n")

	// Registry map
	sb.WriteString("// messageRegistry maps \"messageType/version\" to MessageTypeInfo\n")
	sb.WriteString("var messageRegistry = map[string]MessageTypeInfo{\n")

	for _, pkg := range packages {
		messageType := pkg.Namespace.NamespacePrefix
		version := extractVersionFromPath(pkg.Dir)

		for _, msg := range pkg.Messages {
			if isRootMessage(msg.Name) {
				key := fmt.Sprintf("%s/%s/%s", messageType, version, msg.Name)
				sb.WriteString(fmt.Sprintf("\t\"%s\": {\n", key))
				sb.WriteString(fmt.Sprintf("\t\tType:        reflect.TypeOf(%s.%s{}),\n", pkg.PackageName, msg.Name))
				sb.WriteString(fmt.Sprintf("\t\tNamespace:   %s.Namespace,\n", pkg.PackageName))
				sb.WriteString(fmt.Sprintf("\t\tRootElement: \"%s\",\n", msg.Name))
				sb.WriteString("\t},\n")
			}
		}
	}
	sb.WriteString("}\n\n")

	// Generate all the registry functions
	sb.WriteString(generateRegistryFunctions())

	// Write the file
	return os.WriteFile(registryPath, []byte(sb.String()), 0644)
}

// extractVersionFromPath extracts version from a path like "gen/ddex/ern/v43"
func extractVersionFromPath(path string) string {
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for _, part := range parts {
		if strings.HasPrefix(part, "v") && len(part) > 1 {
			return part
		}
	}
	return ""
}

// generateRegistryFunctions creates all the registry utility functions
func generateRegistryFunctions() string {
	return `// GetRegisteredTypes returns all registered message types
func GetRegisteredTypes() map[string]MessageTypeInfo {
	result := make(map[string]MessageTypeInfo)
	for k, v := range messageRegistry {
		result[k] = v
	}
	return result
}

// New creates a new instance of the specified message type and version
// For message types with multiple root messages, uses the first one found
func New(messageType, version string) (interface{}, error) {
	// Find the first matching message type/version
	prefix := fmt.Sprintf("%s/%s/", messageType, version)
	for key, info := range messageRegistry {
		if strings.HasPrefix(key, prefix) {
			return reflect.New(info.Type).Interface(), nil
		}
	}
	return nil, fmt.Errorf("unknown message type: %s/%s", messageType, version)
}

// NewByMessageName creates a new instance of a specific message by name
func NewByMessageName(messageType, version, messageName string) (interface{}, error) {
	key := fmt.Sprintf("%s/%s/%s", messageType, version, messageName)
	info, ok := messageRegistry[key]
	if !ok {
		return nil, fmt.Errorf("unknown message: %s/%s/%s", messageType, version, messageName)
	}

	return reflect.New(info.Type).Interface(), nil
}

// DetectMessageType attempts to detect the message type, version, and message name from XML data
func DetectMessageType(xmlData []byte) (messageType, version, messageName string, err error) {
	// Parse just enough to get the root element and namespace
	decoder := xml.NewDecoder(strings.NewReader(string(xmlData)))

	for {
		token, err := decoder.Token()
		if err != nil {
			return "", "", "", fmt.Errorf("failed to parse XML: %w", err)
		}

		if startElement, ok := token.(xml.StartElement); ok {
			// Found the root element
			rootElement := startElement.Name.Local
			namespace := startElement.Name.Space

			// If no namespace in the element name, check attributes
			if namespace == "" {
				for _, attr := range startElement.Attr {
					if attr.Name.Local == "xmlns" || strings.HasPrefix(attr.Name.Local, "xmlns:") {
						namespace = attr.Value
						break
					}
				}
			}

			// Match against registered types
			for key, info := range messageRegistry {
				if info.RootElement == rootElement && info.Namespace == namespace {
					parts := strings.Split(key, "/")
					if len(parts) == 3 {
						return parts[0], parts[1], parts[2], nil
					}
				}
			}

			return "", "", "", fmt.Errorf("unknown DDEX message type with root element '%s' and namespace '%s'", rootElement, namespace)
		}
	}
}

// ParseAny automatically detects the message type and parses the XML accordingly
func ParseAny(xmlData []byte) (message interface{}, messageType, version string, err error) {
	// Detect the message type first
	msgType, ver, msgName, err := DetectMessageType(xmlData)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to detect message type: %w", err)
	}

	// Create a new instance of the detected type
	message, err = NewByMessageName(msgType, ver, msgName)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create message instance: %w", err)
	}

	// Unmarshal the XML into the message
	err = xml.Unmarshal(xmlData, message)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return message, msgType, ver, nil
}

// Parse parses XML data for a specific message type and version
func Parse(xmlData []byte, messageType, version string) (interface{}, error) {
	message, err := New(messageType, version)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(xmlData, message)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s/%s: %w", messageType, version, err)
	}

	return message, nil
}

// IsRegistered checks if a message type and version combination is registered
func IsRegistered(messageType, version string) bool {
	prefix := fmt.Sprintf("%s/%s/", messageType, version)
	for key := range messageRegistry {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// GetAvailableTypes returns a list of all available message types and versions
func GetAvailableTypes() []string {
	var types []string
	for key := range messageRegistry {
		types = append(types, key)
	}
	return types
}
`
}
