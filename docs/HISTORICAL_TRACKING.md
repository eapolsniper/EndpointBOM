# Historical Package Tracking

## Overview

EndpointBOM includes **historical tracking** to capture packages that were installed in the recent past, even if they've since been removed. This is crucial for incident response when investigating compromised endpoints.

## Why Historical Tracking?

**Problem:** Traditional scans only see what's currently installed. If a malicious package was installed, used, and then removed before the scan, it won't be detected.

**Solution:** Historical tracking looks back through logs and metadata to find packages installed in the last N days (default: 30).

### Real-World Scenario

1. Attacker compromises npm package "popular-lib" version 2.1.5
2. Developer installs it on Monday: `npm install popular-lib@2.1.5`
3. Developer notices odd behavior, uninstalls it on Tuesday
4. EndpointBOM scan runs on Wednesday
5. **Without historical:** No trace of compromised package
6. **With historical:** Package appears with install date (Monday) marked as "historical"

## What's Tracked

### NPM (Node.js)
- **Source:** `~/.npm/_logs/*-debug*.log`
- **Data:** Package name, version, install date
- **Best Effort:** Parses install commands from debug logs

### Homebrew (macOS)  
- **Source:** `brew info --json=v2 --installed`  
- **Data:** Package name, version, exact install timestamp
- **Highly Accurate:** Homebrew tracks install dates natively!

### Other Package Managers
- **Status:** Can be added in future if logs are available
- **Fallback:** Filesystem timestamps on package directories (good enough!)

## Configuration

### Config File

```yaml
# How many days back to look (default: 30)
historical_lookback_days: 30

# Enable/disable historical tracking (default: true)
include_historical: true

# Include raw log files in zip (default: true)
include_raw_logs: true

# Create zip archive (default: true)
create_zip_archive: true
```

### Command-Line Flags

```bash
# Custom lookback period
sudo endpointbom --historical-days=60

# Disable historical tracking
sudo endpointbom --no-historical

# Don't include raw logs in zip
sudo endpointbom --no-raw-logs

# Don't create zip archive
sudo endpointbom --no-zip
```

## Output Format

### Current vs Historical

Components are marked with `install_type`:

**Current** (installed within last 3 days):
```json
{
  "name": "react",
  "version": "18.2.0",
  "properties": [
    {
      "name": "install_date",
      "value": "2025-12-13T10:30:00Z"
    },
    {
      "name": "install_type",
      "value": "current"
    },
    {
      "name": "source",
      "value": "npm_scan"
    }
  ]
}
```

**Historical** (installed more than 3 days ago):
```json
{
  "name": "malicious-lib",
  "version": "1.0.0",
  "properties": [
    {
      "name": "install_date",
      "value": "2025-12-01T14:20:00Z"
    },
    {
      "name": "install_type",
      "value": "historical"
    },
    {
      "name": "source",
      "value": "npm_log"
    },
    {
      "name": "log_file",
      "value": "2025-12-01-debug-0.log"
    }
  ]
}
```

### Zip Archive Structure

When `create_zip_archive: true`, a comprehensive archive is created:

```
hostname.20251213-150405.scan.zip
├── metadata.json                  # Scan metadata
├── sboms/
│   ├── hostname.timestamp.package-managers.cdx.json
│   ├── hostname.timestamp.applications.cdx.json
│   ├── hostname.timestamp.ide-extensions.cdx.json
│   └── hostname.timestamp.browser-extensions.cdx.json
└── logs/
    ├── npm/
    │   ├── 2025-12-01-debug-0.log
    │   ├── 2025-12-05-debug-0.log
    │   └── 2025-12-10-debug-0.log
    └── brew/
        └── (no logs - brew uses API)
```

**Note:** SBOMs are ALSO kept unzipped in `scans/` directory for easy viewing!

## Use Cases

### 1. Incident Response - Supply Chain Attack

**Scenario:** News breaks about compromised npm package "popular-utils@3.1.4"

**Action:**
```bash
# Scan with 60-day lookback
sudo endpointbom --historical-days=60

# Search all endpoints' SBOMs for the package
grep -r '"name": "popular-utils"' scans/*.cdx.json

# Check if it was installed (even if removed)
# Historical tracking will show install_date
```

**Result:** Found 3 endpoints that had it installed 2 weeks ago, now removed.

### 2. Code Exfiltration Investigation

**Scenario:** Source code leaked, timeline suggests attack 3 weeks ago

