# EndpointBOM

**EndpointBOM** is a comprehensive CLI tool for scanning developer endpoints (Mac and Windows) to capture Bill of Materials (BOM) data. It generates CycloneDX-format SBOM files containing information about installed packages, applications, IDE extensions, and MCP servers. The primary purpose is for use by Information Security teams to import into other tooling for discovering unapproved software or malicious packages.

## Features

### Comprehensive Scanning

- **Package Managers**: Scans all installed packages from popular package managers
  - **Node.js**: npm, yarn, pnpm
  - **Python**: pip (including pip3, python -m pip, python3 -m pip)
  - **Ruby**: gem
  - **Rust**: cargo
  - **PHP**: composer
  - **Go**: go modules
  - **System**: Homebrew (macOS), Chocolatey (Windows)

- **Applications**: Discovers all non-OS applications
  - macOS: `/Applications`, `/System/Applications`, user Applications folders
  - Windows: `Program Files`, `Program Files (x86)`, user-installed apps, Windows Registry
  - Linux: `/usr/share/applications`, user `.local` directories

- **IDE Extensions & Plugins**:
  - **VSCode**: All installed extensions
  - **Cursor**: Extensions and configuration
  - **JetBrains Suite**: IntelliJ IDEA, PyCharm, WebStorm, PhpStorm, GoLand, RubyMine, CLion, Rider, DataGrip, AndroidStudio
  - **Sublime Text**: Packages
  - **Atom**: (framework in place)

- **MCP Servers**: Detects Model Context Protocol servers configured in supported IDEs

- **Browser Extensions**: Scans all major browsers for installed extensions
  - **Chrome**: All extensions with permissions and host access
  - **Firefox**: Add-ons and extensions
  - **Microsoft Edge**: Extensions and web apps
  - **Safari**: Extensions (macOS only)
  - Captures extension IDs, versions, and security-relevant permissions

- **Network Information**: Automatically captures endpoint network details
  - All local IP addresses (IPv4 and IPv6)
  - Public IP address (for endpoint identification)
  - Non-blocking with graceful fallback if unavailable
  - VSCode MCP configurations
  - Cursor MCP configurations
  - Captures command, args count, and env var count (without exposing secrets)

### Security First

- ✅ **No secrets collected**: API keys, tokens, and environment variable values are never captured
- ✅ **Minimal dependencies**: Only 3 external dependencies, all highly trusted
- ✅ **Dependency pinning**: All dependencies pinned to known safe versions
- ✅ **Standard library preferred**: Uses Go standard library whenever possible
- ✅ **No telemetry**: No data is sent anywhere by the tool
- ✅ **Path validation**: Prevents reading sensitive files (passwords, SSH keys, credentials)
- ✅ **Write protection**: Blocks writing to system directories and sensitive locations
- ✅ **Safe defaults**: Excludes sensitive paths by default (.ssh, .aws, .gnupg, etc.)

### Output Format

- **CycloneDX 1.5** JSON format (widely compatible, feature-rich)
- Separate SBOM files by category:
  - `{hostname}.{timestamp}.package-managers.cdx.json`
  - `{hostname}.{timestamp}.applications.cdx.json`
  - `{hostname}.{timestamp}.ide-extensions.cdx.json`
  - `{hostname}.{timestamp}.browser-extensions.cdx.json` (NEW!)
- Includes metadata: hostname, OS version, logged-in users, local IPs, public IP, timestamp

## Installation

### Prerequisites

- Go 1.21 or later

### From Source

```bash
git clone https://github.com/eapolsniper/endpointbom.git
cd endpointbom
make build
```

The binary will be in `bin/endpointbom`.

### Install to System

```bash
make install
```

This installs to `$GOPATH/bin/endpointbom`.

### Build for All Platforms

```bash
make build-all
```

Builds binaries for:
- macOS (Intel and Apple Silicon)
- Windows (x64)
- Linux (x64)

## Usage

### Basic Usage

**Recommended:** Run with admin/root privileges for complete endpoint inventory:

