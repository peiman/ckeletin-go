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
	// Uncomment the following line if your bare application
	// has an action associated with it:
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./ckeletin-go.json5)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the logging level (debug, info, warn, error)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory with name "ckeletin-go" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigName("ckeletin-go")
	}

	viper.SetConfigType("json") // Set the config type to JSON

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Error reading config file:", err)
	}

	config, err := infrastructure.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// Initialize logger
	infrastructure.InitLogger(config.LogLevel)

	// You can use the loaded config here or pass it to other parts of your application
	fmt.Printf("Loaded configuration: %+v\n", config)
}
