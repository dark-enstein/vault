package vlog

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"strings"
	"time"
)

type Logger struct {
	l *zerolog.Logger
}

func New(debug bool) *Logger {
	level := resolveLogger(debug)
	zerolog.SetGlobalLevel(level)
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("  %s  ", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%s", i))
	}
	log := zerolog.New(output).With().Timestamp().Logger()
	return &Logger{l: &log}
}

func resolveLogger(debug bool) zerolog.Level {
	if debug {
		return zerolog.DebugLevel
	}
	return zerolog.InfoLevel
}

func (l *Logger) Logger() *zerolog.Logger {
	return l.l
}

// Printf implements the redis.internal.Logger interface so it can be used
func (l *Logger) Printf(ctx context.Context, format string, v ...interface{}) {
	l.Logger().Printf(format, v)
	return
}
