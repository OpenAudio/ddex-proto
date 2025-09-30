# DDEX Go Library Makefile

.PHONY: all test testdata clean generate-proto generate-proto-go generate fmt buf-lint buf-generate buf-all lint lint-install help

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
	@echo "  generate-ddex  - Run protoc-gen-ddex mega tool (inject tags + extensions)"
	@echo "  buf-lint      - Lint protobuf files with buf"
	@echo "  buf-generate  - Generate Go code from .proto files with buf"
	@echo "  buf-all       - Generate protos from XSD, then Go code from protos"
	@echo ""
	@echo "Tools:"
	@echo "  install-tools - Install DDEX code generation tools locally"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  test          - Run all tests (downloads testdata if needed)"
	@echo "  test-roundtrip - Test XML roundtrip compatibility"
	@echo "  lint          - Run essential quality checks (focuses on dangerous issues)"
	@echo "  lint-install  - Install linting tools"
	@echo "  testdata      - Download DDEX sample files"
	@echo ""
	@echo "Maintenance:"
	@echo "  fmt           - Format all Go code with gofmt"
	@echo "  clean         - Clean generated files and test data"
	@echo "  testdata-refresh - Force re-download test data"

# Generate proto files from XSD
generate-proto:
	@echo "Generating proto files from XSD..."
	go run ./cmd/xsd2proto

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

# Inject XML tags into generated protobuf structs using our custom tool
inject-tags:
	@echo "Injecting tags into generated Go files..."
	@go run ./cmd/protoc-go-inject-tag -input="gen/**/*.pb.go"
	@echo "XML tags injected successfully!"

# Generate Go extensions (enum strings and XML marshaling methods)
generate-go-extensions:
	@echo "Generating enum_strings.go and XML files for Go extensions..."
	@go run ./cmd/ddex-gen ./gen
	@echo "Go extensions generation complete!"

# Alternative: Use the mega tool (does both inject-tags and generate-go-extensions)
generate-ddex:
	@echo "Running protoc-gen-ddex (inject tags + generate extensions)..."
	@go run ./cmd/protoc-gen-ddex ./gen
	@echo "DDEX generation complete!"

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

# Install DDEX code generation tools
install-tools:
	@echo "Installing DDEX code generation tools..."
	go install ./cmd/xsd2proto
	go install ./cmd/protoc-go-inject-tag
	go install ./cmd/ddex-gen
	go install ./cmd/protoc-gen-ddex
	@echo "✓ DDEX tools installed:"
	@echo "  - xsd2proto (XSD to protobuf converter)"
	@echo "  - protoc-go-inject-tag (XML tag injector)"
	@echo "  - ddex-gen (DDEX extensions generator)"
	@echo "  - protoc-gen-ddex (all-in-one tool)"

# Install linting tools used in CI
lint-install:
	@echo "Installing linting tools..."
	go install github.com/client9/misspell/cmd/misspell@latest
	go install github.com/gordonklaus/ineffassign@latest
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@echo "Linting tools installed!"

# Run all linting and quality checks (same as CI)
lint: fmt buf-lint
	@echo "Running quality checks..."
	@echo "Checking gofmt..."
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "❌ The following files are not gofmt'd:"; \
		gofmt -s -l .; \
		echo "Run 'make fmt' to fix formatting issues"; \
		exit 1; \
	fi
	@echo "✅ All files are properly formatted"

	@echo "Running go vet..."
	@if ! go vet ./...; then \
		echo "❌ go vet found issues"; \
		exit 1; \
	fi
	@echo "✅ No go vet issues found"

	@echo "Checking for ineffective assignments (dangerous)..."
	@if ineffassign . | grep -v "gen/" | grep -q "."; then \
		echo "❌ Ineffective assignments found:"; \
		ineffassign . | grep -v "gen/"; \
		echo "These could indicate logic errors"; \
		exit 1; \
	fi
	@echo "✅ No ineffective assignments found"

	@echo "Verifying go.mod is tidy..."
	@go mod tidy
	@if [ "$$(git status --porcelain go.mod go.sum | wc -l)" -gt 0 ]; then \
		echo "❌ go.mod is not tidy"; \
		echo "Run 'go mod tidy' and commit the changes"; \
		git diff go.mod go.sum; \
		exit 1; \
	fi
	@echo "✅ go.mod is tidy"

	@echo "🎉 All quality checks passed!"

# Clean up generated files
clean:
	rm -rf gen/*
	rm -rf proto/*
	rm -rf tmp/
