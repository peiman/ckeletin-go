# Testing Strategy

This document describes the testing approach for ckeletin-go, with a focus on CLI output testing using golden files and structure validation.

## Overview

ckeletin-go uses a multi-layered testing strategy:

1. **Unit Tests** - Test individual functions and components
2. **Golden File Tests** - Snapshot testing for CLI output consistency
3. **Structure Validation Tests** - Integration tests for output structure
4. **Integration Tests** - End-to-end testing of commands

## Golden File Testing

### What Are Golden Files?

Golden files are "reference snapshots" of expected CLI output. Instead of writing dozens of assertions for each line of output, we save the entire expected output to a file and compare test results byte-for-byte against it.

**Analogy:** Think of golden files as "answer keys" for your tests. Just like a teacher compares student answers against the answer key, golden file tests compare actual output against the saved reference.

### How Golden Files Work

```go
// Traditional approach (verbose)
assert.Contains(t, output, "✅ All checks passed")
assert.Contains(t, output, "Code Quality")
assert.Contains(t, output, "Architecture Validation")
// ... 50 more assertions

// Golden file approach (concise)
goldie.Assert(t, "check-summary", []byte(output))
// Compares against testdata/check-summary.golden
```

### Golden Files in This Project

We use golden files to test the **check summary output** (`scripts/check-summary.sh`):

- **File:** `test/integration/testdata/check-summary.golden`
- **Tests:** Component output (not full integration)
- **Speed:** Very fast (~1 second)
- **Purpose:** Ensure summary format stays consistent

### When to Update Golden Files

Update golden files when you **intentionally** change output format:

- Improving summary layout
- Adding new check categories
- Changing emoji or formatting
- Updating success messages

**⚠️ NEVER blindly update without understanding what changed!**

### Workflow for Updating Golden Files

#### 1. Make Your Changes

Edit the output format:
```bash
vim scripts/check-summary.sh
```

#### 2. Update Golden Files

```bash
task test:golden:update
```

This will:
- Run tests with `GOLDEN_UPDATE=1`
- Generate new golden files
- Remind you to review changes

#### 3. Review Changes (CRITICAL!)

```bash
git diff test/integration/testdata/
```

**Ask yourself:**
- ✅ Do these changes match what I intended?
- ✅ Is the new output better than the old?
- ✅ Are there any typos or formatting issues?
- ❌ Are there unexpected changes?

#### 4. Commit Changes

Only commit if changes look correct:

```bash
git add test/integration/testdata/
git commit -m "feat: improve check summary format

- Changed summary header layout
- Added section separators
- Updated golden file to match"
```

### Golden File Best Practices

#### ✅ DO

- **Review golden file changes manually** before committing
- **Test single components** (like summary script) not entire systems
- **Normalize dynamic content** (paths, timestamps, temp directories)
- **Use golden files for user-facing output** that should stay consistent
- **Update golden files with code changes** in the same commit

#### ❌ DON'T

- **Don't blindly run `task test:golden:update`** without reviewing
- **Don't test non-deterministic output** (test parallelism, random data)
- **Don't use golden files for simple validations** (use assertions instead)
- **Don't commit golden files without understanding** what changed
- **Don't test entire integration with golden files** (use structure validation)

### Output Normalization

Golden files need **deterministic output**. We normalize:

1. **Paths** - `/Users/peiman/...` → `./...`
2. **Timings** - `1.23s` → `X.XXs`
3. **Temp Directories** - `/var/folders/.../Test123/` → `/tmp/TEMP_DIR/`

Normalization happens automatically in `test/integration/output_normalizer.go`.

### Troubleshooting

#### Golden File Test Fails

```
FAIL: testdata/check-summary.golden
Expected:
  ✅ All checks passed (15/15)
Got:
  ✅ All checks passed (16/16)
```

**Cause:** Output changed (added a check)

**Solution:**
1. Review: `git diff testdata/check-summary.golden`
2. If intentional: `task test:golden:update`
3. If not: Fix your code

#### Can't Reproduce Golden File

**Problem:** Golden file works on CI but fails locally

**Common causes:**
- Platform differences (Windows vs Mac vs Linux)
- Different Go versions
- Different tool versions

**Solution:**
- Check normalization covers platform differences
- Ensure dev tools are up to date: `task setup`

## Structure Validation Testing

### What is Structure Validation?

Structure validation tests verify the **organization and ordering** of output without caring about exact text. These complement golden files.

**Golden files say:** "Output looks EXACTLY like this"
**Structure validation says:** "Output has THESE sections in THIS order"

### Structure Tests in This Project

Located in `test/integration/task_check_structure_test.go`:

