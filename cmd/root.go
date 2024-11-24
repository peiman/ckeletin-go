// Package cmd implements the command-line interface for the application.
package cmd

import (
	"fmt"
	"os"

	"github.com/peiman/ckeletin-go/internal/infrastructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	logLevel string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "ckeletin-go",
	Short: "A brief description of your application.",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("Hello from ckeletin-go!")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
var Execute = func() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags for use across all commands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./ckeletin-go.json)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", infrastructure.DefaultLogLevel.String(),
		`Set the logging level (trace, debug, info, warn, error, fatal, panic)`)

	// Local flags only for this command
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	// Initialize logger with command line flag value first
	if err := infrastructure.InitLogger(logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		osExit(1)
		return
	}
	logger := infrastructure.GetLogger()

	configManager := infrastructure.NewConfigManager(cfgFile)
	if err := configManager.EnsureConfig(); err != nil {
		logger.Error().Err(err).Msg("Failed to ensure config file exists")
		osExit(1)
		return
	}

	viper.SetConfigFile(configManager.ConfigPath)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		logger.Error().Err(err).Msg("Failed to read config file")
		osExit(1)
		return
	}

	logger.Info().Str("config_file", viper.ConfigFileUsed()).Msg("Using config file")

	config, err := infrastructure.LoadConfig()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load configuration")
		osExit(1)
		return
	}

	// Update log level from config if it wasn't specified on command line
	if !rootCmd.PersistentFlags().Changed("log-level") {
		if err := infrastructure.InitLogger(config.LogLevel.String()); err != nil {
			logger.Error().Err(err).Msg("Failed to update log level from config")
			osExit(1)
			return
		}
		logger = infrastructure.GetLogger()
	}

	logger.Info().Interface("config", config).Msg("Loaded configuration")
}
