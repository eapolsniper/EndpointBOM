# Network Information Collection

## Overview

EndpointBOM automatically collects network information to help identify and track which endpoint a scan came from. This is useful for:
- **Endpoint Identification**: Correlate scans with specific machines in your network
- **Network Inventory**: Know which network segments endpoints are on
- **Remote Work Tracking**: Identify endpoints connecting from different networks
- **Compliance**: Track which systems have been scanned

## What's Collected

### Local IP Addresses

**All non-loopback IPv4 and IPv6 addresses** are collected from the local machine.

**Examples:**
- `192.168.1.100` - Private network address
- `10.0.0.50` - Private network address
- `2001:0db8::1` - IPv6 address (link-local excluded)

**What's Excluded:**
- Loopback addresses (127.0.0.1, ::1)
- Link-local IPv6 addresses (fe80::)

### Public IP Address

The tool attempts to determine the **public IP address** by contacting external services.

**Services Used (in order):**
1. https://api.ipify.org
2. https://ifconfig.me/ip
3. https://icanhazip.com

**Behavior:**
- **Timeout**: 5 seconds per service
- **Fallback**: Tries multiple services for reliability
- **Offline**: If no internet, displays "unavailable" (doesn't fail the scan)
- **Firewall**: If blocked, displays "unavailable"

**Example:**
- `203.0.113.45` - Your organization's public IP

## Privacy Considerations

### What This Means

✅ **IP addresses are included in SBOM files** - These files contain network information for endpoint identification.

⚠️ **Public IP may reveal location** - Your organization's public IP can indicate general location/ISP.

### Best Practices

1. **Secure SBOM Files**: Treat SBOM files as internal inventory data
2. **Network Segmentation**: IP information helps identify network segments
3. **Access Control**: Limit who can access generated SBOM files
4. **Data Retention**: Define retention policies for SBOM files

## Technical Details

### Local IP Detection

```go
// Gets all network interfaces
addrs, err := net.InterfaceAddrs()

// Filters out:
// - Loopback (127.0.0.1, ::1)
// - Link-local IPv6 (fe80::)

// Includes:
// - All IPv4 addresses (including private ranges)
// - Global IPv6 addresses
```

### Public IP Detection

```go
// Non-blocking with 5-second timeout
client := &http.Client{
    Timeout: 5 * time.Second,
}

// Tries multiple services for reliability
// Returns "unavailable" if all fail (no error)
```

### Failures are Non-Fatal

If network detection fails:
- ✅ Scan continues normally
- ✅ Other data is still collected
- ✅ SBOM is still generated
- ⚠️ Network info shows as empty or "unavailable"

## SBOM Output Format

Network information appears in the CycloneDX metadata:

```json
{
  "metadata": {
    "component": {
      "type": "device",
      "name": "macbook-pro",
      "properties": [
        {
          "name": "os",
          "value": "darwin"
        },
        {
          "name": "logged_in_user",
          "value": "john"
        },
        {
          "name": "local_ip",
          "value": "192.168.1.100"
        },
        {
          "name": "local_ip",
          "value": "10.0.0.50"
        },
        {
          "name": "public_ip",
          "value": "203.0.113.45"
        }
      ]
    }
  }
}
```

## Verbose Output

When running with `--verbose`, network information is displayed:

```bash
$ sudo endpointbom --verbose

✓ Running with administrator privileges - scanning all user profiles

Gathering system information...
Hostname: macbook-pro
OS: darwin 13.5.2
Logged in users: [tim]
Local IP(s): [192.168.1.100 10.0.0.50]
Public IP: 203.0.113.45
Output directory: /Users/tim/scans

Running scanner: npm
...
```

## Disabling Network Collection

Currently, network information collection is always enabled. If you need to disable it:

**Option 1: Modify the code**
```go
// In internal/system/system.go, comment out:
// info.LocalIPs = localIPs
// info.PublicIP = publicIP
```

**Option 2: Remove from SBOM**
```go
// In internal/sbom/cyclonedx.go, comment out the network properties section
```

**Future Enhancement**: A `--no-network-info` flag could be added if there's demand.

## Troubleshooting

### No Public IP Detected

**Possible Causes:**
1. **No Internet Connection** - Normal, shows "unavailable"
2. **Firewall Blocking** - Outbound HTTPS blocked
3. **Proxy Required** - HTTP proxy not configured
4. **Service Down** - All three IP services unreachable

**Solution**: This is non-fatal. Scan continues with public IP marked as "unavailable".

### No Local IPs Detected

**Possible Causes:**
1. **Network Disconnected** - No active network interfaces
2. **Virtual Machine** - Network not properly configured
3. **Permissions Issue** - Unable to read network interfaces (rare)

**Solution**: Check network connectivity. If persistent, file an issue.

### Multiple Local IPs

This is **normal** if you have:
- Multiple network interfaces (Ethernet + WiFi)
- VPN connections
- Virtual network adapters (Docker, VMs)
- Both IPv4 and IPv6

All addresses are included for complete endpoint identification.

## Use Cases

### Enterprise Inventory

```
Scenario: Track 1000+ developer laptops across offices

Benefit:
- Identify which network each scan came from
- Correlate with VPN usage
- Track remote vs office endpoints
```

### Remote Work Compliance

```
Scenario: Ensure remote workers have required software

Benefit:
- Public IP shows general location
- Verify VPN usage (corporate IP ranges)
- Track endpoint distribution
```

### Network Segmentation

```
Scenario: Different security zones for dev/staging/prod

Benefit:
- Verify endpoints are on correct networks
- Detect misconfigured systems
- Audit network access
```

## Security Considerations

### Information Disclosure

⚠️ **IP addresses can reveal:**
- Network topology
- Geographic location (approximate)
- ISP/hosting provider
- VPN usage

✅ **Mitigations:**
- Store SBOM files securely
- Limit access to authorized personnel
- Encrypt during transmission
- Follow data retention policies

### Network Traffic

The tool makes **3 HTTPS requests max** to determine public IP:
- Minimal bandwidth (<1KB per request)
- Encrypted (HTTPS)
- No data sent (GET requests only)
- 5-second timeout per request

**Firewall Rules:**
If you want to allow public IP detection:
```
Allow outbound HTTPS to:
- api.ipify.org
- ifconfig.me
- icanhazip.com
```

## References

- **RFC 1918** - Private IPv4 Address Space
- **RFC 4193** - Unique Local IPv6 Addresses  
- **CycloneDX Property Format** - Used for metadata storage

---

**Note**: Network information collection is designed to be helpful for enterprise deployment while respecting privacy and failing gracefully if unavailable.

