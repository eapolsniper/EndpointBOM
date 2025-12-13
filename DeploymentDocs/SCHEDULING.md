# Scheduling EndpointBOM to Run Daily

## Overview

EndpointBOM should run daily (or on your desired schedule) to maintain up-to-date inventory of endpoint software. This guide covers automated scheduling for each operating system.

## macOS

### Option 1: LaunchDaemon (System-Wide, Recommended)

LaunchDaemons run as root and can scan all user profiles.

**Step 1: Create the LaunchDaemon plist**

Create `/Library/LaunchDaemons/com.endpointbom.scan.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.endpointbom.scan</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/endpointbom</string>
        <string>--output</string>
        <string>/var/log/endpointbom/scans</string>
        <string>--verbose</string>
    </array>
    
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>2</integer>
        <key>Minute</key>
        <integer>0</integer>
    </dict>
    
    <key>StandardOutPath</key>
    <string>/var/log/endpointbom/stdout.log</string>
    
    <key>StandardErrorPath</key>
    <string>/var/log/endpointbom/stderr.log</string>
    
    <key>RunAtLoad</key>
    <false/>
    
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    </dict>
</dict>
</plist>
```

**Step 2: Create log directory**

```bash
sudo mkdir -p /var/log/endpointbom/scans
sudo chown root:wheel /var/log/endpointbom
sudo chmod 755 /var/log/endpointbom
```

**Step 3: Set permissions and load**

```bash
sudo chown root:wheel /Library/LaunchDaemons/com.endpointbom.scan.plist
sudo chmod 644 /Library/LaunchDaemons/com.endpointbom.scan.plist
sudo launchctl load /Library/LaunchDaemons/com.endpointbom.scan.plist
```

**Step 4: Verify**

```bash
# Check if loaded
sudo launchctl list | grep endpointbom

# Test run immediately (without waiting for schedule)
sudo launchctl start com.endpointbom.scan

# Check logs
tail -f /var/log/endpointbom/stdout.log
```

**Unload/Stop:**
```bash
sudo launchctl unload /Library/LaunchDaemons/com.endpointbom.scan.plist
```

### Option 2: LaunchAgent (Current User Only)

For non-admin users, use LaunchAgent in `~/Library/LaunchAgents/`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.endpointbom.scan.user</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/endpointbom</string>
        <string>--scan-all-users=false</string>
        <string>--output</string>
        <string>~/endpointbom/scans</string>
    </array>
    
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>2</integer>
        <key>Minute</key>
        <integer>0</integer>
    </dict>
    
    <key>RunAtLoad</key>
    <false/>
</dict>
</plist>
```

Load as current user:
```bash
launchctl load ~/Library/LaunchAgents/com.endpointbom.scan.user.plist
```

### Schedule Options

The `StartCalendarInterval` key supports various schedules:

**Daily at 2am:**
```xml
<key>StartCalendarInterval</key>
<dict>
    <key>Hour</key>
    <integer>2</integer>
    <key>Minute</key>
    <integer>0</integer>
</dict>
```

**Every 6 hours:**
```xml
<key>StartInterval</key>
<integer>21600</integer>
```

**Weekly (Monday at 2am):**
```xml
<key>StartCalendarInterval</key>
<dict>
    <key>Weekday</key>
    <integer>1</integer>
    <key>Hour</key>
    <integer>2</integer>
    <key>Minute</key>
    <integer>0</integer>
</dict>
```

**Multiple times per day:**
```xml
<key>StartCalendarInterval</key>
<array>
    <dict>
        <key>Hour</key>
        <integer>2</integer>
    </dict>
    <dict>
        <key>Hour</key>
        <integer>14</integer>
    </dict>
</array>
```

## Windows

### Option 1: Scheduled Task via PowerShell (Recommended)

**Step 1: Create scheduled task**

Save as `install-endpointbom-task.ps1`:

```powershell
# Create scheduled task for EndpointBOM daily scan

