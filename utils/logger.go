package utils

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logLevel := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(logLevel) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info", "":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "disabled":
		zerolog.SetGlobalLevel(zerolog.Disabled)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		log.Debug().Msgf("Invalid log level '%s' set", logLevel)
	}
}

func NewLogger(strings map[string]string) zerolog.Logger {
	return WithStrings(log.With().Logger(), strings)
}

func WithStrings(logger zerolog.Logger, strings map[string]string) zerolog.Logger {
	ctx := logger.With()
	for k, v := range strings {
		ctx = ctx.Str(k, v)
	}
	return ctx.Logger()
}

func LogEnv() {
	log.Trace().Strs("environment", os.Environ()).Msg("Environment Variables")
}
