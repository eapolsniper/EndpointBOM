# Security Policy

## Security Commitment

EndpointBOM is designed with security as a top priority. The tool is meant to scan endpoints for inventory purposes and does not collect, store, or transmit sensitive information.

## What We Don't Collect

The tool explicitly **does not collect**:
- API keys
- Bearer tokens
- OAuth tokens
- Environment variable values
- Passwords
- Private keys
- Certificates
- Database credentials
- Any other secrets

## What We Do Collect

The tool only collects:
- Package names and versions
- Application names and versions
- IDE extension/plugin names and versions
- MCP server command names (not arguments with secrets)
- File paths
- System metadata (hostname, OS version, logged-in usernames)

## Reporting a Vulnerability

If you discover a security vulnerability in EndpointBOM, please report it responsibly:

### DO:
- Email security@example.com with details
- Provide steps to reproduce if possible
- Allow reasonable time for a fix before public disclosure

### DON'T:
- Open a public GitHub issue for security vulnerabilities
- Exploit the vulnerability
- Disclose the vulnerability publicly before a fix is available

## Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity (critical issues will be prioritized)

## Security Best Practices

When using EndpointBOM:

1. **Run with appropriate permissions**: Only use admin/root when needed
2. **Review output**: Verify SBOM files don't contain unexpected data
3. **Secure SBOM files**: Treat output files as sensitive inventory data
4. **Keep updated**: Use the latest version for security fixes
5. **Audit configuration**: Review exclude paths and disabled scanners

## Dependency Security

EndpointBOM uses minimal dependencies:
- All dependencies are pinned to specific versions
- Dependencies are from trusted, well-maintained projects
- Regular dependency audits are performed

To audit dependencies yourself:
```bash
go list -m all
go mod verify
```

## Build Security

Official releases are:
- Built using GitHub Actions with reproducible builds
- Signed with GPG signatures
- Checksums provided for verification

Verify a release:
```bash
# Download checksum file
curl -L https://github.com/eapolsniper/endpointbom/releases/download/v1.0.0/checksums.txt -o checksums.txt

# Verify binary
shasum -a 256 -c checksums.txt
```

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

Only the latest major version receives security updates.

## Security Features

- **No network calls**: Tool operates entirely offline (unless scanning requires network for package managers)
- **Sandboxed execution**: Tool can run without network permissions
- **No telemetry**: Zero data collection or transmission by default
- **Local output**: SBOMs are written locally, never transmitted
- **Clear logging**: Debug mode shows exactly what the tool is doing

## Known Limitations

1. **Admin privileges**: Some scans require elevated permissions. Use with caution.
2. **File system access**: Tool reads configuration files and installed packages. Ensure it's not run in sensitive directories.

## Compliance

EndpointBOM is designed to be compliant with:
- SBOM standards (CycloneDX)
- Enterprise security policies
- GDPR (no personal data collection)
- SOC 2 requirements (audit logging available)

## Updates

This security policy is reviewed and updated quarterly. Last update: [Current Date]

---

For questions about this security policy, contact: security@example.com

