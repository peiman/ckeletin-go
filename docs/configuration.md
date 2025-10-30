# ckeletin-go Configuration

This document describes all available configuration options for ckeletin-go.

## Configuration Sources

Configuration can be provided in multiple ways, in order of precedence:

1. Command-line flags
2. Environment variables (with prefix `CKELETIN_GO_`)
3. Configuration file (~/.ckeletin-go.yaml)
4. Default values

## Configuration Options

| Key | Type | Default | Environment Variable | Description |
|-----|------|---------|---------------------|-------------|
| `app.log_level` | string | `info` | `CKELETIN_GO_APP_LOG_LEVEL` | Logging level for the application (trace, debug, info, warn, error, fatal, panic). Used as console level if app.log.console_level is not set. |
| `app.log.console_level` | string | `` | `CKELETIN_GO_APP_LOG_CONSOLE_LEVEL` | Console log level (trace, debug, info, warn, error, fatal, panic). If empty, uses app.log_level. |
| `app.log.file_enabled` | bool | `false` | `CKELETIN_GO_APP_LOG_FILE_ENABLED` | Enable file logging to capture detailed logs |
| `app.log.file_path` | string | `./logs/ckeletin-go.log` | `CKELETIN_GO_APP_LOG_FILE_PATH` | Path to the log file (created with secure 0600 permissions) |
| `app.log.file_level` | string | `debug` | `CKELETIN_GO_APP_LOG_FILE_LEVEL` | File log level (trace, debug, info, warn, error, fatal, panic) |
| `app.log.color_enabled` | string | `auto` | `CKELETIN_GO_APP_LOG_COLOR_ENABLED` | Enable colored console output (auto, true, false). Auto detects TTY. |
| `app.log.file_max_size` | int | `100` | `CKELETIN_GO_APP_LOG_FILE_MAX_SIZE` | Maximum size in megabytes before log file is rotated |
| `app.log.file_max_backups` | int | `3` | `CKELETIN_GO_APP_LOG_FILE_MAX_BACKUPS` | Maximum number of old log files to retain |
| `app.log.file_max_age` | int | `28` | `CKELETIN_GO_APP_LOG_FILE_MAX_AGE` | Maximum number of days to retain old log files |
| `app.log.file_compress` | bool | `false` | `CKELETIN_GO_APP_LOG_FILE_COMPRESS` | Compress rotated log files with gzip |
| `app.log.sampling_enabled` | bool | `false` | `CKELETIN_GO_APP_LOG_SAMPLING_ENABLED` | Enable log sampling for high-volume scenarios |
| `app.log.sampling_initial` | int | `100` | `CKELETIN_GO_APP_LOG_SAMPLING_INITIAL` | Number of messages to log per second before sampling |
| `app.log.sampling_thereafter` | int | `100` | `CKELETIN_GO_APP_LOG_SAMPLING_THEREAFTER` | Number of messages to log thereafter per second |
| `app.docs.output_format` | string | `markdown` | `CKELETIN_GO_APP_DOCS_OUTPUT_FORMAT` | Output format for documentation (markdown, yaml) |
| `app.docs.output_file` | string | `` | `CKELETIN_GO_APP_DOCS_OUTPUT_FILE` | Output file for documentation (defaults to stdout) |
| `app.ping.output_message` | string | `Pong` | `CKELETIN_GO_APP_PING_OUTPUT_MESSAGE` | Default message to display for the ping command |
| `app.ping.output_color` | string | `white` | `CKELETIN_GO_APP_PING_OUTPUT_COLOR` | Text color for ping command output (white, red, green, blue, cyan, yellow, magenta) |
| `app.ping.ui` | bool | `false` | `CKELETIN_GO_APP_PING_UI` | Enable interactive UI for the ping command |

## Example Configuration

### YAML Configuration File (.ckeletin-go.yaml)

