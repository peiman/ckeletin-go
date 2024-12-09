package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Init initializes the logger with options from Viper.
// Call this once in rootCmd's PersistentPreRunE or main initialization.
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

	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}).
		With().
		Timestamp().
		Logger()

	return nil
}
