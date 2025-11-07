# ADR-011: License Compliance Strategy

## Status
Accepted

## Context

### The Problem: Legal Risk and Compliance

When building CLI applications, we depend on open-source libraries. Each dependency comes with a license that imposes legal obligations:

- **Permissive licenses** (MIT, Apache-2.0, BSD) allow usage with minimal restrictions
- **Copyleft licenses** (GPL, AGPL) require releasing our code under the same license
- **Unknown licenses** create legal uncertainty and potential liability

**Without automated license checking:**
- Developers unknowingly add dependencies with incompatible licenses
- GPL/AGPL dependencies force entire codebase to become open-source
- Legal issues discovered late (during audit or customer review)
- Manual license reviews are time-consuming and error-prone
- Upstream license changes go undetected

**Real-world consequences:**
- Companies forced to open-source proprietary code
- Last-minute dependency replacements before releases
- Failed customer security/compliance audits
- Legal liability for license violations
- Inability to sell or distribute software

### Why ckeletin-go Needs This

**ckeletin-go is a scaffold** - users build their own projects from it. We must:
1. **Lead by example** - Demonstrate professional license compliance
2. **Protect scaffold users** - Prevent them from inheriting license issues
3. **Enable commercial use** - Scaffold must support proprietary applications
4. **Provide automation** - Make compliance easy, not burdensome
5. **Educate users** - Teach license implications through tooling

### Source vs Binary: The Accuracy Problem

**Challenge**: Not all dependencies in `go.mod` ship in the final binary.

**Example:**
```go
// Source code includes test-only dependency
// +build test
import "github.com/some/test-library"  // LGPL license

// Binary doesn't include this (not compiled)
```

**Two perspectives needed:**
1. **Source-based checking** - Fast feedback during development (go-licenses)
2. **Binary-based checking** - Accurate verification for releases (lichen)

**Without both:**
- Source-only: False positives (flags test-only deps)
- Binary-only: Slow (requires build step for every check)

### Alternatives Considered

**1. No License Checking**
- **Pros**: Zero setup, no tool dependencies
- **Cons**: Legal risk, unprofessional, no user protection
- **Why not**: Unacceptable for production scaffold

**2. Manual License Review**
- **Pros**: Human judgment, handles edge cases
- **Cons**: Time-consuming, error-prone, doesn't scale, not automated
- **Why not**: Can't guarantee every user does this

**3. Single Tool: go-licenses Only**
- **Pros**: Simple, fast, one tool to install
- **Cons**: Source-based only, false positives on test deps
- **Why not**: Inaccurate for final binaries

**4. Single Tool: lichen Only**
- **Pros**: Accurate (binary-based), YAML config
- **Cons**: Requires build first, slower feedback loop
- **Why not**: Too slow for rapid development

**5. Commercial Tools (FOSSA, Snyk, etc.)**
- **Pros**: Comprehensive, great UX, vulnerability scanning
- **Cons**: Costs money, external service dependency, overkill for scaffold
- **Why not**: Creates barrier to entry for scaffold users

**6. Both go-licenses + lichen**
- **Pros**: Fast dev feedback + accurate release verification, defense in depth
- **Cons**: Two tools to install and maintain
- **Why THIS**: Best of both worlds, professional approach

## Decision

We adopt a **dual-tool license compliance strategy**:

1. **go-licenses** for fast development feedback (source-based)
2. **lichen** for accurate release verification (binary-based)
3. **Conservative permissive-only policy** (MIT, Apache-2.0, BSD, ISC)
4. **CI-blocking enforcement** (warn locally, block in CI)
5. **Task orchestrator pattern** following ADR-000

### Architecture

```
Development Workflow:
┌─────────────────────────────────────────────────────────────┐
│ 1. Add Dependency                                           │
│    $ go get github.com/example/package                      │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Fast Source Check (Development)                          │
│    $ task check:license:source                              │
│    → Uses go-licenses (scans go.mod/source)                 │
│    → ~2-5 seconds                                            │
│    → May include test-only dependencies                     │
│    → Catches issues early                                   │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Before Release: Binary Check (Accurate)                  │
│    $ task check:license:binary                              │
│    → Uses lichen (scans compiled binary)                    │
│    → ~10-15 seconds (requires build)                        │
│    → Only runtime dependencies                              │
│    → 100% accurate for shipping                             │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. CI: Both Checks (Comprehensive)                          │
│    $ task check:license (orchestrator)                      │
│    → Runs both source and binary checks                     │
│    → Blocks merge on violations                             │
│    → Defense in depth                                       │
└─────────────────────────────────────────────────────────────┘
```

### Tool Selection

#### go-licenses (google/go-licenses)

