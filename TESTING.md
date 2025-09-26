# DDEX Go Test Suite Documentation

This document explains the comprehensive test suite and what each test category proves about the DDEX Go library's correctness and reliability.

## Test Architecture Overview

The test suite validates three critical aspects of the DDEX Go library:

1. **DDEX Standard Compliance** - Can we correctly parse official DDEX XML files?
2. **XML Roundtrip Integrity** - Do we preserve all data during XML → Proto → XML conversions?
3. **Field Completeness** - Are all required DDEX fields properly mapped and populated?

## Test Categories

### 1. Conformance Tests (`TestDDEXConformance`)

**What it proves**: The library can correctly parse real-world DDEX XML files from the official DDEX consortium.

**Test data**: Official DDEX sample files for each specification:
- **ERN v4.3.2**: Official sample files covering different content types
- **ERN v4.3**: Official sample files with comprehensive content types
  - Audio Album, Video Album, Mixed Media Bundle
  - Simple Audio/Video Singles, Ringtones, DJ Mixes
  - Classical variants and longform musical works
- **ERN v4.2**: Official sample files covering different content types
- **ERN v3.8.3**: Official sample files with catalog list support
- **ERN v3.8.1**: Official sample files with catalog list support
- **MEAD v1.1**: Award metadata examples
- **PIE v1.0**: Party identification and award examples

**What gets validated**:
- XML unmarshaling succeeds without errors
- Core message structure is properly populated
- Required fields contain expected data types
- Message-specific validations (release counts, party counts, awards)

**Critical assertion**: If these tests pass, the library can handle real DDEX data from production systems.

### 2. XML Roundtrip Tests (`TestProtoToXMLRoundTrip`)

**What it proves**: Generated protobuf structs can be marshaled to valid XML and unmarshaled back without data loss.

**Process**:
1. Create a protobuf struct using test constructors
2. Marshal to XML with proper indentation
3. Add XML header for complete DDEX document
4. Unmarshal back to protobuf struct
5. Compare original vs. roundtrip using `reflect.DeepEqual`

**Critical assertion**: Perfect bidirectional conversion between protobuf and XML formats.

### 3. DOM-Level Integrity Tests (`TestXMLRoundTripIntegrity`)

**What it proves**: XML → Proto → XML preserves every element, attribute, and value at the DOM level.

**Advanced validation using `etree` DOM parsing**:
- Counts elements and attributes in original vs. marshaled XML
- Detects missing elements, missing attributes, and value mismatches
- Handles repeated elements and complex nested structures
- Normalizes whitespace and line endings for fair comparison

**Detailed reporting**:
```
Elements: Original=1247, Marshaled=1247
Attributes: Original=89, Marshaled=89
Coverage: 100.0%
```

**Critical assertion**: Zero data loss during XML processing - every piece of information is preserved.

### 4. Field Completeness Tests (`TestFieldCompleteness`)

**What it proves**: Required DDEX fields are properly mapped and populated from real XML data.

**Message-specific validations**:

#### ERN (Electronic Release Notification)
- MessageHeader presence and MessageId population
- MessageSender validation
- ReleaseList contains actual releases
- Release counting across different release types

#### MEAD (Media Enrichment and Description)
- MessageHeader and MessageId validation
- ReleaseInformationList contains release data
- Award metadata completeness

#### PIE (Party Identification and Enrichment)
- MessageHeader validation
- PartyList contains party data
- Award counting across all parties
- Multi-language name support

**Critical assertion**: All business-critical DDEX data is accessible through the Go structs.

### 5. XML Tag Effectiveness Tests (`TestXMLTagsEffectiveness`)

**What it proves**: Generated `@gotags:` XML annotations work correctly for marshaling/unmarshaling.

**Process**:
1. Unmarshal real DDEX XML using generated structs
2. Attempt to marshal back to XML
3. Verify no marshaling errors occur

**Critical assertion**: XML struct tags are properly generated and functional.

### 6. Field Coverage Report (`TestFieldCoverageReport`)

**What it proves**: Comprehensive mapping between XSD schema and generated Go structs.

**Advanced path analysis**:
- Extracts all unique XML paths from original document
- Compares against paths in marshaled output
- Calculates coverage percentage
- Reports uncovered fields for schema completeness analysis

**Example output**:
```
Field Coverage Report:
  Total paths in original: 312
  Paths preserved: 312
  Coverage: 100.0%
```

