# EndpointBOM Changelog

## [0.0.13] - 2025-12-14

See the [GitHub Release](https://github.com/eapolsniper/endpointbom/releases/tag/v0.0.13) for full details.

---



## [0.0.12] - 2025-12-14

See the [GitHub Release](https://github.com/eapolsniper/endpointbom/releases/tag/v0.0.12) for full details.

---



## [0.0.11] - 2025-12-13

See the [GitHub Release](https://github.com/eapolsniper/endpointbom/releases/tag/v0.0.11) for full details.

---



## [1.2.0] - 2025-12-13

### Added - Automated Release Management
- **Automated Versioning** - Uses conventional commits for semantic versioning
- **Automatic Releases** - Builds and publishes releases on every merge to main
- **Multi-Platform Builds** - Generates binaries for macOS (Intel + Apple Silicon), Windows, Linux
- **GoReleaser Integration** - Professional release artifacts with checksums
- **Version Command** - `--version` flag shows build info (version, commit, date)

### Added - Automated Dependency Management
- **Dependabot Configuration** - Automated dependency updates with security focus
- **14-Day Cooldown Policy** - Non-critical updates wait 14 days for community vetting
- **Auto-Merge Workflow** - Critical/high severity fixes merge automatically
- **Dependency Testing** - Comprehensive test suite runs on all dependency updates

### Security - Dependency Update Policy
- ✅ **Critical/High Severity** - Immediate auto-merge after tests pass
- ⏱️ **Medium/Low/Non-Security** - 14-day cooldown before auto-merge
- 🧪 **All Updates** - Full test suite (Linux, macOS, Windows) must pass
- 🔒 **Vulnerability Scanning** - `govulncheck` runs on every update
- 📋 **License Review** - Blocks GPL-3.0 and AGPL-3.0 licenses

### Release Process - Conventional Commits
Version bumps are determined automatically from commit messages:
- **MAJOR** (x.0.0): Breaking changes (`feat!:` or `BREAKING CHANGE:`)
- **MINOR** (0.x.0): New features (`feat:`)
- **PATCH** (0.0.x): Bug fixes, security, dependencies (`fix:`, `security:`, `deps:`)

### Files Added
**Dependency Management:**
- `.github/dependabot.yml` - Dependabot configuration
- `.github/workflows/dependabot-auto-merge.yml` - Auto-merge with cooldown
- `.github/workflows/dependency-test.yml` - Comprehensive dependency tests
- `.github/README.md` - Automation documentation

**Release Management:**
- `.github/workflows/release.yml` - Automated release workflow
- `.goreleaser.yml` - GoReleaser configuration
- `internal/version/version.go` - Version information module
- `.github/CONVENTIONAL_COMMITS.md` - Commit guidelines
- `.github/workflows/release-notes-template.md` - Release notes template

### Benefits
1. **Security** - Critical vulnerabilities patched within hours
2. **Stability** - Community vetting period for non-critical updates
3. **Automation** - Zero manual intervention for releases and dependency updates
4. **Transparency** - All changes tracked via conventional commits and changelogs
5. **Quality** - Every release tested on all platforms before publishing

---

## [1.1.0] - 2025-12-13

### Changed - Browser Scanners Disabled by Default
- **Browser extension scanners are now DISABLED by default**
- Prevents macOS TCC permission popups during automated deployment
- Teams can opt-in to browser scanning if they have TCC configured
- No breaking changes - just change the default behavior

### Added - Browser Extension Scanning
- **Chrome Extension Scanner** - Scans all Chrome/Chromium extensions
- **Firefox Extension Scanner** - Scans Firefox add-ons and extensions
- **Edge Extension Scanner** - Scans Microsoft Edge extensions
- **Safari Extension Scanner** - Scans Safari extensions (macOS only)

### Added - Network Information
- **Local IP Detection** - Captures all local IPv4 and IPv6 addresses
- **Public IP Detection** - Determines external-facing IP address
- Non-blocking with graceful fallback if network unavailable

### Added - Security Enhancements
- **Path Validation Module** - Protects against path traversal attacks
- **Sensitive File Protection** - Blocks access to SSH keys, credentials, secrets
- **System Directory Protection** - Prevents writes to critical system locations
- **Adaptive Privilege Handling** - Gracefully handles admin vs non-admin execution

### Changed - SBOM File Organization
- **Separate Browser Extensions SBOM** - Browser extensions now generate their own dedicated file
  - `{hostname}.{timestamp}.browser-extensions.cdx.json`
- This improves security analysis by isolating browser extension data

### File Structure
```
Before:
- package-managers.cdx.json
- applications.cdx.json
- ide-extensions.cdx.json (included browser extensions)

After:
- package-managers.cdx.json
- applications.cdx.json
- ide-extensions.cdx.json (IDE extensions and MCP servers only)
- browser-extensions.cdx.json (NEW - browser extensions only)
```

### Security Data Captured for Browser Extensions
- Extension ID (unique identifier)
- Extension name and version
- Browser type (chrome, firefox, edge, safari)
- **Permissions** (tabs, webRequest, cookies, storage, etc.)
- **Host Permissions** (which websites can be accessed)
- Manifest version (v2 vs v3)
- Installation location

### Benefits of Separate Browser Extensions File

1. **Focused Security Analysis**
   - Dedicated file for browser extension threat hunting
   - Easier to process with security tools
   - Clear separation of concerns

2. **Incident Response**
   - Quickly analyze only browser extensions during incidents
   - Compare browser extension baselines across fleet
   - Integrate with browser-specific threat intelligence

3. **Compliance & Reporting**
   - Generate browser extension reports independently
   - Track approved vs unapproved extensions
   - Monitor permission changes over time

4. **Performance**
   - Smaller, more focused files
   - Faster parsing for browser-specific queries
   - Better for automated processing

### Example Output

```bash
$ sudo endpointbom --verbose

=== Scan Summary ===
Package Manager Components: 168
Applications: 182
IDE Extensions/Plugins: 5
Browser Extensions: 24
Output Directory: ./scans

=== Generating SBOMs ===
Generated: developer-laptop.20251213-112542.package-managers.cdx.json
Generated: developer-laptop.20251213-112542.applications.cdx.json
Generated: developer-laptop.20251213-112542.ide-extensions.cdx.json
Generated: developer-laptop.20251213-112542.browser-extensions.cdx.json ← NEW!

✓ Scan complete!
```

### Use Cases

**Scenario 1: Malicious Extension Detection**
```bash
# Scan endpoint
sudo endpointbom

# Analyze ONLY browser extensions
cat *.browser-extensions.cdx.json | grep -A 10 '"browser": "chrome"'

# Look for dangerous permissions
grep -i "webRequest.*all_urls" *.browser-extensions.cdx.json
```

**Scenario 2: Fleet-Wide Browser Extension Audit**
```bash
# Collect all browser extension SBOMs
scp endpoints:~/scans/*.browser-extensions.cdx.json ./fleet-audit/

# Find all endpoints with specific extension
grep -r "extension_id.*abcdefgh" fleet-audit/

# Generate report
./analyze-browser-extensions.sh fleet-audit/
```

**Scenario 3: Baseline Comparison**
```bash
# Establish baseline
sudo endpointbom
cp *.browser-extensions.cdx.json baseline/

# Later, compare current state
sudo endpointbom
diff baseline/*.browser-extensions.cdx.json scans/*.browser-extensions.cdx.json
```

### Documentation Updates
- Updated `README.md` with browser extension features
- Created `BROWSER_EXTENSIONS.md` - Comprehensive security guide
- Created `BROWSER_EXTENSION_IMPLEMENTATION.md` - Technical details
- Updated `QUICKSTART.md` with new file structure
- Updated `FEATURES_SUMMARY.md` with complete feature list
- Created `NETWORK_INFO.md` - Network information documentation
- Organized deployment docs into `DeploymentDocs/` folder

### Technical Changes
- Added `BrowserExtensions []Component` to `ScanResult` struct
- Updated component categorization logic to separate browser extensions
- Added browser-extensions SBOM generation in `cyclonedx.go`
- Updated summary output to show browser extension count
- All browser scanners integrated and tested

---

## [1.0.0] - 2025-12-12

### Initial Release
- Package manager scanning (npm, pip, yarn, pnpm, brew, gem, cargo, composer, chocolatey, go)
- Application discovery (macOS, Windows, Linux)
- IDE extension scanning (VSCode, Cursor, JetBrains, Sublime)
- MCP server detection
- CycloneDX 1.5 SBOM generation
- Configuration file support
- CLI flags for customization
- Cross-platform support (macOS, Windows, Linux)
- Adaptive privilege handling
- Path exclusion and scanner disabling
- Verbose and debug modes

---

## Version Numbering

This project follows [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API/output changes
- **MINOR** version for new functionality in a backwards compatible manner
- **PATCH** version for backwards compatible bug fixes

---

**Current Version**: 1.2.0  
**Status**: Production Ready ✅
