# ckeletin-go: Full Strategic & Technical Analysis

**Date:** April 2026
**Methodology:** Multi-dimensional codebase analysis covering architecture, testing, CI/CD, framework mechanism, security, community, and feature gaps.

---

## The State of Things

| Metric | Value |
|--------|-------|
| GitHub stars | 12 |
| Forks | 1 |
| Open issues | 0 |
| Total commits | 228+ since July 2025 |
| Test count | 1,282 tests |
| Coverage | 88.2% (target: 85%) |
| Dependencies | 12 direct |
| CI checks | 23 automated |
| ADRs | 14 |
| Scripts | 37 shell + 1 Go |

**Bottom line:** Technically excellent project with virtually no community. The engineering is A-grade; the adoption is pre-seed.

---

## 1. Architecture & Code Quality

### Strengths

**Ultra-thin command pattern works.** Not just documented — real. `ping.go` is 31 lines. `check.go` is 51. Business logic lives in `internal/`. Enforced by semgrep rules and validation scripts. One of the best Go CLI architectures in open source.

**Configuration system is production-grade.** Centralized registry, auto-generated type-safe constants, no scattered `viper.SetDefault()`. The `config.Key*` constants eliminate an entire class of typo bugs. Generation pipeline (`registry.go` → `generate-config-constants.go` → `keys_generated.go`) is clean.

**Logging is thoughtful.** Dual-stream (console + file), structured (zerolog), with sampling for log storms, path sanitization to prevent credential leaks, proper color/TTY detection. More mature than most production systems.

**`pkg/checkmate` is a genuinely useful library.** 94.1% coverage, thread-safe, TTY-aware, mockable. Could stand alone as its own project.

### Weaknesses

**`internal/check/` is the weakest link.** 48.5% test coverage for the most complex business logic. This is the command that runs 23 quality checks — the heart of the system — and it's the least tested part.

- Shell-based checks via `shellCheck()` are fragile (string parsing of shell output)
- TUI and business logic mixed in `executor.go` (508 lines)
- Timing persistence uses JSON files without atomic writes (data race risk)
- No pre-validation of embedded shell scripts

**Config validation happens too late.** Security checks run after Viper reads the file. Color validation happens in `Execute()`, not at config time.

**Progress system is over-engineered.** 3,357 lines for progress reporting with an event model where half the fields are unused per call. However: it works, it's 98% tested, and refactoring risks breaking things for no user-facing benefit. Leave it.

---

## 2. Testing

### Strengths

- 1,282 tests, 88.2% coverage (exceeds 85% target)
- 139+ table-driven tests with consistent patterns
- DI over mocks (ADR-003) — no gomock/testify-mock, simple manual test implementations
- Golden file testing with `goldie/v2` for CLI output snapshots
- Integration tests that build and run the actual binary
- 32+ benchmarks for hot paths
- No flaky tests detected — no `time.Sleep()`, proper `t.TempDir()` usage

### Gaps

- **Race detection not routine in CI.** Only via `task test:race`, not in standard pipeline. Race conditions can be merged.
- **No signal handling tests.** No SIGINT/SIGTERM testing. For a CLI, graceful shutdown matters.
- **No property-based or fuzz testing in CI.** Fuzz infrastructure exists (`task test:fuzz`) but isn't running.
- **Testify migration incomplete.** Some older tests still use `t.Errorf()` instead of testify assertions.

---

## 3. CI/CD & Tooling

### Strengths

A+ build system maturity:

- 23 automated quality checks in `task check`
- Smart tiered checking: `check:fast` for PRs, `check` for dev, `check:ci` strict
- Multi-platform testing (Linux, macOS, Windows)
- Intelligent caching (pinned tools cached, security tools always fresh)
- Minimal permissions model in GitHub Actions
- Action pinning to specific commits
- GoReleaser with multi-platform builds, SBOM, checksums
- Lefthook pre-commit with parallel execution

### Gaps