- **TestTaskCheckOutputStructure** - Verifies all sections appear in order
- **TestTaskCheckCategoryHeaders** - Ensures all category headers present
- **TestTaskCheckSuccessIndicators** - Validates success markers

### When Structure Tests Catch Issues

Structure validation catches:

- Section reordering (Tests moved before Code Quality)
- Missing sections (Dependencies section removed)
- Wrong success markers (❌ instead of ✅)
- Summary format changes (15/15 → 15 checks)

Golden files **won't** catch these if they're outside the summary component.

### Structure vs Golden Files

| Aspect | Golden Files | Structure Validation |
|--------|-------------|----------------------|
| **Tests** | Component output | Full integration |
| **Speed** | Fast (~1s) | Slower (~19s) |
| **Brittleness** | High (exact match) | Low (pattern match) |
| **What it catches** | Format changes | Ordering/presence changes |
| **Update frequency** | Often (UI changes) | Rarely (structure stable) |

**Use both:** Golden files for component quality, structure tests for integration confidence.

## Test Commands

### Running Tests

```bash
# All tests
task test

# Just golden file tests
task test:golden

# Just structure validation
go test ./test/integration -run Structure -v

# Skip slow tests
go test ./test/integration -short
```

### Updating Golden Files

```bash
# Update golden files
task test:golden:update

# Review changes (REQUIRED!)
git diff test/integration/testdata/

# Commit if looks good
git add test/integration/testdata/
git commit -m "feat: update summary format"
```

### Coverage

```bash
# Generate coverage report
task test:coverage:html

# View coverage
open coverage.html  # macOS
```

## Test Organization

```
test/
├── integration/
│   ├── output_normalizer.go          # Normalization helpers
│   ├── output_normalizer_test.go     # Normalization tests
│   ├── task_check_golden_test.go     # Golden file tests
│   ├── task_check_structure_test.go  # Structure validation
│   └── testdata/
│       └── check-summary.golden      # Golden file snapshots
```

## Testing Philosophy

### Why This Approach?

1. **Golden files** ensure user-visible output quality
2. **Structure validation** ensures integration correctness
3. **Normalization** makes tests portable and reliable
4. **Fast feedback** - golden tests run in ~1 second

### The Testing Pyramid

```
      /\
     /  \  Integration (structure validation)
    /____\
   /      \
  / Golden \ Component tests (golden files)
 /  Files   \
/____________\
  Unit Tests   Fast, focused tests
```

**Bottom (Most):** Unit tests - Fast, many
**Middle:** Golden files - Medium speed, focused on components
**Top (Fewest):** Structure validation - Slower, full integration

## Common Patterns

### Testing a New CLI Output Format

```go
func TestNewFeatureOutput(t *testing.T) {
    output := runCommand(t, "new-feature")
    normalized := NormalizeCheckOutput(output)

    g := goldie.New(t)
    if os.Getenv("GOLDEN_UPDATE") != "" {
        g.Update(t, "new-feature", []byte(normalized))
    } else {
        g.Assert(t, "new-feature", []byte(normalized))
    }
}
```

### Validating Output Structure

```go
func TestNewFeatureStructure(t *testing.T) {
    output := runCommand(t, "new-feature")

    // Verify sections appear
    assert.Contains(t, output, "Header Section")
    assert.Contains(t, output, "Body Section")

    // Verify order
    headerPos := strings.Index(output, "Header")
    bodyPos := strings.Index(output, "Body")
    assert.Less(t, headerPos, bodyPos)
}
```

## References

- **Golden Files Library:** [goldie/v2](https://github.com/sebdah/goldie)
- **ADR-003:** Dependency Injection Over Mocking (.ckeletin/docs/adr/003-dependency-injection-over-mocking.md)
- **Testify Assertions:** [testify/assert](https://github.com/stretchr/testify)

## FAQ

### Q: When should I use golden files vs assertions?

**A:** Use golden files for:
- Multi-line output that should stay consistent
- User-facing messages and formatting
- Output where exact format matters

Use assertions for:
- Simple validations (one or two strings)
- Business logic results
- When exact format doesn't matter

### Q: Why did my golden test fail after merging main?

**A:** Someone updated the output format. Run:
```bash
task test:golden:update
git diff testdata/
```

If changes look correct, commit them. If not, revert their change.

### Q: Can I test the entire `task check` output with golden files?

**A:** No, that's why we have structure validation. Full `task check` output includes test results which run in parallel (non-deterministic order). Use:
- Golden files for the **summary component**
- Structure validation for **full integration**

### Q: How do I add a new golden file test?

1. Write test following pattern in `task_check_golden_test.go`
2. Run with `GOLDEN_UPDATE=1` to create golden file
3. Review golden file content
4. Run test normally to verify it passes
5. Commit test + golden file together
