# Configuration Examples

This directory contains example configuration files for different use cases.

## Available Examples

### 1. Basic Configuration (`basic-config.yaml`)

**Use this if:** You're just getting started or need a simple configuration.

**Features:**
- Minimal working configuration
- Most commonly used settings
- Sensible defaults for local development

**Usage:**
```bash
cp docs/examples/basic-config.yaml config.yaml
./ckeletin-go --config config.yaml
```

### 2. Advanced Configuration (`advanced-config.yaml`)

**Use this if:** You want to understand all available options.

**Features:**
- All configuration options documented
- Explanations for each setting
- Examples of different values
- Environment variable documentation
- Configuration precedence explanation

**Usage:**
```bash
cp docs/examples/advanced-config.yaml config.yaml
# Edit config.yaml and uncomment/modify settings as needed
./ckeletin-go --config config.yaml
```

### 3. Production Configuration (`production-config.yaml`)

**Use this if:** You're deploying to production.

**Features:**
- Production best practices
- File logging with rotation
- Appropriate log levels
- Security considerations
- Deployment checklists for:
  - Standalone servers
  - Docker containers
  - systemd services
  - Kubernetes pods

**Usage:**
```bash
cp docs/examples/production-config.yaml /etc/ckeletin-go/config.yaml
# Or for containerized deployments:
cp docs/examples/production-config.yaml config.yaml
docker run -v $(pwd)/config.yaml:/app/config.yaml ...
```

## Configuration Locations

The application searches for configuration files in this order:

1. Path specified by `--config` flag (highest priority)
2. `./.ckeletin-go.yaml` (current directory)
3. `~/.ckeletin-go.yaml` (user home directory)

If no configuration file is found, the application uses default values from the registry.

## Configuration Precedence

Values are resolved in this order (highest to lowest priority):

1. **Command-line flags** - `--log-level debug`
2. **Environment variables** - `CKELETIN_GO_APP_LOG_LEVEL=debug`
3. **Configuration file** - `app.log_level: debug`
4. **Default values** - From internal registry

## Environment Variables

All configuration options can be set via environment variables:

**Format:** `CKELETIN_GO_<KEY>`
- Replace dots with underscores
- Convert to uppercase

**Examples:**
```bash
export CKELETIN_GO_APP_LOG_LEVEL=debug
export CKELETIN_GO_APP_LOG_FILE_ENABLED=true
export CKELETIN_GO_APP_PING_OUTPUT_MESSAGE="Hello"
```

## Validation

Generate documentation for all available options:

```bash
# Markdown format
./ckeletin-go docs config

# YAML template
./ckeletin-go docs config --format yaml

# Save to file
./ckeletin-go docs config --output docs/configuration.md
```

## Quick Start Examples

### Local Development

```yaml
app:
  log_level: debug
  log:
    file_enabled: false
    color_enabled: true
```

### Docker Container

```yaml
app:
  log:
    file_enabled: false  # Log to stdout for Docker
    console_level: info
    color_enabled: false # Disable colors for log parsers
```

### Production Server

```yaml
app:
  log:
    console_level: info
    file_enabled: true
    file_path: /var/log/ckeletin-go/app.log
    file_level: debug
    file_max_backups: 7
    file_compress: true
```

## Need Help?

- **Full documentation:** See [docs/configuration.md](../configuration.md)
- **Architecture decisions:** See [docs/adr/](../adr/)
- **Development guide:** See [CLAUDE.md](../../CLAUDE.md)
- **Contributing:** See [CONTRIBUTING.md](../../CONTRIBUTING.md)
