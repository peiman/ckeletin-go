package logger

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// loggerMu protects mutable package-level state for thread safety.
	// In practice, Cobra is sequential, but this is correct for a public API.
	loggerMu sync.Mutex
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
	// currentConsoleTimeFormat is the console timestamp layout captured by Init.
	// rebuildLogger reuses it so a runtime level change does not pick up a later edit.
	// Empty means the console line has no timestamp. The file writer is separate.
	currentConsoleTimeFormat string
)

// Init initializes the logger with options from Viper.
// Supports dual logging: console (user-friendly) + file (detailed JSON).
// Call this once in rootCmd's PersistentPreRunE or main initialization.
//
// Configuration:
//   - app.log_level or app.log.console_level: Console log level (default: warn)
//   - app.log.file_enabled: Enable file logging (default: true; false opts out)
//   - app.log.file_path: Log file path (default: ./logs/ckeletin-go.log)
//   - app.log.file_level: File log level (default: debug)
//   - app.log.color_enabled: Color output (auto/true/false, default: auto)
//   - app.log.console_time_format: Console timestamp layout (default: RFC3339; empty omits it)
//
// When the resolved output mode is json, the console writer is
// disabled entirely; only the file (audit) writer receives entries.
//
// Init fails closed when file logging is explicitly enabled but the log file
// cannot be opened: a requested audit trail that cannot exist is an error,
// not a degradation.
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
	if jsonMode() {
		// JSON mode contract (CKSPEC-OUT-004): stdout carries exactly one
		// JSON envelope and stderr stays silent, so the console writer is
		// fully disabled. The file (audit) writer below is unaffected and
		// keeps receiving entries at its configured level. Disabling here,
		// before the writers are built, also keeps Init's own status logs
		// off the console.
		consoleLevel = zerolog.Disabled
	}
	currentConsoleLevel = consoleLevel

	// Determine if color should be enabled
	colorEnabled := isColorEnabled(out)
	currentColorEnabled = colorEnabled
	currentConsoleTimeFormat = loadConsoleTimeFormat()

	// Console writer with filtering
	consoleWriter := newConsoleWriter(out)

	filteredConsole := FilteredWriter{
		Writer:   consoleWriter,
		MinLevel: consoleLevel,
	}

	writers = append(writers, filteredConsole)

	// A second Init must not leave the previous audit file open. Windows
	// cannot delete that path while any handle remains, and only the latest
	// handle is stored in logFile.
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
		currentFileWriter = nil
	}

	// File writer (if enabled)
	if viper.GetBool(config.KeyAppLogFileEnabled) {
		fileLevel := getFileLogLevel()
		currentFileLevel = fileLevel
		filePath := viper.GetString(config.KeyAppLogFilePath)

		fileWriter, err := openLogFileWithRotation(filePath)
		if err == nil {
			// lumberjack opens the file lazily on FIRST write and swallows
			// later write errors, so probe now: an unwritable path must fail
			// Init instead of silently dropping every audit entry.
			_, err = fileWriter.Write(nil)
		}
		if err != nil {
			// Fail closed: file logging was EXPLICITLY enabled, so an audit
			// trail that cannot exist is an error, not a degradation. The old
			// Warn-and-continue path was unobservable in JSON mode (the
			// warning fired into the pre-init logger before log.Logger was
			// replaced), yielding exit 0 with zero audit entries.
			return fmt.Errorf("file logging enabled but log file %s cannot be opened: %w",
				SanitizePath(filePath), err)
		}

		logFile = fileWriter
		currentFileWriter = fileWriter // Store for rebuilding

		filteredFile := FilteredWriter{
			Writer:   fileWriter,
			MinLevel: fileLevel,
		}

		writers = append(writers, filteredFile)
	}

	log.Logger = newLogger(zerolog.MultiLevelWriter(writers...))

	// Set global log level to the most verbose level
	// This allows both writers to filter independently
	globalLevel := getGlobalLogLevel(consoleLevel)
	zerolog.SetGlobalLevel(globalLevel)

	// Log sampling status after logger is configured, so the message goes to
	// the writers configured above instead of the previous global logger
	if viper.GetBool(config.KeyAppLogSamplingEnabled) {
		log.Info().
			Int("initial", viper.GetInt(config.KeyAppLogSamplingInitial)).
			Int("thereafter", viper.GetInt(config.KeyAppLogSamplingThereafter)).
			Msg("Log sampling enabled")
	}

	// Log file logging status after logger is configured
	if viper.GetBool(config.KeyAppLogFileEnabled) && logFile != nil {
		filePath := viper.GetString(config.KeyAppLogFilePath)
		fileLevel := getFileLogLevel()
		log.Debug().
			Str("path", SanitizePath(filePath)).
			Str("file_level", fileLevel.String()).
			Msg("File logging enabled")
	}

	return nil
}

