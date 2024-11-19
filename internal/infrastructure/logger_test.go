package infrastructure

import (
	"bytes"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		want     zerolog.Level
		wantErr  bool
	}{
		{"Debug level", "debug", zerolog.DebugLevel, false},
		{"Info level", "info", zerolog.InfoLevel, false},
		{"Warn level", "warn", zerolog.WarnLevel, false},
		{"Error level", "error", zerolog.ErrorLevel, false},
		{"Invalid level", "invalid", zerolog.InfoLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitLogger(tt.logLevel)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, zerolog.GlobalLevel())
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)
}

func TestSetLogOutput(t *testing.T) {
	// Save the original logger and restore it after the test
	originalLogger := log.Logger
	defer func() {
		log.Logger = originalLogger
		SetLogOutput(os.Stdout) // Reset to stdout for other tests
	}()

	buf := &bytes.Buffer{}
	SetLogOutput(buf)

	// Test different log levels
	tests := []struct {
		level    zerolog.Level
		message  string
		expected bool
	}{
		{zerolog.DebugLevel, "Debug message", true},
		{zerolog.InfoLevel, "Info message", true},
		{zerolog.WarnLevel, "Warn message", true},
		{zerolog.ErrorLevel, "Error message", true},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			zerolog.SetGlobalLevel(tt.level)
			log.Debug().Msg("Debug message")
			log.Info().Msg("Info message")
			log.Warn().Msg("Warn message")
			log.Error().Msg("Error message")

			output := buf.String()
			if tt.expected {
				assert.Contains(t, output, tt.message, "Log should contain %s at %s level", tt.message, tt.level)
			} else {
				assert.NotContains(t, output, tt.message, "Log should not contain %s at %s level", tt.message, tt.level)
			}
			buf.Reset() // Clear the buffer for the next test
		})
	}
}
