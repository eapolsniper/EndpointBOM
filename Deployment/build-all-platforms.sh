#!/bin/bash
# Build script for Dependency-Track Upload Tool
# This script builds binaries for all major platforms

set -e

echo "=========================================="
echo "Dependency-Track Upload Tool - Multi-Platform Build"
echo "=========================================="
echo ""

XOR_KEY=90

xor_encode() {
    local input="$1"
    python3 -c "
import sys
import base64
xor_key = $XOR_KEY
input_str = sys.argv[1]
output = bytes([ord(c) ^ xor_key for c in input_str])
print(base64.b64encode(output).decode())
" "$input"
}

base64_encode() {
    echo -n "$1" | base64
}

# Get URL and API key from arguments or environment variables
DT_URL="${1:-${DT_URL}}"
DT_API_KEY="${2:-${DT_API_KEY}}"

# Check if URL and API key are provided
if [ -z "$DT_URL" ] || [ -z "$DT_API_KEY" ]; then
    echo "⚠️  No URL or API key provided."
    echo "    Building with default values (http://localhost:8081)"
    echo ""
    echo "Usage:"
    echo "  $0 <DEPENDENCY_TRACK_URL> <API_KEY>"
    echo ""
    echo "Example:"
    echo "  $0 https://dtrack.company.com odt_AbCdEfGh123456789"
    echo ""
    echo "Or with environment variables:"
    echo "  DT_URL=https://dtrack.company.com DT_API_KEY=odt_xyz... $0"
    echo ""
    sleep 2
    LDFLAGS=""
else
    echo "🔐 Processing credentials..."
    ENCODED_URL=$(base64_encode "$DT_URL")
    ENCODED_KEY=$(xor_encode "$DT_API_KEY")
    
    echo "   ✅ Ready to build"
    echo ""
    
    LDFLAGS="-X main.encodedURL=$ENCODED_URL -X main.encodedKey=$ENCODED_KEY"
fi

# Create output directory
mkdir -p dist

echo "Building binaries..."
echo ""

# Linux AMD64
echo "📦 Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/dt-upload-linux-amd64 upload-to-dependency-track.go
echo "   ✅ dist/dt-upload-linux-amd64"

# Linux ARM64
echo "📦 Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/dt-upload-linux-arm64 upload-to-dependency-track.go
echo "   ✅ dist/dt-upload-linux-arm64"

# macOS AMD64 (Intel)
echo "📦 Building for macOS (amd64 - Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/dt-upload-macos-amd64 upload-to-dependency-track.go
echo "   ✅ dist/dt-upload-macos-amd64"

# macOS ARM64 (Apple Silicon)
echo "📦 Building for macOS (arm64 - Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/dt-upload-macos-arm64 upload-to-dependency-track.go
echo "   ✅ dist/dt-upload-macos-arm64"

# Windows AMD64
echo "📦 Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/dt-upload-windows-amd64.exe upload-to-dependency-track.go
echo "   ✅ dist/dt-upload-windows-amd64.exe"

# Windows ARM64 (for Surface devices)
echo "📦 Building for Windows (arm64)..."
GOOS=windows GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/dt-upload-windows-arm64.exe upload-to-dependency-track.go
echo "   ✅ dist/dt-upload-windows-arm64.exe"

echo ""
echo "=========================================="
echo "✅ All builds complete!"
echo "=========================================="
echo ""
echo "Binaries are in the dist/ directory:"
ls -lh dist/
echo ""

if [ -n "$DT_URL" ] && [ -n "$DT_API_KEY" ]; then
    echo "🔒 Secrets are embedded in the binaries"
    echo ""
fi

echo "Usage:"
echo "  ./dist/dt-upload-linux-amd64 [scans-directory]"
echo ""
echo "Or override at runtime:"
echo "  DEPENDENCY_TRACK_URL=https://dtrack.company.com DEPENDENCY_TRACK_API_KEY=odt_xyz... ./dist/dt-upload-linux-amd64"
echo ""

