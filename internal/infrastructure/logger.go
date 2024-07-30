package infrastructure

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger
func InitLogger(logLevel string) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
}

// GetLogger returns a new logger instance
func GetLogger() zerolog.Logger {
	return log.Logger
}

// SetLogOutput sets the output destination for the logger
func SetLogOutput(w io.Writer) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339})
}
