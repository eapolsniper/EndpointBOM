# Contributing to EndpointBOM

Thank you for your interest in contributing to EndpointBOM! This document provides guidelines for contributing to the project.

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue with:
- A clear, descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Your environment (OS, Go version, etc.)
- Relevant logs or error messages

### Suggesting Features

We welcome feature suggestions! Please create an issue with:
- A clear description of the feature
- Use case(s) for the feature
- Any relevant examples or mockups

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes**:
   - Follow Go best practices and conventions
   - Add tests for new functionality
   - Update documentation as needed
3. **Ensure your code passes checks**:
   ```bash
   go test ./...
   go vet ./...
   ```
4. **Create a pull request** with:
   - Clear description of changes
   - Reference to related issues
   - Screenshots if applicable

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)

### Getting Started

```bash
# Clone the repository
git clone https://github.com/eapolsniper/endpointbom.git
cd endpointbom

# Download dependencies
go mod download

# Build the project
make build

# Run tests
make test
```

## Code Style

- Follow standard Go formatting (use `gofmt`)
- Write clear, descriptive comments
- Keep functions focused and concise
- Use meaningful variable names

## Adding New Scanners

### Package Manager Scanner

1. Create a new file in `internal/scanners/packagemanagers/`
2. Implement the `Scanner` interface:
   ```go
   type Scanner interface {
       Name() string
       Scan(cfg *config.Config) ([]Component, error)
   }
   ```
3. Add the scanner to the list in `cmd/endpointbom/main.go`
4. Update the README with the new scanner

### IDE Scanner

1. Create a new file in `internal/scanners/ides/`
2. Implement the `Scanner` interface
3. Add support for MCP server detection if applicable
4. Add the scanner to the list in `cmd/endpointbom/main.go`
5. Update the README

## Security Guidelines

- Never collect secrets, API keys, or tokens
- Never transmit data without explicit user configuration
- Use standard library functions when possible
- Minimize external dependencies
- Pin dependency versions

## Testing

- Write unit tests for new functionality
- Ensure tests pass before submitting PR
- Test on multiple operating systems if possible

## Documentation

- Update README.md for user-facing changes
- Add code comments for complex logic
- Update CONTRIBUTING.md if changing contribution process

## Questions?

If you have questions, feel free to:
- Open an issue
- Start a discussion on GitHub Discussions
- Contact the maintainers

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

Thank you for contributing to EndpointBOM! 🎉


