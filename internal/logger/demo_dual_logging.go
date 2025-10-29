// internal/logger/demo_dual_logging.go
//
// Demo program to showcase dual logging functionality.
// Run with: go run internal/logger/demo_dual_logging.go
//
//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"fmt"

	"github.com/peiman/ckeletin-go/internal/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	fmt.Println("=== Dual Logging System Demo ===\n")

	// Create buffers to capture output for demonstration
	var consoleBuf bytes.Buffer
	var fileBuf bytes.Buffer

	// Configure dual logger
	config := logger.DualLoggerConfig{
		ConsoleEnabled: true,
		ConsoleLevel:   zerolog.InfoLevel,
		ConsoleColor:   false, // Disable color for clean demo output
		ConsoleWriter:  &consoleBuf,

		FileEnabled: true,
		FileLevel:   zerolog.DebugLevel,
		FileWriter:  &fileBuf,
	}

	// Initialize dual logger
	dualLogger, cleanup, err := logger.InitDualLogger(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize dual logger")
	}
	defer cleanup()

	// Set as global logger for demo
	log.Logger = dualLogger

	fmt.Println("Logging messages at different levels...\n")

	// Write logs at different levels
	log.Trace().Msg("TRACE: This is a trace message (very verbose)")
	log.Debug().Msg("DEBUG: This is a debug message (detailed info)")
	log.Info().Msg("INFO: Application started successfully")
	log.Info().Str("user", "admin").Int("port", 8080).Msg("INFO: Server configuration")
	log.Warn().Msg("WARN: Configuration file not found, using defaults")
	log.Error().Str("error", "connection timeout").Msg("ERROR: Failed to connect to database")

	// Demonstrate structured logging with context
	log.Info().
		Str("module", "auth").
		Str("action", "login").
		Bool("success", true).
		Msg("User authentication completed")

	log.Debug().
		Str("module", "auth").
		Str("function", "validateToken").
		Int("token_length", 256).
		Msg("Token validation details")

	// Show what was written to console (INFO and above)
	fmt.Println("--- CONSOLE OUTPUT (INFO+ level) ---")
	fmt.Println(consoleBuf.String())

	// Show what was written to file (DEBUG and above)
	fmt.Println("\n--- FILE OUTPUT (DEBUG+ level) ---")
	fmt.Println(fileBuf.String())

	// Demonstrate the filtering
	fmt.Println("\n--- ANALYSIS ---")
	consoleLines := bytes.Count(consoleBuf.Bytes(), []byte("\n"))
	fileLines := bytes.Count(fileBuf.Bytes(), []byte("\n"))

	fmt.Printf("Console messages: %d (INFO, WARN, ERROR only)\n", consoleLines)
	fmt.Printf("File messages: %d (DEBUG, INFO, WARN, ERROR)\n", fileLines)
	fmt.Printf("Filtered messages: %d (TRACE filtered from both, DEBUG filtered from console)\n", fileLines-consoleLines)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("\nKey observations:")
	fmt.Println("1. Console shows user-friendly INFO+ messages with timestamps")
	fmt.Println("2. File captures detailed DEBUG+ messages in JSON format")
	fmt.Println("3. TRACE messages are filtered from both outputs")
	fmt.Println("4. DEBUG messages appear in file but not console")
	fmt.Println("5. Zero allocations for filtered messages (see benchmarks)")
}
