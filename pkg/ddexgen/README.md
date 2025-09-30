# pkg/ddexgen

DDEX code generation library for protobuf-generated Go files.

## Purpose

This package provides programmatic access to DDEX-specific code generation. It's used by the `ddex-gen` and `protoc-gen-ddex` CLI tools, but can also be imported directly.

## What It Generates

1. **enum_strings.go** - String conversion methods for enums
2. ***.xml.go** - XML marshaling/unmarshaling with namespace support
3. **registry.go** - Dynamic message type registry

## Usage

```go
import "github.com/OpenAudio/ddex-proto/pkg/ddexgen"

func main() {
    // Generate DDEX extensions for all .pb.go files in ./gen
    err := ddexgen.Generate("./gen", true) // directory, verbose
    if err != nil {
        log.Fatal(err)
    }
}
```

## Features

- **Automatic detection** - Scans for `.pb.go` files and processes them
- **Enum string methods** - Generates `XMLString()` and `Parse*String()` functions
- **XML marshaling** - Adds proper namespace handling for DDEX compliance
- **Registry** - Dynamic message type detection from XML

## See Also

- **cmd/ddex-gen** - CLI wrapper for this package
- **pkg/injecttag** - XML tag injection (run this first)