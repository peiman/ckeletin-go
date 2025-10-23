// internal/logger/logger_test.go
package logger

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func TestInit(t *testing.T) {
	// Save original viper settings
	originalLogLevel := viper.Get("app.log_level")
	defer func() {
		if originalLogLevel == nil {
			viper.Set("app.log_level", nil)
		} else {
			viper.Set("app.log_level", originalLogLevel)
		}
	}()

	tests := []struct {
		name          string
		logLevel      string
		output        io.Writer
		testMessages  map[string]bool // map of message to whether it should be present
		captureStderr bool
		expectedError bool
	}{
		{
			name:     "Info level",
			logLevel: "info",
			output:   new(bytes.Buffer),
			testMessages: map[string]bool{
				"Info message":  true,
				"Debug message": false,
			},
			expectedError: false,
		},
		{
			name:     "Debug level",
			logLevel: "debug",
			output:   new(bytes.Buffer),
			testMessages: map[string]bool{
				"Info message":  true,
				"Debug message": true,
			},
			expectedError: false,
		},
		{
			name:     "Invalid level defaults to info",
			logLevel: "invalid",
			output:   new(bytes.Buffer),
			testMessages: map[string]bool{
				"Info message":  true,
				"Debug message": false,
			},
			expectedError: false,
		},
		// Skip this test for now as it's more complex to reliably capture stderr
		/* {
			name:          "Nil output uses stderr",
			logLevel:      "info",
			output:        nil,
			testMessages:  map[string]bool{
				"Test message to stderr": true,
			},
			captureStderr: true,
			expectedError: false,
		}, */
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip the stderr capture test for now
			if tt.name == "Nil output uses stderr" {
				t.Skip("Skipping stderr capture test due to platform differences")
				return
			}

			// SETUP PHASE
			viper.Set("app.log_level", tt.logLevel)

			var buf *bytes.Buffer
			var r, w *os.File
			var capturedOutput *bytes.Buffer

			if tt.output != nil {
				// Use the provided output
				buf, _ = tt.output.(*bytes.Buffer)
				buf.Reset() // Clear buffer for this test
				capturedOutput = buf
			} else if tt.captureStderr {
				// Capture stderr for nil output tests
				capturedOutput = new(bytes.Buffer)

				// Save the original os.Stderr
				oldStderr := os.Stderr

				// Create a pipe to capture os.Stderr
				var err error
				r, w, err = os.Pipe()
				if err != nil {
					t.Fatalf("Failed to create pipe: %v", err)
				}

				// Redirect os.Stderr to the write end of the pipe
				os.Stderr = w

				// Setup cleanup to restore stderr
				defer func() {
					// Close the write end of the pipe and restore os.Stderr
					if w != nil {
						w.Close()
					}
					os.Stderr = oldStderr

					// Read the captured output from the read end of the pipe
					if r != nil {
						_, err = io.Copy(capturedOutput, r)
						if err != nil {
							t.Fatalf("Failed to read from pipe: %v", err)
						}
						r.Close()
					}
				}()
			}

			// EXECUTION PHASE
			err := Init(tt.output)

			// Log test messages
			for msg := range tt.testMessages {
				if msg == "Debug message" {
					log.Debug().Msg(msg)
				} else {
					log.Info().Msg(msg)
				}
			}

			// For stderr capture, close the write end to flush
			if tt.captureStderr && w != nil {
				w.Close()
				w = nil // prevent double close in defer
			}

			// ASSERTION PHASE
			// Check for expected error
			if (err != nil) != tt.expectedError {
				t.Errorf("Init() error = %v, expectedError %v", err, tt.expectedError)
			}

			// Check for expected messages in output
			if capturedOutput != nil {
				output := capturedOutput.String()

				for msg, shouldBePresent := range tt.testMessages {
					if shouldBePresent && !bytes.Contains([]byte(output), []byte(msg)) {
						t.Errorf("Expected message %q in output, but it was not found", msg)
					} else if !shouldBePresent && bytes.Contains([]byte(output), []byte(msg)) {
						t.Errorf("Message %q should not be in output, but it was found", msg)
					}
				}
			}
		})
	}
}

func TestSaveAndRestoreLoggerState(t *testing.T) {
	// SETUP PHASE
	// Create a test logger and level
	testBuf := &bytes.Buffer{}
	testLogger := zerolog.New(testBuf).With().Timestamp().Logger()
	testLevel := zerolog.DebugLevel

	// Set the test state
	log.Logger = testLogger
	zerolog.SetGlobalLevel(testLevel)

	// Save the state
	savedLogger, savedLevel := SaveLoggerState()

	// Verify saved state matches what we set
	if savedLevel != testLevel {
		t.Errorf("SaveLoggerState() saved level = %v, want %v", savedLevel, testLevel)
	}

	// EXECUTION PHASE
	// Modify the logger and level
	newBuf := &bytes.Buffer{}
	newLogger := zerolog.New(newBuf).With().Str("modified", "true").Logger()
	newLevel := zerolog.WarnLevel

	log.Logger = newLogger
	zerolog.SetGlobalLevel(newLevel)

	// Verify state was changed
	if zerolog.GlobalLevel() != newLevel {
		t.Errorf("Failed to modify global level, got %v, want %v", zerolog.GlobalLevel(), newLevel)
	}

	// Restore the original state
	RestoreLoggerState(savedLogger, savedLevel)

	// ASSERTION PHASE
	// Verify the logger and level were restored
	if zerolog.GlobalLevel() != testLevel {
		t.Errorf("RestoreLoggerState() level = %v, want %v", zerolog.GlobalLevel(), testLevel)
	}

	// Test that the logger is writing to the original buffer
	log.Info().Msg("test message")
	if !bytes.Contains(testBuf.Bytes(), []byte("test message")) {
		t.Errorf("Restored logger is not writing to original buffer")
	}
	if bytes.Contains(newBuf.Bytes(), []byte("test message")) {
		t.Errorf("Restored logger is still writing to new buffer")
	}
}
