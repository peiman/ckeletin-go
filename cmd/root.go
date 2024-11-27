package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/peiman/ckeletin-go/internal/logger"
	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "ckeletin-go",
		Short: "A scaffold for building professional CLI applications in Go",
		Long:  `ckeletin-go is a scaffold project that helps you kickstart your Go CLI applications.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize the logger
			if err := logger.Init(nil); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
				os.Exit(1)
			}
			return nil
		},
	}
)

func Execute() {
	rootCmd.Version = fmt.Sprintf("%s, commit %s, built at %s", Version, Commit, Date)
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Command execution failed")
		os.Exit(1)
	}
}

func RootCommand() *cobra.Command {
	return rootCmd
}

// cmd/root.go

func init() {
	cobra.OnInitialize(initConfig)

	// Define persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.ckeletin-go.yaml)")
	if err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind flag 'config'")
	}

	rootCmd.PersistentFlags().String("log-level", "info", "Set the log level (trace, debug, info, warn, error, fatal, panic)")
	if err := viper.BindPFlag("app.log_level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind flag 'log-level'")
	}

	// Attach subcommands
	uiRunner := &ui.DefaultUIRunner{}
	rootCmd.AddCommand(NewPingCommand(uiRunner))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(".ckeletin-go")
	}

	// Handle environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set default values for global configurations
	viper.SetDefault("app.log_level", "info")
	// Other global defaults...

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info().Msg("No config file found, using defaults and environment variables")
		} else {
			log.Fatal().Err(err).Msg("Failed to read config file")
		}
	} else {
		log.Info().Str("config_file", viper.ConfigFileUsed()).Msg("Using config file")
	}
}
