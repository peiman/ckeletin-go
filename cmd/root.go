// cmd/root.go

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/peiman/ckeletin-go/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile          string
	Version          = "dev"
	Commit           = ""
	Date             = ""
	binaryName       = "ckeletin-go"
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
		if err := initConfig(); err != nil {
			return err
		}
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
	configPaths := ConfigPaths()
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("Config file (default is %s)", configPaths.DefaultPath))
	if err := viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'config' flag")
	}

	// Legacy log level flag (for backward compatibility)
	RootCmd.PersistentFlags().String("log-level", "info", "Set the log level (trace, debug, info, warn, error, fatal, panic)")
	if err := viper.BindPFlag(config.KeyAppLogLevel, RootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-level'")
	}

	// Dual logging configuration flags
	RootCmd.PersistentFlags().String("log-console-level", "", "Console log level (trace, debug, info, warn, error, fatal, panic). If empty, uses --log-level.")
	if err := viper.BindPFlag(config.KeyAppLogConsoleLevel, RootCmd.PersistentFlags().Lookup("log-console-level")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-console-level'")
	}

	RootCmd.PersistentFlags().Bool("log-file-enabled", false, "Enable file logging to capture detailed logs")
	if err := viper.BindPFlag(config.KeyAppLogFileEnabled, RootCmd.PersistentFlags().Lookup("log-file-enabled")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-file-enabled'")
	}

	RootCmd.PersistentFlags().String("log-file-path", "./logs/ckeletin-go.log", "Path to the log file")
	if err := viper.BindPFlag(config.KeyAppLogFilePath, RootCmd.PersistentFlags().Lookup("log-file-path")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-file-path'")
	}

	RootCmd.PersistentFlags().String("log-file-level", "debug", "File log level (trace, debug, info, warn, error, fatal, panic)")
	if err := viper.BindPFlag(config.KeyAppLogFileLevel, RootCmd.PersistentFlags().Lookup("log-file-level")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-file-level'")
	}

	RootCmd.PersistentFlags().String("log-color", "auto", "Enable colored console output (auto, true, false)")
	if err := viper.BindPFlag(config.KeyAppLogColorEnabled, RootCmd.PersistentFlags().Lookup("log-color")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-color'")
	}

	// Log rotation configuration flags
	RootCmd.PersistentFlags().Int("log-file-max-size", 100, "Maximum size in megabytes before log file is rotated")
	if err := viper.BindPFlag(config.KeyAppLogFileMaxSize, RootCmd.PersistentFlags().Lookup("log-file-max-size")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-file-max-size'")
	}

	RootCmd.PersistentFlags().Int("log-file-max-backups", 3, "Maximum number of old log files to retain")
	if err := viper.BindPFlag(config.KeyAppLogFileMaxBackups, RootCmd.PersistentFlags().Lookup("log-file-max-backups")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-file-max-backups'")
	}

	RootCmd.PersistentFlags().Int("log-file-max-age", 28, "Maximum number of days to retain old log files")
	if err := viper.BindPFlag(config.KeyAppLogFileMaxAge, RootCmd.PersistentFlags().Lookup("log-file-max-age")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-file-max-age'")
	}

	RootCmd.PersistentFlags().Bool("log-file-compress", false, "Compress rotated log files with gzip")
	if err := viper.BindPFlag(config.KeyAppLogFileCompress, RootCmd.PersistentFlags().Lookup("log-file-compress")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-file-compress'")
	}

	// Log sampling configuration flags
	RootCmd.PersistentFlags().Bool("log-sampling-enabled", false, "Enable log sampling for high-volume scenarios")
	if err := viper.BindPFlag(config.KeyAppLogSamplingEnabled, RootCmd.PersistentFlags().Lookup("log-sampling-enabled")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-sampling-enabled'")
	}

	RootCmd.PersistentFlags().Int("log-sampling-initial", 100, "Number of messages to log per second before sampling")
	if err := viper.BindPFlag(config.KeyAppLogSamplingInitial, RootCmd.PersistentFlags().Lookup("log-sampling-initial")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-sampling-initial'")
	}

	RootCmd.PersistentFlags().Int("log-sampling-thereafter", 100, "Number of messages to log thereafter per second")
	if err := viper.BindPFlag(config.KeyAppLogSamplingThereafter, RootCmd.PersistentFlags().Lookup("log-sampling-thereafter")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-sampling-thereafter'")
	}
}

func initConfig() error {
	configPaths := ConfigPaths()

	var configFilePath string
	if cfgFile != "" {
		configFilePath = cfgFile
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		// Check if default config file exists
		defaultConfigPath := filepath.Join(home, configPaths.DefaultFullName)
		if _, err := os.Stat(defaultConfigPath); err == nil {
			configFilePath = defaultConfigPath
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(configPaths.DefaultName)
		viper.SetConfigType(configPaths.Extension)
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
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
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
