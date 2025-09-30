# protoc-go-inject-tag

Forked from [github.com/favadi/protoc-go-inject-tag](https://github.com/favadi/protoc-go-inject-tag) (MIT License)

This fork is maintained as part of the [ddex-proto](https://github.com/OpenAudio/ddex-proto) project.

## Changes from Original

- **Exported API**: All main functions and types are now exported (capitalized) for use as a library
- **Library-first design**: Can be imported and used programmatically by other tools
- **Active maintenance**: Kept up-to-date with latest Go versions and protobuf standards

## Usage

### As a CLI Tool

```bash
go install github.com/OpenAudio/ddex-proto/cmd/protoc-go-inject-tag@latest
protoc-go-inject-tag -input="*.pb.go"
```

### As a Library

```go
import "github.com/OpenAudio/ddex-proto/cmd/protoc-go-inject-tag"

// Parse file and find injection points
areas, err := injecttag.ParseFile(filePath, nil, nil)
if err != nil {
    return err
}

// Write modified file
err = injecttag.WriteFile(filePath, areas, false)
```

## Original License

See [LICENSE](./LICENSE) file for MIT License from original project.

## Attribution

Original work by [@favadi](https://github.com/favadi) and contributors.
Maintained fork by the OpenAudio/ddex-proto team.