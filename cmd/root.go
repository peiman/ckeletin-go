// cmd/root.go
//
// Thread-Safety Notes:
//
// Viper configuration in this application follows a safe initialization pattern:
//  1. All configuration is initialized during startup in PersistentPreRunE (single-threaded)
//  2. Configuration is read-only after initialization completes
//  3. No concurrent writes occur during command execution
//  4. Commands execute sequentially (Cobra's execution model)
//
// This pattern ensures thread-safety without requiring locks or synchronization.
// Viper itself is not thread-safe for writes, but our usage pattern avoids concurrent access.

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/peiman/ckeletin-go/internal/xdg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	configPathMode = ConfigPathModeXDG
	configPathFlag *pflag.Flag
	// Build identity vars injected via ldflags (see Taskfile.yml LDFLAGS and
	// .goreleaser.yml). Without injection they degrade to "unknown" per
	// CKSPEC-OUT-006 — never empty strings.
	Version = versionUnknown
	Commit  = versionUnknown
	Date    = versionUnknown
	// Dirty is "true"/"false" when injected (GoReleaser's {{ .IsGitDirty }});
	// when empty, treeState() derives the state from Version, where Taskfile
	// builds embed a "-dirty" suffix via `git describe --dirty`.
	Dirty            = ""
	binaryName       = "" // MUST be injected via ldflags (see Taskfile.yml LDFLAGS)
	configFileStatus string
	configFileUsed   string

	// Compiled regex patterns for EnvPrefix()
	// Compiled once at package initialization for better performance
	nonAlphanumericRegex = regexp.MustCompile(`[^A-Z0-9]`)
	onlyUnderscoresRegex = regexp.MustCompile(`^_+$`)
)

const (
	// ConfigPathModeXDG searches XDG-style config directory (default).
	// On macOS, this means ~/.config/<app> unless XDG_CONFIG_HOME is set.
	ConfigPathModeXDG = "xdg"
	// ConfigPathModeNative searches the OS-native config directory.
	// On macOS, this means ~/Library/Application Support/<app>.
	ConfigPathModeNative = "native"
	// ConfigPathModeBoth searches both XDG and native directories.
	ConfigPathModeBoth = "both"
)

const (
	// versionUnknown is the graceful-degradation value for build-identity
	// fields when ldflags are not injected (plain `go build`, CKSPEC-OUT-006).
	versionUnknown = "unknown"
	// versionDevFallback is the Taskfile's VERSION fallback when git describe
	// fails at build time; the working-tree state is unknowable then too.
	versionDevFallback = "dev"
	// dirtySuffix is appended to VERSION by `git describe --dirty` (Taskfile
	// LDFLAGS) when the build came from a modified working tree.
	dirtySuffix = "-dirty"

	// Working-tree states surfaced in version output (CKSPEC-OUT-006).
	treeStateDirty   = "dirty"
	treeStateClean   = "clean"
	treeStateUnknown = "unknown"
)

// EnvPrefix returns a sanitized environment variable prefix based on the binary name
func EnvPrefix() string {
	// Convert to uppercase and replace non-alphanumeric characters with underscore
	prefix := strings.ToUpper(binaryName)
	prefix = nonAlphanumericRegex.ReplaceAllString(prefix, "_")

	// Ensure it doesn't start with a number (invalid for env vars)
	if prefix != "" && prefix[0] >= '0' && prefix[0] <= '9' {
		prefix = "_" + prefix
	}

	// Handle case where all characters were special and got replaced
	if onlyUnderscoresRegex.MatchString(prefix) {
		prefix = "_"
	}

	return prefix
}

