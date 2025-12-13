# Public IP Address Security

## Overview

EndpointBOM can optionally gather the public IP address of endpoints by querying external services. This feature includes strong security measures and can be disabled if not needed.

## Security Measures

### 1. Input Validation on Untrusted Services

All responses from external IP services undergo **strict validation**:

```go
// SECURITY: Strict validation of IP addresses from untrusted sources
func isValidPublicIP(ip string) bool {
    // Must be a valid IP address
    parsedIP := net.ParseIP(ip)
    if parsedIP == nil {
        return false
    }
    
    // SECURITY: Reject private/internal IP addresses
    // Public IP services should only return public IPs
    // This prevents potential SSRF or manipulation attacks
    if parsedIP.IsLoopback() || parsedIP.IsPrivate() || 
       parsedIP.IsLinkLocalUnicast() || parsedIP.IsLinkLocalMulticast() || 
       parsedIP.IsMulticast() {
        return false
    }
    
    // Additional checks:
    // - Maximum length validation (45 chars for IPv6)
    // - Character whitelist (only 0-9, a-f, A-F, ., :)
    // - Response size limit (100 bytes max)
    
    return true
}
```

### 2. Response Size Limiting

```go
// SECURITY: Limit response size to prevent memory exhaustion
// An IP address should be max 45 chars (IPv6 with colons)
// We allow 100 bytes to be safe, but prevent large responses
limitedReader := io.LimitReader(resp.Body, 100)
```

### 3. Multiple Service Fallback

Uses multiple trusted services for reliability:
- `https://api.ipify.org`
- `https://ifconfig.me/ip`
- `https://icanhazip.com`

If one service is compromised or returns invalid data, the tool tries the next service.

### 4. Timeout Protection

```go
client := &http.Client{
    Timeout: 5 * time.Second, // 5 second timeout
}
```

Prevents hanging on slow or malicious services.

### 5. Private IP Rejection

**Critical Security Feature:** The tool rejects any private/internal IP addresses from external services.

This prevents potential attacks where a malicious service tries to:
- Return internal IP addresses (SSRF)
- Leak internal network information
- Manipulate SBOM data with fake IPs

**Rejected IP Types:**
- Loopback (127.0.0.1, ::1)
- Private ranges (10.x.x.x, 192.168.x.x, 172.16-31.x.x)
- Link-local (169.254.x.x, fe80::)
- Multicast addresses

## Disabling Public IP Gathering

### Why Disable?

You may want to disable public IP gathering if:
- You don't trust external services
- Your security policy prohibits external network calls
- You're in an air-gapped environment
- You don't need public IP information

### How to Disable

**Option 1: Command Line Flag**

```bash
sudo endpointbom --disable-public-ip
```

**Option 2: Configuration File**

```yaml
# endpointbom.yaml
disable_public_ip: true
```

**Result:**
- No external network calls are made
- Public IP field in SBOM shows: `"disabled"`
- Local IPs are still collected (no external calls needed)

## What Gets Collected

### With Public IP Enabled (Default)

```json
{
  "metadata": {
    "component": {
      "properties": [
        {
          "name": "local_ip",
          "value": "192.168.1.100"
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

### With Public IP Disabled

```json
{
  "metadata": {
    "component": {
      "properties": [
        {
          "name": "local_ip",
          "value": "192.168.1.100"
        },
        {
          "name": "public_ip",
          "value": "disabled"
        }
      ]
    }
  }
}
```

### If External Services Unavailable

```json
{
  "metadata": {
    "component": {
      "properties": [
        {
          "name": "local_ip",
          "value": "192.168.1.100"
        },
        {
          "name": "public_ip",
          "value": "unavailable"
        }
      ]
    }
  }
}
```

## Security Validation Details

### Character Whitelist

Only these characters are allowed in IP addresses:
- Digits: `0-9`
- Hex (for IPv6): `a-f`, `A-F`
- Separators: `.` (IPv4), `:` (IPv6)

Any other characters cause the response to be rejected.

### Length Validation

- **IPv4**: Maximum 15 characters (`xxx.xxx.xxx.xxx`)
- **IPv6**: Maximum 39 characters (8 groups of 4 hex digits)
- **Safety margin**: 45 characters maximum allowed

Responses longer than 45 characters are rejected.

### Format Validation

The IP must:
1. Parse successfully with `net.ParseIP()`
2. Not be a private/internal address
3. Not contain unexpected characters
4. Be within size limits

## Threat Model

### Threats Mitigated

✅ **SSRF (Server-Side Request Forgery)**
- Rejects private IP addresses
- Prevents internal network scanning

✅ **Memory Exhaustion**
- Response size limited to 100 bytes
- Prevents large response attacks

✅ **Code Injection**
- Strict character whitelist
- No shell execution of IP values

✅ **Data Manipulation**
- Multiple service validation
- Format verification

✅ **Denial of Service**
- 5-second timeout per service
- Non-blocking operation
- Graceful failure

### Residual Risks

⚠️ **Service Compromise**
- If all three services are compromised simultaneously
- Mitigation: Use `--disable-public-ip` if concerned

⚠️ **Network Monitoring**
- External services can see your IP address
- Mitigation: This is inherent to the feature; disable if sensitive

## Best Practices

### For High-Security Environments

```yaml
# Recommended configuration for high-security environments
disable_public_ip: true  # Disable external calls
```

### For Air-Gapped Environments

```yaml
# Air-gapped configuration
disable_public_ip: true  # No internet access anyway
```

### For Standard Deployments

```yaml
# Default configuration (public IP enabled)
disable_public_ip: false
```

Public IP gathering is safe with the implemented security measures.

## Compliance Considerations

### Data Privacy

**What's Sent:**
- Nothing! The tool only makes GET requests to IP services
- No endpoint data is transmitted

**What's Received:**
- Only the public IP address (already visible to any website you visit)

### Network Security

**Outbound Connections:**
- HTTPS only (encrypted)
- To well-known IP services
- Can be blocked by firewall if desired

**Firewall Rules:**
If you want to allow public IP detection:
```
Allow outbound HTTPS to:
- api.ipify.org
- ifconfig.me
- icanhazip.com
```

## Verification

### Test Public IP Disabled

```bash
sudo endpointbom --disable-public-ip --verbose

# Should see:
# Public IP: disabled
```

### Test Public IP Enabled

```bash
sudo endpointbom --verbose

# Should see:
# Public IP: xxx.xxx.xxx.xxx
# (or "unavailable" if no internet)
```

### Verify No External Calls

```bash
# Monitor network while running with disabled flag
sudo tcpdump -i any host api.ipify.org &
sudo endpointbom --disable-public-ip

# Should see NO traffic to IP services
```

## Summary

**Public IP gathering is secure by default** with:
- ✅ Strong input validation
- ✅ Response size limiting
- ✅ Private IP rejection
- ✅ Multiple service fallback
- ✅ Timeout protection
- ✅ Can be disabled if not needed

**Disable if:**
- High-security environment
- Air-gapped network
- Security policy prohibits external calls
- Don't need public IP information

---

**Security Contact:** If you discover a security issue with public IP gathering, please report it responsibly via GitHub Security Advisories.