```bash
# macOS/Linux
sudo ./endpointbom

# Windows (run as Administrator)
endpointbom.exe
```

**Without admin privileges:** The tool will automatically scan only the current user and display a warning:

```bash
# macOS/Linux
./endpointbom

# Windows
endpointbom.exe
```

**Output Location**: By default, SBOM files are saved to a `scans/` directory next to the executable. This keeps your scan results organized and prevents them from being accidentally committed to version control.

**Smart Behavior:**
- ✅ **As Admin**: Scans all users on the system (complete endpoint inventory)
- ⚠️ **Without Admin**: Automatically scans only current user with warning message

### Command-Line Options

```bash
endpointbom [flags]

Flags:
  --config string              config file (default is ./endpointbom.yaml)
  --output string              output directory for SBOM files (default: ./scans)
  --debug                      enable debug output
  -v, --verbose                enable verbose output
  --require-admin              require admin/root privileges (fail if not admin, default: false)
  --scan-all-users             scan all user profiles (auto-adjusts if not admin, default: true)
  --exclude strings            paths to exclude from scanning
  --disable strings            scanners to disable (e.g., npm,pip,vscode)
  -h, --help                   help for endpointbom
```

### Examples

#### Scan with verbose output

```bash
sudo endpointbom --verbose
```

#### Scan only current user

```bash
# Just run without sudo - it will automatically scan only current user
./endpointbom

# Or explicitly disable all-users even as admin
sudo endpointbom --scan-all-users=false
```

#### Disable specific scanners

```bash
sudo endpointbom --disable=npm,yarn --disable=vscode
```

#### Specify output directory

```bash
sudo endpointbom --output=/path/to/output

# Or use a custom scans directory
sudo endpointbom --output=./my-scans
```

**Note**: The output directory is validated to prevent writing to sensitive system locations.

#### Use config file

```bash
sudo endpointbom --config=/path/to/config.yaml
```

**Security Note**: Config file paths are validated to prevent reading sensitive system files.

### Configuration File

Create a `endpointbom.yaml` file for persistent configuration:

```yaml
# Paths to exclude from scanning (in addition to built-in sensitive path exclusions)
exclude_paths:
  - /path/to/exclude
  - /opt/sensitive-app

# Scanners to disable
disabled_scanners:
  - npm
  - vscode

# Require admin/root privileges
require_admin: true

# Scan all user profiles (default: true for complete endpoint inventory)
# Set to false to only scan the current user
scan_all_users: true

# Output directory (default: ./scans next to executable)
output_dir: "./scans"

# Enable debug output
debug: false

# Enable verbose output
verbose: false
```

**Built-in Security Exclusions**: The tool automatically excludes sensitive paths:
- SSH keys (`.ssh/`)
- AWS credentials (`.aws/`)
- Kubernetes configs (`.kube/`)
- GPG keys (`.gnupg/`)
- Password files (`/etc/shadow`, etc.)
- Shell history files
- Docker credentials
- And more...

See `configs/example-config.yaml` for a complete example.

## Deployment

For enterprise deployment instructions (Jamf, Intune, GPO), see the **[DeploymentDocs](DeploymentDocs/)** folder.

### Homebrew (macOS)

Once published, the tool can be installed via Homebrew:

```bash
brew install endpointbom
```

*(This requires creating a Homebrew tap - instructions below)*

### Jamf Pro (Enterprise Deployment)

1. **Build the binary** for macOS:
   ```bash
   make build-all
   ```

2. **Create a package** using a tool like `pkgbuild` or Jamf Composer

3. **Upload to Jamf Pro**:
   - Create a new policy
   - Upload the package
   - Set execution frequency (e.g., daily)
   - Configure script to run: `endpointbom --output=/var/log/endpointbom`

4. **Collect results**: Configure the tool to upload SBOMs to a central location (future feature)

### Windows Group Policy / SCCM

Similar deployment pattern for Windows environments using Group Policy or SCCM to distribute and run the tool.

## Dependencies

EndpointBOM uses only 3 external dependencies, all highly trusted:

