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

## References
- `internal/ui/ui.go` - UIRunner interface
- `internal/ui/mock.go` - Test implementation
- `cmd/ping_test.go` - Usage in tests
