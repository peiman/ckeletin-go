# Design Spec: Documentation Identity Overhaul

## Problem

ckeletin-go has evolved into both a **scaffold** (fork and customize via `task init`) AND a **framework** (updatable infrastructure in `.ckeletin/`). Documentation uses "scaffold", "skeleton", "template generator", and "framework" interchangeably across files, creating confusion about what the project actually is.

Key contradictions:
- README says "scaffold" in intro but has a "Framework Architecture" section
- ADR-010 says "not a library" but `.ckeletin/` IS a reusable framework layer
- CONTRIBUTING.md doesn't mention `.ckeletin/` or framework updates at all
- AGENTS.md/CLAUDE.md say "skeleton/template generator"
- cmd/root.go help text says "scaffold project"

## Approved Identity

> **A production-ready Go CLI scaffold powered by an updatable framework layer.**

- **Scaffold** = the entry point. You clone, `task init`, customize. This is how people find and start using the project.
- **Framework** = the ongoing value. `.ckeletin/` provides enforced patterns, reusable infrastructure, and independent updates via `task ckeletin:update`.

The project serves three legitimate use cases:
1. **Runnable CLI** — `ckeletin-go ping` works out of the box (demo/reference)
2. **Project scaffold** — `task init name=myapp` bootstraps a new project
3. **Updatable framework** — `task ckeletin:update` keeps infrastructure current

## Files to Update

### 1. README.md — Full Rewrite (~600 lines target, down from 983)

**Current problems:**
- 983 lines, bloated with detailed code examples that belong in CONTRIBUTING.md
- Identity confusion: "scaffold" intro vs "Framework Architecture" section
- Quickstart doesn't mention `task ckeletin:update`
- "Using the Scaffold" section is disconnected from "Framework Architecture"
- Redundant sections: "Customizing the Module Path" repeats "Using the Scaffold"
- "Options Pattern" section is 50+ lines of code better in CONTRIBUTING.md
- "Tooling Best Practices" is generic advice about Cobra/Viper/Zerolog

**Proposed structure:**

```
# ckeletin-go

[Banner + badges]

## TL;DR (keep — it's effective, update identity language)

## What You Get
NEW section: clearly explain the dual nature
- Scaffold: fork, init, customize
- Framework: updatable .ckeletin/ with enforced patterns
- Show the two-tier directory structure upfront

## Quick Start (consolidate from current Quickstart + Using the Scaffold)
1. Clone + task setup
2. task init name=myapp module=...
3. task build && ./myapp ping
4. (NEW) Mention: "Your project now has an updatable framework layer..."

## Features (trim — current is good but verbose)
- Keep feature list, trim license compliance detail (link to docs/licenses.md)

## Architecture (merge current Architecture + Framework Architecture)
- One section, not two
- Lead with the two-tier model (framework vs project)
- Show directory tree with clear FRAMEWORK / PROJECT annotations
- Layered architecture as a subsection
- ADR references

## Getting Started (trim)
- Prerequisites
- Installation options (keep binary/homebrew/source)
- Remove "Customizing the Module Path" (already covered by task init)
- Remove "Single Source of Truth for Names" (implementation detail for CONTRIBUTING)

## Configuration (keep, minor trim)
- Keep config management, precedence, env vars, flags
- Remove "Adding New Configuration Options" (move to CONTRIBUTING)
- Remove "Automatic Documentation Generation" (move to CONTRIBUTING)

## Commands (keep as-is, minor trim)
- ping, doctor, config validate, check, dev

## Development Workflow (trim)
- Essential task commands
- Tool reproducibility (trim pinned tools list to just the concept + task doctor)
- Pre-commit hooks (1 line)
- CI (1 line)
- Creating Releases (keep but trim)

## Customization (significant trim)
- Keep: Changing program name, adding commands (brief), framework vs project
- MOVE to CONTRIBUTING: Command Implementation Pattern, Options Pattern, detailed code examples
- Remove: Tooling Best Practices (generic advice)

## Framework Updates (NEW)
- How task ckeletin:update works
- What's framework vs project
- When to update

## AI Integration (keep, update terminology)
- Update CLAUDE.md → AGENTS.md + CLAUDE.md references

## Contributing + License + Notes (keep)
```

