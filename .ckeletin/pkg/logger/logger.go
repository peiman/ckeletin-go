package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// logFile holds the open log file handle for cleanup
	logFile io.Closer
	// currentConsoleLevel holds the current console log level for runtime adjustment
	currentConsoleLevel zerolog.Level
	// currentFileLevel holds the current file log level for runtime adjustment
	currentFileLevel zerolog.Level
	// currentConsoleWriter holds the console writer for rebuilding
	currentConsoleWriter io.Writer
	// currentFileWriter holds the file writer for rebuilding
	currentFileWriter io.WriteCloser
	// currentColorEnabled stores whether colors are enabled
	currentColorEnabled bool
)

// Init initializes the logger with options from Viper.
// Supports dual logging: console (user-friendly) + file (detailed JSON).
// Call this once in rootCmd's PersistentPreRunE or main initialization.
//
// Configuration:
//   - app.log_level or app.log.console_level: Console log level (default: info)
//   - app.log.file_enabled: Enable file logging (default: false)
//   - app.log.file_path: Log file path (default: ./logs/ckeletin-go.log)
//   - app.log.file_level: File log level (default: debug)
//   - app.log.color_enabled: Color output (auto/true/false, default: auto)
//
// Example:
//
//	if err := logger.Init(nil); err != nil {
//	    return err
//	}
//	defer logger.Cleanup()
func Init(out io.Writer) error {
	// Default to stderr for logging to avoid polluting stdout (Data Stream)
	if out == nil {
		out = os.Stderr
	}

	// Store console writer for rebuilding
	currentConsoleWriter = out

	var writers []io.Writer

	// Get console log level (with backward compatibility)
	consoleLevel := getConsoleLogLevel()
	currentConsoleLevel = consoleLevel

	// Determine if color should be enabled
	colorEnabled := isColorEnabled(out)
	currentColorEnabled = colorEnabled

	// Console writer with filtering
	consoleWriter := zerolog.ConsoleWriter{
		Out:        out,
		TimeFormat: time.RFC3339,
		NoColor:    !colorEnabled,
	}

	filteredConsole := FilteredWriter{
		Writer:   consoleWriter,
		MinLevel: consoleLevel,
	}

	writers = append(writers, filteredConsole)

	// File writer (if enabled)
	if viper.GetBool(config.KeyAppLogFileEnabled) {
		fileLevel := getFileLogLevel()
		currentFileLevel = fileLevel
		filePath := viper.GetString(config.KeyAppLogFilePath)

		fileWriter, err := openLogFileWithRotation(filePath)
		if err != nil {
			// Log warning but continue with console-only logging
			log.Warn().
				Err(err).
				Str("path", SanitizePath(filePath)).
				Msg("Failed to open log file, continuing with console-only logging")
		} else {
			logFile = fileWriter
			currentFileWriter = fileWriter // Store for rebuilding

			filteredFile := FilteredWriter{
				Writer:   fileWriter,
				MinLevel: fileLevel,
			}

			writers = append(writers, filteredFile)
		}
	}

	// Create multi-writer
	multi := zerolog.MultiLevelWriter(writers...)

	// Create logger with timestamp
	logger := zerolog.New(multi).With().Timestamp().Logger()

	// Apply sampling if enabled
	if viper.GetBool(config.KeyAppLogSamplingEnabled) {
		initial := viper.GetInt(config.KeyAppLogSamplingInitial)
		thereafter := viper.GetInt(config.KeyAppLogSamplingThereafter)
		logger = logger.Sample(&zerolog.BurstSampler{
			Burst:       uint32(initial), //nolint:gosec // Config values are positive
			Period:      time.Second,
			NextSampler: &zerolog.BasicSampler{N: uint32(thereafter)}, //nolint:gosec // Config values are positive
		})
		log.Info().
			Int("initial", initial).
			Int("thereafter", thereafter).
			Msg("Log sampling enabled")
	}

	log.Logger = logger

	// Set global log level to the most verbose level
	// This allows both writers to filter independently
	globalLevel := getGlobalLogLevel(consoleLevel)
	zerolog.SetGlobalLevel(globalLevel)

	// Log file logging status after logger is configured
	if viper.GetBool(config.KeyAppLogFileEnabled) && logFile != nil {
		filePath := viper.GetString(config.KeyAppLogFilePath)
		fileLevel := getFileLogLevel()
		log.Info().
			Str("path", SanitizePath(filePath)).
			Str("file_level", fileLevel.String()).
			Msg("File logging enabled")
	}

	return nil
}

// Cleanup closes any open log files and performs cleanup.
// Should be called with defer after Init().
func Cleanup() {
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			// Can't use logger here as it might be cleaning up
			fmt.Fprintf(os.Stderr, "Warning: failed to close log file: %v\n", err)
		}
		logFile = nil
	}
}

// getConsoleLogLevel determines the console log level from configuration.
// Supports backward compatibility with app.log_level.
func getConsoleLogLevel() zerolog.Level {
	// Try new config key first
	consoleLevelStr := viper.GetString(config.KeyAppLogConsoleLevel)
	if consoleLevelStr == "" {
		// Fall back to legacy config key for backward compatibility
		consoleLevelStr = viper.GetString(config.KeyAppLogLevel)
	}

	level, err := zerolog.ParseLevel(consoleLevelStr)
	if err != nil {
		level = zerolog.InfoLevel
		log.Warn().
			Err(err).
			Str("provided_level", consoleLevelStr).
			Msg("Invalid console log level, defaulting to 'info'")
	}

	return level
}

