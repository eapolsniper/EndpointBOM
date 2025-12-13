# macOS TCC Permissions for EndpointBOM

## ✅ Browser Scanners Disabled by Default

**Good news:** Browser extension scanning is **disabled by default** to prevent TCC permission popups!

This means EndpointBOM will run without any permission prompts on macOS out of the box.

### Enabling Browser Scanning (Optional)

If you want browser extension data, you'll need to:
1. Grant Full Disk Access via MDM (see below)
2. Enable the scanners in your configuration

## 🔴 What Happens If You Enable Browser Scanning

### The Problem

When browser scanning is enabled on macOS, users see a popup:

```
"endpointbom" would like to access files in your Chrome folder.
[Don't Allow] [OK]
```

**This breaks automated deployment!**

### Why This Happens

macOS **Transparency, Consent, and Control (TCC)** protects these directories:
- `~/Library/Application Support/Google/Chrome/`
- `~/Library/Application Support/Microsoft Edge/`
- `~/Library/Safari/`
- `~/Library/Containers/`

**Even running as root (sudo) is not enough on macOS 10.14+!**

## ✅ Solution for Enterprise Deployment

### Answer: Use Jamf Configuration Profile

**Jamf can grant "Full Disk Access" automatically** - no user interaction required!

### Quick Setup (3 Steps)

1. **Create Configuration Profile in Jamf**
   - Navigate to: Computers → Configuration Profiles → New
   - Add: Privacy Preferences Policy Control payload

2. **Configure TCC Access**
   ```
   Identifier: /usr/local/bin/endpointbom
   Identifier Type: Path
   App or Service: SystemPolicyAllFiles
   Access: Allow
   ```

3. **Deploy to Fleet**
   - Scope to all endpoints
   - Deploy BEFORE or WITH the binary

**Result:** No popups, fully automated scanning! ✅

## Detailed Solutions

### Option 1: Full Disk Access (RECOMMENDED for Jamf)

**Pros:**
- ✅ No user prompts
- ✅ Scans all users
- ✅ Complete browser extension data
- ✅ Fully automated

**Cons:**
- ⚠️ Requires MDM (Jamf, Intune, etc.)
- ⚠️ Broad permission (but tool is read-only)

**See:** `JAMF_DEPLOYMENT.md` for complete instructions

### Option 2: Disable Browser Scanning

If TCC access cannot be granted:

```bash
# Disable all browser scanners
endpointbom --disable=chrome-extensions,firefox-extensions,edge-extensions,safari-extensions
```

Or in config file:
```yaml
scanners:
  disabled:
    - chrome-extensions
    - firefox-extensions
    - edge-extensions
    - safari-extensions
```

**Pros:**
- ✅ No TCC permissions needed
- ✅ No popups
- ✅ Works immediately

**Cons:**
- ❌ Misses browser extension data
- ❌ Less security visibility

### Option 3: Scan Current User Only

Run without admin privileges:

```bash
# No sudo - scans only current user
endpointbom --scan-all-users=false
```

**Pros:**
- ✅ No Full Disk Access needed
- ✅ Still scans current user's browser extensions
- ✅ One-time popup per user (they click OK)

**Cons:**
- ⚠️ User must click "OK" on first run
- ❌ Misses other users' extensions
- ❌ Not fully automated

## Technical Details

### What is TCC?

**Transparency, Consent, and Control (TCC)** is macOS's privacy protection system introduced in 10.14 (Mojave).

**Protected Resources:**
- Documents folder
- Downloads folder
- Desktop folder
- **Application Support folders** ← Browser extensions are here!
- Contacts, Calendar, Photos
- Camera, Microphone

### Why Browser Directories Are Protected

Chrome/Safari store extensions in:
```
~/Library/Application Support/Google/Chrome/Default/Extensions/
~/Library/Safari/Extensions/
~/Library/Containers/com.apple.Safari/
```

These are in `~/Library/Application Support/` which is **TCC-protected** to prevent:
- Malware stealing browser data
- Unauthorized access to extensions
- Privacy violations

### TCC System

macOS stores privacy permissions in a protected system database.

**Only MDM or manual user action can grant these permissions!**

You don't need to interact with this system directly - just deploy the Configuration Profile via Jamf and test the tool.

## Jamf Configuration Profile XML

Complete example:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>PayloadContent</key>
    <array>
        <dict>
            <key>PayloadType</key>
            <string>com.apple.TCC.configuration-profile-policy</string>
            <key>PayloadVersion</key>
            <integer>1</integer>
            <key>PayloadIdentifier</key>
            <string>com.company.endpointbom.tcc</string>
            <key>PayloadUUID</key>
            <string>12345678-1234-1234-1234-123456789012</string>
            <key>PayloadEnabled</key>
            <true/>
            <key>PayloadDisplayName</key>
            <string>EndpointBOM Full Disk Access</string>
            <key>Services</key>
            <dict>
                <key>SystemPolicyAllFiles</key>
                <array>
                    <dict>
                        <key>Identifier</key>
                        <string>/usr/local/bin/endpointbom</string>
                        <key>IdentifierType</key>
                        <string>path</string>
                        <key>StaticCode</key>
                        <false/>
                        <key>Allowed</key>
                        <true/>
                    </dict>
                </array>
            </dict>
        </dict>
    </array>
    <key>PayloadDisplayName</key>
    <string>EndpointBOM Privacy Access</string>
    <key>PayloadIdentifier</key>
    <string>com.company.endpointbom</string>
    <key>PayloadType</key>
    <string>Configuration</string>
    <key>PayloadUUID</key>
    <string>87654321-4321-4321-4321-210987654321</string>
    <key>PayloadVersion</key>
    <integer>1</integer>
