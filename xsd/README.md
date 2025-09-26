# DDEX XSD Schema Files

This directory contains the official DDEX XSD schema files used for code generation.

## Schema Sources

All schemas are downloaded directly from the DDEX consortium:

### ERN (Electronic Release Notification) Multiple Versions
- **ERN v4.3.2**: https://service.ddex.net/xml/ern/432/ - Downloaded 2025-01-15
- **ERN v4.3**: https://service.ddex.net/xml/ern/43/ - Downloaded 2025-01-15
- **ERN v4.2**: https://service.ddex.net/xml/ern/42/ - Downloaded 2025-01-15
- **ERN v3.8.3**: https://service.ddex.net/xml/ern/383/ - Downloaded 2025-01-15
- **ERN v3.8.1**: https://service.ddex.net/xml/ern/381/ - Downloaded 2025-01-15
- **Main Schema**: release-notification.xsd (all versions)

### MEAD (Media Enrichment and Description) v1.1  
- **Source**: https://service.ddex.net/xml/mead/11/
- **Main Schema**: media-enrichment-and-description.xsd
- **Downloaded**: 2025-01-15

### PIE (Party Identification and Enrichment) v1.0
- **Source**: https://service.ddex.net/xml/pie/10/
- **Main Schema**: party-identification-and-enrichment.xsd
- **Downloaded**: 2025-01-15

## Directory Structure

```
xsd/
├── avs20200518.xsd             # AVS v2020.05.18 (standalone)
├── avs_20161006.xsd            # AVS v2016.10.06 (standalone)
├── ernv432/                    # ERN v4.3.2 schemas
├── ernv43/                     # ERN v4.3 schemas
├── ernv42/                     # ERN v4.2 schemas
├── ernv383/                    # ERN v3.8.3 schemas
├── ernv381/                    # ERN v3.8.1 schemas
│   ├── release-notification.xsd
│   ├── avs.xsd                 # Allowed Value Sets
│   ├── ddex.xsd               # Common DDEX types
│   └── ... (other dependencies)
│
├── meadv11/                    # MEAD v1.1 schemas
│   ├── media-enrichment-and-description.xsd
│   └── ... (dependencies)
│
└── piev10/                     # PIE v1.0 schemas
    ├── party-identification-and-enrichment.xsd
    └── ... (dependencies)
```

## Schema Processing

- **Filename normalization**: Hyphens converted to underscores for Go compatibility
- **Local references**: Schema location attributes updated to reference local files
- **No modifications**: Schemas are kept as close to original as possible

## Updating Schemas

To update to newer versions:

1. Download new schemas from DDEX service URLs
2. Create version directories (e.g., `ernv44/` for ERN v4.4)
3. Update this README with new sources and dates
4. Update `tools/xsd2proto/main.go` specs array for new versions
5. Add test data to `testdata/ddex/{type}/{version}/`
6. Regenerate code with `make generate`

**Note**: The build system automatically discovers new versions and generates appropriate Protocol Buffer definitions and Go code.

## License

These XSD files are property of the DDEX consortium. See https://ddex.net for license terms.