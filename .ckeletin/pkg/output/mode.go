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
	alias       string
	envPrefix   string
	afterLogger func(error)
}

// settings is process-wide. A CLI resolves one output mode.
var settings flagSettings

// phaseUnset means UseOutputMode has not run for this Execute.
// PhaseEarly is 0, so this stays below both phases.
const phaseUnset Phase = -1

// resolvedPhase is the last phase UseOutputMode finished.
// Execute uses it to tell a finished resolution from a caller that only
// preset the mode.
var resolvedPhase = phaseUnset

// modePinned is true after SetOutputMode. Execute does not replace a
// preset mode when a command fails before PersistentPreRunE.
var modePinned bool

// modeWroteKey is true after UseOutputMode stores the resolved mode in
// viper. The next resolution ignores that write unless the config file
// or a later caller set the key to a different value.
var modeWroteKey bool
var modeWroteValue string

type modeSource int

const (
	sourceFlag modeSource = iota
	sourceAlias
	sourceEnv
	sourceConfig
	sourceDefault
)

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

// UseOutputMode resolves the mode and stores it for IsJSONMode.
// It also writes the resolved value to app.output_format. logger.Init reads
// that key and does not import this package, so an alias flag and a text
// override of config json reach the console writer. The default is not
// written during PhaseEarly: the config file is not loaded yet, and a viper
// write would hide it. Invalid values return an error whose exit code is
// ExitInvalidOutput.
func UseOutputMode(phase Phase) error {
	mode, source, err := resolveMode(phase)
	if err != nil {
		return err
	}
	currentOutputMode = mode
	modePinned = false
	if phase > resolvedPhase {
		resolvedPhase = phase
	}
	if source != sourceDefault || phase == PhaseResolved {
		viper.Set(config.KeyAppOutputFormat, mode)
		modeWroteKey = true
		modeWroteValue = mode
	}
	return nil
}

func resolveMode(phase Phase) (string, modeSource, error) {
	if value, ok := explicitFlag(currentCmd, outputFlagName); ok {
		mode, err := validMode(value)
		return mode, sourceFlag, err
	}
	if settings.alias != "" {
		if value, ok := explicitFlag(currentCmd, settings.alias); ok {
			mode, err := validMode(value)
			return mode, sourceAlias, err
		}
	}
	if value, ok := envMode(); ok {
		mode, err := validMode(value)
		return mode, sourceEnv, err
	}
	// A value this process wrote for the logger is not the config file.
	// InConfig is the file. IsSet covers a caller that set the key itself,
	// which is how tests load configuration without a file.
	if phase == PhaseResolved && configMode() {
		mode, err := validMode(viper.GetString(config.KeyAppOutputFormat))
		return mode, sourceConfig, err
	}
	return "text", sourceDefault, nil
}

func configMode() bool {
	key := config.KeyAppOutputFormat
	if !viper.IsSet(key) {
		modeWroteKey = false
		return false
	}
	if viper.InConfig(key) {
		return true
	}
	if !modeWroteKey {
		return true
	}
	return viper.GetString(key) != modeWroteValue
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
