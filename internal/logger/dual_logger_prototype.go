// internal/logger/dual_logger_prototype.go
//
// Prototype implementation of dual logging system.
// This demonstrates how to configure console and file outputs with different log levels.
//
// This file is a PROTOTYPE and not intended for production use yet.

package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// DualLoggerConfig holds configuration for dual logging setup
type DualLoggerConfig struct {
	// Console configuration
	ConsoleEnabled bool
	ConsoleLevel   zerolog.Level
	ConsoleColor   bool
	ConsoleWriter  io.Writer // defaults to os.Stdout

	// File configuration
	FileEnabled bool
	FileLevel   zerolog.Level
	FilePath    string
	FileWriter  io.Writer // if nil, will open FilePath
}

// InitDualLogger initializes a logger with separate console and file outputs.
// This is a prototype implementation to validate the dual logging approach.
//
// Example usage:
//
//	config := DualLoggerConfig{
//	    ConsoleEnabled: true,
//	    ConsoleLevel:   zerolog.InfoLevel,
//	    ConsoleColor:   true,
//	    FileEnabled:    true,
//	    FileLevel:      zerolog.DebugLevel,
//	    FilePath:       "./logs/app.log",
//	}
//	logger, cleanup, err := InitDualLogger(config)
//	if err != nil {
//	    log.Fatal().Err(err).Msg("Failed to initialize logger")
//	}
//	defer cleanup()
func InitDualLogger(config DualLoggerConfig) (zerolog.Logger, func(), error) {
	var writers []io.Writer
	var cleanupFuncs []func()

	// Console output
	if config.ConsoleEnabled {
		consoleOut := config.ConsoleWriter
		if consoleOut == nil {
			consoleOut = os.Stdout
		}

		consoleWriter := zerolog.ConsoleWriter{
			Out:        consoleOut,
			TimeFormat: time.RFC3339,
			NoColor:    !config.ConsoleColor,
		}

		filteredConsole := FilteredWriter{
			Writer:   consoleWriter,
			MinLevel: config.ConsoleLevel,
		}

		writers = append(writers, filteredConsole)
	}

	// File output
	if config.FileEnabled {
		fileOut := config.FileWriter
		if fileOut == nil && config.FilePath != "" {
			// Open file for logging
			file, err := os.OpenFile(
				config.FilePath,
				os.O_CREATE|os.O_WRONLY|os.O_APPEND,
				0600, // Secure permissions: owner read/write only
			)
			if err != nil {
				return zerolog.Logger{}, nil, err
			}

			fileOut = file
			cleanupFuncs = append(cleanupFuncs, func() {
				_ = file.Close()
			})
		}

		if fileOut != nil {
			filteredFile := FilteredWriter{
				Writer:   fileOut,
				MinLevel: config.FileLevel,
			}

			writers = append(writers, filteredFile)
		}
	}

	// Create multi-writer
	multi := zerolog.MultiLevelWriter(writers...)

	// Create logger with timestamp
	logger := zerolog.New(multi).With().Timestamp().Logger()

	// Cleanup function
	cleanup := func() {
		for _, fn := range cleanupFuncs {
			fn()
		}
	}

	return logger, cleanup, nil
}

// InitDualLoggerGlobal is a convenience function that initializes the global logger
// with dual output configuration.
//
// Example:
//
//	cleanup, err := InitDualLoggerGlobal(DualLoggerConfig{...})
//	if err != nil {
//	    log.Fatal().Err(err).Msg("Failed to initialize logger")
//	}
//	defer cleanup()
//
//	// Now use the global logger
//	log.Info().Msg("This goes to console")
//	log.Debug().Msg("This goes to file only")
func InitDualLoggerGlobal(config DualLoggerConfig) (func(), error) {
	logger, cleanup, err := InitDualLogger(config)
	if err != nil {
		return nil, err
	}

	// Set global logger
	log.Logger = logger

	return cleanup, nil
}
