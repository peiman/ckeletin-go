# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for ckeletin-go.

## What is an ADR?

An Architecture Decision Record (ADR) captures an important architectural decision made along with its context and consequences.

## Reading Guide

**New to ckeletin-go?** Start here for the fastest path to understanding the architecture:

### 1. **Begin with the System Overview** (~20 min)

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Shows WHAT the system is
  - Component structure and interactions
  - Initialization sequence
  - Configuration flow
  - Command execution lifecycle
  - How all ADRs work together

### 2. **Understand the Foundation** (~10 min)

- **[ADR-000](000-task-based-single-source-of-truth.md)** - Task-Based Workflow (Foundational)
  - Why task commands are the single source of truth
  - How local development mirrors CI/CD
  - Pattern enforcement automation

### 3. **Dive into Specific Areas** (Based on your interest)

**Working on commands?**
- [ADR-001](001-ultra-thin-command-pattern.md) - Ultra-thin commands (~20-30 lines)
- [ADR-002](002-centralized-configuration-registry.md) - Config registry
- [ADR-005](005-auto-generated-config-constants.md) - Type-safe constants

**Writing tests?**
- [ADR-003](003-dependency-injection-over-mocking.md) - DI over mocking

**Security or compliance?**
- [ADR-004](004-security-validation-in-config.md) - Security validation
- [ADR-011](011-license-compliance.md) - License compliance

**Infrastructure?**
- [ADR-006](006-structured-logging-with-zerolog.md) - Structured logging
- [ADR-007](007-bubble-tea-for-interactive-ui.md) - Interactive UIs
- [ADR-008](008-release-automation-with-goreleaser.md) - Release automation

### 4. **Making Architectural Changes?**

1. Read **ARCHITECTURE.md** to understand system-wide impact
2. Read relevant ADRs to understand constraints and rationale
3. Validate your changes with `task check` (enforces all ADR patterns)

---

## Document Separation of Concerns

**ARCHITECTURE.md vs ADRs - What's the Difference?**

| Document | Purpose | Contains |
|----------|---------|----------|
| **ARCHITECTURE.md** | System structure (WHAT) | Components, flows, diagrams, interactions |
| **ADRs** | Decisions (WHY) | Context, rationale, alternatives, consequences |

**Example:**
- ARCHITECTURE.md: "Commands are ~20-30 lines and delegate to business logic"
- ADR-001: "WHY commands are thin: testability, separation of concerns, alternatives considered"

This separation prevents duplication and is enforced by `task validate:architecture`.

**Known Gaps:**

Some architectural decisions in ARCHITECTURE.md lack formal ADRs and are marked with `<!-- TODO: ADR -->` comments. These markers:
- Make missing ADRs self-documenting (single source of truth)
- Are excluded from validation failures (allowed WHY statements)
- Track future ADR work without duplicating content

To find all missing ADRs, search ARCHITECTURE.md:
```bash
grep -n "<!-- TODO: ADR" docs/adr/ARCHITECTURE.md
```

This maintains SSOT - the TODO markers are the **only** place where missing ADRs are tracked.

---

## ADR Format

Each ADR follows this structure:

```markdown
# ADR-###: Title

## Status
[Proposed | Accepted | Deprecated | Superseded]

## Context
What is the issue that we're seeing that is motivating this decision or change?

## Decision
What is the change that we're proposing and/or doing?

## Consequences
What becomes easier or more difficult to do because of this change?
```

## Index

- [ADR-000](000-task-based-single-source-of-truth.md) - Task-Based Single Source of Truth (Foundational)
- [ADR-001](001-ultra-thin-command-pattern.md) - Ultra-Thin Command Pattern
- [ADR-002](002-centralized-configuration-registry.md) - Centralized Configuration Registry
- [ADR-003](003-dependency-injection-over-mocking.md) - Dependency Injection Over Mocking
- [ADR-004](004-security-validation-in-config.md) - Security Validation in Configuration
- [ADR-005](005-auto-generated-config-constants.md) - Auto-Generated Configuration Constants
- [ADR-006](006-structured-logging-with-zerolog.md) - Structured Logging with Zerolog
- [ADR-007](007-bubble-tea-for-interactive-ui.md) - Bubble Tea for Interactive UI
- [ADR-008](008-release-automation-with-goreleaser.md) - Release Automation with GoReleaser
- [ADR-009](009-layered-architecture-pattern.md) - Layered Architecture Pattern
- [ADR-010](010-package-organization-strategy.md) - Package Organization Strategy
- [ADR-011](011-license-compliance.md) - License Compliance Strategy (Dual-Tool Approach)
- [ADR-012](012-dev-commands-build-tags.md) - Build Tags for Dev-Only Commands

## Creating a New ADR

1. Copy the template
2. Number it sequentially
3. Fill in all sections thoughtfully
4. Update this index
5. Commit with the changes it describes

## References

- [ADR documentation](https://adr.github.io/)
- [Michael Nygard's original article](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