// ConfigPaths returns configuration paths for the application.
//
// Config file search order (handled by viper):
//  1. --config flag (explicit override)
//  2. ./config.{yaml,yml,json,toml} (project-local config)
//  3. User config directory based on path mode:
//     - xdg (default): $XDG_CONFIG_HOME/<binaryName> or ~/.config/<binaryName>
//     - native: OS-native config path (macOS: ~/Library/Application Support/<binaryName>)
//     - both: xdg first, then native
//
// Viper automatically detects the file format based on extension.
type ConfigPathInfo struct {
	// ConfigName is the base config name without extension (e.g. "config")
	// Viper will search for config.yaml, config.yml, config.json, config.toml
	ConfigName string
	// XDGDir is the XDG-style config directory (e.g. "$XDG_CONFIG_HOME/myapp" or "~/.config/myapp")
	XDGDir string
	// NativeDir is the OS-native config directory (e.g. macOS "~/Library/Application Support/myapp").
	NativeDir string
	// Mode controls which user config directory is searched: xdg, native, or both.
	Mode string
	// SearchPaths lists all viper search paths in priority order.
	SearchPaths []string
}

func ConfigPaths() ConfigPathInfo {
	xdgDir := resolveXDGConfigDir()
	nativeDir, _ := xdg.ConfigDir()

	mode := resolveConfigPathMode()
	paths := ConfigPathInfo{
		ConfigName: "config",
		XDGDir:     xdgDir,
		NativeDir:  nativeDir,
		Mode:       mode,
	}
	paths.SearchPaths = buildConfigSearchPaths(paths)
	return paths
}

func resolveXDGConfigDir() string {
	name := binaryName
	if name == "" {
		name = xdg.GetAppName()
	}

	xdgConfigHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
	if xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, name)
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		fallback, _ := xdg.ConfigDir()
		return fallback
	}

	return filepath.Join(home, ".config", name)
}

func resolveConfigPathMode() string {
	mode := strings.ToLower(strings.TrimSpace(configPathMode))
	if mode == "" {
		mode = ConfigPathModeXDG
	}

	flagChanged := configPathFlag != nil && configPathFlag.Changed
	if !flagChanged {
		envKey := EnvPrefix() + "_CONFIG_PATH_MODE"
		if envMode := strings.ToLower(strings.TrimSpace(os.Getenv(envKey))); envMode != "" {
			mode = envMode
		}
	}

	switch mode {
	case ConfigPathModeXDG, ConfigPathModeNative, ConfigPathModeBoth:
		return mode
	default:
		log.Warn().
			Str("config_path_mode", mode).
			Str("fallback_mode", ConfigPathModeXDG).
			Msg("Invalid config path mode, falling back to default")
		return ConfigPathModeXDG
	}
}

func buildConfigSearchPaths(paths ConfigPathInfo) []string {
	searchPaths := []string{"."}

	addUnique := func(path string) {
		if path == "" {
			return
		}
		for _, existing := range searchPaths {
			if existing == path {
				return
			}
		}
		searchPaths = append(searchPaths, path)
	}

	switch paths.Mode {
	case ConfigPathModeXDG:
		addUnique(paths.XDGDir)
	case ConfigPathModeNative:
		addUnique(paths.NativeDir)
	case ConfigPathModeBoth:
		addUnique(paths.XDGDir)
		addUnique(paths.NativeDir)
	default: // xdg
		addUnique(paths.XDGDir)
	}

	return searchPaths
}

func defaultUserConfigDir(paths ConfigPathInfo) string {
	for _, path := range paths.SearchPaths {
		if path != "." {
			return path
		}
	}
	return ""
}

