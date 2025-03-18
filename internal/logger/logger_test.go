package logger

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "UNKNOWN"},
	}

	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("Expected %s for level %d, got %s", test.expected, test.level, test.level.String())
		}
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"WARN", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"invalid", LevelInfo}, // Default to info for invalid levels
	}

	for _, test := range tests {
		if ParseLogLevel(test.input) != test.expected {
			t.Errorf("Expected level %d for input %s, got %d", test.expected, test.input, ParseLogLevel(test.input))
		}
	}
}

func TestNewLogger(t *testing.T) {
	// Test with stdout
	logger, err := NewLogger("info", "stdout")
	if err != nil {
		t.Errorf("Failed to create logger: %v", err)
	}
	if logger.level != LevelInfo {
		t.Errorf("Expected level to be INFO, got %s", logger.level.String())
	}

	// Test with file
	tempFile, err := os.CreateTemp("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	logger, err = NewLogger("debug", tempFile.Name())
	if err != nil {
		t.Errorf("Failed to create logger: %v", err)
	}
	if logger.level != LevelDebug {
		t.Errorf("Expected level to be DEBUG, got %s", logger.level.String())
	}
}

func TestLoggerLevels(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := &Logger{
		level:  LevelInfo,
		logger: log.New(&buf, "", 0),
	}

	// Test debug level (should be filtered out)
	logger.Debug("Debug message")
	if buf.Len() > 0 {
		t.Errorf("Expected debug message to be filtered out, got: %s", buf.String())
	}

	// Test info level
	buf.Reset()
	logger.Info("Info message")
	if !strings.Contains(buf.String(), "INFO") || !strings.Contains(buf.String(), "Info message") {
		t.Errorf("Expected info message to be logged, got: %s", buf.String())
	}

	// Test warn level
	buf.Reset()
	logger.Warn("Warn message")
	if !strings.Contains(buf.String(), "WARN") || !strings.Contains(buf.String(), "Warn message") {
		t.Errorf("Expected warn message to be logged, got: %s", buf.String())
	}

	// Test error level
	buf.Reset()
	logger.Error("Error message")
	if !strings.Contains(buf.String(), "ERROR") || !strings.Contains(buf.String(), "Error message") {
		t.Errorf("Expected error message to be logged, got: %s", buf.String())
	}
}

func TestGlobalLogger(t *testing.T) {
	// Create a pipe to capture global logger output
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	// Initialize global logger with the pipe
	oldLogger := globalLogger
	defer func() { globalLogger = oldLogger }()

	globalLogger = &Logger{
		level:  LevelDebug,
		logger: log.New(w, "", 0),
	}

	// Start a goroutine to read from the pipe
	done := make(chan bool)
	var output bytes.Buffer
	go func() {
		io.Copy(&output, r)
		done <- true
	}()

	// Test global logger functions
	Debug("Debug message")
	Info("Info message")
	Warn("Warn message")
	Error("Error message")

	// Close the writer to signal EOF to the reader
	w.Close()
	<-done

	// Check output
	outputStr := output.String()
	if !strings.Contains(outputStr, "DEBUG") || !strings.Contains(outputStr, "Debug message") {
		t.Errorf("Expected debug message to be logged, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "INFO") || !strings.Contains(outputStr, "Info message") {
		t.Errorf("Expected info message to be logged, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "WARN") || !strings.Contains(outputStr, "Warn message") {
		t.Errorf("Expected warn message to be logged, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "ERROR") || !strings.Contains(outputStr, "Error message") {
		t.Errorf("Expected error message to be logged, got: %s", outputStr)
	}
}

func TestInitLogger(t *testing.T) {
	// Save the original global logger
	oldLogger := globalLogger
	defer func() { globalLogger = oldLogger }()

	// Initialize logger with stdout
	err := InitLogger("info", "stdout")
	if err != nil {
		t.Errorf("Failed to initialize logger: %v", err)
	}
	if globalLogger == nil {
		t.Errorf("Expected global logger to be initialized")
	}
	if globalLogger.level != LevelInfo {
		t.Errorf("Expected level to be INFO, got %s", globalLogger.level.String())
	}

	// Initialize logger with invalid file
	err = InitLogger("info", "/nonexistent/directory/file.log")
	if err == nil {
		t.Errorf("Expected error when initializing logger with invalid file")
	}
}