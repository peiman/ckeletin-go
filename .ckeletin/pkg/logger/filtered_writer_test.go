package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestFilteredWriter_WriteLevel(t *testing.T) {
	tests := []struct {
		name        string
		minLevel    zerolog.Level
		writeLevel  zerolog.Level
		message     string
		shouldWrite bool
	}{
		// Test with InfoLevel minimum
		{
			name:        "Info minimum - Trace message filtered",
			minLevel:    zerolog.InfoLevel,
			writeLevel:  zerolog.TraceLevel,
			message:     "trace message",
			shouldWrite: false,
		},
		{
			name:        "Info minimum - Debug message filtered",
			minLevel:    zerolog.InfoLevel,
			writeLevel:  zerolog.DebugLevel,
			message:     "debug message",
			shouldWrite: false,
		},
		{
			name:        "Info minimum - Info message written",
			minLevel:    zerolog.InfoLevel,
			writeLevel:  zerolog.InfoLevel,
			message:     "info message",
			shouldWrite: true,
		},
		{
			name:        "Info minimum - Warn message written",
			minLevel:    zerolog.InfoLevel,
			writeLevel:  zerolog.WarnLevel,
			message:     "warn message",
			shouldWrite: true,
		},
		{
			name:        "Info minimum - Error message written",
			minLevel:    zerolog.InfoLevel,
			writeLevel:  zerolog.ErrorLevel,
			message:     "error message",
			shouldWrite: true,
		},
		// Test with DebugLevel minimum
		{
			name:        "Debug minimum - Trace message filtered",
			minLevel:    zerolog.DebugLevel,
			writeLevel:  zerolog.TraceLevel,
			message:     "trace message",
			shouldWrite: false,
		},
		{
			name:        "Debug minimum - Debug message written",
			minLevel:    zerolog.DebugLevel,
			writeLevel:  zerolog.DebugLevel,
			message:     "debug message",
			shouldWrite: true,
		},
		{
			name:        "Debug minimum - Info message written",
			minLevel:    zerolog.DebugLevel,
			writeLevel:  zerolog.InfoLevel,
			message:     "info message",
			shouldWrite: true,
		},
		// Test with ErrorLevel minimum
		{
			name:        "Error minimum - Info message filtered",
			minLevel:    zerolog.ErrorLevel,
			writeLevel:  zerolog.InfoLevel,
			message:     "info message",
			shouldWrite: false,
		},
		{
			name:        "Error minimum - Error message written",
			minLevel:    zerolog.ErrorLevel,
			writeLevel:  zerolog.ErrorLevel,
			message:     "error message",
			shouldWrite: true,
		},
		{
			name:        "Error minimum - Fatal message written",
			minLevel:    zerolog.ErrorLevel,
			writeLevel:  zerolog.FatalLevel,
			message:     "fatal message",
			shouldWrite: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := FilteredWriter{
				Writer:   &buf,
				MinLevel: tt.minLevel,
			}

			// Write test message
			message := []byte(tt.message)
			n, err := writer.WriteLevel(tt.writeLevel, message)

			// Check for errors
			if err != nil {
				t.Errorf("WriteLevel() returned error: %v", err)
			}

			// Check return value (should always return message length)
			if n != len(message) {
				t.Errorf("WriteLevel() returned n = %d, want %d", n, len(message))
			}

			// Check if message was written or filtered
			written := buf.String()
			if tt.shouldWrite {
				if written != tt.message {
					t.Errorf("Expected message to be written, got: %q, want: %q", written, tt.message)
				}
			} else {
				if written != "" {
					t.Errorf("Expected message to be filtered, but got: %q", written)
				}
			}
		})
	}
}

func TestFilteredWriter_Write(t *testing.T) {
	// Test that Write() always passes through (no filtering)
	var buf bytes.Buffer
	writer := FilteredWriter{
		Writer:   &buf,
		MinLevel: zerolog.ErrorLevel, // High minimum level
	}

	message := []byte("test message")
	n, err := writer.Write(message)

	if err != nil {
		t.Errorf("Write() returned error: %v", err)
	}

	if n != len(message) {
		t.Errorf("Write() returned n = %d, want %d", n, len(message))
	}

	if buf.String() != string(message) {
		t.Errorf("Write() wrote %q, want %q", buf.String(), string(message))
	}
}

