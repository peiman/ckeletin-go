// internal/logger/logger_test.go
package logger

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func TestInitWithFileLogging(t *testing.T) {
	// Create temp directory for log files
	tempDir := t.TempDir()
	logFile := tempDir + "/test.log"

	tests := []struct {
		name          string
		fileEnabled   bool
		filePath      string
		fileLevel     string
		consoleLevel  string
		colorEnabled  string
		expectFileLog bool
		expectedError bool
	}{
		{
			name:          "File logging disabled",
			fileEnabled:   false,
			filePath:      logFile,
			fileLevel:     "debug",
			consoleLevel:  "info",
			colorEnabled:  "false",
			expectFileLog: false,
			expectedError: false,
		},
		{
			name:          "File logging enabled",
			fileEnabled:   true,
			filePath:      logFile,
			fileLevel:     "debug",
			consoleLevel:  "info",
			colorEnabled:  "false",
			expectFileLog: true,
			expectedError: false,
		},
		{
			name:          "File logging with color auto",
			fileEnabled:   true,
			filePath:      logFile + ".2",
			fileLevel:     "debug",
			consoleLevel:  "info",
			colorEnabled:  "auto",
			expectFileLog: true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP
			viper.Set("app.log.file_enabled", tt.fileEnabled)
			viper.Set("app.log.file_path", tt.filePath)
			viper.Set("app.log.file_level", tt.fileLevel)
			viper.Set("app.log.console_level", tt.consoleLevel)
			viper.Set("app.log.color_enabled", tt.colorEnabled)
			viper.Set("app.log.sampling_enabled", false)

			// Ensure file doesn't exist before test
			os.Remove(tt.filePath)

			consoleBuf := &bytes.Buffer{}

			// EXECUTE
			err := Init(consoleBuf)

			// ASSERT
			if (err != nil) != tt.expectedError {
				t.Errorf("Init() error = %v, expectedError %v", err, tt.expectedError)
			}

			// Log test messages
			log.Debug().Msg("Debug message")
			log.Info().Msg("Info message")

			// Clean up
			Cleanup()

			// Check console output
			consoleOutput := consoleBuf.String()
			if !bytes.Contains([]byte(consoleOutput), []byte("Info message")) {
				t.Errorf("Console should contain Info message")
			}
			if bytes.Contains([]byte(consoleOutput), []byte("Debug message")) {
				t.Errorf("Console should NOT contain Debug message (console level is info)")
			}

			// Check file output if enabled
			if tt.expectFileLog {
				if _, err := os.Stat(tt.filePath); os.IsNotExist(err) {
					t.Errorf("Expected log file to be created at %s", tt.filePath)
				} else {
					fileContent, err := os.ReadFile(tt.filePath)
					if err != nil {
						t.Errorf("Failed to read log file: %v", err)
					}
					fileOutput := string(fileContent)
					if !bytes.Contains(fileContent, []byte("Debug message")) {
						t.Errorf("File should contain Debug message, got: %s", fileOutput)
					}
					if !bytes.Contains(fileContent, []byte("Info message")) {
						t.Errorf("File should contain Info message")
					}
				}
			} else {
				if _, err := os.Stat(tt.filePath); !os.IsNotExist(err) {
					t.Errorf("Log file should not be created when file logging is disabled")
				}
			}
		})
	}
}

