# ADR-012: Build Tags for Dev-Only Commands

## Status
Accepted

## Context

### The Problem: Developer Tools vs Production Binaries

When building CLI applications, developers need utilities during development that users shouldn't see in production:

- **Config inspection** - View merged configuration, debug config issues
- **Environment health checks** - Verify tools installed, validate setup
- **Debug helpers** - Runtime inspection, troubleshooting tools
- **Code generation** - Scaffold commands, generate test data

**Without proper separation:**
- Production binaries bloated with dev-only code
- Users confused by internal/debug commands in help text
- Accidental usage of dev commands in production
- Increased binary size and attack surface
- No clear distinction between user-facing and dev-facing features

**Real-world consequences:**
- Debug commands exposed in production → security issues
- Internal tooling visible to end users → poor UX
- Dev utilities shipped to customers → unprofessional
- Larger binaries → slower downloads, more disk space
- Support burden from users running dev commands

### Why ckeletin-go Needs This

**ckeletin-go is a scaffold** with unique requirements:

1. **Teach best practices** - Show how to properly separate dev/prod code
2. **Developer ergonomics** - Make local development easy with built-in tools
3. **Production polish** - Ship clean, focused binaries to end users
4. **Zero runtime cost** - Dev features completely removed from prod builds
5. **Standard Go patterns** - Use idiomatic Go build tags, not custom solutions

### Alternatives Considered

**1. Feature Flags (Runtime Toggle)**
```go
if os.Getenv("DEV_MODE") == "true" {
    rootCmd.AddCommand(devCmd)
}
```
- **Pros**: Simple, one binary, easy to toggle
- **Cons**:
  - Dev code shipped in production (security risk)
  - Runtime overhead (env var checks)
  - Easy to accidentally enable in prod
  - Larger binary size
  - Flag bypass possible
- **Why not**: Unacceptable security/size tradeoff

**2. Separate Binary (e.g., ckeletin-go-dev)**
- **Pros**: Complete separation, no confusion
- **Cons**:
  - Two binaries to maintain
  - Duplicate build/test/release pipelines
  - User confusion (which binary to use?)
  - Doubled CI time
  - Code duplication risk
- **Why not**: Maintenance burden too high

**3. Submodule/Plugin System**
```go
// Load dev commands from external module
plugin.Load("github.com/peiman/ckeletin-go-devtools")
```
- **Pros**: Complete modularity, extensible
- **Cons**:
  - Complex architecture
  - Plugin loading overhead
  - Multiple repositories to maintain
  - Overkill for simple dev commands
- **Why not**: Overengineered for this use case

**4. Hidden Commands (No Build Tags)**
```go
devCmd := &cobra.Command{
    Hidden: true,  // Hidden in help, but still in binary
}
```
- **Pros**: Simple, no build complexity
- **Cons**:
  - Code still in production binary
  - Users can still discover and run commands
  - No actual separation
  - Security through obscurity
- **Why not**: Doesn't solve core problem

**5. Build Tags (Conditional Compilation)**
```go
//go:build dev

package cmd

var devCmd = &cobra.Command{...}
```
- **Pros**:
  - Zero runtime overhead
  - Standard Go practice
  - Complete code exclusion from prod
  - Small binary size in prod
  - Compile-time enforcement
- **Cons**:
  - Must test both build modes
  - CI complexity (two builds)
  - Developers must understand build tags
- **Why THIS**: Industry standard, zero overhead, proper separation

## Decision

We adopt **Go build tags** for dev-only commands:

### Build Tag Strategy

**Development builds** (local development):
```bash
go build -tags dev
# Includes dev commands: dev config, dev doctor, etc.
```

**Production builds** (releases):
```bash
go build
# No tags = dev commands completely excluded
```

### Implementation Rules

1. **All dev commands use `//go:build dev` tag**
   - File header: `//go:build dev`
   - Automatically excluded without tag

2. **Default to dev builds locally**
   - `task build` → builds with `-tags dev`
   - Developers get dev tools by default
   - Matches actual development workflow

3. **Explicit production builds**
   - `task build:prod` → builds without tags
   - Production is intentional, not default
   - Prevents accidental dev → prod leakage

4. **CI tests both modes**
   - Test dev build (commands exist)
   - Test prod build (commands hidden)
   - Both must pass before release

5. **GoReleaser releases prod only**
   - No tags in release config
   - Production binaries ship to users
   - Dev tools never in releases

### Initial Dev Commands

**Minimal scope** (two commands to start):

1. **`dev config`** - Configuration inspector
   - List all config keys from registry
   - Show effective config (merged sources)
   - Validate current configuration
   - Export to JSON format

2. **`dev doctor`** - Environment health check
   - Verify required tools installed
   - Check Go version compatibility
   - Validate project structure
   - Check git status and dependencies

More commands can be added later as `dev <subcommand>` following the same pattern.

## Consequences

### Positive

✅ **Zero runtime overhead** - Dev code literally doesn't exist in prod binaries
✅ **Standard Go practice** - Build tags are idiomatic Go, well-documented
✅ **Clean separation** - No runtime flags, env vars, or feature toggles
✅ **Smaller prod binaries** - Dev code excluded at compile time
✅ **Security** - No accidental dev command execution in production
✅ **Developer ergonomics** - Default dev builds make local work easier
✅ **Educational** - Teaches proper Go conditional compilation

