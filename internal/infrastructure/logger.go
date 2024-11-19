package infrastructure

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var loggerMu sync.Mutex

// InitLogger initializes the global logger.
func InitLogger(logLevel string) error {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(level)
	SetLogOutput(os.Stdout)

	return nil
}

// GetLogger returns the global logger instance.
func GetLogger() zerolog.Logger {
	return log.Logger
}

// SetLogOutput sets the output destination for the logger.
func SetLogOutput(w io.Writer) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	output := zerolog.ConsoleWriter{
		Out:        w,
		TimeFormat: time.RFC3339,
		NoColor:    true,
	}
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
}

// SetLogger sets the global logger (for backwards compatibility).
func SetLogger(l *zerolog.Logger) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	log.Logger = *l
}
