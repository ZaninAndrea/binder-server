package log

import (
	"os"
	"path/filepath"
	"time"

	"github.com/jwalton/go-supportscolor"
	"github.com/rs/zerolog"
)

var defaultLogger Logger

func Default() Logger {
	noColor := !supportscolor.Stderr().SupportsColor

	output := ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		NoColor:    noColor,
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			ServiceFieldName,
			zerolog.CallerFieldName,
			zerolog.MessageFieldName,
		},
	}

	// Print caller path relative to current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	divider := colorize(" |", colorCyan, noColor)
	output.FormatCaller = func(i interface{}) string {
		var c string
		if cc, ok := i.(string); ok {
			c = cc
		}
		if len(c) > 0 {
			if rel, err := filepath.Rel(cwd, c); err == nil {
				c = rel
			}
			c = c + divider
		}
		return c
	}

	// Get visible log levels from environment
	var loggerLevel zerolog.Level
	switch os.Getenv("LOGGER_LEVEL") {
	case "TRACE":
		loggerLevel = zerolog.TraceLevel
	case "DEBUG":
		loggerLevel = zerolog.DebugLevel
	case "INFO":
		loggerLevel = zerolog.InfoLevel
	case "WARN":
		loggerLevel = zerolog.WarnLevel
	case "ERROR":
		loggerLevel = zerolog.ErrorLevel
	case "FATAL":
		loggerLevel = zerolog.FatalLevel
	case "PANIC":
		loggerLevel = zerolog.PanicLevel
	default:
		loggerLevel = zerolog.DebugLevel
	}

	logger := zerolog.
		New(output).
		With().
		Timestamp().
		CallerWithSkipFrameCount(3).
		Logger().
		Level(loggerLevel)

	return Logger{
		logger: logger,
	}
}