// Export RootCmd so that tests in other packages can manipulate it without getters/setters.
var RootCmd = &cobra.Command{
	Use:           "",
	Short:         "A production-ready Go CLI application",
	Long:          "",
	SilenceErrors: true, // Errors are handled by main.go, don't print twice
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Bind flags to viper first (must happen before initConfig)
		if err := bindFlags(cmd); err != nil {
			return fmt.Errorf("failed to bind flags: %w", err)
		}

		// Activate JSON output mode as early as it is knowable (flag, then
		// env var) so errors raised DURING config/logger init are routed
		// through the JSON error handler in main.go instead of plain text.
		output.SetOutputMode(earlyOutputMode(cmd))
		output.SetCommandName(cmd.Name())

		// Config-file-driven JSON mode is unknowable until the config is
		// read, so buffer all pre-init log output and resolve it after the
		// final mode is known: discarded in JSON mode, replayed to stderr in
		// text mode.
		preInit := bufferPreInitLogs()

		// Initialize configuration
		if err := initConfig(); err != nil {
			preInit.flush(false)
			return err
		}

		// Apply the final output mode (flag > env > config file) BEFORE the
		// logger is initialized: in JSON mode logger.Init silences ONLY the
		// console writer, so stdout carries exactly one JSON envelope and
		// stderr stays silent, while the audit log file (if enabled) keeps
		// receiving entries (CKSPEC-OUT-004). This second site covers
		// config-file-driven JSON mode, which earlyOutputMode cannot see.
		output.SetOutputMode(viper.GetString(config.KeyAppOutputFormat))

		// Initialize logger with configuration values
		if err := logger.Init(nil); err != nil {
			preInit.flush(false)
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		preInit.flush(true)

		logConfigStatus()
		return nil
	},
}

// earlyOutputMode resolves the output mode from the sources knowable BEFORE
// initConfig runs: the --output flag (highest precedence), then the env var
// mirror of config.KeyAppOutputFormat. A config file can still switch the
// mode afterwards — the post-initConfig SetOutputMode site covers that.
// Resolving early matters for errors raised during init itself: main.go must
// already know it is in JSON mode to emit the error envelope, because the
// post-initConfig site is never reached on those error paths.
func earlyOutputMode(cmd *cobra.Command) string {
	if f := cmd.Root().PersistentFlags().Lookup("output"); f != nil && f.Changed {
		return f.Value.String()
	}
	return strings.TrimSpace(os.Getenv(outputFormatEnvVar()))
}

// outputFormatEnvVar mirrors config.KeyAppOutputFormat through viper's env
// mapping (SetEnvPrefix plus the "." -> "_" key replacer in initConfig): the
// same variable viper itself reads, knowable before initConfig runs.
func outputFormatEnvVar() string {
	return EnvPrefix() + "_" + strings.ToUpper(strings.ReplaceAll(config.KeyAppOutputFormat, ".", "_"))
}

// preInitLogBuffer captures log output emitted before logger.Init installs
// the real writers. Until then zerolog's default logger writes raw JSON to
// stderr, but whether stderr may carry those lines depends on the FINAL
// output mode, which env- or config-file-driven JSON can change after the
// logs were emitted (CKSPEC-OUT-004 keeps stderr silent in JSON mode).
// Buffering defers the decision until flush.
type preInitLogBuffer struct {
	buf  bytes.Buffer
	prev zerolog.Logger
}

// bufferPreInitLogs swaps the global logger for one writing into a buffer and
// returns the buffer for a later flush. Re-entrant: each call snapshots the
// logger it replaces, so repeated Execute() runs (tests) stay independent.
func bufferPreInitLogs() *preInitLogBuffer {
	b := &preInitLogBuffer{prev: log.Logger}
	log.Logger = zerolog.New(&b.buf).With().Timestamp().Logger()
	return b
}

// flush resolves the buffered pre-init logs once the output mode is final:
// JSON mode discards them (stderr must stay silent; no audit file captured
// them, so nothing the file could have kept is lost), text mode replays the
// raw bytes to stderr — exactly the raw-JSON lines text mode printed before
// buffering existed. realLoggerInstalled reports whether logger.Init
// succeeded; when it did not, the previously installed logger is restored so
// later log calls cannot write into an already-flushed buffer.
func (b *preInitLogBuffer) flush(realLoggerInstalled bool) {
	if !realLoggerInstalled {
		log.Logger = b.prev
	}
	if output.IsJSONMode() || b.buf.Len() == 0 {
		return
	}
	_, _ = os.Stderr.Write(b.buf.Bytes())
}

