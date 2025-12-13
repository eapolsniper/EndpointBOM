# Browser Extension Security Scanning

## Overview

Browser extensions have extensive privileges and can access:
- All browsing data and cookies
- Passwords and credentials
- Source code in browser-based IDEs
- API keys in web applications
- Network requests and responses
- Clipboard data
- File system access (with permissions)

**Malicious browser extensions** are a significant attack vector for:
- Credential theft
- Code exfiltration
- Supply chain attacks (injecting malicious code)
- Cryptojacking
- Data exfiltration

EndpointBOM now scans all major browsers for installed extensions to help identify compromised or suspicious extensions during incident response.

## Supported Browsers

### ✅ Google Chrome
- **Locations Scanned:**
  - macOS: `~/Library/Application Support/Google/Chrome/*/Extensions`
  - Windows: `%LOCALAPPDATA%\Google\Chrome\User Data\*\Extensions`
  - Linux: `~/.config/google-chrome/*/Extensions`
- **Data Captured:**
  - Extension name and ID
  - Version
  - Permissions (tabs, webRequest, storage, etc.)
  - Host permissions (which sites it can access)
  - Manifest version

### ✅ Mozilla Firefox  
- **Locations Scanned:**
  - macOS: `~/Library/Application Support/Firefox/Profiles/*/extensions`
  - Windows: `%APPDATA%\Mozilla\Firefox\Profiles\*\extensions`
  - Linux: `~/.mozilla/firefox/*/extensions`
- **Data Captured:**
  - Add-on name and ID
  - Version
  - Permissions
  - Profile association

### ✅ Microsoft Edge
- **Locations Scanned:**
  - macOS: `~/Library/Application Support/Microsoft Edge/*/Extensions`
  - Windows: `%LOCALAPPDATA%\Microsoft\Edge\User Data\*\Extensions`
  - Linux: `~/.config/microsoft-edge/*/Extensions`
- **Data Captured:**
  - Extension name and ID
  - Version
  - Permissions
  - Host permissions
  - Manifest version

### ✅ Safari (macOS only)
- **Locations Scanned:**
  - `~/Library/Safari/Extensions`
  - `~/Library/Containers` (for modern Safari App Extensions)
- **Data Captured:**
  - Extension name
  - Version (when available)
  - Extension type (legacy vs modern)

## What's Captured

### Extension Metadata
```json
{
  "type": "browser-extension",
  "name": "Example Extension",
  "version": "1.2.3",
  "description": "Does something useful",
  "properties": {
    "browser": "chrome",
    "extension_id": "abcdefghijklmnop",
    "manifest_version": "3",
    "permissions": "tabs, storage, webRequest",
    "host_permissions": "*://*.example.com/*, *://*.github.com/*"
  }
}
```

### Security-Relevant Information

**Permissions** - What the extension can do:
- `tabs` - Access to all tabs
- `webRequest` - Intercept/modify network requests
- `storage` - Store data locally
- `cookies` - Access cookies
- `history` - Browse history access
- `downloads` - File download management
- `<all_urls>` - Access to all websites

**Host Permissions** - Which sites it can access:
- `*://*/*` - ALL websites (major red flag)
- `*://*.github.com/*` - All GitHub pages
- `*://localhost/*` - Local development servers

## Use Cases

### 1. Incident Response - Malicious Extension Detection

**Scenario:** A developer's GitHub credentials were stolen

**Investigation:**
```bash
sudo endpointbom --verbose

# Review the ide-extensions.cdx.json
# Look for:
# - Extensions with broad host permissions (*://*/*)
# - Extensions with webRequest + all_urls
# - Unknown/suspicious extension names
# - Recently installed extensions
```

**Red Flags:**
- Extension with access to `*://github.com/*` + `webRequest`
- Extension with `cookies` + `*://*/*`
- Extension names similar to legitimate ones (typosquatting)

### 2. Supply Chain Attack - Compromised Extension