- **Semgrep doesn't run on PRs.** CI config explicitly skips it for pull requests. Custom SAST rules (including ADR enforcement) aren't checked until code hits main.
- **No binary signing or SLSA provenance.** Releases aren't signed. Users can't cryptographically verify binaries.
- **Go version mismatch.** `.go-version` says 1.26.1, `go.mod` says 1.24.2. Needs alignment.
- **No commit message validation in pre-commit.** Conventional commits documented but not enforced via Lefthook.
- **SBOM scanning at release time missing.** SBOMs are generated but not scanned with grype.
- **actionlint download without integrity check.** CI pipes bash from curl with no hash verification.

---

## 4. The Framework Update Mechanism

### What's Elegant

- **Git-based, no extra tooling.** `git checkout ckeletin-upstream/main -- .ckeletin/` is brilliantly simple.
- **True separation.** Framework code never imports from project code.
- **Namespace isolation.** `ckeletin:*` task prefix prevents collisions.
- **Pkg-preserving init.** Derived projects import framework packages as external dependencies.
- **Two-tier ADRs.** Framework decisions (000-099) vs project decisions (100+).

### What's Fragile

**No framework versioning.** Users can't answer "what framework version do I have?" No version constant, no manifest, no tag correlation. Only way to check is `git log -- .ckeletin/ | head`.

**No breaking change detection.** If a framework update changes an API, user code silently breaks. No pre-flight check, no compatibility validation, no migration docs.

**Sed-based module replacement is brittle.** The update mechanism does blind string replacement of module paths in `.go` files. Can corrupt comments, string constants, or partial matches. Go's standard library has `go/ast` support for proper import rewriting — this should be used instead.

**No dry-run or rollback.** Updates commit immediately with `--no-verify`. No preview, no undo.

**No post-update validation.** After updating `.ckeletin/`, the user's code isn't checked for compatibility. A simple `go build ./...` after update would catch most issues.

**Task alias breakage possible.** If framework removes a task that root Taskfile aliases, user discovers this at runtime, not during update.

### Grade: B+ for elegance, C+ for production-readiness

---

## 5. Community & Adoption

12 stars, 1 fork, 3 issues (all closed), 1 external user comment. 228 commits of excellent engineering with no external communication beyond the README.

### Issues