**Action:**
```bash
# Scan with extended lookback
sudo endpointbom --historical-days=30

# Review historical npm packages for suspicious installs
jq '.components[] | select(.properties[] | 
    select(.name == "install_type" and .value == "historical"))' \
    *.package-managers.cdx.json
```

**Result:** Identified a suspicious package installed 21 days ago with webRequest permissions.

### 3. Baseline Drift Detection

**Scenario:** Compare current state against historical baseline

**Action:**
```bash
# Weekly scan with historical tracking
sudo endpointbom --historical-days=7

# Compare against last week's scan
# Flag new packages not in baseline
```

**Result:** Detected unauthorized package installations during the week.

## Limitations

### Best Effort Approach

- ✅ **NPM:** Simple log parsing, may miss some installs
- ✅ **Homebrew:** Native support, very accurate
- ⚠️ **Others:** Not yet implemented (can add if needed)

### Log Availability

- Logs may rotate/delete after time
- Not all package managers keep detailed logs
- Some installs may not be logged (cached installs)

### No Uninstall Tracking

- We track when packages were **installed**
- We don't track when they were **removed**
- A package in historical logs may still be installed

## Privacy & Security

### What's in the Zip?

**Included:**
- CycloneDX SBOM files (same as unzipped in scans/)
- Relevant log files from npm, brew, etc.
- Scan metadata (hostname, date, config used)

**Excluded:**
- No secrets or environment variables
- No source code
- No credentials
- No shell history

### Log File Security

- Only logs within lookback period are included
- Logs are not modified or redacted
- Logs may contain paths, usernames, etc. (normal log content)
- Treat zip files as **internal security data**

### Disabling Features

Don't want certain features? Disable them:

```yaml
include_historical: false  # No historical tracking
include_raw_logs: false    # No logs in zip
create_zip_archive: false  # No zip file created
```

## Performance

### Impact on Scan Time

- **NPM logs:** +1-3 seconds (depends on log size)
- **Homebrew:** +2-4 seconds (API call)
- **Zip creation:** +1-2 seconds

**Total overhead:** ~5-10 seconds (acceptable for improved visibility)

### Log File Size

- NPM logs: Typically 10-50KB per log
- Total zip size: Usually < 5MB
- Large environments may have more logs

## Best Practices

### For Security Teams

1. **Enable Historical Tracking**
   ```yaml
   include_historical: true
   historical_lookback_days: 30
   ```

2. **Collect Zip Archives Centrally**
   ```bash
   # On each endpoint
   sudo endpointbom --output=/central/scans
   
   # Central server collects zips
   scp endpoint:/central/scans/*.zip ./incident-data/
   ```

3. **Review Historical Packages During Incidents**
   ```bash
   # Extract and review
   unzip hostname.scan.zip
   jq '.components[] | select(.properties[] | 
       select(.name == "install_type" and .value == "historical"))' \
       sboms/*.cdx.json
   ```

4. **Archive Zips for Compliance**
   - Keep 90-day rolling archive
   - Index by hostname and date
   - Enable quick searches during incidents

### For Developers

1. **Regular Scans**
   - Let automated scans run (Jamf policy)
   - Don't disable historical tracking without reason

2. **Incident Response**
   - If asked for scan data, provide the .scan.zip file
   - It contains full context for investigation

## Troubleshooting

### No Historical Data Found

**Possible causes:**
- Package manager doesn't keep logs
- Logs were rotated/deleted
- No packages installed in lookback period

**Solution:** This is normal - historical tracking is best-effort.

### Zip File Not Created

**Check:**
1. Is `create_zip_archive: true` in config?
2. Do you have write permissions in output directory?
3. Check for error messages in output

**Debug:**
```bash
sudo endpointbom --verbose --debug
```

### Large Zip Files

**If zips are too large:**
```yaml
include_raw_logs: false  # Disable log inclusion
historical_lookback_days: 7  # Reduce lookback period
```

## Future Enhancements

Potential additions (not yet implemented):

1. **Chocolatey (Windows):** Parse `chocolatey.log`
2. **PIP (Python):** Check dist-info timestamps
3. **Gem (Ruby):** Check gem specification timestamps
4. **Uninstall Tracking:** Track when packages were removed
5. **Diff Reports:** Compare against previous scan

---

**Status:** ✅ Implemented (MVP)  
**Supported:** NPM, Homebrew  
**Version:** 1.1.0+