# Define task parameters
$taskName = "EndpointBOM Daily Scan"
$taskDescription = "Scans endpoint for installed software and generates SBOM files"
$exePath = "C:\Program Files\EndpointBOM\endpointbom.exe"
$arguments = "--output `"C:\ProgramData\EndpointBOM\scans`" --verbose"
$logPath = "C:\ProgramData\EndpointBOM\logs"

# Create log directory
New-Item -ItemType Directory -Force -Path $logPath
New-Item -ItemType Directory -Force -Path "C:\ProgramData\EndpointBOM\scans"

# Create scheduled task action
$action = New-ScheduledTaskAction -Execute $exePath -Argument $arguments

# Create daily trigger at 2am
$trigger = New-ScheduledTaskTrigger -Daily -At 2am

# Run as SYSTEM with highest privileges
$principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -RunLevel Highest

# Task settings
$settings = New-ScheduledTaskSettingsSet `
    -StartWhenAvailable `
    -DontStopOnIdleEnd `
    -AllowStartIfOnBatteries `
    -DontStopIfGoingOnBatteries `
    -ExecutionTimeLimit (New-TimeSpan -Hours 2)

# Register the task
Register-ScheduledTask `
    -TaskName $taskName `
    -Description $taskDescription `
    -Action $action `
    -Trigger $trigger `
    -Principal $principal `
    -Settings $settings `
    -Force

Write-Host "Scheduled task created successfully!"
Write-Host "Task will run daily at 2:00 AM"
Write-Host "Logs: $logPath"
```

**Step 2: Run as Administrator**

```powershell
# Run PowerShell as Administrator
.\install-endpointbom-task.ps1
```

**Step 3: Verify**

```powershell
# Check if task exists
Get-ScheduledTask -TaskName "EndpointBOM Daily Scan"

# Run immediately (test)
Start-ScheduledTask -TaskName "EndpointBOM Daily Scan"

# Check task history
Get-ScheduledTask -TaskName "EndpointBOM Daily Scan" | Get-ScheduledTaskInfo
```

**Uninstall:**
```powershell
Unregister-ScheduledTask -TaskName "EndpointBOM Daily Scan" -Confirm:$false
```

### Option 2: Task Scheduler GUI

1. Open Task Scheduler (`taskschd.msc`)
2. Click "Create Task" (not "Create Basic Task")
3. **General Tab:**
   - Name: `EndpointBOM Daily Scan`
   - Description: `Scans endpoint for installed software`
   - User: `SYSTEM`
   - ✅ Run whether user is logged on or not
   - ✅ Run with highest privileges
4. **Triggers Tab:**
   - New → Daily
   - Start: 2:00:00 AM
   - Recur every: 1 days
5. **Actions Tab:**
   - Action: Start a program
   - Program: `C:\Program Files\EndpointBOM\endpointbom.exe`
   - Arguments: `--output "C:\ProgramData\EndpointBOM\scans" --verbose`
6. **Conditions Tab:**
   - ✅ Start only if computer is on AC power (optional)
   - ✅ Wake the computer to run this task
7. **Settings Tab:**
   - ✅ Allow task to be run on demand
   - Stop task if runs longer than: 2 hours
   - If running task does not end: Stop existing instance
8. Click OK

### Option 3: Command Line (schtasks)

```cmd
schtasks /create ^
  /tn "EndpointBOM Daily Scan" ^
  /tr "\"C:\Program Files\EndpointBOM\endpointbom.exe\" --output \"C:\ProgramData\EndpointBOM\scans\"" ^
  /sc daily ^
  /st 02:00 ^
  /ru SYSTEM ^
  /rl highest ^
  /f
```

## Linux

### Option 1: Cron (Most Common)

**Step 1: Edit root's crontab**

```bash
sudo crontab -e
```

**Step 2: Add cron entry**

