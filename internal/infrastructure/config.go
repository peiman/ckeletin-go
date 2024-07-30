package infrastructure

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration for our program
type Config struct {
	LogLevel string
	Server   ServerConfig
	Database DatabaseConfig
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Port int
	Host string
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (*Config, error) {
	viper.SetConfigName("ckeletin-go") // name of config file (without extension)
	viper.SetConfigType("json")        // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")           // optionally look for config in the working directory

	// Set default values
	viper.SetDefault("LogLevel", "info")
	viper.SetDefault("Server.Port", 8080)
	viper.SetDefault("Server.Host", "localhost")
	viper.SetDefault("Database.Host", "localhost")
	viper.SetDefault("Database.Port", 5432)
	viper.SetDefault("Database.User", "postgres")
	viper.SetDefault("Database.Name", "ckeletin")

	// Environment variable configurations
	viper.SetEnvPrefix("CKELETIN") // will be uppercased automatically
	viper.AutomaticEnv()           // read in environment variables that match

	// Environment variable mappings
	viper.BindEnv("LogLevel", "CKELETIN_LOGLEVEL")
	viper.BindEnv("Server.Port", "CKELETIN_SERVER_PORT")
	viper.BindEnv("Server.Host", "CKELETIN_SERVER_HOST")
	viper.BindEnv("Database.Host", "CKELETIN_DATABASE_HOST")
	viper.BindEnv("Database.Port", "CKELETIN_DATABASE_PORT")
	viper.BindEnv("Database.User", "CKELETIN_DATABASE_USER")
	viper.BindEnv("Database.Password", "CKELETIN_DATABASE_PASSWORD")
	viper.BindEnv("Database.Name", "CKELETIN_DATABASE_NAME")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fmt.Println("No config file found. Using defaults and environment variables.")
		} else {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("error reading config file: %s", err)
		}
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %s", err)
	}

	return &config, nil
}