**Capabilities:**
- Scans `go.mod` and source code
- CSV reports with license info per dependency
- Policy enforcement (`--allowed_licenses`, `--disallowed_types`)
- License file extraction (`save` command)
- Built-in license classification

**Pros:**
- ✅ Actively maintained by Google
- ✅ Fast (~2-5 seconds)
- ✅ No configuration file needed
- ✅ Simple `go install` installation
- ✅ Wide adoption in Go ecosystem
- ✅ Works with Go modules natively

**Cons:**
- ❌ Source-based (may include test/build deps)
- ❌ No YAML config (CLI flags only)
- ❌ May have false positives

**Usage:**
```bash
go-licenses check --allowed_licenses=MIT,Apache-2.0 ./...
go-licenses report ./... > licenses.csv
go-licenses save ./... --save_path=third_party/licenses
```

#### lichen (uw-labs/lichen)

**Capabilities:**
- Analyzes compiled binaries via `go version -m`
- YAML configuration support
- Override mechanism for manual specification
- Exception handling for justified violations
- JSON output support

**Pros:**
- ✅ Binary-based (only shipping dependencies)
- ✅ 100% accurate for releases
- ✅ YAML configuration (user-friendly)
- ✅ Flexible exception mechanism
- ✅ Confidence threshold tuning

**Cons:**
- ❌ Requires pre-built binary (slower)
- ❌ Need build step for each platform
- ❌ More complex setup than go-licenses

**Usage:**
```bash
go build -o myapp
lichen --config=.lichen.yaml myapp
```

### License Policy

**Default: Conservative Permissive-Only**

**✅ Allowed:**
- MIT
- Apache-2.0
- BSD-2-Clause, BSD-3-Clause
- ISC
- 0BSD (Zero-Clause BSD)
- Unlicense

**❌ Denied (Disallowed Types: `forbidden`, `restricted`):**
- GPL-2.0, GPL-3.0 (strong copyleft)
- AGPL-3.0 (network copyleft)
- LGPL-2.0, LGPL-2.1, LGPL-3.0 (weak copyleft)
- SSPL (Server Side Public License)
- EUPL (European Union Public License)
- MPL-2.0 (Mozilla Public License - weak copyleft)

**Rationale:**
- Permissive licenses impose minimal restrictions
- Compatible with commercial/proprietary software
- No source code disclosure requirements
- Scaffold users can build closed-source products
- Conservative = safest default for widest use cases

**Customization:**
Users can override via environment variables or config files if their specific use case allows copyleft licenses.

### Task Integration (ADR-000 Compliance)

Following the orchestrator pattern:

```yaml
# CHECK - Orchestrator and variants
check:license:
  desc: Check all license compliance (orchestrator)
  cmds:
    - task: check:license:source
    - task: check:license:binary

check:license:source:
  desc: Check dependency licenses from source (fast, development)
  cmds:
    - ./scripts/check-licenses-source.sh

check:license:binary:
  desc: Check licenses in compiled binary (accurate, release)
  cmds:
    - ./scripts/check-licenses-binary.sh
  deps: [build]

# GENERATE - Orchestrator and variants
generate:license:
  desc: Generate all license artifacts (orchestrator)
  cmds:
    - task: generate:license:report
    - task: generate:license:files
    - task: generate:attribution

generate:license:report:
  desc: Generate CSV license report
  cmds:
    - ./scripts/generate-license-report.sh

generate:license:files:
  desc: Save license files to third_party/
  cmds:
    - mkdir -p third_party/licenses
    - go-licenses save ./... --save_path=third_party/licenses

generate:attribution:
  desc: Generate NOTICE file with attribution
  cmds:
    - ./scripts/generate-attribution.sh
  deps: [generate:license:report]
```

**Why this structure:**
- `check:license` = simple interface, runs everything
- `check:license:source` = fast feedback during development
- `check:license:binary` = accurate verification before release
- `generate:license` = all artifacts with one command
- Consistent with `check:deps`, `generate:docs` patterns

### Enforcement Strategy

**Local Development:**
- ✅ `check:license:source` runs in `task check` (fast)
- ⚠️  Warnings shown, but non-blocking
- 🎯 Goal: Early feedback without slowing development

**CI/CD:**
- ✅ `check:license` runs (both source + binary)
- ❌ **BLOCKS** on violations
- 📊 Reports uploaded as artifacts
- 🔒 Can't merge non-compliant code

**Weekly Schedule:**
- 🔄 Automated CI run every Sunday
- 📅 Catches upstream license changes
- 📧 Notifications on new violations

## Consequences

### Positive

**1. Legal Protection**
- Automated detection of incompatible licenses
- Can't accidentally ship GPL/AGPL code
- Defense in depth (two tools, two methods)
- Early detection (dev) + accurate verification (release)