```cron
# Run EndpointBOM daily at 2am
0 2 * * * /usr/local/bin/endpointbom --output /var/log/endpointbom/scans >> /var/log/endpointbom/cron.log 2>&1

# Or with verbose output
0 2 * * * /usr/local/bin/endpointbom --output /var/log/endpointbom/scans --verbose >> /var/log/endpointbom/cron.log 2>&1
```

**Step 3: Create log directory**

```bash
sudo mkdir -p /var/log/endpointbom/scans
sudo chown root:root /var/log/endpointbom
sudo chmod 755 /var/log/endpointbom
```

**Verify:**
```bash
# List cron jobs
sudo crontab -l

# Check cron logs
grep endpointbom /var/log/syslog  # Debian/Ubuntu
grep endpointbom /var/log/cron     # RHEL/CentOS
```

### Option 2: Systemd Timer (Modern Linux)

**Step 1: Create service file**

Create `/etc/systemd/system/endpointbom.service`:

```ini
[Unit]
Description=EndpointBOM Scan
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/endpointbom --output /var/log/endpointbom/scans --verbose
StandardOutput=append:/var/log/endpointbom/stdout.log
StandardError=append:/var/log/endpointbom/stderr.log
User=root
Group=root

[Install]
WantedBy=multi-user.target
```

**Step 2: Create timer file**

Create `/etc/systemd/system/endpointbom.timer`:

```ini
[Unit]
Description=EndpointBOM Daily Scan Timer
Requires=endpointbom.service

[Timer]
OnCalendar=daily
OnCalendar=02:00
Persistent=true

[Install]
WantedBy=timers.target
```

**Step 3: Enable and start timer**

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable timer (start on boot)
sudo systemctl enable endpointbom.timer

# Start timer now
sudo systemctl start endpointbom.timer

# Check status
sudo systemctl status endpointbom.timer

# List timers
sudo systemctl list-timers endpointbom*

# Run service immediately (test)
sudo systemctl start endpointbom.service

# Check logs
sudo journalctl -u endpointbom.service -f
```

### Cron Schedule Examples

```cron
# Every day at 2am
0 2 * * * /usr/local/bin/endpointbom

# Every 6 hours
0 */6 * * * /usr/local/bin/endpointbom

# Every Monday at 2am
0 2 * * 1 /usr/local/bin/endpointbom

# Every weekday at 2am
0 2 * * 1-5 /usr/local/bin/endpointbom

# Twice daily (2am and 2pm)
0 2,14 * * * /usr/local/bin/endpointbom
```

## Log Rotation

### macOS

Create `/etc/newsyslog.d/endpointbom.conf`:

```
# logfilename          [owner:group]    mode count size when  flags
/var/log/endpointbom/stdout.log         644  7     10240  *     J
/var/log/endpointbom/stderr.log         644  7     10240  *     J
```

### Linux

Create `/etc/logrotate.d/endpointbom`:

```
/var/log/endpointbom/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 644 root root
}
```

### Windows

Use PowerShell in scheduled task to rotate logs:

```powershell
# Keep last 7 days of logs
Get-ChildItem "C:\ProgramData\EndpointBOM\logs\*.log" | 
    Where-Object {$_.LastWriteTime -lt (Get-Date).AddDays(-7)} | 
    Remove-Item
```

## Monitoring & Alerts

### Verify Scans Are Running

**macOS:**
```bash
# Check LaunchDaemon status
sudo launchctl list | grep endpointbom

# Check recent scans
ls -lt /var/log/endpointbom/scans/ | head -5

# Check logs for errors
grep -i error /var/log/endpointbom/stderr.log
```

**Windows:**
```powershell
# Check task status
Get-ScheduledTask -TaskName "EndpointBOM Daily Scan" | Get-ScheduledTaskInfo

# Check recent scans
Get-ChildItem "C:\ProgramData\EndpointBOM\scans" -File | Sort-Object LastWriteTime -Descending | Select-Object -First 5