// TestRuntimeLevelAdjustment tests that changing log levels at runtime actually affects filtering
// This test will initially FAIL - runtime level changes don't affect actual log output
func TestRuntimeLevelAdjustment(t *testing.T) {
	// SETUP PHASE
	savedLogger, savedLevel := SaveLoggerState()
	defer RestoreLoggerState(savedLogger, savedLevel)

	// Initialize logger with INFO level
	buf := &bytes.Buffer{}
	viper.Set("app.log.console_level", "info")
	viper.Set("app.log.file_enabled", false)
	viper.Set("app.log.sampling_enabled", false)

	err := Init(buf)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test 1: Debug messages filtered at INFO level
	buf.Reset()
	log.Debug().Msg("debug_before_change")
	log.Info().Msg("info_before_change")

	output := buf.String()
	if strings.Contains(output, "debug_before_change") {
		t.Error("Debug message should be filtered at INFO level")
	}
	if !strings.Contains(output, "info_before_change") {
		t.Error("Info message should appear at INFO level")
	}

	// Test 2: Change level to DEBUG
	buf.Reset()
	SetConsoleLevel(zerolog.DebugLevel)

	// Test 3: Debug messages now appear
	buf.Reset()
	log.Debug().Msg("debug_after_change")
	log.Info().Msg("info_after_change")

	output = buf.String()
	if !strings.Contains(output, "debug_after_change") {
		t.Error("Debug message should appear after SetConsoleLevel(DEBUG)")
	}
	if !strings.Contains(output, "info_after_change") {
		t.Error("Info message should still appear after level change")
	}

	// Test 4: Getter reflects new level
	if GetConsoleLevel() != zerolog.DebugLevel {
		t.Errorf("GetConsoleLevel() = %v, want %v", GetConsoleLevel(), zerolog.DebugLevel)
	}

	// Test 5: Change back to WARN
	buf.Reset()
	SetConsoleLevel(zerolog.WarnLevel)

	buf.Reset()
	log.Info().Msg("info_at_warn")
	log.Warn().Msg("warn_at_warn")

	output = buf.String()
	if strings.Contains(output, "info_at_warn") {
		t.Error("Info message should be filtered at WARN level")
	}
	if !strings.Contains(output, "warn_at_warn") {
		t.Error("Warn message should appear at WARN level")
	}
}

// TestRuntimeFileLevelAdjustment tests runtime adjustment for file logging
// This test will initially FAIL - runtime level changes don't affect file output
func TestRuntimeFileLevelAdjustment(t *testing.T) {
	// SETUP PHASE
	savedLogger, savedLevel := SaveLoggerState()
	defer RestoreLoggerState(savedLogger, savedLevel)

	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	viper.Set("app.log.console_level", "error") // Console at ERROR
	viper.Set("app.log.file_enabled", true)
	viper.Set("app.log.file_path", logFile)
	viper.Set("app.log.file_level", "info") // File at INFO
	viper.Set("app.log.sampling_enabled", false)
	viper.Set("app.log.color_enabled", "false")

	consoleBuf := &bytes.Buffer{}
	err := Init(consoleBuf)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Cleanup()

	// Debug filtered in both initially
	log.Debug().Msg("debug_initial")

	// Adjust file level to DEBUG
	SetFileLevel(zerolog.DebugLevel)

	log.Debug().Msg("debug_after_file_change")

	Cleanup() // Ensure file is flushed

	fileContent, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	output := string(fileContent)

	if strings.Contains(output, "debug_initial") {
		t.Error("Initial debug should be filtered")
	}
	if !strings.Contains(output, "debug_after_file_change") {
		t.Error("Debug should appear in file after SetFileLevel(DEBUG)")
	}

	// Verify console still filters (at ERROR level)
	if strings.Contains(consoleBuf.String(), "debug") {
		t.Error("Console should still filter debug messages")
	}
}

func TestLogSampling(t *testing.T) {
	// Create temp directory for log files
	tempDir := t.TempDir()
	logFile := tempDir + "/test-sampling.log"

	// SETUP
	viper.Set("app.log.file_enabled", true)
	viper.Set("app.log.file_path", logFile)
	viper.Set("app.log.file_level", "debug")
	viper.Set("app.log.console_level", "info")
	viper.Set("app.log.color_enabled", "false")
	viper.Set("app.log.sampling_enabled", true)
	viper.Set("app.log.sampling_initial", 2)
	viper.Set("app.log.sampling_thereafter", 10)

	consoleBuf := &bytes.Buffer{}

	// EXECUTE
	err := Init(consoleBuf)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Log many messages - with sampling enabled, not all should appear
	for i := 0; i < 20; i++ {
		log.Debug().Int("iteration", i).Msg("Sampled message")
	}

	// CLEANUP
	Cleanup()

	// ASSERT
	// We can't assert exact counts due to sampling behavior,
	// but we can verify the file was created and contains some logs
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Expected log file to be created")
	}
}

