# XSD to Proto Converter

A sophisticated tool that converts DDEX XSD schemas to Protocol Buffer definitions with `@gotags:` XML annotations for native XML support.

This tool is the core of the DDEX Go generation pipeline, enabling native XML marshal/unmarshal while maintaining high-performance Protocol Buffer serialization. It handles complex XSD features including namespace-aware imports, schema dependencies, and DDEX-specific patterns.

## XSD Primitive to Proto Mappings

Based on analysis of all DDEX schemas in `xsd/`, here are the XSD patterns we need to handle:

### 1. Basic XSD Types → Proto Types

| XSD Type | Proto Type | Example | Notes |
|----------|------------|---------|-------|
| `xs:string` | `string` | `<xs:element name="MessageId" type="xs:string"/>` | Most common type |
| `xs:integer` | `int32` | `<xs:element name="BitsPerSample" type="xs:integer"/>` | Whole numbers |
| `xs:boolean` | `bool` | `<xs:element name="IsProvidedInDelivery" type="xs:boolean"/>` | true/false |
| `xs:decimal` | `string` | `<xs:element name="Value" type="xs:decimal"/>` | Preserve precision |
| `xs:duration` | `string` | `<xs:element name="Duration" type="xs:duration"/>` | ISO 8601 duration (PT30S) |
| `xs:dateTime` | `string` | `<xs:element name="DateTime" type="xs:dateTime"/>` | ISO 8601 datetime |
| `xs:date` | `string` | `<xs:element name="Date" type="xs:date"/>` | ISO 8601 date (YYYY-MM-DD) |
| `xs:anyURI` | `string` | `<xs:element name="URL" type="xs:anyURI"/>` | URI/URL as string |

### 2. XSD Structure → Proto Structure

#### A. Complex Type with Sequence
```xml
<xs:complexType name="MessageHeader">
  <xs:sequence>
    <xs:element name="MessageId" type="xs:string"/>
    <xs:element name="MessageCreatedDateTime" type="xs:dateTime"/>
  </xs:sequence>
</xs:complexType>
```
**→ Proto:**
```protobuf
message MessageHeader {
  // @gotags: xml:"MessageId"
  string message_id = 1;
  // @gotags: xml:"MessageCreatedDateTime"
  string message_created_date_time = 2;
}
```

#### B. Choice → Oneof
```xml
<xs:choice>
  <xs:element name="ISRC" type="xs:string"/>
  <xs:element name="ISWC" type="xs:string"/>
  <xs:element name="ProprietaryId" type="ern:ProprietaryId"/>
</xs:choice>
```
**→ Proto:**
```protobuf
oneof identifier {
  // @gotags: xml:"ISRC"
  string isrc = 1;
  // @gotags: xml:"ISWC"
  string iswc = 2;
  // @gotags: xml:"ProprietaryId"
  ProprietaryId proprietary_id = 3;
}
```

#### C. Cardinality → Repeated
```xml
<!-- maxOccurs="unbounded" becomes repeated -->
<xs:element name="ReleaseAdmin" minOccurs="0" maxOccurs="unbounded" type="ern:ReleaseAdmin"/>
```
**→ Proto:**
```protobuf
// @gotags: xml:"ReleaseAdmin"
repeated ReleaseAdmin release_admin = 1;
```

#### D. Attributes → Fields with attr tag
```xml
<xs:complexType name="SomeType">
  <xs:simpleContent>
    <xs:extension base="xs:string">
      <xs:attribute name="IsDefault" type="xs:boolean" use="required"/>
    </xs:extension>
  </xs:simpleContent>
</xs:complexType>
```
**→ Proto:**
```protobuf
message SomeType {
  // @gotags: xml:",chardata"
  string value = 1;
  // @gotags: xml:"IsDefault,attr"
  bool is_default = 2;
}
```

#### E. Simple Elements
```xml
<xs:element name="URL" minOccurs="0" maxOccurs="unbounded" type="xs:string"/>
```
**→ Proto:**
```protobuf
// @gotags: xml:"URL"
repeated string url = 1;
```

### 3. Cardinality Rules

| XSD | Proto | Meaning |
|-----|-------|---------|
| `minOccurs="0"` | Optional field | Field can be omitted |
| `maxOccurs="unbounded"` | `repeated` | Array/slice of values |
| `minOccurs="1" maxOccurs="1"` | Regular field | Required field (default) |
| `use="required"` (attributes) | Regular field | Required attribute |
| `use="optional"` (attributes) | Optional field | Optional attribute |

### 4. Custom Types → Message References

```xml
<xs:element name="MessageHeader" type="ern:MessageHeader"/>
```
**→ Proto:**
```protobuf
// @gotags: xml:"MessageHeader"
MessageHeader message_header = 1;
```

The namespace prefix (`ern:`) is stripped and the type becomes a message reference.

## Schema Statistics

- **ERN v4.3.2**: 206 complex types, 374 unbounded elements, 91 simpleContent types
- **MEAD v1.1**: 53 simpleContent types  
- **PIE v1.0**: 42 simpleContent types

## Implementation Priority

