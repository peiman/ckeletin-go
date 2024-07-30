package cmd

import (
	"fmt"
	"os"

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
var Execute = func() error {
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

	configManager := infrastructure.NewConfigManager(cfgFile)
	if err := configManager.EnsureConfig(); err != nil {
		logger.Error().Err(err).Msg("Failed to ensure config file exists")
		os.Exit(1)
	}

	viper.SetConfigFile(configManager.ConfigPath)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		logger.Error().Err(err).Msg("Failed to read config file")
		os.Exit(1)
	}

	logger.Info().Str("config_file", viper.ConfigFileUsed()).Msg("Using config file")

	config, err := infrastructure.LoadConfig()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load configuration")
		os.Exit(1)
	}

	logger.Info().Interface("config", config).Msg("Loaded configuration")
}
