package cmd

import (
	"fmt"
	"os"

	"github.com/peiman/ckeletin-go/internal/errors"
	"github.com/peiman/ckeletin-go/internal/infrastructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var logLevel string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ckeletin-go",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from ckeletin-go!")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./ckeletin-go.json)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the logging level (debug, info, warn, error)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Initialize logger
	infrastructure.InitLogger(logLevel)
	logger := infrastructure.GetLogger()

	// Reset viper configuration
	viper.Reset()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory with name "ckeletin-go.json"
		viper.AddConfigPath(".")
		viper.SetConfigName("ckeletin-go")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Warn().Msg("No config file found. Using defaults and environment variables.")
		} else {
			appErr := errors.NewAppError(errors.ErrInvalidConfig, "Error reading config file", err)
			infrastructure.LogError(appErr, "Failed to read config file", nil)
			os.Exit(1)
		}
	} else {
		logger.Info().Str("config_file", viper.ConfigFileUsed()).Msg("Using config file")
	}

	// Reload the configuration
	viper.Set("LogLevel", viper.GetString("LogLevel"))
	viper.Set("Server.Port", viper.GetInt("Server.Port"))
	viper.Set("Server.Host", viper.GetString("Server.Host"))

	config, err := infrastructure.LoadConfig()
	if err != nil {
		appErr := errors.NewAppError(errors.ErrInvalidConfig, "Error loading config", err)
		infrastructure.LogError(appErr, "Failed to load configuration", nil)
		os.Exit(1)
	}

	logger.Info().Interface("config", config).Msg("Loaded configuration")
}