### Negative

⚠️ **Two build modes to test** - CI must verify both dev and prod builds
⚠️ **Learning curve** - Developers must understand build tags (minimal, standard Go)
⚠️ **CI complexity** - Additional job to test dev/prod separation
⚠️ **Breaking change** - `task build` behavior changes (dev default)

### Mitigation Strategies

**For CI complexity:**
- Single focused job: `test-dev-prod` (~2 min)
- Clear failure messages if dev commands leak
- Blocks release if either mode fails

**For learning curve:**
- Document in CLAUDE.md (project guidelines)
- Create `docs/development.md` (dev guide)
- Comments in code explain build tags
- Task commands abstract complexity

**For breaking change:**
- Clear CHANGELOG.md entry
- Update CLAUDE.md prominently
- Semantic commit message
- `task build:prod` available for old behavior

## Implementation

### File Organization

```
cmd/
  dev.go           # //go:build dev - Root dev command
  dev_config.go    # //go:build dev - Config subcommand
  dev_doctor.go    # //go:build dev - Doctor subcommand

internal/dev/
  config.go        # Config inspection logic
  doctor.go        # Health check logic
  (tests)          # Unit tests (no build tags)
```

### Task Commands

**Development (default):**
- `task build` → `-tags dev` (local development)
- `task test` → `-tags dev` (test with dev commands)
- `task run` → runs dev build

**Production:**
- `task build:prod` → no tags (production binary)
- `task test:prod` → no tags (test without dev)

**Explicit (consistency):**
- `task build:dev` → same as `task build`
- `task test:dev` → same as `task test`

### CI Strategy

**New job: `test-dev-prod`**
1. Build dev: `go build -tags dev -o ckeletin-go-dev`
2. Test dev commands exist: `./ckeletin-go-dev dev --help`
3. Run tests with dev tag: `go test -tags dev ./...`
4. Build prod: `go build -o ckeletin-go-prod`
5. Test dev commands hidden: `./ckeletin-go-prod dev` (must fail)
6. Run tests without tags: `go test ./...`

Must pass before `release` job runs.

### GoReleaser Configuration

No changes needed - config already has no build tags:
```yaml
builds:
  - main: ./main.go
    # No tags field = production build
    flags:
      - -trimpath
    # Dev commands excluded automatically
```

Add comment for clarity:
```yaml
# Production builds only (dev commands excluded)
# Dev builds use: go build -tags dev
```

## Build Tag Syntax

### File Header
```go
//go:build dev

package cmd
```

**Important:**
- Must be first line (before package)
- Blank line after directive
- Old syntax `// +build dev` deprecated (Go 1.17+)

### Behavior
- **With tag:** `go build -tags dev` → file included
- **Without tag:** `go build` → file excluded
- Build fails if dev command referenced from non-dev file

## Testing Strategy

### Integration Tests
```go
func TestDevCommandExistsInDevBuild(t *testing.T) {
    // Build with dev tag
    // Verify `./binary dev --help` succeeds
}

func TestDevCommandHiddenInProdBuild(t *testing.T) {
    // Build without tags
    // Verify `./binary dev` fails
}
```

### Unit Tests (No Tags)
```go
// internal/dev/config_test.go - NO build tag
// Tests logic, not conditional compilation
func TestConfigInspector(t *testing.T) { ... }
```

## Migration Path

1. Create internal logic (`internal/dev/*`)
2. Create command files (`cmd/dev*.go`)
3. Update Taskfile.yml
4. Test locally (both modes)
5. Update CI workflow
6. Update GoReleaser comments
7. Create documentation
8. Add integration tests

## Related ADRs

- **ADR-000**: Task-Based Workflow - Task commands provide interface
- **ADR-001**: Ultra-Thin Command Pattern - Dev commands follow same pattern
- **ADR-008**: Release Automation - GoReleaser handles prod builds

## References

- [Go Build Constraints](https://pkg.go.dev/cmd/go#hdr-Build_constraints) - Official Go documentation
- [Build Tags Tutorial](https://www.digitalocean.com/community/tutorials/customizing-go-binaries-with-build-tags) - DigitalOcean guide
- [Conditional Compilation](https://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool) - Dave Cheney's blog

## Future Considerations

**Additional dev commands** (as needed):
- `dev inspect` - Runtime inspection (config, env, build info)
- `dev generate` - Code generation helpers
- `dev benchmark` - Quick performance checks
- `dev migrate` - Data migration tools

**Pattern established** - Easy to add more dev utilities:
1. Create `cmd/dev_<name>.go` with `//go:build dev`
2. Add subcommand under `devCmd`
3. Automatically excluded from production

## Breaking Changes

**⚠️ BREAKING: `task build` now creates dev builds**

**Before:**
```bash
task build  # Production build (no tags)
```

**After:**
```bash
task build       # Dev build (includes dev commands)
task build:prod  # Production build (was: task build)
```

**Rationale:**
- Developers need dev tools during local work
- Production builds are intentional (releases, explicit)
- Matches actual development workflow
- Easy to create prod builds when needed

**Migration:**
- Update any scripts using `task build` for production
- Use `task build:prod` for production builds
- `task build` continues working (just includes dev commands now)