**Scenario:** Popular dev tools extension was compromised (similar to CoPilot incidents)

**Response:**
```bash
# Scan all developer endpoints
sudo endpointbom --output=/central/sboms

# Query all SBOMs for the specific extension ID
grep -r "extension_id.*abcdefghijk" /central/sboms/

# Identify affected endpoints
# Generate remediation list
```

### 3. Code Exfiltration Detection

**Scenario:** Source code leaked, suspect browser extension

**Investigation Checklist:**
- Extensions with `webRequest` on `*://localhost/*`
- Extensions with clipboard access
- Extensions with broad permissions installed recently
- Unknown extensions from untrusted sources

### 4. Cryptojacking Detection

**Scenario:** Developer machines showing high CPU usage

**Investigation:**
- Check for unknown extensions
- Extensions installed outside corporate policies
- Extensions with `background` permission (persistent execution)

## Security Analysis Workflow

### Step 1: Scan Endpoints
```bash
sudo endpointbom
```

### Step 2: Extract Extension Data
```bash
# From the generated ide-extensions.cdx.json
jq '.components[] | select(.type == "browser-extension")' *.cdx.json
```

### Step 3: Identify High-Risk Extensions

**Look for:**
1. **Overly Broad Permissions**
   ```json
   "host_permissions": "*://*/*"
   "permissions": "webRequest, cookies, tabs, <all_urls>"
   ```

2. **Suspicious Names**
   - Similar to legitimate extensions (typosquatting)
   - Generic names like "Helper" or "Extension"
   - Random character strings

3. **Unknown Extensions**
   - Not in your corporate approved list
   - Not in Chrome Web Store / Firefox Add-ons
   - Side-loaded extensions

4. **Dangerous Permission Combinations**
   - `webRequest` + `*://*/*` = Can intercept all traffic
   - `cookies` + `*://*/*` = Can steal all cookies
   - `tabs` + `activeTab` + broad hosts = Can inject code anywhere

### Step 4: Cross-Reference with Threat Intelligence

Check extension IDs against:
- Known malicious extension databases
- CRXcavator (Chrome extension risk scoring)
- Your organization's blocklist
- Recent security advisories

### Step 5: Create Baseline

```bash
# Known-good endpoint scan
sudo endpointbom --output=./baselines/

# Label as "clean baseline"
mv baseline.cdx.json clean-endpoint-baseline.cdx.json
```

### Step 6: Compare Against Baseline

```bash
# Scan potentially compromised endpoint
sudo endpointbom --output=./scans/

# Compare (manual or scripted)
# Identify new extensions not in baseline
```

## Common Malicious Extension Indicators

### 🚩 Red Flags

1. **Excessive Permissions**
   - Requests more permissions than functionality requires
   - `<all_urls>` without clear justification

2. **Obfuscated Code**
   - Extension manifest.json has minimal info
   - Cannot find extension in official stores

3. **Network Behavior**
   - Permissions to intercept requests (`webRequest`)
   - Permissions for specific financial/banking sites

4. **Persistence Mechanisms**
   - `background` permissions
   - `alarms` or `idle` permissions

5. **Data Access**
   - `storage` + broad host permissions
   - `clipboardRead` / `clipboardWrite`
   - `downloads` + suspicious hosts

### ⚠️ Warnings

- Extensions from unknown developers
- Extensions with very few users
- Extensions recently updated with new permissions
- Side-loaded extensions (not from stores)
- Extensions requiring "Developer mode" enabled

## SBOM Output Format

Browser extensions now have their own dedicated SBOM file: `{hostname}.{timestamp}.browser-extensions.cdx.json`

This separation makes it easier to:
- Analyze browser extensions independently
- Process browser-specific security alerts
- Compare browser extension baselines
- Integrate with browser security tools

