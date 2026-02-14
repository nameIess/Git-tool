package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents a log severity level.
type Level string

const (
	LevelInfo    Level = "INFO"
	LevelWarn    Level = "WARN"
	LevelError   Level = "ERROR"
	LevelDebug   Level = "DEBUG"
)

// Logger provides structured logging to a file.
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	filePath string
}

var defaultLogger *Logger

// Init creates a new log file in dir with a timestamped name.
// It also logs system information (OS, Git version, SSH version).
func Init(dir string) (*Logger, error) {
	ts := time.Now().Format("2006-01-02_15-04-05")
	name := fmt.Sprintf("git-setup-%s.log", ts)
	path := filepath.Join(dir, name)

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	l := &Logger{file: f, filePath: path}
	defaultLogger = l

	l.Info("Log file created: %s", path)
	return l, nil
}

// FilePath returns the path of the log file.
func (l *Logger) FilePath() string {
	return l.filePath
}

// Close closes the log file.
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

func (l *Logger) write(level Level, msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	ts := time.Now().Format("2006-01-02 15:04:05.000")
	formatted := fmt.Sprintf(msg, args...)
	line := fmt.Sprintf("[%s] [%s] %s\n", ts, level, formatted)

	if l.file != nil {
		l.file.WriteString(line)
	}
}

// Info logs an informational message.
func (l *Logger) Info(msg string, args ...any) {
	l.write(LevelInfo, msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...any) {
	l.write(LevelWarn, msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...any) {
	l.write(LevelError, msg, args...)
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...any) {
	l.write(LevelDebug, msg, args...)
}

// Default returns the default logger instance.
func Default() *Logger {
	return defaultLogger
}

// Info logs using the default logger.
func Info(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, args...)
	}
}

// Warn logs using the default logger.
func Warn(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, args...)
	}
}

// Error logs using the default logger.
func Error(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, args...)
	}
}

// Debug logs using the default logger.
func Debug(msg string, args ...any) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, args...)
	}
}
