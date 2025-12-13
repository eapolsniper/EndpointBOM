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


### Output Format

- **CycloneDX 1.5** JSON format (widely compatible, feature-rich)
- Separate SBOM files by category:
  - `{hostname}.{timestamp}.package-managers.cdx.json`
  - `{hostname}.{timestamp}.applications.cdx.json`
  - `{hostname}.{timestamp}.ide-extensions.cdx.json`
  - `{hostname}.{timestamp}.browser-extensions.cdx.json` 
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

It's recommended to run with admin/root privileges for complete endpoint inventory, but if the system is used only by one person, the inventory should be nearly complete and "good enough" for the purpose of this tool.

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

### Enterprise Deployment

For enterprise deployment and automated scheduling, see **[DeploymentDocs](DeploymentDocs/)**:

- **[SCHEDULING.md](DeploymentDocs/SCHEDULING.md)** - Set up daily automated scans (macOS, Windows, Linux)
- **[JAMF_DEPLOYMENT.md](DeploymentDocs/JAMF_DEPLOYMENT.md)** - Jamf Pro deployment guide
- **[MACOS_TCC_PERMISSIONS.md](DeploymentDocs/MACOS_TCC_PERMISSIONS.md)** - macOS permissions for browser scanning


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


**Legend**:
- ✅ Fully supported
- ⚠️ Partially supported
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
   

2. **Project-Level Dependencies**: The tool scans globally installed packages, not project-specific installer files that haven't been installed `package.json`, `requirements.txt`, etc.

3. **Transitive Dependencies**: Not all package managers support querying transitive dependencies. Where possible (npm, pnpm, yarn, go), we capture them.

4. **Windows Registry Parsing**: Currently simplified. Full Windows registry parsing for application versions will be enhanced in future versions.

5. **Application Versions**: Some applications may show "unknown" version if version metadata isn't easily accessible.

## Roadmap

- [ ] Upload SBOM files to remote endpoints (S3, HTTP, etc.)


## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

Please follow standard Go coding conventions and ensure all tests pass before submitting.

## Security

If you discover a security vulnerability, please use the Security reporting tab in GitHub instead of using the issue tracker.

### Automated Dependency Management

This project uses **Dependabot** with automated testing and a **14-day cooldown policy** for dependency updates:

- ✅ **Critical/High Severity** - Security fixes merge automatically after tests pass
- ⏱️ **Medium/Low/Non-Security** - 14-day cooldown for community vetting before auto-merge
- 🧪 **All Updates** - Comprehensive test suite (Linux, macOS, Windows) required
- 🔒 **Vulnerability Scanning** - `govulncheck` runs on every dependency update
- 📋 **License Review** - Blocks incompatible licenses (GPL-3.0, AGPL-3.0)

For more details on the automation workflows, see [`.github/README.md`](.github/README.md).

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
