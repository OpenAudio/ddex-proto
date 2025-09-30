# ddex-gen

DDEX code generator for protobuf-generated files.

## What It Does

Generates DDEX-specific Go code extensions for `.pb.go` files:

1. **enum_strings.go** - String conversion methods for enums (`XMLString()`, parsers)
2. ***.xml.go** - XML marshaling methods with namespace support (`MarshalXML`, `UnmarshalXML`)
3. **registry.go** - Dynamic message type registry for auto-detection

## Installation

```bash
go install github.com/OpenAudio/ddex-proto/cmd/ddex-gen@latest
```

## Usage

After generating `.pb.go` files with `buf generate`:

```bash
# Default: processes ./gen directory
ddex-gen

# Specify custom directory
ddex-gen ./my-gen-dir

# Verbose mode
ddex-gen -verbose ./gen
```

## Example Workflow

```bash
# 1. Generate protobuf Go code
buf generate

# 2. Generate DDEX extensions
ddex-gen

# Now your code has:
# - gen/ddex/ern/v432/enum_strings.go
# - gen/ddex/ern/v432/v432.xml.go
# - gen/registry.go
```

## As a Library

```go
import "github.com/OpenAudio/ddex-proto/pkg/ddexgen"

err := ddexgen.Generate("./gen", true) // dir, verbose
if err != nil {
    log.Fatal(err)
}
```

## See Also

- **protoc-go-inject-tag** - Inject XML struct tags (run this first)
- **protoc-gen-ddex** - All-in-one tool (does both)