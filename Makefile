# DDEX Go Library Makefile

.PHONY: all test testdata clean generate-proto generate-proto-go generate fmt buf-lint buf-generate buf-all help

# Default target
help:
	@echo "DDEX Go Library - Makefile targets:"
	@echo ""
	@echo "Complete workflows:"
	@echo "  all           - Clean, generate everything, and test (full verification)"
	@echo ""
	@echo "Generation:"
	@echo "  generate-proto - Generate .proto files from XSD (proto/ directory)"
	@echo "  generate-proto-go - Generate Go structs from .proto files (gen/ directory)"
	@echo "  generate       - Generate proto files and Go code"
	@echo "  buf-lint      - Lint protobuf files with buf"
	@echo "  buf-generate  - Generate Go code from .proto files with buf"
	@echo "  buf-all       - Generate protos from XSD, then Go code from protos"
	@echo ""
	@echo "Testing:"
	@echo "  test          - Run all tests (downloads testdata if needed)"
	@echo "  test-roundtrip - Test XML roundtrip compatibility"
	@echo "  testdata      - Download DDEX sample files"
	@echo ""
	@echo "Maintenance:"
	@echo "  fmt           - Format all Go code with gofmt"
	@echo "  clean         - Clean generated files and test data"
	@echo "  testdata-refresh - Force re-download test data"

# Generate proto files from XSD
generate-proto:
	@echo "Generating proto files from XSD..."
	go run tools/xsd2proto/main.go

# Generate Go structs from proto files
generate-proto-go:
	@echo "Generating Go structs from proto files..."
	buf generate
	@echo "Injecting XML tags with protoc-go-inject-tag..."
	@$(MAKE) inject-tags
	@echo "Generating Go extensions (enums and XML methods)..."
	@$(MAKE) generate-go-extensions

# Generate everything
generate: generate-proto generate-proto-go fmt
	@echo "All generation complete!"

# Format all Go code
fmt:
	@echo "Formatting Go code..."
	gofmt -s -w .

# Lint protobuf files with buf
buf-lint:
	@echo "Linting protobuf files..."
	buf lint

# Generate Go code from protobuf files
buf-generate: 
	@echo "Generating Go code from protobuf files..."
	buf generate
	@echo "Injecting XML tags with protoc-go-inject-tag..."
	@$(MAKE) inject-tags
	@echo "Generating Go extensions (enums and XML methods)..."
	@$(MAKE) generate-go-extensions

# Inject XML tags into generated protobuf structs using protoc-go-inject-tag
inject-tags:
	@echo "Injecting tags into generated Go files..."
	@for file in $$(find gen -name "*.pb.go" 2>/dev/null); do \
		echo "  Processing $$file..."; \
		protoc-go-inject-tag -input=$$file 2>/dev/null || true; \
	done
	@echo "XML tags injected successfully!"

# Generate Go extensions (enum strings and XML marshaling methods)
generate-go-extensions:
	@echo "Generating enum_strings.go and XML files for Go extensions..."
	go run tools/generate-go-extensions/main.go
	@echo "Go extensions generation complete!"

# Complete protobuf workflow: XSD -> proto -> Go with XML tags
buf-all: generate-proto buf-lint buf-generate
	@echo "Complete protobuf generation workflow complete!"

# Complete workflow: clean, generate everything, and test
all: clean generate test
	@echo "Full clean → generate → test cycle complete!"

# Run all tests including comprehensive validation
test:
	go test -v -count=1 ./...

# Run comprehensive tests against DDEX samples
test-comprehensive:
	go test -v -run TestConformance ./...
	go test -v -run TestRoundTrip ./...
	go test -v -run TestFieldCompleteness ./...

# Run performance benchmarks
benchmark:
	go test -bench=. -benchmem ./...

# Test roundtrip compatibility between pure Go and proto-generated Go
test-roundtrip:
	go test -v ./test/roundtrip/...

# Clean up generated files
clean:
	rm -rf gen/*
	rm -rf proto/*
	rm -rf tmp/