</dict>
</plist>
```

**Upload this to Jamf and deploy to your fleet!**

## Testing TCC Access

### Before Configuration Profile

```bash
$ sudo endpointbom --verbose

Running scanner: chrome-extensions
[POPUP APPEARS] ← User must click OK
  Found 14 components
```

### After Configuration Profile

```bash
$ sudo endpointbom --verbose

Running scanner: chrome-extensions
  Found 14 components  ← No popup!
```

### Verify TCC Permission

Simply run the tool and check if browser extensions are found:

```bash
# If TCC is configured correctly, you'll see browser data
sudo endpointbom --verbose

# Look for:
# Running scanner: chrome-extensions
#   Found XX components  ← Should find extensions
```

If you see "Found 0 components" or permission errors, the TCC profile needs to be redeployed.

## Deployment Workflow

### Recommended Order

1. **Deploy Configuration Profile** (Day 1)
   - Push TCC profile via Jamf
   - Wait 1-2 hours for profile to apply
   - Verify profile installation

2. **Deploy Binary** (Day 2)
   - Install endpointbom to `/usr/local/bin/`
   - Set permissions: `chmod +x`
   - Test on pilot group

3. **Schedule Scans** (Day 3)
   - Create Jamf policy for daily scans
   - Set trigger: Recurring Check-In
   - Monitor for errors

### Verification Steps

```bash
# 1. Check profile is installed
sudo profiles -P | grep -i endpointbom

# 2. Test manual run
sudo /usr/local/bin/endpointbom --verbose

# 3. Check for browser extension data
ls -lh /var/log/endpointbom/scans/*.browser-extensions.cdx.json

# 4. Verify no errors
cat /var/log/endpointbom/scans/*.log
```

## Troubleshooting

### Issue: Popup Still Appears

**Cause:** Configuration profile not applied

**Solution:**
```bash
# Force profile refresh
sudo profiles renew -type enrollment

# Reboot endpoint
sudo reboot

# Check profile status
sudo profiles status -type enrollment
```

### Issue: Permission Denied Errors

**Cause:** Not running as root

**Solution:**
```bash
# Must use sudo
sudo /usr/local/bin/endpointbom
```

### Issue: Some Extensions Missing

**Cause:** Partial TCC access or user-specific extensions

**Solution:**
```bash
# Check which directories are accessible
sudo /usr/local/bin/endpointbom --verbose --debug 2>&1 | grep "Error scanning"

# Verify Full Disk Access is granted (not just specific folders)
```

### Issue: Works Manually, Fails in Jamf

**Cause:** Jamf policy running as different user

**Solution:**
```bash
# In Jamf policy script, explicitly use sudo
#!/bin/bash
sudo /usr/local/bin/endpointbom --output=/var/log/endpointbom/scans
exit 0
```

## Security Considerations

### Why Full Disk Access is Safe

1. **EndpointBOM is read-only**
   - Never writes to browser directories
   - Only reads extension metadata
   - No modification of browser data

2. **No secrets collected**
   - Skips passwords, cookies, session data
   - Only collects extension names, versions, permissions
   - See `SECURITY_IMPROVEMENTS.md`

3. **Path validation**
   - Blocks access to sensitive files
   - Prevents path traversal attacks
   - Validates all file operations

### Least Privilege Alternative

If Full Disk Access is too broad:

**Option:** Only scan current user
```bash
# Run as user (not sudo)
endpointbom --scan-all-users=false
```

**Trade-off:**
- ✅ No Full Disk Access needed
- ✅ Still gets browser extension data
- ❌ User must approve once
- ❌ Misses other users' data

## macOS Version Matrix

| macOS Version | TCC Required | Root Enough? | Solution |
|--------------|--------------|--------------|----------|
| 10.13 or earlier | No | Yes | Just use sudo |
| 10.14 (Mojave) | Yes | **No** | Config Profile |
| 10.15 (Catalina) | Yes | **No** | Config Profile |
| 11.0 (Big Sur) | Yes | **No** | Config Profile |
| 12.0 (Monterey) | Yes | **No** | Config Profile |
| 13.0 (Ventura) | Yes | **No** | Config Profile |
| 14.0 (Sonoma) | Yes | **No** | Config Profile |

**Key Takeaway:** On modern macOS (10.14+), **sudo is not enough!** You must use MDM to grant Full Disk Access.

## Alternative: Windows Deployment

**Good news:** Windows doesn't have TCC!

On Windows:
- ✅ Running as Administrator is sufficient
- ✅ No permission prompts
- ✅ No configuration profiles needed

Browser extension scanning works immediately on Windows with admin rights.

## Summary

### For Jamf Deployment (RECOMMENDED)

1. ✅ **Deploy Configuration Profile** granting Full Disk Access
2. ✅ **Install binary** to `/usr/local/bin/endpointbom`
3. ✅ **Schedule daily scans** via Jamf policy
4. ✅ **No user interaction required** - fully automated!

### For Quick Testing

```bash
# Option 1: Disable browsers (no TCC needed)
sudo endpointbom --disable=chrome-extensions,firefox-extensions,edge-extensions,safari-extensions

# Option 2: Scan current user only (one-time popup)
endpointbom --scan-all-users=false
```

### Documentation

- **Complete Jamf Guide:** `JAMF_DEPLOYMENT.md`
- **Security Details:** `SECURITY_IMPROVEMENTS.md`
- **General Usage:** `docs/USAGE.md`

---

**TL;DR:** Use Jamf Configuration Profile to grant Full Disk Access. No popups, fully automated! 🚀

**Status:** Production Ready with MDM  
**Last Updated:** 2025-12-13

