# EndpointBOM Quick Start Guide

## What is EndpointBOM?

EndpointBOM is a CLI tool that scans developer endpoints (Mac, Windows, Linux) to generate comprehensive Software Bill of Materials (SBOM) files in CycloneDX format.

## What Gets Scanned?

- ✅ **Package Managers**: npm, pip, yarn, pnpm, brew, gem, cargo, composer, chocolatey, go
- ✅ **Applications**: All installed apps (excluding OS components)
- ✅ **IDE Extensions**: VSCode, Cursor, JetBrains, Sublime
- ✅ **Browser Extensions**: Chrome, Firefox, Edge, Safari (with permissions)
- ✅ **MCP Servers**: Detected in supported IDEs
- ✅ **Network Info**: Local & public IP addresses for endpoint tracking

## Quick Start

### 1. Download

```bash
# macOS (ARM)
wget https://github.com/eapolsniper/endpointbom/releases/latest/download/endpointbom-darwin-arm64

# macOS (Intel)
wget https://github.com/eapolsniper/endpointbom/releases/latest/download/endpointbom-darwin-amd64

# Linux
wget https://github.com/eapolsniper/endpointbom/releases/latest/download/endpointbom-linux-amd64

# Windows
# Download from: https://github.com/eapolsniper/endpointbom/releases/latest
```

### 2. Make Executable (macOS/Linux)

```bash
chmod +x endpointbom-*
sudo mv endpointbom-* /usr/local/bin/endpointbom
```

### 3. Run

```bash
# macOS/Linux (as admin to scan all users)
sudo endpointbom

# Windows (as Administrator)
endpointbom.exe
```

SBOM files will be created in the `scans/` directory:
- `{hostname}.{timestamp}.package-managers.cdx.json` - NPM, pip, brew packages
- `{hostname}.{timestamp}.applications.cdx.json` - Installed applications
- `{hostname}.{timestamp}.ide-extensions.cdx.json` - VSCode, Cursor extensions, MCP servers
- `{hostname}.{timestamp}.browser-extensions.cdx.json` - Chrome, Firefox, Edge, Safari extensions

## Common Commands

### Basic Scan

```bash
# Full scan (all scanners)
sudo endpointbom
```

### Verbose Output

```bash
# See what's being scanned
sudo endpointbom --verbose
```

### Custom Output Directory

```bash
# Save to specific location
sudo endpointbom --output=/var/log/endpointbom/scans
```

### Disable Specific Scanners

```bash
# Skip slow scanners
sudo endpointbom --disable=npm,pip,brew

# Skip browser scanning (faster scan)
sudo endpointbom --disable=chrome-extensions,firefox-extensions,edge-extensions,safari-extensions
```

### Scan Current User Only (No Admin)

```bash
# Run without sudo - scans only your profile
endpointbom --scan-all-users=false
```

### Disable Public IP Gathering

```bash
# Don't query external services for public IP
sudo endpointbom --disable-public-ip
```

### Debug Mode

```bash
# Show errors during development
sudo endpointbom --debug
```

## Configuration File

Create `endpointbom.yaml` in the same directory as the binary:

```yaml
# Output directory
output_dir: "./scans"

# Disable specific scanners
disabled_scanners:
  - chrome-extensions  # Disabled by default (requires TCC on macOS)
  - firefox-extensions
  - edge-extensions
  - safari-extensions

# Paths to exclude
exclude_paths:
  - /tmp
  - /Users/*/Library/Caches

# Scan settings
scan_all_users: true
require_admin: false
disable_public_ip: false

# Output settings
verbose: false
debug: false
```

## Output Examples

### Scan Summary

```
=== Scan Summary ===
Package Manager Components: 234
Applications: 87
IDE Extensions/Plugins: 22
Browser Extensions: 0 (disabled by default)
Output Directory: /Users/jsmith/scans

=== Generating SBOMs ===
Generated: developer-laptop.20251213-120000.package-managers.cdx.json
Generated: developer-laptop.20251213-120000.applications.cdx.json
Generated: developer-laptop.20251213-120000.ide-extensions.cdx.json

✓ Scan complete!
```

### SBOM File Structure

```json
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "metadata": {
    "timestamp": "2025-12-13T12:00:00Z",
    "component": {
      "type": "device",
      "name": "developer-macbook",
      "properties": [
        {"name": "os", "value": "darwin"},
        {"name": "logged_in_user", "value": "jsmith"},
        {"name": "local_ip", "value": "192.168.1.100"},
        {"name": "public_ip", "value": "203.0.113.45"}
      ]
    }
  },
  "components": [
    {
      "type": "library",
      "name": "react",
      "version": "18.2.0"
    }
  ]
}
```

## Troubleshooting

### No Components Found

**Issue:** Scanner returns 0 components

**Solutions:**
- Check if package manager is installed (`which npm`, `which pip`)
- Run with `--debug` to see errors
- Verify you have permissions to access directories

### Permission Denied

**Issue:** Cannot access certain directories

**Solutions:**
- Run with `sudo` (macOS/Linux) or as Administrator (Windows)
- Use `--scan-all-users=false` to scan only current user
- Add paths to exclude list if needed

### macOS Permission Popups

**Issue:** Popup asking for permission to access Chrome folder

**Solutions:**
- Browser scanners are disabled by default (no popups)
- To enable: See `DeploymentDocs/ENABLING_BROWSER_SCANNING.md`
- For Jamf deployment: See `DeploymentDocs/JAMF_DEPLOYMENT.md`

## Next Steps

- **Full Documentation**: See `docs/USAGE.md`
- **Deployment Guide**: See `DeploymentDocs/README.md`
- **Security Features**: See `docs/SECURITY_IMPROVEMENTS.md`
- **Browser Extensions**: See `docs/BROWSER_EXTENSIONS.md`

---

**Copyright © Tim Jensen (EapolSniper)**  
**GitHub**: https://github.com/eapolsniper/endpointbom