- **GitHub topics are stale.** Current: `boilerplate`, `cli`, `go`, `golang`, `skaffold`, `skeleton`. Missing: `ai-agent`, `claude-code`, `scaffold`, `framework`, `cobra`, `viper`. `skaffold` is confusing (Google's Kubernetes tool).
- **GitHub description doesn't mention AI.** Missing the differentiator.
- **No demo or showcase.** No GIF/asciinema of "tell Claude to add a command."
- **No blog post or launch announcement.** Zero external communication.

---

## 6. Security

### Strengths

- govulncheck for dependency vulnerabilities
- gitleaks for secrets scanning
- semgrep with custom SAST rules (CWE-mapped)
- Trivy for filesystem scanning
- GPL/AGPL license blocking
- Path sanitization in logging
- File permission validation
- Config file size limits (DoS prevention)
- gosec linter enabled

### Gaps

- SBOM scanning at release time (generated but not scanned with grype)
- Binary signing (cosign)
- SLSA provenance for supply chain verification
- Semgrep not running on PRs

---

## 7. Feature Gaps

### `task generate:command` is incomplete

Currently generates a basic stub. Should generate the full pattern: `cmd/<name>.go` + `internal/<name>/<name>.go` + `internal/<name>/<name>_test.go` + `internal/config/commands/<name>_config.go` + test scaffolding. Most common developer workflow is still 8 manual steps.

### No `task ckeletin:doctor` for framework health

Should report: framework version, update availability, local modifications in `.ckeletin/`, import consistency.

### Windows support is ambiguous

AGENTS.md says "not officially supported" but CI tests on Windows and GoReleaser builds Windows binaries. Needs clarification.

### No plugin/extension system

Framework is all-or-nothing. No composable plugins for database, gRPC, etc. Premature for current adoption level but worth noting.

---

## 8. The Honest Scorecard

| Dimension | Grade | Notes |
|-----------|-------|-------|
| Architecture | A | Clean, enforced, AI-agent ready |
| Code Quality | B+ | Strong overall, `check/` is weak (48.5%) |
| Testing | A- | 88.2%, 1,282 tests, no flaky tests |
| CI/CD | A | 23 checks, multi-platform, smart caching |
| Framework Update | B- | Elegant design, fragile execution |
| Documentation | A- | Excellent after value rebalance |
| Security | A- | Multi-layer scanning, missing signing |
| Community/Adoption | D | 12 stars, no external communication |
| Developer Experience | B+ | Good for humans, excellent for AI agents |

**Overall: B+ — technically excellent project that almost nobody knows about.**

---

## 9. Prioritized Recommendations

### Immediate (This Week)

| # | What | Why | Effort |
|---|------|-----|--------|
| 1 | Fix Go version mismatch (.go-version vs go.mod) | Bug | 5 min |
| 2 | Update GitHub topics | Discoverability | 5 min |
| 3 | Update GitHub repo description | AI differentiator | 5 min |
| 4 | Enable semgrep on PRs | Security gap | 10 min |

### Short-Term (This Month)

| # | What | Why | Effort |
|---|------|-----|--------|
| 5 | Add `.ckeletin/VERSION` + `task ckeletin:version` | Framework versioning — #1 missing piece | 1-2 hours |
| 6 | Add `--dry-run` to `task ckeletin:update` | Users need safe preview | 1-2 hours |
| 7 | Improve `task generate:command` to scaffold full pattern | Most common workflow is manual | 4-6 hours |
| 8 | Add race detection to CI test run | Concurrency bugs can merge | 30 min |
| 9 | Improve `internal/check/` test coverage (48.5% → 75%) | Core feature credibility gap | 4-6 hours |
| 10 | Add post-update `go build ./...` check | Catch breaking framework changes | 30 min |
| 11 | Add `.ckeletin/CHANGELOG.md` for framework changes | Breaking change documentation | 1 hour |

### Medium-Term (This Quarter)

| # | What | Why | Effort |
|---|------|-----|--------|
| 12 | Create asciinema demo of AI-agent workflow | Highest-ROI marketing asset | 2-3 hours |
| 13 | Write "How to make any codebase AI-agent ready" blog post | Thought leadership positioning | 1 day |
| 14 | Add binary signing (cosign) | Supply chain security | 4-6 hours |
| 15 | Add commit message validation to Lefthook | Enforce conventional commits | 1-2 hours |
| 16 | Clarify Windows support story | Currently ambiguous | 1-2 hours |
| 17 | Add `task ckeletin:doctor` (framework health check) | Makes update story tangible | 4-6 hours |
| 18 | Add config-time validation for user-facing values | Earlier, better error messages | 2-3 hours |
| 19 | Extract TUI from `internal/check/executor.go` | Testability of core logic | 4-6 hours |
| 20 | Add SBOM scanning (grype) to release workflow | Vulnerability scanning of shipped binaries | 2-3 hours |

### Long-Term (Next Quarter+)

| # | What | Why | Effort |
|---|------|-----|--------|
| 21 | Replace sed with `go/ast` for import rewriting | Eliminates string replacement bugs in framework update and scaffold init | 1-2 days |
| 22 | Pre-update compatibility checking | Framework updates become safe at scale | 1 week |
| 23 | Weekly fuzz testing in CI | Catches edge cases humans miss | 2-3 hours |
| 24 | Plugin/extension architecture | Composability (only if adoption warrants) | 2+ weeks |
| 25 | Add SLSA provenance generation | Full supply chain verification | 1-2 days |
| 26 | Migrate remaining tests to testify | Consistent assertion patterns | 2-3 hours |
| 27 | Add signal handling tests (SIGINT/SIGTERM) | Graceful shutdown coverage | 2-3 hours |
| 28 | Add atomic writes to timing persistence | Eliminate data race risk in check/ | 1-2 hours |