**2. Professional Standard**
- Enterprise-grade compliance baked into scaffold
- Users inherit good practices
- Audit-ready from day one
- SBOM generation support

**3. Dual-Tool Benefits**
- **Speed**: go-licenses gives fast feedback (~2-5s)
- **Accuracy**: lichen verifies actual shipping dependencies
- **Confidence**: Two tools cross-validate
- **Flexibility**: Choose tool based on context

**4. Developer Experience**
- Clear error messages with remediation steps
- Fast local checks don't slow development
- Comprehensive CI checks prevent issues
- Easy customization via env vars or YAML

**5. Educational Value**
- Users learn about license implications
- Scaffold demonstrates proper compliance
- Documentation explains "why" not just "how"
- Encourages conscious dependency choices

**6. Task Pattern Consistency**
- Follows ADR-000 orchestrator pattern perfectly
- Clear `action:target:variant` naming
- Integrates seamlessly with existing `task check`
- Example of pattern scaling to complexity

### Negative

**1. Tool Dependencies**
- Must install go-licenses AND lichen
- Two tools to maintain and update
- Potential version compatibility issues
- Binary size increase for CI containers

**2. Build Requirement**
- Binary check requires `task build` first
- Slower than source-only checking
- Cross-platform binaries need multiple checks
- CI time increase (~10-15s per platform)

**3. False Positives Possible**
- Source check may flag test-only dependencies
- Lichen confidence threshold may misidentify licenses
- Custom/dual licenses need manual override
- Requires occasional manual intervention

**4. Conservative Policy May Be Restrictive**
- Default denies LGPL/MPL (some users may need)
- May force finding alternatives to useful libraries
- Customization required for non-commercial projects
- Some false "violations" for permissive projects

**5. Learning Curve**
- Users must understand license types
- Need to know when to use source vs binary check
- Configuration files (.lichen.yaml) require learning
- May be overkill for hobby projects

### Mitigations

**1. Tooling Made Easy**
- `task setup` installs both tools automatically
- `task doctor` verifies installation
- Clear installation instructions in error messages
- Caching in CI reduces repeated downloads

**2. Performance Optimization**
- Source check used in `task check` (fast path)
- Binary check optional for local dev
- CI caches build artifacts
- Parallel CI jobs reduce perceived slowness

**3. Documentation & Education**
- Comprehensive `docs/licenses.md` user guide
- ADR explains rationale and alternatives
- CLAUDE.md includes quick reference
- Examples for common scenarios

**4. Customization Support**
- Environment variables for quick overrides
- `.lichen.yaml` for complex policies
- Override mechanisms for edge cases
- Clear upgrade path if users outgrow defaults

**5. Graduated Enforcement**
- Start with warnings in local dev
- Block only in CI (safety net, not obstacle)
- Weekly checks catch upstream changes
- Users can temporarily disable if needed

## Implementation

### Scripts

**scripts/check-licenses-source.sh**
- Uses go-licenses with CLI flags
- Environment variable overrides (LICENSE_ALLOWED, LICENSE_DISALLOWED)
- Default conservative policy
- Clear error messages with next steps

**scripts/check-licenses-binary.sh**
- Uses lichen with .lichen.yaml config
- Requires binary build first
- JSON output parsing for CI integration
- Platform-specific binary checking

**scripts/generate-license-report.sh**
- Generates CSV report via go-licenses
- Human-readable summary to console
- Error log for debugging
- Artifact for compliance audits

**scripts/generate-attribution.sh**
- Parses CSV report
- Generates NOTICE file
- Includes copyright, license type, URLs
- Required for Apache-2.0 dependencies

### Configuration

**.lichen.yaml**
- Threshold: 0.80 (80% confidence)
- Allowed licenses list
- Override section for edge cases
- Exception handling
- Well-commented for user customization

### Documentation

**docs/licenses.md**
- What is license compliance
- Default policy and rationale
- How to check licenses
- Handling violations
- Customization guide
- License types explained
- Attribution requirements

## Related ADRs

- [ADR-000](000-task-based-single-source-of-truth.md) - Task orchestrator pattern for license checking
- [ADR-004](004-security-validation-in-config.md) - Security validation complements compliance checking
- [ADR-008](008-release-automation-with-goreleaser.md) - Binary checks integrate with release process

## References

- [google/go-licenses](https://github.com/google/go-licenses) - Source-based license checking
- [lichen](https://github.com/uw-labs/lichen) - Binary-based license analysis
- [SPDX License List](https://spdx.org/licenses/) - Standard license identifiers
- [Choose a License](https://choosealicense.com/) - License comparison guide
- [TLDRLegal](https://www.tldrlegal.com/) - License summaries
- [Open Source Guide: Legal](https://opensource.guide/legal/) - Legal considerations