```json
{
  "components": [
    {
      "type": "browser-extension",
      "name": "Example Developer Tools",
      "version": "2.3.1",
      "description": "Helps developers code faster",
      "properties": [
        {
          "name": "browser",
          "value": "chrome"
        },
        {
          "name": "extension_id",
          "value": "abcdefghijklmnop"
        },
        {
          "name": "permissions",
          "value": "tabs, storage, webRequest"
        },
        {
          "name": "host_permissions",
          "value": "*://github.com/*, *://localhost/*"
        }
      ]
    }
  ]
}
```

## Limitations

### Firefox
- Only scans unpacked extensions in the profile directory
- `.xpi` files (packed extensions) are not extracted/scanned
- Extension store data not included

### Safari
- macOS only
- Modern Safari App Extensions may not have complete metadata
- Legacy extensions more fully supported

### All Browsers
- Only scans currently installed extensions
- Does not track historical installations
- Does not capture when extension was installed (filesystem timestamps may help)
- Does not scan browser profiles that aren't on disk

## Best Practices

### For Security Teams

1. **Maintain Approved Extension List**
   ```yaml
   approved_extensions:
     chrome:
       - abcdefghijklmnop  # Extension ID
       - bcdefghijklmnopq
     firefox:
       - extension@developer.mozilla.org
   ```

2. **Regular Scanning**
   - Daily automated scans
   - Store results centrally
   - Alert on new/unknown extensions

3. **Baseline Comparison**
   - Establish known-good baselines
   - Flag deviations from baseline
   - Review permission changes

4. **Policy Enforcement**
   - Block extension installation via GPO/MDM
   - Require approval for new extensions
   - Remove unauthorized extensions automatically

### For Developers

1. **Review Extension Permissions**
   - Only install extensions you trust
   - Review permissions before installing
   - Remove unused extensions

2. **Source Verification**
   - Install from official stores only
   - Check developer reputation
   - Read extension reviews

3. **Regular Audits**
   - Quarterly review of installed extensions
   - Remove unnecessary extensions
   - Update to latest versions

## Integration with Other Tools

### SIEM Integration
```bash
# Parse SBOM and send to SIEM
jq '.components[] | select(.type == "browser-extension")' *.cdx.json | \
  curl -X POST -H "Content-Type: application/json" \
       -d @- https://siem.example.com/api/extensions
```

### Ticket Creation
```bash
# Find high-risk extensions and create tickets
jq '.components[] | select(.type == "browser-extension" and 
    (.properties[] | select(.name == "host_permissions" and 
    contains("*://*/*"))))' *.cdx.json
```

### Reporting
```bash
# Generate summary report
echo "Extension Summary:"
jq -r '.components[] | select(.type == "browser-extension") | 
       "\(.properties[] | select(.name == "browser") | .value): \(.name)"' \
       *.cdx.json | sort | uniq -c
```

## Real-World Attack Examples

### Example 1: Octopus Scanner (2020)
- **Target:** Developer extensions
- **Method:** Compromised legitimate extensions
- **Impact:** Stole source code from GitHub repos
- **Detection:** Extensions with GitHub access + webRequest

### Example 2: CryptBot Malware
- **Target:** Cryptocurrency wallets
- **Method:** Fake wallet extensions
- **Impact:** Stole crypto credentials
- **Detection:** Extensions with broad storage + clipboard access

### Example 3: Developer Tool Impersonation
- **Target:** Web developers
- **Method:** Typosquatting popular tools
- **Impact:** Injected malicious code into builds
- **Detection:** Similar names to legitimate extensions

## References

- [Chrome Extension Security Best Practices](https://developer.chrome.com/docs/extensions/mv3/security/)
- [Firefox Extension Security](https://extensionworkshop.com/documentation/develop/build-a-secure-extension/)
- [OWASP Browser Extension Security](https://owasp.org/www-community/vulnerabilities/Browser_extension_vulnerabilities)
- [CRXcavator - Chrome Extension Risk Scoring](https://crxcavator.io/)

---

**Note:** Browser extension scanning is designed to help security teams identify potentially malicious or compromised extensions during incident response. Always cross-reference findings with threat intelligence and organizational security policies.