```yaml
app:
  # Logging level for the application (trace, debug, info, warn, error, fatal, panic). Used as console level if app.log.console_level is not set.
  log_level: debug

  log:
    # Console log level (trace, debug, info, warn, error, fatal, panic). If empty, uses app.log_level.
    console_level: info

    # Enable file logging to capture detailed logs
    file_enabled: true

    # Path to the log file (created with secure 0600 permissions)
    file_path: /var/log/ckeletin-go/app.log

    # File log level (trace, debug, info, warn, error, fatal, panic)
    file_level: debug

    # Enable colored console output (auto, true, false). Auto detects TTY.
    color_enabled: true

    # Maximum size in megabytes before log file is rotated
    file_max_size: 100

    # Maximum number of old log files to retain
    file_max_backups: 3

    # Maximum number of days to retain old log files
    file_max_age: 28

    # Compress rotated log files with gzip
    file_compress: true

    # Enable log sampling for high-volume scenarios
    sampling_enabled: true

    # Number of messages to log per second before sampling
    sampling_initial: 100

    # Number of messages to log thereafter per second
    sampling_thereafter: 100

  docs:
    # Output format for documentation (markdown, yaml)
    output_format: yaml

    # Output file for documentation (defaults to stdout)
    output_file: /path/to/output.md

  ping:
    # Default message to display for the ping command
    output_message: Hello World!

    # Text color for ping command output (white, red, green, blue, cyan, yellow, magenta)
    output_color: green

    # Enable interactive UI for the ping command
    ui: true

```

### Environment Variables

```bash
# Logging level for the application (trace, debug, info, warn, error, fatal, panic). Used as console level if app.log.console_level is not set.
export CKELETIN_GO_APP_LOG_LEVEL=debug

# Console log level (trace, debug, info, warn, error, fatal, panic). If empty, uses app.log_level.
export CKELETIN_GO_APP_LOG_CONSOLE_LEVEL=info

# Enable file logging to capture detailed logs
export CKELETIN_GO_APP_LOG_FILE_ENABLED=true

# Path to the log file (created with secure 0600 permissions)
export CKELETIN_GO_APP_LOG_FILE_PATH=/var/log/ckeletin-go/app.log

# File log level (trace, debug, info, warn, error, fatal, panic)
export CKELETIN_GO_APP_LOG_FILE_LEVEL=debug

# Enable colored console output (auto, true, false). Auto detects TTY.
export CKELETIN_GO_APP_LOG_COLOR_ENABLED=true

# Maximum size in megabytes before log file is rotated
export CKELETIN_GO_APP_LOG_FILE_MAX_SIZE=100

# Maximum number of old log files to retain
export CKELETIN_GO_APP_LOG_FILE_MAX_BACKUPS=3

# Maximum number of days to retain old log files
export CKELETIN_GO_APP_LOG_FILE_MAX_AGE=28

# Compress rotated log files with gzip
export CKELETIN_GO_APP_LOG_FILE_COMPRESS=true

# Enable log sampling for high-volume scenarios
export CKELETIN_GO_APP_LOG_SAMPLING_ENABLED=true

# Number of messages to log per second before sampling
export CKELETIN_GO_APP_LOG_SAMPLING_INITIAL=100

# Number of messages to log thereafter per second
export CKELETIN_GO_APP_LOG_SAMPLING_THEREAFTER=100

# Output format for documentation (markdown, yaml)
export CKELETIN_GO_APP_DOCS_OUTPUT_FORMAT=yaml

# Output file for documentation (defaults to stdout)
export CKELETIN_GO_APP_DOCS_OUTPUT_FILE=/path/to/output.md

# Default message to display for the ping command
export CKELETIN_GO_APP_PING_OUTPUT_MESSAGE=Hello World!

# Text color for ping command output (white, red, green, blue, cyan, yellow, magenta)
export CKELETIN_GO_APP_PING_OUTPUT_COLOR=green

# Enable interactive UI for the ping command
export CKELETIN_GO_APP_PING_UI=true

```