1. **complexType with sequence** - Most common pattern
2. **Basic type mapping** - Handle xs:string, xs:integer, etc.
3. **Cardinality** - Handle minOccurs/maxOccurs  
4. **Attributes** - Handle simpleContent with attributes
5. **Choice** - Handle oneof patterns
6. **Custom types** - Reference other messages

## Architecture

### Supported DDEX Specifications

The tool processes these DDEX specifications (defined in `specs` array):

**Current Support**: 5 ERN versions, 2 AVS versions, MEAD v1.1, and PIE v1.0

- **AVS (Allowed Value Sets)**
  - `v20200518` - AVS v2020.05.18 (`avs20200518.xsd`)
  - `v20161006` - AVS v2016.10.06 (`avs_20161006.xsd`)
- **ERN (Electronic Release Notification)**
  - `v4.3.2` - ERN v4.3.2 (`release-notification.xsd`)
  - `v4.3` - ERN v4.3 (`release-notification.xsd`)
  - `v4.2` - ERN v4.2 (`release-notification.xsd`)
  - `v3.8.3` - ERN v3.8.3 (`release-notification.xsd`)
  - `v3.8.1` - ERN v3.8.1 (`release-notification.xsd`)
- **MEAD (Media Enrichment and Description)**
  - `v1.1` - MEAD v1.1 (`media-enrichment-and-description.xsd`)
- **PIE (Party Identification and Enrichment)**
  - `v1.0` - PIE v1.0 (`party-identification-and-enrichment.xsd`)

### Processing Pipeline

1. **Schema Graph Loading**: Recursively loads XSD schemas following `xs:import` and `xs:include` dependencies
2. **Namespace Bundling**: Groups schema components by target namespace for proper proto package organization
3. **AVS Version Detection**: Automatically detects which AVS version each schema imports
4. **Proto Generation**: Generates one `.proto` file per namespace with proper imports and `@gotags:` annotations

### Key Features

- **Namespace-aware imports**: Properly handles cross-namespace references between DDEX specifications
- **AVS version context**: Tracks which AVS version each schema uses and generates appropriate imports
- **Deduplication**: Handles repeated field names and prevents duplicate type generation
- **XML compliance**: Generates `@gotags:` comments for `protoc-go-inject-tag` processing
- **Root element handling**: Adds namespace attributes (`xmlns:`, `xsi:`, `schemaLocation`) to root elements

## Output Structure

Generated `.proto` files are organized by namespace:

```
proto/
├── ddex/avs/v20200518/v20200518.proto   # AVS v2020.05.18 enums
├── ddex/avs/v20161006/v20161006.proto   # AVS v2016.10.06 enums
├── ddex/ern/v432/v432.proto             # ERN v4.3.2 messages
├── ddex/ern/v43/v43.proto               # ERN v4.3 messages
├── ddex/ern/v42/v42.proto               # ERN v4.2 messages
├── ddex/ern/v383/v383.proto             # ERN v3.8.3 messages
├── ddex/ern/v381/v381.proto             # ERN v3.8.1 messages
├── ddex/mead/v11/v11.proto              # MEAD v1.1 messages
└── ddex/pie/v10/v10.proto               # PIE v1.0 messages
```

Each `.proto` file includes:
- Package declaration with versioned namespace
- Go package option pointing to `gen/` directory
- Proper imports for cross-namespace dependencies
- Message definitions with `@gotags:` XML annotations
- Enum definitions for simple types with restrictions

## Usage

The tool runs automatically as part of the build process:

```bash
# Generate all proto files from XSD schemas
make generate-proto

# Or run directly
go run tools/xsd2proto/main.go
```

## Implementation Details

### XSD Feature Support

- **Complex Types**: Converted to proto messages with proper field numbering
- **Simple Types with Enumerations**: Converted to proto enums with UNSPECIFIED default
- **Sequences**: Elements become message fields with appropriate cardinality
- **Choices**: Flattened into parent message fields (not oneof for XML compatibility)
- **Attributes**: Become message fields with `xml:",attr"` tags
- **Simple Content**: Base type becomes `value` field with `xml:",chardata"` tag
- **Cardinality**: `maxOccurs="unbounded"` becomes `repeated` fields

### Namespace Handling

- **Target Namespace Detection**: Each XSD schema's `targetNamespace` determines proto package
- **Import Resolution**: Follows `xs:import` declarations to load dependencies
- **Include Resolution**: Follows `xs:include` declarations for same-namespace components
- **AVS Version Mapping**: Detects AVS imports and maps to appropriate versioned packages

### Field Generation

- **Deduplication**: Prevents field name conflicts within messages
- **Repeated Field Optimization**: Merges multiple same-name elements into single repeated field
- **XML Tag Preservation**: Maintains original XML element/attribute names in `@gotags:`
- **Type Mapping**: Maps XSD types to appropriate proto types (see mapping table above)

## Goal

Generate `.proto` files that produce Go structs with native XML marshaling support, providing:
- Complete DDEX XSD compliance for XML serialization
- High-performance Protocol Buffer binary serialization
- Standard JSON serialization support
- Full bidirectional conversion between all three formats

The generated Go structs maintain perfect compatibility with DDEX XML standards while enabling modern microservice architectures through Protocol Buffer support.