package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// Level represents the severity level of a log message
type Level int

const (
	// LevelDebug for detailed debugging information
	LevelDebug Level = iota
	// LevelInfo for general operational information
	LevelInfo
	// LevelWarn for warning messages
	LevelWarn
	// LevelError for error messages
	LevelError
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger is a simple logger that supports different log levels
type Logger struct {
	level  Level
	logger *log.Logger
}

// NewLogger creates a new logger with the specified level and output
func NewLogger(level string, output string) (*Logger, error) {
	var writer io.Writer
	
	if output == "stdout" {
		writer = os.Stdout
	} else {
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		writer = file
	}
	
	return &Logger{
		level:  ParseLogLevel(level),
		logger: log.New(writer, "", 0),
	}, nil
}

// log logs a message with the specified level
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s: %s", timestamp, level.String(), message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Global logger instance
var globalLogger *Logger

// InitLogger initializes the global logger
func InitLogger(level string, output string) error {
	logger, err := NewLogger(level, output)
	if err != nil {
		return err
	}
	
	globalLogger = logger
	return nil
}

// GlobalDebug logs a debug message to the global logger
func GlobalDebug(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(format, args...)
	}
}

// GlobalInfo logs an info message to the global logger
func GlobalInfo(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(format, args...)
	}
}

// GlobalWarn logs a warning message to the global logger
func GlobalWarn(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(format, args...)
	}
}

// GlobalError logs an error message to the global logger
func GlobalError(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(format, args...)
	}
}

// Convenience aliases for global logger functions
var (
	Debug = GlobalDebug
	Info  = GlobalInfo
	Warn  = GlobalWarn
	Error = GlobalError
)