# macOS TCC - Quick Reference Card

## ✅ Default Behavior: No Popups!

**Browser scanners are DISABLED by default** - EndpointBOM runs without permission prompts.

## 🔍 Enabling Browser Scanning (Optional)

If you want browser extension data, you'll see this popup:

```
"endpointbom" would like to access files in your Chrome folder.
[Don't Allow] [OK]
```

**This breaks automated deployment if browser scanning is enabled!**

## ✅ Solution for Jamf

### 3-Step Fix

**1. Create Configuration Profile in Jamf**
```
Computers → Configuration Profiles → New
Add: Privacy Preferences Policy Control
```

**2. Configure Settings**
```
Identifier: /usr/local/bin/endpointbom
Identifier Type: Path
App or Service: SystemPolicyAllFiles
Access: Allow
```

**3. Deploy & Test**
```bash
# Deploy profile first, then binary
sudo endpointbom --verbose
# Should see: "Found XX components" with NO popup
```

## 📋 Quick Answers

### Q: Will running as root fix this?
**A: NO!** On macOS 10.14+, even root needs TCC permissions.

### Q: Can we skip the popup?
**A: YES!** Use Jamf Configuration Profile (above).

### Q: What if we can't use Jamf profiles?
**A: Disable browser scanning:**
```bash
endpointbom --disable=chrome-extensions,firefox-extensions,edge-extensions,safari-extensions
```

### Q: Does this affect Windows?
**A: NO!** Windows doesn't have TCC. Running as Administrator is sufficient.

### Q: Is Full Disk Access safe?
**A: YES!** EndpointBOM is read-only and has path validation. See `SECURITY_IMPROVEMENTS.md`.

## 🎯 Deployment Checklist

- [ ] Create TCC Configuration Profile in Jamf
- [ ] Set Identifier to `/usr/local/bin/endpointbom`
- [ ] Grant `SystemPolicyAllFiles` access
- [ ] Deploy profile to test group
- [ ] Wait 1-2 hours for profile to apply
- [ ] Test: `sudo endpointbom --verbose`
- [ ] Verify: No popup appears
- [ ] Deploy to production fleet

## 📚 Full Documentation

- **Complete Jamf Guide:** `JAMF_DEPLOYMENT.md`
- **TCC Details:** `MACOS_TCC_PERMISSIONS.md`
- **Security Info:** `SECURITY_IMPROVEMENTS.md`

## 🚀 Alternative: Disable Browsers

If TCC is not an option:

```yaml
# endpointbom.yaml
scanners:
  disabled:
    - chrome-extensions
    - firefox-extensions
    - edge-extensions
    - safari-extensions
```

**Trade-off:** Loses browser extension visibility, but no TCC needed.

---

**TL;DR:** Use Jamf Configuration Profile to grant Full Disk Access. Problem solved! ✅

