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
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	log.Logger = zerolog.New(output).With().
		Timestamp().
		Caller().
		Logger()
}

// GetLogger returns a new logger instance
func GetLogger() zerolog.Logger {
	return log.Logger
}

// SetLogOutput sets the output destination for the logger
func SetLogOutput(w io.Writer) {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        w,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	})
}

// LogError logs an error with additional context
func LogError(err error, message string, fields map[string]interface{}) {
	event := log.Error().Err(err).Str("message", message)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Send()
}
