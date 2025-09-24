# Contributing to DDEX Go

Thank you for your interest in contributing to DDEX Go! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Process](#contributing-process)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Requirements](#testing-requirements)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)

## Code of Conduct

This project adheres to a code of conduct that ensures a welcoming environment for all contributors. By participating, you agree to abide by these standards:

- Be respectful and inclusive
- Focus on constructive feedback
- Assume good intentions
- Report unacceptable behavior to the maintainers

## Getting Started

### Prerequisites

- Go 1.25.0 or later
- Git
- Make
- `buf` CLI for Protocol Buffer management
- `protoc-go-inject-tag` for XML tag injection

### Installation

```bash
# Clone the repository
git clone https://github.com/OpenAudio/ddex-proto.git
cd ddex-proto

# Install dependencies
go mod download

# Install required tools
go install github.com/bufbuild/buf/cmd/buf@latest
go install github.com/favadi/protoc-go-inject-tag@latest

# Run tests to verify setup
make test
```

## Development Setup

### Repository Structure

```
ddex-go/
├── gen/                     # Generated Go code from protobuf
├── proto/                   # Protocol Buffer definitions
├── tools/                   # Code generation tools
│   ├── xsd2proto/          # XSD to protobuf converter
│   └── generate-go-extensions/ # Go extensions generator
├── testdata/               # Test data (official DDEX samples)
├── xsd/                    # Original DDEX XSD schemas
├── examples/               # Usage examples
└── docs/                   # Additional documentation
```

### Build System

The project uses a sophisticated build system:

```bash
# Generate everything from XSD schemas
make generate

# Individual generation steps
make generate-proto        # XSD → .proto files
make generate-proto-go     # .proto → Go structs with XML tags

# Testing
make test                  # All tests
make test-comprehensive    # Conformance + roundtrip + completeness
make benchmark            # Performance benchmarks

# Maintenance
make clean                # Clean generated files
make help                 # Show all available targets
```

## Contributing Process

### 1. Issues and Discussions

- **Bug Reports**: Use the issue template and include reproduction steps
- **Feature Requests**: Describe the use case and expected behavior
- **Questions**: Start with GitHub Discussions for general questions
- **DDEX Standard Changes**: Reference official DDEX documentation

### 2. Development Workflow

1. **Fork and Clone**
   ```bash
   git fork https://github.com/OpenAudio/ddex-proto.git
   git clone https://github.com/yourusername/ddex-proto.git
   cd ddex-proto
   git remote add upstream https://github.com/OpenAudio/ddex-proto.git
   ```

2. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/issue-number
   ```

3. **Make Changes**
   - Follow code style guidelines
   - Add tests for new functionality
   - Update documentation as needed

4. **Test Your Changes**
   ```bash
   make test                    # Run all tests
   make test-comprehensive     # Run comprehensive validation
   make benchmark             # Check performance impact
   go mod tidy                # Clean up dependencies
   ```

5. **Commit and Push**
   ```bash
   git add .
   git commit -m "feat: add support for ERN v4.4"  # Follow conventional commits
   git push origin feature/your-feature-name
   ```

## Code Style Guidelines

### Go Code Style

- Follow standard Go conventions (`gofmt`, `golint`, `go vet`)
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Prefer composition over inheritance
- Handle errors explicitly

### Protocol Buffer Style

- Use PascalCase for message names: `NewReleaseMessage`
- Use snake_case for field names: `message_header`
- Add `@gotags:` comments for XML marshaling
- Include field numbers sequentially starting from 1

### Generated Code

- **Never edit generated files directly**
- Make changes to XSD schemas or generation tools
- Regenerate using `make generate`
- Commit both source changes and generated output

## Testing Requirements

### Test Categories

All contributions must include appropriate tests:

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **Conformance Tests**: Validate against official DDEX samples
4. **Roundtrip Tests**: Ensure XML ↔ protobuf ↔ JSON integrity

### Test Guidelines

- **100% test coverage** for new functionality
- **Zero data loss tolerance** - any missing XML elements/attributes = test failure
- **Use official DDEX samples** when available
- **Benchmark performance-critical code**
- **Test edge cases and error conditions**

### Running Tests

```bash
# Full test suite
make test

# Specific test categories
go test -v -run TestDDEXConformance ./...
go test -v -run TestXMLRoundTripIntegrity ./...
go test -v -run TestFieldCompleteness ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Documentation

### Required Documentation

- **Code Comments**: All exported functions must have documentation comments
- **README Updates**: Update examples and feature lists
- **TESTING.md**: Document new test categories or validation approaches
- **Tool Documentation**: Update tool READMEs if you modify generation logic

### Documentation Style

- Use clear, concise language
- Include code examples for new features
- Reference official DDEX documentation where appropriate
- Update table of contents for new sections

## Submitting Changes

### Pull Request Guidelines

1. **Clear Title and Description**
   - Use conventional commit format: `feat:`, `fix:`, `docs:`, etc.
   - Describe what changes and why
   - Reference related issues

2. **PR Checklist**
   - [ ] Tests pass (`make test`)
   - [ ] Code follows style guidelines
   - [ ] Documentation updated
   - [ ] No breaking changes (or clearly documented)
   - [ ] Generated code included if applicable

3. **PR Description Template**
   ```markdown
   ## Summary
   Brief description of changes

   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update

   ## Testing
   - [ ] Added tests for new functionality
   - [ ] All existing tests pass
   - [ ] Benchmarks show acceptable performance

   ## DDEX Compliance
   - [ ] Changes maintain DDEX standard compliance
   - [ ] XML roundtrip integrity preserved
   - [ ] Field completeness validated
   ```

### Review Process

1. **Automated Checks**: CI/CD pipeline runs tests and quality checks
2. **Code Review**: Maintainers review for correctness and style
3. **DDEX Validation**: Ensure changes maintain standard compliance
4. **Documentation Review**: Verify documentation completeness

### Merge Requirements

- All CI checks pass
- At least one maintainer approval
- No unresolved review comments
- Up-to-date with main branch

## Special Considerations

### DDEX Standard Compliance

- **Preserve XML structure**: Changes must maintain official DDEX XML format
- **Schema compatibility**: New versions should be backwards compatible
- **Official samples**: Test against real DDEX consortium samples
- **Data integrity**: Zero tolerance for data loss during conversions

### Breaking Changes

Breaking changes require:
- Major version bump
- Migration guide
- Deprecation notices for removed features
- Extended review period

### Performance

- Benchmark performance-critical changes
- Maintain or improve parsing/marshaling speed
- Consider memory allocation patterns
- Document performance implications

## Getting Help

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and community support
- **Documentation**: Check README.md, TESTING.md, and tool READMEs
- **DDEX Official**: Reference [ddex.net](https://ddex.net) for standard questions

## Recognition

Contributors will be:
- Listed in repository contributors
- Mentioned in release notes for significant contributions
- Credited in documentation for major features

Thank you for contributing to DDEX Go! Your efforts help improve metadata exchange across the music industry.