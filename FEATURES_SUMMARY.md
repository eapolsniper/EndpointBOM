# EndpointBOM Features Summary

## ✅ Implemented Features

### 1. Core Scanning Capabilities

#### Package Managers
- **npm** (Node.js) - Global packages + transitive dependencies
- **pip** (Python) - System packages via pip freeze
- **yarn** (Node.js) - Global packages
- **pnpm** (Node.js) - Global packages
- **Homebrew** (macOS) - All installed packages
- **gem** (Ruby) - System gems
- **cargo** (Rust) - Installed crates
- **composer** (PHP) - Global packages
- **chocolatey** (Windows) - Installed packages
- **Go modules** - Go packages

#### Applications
- **macOS**: All applications in /Applications and ~/Applications
- **Windows**: Installed applications via registry and common locations
- **Linux**: Applications via dpkg, rpm, and common paths
- Excludes OS components and system utilities

#### IDE Extensions & Plugins
- **VSCode**: All installed extensions with metadata
- **Cursor**: Extensions and configuration
- **JetBrains Suite**: IntelliJ IDEA, PyCharm, WebStorm, PhpStorm, GoLand, RubyMine, CLion, Rider, DataGrip, AndroidStudio
- **Sublime Text**: Installed packages
- **MCP Servers**: Model Context Protocol servers detected in supported IDEs

#### 🆕 Browser Extensions (NEW!)
- **Chrome**: All extensions with:
  - Extension ID, name, version
  - Permissions (tabs, webRequest, cookies, storage, etc.)
  - Host permissions (which websites it can access)
  - Manifest version (v2 vs v3)
  
- **Firefox**: Add-ons and extensions:
  - Add-on ID, name, version
  - Permissions
  - Profile association
  
- **Microsoft Edge**: Extensions:
  - Extension ID, name, version
  - Permissions and host permissions
  - Manifest version
  
- **Safari** (macOS only): Extensions:
  - Legacy .safariextension format
  - Modern Safari App Extensions
  - Extension type identification

### 2. System Metadata Collection

- **Hostname**: Full system hostname
- **OS & Version**: Operating system and version
- **Logged-in Users**: All currently logged-in users
- **Local IP Addresses**: All IPv4 and IPv6 addresses (excluding loopback)
- **Public IP Address**: External-facing IP address
- **Scan Timestamp**: When the scan was performed
- **Scan Category**: Type of SBOM (applications, package-managers, ide-extensions)

### 3. Security Features

#### Path Validation & Security
- **Sensitive Path Protection**: Blocks access to sensitive system files
  - `/etc/shadow`, `/etc/passwd`
  - SSH keys (`~/.ssh/id_rsa`, `~/.ssh/id_ed25519`)
  - Cloud credentials (`~/.aws/credentials`, `~/.config/gcloud/`)
  - Application secrets (`.env`, `.secrets`)
  
- **Path Traversal Prevention**: Sanitizes and validates all file paths
- **System Directory Protection**: Prevents writes to critical system directories
- **Built-in Path Exclusions**: Automatically excludes sensitive directories

#### Adaptive Privilege Handling
- **Automatic Detection**: Detects if running as admin/root
- **Smart Fallback**: If not admin, scans current user only (with warning)
- **No Hard Failures**: Never fails due to privilege issues
- **Clear Messaging**: Explains privilege status and scan scope

#### Dependency Security
- **Pinned Versions**: All dependencies use pinned versions
- **Minimal Dependencies**: Only essential, well-maintained libraries
- **Trusted Sources**: Official Go modules from reputable sources

### 4. Output & Format

#### CycloneDX 1.5 SBOM
- **Widely Compatible**: Industry standard format
- **Component Categories**: Applications, libraries, browser extensions
- **Rich Metadata**: Version, description, location, permissions
- **Properties**: Custom properties for browser-specific data

#### File Organization
```
scans/
├── hostname.timestamp.applications.cdx.json
├── hostname.timestamp.package-managers.cdx.json
├── hostname.timestamp.ide-extensions.cdx.json
└── hostname.timestamp.browser-extensions.cdx.json
```

#### SBOM Categories
1. **applications**: Non-OS applications installed on the system
2. **package-managers**: Packages from npm, pip, brew, etc.
3. **ide-extensions**: IDE extensions, plugins, MCP servers
4. **browser-extensions**: Browser extensions from Chrome, Firefox, Edge, Safari

