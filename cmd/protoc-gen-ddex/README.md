# protoc-gen-ddex

**All-in-one** post-processor for DDEX protobuf-generated Go code.

Combines XML tag injection + DDEX code generation in a single command.

## What It Does

1. **Injects XML struct tags** - Reads `@gotags` comments and adds them to `.pb.go` files
2. **Generates enum strings** - `XMLString()` methods and parsers for enums
3. **Generates XML methods** - `MarshalXML`/`UnmarshalXML` with namespace support
4. **Generates registry** - Dynamic message type detection and parsing

## Installation

```bash
go install github.com/OpenAudio/ddex-proto/cmd/protoc-gen-ddex@latest
```

## Usage

```bash
# Default: processes ./gen directory
protoc-gen-ddex

# Specify custom directory
protoc-gen-ddex ./my-gen-dir

# Verbose mode
protoc-gen-ddex -verbose

# Show version
protoc-gen-ddex -version
```

## Complete Workflow

```bash
# 1. Generate .proto files from buf.build/openaudio/ddex
buf generate

# 2. Post-process (adds XML support)
protoc-gen-ddex

# Done! Your code now has full DDEX XML support
```

## What Gets Generated

```
gen/
├── ddex/ern/v432/
│   ├── v432.pb.go           # Modified (XML tags injected)
│   ├── enum_strings.go       # NEW (enum methods)
│   └── v432.xml.go          # NEW (XML marshaling)
└── registry.go              # NEW (dynamic registry)
```

## Future Features

- DDEX validation rules (e.g., reference resolution in ERN messages)
- Configurable validation options

## Individual Tools

If you need finer control, use the individual tools:

- **protoc-go-inject-tag** - Just inject XML tags
- **ddex-gen** - Just generate DDEX extensions

## For External Users

If you're generating code from `buf.build/openaudio/ddex` in your own repository:

```bash
# In your project
buf generate  # Downloads schemas from buf.build

# Add XML support
go install github.com/OpenAudio/ddex-proto/cmd/protoc-gen-ddex@latest
protoc-gen-ddex ./gen

# Now you have the same XML capabilities as this library!
```