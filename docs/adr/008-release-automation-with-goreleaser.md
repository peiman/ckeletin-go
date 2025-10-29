# ADR-008: Release Automation with GoReleaser

## Status
Accepted

## Context

As the project matures, we need a professional release process that:
- Builds binaries for multiple platforms (Linux, macOS, Windows)
- Supports multiple architectures (amd64, arm64)
- Generates release artifacts (archives, checksums, SBOMs)
- Automates GitHub releases with proper changelog
- Provides easy installation methods for users (Homebrew, direct download)
- Ensures reproducible and verifiable builds

Manual release processes are:
- Time-consuming and error-prone
- Inconsistent across releases
- Limited to single-platform builds
- Difficult to verify and audit
- Missing modern distribution channels

## Decision

We adopt **GoReleaser** as our release automation tool with the following configuration:

### Multi-Platform Builds
- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

### Release Artifacts
1. **Binaries**: Cross-compiled for all platforms
2. **Archives**: tar.gz (Unix) and zip (Windows)
3. **Checksums**: SHA256 checksums for verification
4. **SBOM**: Software Bill of Materials in SPDX format

### Distribution Channels
1. **GitHub Releases**: Automated release creation with changelog
2. **Homebrew Tap**: `peiman/tap/ckeletin-go` for macOS/Linux users
3. **Direct Downloads**: Platform-specific archives from releases

### Version Information
Maintain existing ldflags injection pattern:
- `cmd.binaryName` - Project name
- `cmd.Version` - Git tag version
- `cmd.Commit` - Git commit hash
- `cmd.Date` - Build date

### Development Workflow
- **Local Testing**: `task release:test` - Snapshot builds without publishing
- **Local Builds**: `task release:build` - Build without release
- **Clean Artifacts**: `task release:clean` - Remove build artifacts
- **CI/CD**: Automatic release on semantic version tags

## Consequences

### Positive

- **Professional Distribution**: Users get native binaries for their platform
- **Security**: Checksums and SBOMs for verification and compliance
- **Automation**: Reduces manual work and human error
- **Consistency**: Every release follows the same process
- **Discoverability**: Homebrew makes installation easier for users
- **Scalability**: Easy to add new platforms or distribution channels
- **Reproducibility**: Consistent build environment in CI

### Negative

- **Complexity**: Additional tool and configuration to maintain
- **Dependencies**: Requires GoReleaser in development environment
- **Secrets Management**: Needs GitHub tokens for releases and Homebrew tap
- **Build Time**: Cross-compilation takes longer than single-platform builds

### Mitigations

- **Documentation**: Clear release process in README.md
- **Task Commands**: `task release:*` abstracts GoReleaser complexity
- **CI Integration**: Automated process requires minimal manual intervention
- **Local Testing**: Snapshot builds allow testing without tags
- **Version Check**: `task release:check` validates GoReleaser installation

## Implementation Details

### Configuration
- `.goreleaser.yml` - Main configuration file
- `.github/workflows/ci.yml` - CI integration with `goreleaser-action`
- `Taskfile.yml` - Development tasks for local testing

### GitHub Secrets Required
- `CKELETIN_GITHUB_TOKEN` - For creating GitHub releases
- `HOMEBREW_TAP_GITHUB_TOKEN` - For updating Homebrew tap

### Release Process
1. Ensure all changes are committed and pushed
2. Run quality checks: `task check`
3. Update CHANGELOG.md with release notes
4. Create and push semantic version tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
5. Push tag: `git push origin v1.0.0`
6. CI automatically builds and releases via GoReleaser

### Semantic Versioning
Follow [Semantic Versioning 2.0.0](https://semver.org/):
- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features, backwards compatible)
- `v1.1.1` - Patch release (bug fixes)
- `v1.0.0-beta.1` - Pre-release versions

## Compliance Validation

Test GoReleaser configuration locally:

```bash
# Check if GoReleaser is installed
task release:check

# Test build without releasing
task release:test

# Clean artifacts
task release:clean
```

## Related ADRs

None currently. This ADR establishes release automation as a foundational practice.

## References

- [GoReleaser Documentation](https://goreleaser.com/)
- [Semantic Versioning](https://semver.org/)
- `.goreleaser.yml` - Configuration file
- `Taskfile.yml` - Release tasks
- `.github/workflows/ci.yml` - CI integration