**Critical assertion**: Complete schema coverage - no XSD fields are unmapped.

## Performance Benchmarks (`BenchmarkDDEX`)

**What it proves**: The library performs efficiently for production use.

**Measured operations**:
- **Parse benchmarks**: XML → protobuf struct unmarshaling speed
- **Marshal benchmarks**: Protobuf → XML marshaling speed
- **Memory benchmarks**: Memory allocation patterns (`-benchmem`)

**All three DDEX specifications** (ERN, MEAD, PIE) are benchmarked separately.

**Critical assertion**: Performance is suitable for high-throughput DDEX processing.

## Test Data

### Official DDEX Samples (High Confidence)
- **ERN**: Official DDEX consortium sample files for all supported versions - complete accuracy guarantee
- Real-world complexity with nested structures, optional fields, and edge cases
- Comprehensive coverage across ERN v3.8.1, v3.8.3, v4.2, v4.3, and v4.3.2

### Created Examples (Representative)
- **MEAD/PIE**: Manually created but representative examples
- Covers core functionality and common use cases
- Based on official DDEX specification patterns

### Automatic Test Discovery
- **Dynamic test generation**: Tests automatically discover all message types and versions from testdata
- **Registry-based parsing**: Uses auto-generated registry for type-safe unmarshaling
- **Universal validation**: Single test framework validates all DDEX specifications

## What the Tests Prove Collectively

### ✅ **DDEX Standard Compliance**
The library correctly implements all three major DDEX specifications (ERN, MEAD, PIE) according to official consortium standards.

### ✅ **Production Readiness**
Real-world DDEX XML files from music industry systems can be processed without data loss or corruption.

### ✅ **Bidirectional Conversion**
Perfect XML ↔ protobuf ↔ JSON conversion maintains complete data integrity.

### ✅ **Schema Completeness**
All XSD schema elements are mapped to Go structs - no business data is lost.

### ✅ **Type Safety**
Strong typing with comprehensive validation ensures data consistency.

### ✅ **Performance**
Efficient processing suitable for high-volume DDEX message handling.

## Running the Tests

```bash
# Run all tests including comprehensive validation
make test

# Run specific test categories
make test-comprehensive  # Conformance + roundtrip + completeness
make test-roundtrip     # XML bidirectional conversion tests
make benchmark          # Performance benchmarks

# Individual test suites
go test -v -run TestDDEX ./...                    # Auto-discovered message type tests
go test -v -run TestXMLRoundTripIntegrity ./...
go test -v -run TestFieldCompleteness ./...
```

## Test Philosophy

The test suite follows a **zero-tolerance approach** to data integrity:

- **Any missing XML element** = test failure
- **Any missing attribute** = test failure
- **Any value mismatch** = test failure
- **Any unmarshaling error** = test failure
- **Automatic test discovery** = comprehensive coverage without manual maintenance

This ensures the library meets the exacting standards required for music industry metadata exchange, where data accuracy is critical for rights management, royalty distribution, and content delivery.

## New Test Architecture Benefits

### Automatic Message Type Discovery
The test framework now automatically discovers all supported message types and versions by scanning the `testdata/ddex/` directory structure. This means:
- New DDEX versions can be added simply by adding test files
- No manual test registration required
- Comprehensive validation across all supported specifications

### Registry-Based Parsing
The new auto-generated registry (`gen/registry.go`) provides:
- Type-safe parsing for all DDEX message types
- Automatic message type detection from XML content
- Unified parsing interface across ERN, MEAD, and PIE specifications

### Unified Test Framework
A single `TestDDEX` function now validates all message types and versions:
- Consistent test methodology across all DDEX specifications
- Automatic test file discovery and filtering
- Simplified maintenance and extension

## Adding New Types and Tests

### Adding New DDEX Versions
1. **Add XSD schemas** to `xsd/` directory (e.g., `xsd/ernv44/`)
2. **Add test files** to `testdata/ddex/{type}/{version}/` (e.g., `testdata/ddex/ern/v44/sample.xml`)
3. **Update specs array** in `tools/xsd2proto/main.go`
4. **Run generation**: `make generate`

### Adding New Test Files
Simply drop XML files into the appropriate `testdata/ddex/{type}/{version}/` directory. The test framework automatically discovers and validates them.

### What You Get Automatically
- Conformance testing against real XML files
- XML roundtrip validation (XML → Proto → XML)
- Field completeness validation
- Performance benchmarking
- Registry-based type-safe parsing