// internal/logger/logger_test.go
package logger

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func TestInit(t *testing.T) {
	buf := new(bytes.Buffer)
	viper.Set("app.log_level", "info")
	if err := Init(buf); err != nil {
		t.Fatalf("Init returned an error: %v", err)
	}
	log.Info().Msg("Test message")

	if !bytes.Contains(buf.Bytes(), []byte("Test message")) {
		t.Errorf("Expected 'Test message' in log output")
	}

	// Test with invalid log level
	viper.Set("app.log_level", "invalid")
	buf.Reset()
	if err := Init(buf); err != nil {
		t.Fatalf("Init returned an error: %v", err)
	}
	log.Info().Msg("Test message with invalid level")

	if !bytes.Contains(buf.Bytes(), []byte("Test message with invalid level")) {
		t.Errorf("Expected 'Test message with invalid level' in log output")
	}

	// Test with 'debug' log level
	viper.Set("app.log_level", "debug")
	buf.Reset()
	if err := Init(buf); err != nil {
		t.Fatalf("Init returned an error: %v", err)
	}
	log.Debug().Msg("Debug message")

	if !bytes.Contains(buf.Bytes(), []byte("Debug message")) {
		t.Errorf("Expected 'Debug message' in log output")
	}
}