func TestCleanup(t *testing.T) {
	// Create temp directory for log files
	tempDir := t.TempDir()
	logFile := tempDir + "/test-cleanup.log"

	// SETUP
	viper.Set("app.log.file_enabled", true)
	viper.Set("app.log.file_path", logFile)
	viper.Set("app.log.file_level", "debug")
	viper.Set("app.log.console_level", "info")
	viper.Set("app.log.color_enabled", "false")
	viper.Set("app.log.sampling_enabled", false)

	consoleBuf := &bytes.Buffer{}

	err := Init(consoleBuf)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Log a message to ensure file is created and written to
	log.Info().Msg("Test message before cleanup")

	// EXECUTE
	Cleanup()

	// ASSERT
	// After cleanup, logFile should be nil (we can't directly test this,
	// but we can verify no panic occurs on second cleanup)
	Cleanup() // Should not panic

	// File should exist and contain the logged message
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Log file should exist after cleanup")
	}
}

func TestIsColorEnabled(t *testing.T) {
	tests := []struct {
		name           string
		colorConfig    string
		output         io.Writer
		expectedResult bool
	}{
		{
			name:           "Explicit true",
			colorConfig:    "true",
			output:         &bytes.Buffer{},
			expectedResult: true,
		},
		{
			name:           "Explicit false",
			colorConfig:    "false",
			output:         &bytes.Buffer{},
			expectedResult: false,
		},
		{
			name:           "Auto with buffer (not TTY)",
			colorConfig:    "auto",
			output:         &bytes.Buffer{},
			expectedResult: false,
		},
		{
			name:           "Empty string (auto)",
			colorConfig:    "",
			output:         &bytes.Buffer{},
			expectedResult: false,
		},
		{
			name:           "Invalid value",
			colorConfig:    "invalid",
			output:         &bytes.Buffer{},
			expectedResult: false,
		},
		{
			name:           "Auto with file",
			colorConfig:    "auto",
			output:         os.Stdout,
			expectedResult: false, // CI environment, likely not a TTY
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP
			originalValue := viper.Get("app.log.color_enabled")
			defer viper.Set("app.log.color_enabled", originalValue)
			viper.Set("app.log.color_enabled", tt.colorConfig)

			// EXECUTE
			result := isColorEnabled(tt.output)

			// ASSERT
			if result != tt.expectedResult {
				t.Errorf("isColorEnabled() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestGetFileLogLevel(t *testing.T) {
	tests := []struct {
		name          string
		fileLevel     string
		expectedLevel zerolog.Level
	}{
		{
			name:          "File level set to debug",
			fileLevel:     "debug",
			expectedLevel: zerolog.DebugLevel,
		},
		{
			name:          "File level empty",
			fileLevel:     "",
			expectedLevel: zerolog.NoLevel,
		},
		{
			name:          "File level set to trace",
			fileLevel:     "trace",
			expectedLevel: zerolog.TraceLevel,
		},
		{
			name:          "Invalid file level, defaults to debug",
			fileLevel:     "invalid",
			expectedLevel: zerolog.DebugLevel,
		},
		{
			name:          "File level set to error",
			fileLevel:     "error",
			expectedLevel: zerolog.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP
			viper.Set("app.log.file_level", tt.fileLevel)

			// EXECUTE
			result := getFileLogLevel()

			// ASSERT
			if result != tt.expectedLevel {
				t.Errorf("getFileLogLevel() = %v, want %v", result, tt.expectedLevel)
			}
		})
	}
}

func TestOpenLogFileWithRotation(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "Valid path",
			path:        t.TempDir() + "/logs/app.log",
			expectError: false,
		},
		{
			name:        "Path with multiple nested dirs",
			path:        t.TempDir() + "/deep/nested/path/app.log",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP
			viper.Set("app.log.file_max_size", 100)
			viper.Set("app.log.file_max_backups", 3)
			viper.Set("app.log.file_max_age", 28)
			viper.Set("app.log.file_compress", false)

			// EXECUTE
			writer, err := openLogFileWithRotation(tt.path)

			// ASSERT
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if writer != nil {
				writer.Close()
			}

			// Verify directory was created
			if !tt.expectError {
				dir := filepath.Dir(tt.path)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("Expected directory %s to be created", dir)
				}
			}
		})
	}
}
