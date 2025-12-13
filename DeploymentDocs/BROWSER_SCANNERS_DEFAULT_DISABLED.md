# Browser Scanners: Disabled by Default

## ✅ Change Summary

**Browser extension scanners are now DISABLED by default** (as of v1.1.0)

This prevents macOS TCC permission popups and makes deployment easier.

## What Changed

### Before (v1.0)
```yaml
# All scanners enabled by default
disabled_scanners: []

# Result on macOS: Permission popup appears
```

### After (v1.1)
```yaml
# Browser scanners disabled by default
disabled_scanners:
  - chrome-extensions
  - firefox-extensions
  - edge-extensions
  - safari-extensions

# Result on macOS: No popups, works immediately ✅
```

## Impact

### ✅ Benefits

1. **No Permission Popups**
   - Works out-of-the-box on macOS
   - No TCC configuration required for basic scanning
   - Automated Jamf deployment works immediately

2. **Easier Deployment**
   - Install and run without additional setup
   - No MDM configuration needed (unless you want browser data)
   - Works on all macOS versions

3. **Opt-In Model**
   - Teams that need browser data can enable it
   - Teams without TCC access aren't blocked
   - Flexibility for different environments

### ⚠️ Trade-offs

1. **Less Data by Default**
   - Browser extension data not collected unless enabled
   - Teams must opt-in to get full security visibility
   - Requires explicit configuration change

2. **Two-Step Process for Browser Data**
   - Step 1: Configure TCC (macOS only)
   - Step 2: Enable scanners in config
   - More complex than "just works"

## Who Should Enable Browser Scanning?

### ✅ Enable If:
- You have Jamf or other MDM that can deploy TCC profiles
- You need browser extension data for security analysis
- You're investigating incidents like Shai Hulud or malicious extensions
- You're on Windows (no TCC issues)

### ⚠️ Keep Disabled If:
- You can't deploy TCC Configuration Profiles
- You want simplest possible deployment
- You don't need browser extension visibility
- You're testing and don't want popups

## How to Enable

### Quick Method (CLI)

```bash
# Enable all scanners (including browsers)
sudo endpointbom --disable=""
```

### Permanent Method (Config File)

```yaml
# /usr/local/etc/endpointbom.yaml
disabled_scanners: []  # Empty list = all enabled
```

Or remove specific scanners:

```yaml
disabled_scanners:
  # Browser scanners removed from list
  # - chrome-extensions  ← Commented out
  # - firefox-extensions ← Commented out
  # - edge-extensions    ← Commented out
  # - safari-extensions  ← Commented out
```

## Deployment Examples

### Example 1: Default Deployment (No Browser Data)

```bash
# Jamf policy script
#!/bin/bash
sudo /usr/local/bin/endpointbom --output=/var/log/endpointbom/scans
exit 0
```

**Result:**
- ✅ Works immediately
- ✅ No popups
- ✅ Gets packages, applications, IDE extensions
- ❌ No browser extension data

### Example 2: Full Deployment (With Browser Data)

```bash
# Jamf policy script
#!/bin/bash

# Deploy config with browsers enabled
cat > /usr/local/etc/endpointbom.yaml <<EOF
disabled_scanners: []
output_dir: /var/log/endpointbom/scans
EOF

# Run scan
sudo /usr/local/bin/endpointbom --config=/usr/local/etc/endpointbom.yaml
exit 0
```

**Prerequisites:**
- TCC Configuration Profile deployed first
- See `JAMF_DEPLOYMENT.md`

**Result:**
- ✅ Gets ALL data including browser extensions
- ✅ No popups (TCC configured)
- ✅ Full security visibility

### Example 3: Windows Deployment (Enable Browsers)

```powershell
# Windows deployment script
# No TCC issues - safe to enable all scanners

C:\ProgramData\endpointbom\endpointbom.exe --disable="" --output="C:\ProgramData\endpointbom\scans"
```

**Result:**
- ✅ Gets ALL data including browser extensions
- ✅ No permission issues on Windows
- ✅ Full security visibility

## Testing

### Test Default Behavior (Browsers Disabled)

```bash
$ sudo endpointbom --verbose

Skipping disabled scanner: chrome-extensions
Skipping disabled scanner: firefox-extensions
Skipping disabled scanner: edge-extensions
Skipping disabled scanner: safari-extensions

=== Scan Summary ===
Browser Extensions: 0  ← Zero (disabled)
```

### Test Enabled Behavior (Browsers Enabled)

```bash
$ sudo endpointbom --verbose --disable=""

Running scanner: chrome-extensions
  Found 14 components
Running scanner: firefox-extensions
  Found 0 components
Running scanner: edge-extensions
  Found 0 components
Running scanner: safari-extensions
  Found 10 components

=== Scan Summary ===
Browser Extensions: 24  ← Non-zero (enabled)
```

## Migration Guide

### If You're Already Using v1.0

**No action required!** The change is backwards compatible.

**If you want browser data:**
1. Configure TCC (macOS only) - see `JAMF_DEPLOYMENT.md`
2. Update config to enable browsers - see `ENABLING_BROWSER_SCANNING.md`

**If you don't need browser data:**
- Nothing to do! Default behavior works for you.

## FAQ

### Q: Will this break my existing deployment?
**A:** No! It's backwards compatible. Existing configs will continue to work.

### Q: Can I still get browser extension data?
**A:** Yes! Just enable the scanners. See `ENABLING_BROWSER_SCANNING.md`.

### Q: Why not keep them enabled by default?
**A:** macOS TCC popups break automated deployment. Disabled by default = works everywhere immediately.

### Q: Does this affect Windows?
**A:** Windows doesn't have TCC issues. You can safely enable browser scanners on Windows without any additional configuration.

### Q: What about other scanners (npm, pip, etc.)?
**A:** All other scanners remain enabled by default. Only browser scanners are disabled.

### Q: How do I know if browser scanning is enabled?
**A:** Check the scan summary:
```
Browser Extensions: 0  ← Disabled
Browser Extensions: 24 ← Enabled
```

## Documentation

- **How to Enable:** `ENABLING_BROWSER_SCANNING.md`
- **TCC Details:** `MACOS_TCC_PERMISSIONS.md`
- **Jamf Deployment:** `JAMF_DEPLOYMENT.md`
- **Browser Security:** `BROWSER_EXTENSIONS.md`
- **Quick Reference:** `TCC_QUICK_REFERENCE.md`

## Summary

**Default Behavior:** Browser scanners disabled → No TCC issues ✅  
**To Enable:** Configure TCC (macOS) + Update config  
**Windows:** No TCC needed, safe to enable anytime  
**Backwards Compatible:** Existing deployments unaffected  

This change makes EndpointBOM **easier to deploy** while still providing **opt-in access** to browser extension data for teams that need it.

---

**Version:** 1.1.0  
**Change Date:** 2025-12-13  
**Status:** Production Ready ✅