# Check for errors
Select-String -Path "C:\ProgramData\EndpointBOM\logs\*.log" -Pattern "error" -CaseSensitive:$false
```

**Linux:**
```bash
# Check systemd timer status
sudo systemctl status endpointbom.timer

# Or check cron
sudo journalctl -u cron | grep endpointbom

# Check recent scans
ls -lt /var/log/endpointbom/scans/ | head -5
```

### Alert on Missing Scans

Create a monitoring script that alerts if no scan in last 25 hours:

**Bash (macOS/Linux):**
```bash
#!/bin/bash
SCAN_DIR="/var/log/endpointbom/scans"
LATEST_SCAN=$(find "$SCAN_DIR" -name "*.cdx.json" -type f -mtime -1 | wc -l)

if [ "$LATEST_SCAN" -eq 0 ]; then
    echo "ERROR: No EndpointBOM scan in last 24 hours!"
    # Send alert (email, Slack, PagerDuty, etc.)
    exit 1
fi

echo "OK: Recent scan found"
exit 0
```

**PowerShell (Windows):**
```powershell
$scanDir = "C:\ProgramData\EndpointBOM\scans"
$latestScan = Get-ChildItem $scanDir -Filter "*.cdx.json" | 
    Where-Object {$_.LastWriteTime -gt (Get-Date).AddHours(-25)}

if ($latestScan.Count -eq 0) {
    Write-Error "No EndpointBOM scan in last 24 hours!"
    # Send alert
    exit 1
}

Write-Host "OK: Recent scan found"
exit 0
```

## Troubleshooting

### macOS: LaunchDaemon Not Running

```bash
# Check if loaded
sudo launchctl list | grep endpointbom

# Check logs
tail -50 /var/log/endpointbom/stderr.log

# Test manually
sudo /usr/local/bin/endpointbom --output /var/log/endpointbom/scans --verbose

# Reload
sudo launchctl unload /Library/LaunchDaemons/com.endpointbom.scan.plist
sudo launchctl load /Library/LaunchDaemons/com.endpointbom.scan.plist
```

### Windows: Scheduled Task Failing

```powershell
# Check task history
Get-ScheduledTask -TaskName "EndpointBOM Daily Scan" | Get-ScheduledTaskInfo

# View event log
Get-WinEvent -LogName "Microsoft-Windows-TaskScheduler/Operational" | 
    Where-Object {$_.Message -like "*EndpointBOM*"} | 
    Select-Object -First 10

# Test manually
& "C:\Program Files\EndpointBOM\endpointbom.exe" --output "C:\ProgramData\EndpointBOM\scans" --verbose
```

### Linux: Cron Not Running

```bash
# Check if cron service is running
sudo systemctl status cron    # Debian/Ubuntu
sudo systemctl status crond   # RHEL/CentOS

# Check cron logs
grep endpointbom /var/log/syslog
grep CRON /var/log/syslog | grep endpointbom

# Test manually
sudo /usr/local/bin/endpointbom --output /var/log/endpointbom/scans --verbose
```

## Best Practices

1. **Run at low-usage times** (2am-4am typical)
2. **Monitor for failures** (check daily that scans complete)
3. **Rotate logs** (don't let logs grow unbounded)
4. **Test first** (run manually before automating)
5. **Set timeouts** (max 2 hours for scan to complete)
6. **Alert on failures** (integrate with monitoring system)
7. **Centralize results** (push SBOMs to central location)

## Enterprise Deployment

For enterprise deployments (Jamf, Intune, SCCM), see:
- **[JAMF_DEPLOYMENT.md](JAMF_DEPLOYMENT.md)** - Jamf Pro deployment
- Deploy scheduling configuration via MDM
- Centralize SBOM collection
- Monitor compliance dashboards

---

**Related Documentation:**
- [JAMF_DEPLOYMENT.md](JAMF_DEPLOYMENT.md) - Jamf deployment guide
- [README.md](../README.md) - Main documentation
- [docs/USAGE.md](../docs/USAGE.md) - Usage guide

