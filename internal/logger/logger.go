package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

// Init initializes the global logger
func Init(logPath string, level string, context map[string]interface{}) error {
	// Handle home directory expansion if necessary
	if strings.HasPrefix(logPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home dir: %w", err)
		}
		logPath = filepath.Join(home, logPath[2:])
	}

	// Ensure directory exists
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	var zLevel zerolog.Level
	switch strings.ToLower(level) {
	case "debug":
		zLevel = zerolog.DebugLevel
	case "info":
		zLevel = zerolog.InfoLevel
	case "warn":
		zLevel = zerolog.WarnLevel
	case "error":
		zLevel = zerolog.ErrorLevel
	default:
		zLevel = zerolog.InfoLevel
	}

	c := zerolog.New(f).Level(zLevel).With().Timestamp()
	if context != nil {
		for k, v := range context {
			c = c.Interface(k, v)
		}
	}
	log = c.Logger()
	return nil
}

// Debug logs a debug message
func Debug(msg string, fields map[string]interface{}) {
	e := log.Debug()
	addFields(e, fields)
	e.Msg(msg)
}

// Info logs an info message
func Info(msg string, fields map[string]interface{}) {
	e := log.Info()
	addFields(e, fields)
	e.Msg(msg)
}

// Warn logs a warning message
func Warn(msg string, fields map[string]interface{}) {
	e := log.Warn()
	addFields(e, fields)
	e.Msg(msg)
}

// Error logs an error message
func Error(msg string, err error, fields map[string]interface{}) {
	e := log.Error()
	if err != nil {
		e = e.Err(err)
	}
	addFields(e, fields)
	e.Msg(msg)
}

func addFields(e *zerolog.Event, fields map[string]interface{}) {
	if fields == nil {
		return
	}
	for k, v := range fields {
		e.Interface(k, v)
	}
}

// SetOutput changes the logger output (useful for testing)
func SetOutput(w io.Writer) {
	log = log.Output(w)
}
