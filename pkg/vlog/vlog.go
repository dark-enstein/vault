// Package vlog offers a structured, pluggable logging interface built upon rs/zerolog,
// designed for simplicity, flexibility, and contextual logging. It abstracts logging
// complexity and provides customizable log level management and output formatting.
// The package provides seamless integration with standard libraries and supports
// context-based logging, making it suitable for applications ranging from simple
// command-line tools to large-scale microservices.

package vlog

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"strings"
	"time"
)

// Logger is a lightweight wrapper around zerolog.Logger, facilitating
// advanced logging capabilities such as dynamic log level setting and
// custom log formatting. It encapsulates the zerolog.Logger to leverage
// its high performance and concurrency-safe logging features while
// providing a simplified and intuitive API for application logging.
type Logger struct {
	l *zerolog.Logger // embedded zerolog Logger
}

// New creates and configures a new Logger instance. It takes a boolean flag 'debug'
// to set the appropriate logging level. The logger is set up with a console writer
// that formats output to a human-friendly, readable format. It is the primary
// entry point for creating a logger with the vlog package.
func New(debug bool) *Logger {
	// Set the appropriate logging level based on the debug flag.
	level := resolveLogger(debug)
	// Apply the level globally to ensure consistency across the application.
	zerolog.SetGlobalLevel(level)

	// Configure the output format for log messages, with specific formatting
	// for levels, messages, and fields to enhance readability and scanning.
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		// Format log levels to uppercase.
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		// Format messages for readability.
		return fmt.Sprintf("  %s  ", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		// Append a colon to field names.
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		// Format field values to uppercase.
		return strings.ToUpper(fmt.Sprintf("%s", i))
	}

	// Initialize the zerolog logger with the configured output and decorators.
	log := zerolog.New(output).With().Timestamp().Logger()

	return &Logger{l: &log} // Return the configured Logger instance
}

// resolveLogger determines the zerolog.Level based on the provided debug flag.
// It serves as a helper function to map the boolean 'debug' flag to the corresponding
// zerolog level, abstracting away the zerolog specifics from the package consumer.
func resolveLogger(debug bool) zerolog.Level {
	if debug {
		return zerolog.DebugLevel
	}
	return zerolog.InfoLevel
}

// Logger retrieves the underlying zerolog.Logger, providing direct access
// to its advanced functionality if required. This allows users of vlog to
// utilize zerolog's full capabilities while defaulting to simplified use.
func (l *Logger) Logger() *zerolog.Logger {
	return l.l
}

// Printf provides a formatted output function that adheres to the Logger interface
// required by various packages like database/sql. It allows vlog to be easily plugged
// into other libraries or frameworks that accept standard loggers. Contextual information
// from the context.Context parameter can be utilized here for more contextual logging.
func (l *Logger) Printf(ctx context.Context, format string, v ...interface{}) {
	// The Printf method is a simple passthrough to zerolog's Printf method,
	// maintaining compatibility with interfaces expecting a Printf method.
	l.Logger().Printf(format, v...)
}
