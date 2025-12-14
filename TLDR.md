# EndpointBOM - Quick Start

## Install

```bash
# Download latest release
curl -L https://github.com/eapolsniper/EndpointBOM/releases/latest/download/endpointbom_$(uname -s)_$(uname -m).tar.gz -o endpointbom.tar.gz

# Extract
tar -xzf endpointbom.tar.gz


## Run Scan

```bash
# Basic scan (Recommended for most use cases)
./endpointbom

## Common Options

```bash
# Scan with all browsers, shouldn't cause a permissions popup but it might, so disabled by default
./endpointbom --enable browser-extensions

# Enable public IP lookup, generates external service interaction
./endpointbom --fetch-public-ip

# Debug output
./endpointbom --verbose --debug
```
## Enable Browser Extensions (Optional)

```bash
# macOS: Grant Full Disk Access first
# System Settings → Privacy & Security → Full Disk Access → Add endpointbom

# Run with browser extensions enabled
./endpointbom --enable browser-extensions
```

# Output: scans/*.cdx.json and scans/*.zip
```

## Output Files

- `scans/hostname.timestamp.package-managers.cdx.json` - npm, pip, brew, etc.
- `scans/hostname.timestamp.applications.cdx.json` - Installed apps
- `scans/hostname.timestamp.ide-extensions.cdx.json` - VSCode, Cursor, etc.
- `scans/hostname.timestamp.browser-extensions.cdx.json` - Chrome, Firefox, etc. (disabled by default)
- `scans/hostname.timestamp.scan.zip` - All files archived, including log files from package managers

Note: The .zip file is not actually used by Dependency-track or any other tool. These are package manager logs and the json files for investigation purposes later if needed. You can either ignore these, or upload them to a server somewhere and store for a period of time incase they're needed. 

## Upload to Dependency-Track (Optional)

### Option 1: Python Script (Simple, requires python, API key is exposed but permissions are limited)

```bash
cd Deployment

# Edit UploadToDependencyTrack.py - set these:
DEPENDENCY_TRACK_URL = "https://your-dtrack-url.com"
API_KEY = "odt_your_api_key_here"

# Run
python3 UploadToDependencyTrack.py
```

### Option 2: Go Binary (No Dependencies)

```bash
cd Deployment

# Build with your credentials
./build-all-platforms.sh https://your-dtrack-url.com odt_your_api_key_here

# Run
./dist/dt-upload-linux-amd64
```

## Create Dependency-Track API Key

1. Login to Dependency-Track
2. Go to **Settings** → **Access Management** → **Teams**
3. Select or create a team
4. Add permissions:
   - `BOM_UPLOAD`
   - `PROJECT_CREATION_UPLOAD`
   - `VIEW_PORTFOLIO`
5. Click **API Keys** → **Generate New Key**
6. Copy the key (starts with `odt_`)

## That's It!

For detailed docs, see `README.md` or `Deployment/README.md`

