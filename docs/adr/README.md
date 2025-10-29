# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for ckeletin-go.

## What is an ADR?

An Architecture Decision Record (ADR) captures an important architectural decision made along with its context and consequences.

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

- [ADR-001](001-ultra-thin-command-pattern.md) - Ultra-Thin Command Pattern
- [ADR-002](002-centralized-configuration-registry.md) - Centralized Configuration Registry
- [ADR-003](003-dependency-injection-over-mocking.md) - Dependency Injection Over Mocking
- [ADR-004](004-security-validation-in-config.md) - Security Validation in Configuration
- [ADR-005](005-auto-generated-config-constants.md) - Auto-Generated Configuration Constants
- [ADR-006](006-structured-logging-with-zerolog.md) - Structured Logging with Zerolog
- [ADR-007](007-bubble-tea-for-interactive-ui.md) - Bubble Tea for Interactive UI

## Creating a New ADR

1. Copy the template
2. Number it sequentially
3. Fill in all sections thoughtfully
4. Update this index
5. Commit with the changes it describes

## References

- [ADR documentation](https://adr.github.io/)
- [Michael Nygard's original article](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
