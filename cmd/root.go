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
)

// EnvPrefix returns a sanitized environment variable prefix based on the binary name
func EnvPrefix() string {
	// Convert to uppercase and replace non-alphanumeric characters with underscore
	prefix := strings.ToUpper(binaryName)
	re := regexp.MustCompile(`[^A-Z0-9]`)
	prefix = re.ReplaceAllString(prefix, "_")

	// Ensure it doesn't start with a number (invalid for env vars)
	if prefix != "" && prefix[0] >= '0' && prefix[0] <= '9' {
		prefix = "_" + prefix
	}

	// Handle case where all characters were special and got replaced
	re = regexp.MustCompile(`^_+$`)
	if re.MatchString(prefix) {
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
	RootCmd.Version = fmt.Sprintf("%s, commit %s, built at %s", Version, Commit, Date)
	return RootCmd.Execute()
}

func init() {
	configPaths := ConfigPaths()
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("Config file (default is %s)", configPaths.DefaultPath))
	if err := viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'config' flag")
	}

	RootCmd.PersistentFlags().String("log-level", "info", "Set the log level (trace, debug, info, warn, error, fatal, panic)")
	if err := viper.BindPFlag("app.log_level", RootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-level'")
	}
}

func initConfig() error {
	configPaths := ConfigPaths()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(configPaths.DefaultName)
		viper.SetConfigType(configPaths.Extension)
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

// getConfigValue retrieves a configuration value with the following precedence:
// 1. Command line flag (if set)
// 2. Configuration from viper (environment variable or config file)
// This consolidates the common pattern of checking if a flag is set and using its value
func getConfigValue[T any](cmd *cobra.Command, flagName string, viperKey string) T {
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
				// We need to use interface conversion since Go can't directly assign
				// to the type parameter value - this is a bit verbose but type-safe
				value = any(v).(T)
			}
		case bool:
			if v, err := cmd.Flags().GetBool(flagName); err == nil {
				value = any(v).(T)
			}
		case int:
			if v, err := cmd.Flags().GetInt(flagName); err == nil {
				value = any(v).(T)
			}
		case float64:
			if v, err := cmd.Flags().GetFloat64(flagName); err == nil {
				value = any(v).(T)
			}
		case []string:
			if v, err := cmd.Flags().GetStringSlice(flagName); err == nil {
				value = any(v).(T)
			}
		}
	}

	return value
}

// getKeyValue retrieves a configuration value from Viper by key only.
// Flags should already be bound to Viper; this function prefers Viper's merged value.
func getKeyValue[T any](viperKey string) T {
	var zero T
	if v := viper.Get(viperKey); v != nil {
		if typedValue, ok := v.(T); ok {
			return typedValue
		}
	}
	return zero
}
