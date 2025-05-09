# Test Fixtures

**ckeletin-go** includes a comprehensive set of test fixtures to facilitate testing different aspects of the application.

## Available Test Fixtures

The `testdata/` directory contains:

- **Configuration Files**: Sample configs in various formats (YAML, JSON) and states (valid, invalid, empty, partial)
- **Expected Outputs**: Expected command outputs for comparison in tests
- **Documentation**: A README explaining the purpose and structure of test fixtures

### Configuration Files

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

### Expected Outputs

The `expected_outputs/` directory contains the expected output of commands for testing:

| Filename | Purpose |
|----------|---------|
| `docs_markdown.md` | Expected markdown output of the docs command |
| `docs_yaml.yaml` | Expected YAML output of the docs command |

## Using Test Fixtures in Tests

Test fixtures can be used in your tests to:

1. **Test Configuration Loading**: Use different config formats and scenarios

   ```go
   // Load a test config file
   viper.SetConfigFile("../../testdata/config.yaml")
   ```

2. **Test Error Handling**: Use invalid configs to verify error handling

   ```go
   // Load an invalid config file
   viper.SetConfigFile("../../testdata/invalid_config.yaml")
   ```

3. **Verify Command Output**: Compare command output with expected outputs

   ```go
   // Compare output with expected
   expected, err := os.ReadFile("../../testdata/expected_outputs/docs_markdown.md")
   ```

4. **Test UI Components**: Use UI-specific configs

## Adding New Test Fixtures

When adding new functionality that needs testing:

1. Add appropriate test fixtures to the `testdata/` directory
2. Document the fixtures in `testdata/README.md`
3. Reference the fixtures in your tests

## Best Practices for Test Fixtures

1. **Keep Test Fixtures in Sync with Code**: Update test fixtures when the corresponding code changes
2. **Document Purpose**: Include comments explaining the purpose of each test fixture
3. **Use Realistic Data**: Make test fixtures representative of real-world usage
4. **Version Control**: Always commit test fixtures alongside code changes
5. **Isolation**: Each test fixture should be designed for a specific test case
