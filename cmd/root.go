// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version information (populated via ldflags)
	Version = "dev"
	Commit  = ""
	Date    = ""

	cfgFile string

	// rootCmd represents the base command
	rootCmd = &cobra.Command{
		Use:   "ckeletin-go",
		Short: "A scaffold for building professional CLI applications in Go",
		Long:  `ckeletin-go is a scaffold project that helps you kickstart your Go CLI applications.`,
	}
)

// Execute adds all child commands to the root command
func Execute() {
	// Handle version flag
	rootCmd.Version = fmt.Sprintf("%s, commit %s, built at %s", Version, Commit, Date)

	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Command execution failed")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Define a persistent flag for specifying the config file
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ckeletin-go.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		// Use the config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ckeletin-go" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigName(".ckeletin-go")
	}

	viper.AutomaticEnv() // Read in environment variables that match

	// Set default configuration values
	viper.SetDefault("app.output_message", "Pong")
	viper.SetDefault("app.output_color", "white")
	viper.SetDefault("app.log_level", "info")
	viper.SetDefault("app.ui", false)

	// If a config file is found, read it
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