// logConfigStatus reports how configuration was resolved, once the real
// logger is installed; in JSON mode this reaches only the audit log file
// (the console writer is disabled).
func logConfigStatus() {
	if configFileStatus == "" {
		return
	}
	if configFileUsed != "" {
		log.Info().Str("config_file", logger.SanitizePath(configFileUsed)).Msg(configFileStatus)
		return
	}
	log.Debug().Msg(configFileStatus)
}

func Execute() error {
	// Ensure logger cleanup on exit
	defer logger.Cleanup()

	RootCmd.Version = versionString()
	// Register the --version flag NOW (cobra normally defers this until after
	// command lookup): without it, `--version --output json` makes stripFlags
	// treat the unknown --version as value-taking, swallow --output, and fail
	// with `unknown command "json"`.
	RootCmd.InitDefaultVersionFlag()
	return RootCmd.Execute()
}

// orUnknown backstops build-identity fields at runtime: a pipeline that
// injects EMPTY strings via ldflags overrides the package-var "unknown"
// defaults, and CKSPEC-OUT-006 forbids empty fields in version output.
func orUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return versionUnknown
	}
	return value
}

// versionString composes the build identity shown by --version: semantic
// version, commit, build date, and working-tree state (CKSPEC-OUT-006).
func versionString() string {
	return fmt.Sprintf("%s, commit %s, built at %s, tree %s",
		orUnknown(Version), orUnknown(Commit), orUnknown(Date), treeState())
}

// treeState resolves whether the build came from a dirty working tree
// (CKSPEC-OUT-006). Precedence:
//  1. Explicit Dirty ldflag — GoReleaser injects {{ .IsGitDirty }} ("true"/"false").
//  2. The "-dirty" suffix `git describe --dirty` embeds in Version (Taskfile builds).
//  3. Unknown — no build identity was injected (plain `go build`), or the
//     Taskfile fell back to "dev" because git describe failed.
func treeState() string {
	switch strings.ToLower(strings.TrimSpace(Dirty)) {
	case "true":
		return treeStateDirty
	case "false":
		return treeStateClean
	}
	version := orUnknown(Version)
	if strings.HasSuffix(version, dirtySuffix) {
		return treeStateDirty
	}
	if version == versionUnknown || version == versionDevFallback {
		return treeStateUnknown
	}
	return treeStateClean
}

// versionData is the machine-readable --version payload for --output json.
type versionData struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	Tree    string `json:"tree"`
}

// renderVersion renders the --version output. Cobra handles the version flag
// BEFORE PersistentPreRunE runs, so JSON mode is detected from the parsed
// --output flag directly rather than via output.IsJSONMode() (not yet set).
func renderVersion(cmd *cobra.Command) string {
	if outputFlag := cmd.Root().PersistentFlags().Lookup("output"); outputFlag != nil && outputFlag.Value.String() == "json" {
		envelope := output.JSONEnvelope{
			Status:  "success",
			Command: cmd.Name(),
			Data: versionData{
				Version: orUnknown(Version),
				Commit:  orUnknown(Commit),
				Date:    orUnknown(Date),
				Tree:    treeState(),
			},
		}
		var buf bytes.Buffer
		if err := output.RenderJSON(&buf, envelope); err == nil {
			return buf.String()
		}
		// Unreachable in practice (versionData always marshals); fall through
		// to text so --version never produces empty output.
	}
	return fmt.Sprintf("%s version %s\n", cmd.DisplayName(), cmd.Version)
}

