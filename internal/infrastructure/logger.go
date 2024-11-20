package infrastructure

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var loggerMu sync.Mutex

// InitLogger initializes the global logger.
func InitLogger(level string) error {
	parsedLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", level, err)
	}

	zerolog.SetGlobalLevel(parsedLevel)
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
