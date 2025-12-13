# Enabling Browser Extension Scanning

## Default Behavior

**Browser extension scanners are DISABLED by default** to prevent macOS TCC permission popups.

This means EndpointBOM works out-of-the-box without any permission issues! ✅

## Why Disabled by Default?

On macOS 10.14+, accessing browser directories requires **Full Disk Access** (TCC permissions):
- Without TCC: Users see permission popups
- Popups break automated Jamf deployment
- Not all environments can configure TCC

**Solution:** Disabled by default, opt-in for teams that need it.

## How to Enable Browser Scanning

### Step 1: Configure TCC Permissions (macOS Only)

**Required for macOS 10.14+**

Deploy a Configuration Profile via Jamf/MDM:

```xml
<!-- See JAMF_DEPLOYMENT.md for complete XML -->
<key>SystemPolicyAllFiles</key>
<array>
    <dict>
        <key>Identifier</key>
        <string>/usr/local/bin/endpointbom</string>
        <key>IdentifierType</key>
        <string>path</string>
        <key>Allowed</key>
        <true/>
    </dict>
</array>
```

**Windows:** No TCC needed - skip to Step 2

### Step 2: Enable Browser Scanners

**Option A: Via Configuration File (Recommended)**

Edit `endpointbom.yaml`:

```yaml
# Remove browser scanners from disabled list
disabled_scanners:
  # - chrome-extensions    ← Remove or comment out
  # - firefox-extensions   ← Remove or comment out
  # - edge-extensions      ← Remove or comment out
  # - safari-extensions    ← Remove or comment out
  
  # Other scanners you want to disable
  # - npm
  # - pip
```

Or set to empty:

```yaml
disabled_scanners: []  # Enables all scanners including browsers
```

**Option B: Via Command Line**

```bash
# Specify a config file with browsers enabled
sudo endpointbom --config=/path/to/config-with-browsers-enabled.yaml

# Or disable other scanners (default browser disable won't apply)
sudo endpointbom --disable=npm,pip
# Note: Specifying --disable overrides defaults, so browsers will be enabled
```

**Option C: Via Environment Variable**

```bash
# Set in Jamf policy script
export ENDPOINTBOM_DISABLED_SCANNERS=""
sudo endpointbom
```

## Verification

### Test Browser Scanning is Enabled

```bash
$ sudo endpointbom --verbose

# Should see:
Running scanner: chrome-extensions
  Found XX components
Running scanner: firefox-extensions
  Found XX components
Running scanner: edge-extensions
  Found XX components
Running scanner: safari-extensions
  Found XX components

Browser Extensions: XX  ← Non-zero count
```

### Check SBOM Output

```bash
$ ls -lh scans/*.browser-extensions.cdx.json

# Should see:
hostname.timestamp.browser-extensions.cdx.json  (file exists and has size)
```

### Verify No Popups (macOS)

On macOS, if TCC is configured correctly:
- ✅ No permission popups appear
- ✅ Browser extensions are found
- ✅ SBOM file is generated

If you see popups:
- ❌ TCC Configuration Profile not applied
- See `JAMF_DEPLOYMENT.md` for troubleshooting

## Deployment Scenarios

### Scenario 1: Enterprise with Jamf (Enable Browser Scanning)

```yaml
# 1. Deploy TCC Configuration Profile via Jamf
# 2. Deploy endpointbom with custom config

# /usr/local/etc/endpointbom.yaml
disabled_scanners: []  # Enable all scanners
output_dir: /var/log/endpointbom/scans
scan_all_users: true
```

**Result:** Full visibility including browser extensions ✅

### Scenario 2: Enterprise without TCC Access (Default)

```yaml
# Use default configuration (browser scanners disabled)

# /usr/local/etc/endpointbom.yaml
disabled_scanners:
  - chrome-extensions
  - firefox-extensions
  - edge-extensions
  - safari-extensions
```

**Result:** No TCC issues, but no browser extension data ⚠️

### Scenario 3: Windows Environment (Enable Browser Scanning)

```yaml
# Windows doesn't have TCC - safe to enable

# C:\ProgramData\endpointbom\config.yaml
disabled_scanners: []  # Enable all scanners including browsers
output_dir: C:\ProgramData\endpointbom\scans
```

