# Configuration Reference

This document describes all configuration options for the application.

## Global Configuration

| Option | Environment Variable | Default | Description |
|--------|---------------------|---------|-------------|
| `app.log_level` | `APP_LOG_LEVEL` | `info` | The log level (trace, debug, info, warn, error, fatal) |
| `app.timeout` | `APP_TIMEOUT` | `30s` | Global timeout for operations |

## Ping Command

| Option | Environment Variable | Default | Description |
|--------|---------------------|---------|-------------|
| `app.ping.output_message` | `APP_PING_OUTPUT_MESSAGE` | `Pong!` | Message to display |
| `app.ping.output_color` | `APP_PING_OUTPUT_COLOR` | `green` | Color to use for output |
| `app.ping.ui` | `APP_PING_UI` | `false` | Whether to use interactive UI |

## Docs Command

| Option | Environment Variable | Default | Description |
|--------|---------------------|---------|-------------|
| `app.docs.format` | `APP_DOCS_FORMAT` | `markdown` | Output format (markdown, yaml) |
| `app.docs.output_file` | `APP_DOCS_OUTPUT_FILE` | `` | File to write output to (stdout if empty) 