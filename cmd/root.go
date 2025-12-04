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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile          string
	Version          = "dev"
	Commit           = ""
	Date             = ""
	binaryName       = "" // MUST be injected via ldflags (see Taskfile.yml LDFLAGS)
	configFileStatus string
	configFileUsed   string

	// Compiled regex patterns for EnvPrefix()
	// Compiled once at package initialization for better performance
	nonAlphanumericRegex = regexp.MustCompile(`[^A-Z0-9]`)
	onlyUnderscoresRegex = regexp.MustCompile(`^_+$`)
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

// ConfigPaths returns standard paths and filenames for config files based on the binary name
func ConfigPaths() struct {
	// Default config name with dot prefix (e.g. ".myapp")
	DefaultName string
	// Config file extension
	Extension string
	// Default full config name (e.g. ".myapp.yaml")
	DefaultFullName string
	// Default config file with home directory (e.g. "$HOME/.myapp.yaml")
	DefaultPath string
	// Default ignore pattern for gitignore (e.g. "myapp.yaml")
	IgnorePattern string
} {
	ext := "yaml"
	defaultName := fmt.Sprintf(".%s", binaryName)
	defaultFullName := fmt.Sprintf("%s.%s", defaultName, ext)

	home, err := os.UserHomeDir()
	defaultPath := defaultFullName // Fallback if home dir not available
	if err == nil {
		defaultPath = filepath.Join(home, defaultFullName)
	}

	// Used for .gitignore - without leading dot
	ignorePattern := fmt.Sprintf("%s.%s", binaryName, ext)

	return struct {
		DefaultName     string
		Extension       string
		DefaultFullName string
		DefaultPath     string
		IgnorePattern   string
	}{
		DefaultName:     defaultName,
		Extension:       ext,
		DefaultFullName: defaultFullName,
		DefaultPath:     defaultPath,
		IgnorePattern:   ignorePattern,
	}
}

// Export RootCmd so that tests in other packages can manipulate it without getters/setters.
var RootCmd = &cobra.Command{
	Use:   binaryName,
	Short: "A scaffold for building professional CLI applications in Go",
	Long: fmt.Sprintf(`%s is a scaffold project that helps you kickstart your Go CLI applications.
It integrates Cobra, Viper, Zerolog, and Bubble Tea, along with a testing framework.`, binaryName),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Bind flags to viper first (must happen before initConfig)
		if err := bindFlags(cmd); err != nil {
			return fmt.Errorf("failed to bind flags: %w", err)
		}

		// Initialize configuration
		if err := initConfig(); err != nil {
			return err
		}

		// Initialize logger with configuration values
		if err := logger.Init(nil); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		// Log config status after logger is initialized
		if configFileStatus != "" {
			if configFileUsed != "" {
				log.Info().Str("config_file", logger.SanitizePath(configFileUsed)).Msg(configFileStatus)
			} else {
				log.Debug().Msg(configFileStatus)
			}
		}

		return nil
	},
}

func Execute() error {
	// Ensure logger cleanup on exit
	defer logger.Cleanup()

	RootCmd.Version = fmt.Sprintf("%s, commit %s, built at %s", Version, Commit, Date)
	return RootCmd.Execute()
}

func init() {
	// Fallback for development/testing when ldflags aren't injected
	// Production builds MUST inject binaryName via ldflags (see Taskfile.yml LDFLAGS)
	if binaryName == "" {
		binaryName = "ckeletin-go"
	}

	configPaths := ConfigPaths()

	// Define all persistent flags (flag definitions only - bindings happen in bindFlags())
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("Config file (default is %s)", configPaths.DefaultPath))

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

	// Return combined error if any bindings failed
	if len(errs) > 0 {
		return fmt.Errorf("failed to bind %d flag(s): %v", len(errs), errs)
	}

	return nil
}

func initConfig() error {
	configPaths := ConfigPaths()

	var configFilePath string
	if cfgFile != "" {
		configFilePath = cfgFile
		viper.SetConfigFile(cfgFile)
	} else {
		// Add current directory first (highest priority after --config)
		viper.AddConfigPath(".")

		// Add home directory with fallback for containerized environments
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		// Note: No error if HOME is unavailable - will search current dir only

		viper.SetConfigName(configPaths.DefaultName)
		viper.SetConfigType(configPaths.Extension)

		// Try to determine which config will be used (for security validation)
		// Check current directory first
		currentDirConfig := configPaths.DefaultFullName
		if _, err := os.Stat(currentDirConfig); err == nil {
			configFilePath = currentDirConfig
		} else if home != "" {
			// Check home directory if current dir doesn't have config
			homeConfig := filepath.Join(home, configPaths.DefaultFullName)
			if _, err := os.Stat(homeConfig); err == nil {
				configFilePath = homeConfig
			}
		}
	}

	// Security validation if config file path is known
	if configFilePath != "" {
		// Validate file security (size and permissions)
		if err := config.ValidateConfigFileSecurity(configFilePath, config.MaxConfigFileSize); err != nil {
			log.Error().Err(err).Str("path", configFilePath).Msg("Config file security validation failed")
			return fmt.Errorf("config file security validation failed: %w", err)
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
		log.Error().Int("error_count", len(errs)).Msg("Invalid default configuration values detected")
		for i, err := range errs {
			log.Error().Int("error_num", i+1).Err(err).Msg("Default validation error")
		}
		return fmt.Errorf("configuration has %d invalid default value(s) - this is a programming error", len(errs))
	}

	if err := viper.ReadInConfig(); err != nil {
		var configNotFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &configNotFoundErr) {
			configFileStatus = "No config file found, using defaults and environment variables"
		} else {
			// This error needs to be reported immediately
			log.Error().Err(err).Msg("Failed to read config file")
			return fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		configFileStatus = "Using config file"
		configFileUsed = viper.ConfigFileUsed()
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

	// Get the value from viper first (this will be from config file or env var)
	if v := viper.Get(viperKey); v != nil {
		if typedValue, ok := v.(T); ok {
			value = typedValue
		}
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
	if v := viper.Get(viperKey); v != nil {
		if typedValue, ok := v.(T); ok {
			return typedValue
		}
	}
	return zero
}