**Result:** Full visibility including browser extensions ✅

### Scenario 4: Mixed Mac/Windows Fleet

**Jamf Policy Script:**

```bash
#!/bin/bash

# Detect OS and use appropriate config
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS - check if config file exists (deployed with TCC profile)
    if [ -f /usr/local/etc/endpointbom-with-browsers.yaml ]; then
        # Config exists - browsers enabled
        sudo endpointbom --config=/usr/local/etc/endpointbom-with-browsers.yaml
    else
        # No config - use defaults (browsers disabled)
        sudo endpointbom
    fi
else
    # Windows - no TCC issues, enable all scanners
    endpointbom.exe --config=C:\ProgramData\endpointbom\config.yaml
fi
```

## Configuration Examples

### Minimal Config (Enable Browsers Only)

```yaml
disabled_scanners: []
```

### Selective Scanning (Browsers + Applications Only)

```yaml
disabled_scanners:
  - npm
  - pip
  - yarn
  - pnpm
  - brew
  - gem
  - cargo
  - composer
  - chocolatey
  - go
  - vscode
  - cursor
  - jetbrains
  - sublime
# Browser scanners NOT in list = enabled
```

### Maximum Performance (Disable Slow Scanners)

```yaml
disabled_scanners:
  - npm          # Can be slow with many packages
  - pip          # Can be slow
  - brew         # Can be slow on macOS
  # Browsers enabled for security visibility
```

## Troubleshooting

### Issue: Browser Extensions Still Zero

**Cause:** Scanners still disabled

**Solution:**
```bash
# Check config
cat /usr/local/etc/endpointbom.yaml | grep -A 10 disabled_scanners

# Verify chrome-extensions is NOT in the list
# If it is, remove it and re-run
```

### Issue: Permission Popups on macOS

**Cause:** TCC not configured

**Solution:**
1. Deploy TCC Configuration Profile via Jamf (see `JAMF_DEPLOYMENT.md`)
2. Wait 15-30 minutes for profile to apply
3. Test by running the tool - popups should not appear

### Issue: Works Manually, Not in Jamf

**Cause:** Config file not found by Jamf policy

**Solution:**
```bash
# In Jamf policy, specify config explicitly
sudo endpointbom --config=/usr/local/etc/endpointbom.yaml
```

## Security Considerations

### Why Enable Browser Scanning?

**High-Value Security Data:**
- Browser extensions can access ALL web traffic
- Malicious extensions are a common attack vector
- Critical for incident response (Shai Hulud, etc.)
- Permissions data shows risk level

**Examples of Dangerous Permissions:**
- `webRequest` + `<all_urls>` = Can intercept all traffic
- `cookies` + `*://*/*` = Can steal all cookies
- `nativeMessaging` = Can execute native code

### Why Keep It Disabled?

**Operational Simplicity:**
- No TCC configuration needed
- Works immediately on any macOS system
- No permission popups
- Easier deployment

**Trade-off:** Less security visibility

## Recommendation

### For Security Teams
✅ **Enable browser scanning** if you can configure TCC
- Deploy Configuration Profile via Jamf
- Enable scanners in config
- Get critical security visibility

### For IT Teams Without MDM
⚠️ **Keep browsers disabled** (default)
- Avoid deployment complications
- Still get package manager + application data
- Consider enabling on Windows only

### For Testing
🔧 **Enable manually** for testing
- User clicks "OK" on popup (one-time)
- Test browser scanning functionality
- Decide if worth deploying TCC for production

## Summary

**Default:** Browser scanners disabled → No TCC issues ✅  
**To Enable:** Configure TCC + Remove from disabled list  
**Windows:** No TCC needed, safe to enable  
**macOS:** Requires Full Disk Access via MDM  

**Documentation:**
- `JAMF_DEPLOYMENT.md` - Complete TCC setup
- `MACOS_TCC_PERMISSIONS.md` - TCC details
- `BROWSER_EXTENSIONS.md` - Security analysis guide

---

**Status:** Browser scanning is opt-in for teams that need it  
**Last Updated:** 2025-12-13

