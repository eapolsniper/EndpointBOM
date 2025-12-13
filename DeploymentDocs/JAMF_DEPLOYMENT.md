# Jamf Deployment Guide for EndpointBOM

## macOS Privacy Permissions (TCC) - Critical!

⚠️ **EndpointBOM requires Full Disk Access to scan browser extensions on macOS 10.14+**

Without this permission:
- Users will see popup prompts asking for permission
- Automated scans will fail to access browser data
- Browser extension scanning will be incomplete

## Solution 1: Grant Full Disk Access via Configuration Profile (RECOMMENDED)

### Step 1: Create Privacy Preferences Policy Control Profile

1. **In Jamf Pro Console:**
   - Navigate to: **Computers** → **Configuration Profiles** → **New**
   - Name: "EndpointBOM Full Disk Access"
   - Category: Security & Privacy

2. **Add Privacy Preferences Policy Control Payload:**
   - Click **+ Add** → **Privacy Preferences Policy Control**
   - Click **+ Configure** to add a new entry

3. **Configure the Entry:**
   ```
   Identifier: com.eapolsniper.endpointbom
   Identifier Type: Bundle ID
   Code Requirement: identifier "com.eapolsniper.endpointbom"
   App or Service: SystemPolicyAllFiles
   Access: Allow
   ```

4. **Deployment:**
   - Scope to target computers
   - Deploy before or alongside the binary

### Step 2: Get Code Requirement (For Signed Binaries)

If your binary is signed:
```bash
codesign -dr - /path/to/endpointbom
```

Example output:
```
designated => identifier "com.eapolsniper.endpointbom" and anchor apple generic
```

Use this exact string in the Configuration Profile.

### Complete Configuration Profile XML

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
            <string>GENERATE-UUID-HERE</string>
            <key>PayloadEnabled</key>
            <true/>
            <key>PayloadDisplayName</key>
            <string>EndpointBOM Privacy Access</string>
            <key>Services</key>
            <dict>
                <key>SystemPolicyAllFiles</key>
                <array>
                    <dict>
                        <key>Identifier</key>
                        <string>/usr/local/bin/endpointbom</string>
                        <key>IdentifierType</key>
                        <string>path</string>
                        <key>CodeRequirement</key>
                        <string>identifier "/usr/local/bin/endpointbom" and anchor apple generic</string>
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
    <string>EndpointBOM Full Disk Access</string>
    <key>PayloadIdentifier</key>
    <string>com.company.endpointbom</string>
    <key>PayloadType</key>
    <string>Configuration</string>
    <key>PayloadUUID</key>
    <string>GENERATE-UUID-HERE</string>
    <key>PayloadVersion</key>
    <integer>1</integer>
</dict>
</plist>
```

## Solution 2: Using Path-Based Access (For Unsigned Binaries)

If your binary is **not signed** (current state):

### Configuration Profile Settings:
```
Identifier: /usr/local/bin/endpointbom
Identifier Type: Path
App or Service: SystemPolicyAllFiles
Access: Allow
```

**Note:** Path-based access works but is less secure than bundle ID-based access.

## Solution 3: Manual Grant (Testing Only)

For testing before Jamf deployment:

1. **System Settings** → **Privacy & Security** → **Full Disk Access**
2. Click the **+** button
3. Navigate to `/usr/local/bin/` (press Cmd+Shift+G)
4. Select `endpointbom`
5. Enable the checkbox

⚠️ **This is manual - not suitable for fleet deployment!**

## Jamf Policy Configuration

### Complete Deployment Policy

```bash
#!/bin/bash

# EndpointBOM Jamf Deployment Script

# Install binary
curl -L https://github.com/eapolsniper/endpointbom/releases/latest/download/endpointbom-darwin-arm64 \
  -o /usr/local/bin/endpointbom

chmod +x /usr/local/bin/endpointbom

# Deploy configuration file (optional)
cat > /usr/local/etc/endpointbom.yaml <<EOF
output_dir: /var/log/endpointbom/scans
require_admin: true
scan_all_users: true
scanners:
  disabled: []
paths:
  exclude:
    - /System
    - /Library/Caches
EOF

# Create output directory
mkdir -p /var/log/endpointbom/scans
chmod 755 /var/log/endpointbom/scans

# Test run (will fail if TCC not granted)
/usr/local/bin/endpointbom --output=/var/log/endpointbom/scans

