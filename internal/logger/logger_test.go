// internal/logger/logger_test.go
package logger

import (
	"bytes"
	"io"
	"os"
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

func TestInit_ValidLogLevel(t *testing.T) {
	buf := new(bytes.Buffer)
	viper.Set("app.log_level", "debug")

	err := Init(buf)
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	log.Debug().Msg("Debug message")
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Debug message")) {
		t.Errorf("Expected 'Debug message' in log output")
	}
}

func TestInit_InvalidLogLevel(t *testing.T) {
	buf := new(bytes.Buffer)
	viper.Set("app.log_level", "invalid")

	err := Init(buf)
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	log.Info().Msg("Info message")
	log.Debug().Msg("Debug message")

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("Info message")) {
		t.Errorf("Expected 'Info message' in log output")
	}
	if bytes.Contains([]byte(output), []byte("Debug message")) {
		t.Errorf("Did not expect 'Debug message' in log output")
	}
}

func TestInit_NilOutput(t *testing.T) {
	// Save the original os.Stderr
	oldStderr := os.Stderr

	// Create a pipe to capture os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Redirect os.Stderr to the write end of the pipe
	os.Stderr = w

	// Initialize the logger with nil output
	if err := Init(nil); err != nil {
		t.Fatalf("Failed to initialize logger with nil output: %v", err)
	}

	// Log a message to test the output
	log.Info().Msg("Test message to stderr")

	// Close the write end of the pipe and restore os.Stderr
	w.Close()
	os.Stderr = oldStderr

	// Read the captured output from the read end of the pipe
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	// Close the read end of the pipe
	r.Close()

	// Verify that the output contains the test message
	if !bytes.Contains(buf.Bytes(), []byte("Test message to stderr")) {
		t.Errorf("Expected 'Test message to stderr' in output, got '%s'", buf.String())
	}
}