func init() {
	// Fallback for development/testing when ldflags aren't injected
	// Production builds MUST inject binaryName via ldflags (see Taskfile.yml LDFLAGS)
	if binaryName == "" {
		binaryName = "ckeletin-go"
	}

	// Initialize XDG paths with app name (single source of truth)
	xdg.SetAppName(binaryName)

	// Update RootCmd with the resolved binaryName.
	// Package-level var declarations capture binaryName="" before init() runs,
	// so we need to set these after the fallback is applied.
	RootCmd.Use = binaryName
	RootCmd.Long = fmt.Sprintf(`%s is a production-ready Go CLI application built with ckeletin-go.
Powered by Cobra, Viper, Zerolog, and Bubble Tea with enforced architecture patterns.`, binaryName)

	// Route --version through renderVersion so it honors --output json.
	// Cobra resolves the version flag before PersistentPreRunE, so the
	// template func is the only hook that runs early enough.
	cobra.AddTemplateFunc("ckeletinRenderVersion", renderVersion)
	RootCmd.SetVersionTemplate(`{{ckeletinRenderVersion .}}`)

	configPaths := ConfigPaths()

	// Define all persistent flags (flag definitions only - bindings happen in bindFlags())
	searchTargets := make([]string, 0, len(configPaths.SearchPaths))
	for _, path := range configPaths.SearchPaths {
		if path == "." {
			searchTargets = append(searchTargets, "./config.yaml")
			continue
		}
		searchTargets = append(searchTargets, filepath.Join(path, "config.yaml"))
	}

	configHelp := "Config file (searches: " + strings.Join(searchTargets, ", ")
	if configHelp == "Config file (searches: " {
		configHelp += "./config.yaml"
	}
	configHelp += ")"
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", configHelp)
	RootCmd.PersistentFlags().StringVar(&configPathMode, "config-path-mode", ConfigPathModeXDG,
		"Config path mode when --config is not set (xdg, native, both)")
	configPathFlag = RootCmd.PersistentFlags().Lookup("config-path-mode")

	// Legacy log level flag (for backward compatibility)
	RootCmd.PersistentFlags().String("log-level", "info", "Set the log level (trace, debug, info, warn, error, fatal, panic)")

	// Dual logging configuration flags
	RootCmd.PersistentFlags().String("log-console-level", "", "Console log level (trace, debug, info, warn, error, fatal, panic). If empty, uses --log-level.")
	RootCmd.PersistentFlags().Bool("log-file-enabled", false, "Enable file logging to capture detailed logs")
	RootCmd.PersistentFlags().String("log-file-path", "./logs/ckeletin-go.log", "Path to the log file")
	RootCmd.PersistentFlags().String("log-file-level", "debug", "File log level (trace, debug, info, warn, error, fatal, panic)")
	RootCmd.PersistentFlags().String("log-color", "auto", "Enable colored console output (auto, true, false)")

	// Log rotation configuration flags
	RootCmd.PersistentFlags().Int("log-file-max-size", 100, "Maximum size in megabytes before log file is rotated")
	RootCmd.PersistentFlags().Int("log-file-max-backups", 3, "Maximum number of old log files to retain")
	RootCmd.PersistentFlags().Int("log-file-max-age", 28, "Maximum number of days to retain old log files")
	RootCmd.PersistentFlags().Bool("log-file-compress", false, "Compress rotated log files with gzip")

	// Log sampling configuration flags
	RootCmd.PersistentFlags().Bool("log-sampling-enabled", false, "Enable log sampling for high-volume scenarios")
	RootCmd.PersistentFlags().Int("log-sampling-initial", 100, "Number of messages to log per second before sampling")

	RootCmd.PersistentFlags().Int("log-sampling-thereafter", 100, "Number of messages to log thereafter per second")

	// Output format flag
	RootCmd.PersistentFlags().String("output", "text", "Output format: text (human-readable) or json (machine-readable)")
}

