package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"
)

var (
	logFile     *os.File
	logger      *log.Logger
	once        sync.Once
	initialized bool
)

// LogLevel represents the severity of the log message.
type LogLevel int

const (
	INFO LogLevel = iota
	WARNING
	ERROR
	CRITICAL
	DEBUG
)

func (l LogLevel) String() string {
	return [...]string{"INFO", "WARNING", "ERROR", "CRITICAL", "DEBUG"}[l]
}

// Init initializes the logger.
func Init(workingDir string) error {
	var initErr error
	once.Do(func() {
		logDir, err := getOrCreateLogDir(workingDir)
		if err != nil {
			initErr = fmt.Errorf("failed to get or create log directory: %v", err)
			return
		}

		logPath := filepath.Join(logDir, "netutil.log")
		logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %v", err)
			return
		}

		logger = log.New(logFile, "", log.Ldate|log.Ltime)
		initialized = true

		// Add an initial log entry
		logger.Printf("[INFO] Logging initialized at %s", time.Now().Format("2006-01-02 15:04:05"))
	})
	return initErr
}

// getOrCreateLogDir retrieves or creates the log directory.
func getOrCreateLogDir(workingDir string) (string, error) {
	entries, err := os.ReadDir(workingDir)
	if err != nil {
		return "", err
	}

	logDirRegex := regexp.MustCompile(`(?i)log`)
	for _, entry := range entries {
		if entry.IsDir() && logDirRegex.MatchString(entry.Name()) {
			return filepath.Join(workingDir, entry.Name()), nil
		}
	}

	// If no log directory found in working directory, create one
	newLogDir := filepath.Join(workingDir, "logs")
	err = os.MkdirAll(newLogDir, 0755)
	if err != nil {
		return "", err
	}

	return newLogDir, nil
}

// Log logs a message with the specified severity level.
func Log(level LogLevel, format string, v ...interface{}) {
	if !initialized {
		fmt.Fprintf(os.Stderr, "[%s] %s\n", level.String(), fmt.Sprintf(format, v...))
		return
	}
	_, file, line, _ := runtime.Caller(1)
	msg := fmt.Sprintf(format, v...)
	logger.Printf("[%s] %s:%d: %s", level.String(), filepath.Base(file), line, msg)
}

// Info logs an informational message.
func Info(format string, v ...interface{}) {
	Log(INFO, format, v...)
}

// Warning logs a warning message.
func Warning(format string, v ...interface{}) {
	Log(WARNING, format, v...)
}

// Error logs an error message.
func Error(format string, v ...interface{}) {
	Log(ERROR, format, v...)
}

// Critical logs a critical error message.
func Critical(format string, v ...interface{}) {
	Log(CRITICAL, format, v...)
}

// Debug logs a debug message.
func Debug(format string, v ...interface{}) {
	Log(DEBUG, format, v...)
}

// Close closes the log file.
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