**Key terminology changes:**
- Line 5: "scaffold" → "A production-ready Go CLI scaffold powered by an updatable framework layer"
- Line 114: Remove "skeleton" reference, use "scaffold with framework" language
- Line 165: "you're running the scaffold" → "you're running your new CLI"
- Line 980: "offered by this scaffold" → "offered by ckeletin-go"
- All instances of "scaffold" used alone → add "framework" context where appropriate

### 2. AGENTS.md — Terminology Fix (~5 lines changed)

**Change:** Line 5 "Go CLI skeleton/template generator" → identity consistent with README

**Proposed:**
```markdown
**ckeletin-go** is a production-ready Go CLI scaffold powered by an updatable framework layer. Key characteristics:
```

Also add one line explaining the dual nature:
```markdown
The `.ckeletin/` directory contains the framework (config, logging, validation, tasks).
Your code lives in `cmd/`, `internal/`, `pkg/`. Framework updates via `task ckeletin:update`.
```

### 3. CLAUDE.md — Minimal Change (~1 line)

CLAUDE.md points to AGENTS.md for project description. No direct identity claim to fix. Just verify the reference is correct after AGENTS.md changes.

### 4. CONTRIBUTING.md — Framework Awareness (~100 lines of changes)

**Current problems:**
- No mention of `.ckeletin/` directory
- Directory tree (line 71-77) shows old structure without `.ckeletin/`
- No mention of `task ckeletin:update`
- No mention of two-tier ADR system
- Coverage thresholds inconsistent (says 80% overall, AGENTS.md says 85%)

