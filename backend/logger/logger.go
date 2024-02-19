package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	Log *zerolog.Logger
}

func NewLogger(env string) *zerolog.Logger {
	if env == "development" {
		logger := log.Output(zerolog.ConsoleWriter{
			Out: os.Stdout,
		})

		logger.With().Timestamp()

		return &logger
	}
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	return &logger
}

func (log *Logger) InfoLog(jsonEncoded []byte) {
	log.Log.Info().RawJSON("value", jsonEncoded)
}
