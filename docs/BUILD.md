# Building EndpointBOM

This guide explains how to build EndpointBOM from source.

## Prerequisites

- **Go 1.21 or later**: [Download Go](https://golang.org/dl/)
- **Make** (optional): For using the Makefile
- **Git**: For cloning the repository

## Quick Start

```bash
# Clone the repository
git clone https://github.com/eapolsniper/endpointbom.git
cd endpointbom

# Download dependencies
go mod download

# Build for your platform
make build

# Or build manually
go build -o bin/endpointbom cmd/endpointbom/main.go
```

## Building for Different Platforms

### Using Make

Build for all platforms:
```bash
make build-all
```

This creates binaries for:
- `bin/endpointbom-darwin-amd64` (macOS Intel)
- `bin/endpointbom-darwin-arm64` (macOS Apple Silicon)
- `bin/endpointbom-windows-amd64.exe` (Windows)
- `bin/endpointbom-linux-amd64` (Linux)

### Using Go Directly

#### macOS (Intel)
```bash
GOOS=darwin GOARCH=amd64 go build -o bin/endpointbom-darwin-amd64 cmd/endpointbom/main.go
```

#### macOS (Apple Silicon)
```bash
GOOS=darwin GOARCH=arm64 go build -o bin/endpointbom-darwin-arm64 cmd/endpointbom/main.go
```

#### Windows
```bash
GOOS=windows GOARCH=amd64 go build -o bin/endpointbom-windows-amd64.exe cmd/endpointbom/main.go
```

#### Linux
```bash
GOOS=linux GOARCH=amd64 go build -o bin/endpointbom-linux-amd64 cmd/endpointbom/main.go
```

## Building with GoReleaser

For production releases with proper versioning and signing:

```bash
# Install GoReleaser
brew install goreleaser

# Create a release
goreleaser release --snapshot --clean
```

## Installing Locally

Install to your `$GOPATH/bin`:

```bash
make install
```

Or manually:
```bash
go install ./cmd/endpointbom
```

## Verifying the Build

After building, verify the binary works:

```bash
./bin/endpointbom --help
```

You should see the help output.

## Development Build

For development with faster iteration:

```bash
# Build and run
make run

# Or
go run cmd/endpointbom/main.go --help
```

## Build Options

### Static Binary

To create a fully static binary (useful for distribution):

```bash
CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o bin/endpointbom cmd/endpointbom/main.go
```

### With Version Information

Add version information to the build:

```bash
VERSION=1.0.0
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD)

go build \
  -ldflags "-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}" \
  -o bin/endpointbom \
  cmd/endpointbom/main.go
```

### Optimized for Size

Create a smaller binary:

```bash
go build -ldflags="-s -w" -o bin/endpointbom cmd/endpointbom/main.go
```

Then optionally compress with UPX:
```bash
upx --best --lzma bin/endpointbom
```

## Troubleshooting

### "go: command not found"

Install Go from https://golang.org/dl/

### "cannot find package"

Download dependencies:
```bash
go mod download
go mod tidy
```

### Build Errors

Ensure you're using Go 1.21 or later:
```bash
go version
```

### Permission Denied

Make the binary executable:
```bash
chmod +x bin/endpointbom
```

## CI/CD Integration

### GitHub Actions

Example workflow (`.github/workflows/build.yml`):

```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go mod download
      - run: make build-all
      - uses: actions/upload-artifact@v3
        with:
          name: binaries
          path: bin/
```

### Docker Build

Example Dockerfile:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o endpointbom cmd/endpointbom/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/endpointbom /usr/local/bin/
ENTRYPOINT ["endpointbom"]
```

## Distribution

### Creating a Homebrew Tap

1. Create a GitHub repository: `homebrew-tap`
2. Use GoReleaser to automatically create the formula
3. Users can install with: `brew install eapolsniper/tap/endpointbom`

### Creating a Chocolatey Package

1. Create a `.nuspec` file
2. Package with: `choco pack`
3. Publish to Chocolatey.org

### Creating a Debian Package

Use `nfpm` or `fpm`:

```bash
nfpm pkg --packager deb --target .
```

## Next Steps

- See [USAGE.md](USAGE.md) for usage instructions
- See [CONTRIBUTING.md](../CONTRIBUTING.md) for contributing guidelines
- See [README.md](../README.md) for project overview

