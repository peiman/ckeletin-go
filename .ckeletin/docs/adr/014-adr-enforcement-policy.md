# ADR-014: ADR Enforcement Policy

## Status
Accepted

## Context

The project has 14 ADRs covering architecture, code quality, and workflow decisions. Enforcement varies widely: some ADRs have multiple automated validators (`task validate:*` scripts, semgrep rules, linter checks), while others rely purely on developer discipline (honor system).

Inconsistent enforcement means architectural decisions silently erode over time. Developers may not even know a rule exists if nothing prevents them from breaking it.

## Decision

### 1. Every ADR MUST have automated enforcement where technically feasible

If a decision can be checked by a tool, it must be. The enforcement ladder below lists mechanisms in order of preference (strongest first):

| Level | Mechanism | Example |
|-------|-----------|---------|
| 1 | **Compile-time** | Build tags, type system constraints |
| 2 | **Linter** | golangci-lint rules, go-arch-lint |
| 3 | **Semgrep** | Pattern-based SAST rules |
| 4 | **Validator script** | `task validate:*` scripts |
| 5 | **CI-only** | Checks that run only in CI pipeline |
| 6 | **Honor system** | Documented but not automated (last resort) |

Prefer higher levels — they catch violations earlier and require less effort to maintain.

### 2. ADRs where automation isn't possible MUST document the gap

If an ADR cannot be fully automated, it must explicitly state:
- What IS automated (and by which tool)
- What remains honor system
- Why automation isn't feasible

### 3. Every ADR MUST include a standardized `## Enforcement` section

This section documents how the ADR is enforced, which tools run, and what gaps remain.

### 4. A living enforcement audit table tracks status

The table below is the single source of truth for enforcement coverage.

## Enforcement Audit

| ADR | Title | Enforcement | Status | Gap |
|-----|-------|------------|--------|-----|
| 000 | Task naming | `validate:task-naming` | Full | — |
| 001 | Ultra-thin commands | `validate:commands` + semgrep `no-os-exit` | Full | — |
| 002 | Config registry | `validate:defaults` + `validate:config-consumption` | Full | — |
| 003 | DI over mocking | Layering (go-arch-lint) + type system + coverage + semgrep `no-mock-frameworks` (advisory) | Full | — |
| 004 | Security validation | `validate:security` + gosec + semgrep | Full | — |
| 005 | Config constants | `validate:constants` | Full | — |
| 006 | Structured logging | zerologlint + semgrep `no-fmt-print` + semgrep `log-error-and-return` | Full | — |
| 007 | Bubble Tea UI | *None* | Honor system | Technology choice — low value to automate |
| 008 | GoReleaser releases | `test:release` validates config | Partial | Config validation only, not process |
| 009 | Layered architecture | `validate:layering` (go-arch-lint) | Full | — |
| 010 | Package organization | `validate:package-organization` | Full | — |
| 011 | License compliance | `check:license` (dual-tool) | Full | — |
| 012 | Dev build tags | `validate:dev-build-tags` | Full | — |
| 013 | Structured output | `validate:output` | Full | — |
| 014 | Enforcement policy | This ADR + audit table | Meta | — |

## Consequences

### Positive
- Architectural decisions are enforced consistently, not just documented
- New contributors discover constraints through tooling, not by reading every ADR
- The audit table makes enforcement gaps visible and trackable
- Enforcement ladder provides clear guidance on where to invest automation effort

### Negative
- More validator scripts and semgrep rules to maintain
- Some ADRs (007, 008) will remain partially honor-system
- Overhead of keeping the audit table current

### Mitigations
- Validators are simple shell scripts, easy to maintain
- Semgrep rules are declarative YAML
- The audit table is updated when ADRs change — it's part of the ADR review process

## References
- All ADR files in `.ckeletin/docs/adr/`
- `.semgrep.yml` for semgrep rules
- `.ckeletin/Taskfile.yml` for validator tasks
- `.ckeletin/scripts/validate-*.sh` for validator implementations
