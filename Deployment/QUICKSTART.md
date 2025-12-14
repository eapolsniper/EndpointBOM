# Quick Start Guide

## For Organizations: Building Your Custom Binary

### Step 1: Install Go (if not already installed)

```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Windows
# Download from https://go.dev/dl/ and run installer
```

### Step 2: Build Your Binaries

```bash
cd Deployment

# Pass your Dependency-Track URL and API key to the build script
./build-all-platforms.sh https://dtrack.yourcompany.com odt_AbCdEfGh123456789
```

Or use environment variables:

```bash
DT_URL=https://dtrack.yourcompany.com \
DT_API_KEY=odt_AbCdEfGh123456789 \
./build-all-platforms.sh
```

This creates binaries for all platforms in the `dist/` directory:
- `dt-upload-linux-amd64`
- `dt-upload-macos-arm64`
- `dt-upload-macos-amd64`
- `dt-upload-windows-amd64.exe`
- And more...

### Step 3: Distribute to Developers

Copy the appropriate binary to your developers:

```bash
# Via file share
cp dist/dt-upload-windows-amd64.exe \\fileserver\tools\security\

# Via MDM (Jamf, Intune, etc.)
# Upload to your MDM console and deploy

# Via internal package repository
# Add to your internal package repo
```

---

## For Developers: Using the Binary

### Basic Usage

```bash
# From your EndpointBOM project root
# (The tool expects a scans/ directory)
./dt-upload-linux-amd64

# Or specify a custom directory
./dt-upload-linux-amd64 /path/to/scans
```

### What Happens

The tool will:
1. Find all `*.cdx.json` files in the scans directory
2. Group them by hostname
3. Create a parent "DEVICE" project in Dependency-Track
4. Create child projects for each SBOM type
5. Upload all SBOMs
6. Monitor processing status
7. Display results with a link to view in Dependency-Track

**Example output:**
```
📦 Creating project: Tims-MacBook-Pro.local v latest
   ✅ Created project UUID: abc123...

📤 Uploading BOM: package-managers.cdx.json
   ✅ BOM uploaded successfully

✅ Successfully uploaded 4 BOMs for Tims-MacBook-Pro.local
🔗 View in Dependency-Track:
   https://dtrack.yourcompany.com/projects/abc123...
```

---

## Troubleshooting

### "No SBOM files found"

Make sure you're in the right directory:

```bash
# Check where you are
pwd

# Should show something like: /Users/yourname/EndpointBOM/EndpointBOM

# Check if scans directory exists
ls -la scans/

# If you're somewhere else, either:
# 1. cd to the right place
# 2. Or specify the scans directory:
./dt-upload /full/path/to/scans
```

### "Failed to create project: 401 Unauthorized"

Your API key is invalid or expired.

**Fix:**
1. Log into Dependency-Track
2. Go to **Settings → Access Management → Teams**
3. Find or create a team with permissions:
   - `BOM_UPLOAD`
   - `PROJECT_CREATION_UPLOAD`
   - `VIEW_PORTFOLIO`
4. Generate a new API key
5. Rebuild the binary with the new key

### "Connection refused"

Dependency-Track is not running or the URL is wrong.

**Fix:**
```bash
# Test the connection
curl https://dtrack.yourcompany.com/api/version

# Should return something like:
# {"version":"4.11.0"}

# If it fails, check:
# - Is Dependency-Track running?
# - Is the URL correct?
# - Is there a firewall blocking the connection?
# - Do you need to be on the corporate VPN?
```

### macOS "cannot be opened because the developer cannot be verified"

macOS Gatekeeper blocks unsigned binaries.

**Fix:**
```bash
# Option 1: Remove quarantine attribute
xattr -d com.apple.quarantine dt-upload-macos-arm64

# Option 2: Allow in System Settings
# Right-click the binary → Open → Click "Open Anyway"
```

### Windows SmartScreen Warning

Windows may warn about running an "unrecognized app".

**Fix:**
1. Click "More info"
2. Click "Run anyway"

**For IT admins**: Sign the binary with your organization's code signing certificate to avoid this warning.

---

## Integration with CI/CD

### GitHub Actions

```yaml
name: Upload to Dependency-Track

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM

jobs:
  upload:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      
      - name: Download dt-upload tool
        run: |
          # Download your organization's pre-built binary
          curl -O https://internal-server.com/tools/dt-upload-linux-amd64
          chmod +x dt-upload-linux-amd64
      
      - name: Run EndpointBOM scan
        run: ./bin/endpointbom
      
      - name: Upload to Dependency-Track
        run: ./dt-upload-linux-amd64
```

### Jenkins

```groovy
pipeline {
    agent any
    
    stages {
        stage('Scan') {
            steps {
                sh './bin/endpointbom'
            }
        }
        
        stage('Upload to Dependency-Track') {
            steps {
                sh './dt-upload-linux-amd64'
            }
        }
    }
}
```

---

## Security Best Practices

1. **Rotate API Keys Regularly** (every 90 days)
2. **Use Team-Specific Keys** (don't share one key across all teams)
3. **Limit Permissions** (only grant necessary permissions)
4. **Monitor Usage** (check Dependency-Track audit logs)
5. **Use HTTPS** (never use http:// in production)
6. **Restrict Binary Distribution** (only share with authorized developers)

---

## Support

- **EndpointBOM Issues**: GitHub repository
- **Dependency-Track Issues**: https://docs.dependencytrack.org/
- **Build Issues**: Check Go installation with `go version`

