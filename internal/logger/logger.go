// internal/logger/logger.go
package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Init initializes the logger
func Init() {
	levelStr := viper.GetString("app.log_level")
	if levelStr == "" {
		levelStr = "info"
	}
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}
