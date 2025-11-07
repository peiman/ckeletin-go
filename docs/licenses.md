# License Compliance Guide

> **tl;dr**: ckeletin-go automatically checks dependency licenses to prevent legal issues. Run `task check:license:source` after adding dependencies. Conservative permissive-only policy by default (MIT, Apache-2.0, BSD, ISC allowed; GPL/AGPL denied).

**See also:**
- [ADR-011](adr/011-license-compliance.md) - Full strategy and rationale
- [CLAUDE.md](../CLAUDE.md#license-compliance) - Quick reference for developers
- [.lichen.yaml](../.lichen.yaml) - Binary check configuration

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Why License Compliance Matters](#why-license-compliance-matters)
3. [Default Policy](#default-policy)
4. [Two-Tier Checking System](#two-tier-checking-system)
5. [Common Workflows](#common-workflows)
6. [Handling Violations](#handling-violations)
7. [Customizing Policy](#customizing-policy)
8. [License Types Explained](#license-types-explained)
9. [Generating Reports](#generating-reports)
10. [Troubleshooting](#troubleshooting)

---

## Quick Start

### After Adding a Dependency

```bash
# Add dependency
go get github.com/example/package

# Check licenses (fast, 2-5 seconds)
task check:license:source

# If all clear, commit
task check  # Includes license check
git add . && git commit -m "feat: add example package"
```

### Before a Release

```bash
# Build binary
task build

# Check binary licenses (accurate, 10-15 seconds)
task check:license:binary

# Generate license artifacts
task generate:license
# Creates: reports/licenses.csv, third_party/licenses/, NOTICE
```

---

## Why License Compliance Matters

### The Problem

**Open-source licenses impose legal obligations.** Some licenses (GPL, AGPL) require you to:
- Release your entire application as open-source
- Provide source code to all users
- License your code under the same terms

**This is a problem if you're building:**
- Commercial/proprietary software
- Closed-source internal tools
- SaaS applications (AGPL applies to network use)

### The Solution

**Automated license checking** catches incompatible licenses before they become legal problems:

- ✅ Prevents accidentally using GPL/AGPL dependencies
- ✅ Provides legal peace of mind for commercial use
- ✅ Catches issues during development, not audits
- ✅ Generates compliance reports for customers/auditors
- ✅ Educates developers about license implications

---

## Default Policy

ckeletin-go uses a **conservative permissive-only policy** by default.

### ✅ Allowed Licenses

| License | Type | Use Case |
|---------|------|----------|
| MIT | Permissive | Most common, simple |
| Apache-2.0 | Permissive | Patent grant, NOTICE required |
| BSD-2-Clause | Permissive | Simple BSD |
| BSD-3-Clause | Permissive | BSD with non-endorsement |
| ISC | Permissive | Like MIT, simpler wording |
| 0BSD | Public Domain | No restrictions |
| Unlicense | Public Domain | Explicit public domain |

**What this means:** Use freely, modify, distribute, including in commercial/proprietary software. Only requirement: include copyright notice and license text.

### ❌ Denied Licenses

| License | Type | Why Denied |
|---------|------|------------|
| GPL-2.0, GPL-3.0 | Strong Copyleft | Forces entire app to be GPL |
| AGPL-3.0 | Network Copyleft | Forces source release even for SaaS |
| LGPL-2.0, LGPL-2.1, LGPL-3.0 | Weak Copyleft | Complex linking requirements |
| MPL-2.0 | Weak Copyleft | File-level copyleft, conservatively denied |
| SSPL | Network Copyleft | Server-side restrictions |

**What this means:** These licenses require source code disclosure or impose restrictions incompatible with proprietary software.

### ⚠️ Unknown Licenses

**Treated as violations** until manually reviewed and overridden in `.lichen.yaml`.

---

## Two-Tier Checking System

ckeletin-go uses **two tools** for different purposes:

### Source-Based: go-licenses (Development)

**Purpose:** Fast feedback during development

**How it works:**
- Scans `go.mod` and source code
- Checks ALL dependencies (including test-only)
- ~2-5 seconds

**Use:**
```bash
task check:license:source
```

**When:**
- After `go get`
- During development
- In local `task check`

**Limitation:** May include dependencies not shipped in binary (false positives)

### Binary-Based: lichen (Release)

**Purpose:** Accurate verification for releases

**How it works:**
- Scans compiled binary via `go version -m`
- Checks ONLY runtime dependencies
- ~10-15 seconds (requires build)

**Use:**
```bash
task check:license:binary
```

**When:**
- Before releases
- Before tagging versions
- Final verification

**Benefit:** 100% accurate for what ships to customers

### Orchestrator: Both (CI)

**Purpose:** Comprehensive check with defense in depth

**Use:**
```bash
task check:license  # Runs both
```

**When:**
- In CI/CD (automated)
- Before major commits
- When paranoid about compliance

---

## Common Workflows

### Adding a New Dependency

```bash
# 1. Add dependency
go get github.com/spf13/viper

# 2. Check licenses (fast)
task check:license:source

# 3. If error, see "Handling Violations" section
# 4. If clean, proceed
task check
git add go.mod go.sum
git commit -m "feat: add viper for configuration"
```

### Updating Dependencies

```bash
# 1. Update
go get -u ./...

# 2. Check licenses (may have changed!)
task check:license:source

# 3. Review changes
task generate:license:report
git diff reports/licenses.csv

# 4. If clean, commit
task check
git add go.mod go.sum
git commit -m "chore: update dependencies"
```

### Before a Release

```bash
# 1. Build binary
task build

# 2. Accurate check
task check:license:binary

# 3. Generate compliance artifacts
task generate:license
# Creates:
#   - reports/licenses.csv (audit report)
#   - third_party/licenses/ (all license texts)
#   - NOTICE (attribution file)

# 4. Review
cat NOTICE
head reports/licenses.csv

# 5. Tag release
git tag v1.0.0
git push --tags
```

---

## Handling Violations

### Example Violation

```bash
$ task check:license:source

❌ License compliance check failed

Found disallowed licenses:
  github.com/example/gpl-lib@v1.0.0: GPL-3.0 (forbidden)
```

### Resolution Steps

#### Option 1: Remove the Dependency (Preferred)

```bash
# Remove dependency
go get github.com/example/gpl-lib@none
go mod tidy

# Verify
task check:license:source
```

#### Option 2: Find an Alternative

1. Search [pkg.go.dev](https://pkg.go.dev) for alternatives
2. Filter by license: Look for MIT, Apache-2.0, BSD
3. Check "Similar packages" section
4. Replace import statements
5. Verify: `task check:license:source`

**Example:**
```bash
# Instead of GPL library
go get github.com/example/gpl-lib@none

# Use MIT alternative
go get github.com/alternative/mit-lib
```

#### Option 3: Override Policy (If Justified)

**⚠️ Only if:**
- You understand the license implications
- Your project allows that license type
- You have legal approval (for commercial projects)

**Temporary override:**
```bash
LICENSE_ALLOWED="MIT,Apache-2.0,BSD-3-Clause,MPL-2.0" \
  task check:license:source
```

**Permanent override (edit `.lichen.yaml`):**
```yaml
allow:
  - "MIT"
  - "Apache-2.0"
  - "MPL-2.0"  # Added after legal review
```

#### Option 4: Add Exception (Justified Cases)

**Edit `.lichen.yaml`:**
```yaml
exceptions:
  licenseNotPermitted:
    - path: "github.com/legacy/library"
      licenses: ["LGPL-3.0"]
      # Justification: Core functionality, no alternative,
      #                dynamic linking maintained per LGPL requirements
```

**Document:**
- Why exception is needed
- What alternatives were considered
- How license requirements are satisfied
- Who approved (for commercial projects)

---

## Customizing Policy

### Via Environment Variables (Quick)

**Allow additional licenses:**
```bash
export LICENSE_ALLOWED="MIT,Apache-2.0,BSD-3-Clause,MPL-2.0,LGPL-3.0"
task check:license:source
```

**Change disallowed types:**
```bash
export LICENSE_DISALLOWED="forbidden"  # Only block GPL, not LGPL
task check:license:source
```

**Make permanent:**
Edit `scripts/check-licenses-source.sh`:
```bash
ALLOWED_LICENSES="${LICENSE_ALLOWED:-MIT,Apache-2.0,BSD-3-Clause,MPL-2.0}"
```

### Via .lichen.yaml (Binary Checks)

**Edit `.lichen.yaml`:**
```yaml
threshold: .80  # Confidence for license detection

allow:
  - "MIT"
  - "Apache-2.0"
  - "MPL-2.0"  # Add weak copyleft

override:
  - path: "github.com/example/package"
    licenses: ["MIT"]  # Override mis-detected license

exceptions:
  licenseNotPermitted:
    - path: "github.com/vendor/legacy"
      licenses: ["LGPL-3.0"]
```

**See `.lichen.yaml` for complete examples and comments.**

---

## License Types Explained

### Permissive Licenses (Low Risk)

**Examples:** MIT, Apache-2.0, BSD, ISC

**What they allow:**
- ✅ Commercial use
- ✅ Modification
- ✅ Distribution
- ✅ Sublicensing
- ✅ Private use
- ✅ Closed-source derivatives

**Requirements:**
- Include copyright notice
- Include license text
- NOTICE file (Apache-2.0 only)

**Risk Level:** ⬜ Low - Safe for commercial/proprietary use

### Weak Copyleft (Medium Risk)

**Examples:** LGPL, MPL

**What they require:**
- Modified **library** source must be provided
- Application code can remain proprietary
- Dynamic linking usually OK
- Static linking may trigger copyleft

**Requirements:**
- Source for modified library files
- License preservation
- Complex compliance (consult legal)

**Risk Level:** ⚠️ Medium - Consult legal counsel

### Strong Copyleft (High Risk)

**Examples:** GPL-2.0, GPL-3.0

**What they require:**
- **Entire application** must be GPL
- All source code provided to users
- Cannot combine with proprietary code
- Derivative works must be GPL

**Requirements:**
- Full source disclosure
- License entire codebase as GPL
- Provide source to all users

**Risk Level:** ❌ High - Incompatible with proprietary software

### Network Copyleft (Very High Risk)

**Examples:** AGPL-3.0, SSPL

**What they require:**
- Same as GPL, **plus**
- Source provided even for SaaS/network use
- Triggered by offering software over network
- Includes infrastructure code (SSPL)

**Requirements:**
- Full source disclosure
- Applies to SaaS deployments
- Very broad requirements

**Risk Level:** 🚫 Very High - Avoid for SaaS/commercial

---

## Generating Reports

### CSV Report (Audits)

```bash
task generate:license:report
# Output: reports/licenses.csv

# View
cat reports/licenses.csv | column -t -s ','

# Opens in Excel/Numbers
open reports/licenses.csv
```

**Contains:** Package, URL, License

### License Files (Distribution)

```bash
task generate:license:files
# Output: third_party/licenses/

# View structure
tree third_party/licenses

# Include in releases (.goreleaser.yml):
archives:
  files:
    - LICENSE
    - NOTICE
    - third_party/**
```

### NOTICE File (Attribution)

```bash
task generate:attribution
# Output: NOTICE

# View
cat NOTICE

# Required for:
# - Binary distributions
# - Apache-2.0 dependencies
# - Customer compliance requests
```

### All Artifacts

```bash
task generate:license
# Generates: report, files, NOTICE
```

---

## Troubleshooting

### "go-licenses not installed"

```bash
# Install
go install github.com/google/go-licenses/v2@latest

# Or
task setup
```

### "lichen not installed"

```bash
# Install
go install github.com/uw-labs/lichen@latest

# Or
task setup
```

### "Binary not found"

```bash
# lichen requires a built binary
task build

# Then
task check:license:binary
```

### "Unknown license" / Low Confidence

**Causes:**
- No LICENSE file in repository
- Non-standard license text
- Dual/multiple licenses

**Solutions:**

1. **Check manually:**
   - Visit repository on GitHub
   - Look for LICENSE, COPYING, README
   - Check go.pkg.dev page

2. **Override in `.lichen.yaml`:**
   ```yaml
   override:
     - path: "github.com/example/package"
       version: "v1.2.3"
       licenses: ["MIT"]
       # Verified: GitHub repo shows MIT license
   ```

### False Positives (Test Dependencies)

**Problem:** Source check flags test-only dependency

**Solution:** Use binary check for accurate results
```bash
task check:license:binary
# Only checks runtime dependencies
```

### License Changed Upstream

**Weekly CI check catches this**

**Response:**
1. Review new license
2. If incompatible, pin to old version or find alternative
3. Update `.lichen.yaml` if new license is acceptable

---

## Integration with CI

License checks run automatically in CI (`.github/workflows/ci.yml`):

```yaml
- name: Check Licenses
  run: task check:license

- name: Generate Report
  if: always()
  run: task generate:license:report

- name: Upload Report
  uses: actions/upload-artifact@v4
  with:
    name: license-report
    path: reports/
```

**Behavior:**
- ✅ Runs on every PR
- ❌ **Blocks** merge on violations
- 📊 Reports uploaded as artifacts
- 📅 Weekly schedule catches upstream changes

---

## Additional Resources

- **ADR-011:** [License Compliance Strategy](adr/011-license-compliance.md) - Full rationale
- **CLAUDE.md:** [License Compliance Section](../CLAUDE.md#license-compliance) - Developer quick reference
- **Taskfile.yml:** Search for `check:license` tasks
- **SPDX:** [License List](https://spdx.org/licenses/) - Standard license identifiers
- **Choose a License:** [choosealicense.com](https://choosealicense.com/) - License comparison
- **TLDRLegal:** [tldrlegal.com](https://tldrlegal.com/) - License summaries

---

## Quick Reference Card

| Scenario | Command | Time |
|----------|---------|------|
| After `go get` | `task check:license:source` | ~2-5s |
| Before release | `task check:license:binary` | ~10-15s |
| In CI | `task check:license` | ~15-20s |
| Generate report | `task generate:license:report` | ~5s |
| Generate artifacts | `task generate:license` | ~10s |

**Default Policy:** Permissive only (MIT, Apache-2.0, BSD, ISC)
**Customization:** `.lichen.yaml` or environment variables
**Help:** See [ADR-011](adr/011-license-compliance.md) for full details