# Exit successfully
exit 0
```

### Schedule Daily Scans

**Jamf Policy Settings:**
- **Trigger:** Recurring Check-In
- **Frequency:** Once per day
- **Execution Frequency:** Ongoing

**Script:**
```bash
#!/bin/bash

# Run EndpointBOM scan
/usr/local/bin/endpointbom --output=/var/log/endpointbom/scans

# Upload to central server (optional)
# scp /var/log/endpointbom/scans/*.cdx.json user@server:/central/scans/

exit 0
```

## Verification

### Test TCC Access

```bash
# This will show if the tool has Full Disk Access
sudo /usr/local/bin/endpointbom --verbose --disable=npm,pip,brew

# Should see:
# Running scanner: chrome-extensions
#   Found XX components
# (No permission prompts!)
```

### Check if Permissions Work

```bash
# Simply run the tool and check the output
sudo endpointbom --verbose

# If browser scanners are enabled and TCC is working:
# Running scanner: chrome-extensions
#   Found XX components

# If not working:
# Running scanner: chrome-extensions
#   Found 0 components  (or permission errors)
```

## Alternative: Disable Browser Scanning

If TCC permissions cannot be granted, disable browser scanning:

```yaml
# /usr/local/etc/endpointbom.yaml
scanners:
  disabled:
    - chrome-extensions
    - firefox-extensions
    - edge-extensions
    - safari-extensions
```

Or via CLI:
```bash
endpointbom --disable=chrome-extensions,firefox-extensions,edge-extensions,safari-extensions
```

## Troubleshooting

### Issue: Prompts Still Appear

**Cause:** Configuration profile not applied correctly

**Solution:**
```bash
# Check if profile is installed
sudo profiles -P

# Look for your TCC profile
# If not found, reinstall via Jamf

# Force profile refresh
sudo profiles renew -type enrollment
```

### Issue: Permission Denied Errors

**Cause:** Running without sudo/root

**Solution:**
```bash
# Must run as root
sudo /usr/local/bin/endpointbom
```

### Issue: Some Extensions Not Found

**Cause:** Partial TCC access

**Solution:**
```bash
# Check verbose output
sudo /usr/local/bin/endpointbom --verbose --debug

# Look for specific errors
# Grant additional TCC permissions if needed
```

## macOS Version Compatibility

| macOS Version | TCC Required | Solution |
|--------------|--------------|----------|
| 10.13 (High Sierra) or earlier | No | Works with sudo |
| 10.14 (Mojave) | Yes | Configuration Profile |
| 10.15 (Catalina) | Yes | Configuration Profile |
| 11.0 (Big Sur) | Yes | Configuration Profile |
| 12.0 (Monterey) | Yes | Configuration Profile |
| 13.0 (Ventura) | Yes | Configuration Profile |
| 14.0 (Sonoma) | Yes | Configuration Profile |

## Best Practices

1. **Deploy TCC Profile First**
   - Install configuration profile before binary
   - Allow 5-10 minutes for profile to apply
   - Verify access before scheduling scans

2. **Test on Pilot Group**
   - Deploy to test group first
   - Verify no prompts appear
   - Check SBOM output quality

3. **Monitor Deployment**
   - Check Jamf logs for errors
   - Verify SBOM files are generated
   - Alert on missing browser extension data

4. **Code Signing (Recommended)**
   - Sign binary with Apple Developer certificate
   - Use bundle ID-based TCC access
   - More secure than path-based access

## Security Considerations

**Why Full Disk Access?**
- Browser extensions contain sensitive security data
- Chrome/Safari directories are TCC-protected
- Required to scan all user profiles

**Least Privilege Alternative:**
- Only scan current user (no Full Disk Access needed)
- Use `--scan-all-users=false` flag
- Trade-off: miss extensions from other users

**Risk Mitigation:**
- Binary is read-only
- No secrets are collected
- Path validation prevents sensitive file access
- See SECURITY_IMPROVEMENTS.md for details

## Support

**If deployment fails:**
1. Check Jamf policy logs
2. Verify TCC profile is installed
3. Test manual run with sudo
4. Check macOS version compatibility
5. Review `/var/log/endpointbom/` for errors

---

**Last Updated:** 2025-12-13  
**Tested On:** macOS 14.6 (Sonoma)  
**Jamf Version:** 10.x and 11.x

