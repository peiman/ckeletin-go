// internal/logger/logger.go
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Init initializes the logger with the given output writer.
// If out is nil, it defaults to os.Stderr.
func Init(out io.Writer) error {
	if out == nil {
		out = os.Stderr
	}

	logLevelStr := viper.GetString("app.log_level")
	level, err := zerolog.ParseLevel(logLevelStr)
	if err != nil {
		level = zerolog.InfoLevel
		log.Warn().
			Err(err).
			Str("provided_level", logLevelStr).
			Msg("Invalid log level provided, defaulting to 'info'")
	}
	zerolog.SetGlobalLevel(level)

	// Configure the logger to write to 'out' and set time format
	log.Logger = zerolog.New(out).
		With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{
			Out:        out,
			TimeFormat: time.RFC3339,
		})

	return nil
}