1. **github.com/CycloneDX/cyclonedx-go** (v0.8.0)
   - Official CycloneDX library
   - Maintained by OWASP CycloneDX team
   - 50+ contributors

2. **github.com/spf13/cobra** (v1.8.0)
   - Industry-standard CLI framework
   - 35k+ stars on GitHub
   - Maintained by Google employees + community

3. **gopkg.in/yaml.v3** (v3.0.1)
   - Standard YAML parser for Go
   - Part of go-yaml organization
   - Widely used in Go ecosystem

All dependencies are pinned to specific versions in `go.mod`.

## Supported Package Managers

| Package Manager | Supported | Transitive Dependencies | Notes |
|----------------|-----------|-------------------------|-------|
| npm            | ✅        | ✅ (via `--all`)        | Global packages only |
| pip            | ✅        | ⚠️ (via `pip show`)     | Lists direct dependencies |
| yarn           | ✅        | ✅                      | Global packages only |
| pnpm           | ✅        | ✅                      | Global packages only |
| Homebrew       | ✅        | ❌                      | Formulae and casks |
| gem            | ✅        | ❌                      | All versions listed |
| cargo          | ✅        | ❌                      | Installed binaries |
| composer       | ✅        | ❌                      | Global packages only |
| chocolatey     | ✅        | ❌                      | Windows only |
| go             | ✅        | ✅                      | Module dependencies |
| poetry         | ⚠️        | ❌                      | Not yet implemented |
| pipenv         | ⚠️        | ❌                      | Not yet implemented |
| conda          | ⚠️        | ❌                      | Not yet implemented |
| maven          | ⚠️        | ❌                      | Not yet implemented |
| gradle         | ⚠️        | ❌                      | Not yet implemented |
| nuget          | ⚠️        | ❌                      | Not yet implemented |

**Legend**:
- ✅ Fully supported
- ⚠️ Planned/partially supported
- ❌ Not applicable or not supported

## Known Limitations

1. **Package Managers Not Yet Supported**:
   - poetry (Python)
   - pipenv (Python)
   - conda (Python)
   - uv (Python)
   - maven (Java)
   - gradle (Java)
   - nuget (.NET)
   - bun (JavaScript)
   
   These will be added in future versions based on community feedback.

2. **Project-Level Dependencies**: The tool scans globally installed packages, not project-specific `package.json`, `requirements.txt`, etc.

3. **Transitive Dependencies**: Not all package managers support querying transitive dependencies. Where possible (npm, pnpm, yarn, go), we capture them.

4. **Windows Registry Parsing**: Currently simplified. Full Windows registry parsing for application versions will be enhanced in future versions.

5. **Application Versions**: Some applications may show "unknown" version if version metadata isn't easily accessible.

## Roadmap

### Version 1.1
- [ ] Add support for poetry, pipenv, conda
- [ ] Add support for maven, gradle, nuget
- [ ] Enhanced Windows Registry parsing for accurate version information
- [ ] Add Atom IDE scanner

### Version 1.2
- [ ] Upload SBOM files to remote endpoints (S3, HTTP, etc.)
- [ ] Integration with vulnerability databases
- [ ] Docker image for containerized scanning

### Version 2.0
- [ ] Web UI for viewing SBOMs
- [ ] Historical tracking and diff reports
- [ ] Compliance reporting

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Security

If you discover a security vulnerability, please email security@example.com instead of using the issue tracker.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

**Tim Jensen (EapolSniper)**
- GitHub: [@eapolsniper](https://github.com/eapolsniper)
- Project: [EndpointBOM](https://github.com/eapolsniper/endpointbom)

## Acknowledgments

- CycloneDX for the SBOM standard
- OWASP for security best practices
- All the open-source projects that make this tool possible

## Support

- **Issues**: https://github.com/eapolsniper/endpointbom/issues
- **Discussions**: https://github.com/eapolsniper/endpointbom/discussions

---

**Note**: This tool is designed for inventory and security scanning purposes. It does not collect sensitive information such as API keys, tokens, or environment variable values. All scanning is performed locally, and no data is transmitted without explicit configuration.