### 5. Configuration & Flexibility

#### Command-Line Flags
```bash
--config FILE          # Custom config file path
--output DIR           # Output directory for SBOMs
--verbose, -v          # Verbose output
--debug                # Debug mode with error details
--disable SCANNERS     # Disable specific scanners
--exclude PATHS        # Exclude specific paths
--require-admin        # Fail if not admin (default: false)
--scan-all-users       # Scan all user profiles (default: true if admin)
```

#### Configuration File (YAML)
```yaml
scanners:
  disabled:
    - npm
    - chrome-extensions
    
paths:
  exclude:
    - /Users/*/Library/Caches
    - /tmp
    
output_dir: /central/sboms
require_admin: false
scan_all_users: true
```

### 6. Use Cases

#### Incident Response
- **Malicious Extension Detection**: Identify suspicious browser extensions
- **Compromised Package Detection**: Find malicious npm/pip packages
- **Supply Chain Attacks**: Detect Shai Hulud-style compromises
- **Code Exfiltration**: Identify extensions with dangerous permissions

#### Compliance & Inventory
- **Software Inventory**: Complete list of all software on endpoints
- **Version Tracking**: Know exact versions for vulnerability management
- **License Compliance**: Inventory for license auditing (future)
- **Audit Trails**: Timestamped records of endpoint state

#### Security Operations
- **Baseline Comparison**: Compare against known-good state
- **Threat Hunting**: Search for specific packages/extensions across fleet
- **Anomaly Detection**: Identify outlier configurations
- **Automated Scanning**: Daily/weekly scans via Jamf or similar

## 🔍 Security-Relevant Browser Extension Data

### High-Risk Permission Combinations

**Critical Permissions to Flag:**
1. `webRequest` + `<all_urls>` = Can intercept ALL network traffic
2. `cookies` + `*://*/*` = Can steal cookies from all sites
3. `tabs` + broad host permissions = Can inject code anywhere
4. `nativeMessaging` = Can execute native code
5. `webNavigation` = Can track all browsing

**Example from Your System:**
```json
{
  "name": "Adobe Acrobat Extension",
  "permissions": "webRequest, cookies, tabs, storage",
  "host_permissions": "<all_urls>"
}
```
This extension can:
- See all your web traffic (webRequest)
- Access all your cookies (cookies + all_urls)
- Inject code into any website (tabs + all_urls)

### Analysis Workflow

1. **Scan Endpoints**
   ```bash
   sudo endpointbom --verbose
   ```

2. **Extract Browser Extensions**
   ```bash
   grep '"browser":' *.ide-extensions.cdx.json
   ```

3. **Flag High-Risk Extensions**
   - Extensions with `<all_urls>` or `*://*/*`
   - Extensions with `webRequest` + broad permissions
   - Unknown extensions (not in approved list)

4. **Cross-Reference with Threat Intel**
   - Known malicious extension IDs
   - Typosquatting (similar names to legit extensions)
   - Extensions from untrusted developers

5. **Investigate Suspicious Findings**
   - When was it installed? (filesystem timestamps)
   - Who installed it? (user profile)
   - What permissions does it have?
   - Is it in Chrome Web Store / Firefox Add-ons?

## 📊 Sample Scan Output

```
Gathering system information...
Hostname: developer-macbook.local
OS: darwin 14.6
Logged in users: [jsmith]
Local IP(s): [192.168.1.100, 10.0.0.50]
Public IP: 203.0.113.45
Output directory: /Users/jsmith/scans

Running scanner: chrome-extensions
  Found 12 components
Running scanner: firefox-extensions
  Found 3 components
Running scanner: safari-extensions
  Found 5 components

=== Scan Summary ===
Package Manager Components: 234
Applications: 87
IDE Extensions/Plugins: 22
Browser Extensions: 20
Output Directory: /Users/jsmith/scans

=== Generating SBOMs ===
Generated: developer-macbook.local.20251213-120000.package-managers.cdx.json
Generated: developer-macbook.local.20251213-120000.applications.cdx.json
Generated: developer-macbook.local.20251213-120000.ide-extensions.cdx.json
Generated: developer-macbook.local.20251213-120000.browser-extensions.cdx.json

✓ Scan complete!
```

