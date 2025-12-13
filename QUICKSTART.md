# EndpointBOM Quick Start Guide

Get up and running with EndpointBOM in 5 minutes!

## Prerequisites

- Go 1.21+ installed ([Download](https://golang.org/dl/))
- Admin/root access to your machine

## Step 1: Install

### From Source

```bash
# Clone the repository
git clone https://github.com/eapolsniper/endpointbom.git
cd endpointbom

# Build
make build

# Or without Make
go build -o bin/endpointbom cmd/endpointbom/main.go
```

### Via Homebrew (macOS)

```bash
# Coming soon
brew install endpointbom
```

## Step 2: Run Your First Scan

**For complete endpoint inventory (recommended):**
```bash
# macOS/Linux - with sudo for all users
sudo ./bin/endpointbom

# Windows - as Administrator for all users
bin\endpointbom.exe
```

**Or run without admin (scans current user only):**
```bash
# macOS/Linux - no sudo needed
./bin/endpointbom

# Windows - no admin needed
bin\endpointbom.exe
```

The tool will automatically adapt and show you what it's scanning!

## Step 3: View Results

After the scan completes, you'll see SBOM files in the current directory:

```
hostname.20240115-143022.package-managers.cdx.json
hostname.20240115-143022.applications.cdx.json
hostname.20240115-143022.ide-extensions.cdx.json
```

These are CycloneDX JSON files containing your endpoint's bill of materials!

## What Gets Scanned?

- ✅ **Package Managers**: npm, pip, yarn, pnpm, brew, gem, cargo, composer, chocolatey, go
- ✅ **Applications**: All installed apps (excluding OS components)
- ✅ **IDE Extensions**: VSCode, Cursor, JetBrains, Sublime
- ✅ **Browser Extensions**: Chrome, Firefox, Edge, Safari (with permissions)
- ✅ **MCP Servers**: Detected in supported IDEs
- ✅ **Network Info**: Local & public IP addresses for endpoint tracking

## Common Use Cases

### Scan Without Admin (Current User Only)

```bash
# Just run it - no flags needed!
./endpointbom

# It will automatically scan current user and show a friendly warning
```

### Verbose Output

```bash
sudo endpointbom --verbose
```

### Save to Specific Directory

```bash
sudo endpointbom --output=/path/to/sboms
```

### Disable Specific Scanners

```bash
sudo endpointbom --disable=npm,vscode
```

## Next Steps

- Read the [full README](README.md) for all features
- Check [USAGE.md](docs/USAGE.md) for detailed usage instructions
- See [BUILD.md](docs/BUILD.md) for building from source
- Review [SECURITY.md](SECURITY.md) for security details

## Troubleshooting

**"Permission denied"**: Run with `sudo` (macOS/Linux) or as Administrator (Windows)

**"go: command not found"**: Install Go from https://golang.org/dl/

**"No components found"**: Install package managers (npm, pip, etc.) and ensure they're in PATH

For more help, see [USAGE.md](docs/USAGE.md) or open an issue on GitHub.

## What's Next?

1. ✅ Generate your first SBOM
2. ⬜ Set up automated daily scans
3. ⬜ Integrate with vulnerability scanning tools
4. ⬜ Import into inventory management

Happy scanning! 🚀

