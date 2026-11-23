package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const outputFlagName = "output"

// Phase selects which sources UseOutputMode may read.
type Phase int

const (
	// PhaseEarly is flag, alias, then environment. Config is not loaded yet.
	PhaseEarly Phase = iota
	// PhaseResolved adds the config file and the registry default. Call it
	// after viper has read the file and before logger.Init.
	PhaseResolved
)

type flagSettings struct {
	alias     string
	envPrefix string
}

// settings is process-wide. A CLI resolves one output mode.
var settings flagSettings

// Option configures RegisterOutputFlag.
type Option func(*flagSettings)

// WithAliasFlag treats a command's existing flag of that name as --output.
// The framework does not register the alias. An explicit --output wins.
// A changed alias wins over the environment and the config file.
func WithAliasFlag(name string) Option {
	return func(s *flagSettings) {
		s.alias = name
	}
}

// WithEnvPrefix is the viper env prefix (the binary name, sanitized).
// The variable is PREFIX_APP_OUTPUT_FORMAT, the same name viper reads.
func WithEnvPrefix(prefix string) Option {
	return func(s *flagSettings) {
		s.envPrefix = prefix
	}
}

// RegisterOutputFlag registers the persistent --output text|json flag and
// resolves the mode at the start of PersistentPreRunE, before the command's
// own pre-run. Call it once from the application init, before flag defaults
// are copied onto the flag.
func RegisterOutputFlag(root *cobra.Command, opts ...Option) {
	if root == nil {
		return
	}
	for _, opt := range opts {
		opt(&settings)
	}
	if root.PersistentFlags().Lookup(outputFlagName) == nil {
		root.PersistentFlags().String(outputFlagName, "text",
			"Output format: text (human-readable) or json (machine-readable)")
	}
	if root.Annotations == nil {
		root.Annotations = map[string]string{}
	}
	if root.Annotations["ckeletin.output.wrapped"] == "1" {
		return
	}
	root.Annotations["ckeletin.output.wrapped"] = "1"

	prev := root.PersistentPreRunE
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		SetCommandFrom(cmd)
		if err := UseOutputMode(PhaseEarly); err != nil {
			return err
		}
		if prev != nil {
			return prev(cmd, args)
		}
		return nil
	}
}

// UseOutputMode resolves the mode and stores it for IsJSONMode and logger.Init.
// Invalid values return an error whose exit code is ExitInvalidOutput.
func UseOutputMode(phase Phase) error {
	mode, err := resolveMode(phase)
	if err != nil {
		return err
	}
	SetOutputMode(mode)
	return nil
}

func resolveMode(phase Phase) (string, error) {
	if value, ok := explicitFlag(currentCmd, outputFlagName); ok {
		return validMode(value)
	}
	if settings.alias != "" {
		if value, ok := explicitFlag(currentCmd, settings.alias); ok {
			return validMode(value)
		}
	}
	if value, ok := envMode(); ok {
		return validMode(value)
	}
	if phase == PhaseResolved && viper.IsSet(config.KeyAppOutputFormat) {
		return validMode(viper.GetString(config.KeyAppOutputFormat))
	}
	return "text", nil
}

func explicitFlag(cmd *cobra.Command, name string) (string, bool) {
	if cmd == nil || name == "" {
		return "", false
	}
	flag := cmd.Flags().Lookup(name)
	if flag == nil || !flag.Changed {
		return "", false
	}
	return flag.Value.String(), true
}

func envMode() (string, bool) {
	if settings.envPrefix == "" {
		return "", false
	}
	key := settings.envPrefix + "_" + strings.ToUpper(strings.ReplaceAll(config.KeyAppOutputFormat, ".", "_"))
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", false
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	return value, true
}

func validMode(value string) (string, error) {
	switch value {
	case "text", "json":
		return value, nil
	default:
		return "", &exitError{
			code: ExitInvalidOutput,
			msg:  fmt.Sprintf("invalid output format %q: must be \"text\" or \"json\"", value),
		}
	}
}

type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string { return e.msg }

func (e *exitError) ExitCode() int { return e.code }
