package infrastructure

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		want     zerolog.Level
	}{
		{"Debug level", "debug", zerolog.DebugLevel},
		{"Info level", "info", zerolog.InfoLevel},
		{"Warn level", "warn", zerolog.WarnLevel},
		{"Error level", "error", zerolog.ErrorLevel},
		{"Invalid level", "invalid", zerolog.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitLogger(tt.logLevel)
			assert.Equal(t, tt.want, zerolog.GlobalLevel())
		})
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger)
}

func TestSetLogOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	SetLogOutput(buf)

	logger := GetLogger()
	logger.Info().Msg("test message")

	assert.Contains(t, buf.String(), "test message")
}