// bindFlags binds all persistent flags to viper configuration keys.
// This function is called from PersistentPreRunE to allow proper error handling.
// Unlike the previous init() pattern with log.Fatal(), this returns errors that can be
// handled gracefully and makes the code testable.
func bindFlags(cmd *cobra.Command) error {
	var errs []error

	// Helper function to collect binding errors
	// Use cmd.Root() to get flags from RootCmd even when called from subcommands
	bindFlag := func(key string, flagName string) {
		if err := viper.BindPFlag(key, cmd.Root().PersistentFlags().Lookup(flagName)); err != nil {
			errs = append(errs, fmt.Errorf("bind flag %q to key %q: %w", flagName, key, err))
		}
	}

	// Bind all flags to their viper keys
	bindFlag("config", "config")
	bindFlag(config.KeyAppLogLevel, "log-level")
	bindFlag(config.KeyAppLogConsoleLevel, "log-console-level")
	bindFlag(config.KeyAppLogFileEnabled, "log-file-enabled")
	bindFlag(config.KeyAppLogFilePath, "log-file-path")
	bindFlag(config.KeyAppLogFileLevel, "log-file-level")
	bindFlag(config.KeyAppLogColorEnabled, "log-color")
	bindFlag(config.KeyAppLogFileMaxSize, "log-file-max-size")
	bindFlag(config.KeyAppLogFileMaxBackups, "log-file-max-backups")
	bindFlag(config.KeyAppLogFileMaxAge, "log-file-max-age")
	bindFlag(config.KeyAppLogFileCompress, "log-file-compress")
	bindFlag(config.KeyAppLogSamplingEnabled, "log-sampling-enabled")
	bindFlag(config.KeyAppLogSamplingInitial, "log-sampling-initial")
	bindFlag(config.KeyAppLogSamplingThereafter, "log-sampling-thereafter")
	bindFlag(config.KeyAppOutputFormat, "output")

	// Return combined error if any bindings failed
	if len(errs) > 0 {
		return fmt.Errorf("failed to bind %d flag(s): %v", len(errs), errs)
	}

	return nil
}

func initConfig() error {
	configPaths := ConfigPaths()

	if cfgFile != "" {
		// Explicit --config flag takes highest priority
		viper.SetConfigFile(cfgFile)
	} else {
		// Let viper search for config files in priority order
		// Viper will look for config.yaml, config.yml, config.json, config.toml, etc.
		viper.SetConfigName(configPaths.ConfigName)

		// Search paths based on configured path mode.
		for _, searchPath := range configPaths.SearchPaths {
			viper.AddConfigPath(searchPath)
		}
	}

	// Set up environment variable handling with proper prefix
	envPrefix := EnvPrefix()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set default values from registry
	// IMPORTANT: Never set defaults directly with viper.SetDefault() here.
	// All defaults MUST be defined in internal/config/registry.go
	//
	// Thread-safety: This is called during startup before any concurrent access.
	// No synchronization needed as all config writes happen here in PersistentPreRunE.
	config.SetDefaults()

	// Validate default values to ensure they don't exceed limits
	// This catches programming errors in default value definitions
	if errs := config.ValidateAllConfigValues(viper.AllSettings()); len(errs) > 0 {
		log.Debug().Int("error_count", len(errs)).Msg("Invalid default configuration values detected")
		for i, err := range errs {
			log.Debug().Int("error_num", i+1).Err(err).Msg("Default validation error")
		}
		return fmt.Errorf("configuration has %d invalid default value(s) - this is a programming error", len(errs))
	}

	if err := viper.ReadInConfig(); err != nil {
		var configNotFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &configNotFoundErr) {
			configFileStatus = "No config file found, using defaults and environment variables"
		} else {
			// This error needs to be reported immediately
			log.Debug().Err(err).Msg("Failed to read config file")
			return fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		configFileStatus = "Using config file"
		configFileUsed = viper.ConfigFileUsed()

		// Security validation after viper finds and reads the config
		if err := config.ValidateConfigFileSecurity(configFileUsed, config.MaxConfigFileSize); err != nil {
			log.Debug().Err(err).Str("path", configFileUsed).Msg("Config file security validation failed")
			return fmt.Errorf("config file security validation failed: %w", err)
		}
	}

	// Validate registered config options (colors, log levels, etc.)
	if errs := config.ValidateRegisteredOptions(); len(errs) > 0 {
		for _, err := range errs {
			log.Debug().Err(err).Msg("Config validation error")
		}
		return fmt.Errorf("configuration validation failed: %w", errs[0])
	}

	return nil
}

