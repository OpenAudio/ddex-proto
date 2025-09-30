# pkg/injecttag

Struct tag injection library for Go source files.

Forked from [github.com/favadi/protoc-go-inject-tag](https://github.com/favadi/protoc-go-inject-tag) (MIT License) with exported API for library use.

## Purpose

This package provides programmatic access to struct tag injection. It reads special comments in Go files and injects them as struct tags.

## Usage

```go
import "github.com/OpenAudio/ddex-proto/pkg/injecttag"

func main() {
    // Read the generated .pb.go file
    src, err := os.ReadFile("generated.pb.go")
    if err != nil {
        log.Fatal(err)
    }

    // Parse file and find injection points
    areas, err := injecttag.ParseFile("generated.pb.go", src, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Write modified file with injected tags
    err = injecttag.WriteFile("generated.pb.go", areas, false)
    if err != nil {
        log.Fatal(err)
    }
}
```

## API

**Main Functions:**
- `ParseFile(inputPath string, src interface{}, xxxSkip []string) ([]TextArea, error)`
- `WriteFile(inputPath string, areas []TextArea, removeTagComment bool) error`
- `Logf(format string, v ...interface{})`

**Types:**
- `TextArea` - Represents an injection point
- `Verbose bool` - Controls verbose logging

## See Also

- **cmd/protoc-go-inject-tag** - CLI wrapper
- **pkg/ddexgen** - DDEX code generation

## Attribution

Original work by [@favadi](https://github.com/favadi) and contributors.  
Maintained fork by the OpenAudio/ddex-proto team.

See [LICENSE](./LICENSE) for MIT License.
