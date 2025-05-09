# Test Fixtures

This directory contains test fixtures used for testing various components of the ckeletin-go CLI application.

## Config Files

| Filename | Purpose |
|----------|---------|
| `config.yaml` | Basic configuration file used for general testing |
| `config.json` | JSON configuration for testing format compatibility |
| `invalid_config.yaml` | Intentionally invalid YAML for testing error handling |
| `empty_config.yaml` | Empty config file for testing default values |
| `partial_config.yaml` | Partial config for testing default value merging |
| `docs_config.yaml` | Configuration for testing the docs command |
| `env_override.yaml` | Base config for testing environment variable overrides |
| `ui_test_config.yaml` | Configuration for testing UI components |
| `logger_test_config.yaml` | Configuration for testing logging systems |

## Usage in Tests

These fixtures are meant to be used in tests to:

1. Validate configuration loading and parsing
2. Test error handling with invalid configs
3. Verify default values are applied correctly
4. Test environment variable overrides
5. Test UI component configurations
6. Test logger configurations

## Adding New Test Fixtures

When adding new test fixtures, follow these guidelines:

1. Use clear, descriptive filenames
2. Include comments in the fixture explaining its purpose
3. Update this README with details about the new fixture
4. Make sure fixtures follow the same structure as similar existing fixtures

## Mock Data

For more complex tests that require mock data beyond configuration, consider:

1. Adding subdirectories for specific test cases
2. Using consistent naming patterns
3. Documenting the structure and purpose of mock data 