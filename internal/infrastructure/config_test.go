package infrastructure

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
	viper.Reset()
}

func (s *ConfigTestSuite) TestLoadConfig() {
	tests := []struct {
		name         string
		configValues map[string]interface{}
		expected     *Config
		expectError  bool
	}{
		{
			name: "Default Values",
			configValues: map[string]interface{}{
				"logLevel": "info", // Test string format
			},
			expected: &Config{
				LogLevel: DefaultLogLevel,
				Ping: PingConfig{
					DefaultCount:  DefaultPingCount,
					OutputMessage: DefaultPingMessage,
					ColoredOutput: false,
				},
			},
		},
		{
			name: "Custom Values with Numeric Level",
			configValues: map[string]interface{}{
				"logLevel":           int8(zerolog.DebugLevel), // Test numeric format
				"ping.defaultCount":  5,
				"ping.outputMessage": "hello",
				"ping.coloredOutput": true,
			},
			expected: &Config{
				LogLevel: zerolog.DebugLevel,
				Ping: PingConfig{
					DefaultCount:  5,
					OutputMessage: "hello",
					ColoredOutput: true,
				},
			},
		},
		{
			name: "Custom Values with String Level",
			configValues: map[string]interface{}{
				"logLevel":           "debug", // Test string format
				"ping.defaultCount":  5,
				"ping.outputMessage": "hello",
				"ping.coloredOutput": true,
			},
			expected: &Config{
				LogLevel: zerolog.DebugLevel,
				Ping: PingConfig{
					DefaultCount:  5,
					OutputMessage: "hello",
					ColoredOutput: true,
				},
			},
		},
		{
			name: "Invalid String Level",
			configValues: map[string]interface{}{
				"logLevel": "invalid",
			},
			expectError: true,
		},
		{
			name: "Invalid Numeric Level",
			configValues: map[string]interface{}{
				"logLevel": int8(-10),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			viper.Reset()

			for k, v := range tt.configValues {
				viper.Set(k, v)
			}

			config, err := LoadConfig()
			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expected, config)
			}
		})
	}
}
