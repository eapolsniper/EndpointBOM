# EndpointBOM Usage Guide

Complete guide for using EndpointBOM to scan endpoints and generate SBOMs.

## Table of Contents

- [Quick Start](#quick-start)
- [Command-Line Interface](#command-line-interface)
- [Configuration](#configuration)
- [Understanding the Output](#understanding-the-output)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)

## Quick Start

### First Run

**Recommended:** Run with admin privileges for complete endpoint inventory:

```bash
# macOS/Linux
sudo ./endpointbom

# Windows (run as Administrator)
endpointbom.exe
```

**Or run without admin** - the tool will automatically scan only your user profile:

```bash
# macOS/Linux
./endpointbom

# Windows
endpointbom.exe
```

This will:
1. Scan all package managers
2. Discover all applications
3. Find IDE extensions
4. Generate SBOM files in the scans/ directory

**Smart Behavior:**
- Running as admin → Scans all users (complete inventory)
- Running without admin → Scans current user only (shows warning)

### Understanding the Output

After running, you'll see files in the `scans/` directory:
```
scans/hostname.20240115-143022.package-managers.cdx.json
scans/hostname.20240115-143022.applications.cdx.json
scans/hostname.20240115-143022.ide-extensions.cdx.json
```

Each file is a CycloneDX SBOM in JSON format. The `scans/` directory is automatically created next to the executable and is excluded from git by default.

## Command-Line Interface

### Global Flags

```bash
endpointbom [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | | `./endpointbom.yaml` | Configuration file path (validated) |
| `--output` | | `./scans` | Output directory for SBOMs (next to executable) |
| `--debug` | | `false` | Enable debug output |
| `--verbose` | `-v` | `false` | Enable verbose output |
| `--require-admin` | | `false` | Fail if not running as admin (strict mode) |
| `--scan-all-users` | | `true` | Scan all user profiles (auto-adjusts if not admin) |
| `--exclude` | | `[]` | Paths to exclude (repeatable) |
| `--disable` | | `[]` | Scanners to disable (repeatable) |
| `--help` | `-h` | | Show help |

**Smart Privilege Handling:**
- ✅ **As Admin**: Automatically scans all users on the system
- ⚠️ **Without Admin**: Automatically scans current user only with warning
- Use `--require-admin=true` to force admin requirement (fails if not admin)

**Security Notes:**
- Config and output paths are validated to prevent access to sensitive files
- Built-in exclusions protect sensitive paths (.ssh, .aws, .gnupg, etc.)
- Writing to system directories is blocked

### Examples

#### Basic Scan with Verbose Output

```bash
sudo endpointbom --verbose
```

Output:
```
Gathering system information...
Hostname: macbook-pro
OS: darwin 13.5.2
Logged in users: [john]
Running scanner: npm
  Found 45 components
Running scanner: pip
  Found 128 components
...
```

#### Scan Current User Only

```bash
# Simply run without sudo - automatically scans current user
./endpointbom

# Or explicitly with admin but limit to current user
sudo endpointbom --scan-all-users=false
```

The tool automatically scans only the current user when not running as admin. A clear warning message is displayed to inform you of the limited scope.

#### Disable Specific Scanners

```bash
sudo endpointbom --disable=npm --disable=yarn --disable=vscode
```

Or disable multiple at once:
```bash
sudo endpointbom --disable=npm,yarn,vscode
```

#### Exclude Paths

```bash
sudo endpointbom --exclude=/opt/temp --exclude=/Users/guest
```

#### Custom Output Directory

```bash
sudo endpointbom --output=/var/log/sboms

# Or relative to current directory
sudo endpointbom --output=./my-scans
```

**Security**: The output path is validated. Attempts to write to system directories like `/etc`, `/bin`, `C:\Windows` will be blocked.

#### Debug Mode

```bash
sudo endpointbom --debug
```

Shows detailed error messages and scanner progress.

## Configuration

### Configuration File

Create `endpointbom.yaml` in the working directory or specify with `--config`:

```yaml
# Paths to exclude from scanning
exclude_paths:
  - /Users/test-user
  - /Applications/Testing.app

# Disable specific scanners
disabled_scanners:
  - npm
  - vscode

# Require admin privileges
require_admin: true

# Scan all user profiles
scan_all_users: true

# Output directory
output_dir: "/var/log/sboms"

# Logging
debug: false
verbose: true
```

### Environment-Specific Configs

Create different configs for different environments:

**Development** (`config-dev.yaml`):
```yaml
require_admin: false
scan_all_users: false  # Only scan current user for development
verbose: true
output_dir: "./scans"
disabled_scanners: []
exclude_paths: []  # Uses built-in sensitive path exclusions
```

**Production** (`config-prod.yaml`):
```yaml
require_admin: true
scan_all_users: true  # Scan all users for complete endpoint inventory
verbose: false
output_dir: "/var/log/endpoint-boms"
disabled_scanners: []
# Add additional exclusions beyond built-in ones
exclude_paths:
  - /opt/proprietary-app
  - /home/service-accounts
```

Use with:
```bash
sudo endpointbom --config=config-prod.yaml
```

## Understanding the Output

### SBOM File Structure

Each SBOM file follows CycloneDX 1.5 format:

```json
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "serialNumber": "urn:uuid:...",
  "version": 1,
  "metadata": {
    "timestamp": "2024-01-15T14:30:22Z",
    "component": {
      "type": "device",
      "name": "macbook-pro",
      "version": "13.5.2",
      "properties": [
        {"name": "os", "value": "darwin"},
        {"name": "os_version", "value": "13.5.2"},
        {"name": "logged_in_user", "value": "john"}
      ]
    }
  },
  "components": [
    {
      "type": "library",
      "name": "express",
      "version": "4.18.2",
      "properties": [
        {"name": "package_manager", "value": "npm"},
        {"name": "location", "value": "/usr/local/lib/node_modules/express"}
      ]
    }
  ]
}
```

### File Categories

**Package Managers SBOM** (`*-package-managers.cdx.json`):
- All packages from npm, pip, yarn, etc.
- Includes transitive dependencies where available
- Contains version information

**Applications SBOM** (`*-applications.cdx.json`):
- Non-OS applications installed on the system
- Desktop applications
- User-installed software

**IDE Extensions SBOM** (`*-ide-extensions.cdx.json`):
- VSCode extensions
- JetBrains plugins
- Cursor extensions
- Sublime packages
- MCP servers

## Advanced Usage

### Automated Scanning

#### Daily Cron Job (macOS/Linux)

```bash
# Edit crontab
sudo crontab -e

# Add daily scan at 2 AM
0 2 * * * /usr/local/bin/endpointbom --output=/var/log/sboms > /var/log/endpointbom.log 2>&1
```

#### Windows Task Scheduler

Create a scheduled task:
```powershell
$action = New-ScheduledTaskAction -Execute "C:\Program Files\endpointbom\endpointbom.exe" -Argument "--output=C:\Logs\sboms"
$trigger = New-ScheduledTaskTrigger -Daily -At 2am
Register-ScheduledTask -Action $action -Trigger $trigger -TaskName "EndpointBOM Daily Scan" -RunLevel Highest
```

### Integration with Other Tools

#### Upload to S3

```bash
#!/bin/bash
sudo endpointbom --output=/tmp/sboms
aws s3 cp /tmp/sboms/ s3://my-bucket/sboms/ --recursive
rm -rf /tmp/sboms/*.cdx.json
```

#### Send to SIEM

```bash
#!/bin/bash
sudo endpointbom --output=/tmp/sboms
for file in /tmp/sboms/*.cdx.json; do
    curl -X POST -H "Content-Type: application/json" \
         -d @"$file" \
         https://siem.example.com/api/sbom
done
```

### Jamf Pro Integration

Create a script in Jamf:

```bash
#!/bin/bash

# Run EndpointBOM
/usr/local/bin/endpointbom --output=/var/log/sboms

# Upload to central server
# Add your upload logic here
```

Set as a policy to run daily.

## Troubleshooting

### Common Issues

#### "Permission denied"

**Problem**: Not running with admin privileges

**Solution**:
```bash
# macOS/Linux
sudo endpointbom

# Windows: Run as Administrator
```

#### "No components found"

**Problem**: Package managers not installed or not in PATH

**Solution**:
- Install relevant package managers (npm, pip, etc.)
- Ensure they're in your PATH
- Run with `--debug` to see which scanners are skipping

#### "Config file not found"

**Problem**: Config file doesn't exist

**Solution**:
- Create `endpointbom.yaml` or specify path with `--config`
- Or don't use a config file (defaults will be used)

#### "Output directory permission denied"

**Problem**: Can't write to output directory

**Solution**:
```bash
# Create directory with proper permissions
sudo mkdir -p /var/log/sboms
sudo chmod 755 /var/log/sboms

# Or use a directory you have access to
endpointbom --output=./sboms
```

### Debug Mode

Enable debug mode to see detailed information:

```bash
sudo endpointbom --debug
```

This shows:
- Which scanners are running
- Why scanners are skipped
- Errors encountered during scanning
- Parsing errors

### Verbose Mode

Enable verbose mode for progress information:

```bash
sudo endpointbom --verbose
```

This shows:
- System information
- Scanner progress
- Component counts
- File generation status

## Best Practices

1. **Regular Scans**: Run daily or weekly to track changes
2. **Version Control**: Store SBOM files in version control to track changes over time
3. **Centralized Storage**: Upload SBOMs to a central location for analysis
4. **Access Control**: Protect SBOM files as they contain inventory data
5. **Review Output**: Periodically review generated SBOMs for accuracy
6. **Keep Updated**: Use the latest version of EndpointBOM

## Next Steps

- Review generated SBOM files
- Integrate with your vulnerability scanning tools
- Set up automated scanning
- Import into inventory management systems

## Getting Help

- GitHub Issues: https://github.com/eapolsniper/endpointbom/issues
- Documentation: See README.md
- Community: GitHub Discussions

