# protoc-go-inject-tag

Inject custom struct tags into protobuf-generated Go files.

Forked from [github.com/favadi/protoc-go-inject-tag](https://github.com/favadi/protoc-go-inject-tag) (MIT License) and maintained as part of the [ddex-proto](https://github.com/OpenAudio/ddex-proto) project.

## What It Does

Reads `// @gotags:` comments in `.proto` files and injects them as struct tags in the generated `.pb.go` files.

**Example:**

```protobuf
message Release {
  // @gotags: xml:"ReleaseId"
  string release_id = 1;

  // @gotags: xml:"Title"
  string title = 2;
}
```

Generates:

```go
type Release struct {
    ReleaseId string `protobuf:"..." json:"..." xml:"ReleaseId"`
    Title     string `protobuf:"..." json:"..." xml:"Title"`
}
```

## Installation

```bash
go install github.com/OpenAudio/ddex-proto/cmd/protoc-go-inject-tag@latest
```

## Usage

### CLI

```bash
# Process all .pb.go files
protoc-go-inject-tag -input="gen/**/*.pb.go"

# Single file
protoc-go-inject-tag -input="gen/ddex/ern/v432/v432.pb.go"

# Verbose mode
protoc-go-inject-tag -input="*.pb.go" -verbose
```

### As a Library

```go
import "github.com/OpenAudio/ddex-proto/pkg/injecttag"

// Parse file and find injection points
src, _ := os.ReadFile("file.pb.go")
areas, err := injecttag.ParseFile("file.pb.go", src, nil)
if err != nil {
    return err
}

// Write modified file
err = injecttag.WriteFile("file.pb.go", areas, false)
```

## Changes from Original

- **Exported API**: Functions and types are now exported for library use
- **Library-first design**: Can be imported by other tools
- **Active maintenance**: Updated for latest Go/protobuf versions

## See Also

- **ddex-gen** - Generate DDEX extensions (run this after)
- **protoc-gen-ddex** - All-in-one tool (does both)

## Attribution

Original work by [@favadi](https://github.com/favadi) and contributors.
Maintained fork by the OpenAudio/ddex-proto team.

See [LICENSE](./LICENSE) for MIT License.