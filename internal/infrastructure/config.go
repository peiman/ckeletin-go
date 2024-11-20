// Package infrastructure handles all infrastructure-related operations.
package infrastructure

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Config holds all configuration for our program.
type Config struct {
	LogLevel zerolog.Level `mapstructure:"logLevel" json:"logLevel"`
	Ping     PingConfig    `mapstructure:"ping" json:"ping"`
}

// PingConfig holds configuration for the ping command
type PingConfig struct {
	DefaultCount  int    `mapstructure:"defaultCount" json:"defaultCount"`
	OutputMessage string `mapstructure:"outputMessage" json:"outputMessage"`
	ColoredOutput bool   `mapstructure:"coloredOutput" json:"coloredOutput"`
}

// decodeLevelHook helps viper convert strings or integers to zerolog.Level
func decodeLevelHook() viper.DecoderConfigOption {
	return viper.DecodeHook(
		func(_ reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
			if t != reflect.TypeOf(zerolog.Level(0)) {
				return data, nil
			}

			switch v := data.(type) {
			case string:
				level, err := zerolog.ParseLevel(strings.ToLower(v))
				if err != nil {
					return nil, fmt.Errorf("invalid log level %q: %w", v, err)
				}
				return level, nil
			case int, int8, int16, int32, int64:
				// Convert to int8 for zerolog.Level
				level := zerolog.Level(reflect.ValueOf(v).Int())
				if level < zerolog.TraceLevel || level > zerolog.Disabled {
					return nil, fmt.Errorf("invalid log level: %d", level)
				}
				return level, nil
			default:
				return nil, fmt.Errorf("invalid log level type: %T", data)
			}
		},
	)
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (*Config, error) {
	// Set defaults using zerolog's constants
	viper.SetDefault("logLevel", DefaultLogLevel)
	viper.SetDefault("ping.defaultCount", DefaultPingCount)
	viper.SetDefault("ping.outputMessage", DefaultPingMessage)
	viper.SetDefault("ping.coloredOutput", false)

	var config Config
	if err := viper.Unmarshal(&config, decodeLevelHook()); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &config, nil
}
