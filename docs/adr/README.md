# Project Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for your **project-specific** decisions.

## Two-Tier ADR System

| Location | Numbers | Purpose |
|----------|---------|---------|
| `.ckeletin/docs/adr/` | 000-099 | **Framework** decisions (updated with framework) |
| `docs/adr/` | 100+ | **Project** decisions (your custom choices) |

## Framework ADRs (000-099)

Framework ADRs document decisions about the ckeletin infrastructure:

- [ADR-000](.ckeletin/docs/adr/000-task-based-single-source-of-truth.md) - Task-Based Workflow (Foundational)
- [ADR-001](.ckeletin/docs/adr/001-ultra-thin-command-pattern.md) - Ultra-Thin Command Pattern
- [ADR-002](.ckeletin/docs/adr/002-centralized-configuration-registry.md) - Centralized Configuration Registry
- [ADR-003](.ckeletin/docs/adr/003-dependency-injection-over-mocking.md) - Dependency Injection Over Mocking
- [ADR-004](.ckeletin/docs/adr/004-security-validation-in-config.md) - Security Validation in Configuration
- [ADR-005](.ckeletin/docs/adr/005-auto-generated-config-constants.md) - Auto-Generated Config Constants
- [ADR-006](.ckeletin/docs/adr/006-structured-logging-with-zerolog.md) - Structured Logging with Zerolog
- [ADR-007](.ckeletin/docs/adr/007-bubble-tea-for-interactive-ui.md) - Bubble Tea for Interactive UI
- [ADR-008](.ckeletin/docs/adr/008-release-automation-with-goreleaser.md) - Release Automation with GoReleaser
- [ADR-009](.ckeletin/docs/adr/009-layered-architecture-pattern.md) - Layered Architecture Pattern
- [ADR-010](.ckeletin/docs/adr/010-package-organization-strategy.md) - Package Organization Strategy
- [ADR-011](.ckeletin/docs/adr/011-license-compliance.md) - License Compliance Strategy

See [.ckeletin/docs/adr/](../../.ckeletin/docs/adr/) for full framework documentation.

## Project ADRs (100+)

Document your project-specific architectural decisions here:

```markdown
docs/adr/
├── README.md        # This file
├── TEMPLATE.md      # Template for new ADRs
├── 100-*.md         # Your first project ADR
├── 101-*.md         # Your second project ADR
└── ...
```

### Examples of Project ADRs

- Database choice (PostgreSQL vs MySQL vs SQLite)
- API design patterns (REST vs GraphQL)
- Authentication strategy (JWT vs sessions)
- Deployment architecture (Kubernetes vs serverless)
- Third-party service integrations

## Creating a New Project ADR

1. Copy the template:
   ```bash
   cp docs/adr/TEMPLATE.md docs/adr/100-my-decision.md
   ```

2. Fill in all sections:
   - Status, Context, Decision, Consequences

3. Update this README's index (below)

4. Commit with your implementation

## Project ADR Index

*(Add your project ADRs here as you create them)*

<!-- Example:
- [ADR-100](100-database-selection.md) - Database Selection (PostgreSQL)
- [ADR-101](101-api-versioning.md) - API Versioning Strategy
-->

## ADR Format

Each ADR follows this structure:

```markdown
# ADR-###: Title

## Status
[Proposed | Accepted | Deprecated | Superseded]

## Context
What is the issue motivating this decision?

## Decision
What is the change we're proposing/doing?

## Consequences
What becomes easier or more difficult?
```

## References

- [ADR documentation](https://adr.github.io/)
- [Michael Nygard's original article](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