// setupCommandConfig creates a PreRunE function that integrates with the root PersistentPreRunE
// to provide consistent configuration initialization with command-specific behavior.
// This pattern ensures that:
// 1. Root configuration is initialized first
// 2. Command-specific configuration is applied
// 3. Parent command's PreRunE is always called to maintain inheritance
func setupCommandConfig(cmd *cobra.Command) {
	// Store original PreRunE if it exists
	originalPreRunE := cmd.PreRunE

	// Create new PreRunE that applies command-specific configuration
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		// Call original PreRunE if it exists
		if originalPreRunE != nil {
			if err := originalPreRunE(c, args); err != nil {
				return err
			}
		}

		// Debug log that we're configuring this command
		log.Debug().Str("command", c.Name()).Msg("Applying command-specific configuration")

		// The common viper environment setup is already done in root's PersistentPreRunE
		// via the initConfig() function, so we don't need to repeat it here

		// IMPORTANT: Never set defaults directly with viper.SetDefault() here or in command files.
		// All defaults MUST be defined in internal/config/registry.go

		return nil
	}
}

// getConfigValueWithFlags retrieves a configuration value with the following precedence:
//  1. Command line flag (if explicitly set via --flagName)
//  2. Configuration from viper (environment variable or config file)
//  3. Zero value of type T (if neither flag nor config is set)
//
// The function uses type parameters to provide type-safe configuration retrieval.
// It handles type assertions safely, logging warnings if type conversion fails.
//
// Supported types: string, bool, int, float64, []string
//
// Example usage:
//
//	message := getConfigValueWithFlags[string](cmd, "message", "app.ping.output_message")
//	enabled := getConfigValueWithFlags[bool](cmd, "ui", "app.ping.ui")
//
// Parameters:
//   - cmd: The cobra.Command instance containing flags
//   - flagName: The name of the command-line flag (e.g., "message")
//   - viperKey: The viper configuration key (e.g., "app.ping.output_message")
//
// Returns:
//   - The configuration value of type T, or zero value if not found
func getConfigValueWithFlags[T any](cmd *cobra.Command, flagName string, viperKey string) T {
	var value T

	// Get the value from viper first (config file or env var), coerced to T.
	if v, ok := coerceViperValue[T](viperKey); ok {
		value = v
	}

	// If the flag was explicitly set, override the viper value
	if cmd.Flags().Changed(flagName) {
		// Handle different types appropriately
		switch any(value).(type) {
		case string:
			if v, err := cmd.Flags().GetString(flagName); err == nil {
				// Use safe type assertion with two-value form
				if convertedVal, ok := any(v).(T); ok {
					value = convertedVal
				} else {
					log.Warn().
						Str("flag", flagName).
						Str("expected_type", fmt.Sprintf("%T", value)).
						Str("actual_type", fmt.Sprintf("%T", v)).
						Msg("Type assertion failed for string flag, using current value")
				}
			}
		case bool:
			if v, err := cmd.Flags().GetBool(flagName); err == nil {
				if convertedVal, ok := any(v).(T); ok {
					value = convertedVal
				} else {
					log.Warn().
						Str("flag", flagName).
						Str("expected_type", fmt.Sprintf("%T", value)).
						Str("actual_type", fmt.Sprintf("%T", v)).
						Msg("Type assertion failed for bool flag, using current value")
				}
			}
		case int:
			if v, err := cmd.Flags().GetInt(flagName); err == nil {
				if convertedVal, ok := any(v).(T); ok {
					value = convertedVal
				} else {
					log.Warn().
						Str("flag", flagName).
						Str("expected_type", fmt.Sprintf("%T", value)).
						Str("actual_type", fmt.Sprintf("%T", v)).
						Msg("Type assertion failed for int flag, using current value")
				}
			}
		case float64:
			if v, err := cmd.Flags().GetFloat64(flagName); err == nil {
				if convertedVal, ok := any(v).(T); ok {
					value = convertedVal
				} else {
					log.Warn().
						Str("flag", flagName).
						Str("expected_type", fmt.Sprintf("%T", value)).
						Str("actual_type", fmt.Sprintf("%T", v)).
						Msg("Type assertion failed for float64 flag, using current value")
				}
			}
		case []string:
			if v, err := cmd.Flags().GetStringSlice(flagName); err == nil {
				if convertedVal, ok := any(v).(T); ok {
					value = convertedVal
				} else {
					log.Warn().
						Str("flag", flagName).
						Str("expected_type", fmt.Sprintf("%T", value)).
						Str("actual_type", fmt.Sprintf("%T", v)).
						Msg("Type assertion failed for string slice flag, using current value")
				}
			}
		}
	}

	return value
}

