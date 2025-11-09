# Test Fixtures

This directory contains test fixtures used for testing various components of the ckeletin-go CLI application.

## Directory Structure

```
testdata/
├── README.md          # This file
├── config/            # Configuration-related test fixtures
│   ├── valid.yaml     # Basic valid configuration (YAML format)
│   ├── valid.json     # Basic valid configuration (JSON format)
│   ├── partial.yaml   # Partial config for testing default value merging
│   ├── env_override.yaml # Base config for testing environment variable overrides
│   ├── invalid.yaml   # Intentionally invalid YAML for testing error handling
│   └── empty.yaml     # Empty config file for testing default values
├── docs/              # Documentation generation test fixtures
│   ├── config.yaml    # Configuration for testing the docs command
│   └── expected/      # Expected documentation output files
│       ├── docs_markdown.md # Expected markdown output
│       └── docs_yaml.yaml   # Expected YAML output
├── logger/            # Logging system test fixtures
│   └── config.yaml    # Configuration for testing logging systems
└── ui/                # UI component test fixtures
    └── config.yaml    # Configuration for testing UI components
```

## Usage in Tests

These fixtures are organized by component to:

1. **Improve Scalability**: Easy to add new fixtures for specific components
2. **Better Organization**: Clear separation between different test categories
3. **Reduce Naming Conflicts**: Multiple files can be named `config.yaml` in different contexts
4. **Self-Documenting**: Directory structure indicates fixture purpose

### Example Usage

```go
// In cmd/ping_test.go
testFixturePath: "../testdata/config/valid.yaml"

// In docs/generator_test.go
testFixturePath: "../testdata/docs/config.yaml"
expectedOutput: "../testdata/docs/expected/docs_markdown.md"
```

## Config Files Reference

### config/ (General Configuration)

| Filename | Purpose |
|----------|---------|
| `valid.yaml` | Basic valid configuration for general testing |
| `valid.json` | JSON configuration for testing format compatibility |
| `partial.yaml` | Partial config for testing default value merging |
| `env_override.yaml` | Base config for testing environment variable overrides |
| `invalid.yaml` | Intentionally invalid YAML for testing error handling |
| `empty.yaml` | Empty config file for testing default values |

### docs/ (Documentation Generation)

| Filename | Purpose |
|----------|---------|
| `config.yaml` | Configuration for testing the docs command |
| `expected/docs_markdown.md` | Expected markdown output for validation |
| `expected/docs_yaml.yaml` | Expected YAML output for validation |

### logger/ (Logging System)

| Filename | Purpose |
|----------|---------|
| `config.yaml` | Configuration for testing dual logging, rotation, sampling |

### ui/ (UI Components)

| Filename | Purpose |
|----------|---------|
| `config.yaml` | Configuration for testing Bubble Tea UI components |

## Adding New Test Fixtures

When adding new test fixtures, follow these guidelines:

1. **Choose the Right Directory**: Place fixtures in the subdirectory matching the component being tested
   - General config tests → `config/`
   - Component-specific tests → `<component>/`
   - New component? Create a new subdirectory

2. **Use Clear Filenames**:
   - Prefer descriptive names: `invalid.yaml`, `partial.yaml`
   - Use `config.yaml` for the primary fixture in component directories
   - Use subdirectories like `expected/` for output validation files

3. **Document Your Fixtures**:
   - Include comments in the fixture file explaining its purpose
   - Update this README with details about the new fixture
   - Add table entries for new files

4. **Maintain Consistency**:
   - Follow the same structure as similar existing fixtures
   - Use YAML format unless testing JSON specifically
   - Include meaningful test data that represents real-world scenarios

## Best Practices

### Creating New Component Fixtures

If you're adding tests for a new component (e.g., `scaffold`):

```bash
# 1. Create component directory
mkdir testdata/scaffold

# 2. Add fixtures
cat > testdata/scaffold/config.yaml <<EOF
# Configuration for scaffold command tests
app:
  scaffold:
    template: "basic"
    output_dir: "./output"
EOF

# 3. Add expected outputs if needed
mkdir testdata/scaffold/expected
echo "expected content" > testdata/scaffold/expected/template.go

# 4. Update this README
# Add scaffold/ section to the table above
```

### Security Fixtures

For security-related tests (e.g., testing file permission validation):

- Place in `config/` with descriptive names
- Example: `config/world_writable.yaml`, `config/oversized.yaml`
- Document security implications in fixture comments

## Migration Notes

**November 2025**: Reorganized from flat structure to subdirectories for better scalability.

Previous paths → New paths:
- `config.yaml` → `config/valid.yaml`
- `config.json` → `config/valid.json`
- `partial_config.yaml` → `config/partial.yaml`
- `docs_config.yaml` → `docs/config.yaml`
- `expected_outputs/` → `docs/expected/`
- `logger_test_config.yaml` → `logger/config.yaml`
- `ui_test_config.yaml` → `ui/config.yaml`

All test files have been updated to reference the new paths.