// Cleanup closes any open log files and performs cleanup.
// Should be called with defer after Init().
func Cleanup() {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if logFile != nil {
		if err := logFile.Close(); err != nil && !jsonMode() {
			// Can't use logger here as it might be cleaning up. In JSON mode
			// stderr must stay silent (CKSPEC-OUT-004), so the warning is
			// suppressed — the close failure cannot reach the audit file
			// either, since that is the very file that failed to close.
			fmt.Fprintf(os.Stderr, "Warning: failed to close log file: %v\n", err)
		}
		logFile = nil
	}
}

// jsonMode reports the resolved output mode without importing the output
// package. The output package imports the command framework, and the logger
// must stay free of that import so business logic can depend on logging
// without depending on rendering.
func jsonMode() bool {
	return strings.EqualFold(strings.TrimSpace(viper.GetString(config.KeyAppOutputFormat)), "json")
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
	if err != nil || consoleLevelStr == "" {
		level = zerolog.WarnLevel
		if consoleLevelStr != "" {
			log.Warn().
				Err(err).
				Str("provided_level", consoleLevelStr).
				Msg("Invalid console log level, defaulting to 'warn'")
		}
	}

	return level
}

// getFileLogLevel determines the file log level from configuration.
// app.log.file_level does not follow app.log_level. An empty value is the
// documented default. zerolog.ParseLevel treats "" as NoLevel with a nil
// error, which would drop debug lines, so empty is handled before parsing.
func getFileLogLevel() zerolog.Level {
	fileLevelStr := strings.TrimSpace(viper.GetString(config.KeyAppLogFileLevel))
	if fileLevelStr == "" {
		return zerolog.DebugLevel
	}

	level, err := zerolog.ParseLevel(fileLevelStr)
	if err != nil {
		log.Warn().
			Err(err).
			Str("provided_level", fileLevelStr).
			Msg("Invalid file log level, defaulting to 'debug'")
		return zerolog.DebugLevel
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

// clampSamplingValue converts a sampling config value to the uint32 zerolog
// requires, clamping out-of-range values to [1, math.MaxUint32]. The registry
// validators (ValidatePositiveInt) are the primary defense; this clamp is
// defense in depth for values that bypass config validation (e.g. direct
// viper.Set), where a negative value would wrap to ~4 billion and zero would
// make zerolog's BasicSampler drop every event after the burst.
func clampSamplingValue(n int) uint32 {
	v := int64(n)
	if v < 1 {
		return 1
	}
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}

// loadConsoleTimeFormat reads app.log.console_time_format.
// Unset (no default installed) stays RFC3339, which is what Init hardcoded before
// the key existed. A configured empty string, including a project default, drops
// the console timestamp. The file writer does not use this value.
func loadConsoleTimeFormat() string {
	if !viper.IsSet(config.KeyAppLogConsoleTimeFormat) {
		return time.RFC3339
	}
	return viper.GetString(config.KeyAppLogConsoleTimeFormat)
}

// newConsoleWriter builds the console writer from the format and color captured
// by Init. An empty layout removes the timestamp part. Zerolog would otherwise
// treat an empty TimeFormat as time.Kitchen.
func newConsoleWriter(out io.Writer) zerolog.ConsoleWriter {
	w := zerolog.ConsoleWriter{
		Out:        out,
		TimeFormat: currentConsoleTimeFormat,
		NoColor:    !currentColorEnabled,
	}
	if currentConsoleTimeFormat == "" {
		w.PartsExclude = []string{zerolog.TimestampFieldName}
	}
	return w
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

// sharedSink is the one console and file destination. Component loggers hold
// this pointer, so a level change reaches loggers created before it.
var sharedSink = &levelSink{}

// defaultFields are copied onto every logger built after they are set.
var defaultFields = map[string]string{}

type levelSink struct {
	mu sync.Mutex
	w  zerolog.LevelWriter
}

func (s *levelSink) set(w zerolog.LevelWriter) {
	s.mu.Lock()
	s.w = w
	s.mu.Unlock()
}

func (s *levelSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	w := s.w
	s.mu.Unlock()
	if w == nil {
		return len(p), nil
	}
	return w.Write(p)
}

func (s *levelSink) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	s.mu.Lock()
	w := s.w
	s.mu.Unlock()
	if w == nil {
		return len(p), nil
	}
	return w.WriteLevel(level, p)
}

func newLogger(w zerolog.LevelWriter) zerolog.Logger {
	sharedSink.set(w)
	ctx := zerolog.New(sharedSink).With().Timestamp()
	for key, value := range defaultFields {
		ctx = ctx.Str(key, value)
	}
	logger := ctx.Logger()
	if viper.GetBool(config.KeyAppLogSamplingEnabled) {
		initial := viper.GetInt(config.KeyAppLogSamplingInitial)
		thereafter := viper.GetInt(config.KeyAppLogSamplingThereafter)
		logger = logger.Sample(&zerolog.BurstSampler{
			Burst:       clampSamplingValue(initial),
			Period:      time.Second,
			NextSampler: &zerolog.BasicSampler{N: clampSamplingValue(thereafter)},
		})
	}
	return logger
}

// SetDefaultFields puts fields on every later line from the global logger and
// from component loggers created after this call. A field on the event with
// the same key wins.
func SetDefaultFields(fields map[string]string) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	defaultFields = make(map[string]string, len(fields))
	for key, value := range fields {
		defaultFields[key] = value
	}
	if currentConsoleWriter != nil {
		rebuildLogger()
	}
}

// Component returns a logger for one part of the program. It writes to the
// same console and file as the global logger, with component set to name.
// It does not open a file. SetConsoleLevel and SetFileLevel apply to a
// component logger created before or after the change.
func Component(name string) zerolog.Logger {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	return log.Logger.With().Str("component", name).Logger()
}

// rebuildLogger recreates the logger with current level settings.
// This is called when runtime log level adjustments are made.
func rebuildLogger() {
	var writers []io.Writer

	// Recreate console writer with the format captured at Init.
	consoleWriter := newConsoleWriter(currentConsoleWriter)

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

	// Update global logger. Component loggers keep this sink, so the new
	// levels apply to ones created before this rebuild.
	log.Logger = newLogger(zerolog.MultiLevelWriter(writers...))

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
	loggerMu.Lock()
	defer loggerMu.Unlock()
	currentConsoleLevel = level
	rebuildLogger()
	// Logged under loggerMu: log.Logger is rebuilt under this mutex, so
	// reading it after unlocking would race with concurrent Set*Level calls
	log.Info().
		Str("level", level.String()).
		Msg("Console log level changed")
}

// SetFileLevel dynamically changes the file log level at runtime.
// This allows adjusting file log verbosity without restarting the application.
func SetFileLevel(level zerolog.Level) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	currentFileLevel = level
	rebuildLogger()
	// Logged under loggerMu: see SetConsoleLevel
	log.Info().
		Str("level", level.String()).
		Msg("File log level changed")
}

// GetConsoleLevel returns the current console log level.
func GetConsoleLevel() zerolog.Level {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	return currentConsoleLevel
}

// GetFileLevel returns the current file log level.
func GetFileLevel() zerolog.Level {
	loggerMu.Lock()
	defer loggerMu.Unlock()
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
	// Close the audit file opened after the save. On Windows a test temp
	// directory cannot be removed while lumberjack still holds that file.
	Cleanup()
	log.Logger = logger
	zerolog.SetGlobalLevel(level)
}
