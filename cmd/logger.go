package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ConfigureLogger sets globals and returns a logger
func ConfigureLogger() zerolog.Logger {
	// set default time
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// default info level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// default to human readable log output
	if !jsonFormat {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		output.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		}
		return zerolog.New(output).With().Timestamp().Logger()
	}
	return log.Logger
}
