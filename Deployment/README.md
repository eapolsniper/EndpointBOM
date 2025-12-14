# Dependency-Track Upload Tools

This directory contains tools to upload EndpointBOM SBOM files to [Dependency-Track](https://dependencytrack.org/).

## Available Tools

### 1. Python Script (Original)
**File**: `UploadToDependencyTrack.py`

- **Pros**: Feature-rich, easy to modify
- **Cons**: Requires Python 3 installation and dependencies
- **Best for**: Development, testing, one-off uploads

**Requirements**:
```bash
pip install requests python-dateutil
```

**Usage**:
```bash
# Edit the script to set your URL and API key
vim UploadToDependencyTrack.py

# Run from the project root (expects scans/ directory)
python3 Deployment/UploadToDependencyTrack.py
```

---

### 2. Go Standalone Binary (Recommended for Enterprise)
**File**: `upload-to-dependency-track.go`

- **Pros**: 
  - No dependencies required
  - Single binary for each platform
  - Secrets can be embedded at build time
  - Works on Windows, Linux, macOS
- **Cons**: Requires Go to compile (but end users don't need Go)
- **Best for**: Enterprise deployment, distributing to developers

## Building the Go Binary

### Option A: Build with Defaults (Not Recommended)

Build without embedded secrets (uses defaults from code):

```bash
cd Deployment

# For your current platform
go build -o dt-upload upload-to-dependency-track.go
```

This will use the default URL (`http://localhost:8081`) and API key. Only useful for local testing.

### Option B: Build with Embedded Secrets (Recommended)

```bash
cd Deployment

# Build with your secrets
./build-all-platforms.sh https://dtrack.company.com odt_your_api_key_here
```

Or use environment variables:

```bash
DT_URL=https://dtrack.company.com \
DT_API_KEY=odt_your_api_key_here \
./build-all-platforms.sh
```

Binaries will be created in `dist/`:
- `dt-upload-linux-amd64` - Linux (x86_64)
- `dt-upload-linux-arm64` - Linux (ARM64)
- `dt-upload-macos-amd64` - macOS Intel
- `dt-upload-macos-arm64` - macOS Apple Silicon
- `dt-upload-windows-amd64.exe` - Windows (x86_64)
- `dt-upload-windows-arm64.exe` - Windows (ARM64)

## Usage

### Quick Start

```bash
# From your EndpointBOM project root (where scans/ directory is)
./Deployment/dist/dt-upload-linux-amd64

# Or specify a custom scans directory
./Deployment/dist/dt-upload-linux-amd64 /path/to/scans
```

## How the Tool Works

### Project Hierarchy

The tool creates a hierarchical structure in Dependency-Track:

```
📦 developer-laptop.local (DEVICE) - Parent
   ├── 📦 developer-laptop.local - package-managers (LIBRARY)
   ├── 📦 developer-laptop.local - applications (APPLICATION)
   ├── 📦 developer-laptop.local - ide-extensions (LIBRARY)
   └── 📦 developer-laptop.local - browser-extensions (LIBRARY)
```

### Versioning

- **Parent Project**: Uses version `latest` (always the same parent)
- **Child Projects**: Use timestamp-based versions (e.g., `2025-12-13-1654`) for tracking changes over time

### Features

1. **Automatic Deduplication**: Won't create duplicate projects
2. **Progress Monitoring**: Tracks BOM processing status
3. **Metadata Extraction**: Includes OS, user, IPs in project properties
4. **Multi-Platform Support**: Single binary works on Windows, Linux, macOS
5. **Embedded Secrets**: API keys and URLs can be embedded at build time

## Security Considerations

### Security Considerations

**Important**: Embedded secrets are not encrypted. For true security:

- Restrict upload binary distribution to trusted systems/users.


## Troubleshooting

### "No SBOM files found"

Make sure you're running the tool from your EndpointBOM project root, or specify the scans directory:

```bash
./dt-upload /full/path/to/scans
```

### "Failed to create project: 401"

Your API key is invalid or expired. Check:
1. API key in Dependency-Track (Settings → Access Management → Teams)
2. The key has proper permissions (BOM_UPLOAD, PROJECT_CREATION_UPLOAD)

### "Connection refused"

Dependency-Track is not running or URL is incorrect. Verify:
```bash
curl http://localhost:8081/api/version
```

### Binary won't run on macOS

macOS Gatekeeper may block unsigned binaries:
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine dt-upload-macos-arm64

# Or run with explicit permission
spctl --add dt-upload-macos-arm64
```

## Development

### Modifying the Go Tool

Key areas in the source code:

- **API calls**: `DependencyTrackClient` methods
- **Metadata extraction**: `ExtractMetadata()` function

After making changes:
```bash
go build -o dt-upload upload-to-dependency-track.go
```

### Testing

```bash
# Test with default values (localhost)
go run upload-to-dependency-track.go
```

## License

Same as EndpointBOM project. See main repository LICENSE file.

## Support

For issues related to:
- **EndpointBOM**: Open an issue in the main repository
- **Dependency-Track**: See https://docs.dependencytrack.org/

