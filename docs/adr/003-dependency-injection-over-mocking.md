# ADR-003: Dependency Injection Over Mocking

## Status
Accepted

## Context

Testing often requires mocking external dependencies like UI frameworks, file systems, and loggers. Traditional approaches use mocking frameworks which add complexity and make tests fragile.

## Decision

Use **dependency injection with interfaces** instead of mocking frameworks:

- Define interfaces for external dependencies (UIRunner, io.Writer)
- Inject dependencies via constructor parameters
- Use real implementations in production
- Use simple test implementations for testing

### Example

```go
// internal/ui/ui.go
type UIRunner interface {
    RunUI(message, color string) error
}

// internal/ping/ping.go
type Executor struct {
    uiRunner ui.UIRunner  // Interface, not concrete type
    writer   io.Writer     // Standard interface
}

func NewExecutor(cfg Config, uiRunner ui.UIRunner, writer io.Writer) *Executor {
    return &Executor{cfg: cfg, uiRunner: uiRunner, writer: writer}
}

// Testing
func TestPing(t *testing.T) {
    mockUI := &mockUIRunner{}  // Simple struct
    executor := ping.NewExecutor(cfg, mockUI, &bytes.Buffer{})
    executor.Execute()
}
```

## Consequences

### Positive
- Simple, understandable tests
- No mocking framework dependency
- Interfaces clarify dependencies
- Easy to swap implementations
- Tests remain maintainable

### Negative
- Manual interface implementation for tests
- More code to write initially

## Enforcement

Dependency injection is enforced through multiple architectural layers rather than a dedicated validation script:

**1. Layered Architecture Validation** (ADR-009)
```bash
task validate:layering  # Prevents business logic from importing CLI frameworks
```
- Business logic cannot import `cobra` or `cmd/`
- Forces interface-based dependencies as consequence
- go-arch-lint enforces separation automatically

**2. Testing Requirements**
- Minimum 80% coverage requirement
- Interface-based code is naturally more testable
- Coverage targets incentivize proper DI patterns

**3. Code Organization**
- `internal/ui/ui.go` defines `UIRunner` interface (not concrete type)
- `internal/ui/mock.go` provides test implementation
- Pattern demonstrated in reference implementations

**4. Compile-Time Enforcement**
- Go type system enforces interface contracts
- Constructor functions require interface parameters
- Misuse causes compile errors, not runtime failures

**Why No Dedicated Script:**
DI is enforced naturally through Go's type system and the layered architecture. Adding a script would duplicate what the compiler and go-arch-lint already verify. The pattern is "enforced by design" rather than "enforced by validation."

## References
- `internal/ui/ui.go` - UIRunner interface
- `internal/ui/mock.go` - Test implementation
- `cmd/ping_test.go` - Usage in tests
