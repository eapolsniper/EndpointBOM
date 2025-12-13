# GitHub Automation

This directory contains GitHub-specific automation configuration for dependency management and releases.

## Dependabot Configuration

### `dependabot.yml`

Configures automated dependency updates for:
- **Go modules** - Daily checks for security and version updates
- **GitHub Actions** - Weekly checks for workflow updates

**Key Features:**
- Groups security updates together
- Ignores major version bumps (requires manual review)
- Automatically labels PRs
- Assigns to @eapolsniper for review

### Cooldown Policy

Updates are subject to a **14-day cooldown period** unless:
- ✅ Critical or high severity security fixes (merged immediately)
- ✅ Update is older than 14 days

This ensures community vetting of new releases before automatic adoption.

## Workflows

### `dependabot-auto-merge.yml`

Automatically merges Dependabot PRs based on:
1. **Severity**: Critical/high security fixes merge immediately
2. **Age**: Non-critical updates wait 14 days (cooldown period)
3. **Tests**: Only merges if all tests pass

**Runs:**
- On every Dependabot PR open/update
- Daily at 3am UTC (checks PRs that have passed cooldown)

**Labels:**
- `cooldown-active` - PR is in 14-day cooldown period
- `dependencies` - All dependency updates
- `security` - Security-related updates

### `dependency-test.yml`

Tests all dependency updates before merging:
- ✅ Runs on Linux, macOS, and Windows
- ✅ Verifies `go.mod` and `go.sum` integrity
- ✅ Runs full test suite with race detection
- ✅ Builds binary and tests basic functionality
- ✅ Checks for known vulnerabilities with `govulncheck`
- ✅ Runs dependency review (licenses, known issues)

## How It Works

### For Critical/High Security Updates:

```
1. Dependabot opens PR
2. dependabot-auto-merge.yml detects severity
3. Adds comment: "🚨 Critical/high severity - merging immediately"
4. dependency-test.yml runs tests
5. If tests pass → auto-merges
6. If tests fail → requires manual review
```

### For Medium/Low/Non-Security Updates:

```
1. Dependabot opens PR
2. dependabot-auto-merge.yml checks PR age
3. If < 14 days:
   - Adds "cooldown-active" label
   - Comments with days remaining
   - Waits for cooldown
4. Daily cron job checks all cooldown PRs
5. When 14 days pass:
   - Removes cooldown label
   - dependency-test.yml runs tests
   - If tests pass → auto-merges
```

## Manual Override

You can manually merge any PR at any time:
1. Review the PR and tests
2. Click "Merge pull request"
3. This bypasses the cooldown period

## Configuration

### Adjust Cooldown Period

Edit `dependabot-auto-merge.yml` and change:
```yaml
elif [[ $DAYS_OLD -ge 14 ]]; then  # Change 14 to desired days
```

### Disable Auto-Merge

To disable auto-merge entirely:
1. Delete `.github/workflows/dependabot-auto-merge.yml`
2. Dependabot will still create PRs for manual review

### Change Schedule

Edit `dependabot.yml`:
```yaml
schedule:
  interval: "daily"  # Change to "weekly" or "monthly"
  time: "02:00"     # Change time (UTC)
```

## Monitoring

### Check Dependabot Status

```bash
# List all open Dependabot PRs
gh pr list --author "dependabot[bot]"

# Check PRs in cooldown
gh pr list --label "cooldown-active"

# View Dependabot logs
gh api /repos/eapolsniper/endpointbom/dependabot/alerts
```

### View Workflow Runs

```bash
# List recent workflow runs
gh run list --workflow=dependabot-auto-merge.yml

# View specific run
gh run view <run-id>
```

## Troubleshooting

### PR Not Auto-Merging

Check:
1. Are all status checks passing?
2. Is branch protection enabled?
3. Is auto-merge enabled in repo settings?
4. Check workflow logs: `gh run view <run-id>`

### Cooldown Not Working

Check:
1. Workflow has correct permissions (`contents: write`, `pull-requests: write`)
2. `GITHUB_TOKEN` has sufficient permissions
3. Check cron job ran: `gh run list --workflow=dependabot-auto-merge.yml`

### Tests Failing

Review `dependency-test.yml` logs:
```bash
gh run list --workflow=dependency-test.yml
gh run view <run-id> --log-failed
```

## Security

- ✅ Uses `GITHUB_TOKEN` (no personal access tokens needed)
- ✅ Minimal permissions (contents: write, pull-requests: write)
- ✅ All actions pinned to specific versions
- ✅ Dependency review checks licenses and vulnerabilities

---

**Last Updated:** 2025-12-13  
**Maintainer:** @eapolsniper

