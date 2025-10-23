# cmd/ - Command Definitions

This directory contains command files following the **ultra-thin command pattern**.

## Overview

Command files in this directory are intentionally minimal (~20-30 lines) and serve as thin CLI wrappers. All business logic, configuration, and metadata are separated into dedicated packages.

## Directory Structure

```
cmd/
├── README.md              # This file
├── root.go               # Root command (framework file)
├── flags.go              # Flag registration helpers (framework file)
├── helpers.go            # Command creation helpers (framework file)
├── ping.go               # Example ultra-thin command
└── docs.go               # Example ultra-thin command with subcommands
```

## Framework vs User Files

### Framework Files (DO NOT EDIT unless modifying the framework)
- `root.go` - Root command setup
- `flags.go` - Flag registration logic
- `helpers.go` - NewCommand() and MustAddToRoot() helpers

### User Files (Edit these when adding commands)
- `<command>.go` - Individual command files

## Creating a New Command

### Option 1: Using the Generator (Recommended)

```bash
task generate:command name=mycommand
```

This creates:
- `cmd/mycommand.go` - Ultra-thin command file
- `internal/config/commands/mycommand_config.go` - Metadata + options

### Option 2: Manual Creation

1. **Create command config** in `internal/config/commands/mycommand_config.go`:

```go
package commands

import "github.com/peiman/ckeletin-go/internal/config"

// MycommandMetadata defines all metadata for the mycommand command
var MycommandMetadata = config.CommandMetadata{
    Use:   "mycommand",
    Short: "Short description",
    Long:  `Long description...`,
    ConfigPrefix: "app.mycommand",
    FlagOverrides: map[string]string{
        "app.mycommand.some_option": "flag-name",
    },
}

func MycommandOptions() []config.ConfigOption {
    return []config.ConfigOption{
        {
            Key:          "app.mycommand.some_option",
            DefaultValue: "default",
            Description:  "Option description",
            Type:         "string",
        },
    }
}

func init() {
    config.RegisterOptionsProvider(MycommandOptions)
}
```

2. **Create command file** in `cmd/mycommand.go`:

```go
package cmd

import (
    "github.com/peiman/ckeletin-go/internal/config/commands"
    "github.com/peiman/ckeletin-go/internal/mycommand"
    "github.com/spf13/cobra"
)

var mycommandCmd = NewCommand(commands.MycommandMetadata, runMycommand)

func init() {
    MustAddToRoot(mycommandCmd)
}

func runMycommand(cmd *cobra.Command, args []string) error {
    cfg := mycommand.Config{
        SomeOption: getConfigValueWithFlags[string](cmd, "flag-name", "app.mycommand.some_option"),
    }
    return mycommand.NewExecutor(cfg, cmd.OutOrStdout()).Execute()
}
```

3. **Create business logic** in `internal/mycommand/mycommand.go`:

```go
package mycommand

import "io"

// Config holds configuration for mycommand business logic
type Config struct {
    SomeOption string
}

type Executor struct {
    cfg    Config
    writer io.Writer
}

func NewExecutor(cfg Config, writer io.Writer) *Executor {
    return &Executor{cfg: cfg, writer: writer}
}

func (e *Executor) Execute() error {
    // Business logic here
    return nil
}
```

## Ultra-Thin Pattern Rules

### ✅ DO

- Use `NewCommand()` to create commands from metadata
- Use `MustAddToRoot()` to register commands
- Keep command files ~20-30 lines
- Move all business logic to `internal/<command>/`
- Define metadata in `internal/config/commands/<command>_config.go`
- Use dependency injection (pass io.Writer, etc.)
- Add tests for business logic in `internal/<command>/<command>_test.go`

### ❌ DON'T

- Hardcode command metadata (Use, Short, Long) in cmd files
- Put business logic in cmd files
- Manually call `RegisterFlagsForPrefixWithOverrides()` (NewCommand does this)
- Manually call `RootCmd.AddCommand()` (MustAddToRoot does this)
- Set defaults with `viper.SetDefault()` (use config registry)

## Validation

Run the validation script to ensure commands follow the pattern:

```bash
task validate-commands
```

### Whitelisting Commands

If you need to deviate from the pattern (e.g., complex command hierarchy), add this comment to the command file:

```go
// ckeletin:allow-custom-command
```

## Examples

- **Simple command**: `cmd/ping.go` (~30 lines)
- **Command with subcommands**: `cmd/docs.go` (~48 lines)

## Architecture Benefits

- **Separation of Concerns**: CLI wiring separate from business logic
- **Testability**: Business logic easily testable without Cobra
- **Consistency**: All commands follow same pattern
- **Maintainability**: Metadata and options centralized
- **Discoverability**: Single source of truth for each command

## Related Files

- `internal/config/command_metadata.go` - CommandMetadata struct definition
- `internal/config/command_options.go` - ConfigOption struct definition
- `internal/config/commands/` - All command configs (metadata + options)
- `cmd/helpers.go` - Framework helpers for creating commands
- `scripts/validate-command-patterns.sh` - Validation script
