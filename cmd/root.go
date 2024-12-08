package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/peiman/ckeletin-go/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// The following variables are set by ldflags at build time.
// For example:
// -X 'github.com/peiman/ckeletin-go/cmd.binaryName=your-binary'
// -X 'github.com/peiman/ckeletin-go/cmd.Version=1.0.0'
var (
	cfgFile    string
	Version    = "dev"
	Commit     = ""
	Date       = ""
	binaryName = "ckeletin-go" // default, overridden by ldflags if desired
)

var rootCmd = &cobra.Command{
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
		return nil
	},
}

func Execute() error {
	rootCmd.Version = fmt.Sprintf("%s, commit %s, built at %s", Version, Commit, Date)
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("Config file (default is $HOME/.%s.yaml)", binaryName))
	if err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'config' flag")
	}

	rootCmd.PersistentFlags().String("log-level", "info", "Set the log level (trace, debug, info, warn, error, fatal, panic)")
	if err := viper.BindPFlag("app.log_level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'log-level'")
	}
}

func initConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(fmt.Sprintf(".%s", binaryName))
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("app.log_level", "info")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info().Msg("No config file found, using defaults and environment variables")
		} else {
			log.Error().Err(err).Msg("Failed to read config file")
			return fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		log.Info().Str("config_file", viper.ConfigFileUsed()).Msg("Using config file")
	}

	return nil
}

// For testing main and commands
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func SetRootCmd(cmd *cobra.Command) {
	rootCmd = cmd
}