func TestFilteredWriter_WithZerolog(t *testing.T) {
	// Integration test: FilteredWriter with actual zerolog logger
	var consoleBuf bytes.Buffer
	var fileBuf bytes.Buffer

	// Console writer: INFO and above
	consoleWriter := FilteredWriter{
		Writer:   &consoleBuf,
		MinLevel: zerolog.InfoLevel,
	}

	// File writer: DEBUG and above
	fileWriter := FilteredWriter{
		Writer:   &fileBuf,
		MinLevel: zerolog.DebugLevel,
	}

	// Create logger with both writers
	multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	logger := zerolog.New(multi).With().Timestamp().Logger()

	// Write logs at different levels
	logger.Trace().Msg("trace message")
	logger.Debug().Msg("debug message")
	logger.Info().Msg("info message")
	logger.Warn().Msg("warn message")
	logger.Error().Msg("error message")

	// Check console output (should have info, warn, error)
	consoleOutput := consoleBuf.String()
	if !strings.Contains(consoleOutput, "info message") {
		t.Error("Console should contain 'info message'")
	}
	if !strings.Contains(consoleOutput, "warn message") {
		t.Error("Console should contain 'warn message'")
	}
	if !strings.Contains(consoleOutput, "error message") {
		t.Error("Console should contain 'error message'")
	}
	if strings.Contains(consoleOutput, "trace message") {
		t.Error("Console should NOT contain 'trace message'")
	}
	if strings.Contains(consoleOutput, "debug message") {
		t.Error("Console should NOT contain 'debug message'")
	}

	// Check file output (should have debug, info, warn, error)
	fileOutput := fileBuf.String()
	if !strings.Contains(fileOutput, "debug message") {
		t.Error("File should contain 'debug message'")
	}
	if !strings.Contains(fileOutput, "info message") {
		t.Error("File should contain 'info message'")
	}
	if !strings.Contains(fileOutput, "warn message") {
		t.Error("File should contain 'warn message'")
	}
	if !strings.Contains(fileOutput, "error message") {
		t.Error("File should contain 'error message'")
	}
	if strings.Contains(fileOutput, "trace message") {
		t.Error("File should NOT contain 'trace message'")
	}
}

func TestFilteredWriter_LevelComparison(t *testing.T) {
	// Verify our understanding of zerolog level comparison
	// Higher numeric values = higher severity
	levels := []struct {
		level zerolog.Level
		value int8
	}{
		{zerolog.TraceLevel, int8(zerolog.TraceLevel)},
		{zerolog.DebugLevel, int8(zerolog.DebugLevel)},
		{zerolog.InfoLevel, int8(zerolog.InfoLevel)},
		{zerolog.WarnLevel, int8(zerolog.WarnLevel)},
		{zerolog.ErrorLevel, int8(zerolog.ErrorLevel)},
		{zerolog.FatalLevel, int8(zerolog.FatalLevel)},
		{zerolog.PanicLevel, int8(zerolog.PanicLevel)},
	}

	t.Logf("Zerolog level values (higher = higher severity):")
	for _, l := range levels {
		t.Logf("  %s = %d", l.level, l.value)
	}

	// Verify comparison logic: higher numeric values = higher severity
	if !(zerolog.ErrorLevel > zerolog.WarnLevel) {
		t.Error("Expected ErrorLevel > WarnLevel (higher severity)")
	}
	if !(zerolog.WarnLevel > zerolog.InfoLevel) {
		t.Error("Expected WarnLevel > InfoLevel (higher severity)")
	}
	if !(zerolog.InfoLevel > zerolog.DebugLevel) {
		t.Error("Expected InfoLevel > DebugLevel (higher severity)")
	}
	if !(zerolog.DebugLevel > zerolog.TraceLevel) {
		t.Error("Expected DebugLevel > TraceLevel (higher severity)")
	}
}
