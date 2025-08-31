package infrastructure

import (
	"fmt"
	"log"
	"time"
)

// LogLevel represents different log levels
type LogLevel string

const (
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarning LogLevel = "WARNING"
	LogLevelError   LogLevel = "ERROR"
	LogLevelSuccess LogLevel = "SUCCESS"
	LogLevelDebug   LogLevel = "DEBUG"
)

// Logger provides structured logging with timestamps and levels
type Logger struct {
	prefix string
}

// NewLogger creates a new logger with optional prefix
func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}

// formatMessage formats a log message with timestamp and level
func (l *Logger) formatMessage(level LogLevel, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if l.prefix != "" {
		return fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level, l.prefix, message)
	}
	return fmt.Sprintf("[%s] [%s] %s", timestamp, level, message)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Print(l.formatMessage(LogLevelInfo, message))
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Print(l.formatMessage(LogLevelWarning, message))
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Print(l.formatMessage(LogLevelError, message))
}

// Success logs a success message
func (l *Logger) Success(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Print(l.formatMessage(LogLevelSuccess, message))
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Print(l.formatMessage(LogLevelDebug, message))
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Fatal(l.formatMessage(LogLevelError, message))
}

// Global logger instances
var (
	ServerLogger = NewLogger("SERVER")
	AgentLogger  = NewLogger("AGENT")
	DefaultLogger = NewLogger("")
)