## 🚀 Deployment Options

### Homebrew (macOS)
```bash
brew tap yourusername/endpointbom
brew install endpointbom
```

### Jamf (Enterprise Deployment)
1. Package the binary
2. Create Jamf policy
3. Set to run daily/weekly
4. Collect SBOMs centrally

### Manual Installation
```bash
# Download release
wget https://github.com/eapolsniper/endpointbom/releases/latest/download/endpointbom-darwin-amd64

# Make executable
chmod +x endpointbom-darwin-amd64

# Run
sudo ./endpointbom-darwin-amd64
```

## 📚 Documentation

- **README.md** - Project overview and quick start
- **QUICKSTART.md** - Fast setup and common commands
- **docs/USAGE.md** - Comprehensive usage guide
- **docs/BUILD.md** - Building from source
- **BROWSER_EXTENSIONS.md** - Browser extension security guide
- **NETWORK_INFO.md** - Network information collection details
- **SECURITY_IMPROVEMENTS.md** - Security enhancements documentation

## 🎯 Real-World Scenarios

### Scenario 1: Malicious Extension Response
**Problem**: Developer's GitHub credentials stolen

**Action**:
```bash
# Scan the compromised endpoint
sudo endpointbom --verbose

# Review browser extensions in ide-extensions SBOM
# Look for extensions with github.com host permissions + webRequest
```

**Result**: Identified "Dev Tools Helper" extension with suspicious permissions

### Scenario 2: Supply Chain Attack
**Problem**: News breaks about compromised npm package "popular-lib"

**Action**:
```bash
# Scan all developer machines
for host in $(cat developer-hosts.txt); do
  ssh $host 'sudo endpointbom --output=/tmp/scans'
  scp $host:/tmp/scans/*.cdx.json ./central-scans/
done

# Search all SBOMs for the compromised package
grep -r '"name": "popular-lib"' central-scans/
```

**Result**: Identified 12 endpoints with vulnerable version

### Scenario 3: Code Exfiltration Investigation
**Problem**: Source code leaked, suspect browser extension

**Action**:
```bash
# Review browser extensions with dangerous permissions
# Flag extensions with:
# - webRequest + localhost access
# - clipboard permissions
# - broad host permissions on code hosting sites
```

**Result**: Found extension with `*://github.com/*` + `webRequest` + `clipboard`

## ⚡ Performance

- **Fast**: Completes in 10-30 seconds (depending on installed software)
- **Non-blocking**: Network calls (public IP) have 5-second timeout
- **Efficient**: Scans only relevant directories
- **Graceful**: Skips inaccessible locations without failing

## 🔮 Future Enhancements (Potential)

1. **Git Repository Scanning** - Find all cloned repos with remote URLs
2. **Running Process Inventory** - Active processes and their binaries
3. **Scheduled Tasks** - Cron jobs, LaunchAgents, Windows Tasks
4. **SSH Key Inventory** - List keys (fingerprints only, not content)
5. **Container Scanning** - Docker images and running containers
6. **Network Connections** - Active connections and listening ports
7. **Baseline Comparison** - Compare against known-good baseline
8. **Central Collection API** - Submit SBOMs to central server
9. **Threat Intelligence Integration** - Check against known malicious indicators

## 🏆 Why EndpointBOM?

### For Security Teams
✅ **Complete Visibility**: Know exactly what's on every endpoint  
✅ **Incident Response**: Quickly identify compromised endpoints  
✅ **Compliance**: Automated software inventory  
✅ **Threat Hunting**: Search across entire fleet  

### For Developers
✅ **Non-Intrusive**: Fast, doesn't slow down work  
✅ **Transparent**: Clear about what it collects  
✅ **Safe**: No secrets collected, path validation  
✅ **Flexible**: Can run as user or admin  

### For DevSecOps
✅ **CI/CD Ready**: Can integrate into pipelines  
✅ **Machine-Readable**: Standard SBOM format  
✅ **Extensible**: Easy to add new scanners  
✅ **Cross-Platform**: Mac, Windows, Linux  

---

**Last Updated**: 2025-12-13  
**Version**: 1.0.0 with Browser Extension Support  
**Status**: Production Ready ✅