// getFileLogLevel determines the file log level from configuration.
func getFileLogLevel() zerolog.Level {
	fileLevelStr := viper.GetString(config.KeyAppLogFileLevel)

	level, err := zerolog.ParseLevel(fileLevelStr)
	if err != nil {
		level = zerolog.DebugLevel
		log.Warn().
			Err(err).
			Str("provided_level", fileLevelStr).
			Msg("Invalid file log level, defaulting to 'debug'")
	}

	return level
}

// getGlobalLogLevel determines the global log level.
// This should be the most verbose level to allow per-writer filtering.
func getGlobalLogLevel(consoleLevel zerolog.Level) zerolog.Level {
	if !viper.GetBool(config.KeyAppLogFileEnabled) {
		return consoleLevel
	}

	fileLevel := getFileLogLevel()

	// Return the more verbose level (lower numeric value)
	// TraceLevel=-1, DebugLevel=0, InfoLevel=1, etc.
	if fileLevel < consoleLevel {
		return fileLevel
	}

	return consoleLevel
}

// isColorEnabled determines if colored output should be enabled.
func isColorEnabled(out io.Writer) bool {
	colorConfig := viper.GetString(config.KeyAppLogColorEnabled)

	switch colorConfig {
	case "true":
		return true
	case "false":
		return false
	case "auto", "":
		// Auto-detect: check if output is a TTY
		if file, ok := out.(*os.File); ok {
			return isatty.IsTerminal(file.Fd())
		}
		return false
	default:
		log.Warn().
			Str("value", colorConfig).
			Msg("Invalid color_enabled value, using auto-detection")
		return false
	}
}

// openLogFileWithRotation opens the log file with lumberjack rotation support.
// Creates parent directories if they don't exist.
func openLogFileWithRotation(path string) (io.WriteCloser, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Configure lumberjack for log rotation
	logger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    viper.GetInt(config.KeyAppLogFileMaxSize),    // megabytes
		MaxBackups: viper.GetInt(config.KeyAppLogFileMaxBackups), // number of backups
		MaxAge:     viper.GetInt(config.KeyAppLogFileMaxAge),     // days
		Compress:   viper.GetBool(config.KeyAppLogFileCompress),  // compress with gzip
	}

	return logger, nil
}

// rebuildLogger recreates the logger with current level settings.
// This is called when runtime log level adjustments are made.
func rebuildLogger() {
	var writers []io.Writer

	// Recreate console writer with current level
	consoleWriter := zerolog.ConsoleWriter{
		Out:        currentConsoleWriter,
		TimeFormat: time.RFC3339,
		NoColor:    !currentColorEnabled,
	}

	filteredConsole := FilteredWriter{
		Writer:   consoleWriter,
		MinLevel: currentConsoleLevel,
	}

	writers = append(writers, filteredConsole)

	// Add file writer if it exists
	if currentFileWriter != nil {
		filteredFile := FilteredWriter{
			Writer:   currentFileWriter,
			MinLevel: currentFileLevel,
		}
		writers = append(writers, filteredFile)
	}

	// Create multi-writer
	multi := zerolog.MultiLevelWriter(writers...)

	// Create logger with timestamp
	logger := zerolog.New(multi).With().Timestamp().Logger()

	// Apply sampling if enabled
	if viper.GetBool(config.KeyAppLogSamplingEnabled) {
		initial := viper.GetInt(config.KeyAppLogSamplingInitial)
		thereafter := viper.GetInt(config.KeyAppLogSamplingThereafter)
		logger = logger.Sample(&zerolog.BurstSampler{
			Burst:       uint32(initial), //nolint:gosec // Config values are positive
			Period:      time.Second,
			NextSampler: &zerolog.BasicSampler{N: uint32(thereafter)}, //nolint:gosec // Config values are positive
		})
	}

	// Update global logger
	log.Logger = logger

	// Update global log level to the most verbose level
	globalLevel := currentConsoleLevel
	if currentFileWriter != nil && currentFileLevel < currentConsoleLevel {
		globalLevel = currentFileLevel
	}
	zerolog.SetGlobalLevel(globalLevel)
}

// SetConsoleLevel dynamically changes the console log level at runtime.
// This allows adjusting verbosity without restarting the application.
func SetConsoleLevel(level zerolog.Level) {
	currentConsoleLevel = level
	rebuildLogger()
	log.Info().
		Str("level", level.String()).
		Msg("Console log level changed")
}

// SetFileLevel dynamically changes the file log level at runtime.
// This allows adjusting file log verbosity without restarting the application.
func SetFileLevel(level zerolog.Level) {
	currentFileLevel = level
	rebuildLogger()
	log.Info().
		Str("level", level.String()).
		Msg("File log level changed")
}

// GetConsoleLevel returns the current console log level.
func GetConsoleLevel() zerolog.Level {
	return currentConsoleLevel
}

// GetFileLevel returns the current file log level.
func GetFileLevel() zerolog.Level {
	return currentFileLevel
}

// SaveLoggerState returns the current global logger and log level for later restoration.
// This is useful in tests to avoid global state pollution.
// Use with RestoreLoggerState in a defer statement:
//
//	savedLogger, savedLevel := logger.SaveLoggerState()
//	defer logger.RestoreLoggerState(savedLogger, savedLevel)
func SaveLoggerState() (zerolog.Logger, zerolog.Level) {
	return log.Logger, zerolog.GlobalLevel()
}

// RestoreLoggerState restores the global logger and log level to a previously saved state.
// This is useful in tests to avoid global state pollution.
func RestoreLoggerState(logger zerolog.Logger, level zerolog.Level) {
	log.Logger = logger
	zerolog.SetGlobalLevel(level)
}