**Changes:**
1. **Update directory tree** (line 71-77): Add `.ckeletin/` with annotation
2. **Add "Framework vs Project" section** after Getting Started:
   - What lives in `.ckeletin/` (don't edit)
   - What's yours to edit (cmd/, internal/, pkg/, docs/adr/100+)
   - How framework updates work
3. **Add reference to two-tier ADR system**: Framework ADRs (000-099) vs Project ADRs (100+)
4. **Fix coverage threshold**: CONTRIBUTING.md line 385 says 80% overall minimum, but AGENTS.md/CLAUDE.md say 85%. Update CONTRIBUTING.md to 85% to match
5. **Move detailed code examples here** from README:
   - "Options Pattern for Command Configuration"
   - "Command Implementation Pattern" details
   - "Adding New Configuration Options"
6. **Update "read the architecture documentation" list** (line 55-58): Add AGENTS.md

### 5. ADR-010 — Substantial Revision (Identity + pkg/ Reality)

ADR-010 has an internal contradiction: the Context says "not a library" while the Decision says "with optional public packages" — and `pkg/checkmate/` IS a real public importable library. The ADR needs revision across multiple sections.

**Section: Context (lines 22-34)**

Current: "ckeletin-go is a CLI application skeleton, not a library"

Proposed:
```markdown
**ckeletin-go is a production-ready Go CLI scaffold powered by an updatable framework layer.**
It is primarily a CLI tool, but it also hosts public reusable packages in `pkg/` (e.g., `checkmate`).
This dual nature should be:
1. **Visible** — Structure shows "CLI-first, with optional public packages"
2. **Enforced** — `pkg/` packages must be standalone (no `internal/` imports)
3. **Documented** — Clear criteria for what belongs in `pkg/`
4. **Validated** — Automated checks enforce boundaries
```

Also revise lines 30-34 ("Without clear organization"):
```markdown
Without clear organization:
- Business logic might leak into root directory
- `pkg/` packages might depend on `internal/`, breaking standalone reusability
- Framework code (`.ckeletin/`) might get mixed with project code
- New developers won't know where code belongs
```

**Section: Alternative 1 (line 47)**

Current: "Why not: We're explicitly NOT a library"

Proposed: "Why chosen: We adopt this traditional layout — ckeletin-go is CLI-first but also provides public packages (e.g., `pkg/checkmate/`)"

(This changes Alternative 1 from "rejected" to "chosen with modifications", which is what actually happened.)

**Section: Consequences — Positive (lines 193-207)**

Current line 195: "No confusion about whether to import as library"
Proposed: "Clear intent — CLI is the product, `pkg/` packages are bonus reusable components"

Current lines 204-207 ("Prevents Scope Creep"):
```markdown
**3. Intentional Public API**
- `pkg/` packages require conscious decision and criteria checklist
- Forces quality commitment: docs, tests, API stability
- Maintains focus on CLI while allowing reusable components
```

**Section: Consequences — Negative (line 221)**

Current: "API Maintenance Burden (if using `pkg/`)"
Add note: "This is a real cost — `pkg/checkmate/` is actively maintained as a public package. This is an intentional trade-off."

### 6. ARCHITECTURE.md — Terminology Fix (~3 lines)

**Change line 29:**
```
"scaffold" → "production-ready Go CLI scaffold with an updatable framework layer"
```

**Update "Last Updated" date.**

### 7. cmd/root.go — Help Text Fix (~4 lines)

**Short field (line 208):**
```go
// Current:
Short: "A scaffold for building professional CLI applications in Go",
// Proposed:
Short: "A production-ready Go CLI application",
```

**Long field (line 262-263):**
```go
// Current:
RootCmd.Long = fmt.Sprintf(`%s is a scaffold project that helps you kickstart your Go CLI applications.
It integrates Cobra, Viper, Zerolog, and Bubble Tea, along with a testing framework.`, binaryName)
// Proposed:
RootCmd.Long = fmt.Sprintf(`%s is a production-ready Go CLI application built with ckeletin-go.
Powered by Cobra, Viper, Zerolog, and Bubble Tea with enforced architecture patterns.`, binaryName)
```

Note: After `task init`, `binaryName` will be the user's app name, so the description should describe the built app, not the scaffold. The Short field should be generic enough to work for any derived project.

## What NOT to Change

- **.ckeletin/README.md** — Already correctly says "framework". Leave it.
- **CHANGELOG.md** — Historical record. Don't retroactively change terminology.
- **ADRs 000-009, 011-014** — Only change ADR-010. Others don't have identity claims.
- **.goreleaser.yml** — Already uses "framework" in comments. Fine as-is.

## Content Movement Plan

| Content | FROM | TO | Reason |
|---------|------|----|--------|
| Options Pattern (50+ lines of code) | README | CONTRIBUTING.md | Developer guide, not user-facing |
| Command Implementation Pattern details | README | CONTRIBUTING.md | Developer guide |
| Adding New Configuration Options | README | CONTRIBUTING.md | Developer guide |
| Single Source of Truth for Names | README | CONTRIBUTING.md | Implementation detail |
| Tooling Best Practices | README | DELETE | Generic advice, adds no value |

## Terminology Consistency Rules

| Context | Use | Don't Use |
|---------|-----|-----------|
| Project identity (one-liner) | "production-ready Go CLI scaffold powered by an updatable framework layer" | "skeleton", "template generator", "boilerplate" |
| Referring to `.ckeletin/` | "framework layer" or "framework infrastructure" | "scaffold", "template" |
| Referring to `task init` workflow | "scaffold" or "project scaffold" | "framework", "generator" |
| Referring to `task ckeletin:update` | "framework update" | "scaffold update" |
| The project's dual nature | "scaffold and framework" | either term alone without context |

## Implementation Order

1. **AGENTS.md** — Quick fix, sets the terminology baseline
2. **CLAUDE.md** — Verify reference is correct
3. **cmd/root.go** — Quick code fix
4. **ADR-010** — Revise identity paragraph
5. **ARCHITECTURE.md** — Quick terminology fix
6. **CONTRIBUTING.md** — Add framework awareness + absorb README content
7. **README.md** — Full rewrite (depends on CONTRIBUTING absorbing content first)
8. **Run `task check`** — Verify nothing broke
9. **Commit all together**

## Verification

1. **Grep for inconsistencies:** `grep -rn "skeleton\|template generator\|boilerplate" *.md docs/ .ckeletin/docs/ cmd/root.go` — should find zero hits in project identity contexts. Exceptions: CHANGELOG.md (historical), ADR-001/ADR-009 (generic English usage of "boilerplate"/"skeletons" in technical discussion, not identity claims).
2. **`task check`** must pass (cmd/root.go change triggers tests)
3. **Golden files:** The cmd/root.go help text change does NOT affect golden files (golden files contain check result output, not help text). No golden file update needed.
4. **Manual review:** Read the new README as a first-time visitor. Does the dual identity come through clearly?
5. **Cross-reference check:** AGENTS.md, CLAUDE.md, README.md, CONTRIBUTING.md, ARCHITECTURE.md all use consistent terminology
6. **GitHub repo description:** Run `gh repo edit --description "A production-ready Go CLI scaffold powered by an updatable framework layer"` to update the GitHub repo description

## Notes for Implementer

- **Line numbers are approximate** — files may have been modified since this spec was written. Search for the quoted text strings rather than relying on line numbers.
- **Content moved to CONTRIBUTING.md** should be placed after the existing "Adding a New Command" section (line 119), which is the natural home for implementation patterns.
- **README section-level detail:** The README rewrite outline provides structure and guidance. The implementer should use judgment on prose while preserving all technical content from the current README. Key constraint: target ~600 lines (down from 983). The trim comes from moving developer-facing content to CONTRIBUTING.md and removing generic advice.

## Risks

- **README rewrite is large** — risk of losing valuable content. Mitigation: content moves to CONTRIBUTING, not deleted. Run content completeness audit after rewrite.
- **cmd/root.go change affects tests** — unit tests may reference the Short/Long text. Mitigation: search for "scaffold project" in test files and update if found.
- **SEO impact** — removing "skeleton" may affect discoverability. Mitigation: keep "scaffold" which serves the same search intent.

## Related Engineering Change: pkg/ Cleanup on Scaffold Init

**This is a separate implementation task, not part of the docs overhaul, but documented here because it emerged from the same analysis.**

### Problem

When `task init name=myapp` scaffolds a new project, `pkg/checkmate/` is inherited under the new module path. But checkmate is a ckeletin-go library — it shouldn't live in derived projects as if they authored it.

### Design

1. **Keep `pkg/checkmate/` in ckeletin-go** as a public library at `github.com/peiman/ckeletin-go/pkg/checkmate`
2. **Change `internal/check/` imports** to reference checkmate from the official ckeletin-go module, not a local copy
3. **Modify `scaffold-init.go`** to:
   - NOT replace `github.com/peiman/ckeletin-go/pkg/checkmate` import paths (add to skip list)
   - Clean out `pkg/` directory after init (remove ckeletin-go's packages)
   - Ensure `go.mod` includes `github.com/peiman/ckeletin-go` as a dependency
4. **Derived projects** get a clean `pkg/` directory and import checkmate as an external dependency

### Files Affected

- `.ckeletin/scripts/scaffold-init.go` — Add pkg/ cleanup step, skip checkmate imports
- `internal/check/*.go` — Change checkmate import path to absolute `github.com/peiman/ckeletin-go/pkg/checkmate`
- Any other files that import from `pkg/checkmate/`
- `go.mod` — Will need ckeletin-go as a dependency after init

### Implications

- Derived projects depend on ckeletin-go module (for checkmate). This is acceptable — the framework relationship is explicit.
- checkmate versioning now matters: derived projects pin to a ckeletin-go version
- The validate-package-organization script may need updates to handle external pkg/ imports
- Tests in `internal/check/` need verification after import path change
