# EndpointBOM Deployment Documentation

This folder contains deployment-specific documentation for enterprise environments.

## 📁 Documents

### Getting Started
- **[SCHEDULING.md](SCHEDULING.md)** - **START HERE** - How to schedule daily scans on macOS, Windows, and Linux
- **[JAMF_DEPLOYMENT.md](JAMF_DEPLOYMENT.md)** - Jamf Pro deployment guide (macOS enterprise)

### Internal Release Management (Not Published)
- **[GITHUB_RELEASE_PROCESS.md](GITHUB_RELEASE_PROCESS.md)** - How to create and publish GitHub releases
- **[CONVENTIONAL_COMMITS.md](CONVENTIONAL_COMMITS.md)** - Commit message guidelines for automated versioning
- **[RELEASE_NOTES_TEMPLATE.md](RELEASE_NOTES_TEMPLATE.md)** - Template for release notes

### Quick Reference
- **[TCC_QUICK_REFERENCE.md](TCC_QUICK_REFERENCE.md)** - Quick reference card for macOS TCC permissions
- **[BROWSER_SCANNERS_DEFAULT_DISABLED.md](BROWSER_SCANNERS_DEFAULT_DISABLED.md)** - Why browser scanners are disabled by default

### macOS Deployment
- **[MACOS_TCC_PERMISSIONS.md](MACOS_TCC_PERMISSIONS.md)** - Complete guide to macOS TCC permissions

### Browser Extensions
- **[ENABLING_BROWSER_SCANNING.md](ENABLING_BROWSER_SCANNING.md)** - How to enable browser extension scanning
- **[BROWSER_EXTENSIONS_SEPARATION.md](BROWSER_EXTENSIONS_SEPARATION.md)** - Why browser extensions have separate SBOM files

## 🎯 Common Scenarios

### Scenario 1: Basic Deployment with Daily Scans

**Goal:** Deploy tool to scan packages, applications, IDE extensions daily

**Steps:**
1. Download binary from GitHub releases
2. Install to system path (`/usr/local/bin/` or `C:\Program Files\`)
3. Set up daily schedule (see [SCHEDULING.md](SCHEDULING.md))
4. Done! No TCC permissions needed (unless enabling browser scanning)

**Result:** Works immediately, automated daily scans ✅

### Scenario 2: Enterprise Jamf Deployment (macOS)

**Goal:** Deploy to all macOS endpoints via Jamf with daily scans

**Steps:**
1. Create package with binary
2. Deploy via Jamf policy
3. Deploy LaunchDaemon for daily execution (see [JAMF_DEPLOYMENT.md](JAMF_DEPLOYMENT.md))
4. Optional: Deploy TCC profile for browser scanning

**Result:** Automated enterprise-wide scanning ✅

### Scenario 3: Full Deployment with Browser Scanning

**Goal:** Get complete security visibility including browser extensions

**Steps:**
1. Deploy binary
2. Set up daily schedule (see [SCHEDULING.md](SCHEDULING.md))
3. **macOS**: Deploy TCC Configuration Profile (see [JAMF_DEPLOYMENT.md](JAMF_DEPLOYMENT.md))
4. Deploy config file with browsers enabled (see [ENABLING_BROWSER_SCANNING.md](ENABLING_BROWSER_SCANNING.md))

**Result:** Full data including browser extensions, no popups ✅

### Scenario 4: Windows Deployment

**Goal:** Deploy to Windows endpoints with daily scans

**Steps:**
1. Download Windows binary from GitHub releases
2. Deploy to `C:\Program Files\EndpointBOM\endpointbom.exe`
3. Create scheduled task (see [SCHEDULING.md](SCHEDULING.md#windows))
4. Optional: Enable browser scanning (no TCC issues on Windows)

**Result:** Works immediately with automated daily scans ✅

## 🔧 Configuration

### Default Behavior
- ✅ Scans packages, applications, IDE extensions
- ❌ Browser scanning disabled (to avoid macOS TCC popups)
- ✅ Works immediately on all platforms

### Enable Browser Scanning

**macOS:**
1. Deploy TCC profile (Full Disk Access)
2. Update config: `disabled_scanners: []`

**Windows:**
1. Update config: `disabled_scanners: []`
2. No additional permissions needed

## 📊 Output

### SBOM Files Generated

```
scans/
├── hostname.timestamp.package-managers.cdx.json
├── hostname.timestamp.applications.cdx.json
├── hostname.timestamp.ide-extensions.cdx.json
└── hostname.timestamp.browser-extensions.cdx.json (if enabled)
```

### Default Output Location

- **macOS/Linux:** `./scans/` (relative to binary)
- **Windows:** `.\scans\` (relative to binary)
- **Customizable:** Use `--output` flag or config file

## 🆘 Troubleshooting

### Issue: Permission Popups on macOS

**Solution:** Browser scanners are disabled by default. If you enabled them, deploy TCC Configuration Profile via Jamf.  
**See:** [MACOS_TCC_PERMISSIONS.md](MACOS_TCC_PERMISSIONS.md)

### Issue: No Browser Extension Data

**Solution:** Browser scanners are disabled by default. To enable, see [ENABLING_BROWSER_SCANNING.md](ENABLING_BROWSER_SCANNING.md)

### Issue: Jamf Deployment Fails

**Solution:** Check Jamf policy logs and ensure binary has execute permissions.  
**See:** [JAMF_DEPLOYMENT.md](JAMF_DEPLOYMENT.md)

## 📚 Related Documentation

### Main Documentation (in project root)
- **[README.md](../README.md)** - Project overview
- **[QUICKSTART.md](../QUICKSTART.md)** - Quick start guide
- **[docs/USAGE.md](../docs/USAGE.md)** - Complete usage documentation

### Technical Documentation
- **[BROWSER_EXTENSIONS.md](../BROWSER_EXTENSIONS.md)** - Browser extension security analysis
- **[SECURITY_IMPROVEMENTS.md](../SECURITY_IMPROVEMENTS.md)** - Security enhancements
- **[NETWORK_INFO.md](../NETWORK_INFO.md)** - Network information collection

## 🎓 Best Practices

1. **Start Simple**
   - Deploy without browser scanning first
   - Verify basic functionality
   - Add browser scanning later if needed

2. **Test Before Production**
   - Deploy to pilot group first
   - Verify SBOM files are generated
   - Check for any permission issues

3. **Monitor Deployment**
   - Check Jamf/Intune logs
   - Verify scans run daily
   - Alert on missing or failed scans

4. **Secure SBOM Files**
   - Store centrally with restricted access
   - Encrypt during transmission
   - Follow data retention policies

---

**For questions or issues, see the main [README.md](../README.md) for support options.**

