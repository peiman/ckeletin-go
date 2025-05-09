# ckeletin-go Configuration

This document describes all available configuration options for ckeletin-go.

## Configuration Sources

Configuration can be provided in multiple ways, in order of precedence:

1. Command-line flags
2. Environment variables (with prefix `CKELETIN_GO_`)
3. Configuration file (/Users/peiman/.ckeletin-go.yaml)
4. Default values

## Configuration Options

| Key | Type | Default | Environment Variable | Description |
|-----|------|---------|---------------------|-------------|
| `app.log_level` | string | `info` | `CKELETIN_GO_APP_LOG_LEVEL` | Logging level for the application (trace, debug, info, warn, error, fatal, panic) |
| `app.ping.output_message` | string | `Pong` | `CKELETIN_GO_APP_PING_OUTPUT_MESSAGE` | Default message to display for the ping command |
| `app.ping.output_color` | string | `white` | `CKELETIN_GO_APP_PING_OUTPUT_COLOR` | Text color for ping command output (white, red, green, blue, cyan, yellow, magenta) |
| `app.ping.ui` | bool | `false` | `CKELETIN_GO_APP_PING_UI` | Enable interactive UI for the ping command |

## Example Configuration

### YAML Configuration File (.ckeletin-go.yaml)

```yaml
app:
  # Logging level for the application (trace, debug, info, warn, error, fatal, panic)
  log_level: debug

  # Default message to display for the ping command
  ping.output_message: Hello World!

  # Text color for ping command output (white, red, green, blue, cyan, yellow, magenta)
  ping.output_color: green

  # Enable interactive UI for the ping command
  ping.ui: true

```

### Environment Variables

```bash
# Logging level for the application (trace, debug, info, warn, error, fatal, panic)
export CKELETIN_GO_APP_LOG_LEVEL=debug

# Default message to display for the ping command
export CKELETIN_GO_APP_PING_OUTPUT_MESSAGE=Hello World!

# Text color for ping command output (white, red, green, blue, cyan, yellow, magenta)
export CKELETIN_GO_APP_PING_OUTPUT_COLOR=green

# Enable interactive UI for the ping command
export CKELETIN_GO_APP_PING_UI=true

```