// getKeyValue retrieves a configuration value from Viper by key only.
//
// This function is used when flags are already bound to Viper and you want to
// retrieve the merged value (environment variables, config file, or defaults).
// It does NOT check command-line flags directly - use getConfigValueWithFlags for that.
//
// The function returns the zero value of type T if the key is not found or
// if type conversion fails.
//
// Supported types: any type T that can be stored in Viper
//
// Example usage:
//
//	format := getKeyValue[string]("app.docs.output_format")
//	count := getKeyValue[int]("app.max_items")
//
// Parameters:
//   - viperKey: The full viper configuration key (e.g., "app.docs.output_format")
//
// Returns:
//   - The configuration value of type T, or zero value if not found/conversion fails
func getKeyValue[T any](viperKey string) T {
	var zero T
	if v, ok := coerceViperValue[T](viperKey); ok {
		return v
	}
	return zero
}

// coerceViperValue reads viperKey and coerces the stored value to type T.
//
// Environment variables always arrive through viper as strings, so a plain
// viper.Get + Go type assertion (v.(bool)) silently fails for non-string config
// (bool/int/float/[]string) and drops the value to its zero. spf13/cast coerces
// (e.g. "true" -> true, "42" -> 42), which is the behavior users expect from
// env-var configuration.
//
// A value that genuinely cannot be coerced (e.g. a typo'd env var like
// FAIL_FAST=yse) is NOT silently dropped to zero — it is logged at WARN and the
// key is reported as unset so a default or flag can still win. Presence is
// detected with viper.Get != nil (which works for env-only keys, unlike
// viper.IsSet); the second return value reports whether a usable value was found.
func coerceViperValue[T any](viperKey string) (T, bool) {
	var zero T
	raw := viper.Get(viperKey)
	if raw == nil {
		return zero, false
	}

	var (
		val any
		err error
	)
	switch any(zero).(type) {
	case string:
		val, err = cast.ToStringE(raw)
	case bool:
		val, err = cast.ToBoolE(raw)
	case int:
		val, err = cast.ToIntE(raw)
	case float64:
		val, err = cast.ToFloat64E(raw)
	case []string:
		val, err = cast.ToStringSliceE(raw)
	default:
		if typedValue, ok := raw.(T); ok {
			return typedValue, true
		}
		return zero, false
	}

	if err != nil {
		log.Warn().
			Str("key", viperKey).
			Interface("value", raw).
			Err(err).
			Msg("config value could not be coerced to its declared type; ignoring it")
		return zero, false
	}
	return val.(T), true
}
