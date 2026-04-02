# ADR-100: Plugin/Extension Architecture

## Status

Deferred

## Context

The ckeletin-go framework is currently all-or-nothing: users get the full `.ckeletin/` framework layer or nothing. There is no mechanism to add composable plugins (e.g., database support, gRPC scaffolding, API client generation) that extend the framework.

Several users or use cases might benefit from a plugin system:
- Adding database migrations and ORM setup
- Adding gRPC/protobuf scaffolding
- Adding API client generation
- Adding authentication middleware patterns
- Adding observability (OpenTelemetry) scaffolding

However, the project is at an early adoption stage (12 stars, 1 fork as of April 2026), and building a plugin system adds significant architectural complexity.

## Decision

Defer the plugin/extension architecture until adoption criteria are met:

1. **3+ external users** request composable extensions
2. **Adoption reaches 50+ stars** (indicating sufficient community interest)
3. **A concrete plugin use case** emerges that cannot be reasonably solved by adding code to `internal/`

Until then, users who need extensions should add them directly to their `internal/` packages, following the existing patterns.

## Consequences

### Positive

- Framework stays simple and maintainable
- No premature abstraction
- Development effort focuses on core quality (testing, enforcement, AI-agent readiness)
- Users learn the patterns by building directly in `internal/`

### Negative

- Users who want reusable extensions must build them ad hoc
- No ecosystem of shared plugins
- May slow adoption if users see the framework as inflexible

### When to Revisit

Review this decision when any of the three adoption criteria are met. At that point, investigate:
- Plugin discovery and loading mechanism
- Plugin interface contract and versioning
- Compatibility with `task ckeletin:update`
- Plugin testing and quality enforcement
- How plugins interact with the AI-agent configuration stack
