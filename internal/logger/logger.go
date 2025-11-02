package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	debugEnabled bool
	logFile      *os.File
	logger       *log.Logger
)

// Init initializes the logger with the specified debug flag
// Creates log directory and file if they don't exist
// Overwrites existing log file on each start
func Init(debug bool) error {
	debugEnabled = debug

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create log directory path
	logDir := filepath.Join(homeDir, ".local", "share", "smpe_ls")

	// Create directories if they don't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create/overwrite log file
	logPath := filepath.Join(logDir, "smpe_ls.log")
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	// Initialize logger
	logger = log.New(logFile, "", log.LstdFlags)

	Info("smpe_ls started")
	if debugEnabled {
		Info("Debug mode enabled")
	}

	return nil
}

// Close closes the log file
func Close() error {
	if logFile != nil {
		Info("smpe_ls shutting down")
		return logFile.Close()
	}
	return nil
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if logger != nil {
		msg := fmt.Sprintf(format, v...)
		logger.Printf("[INFO] %s", msg)
	}
}

// Debug logs a debug message (only if debug is enabled)
func Debug(format string, v ...interface{}) {
	if debugEnabled && logger != nil {
		msg := fmt.Sprintf(format, v...)
		logger.Printf("[DEBUG] %s", msg)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if logger != nil {
		msg := fmt.Sprintf(format, v...)
		logger.Printf("[ERROR] %s", msg)
	}
}

// Fatal logs a fatal message and exits
func Fatal(format string, v ...interface{}) {
	if logger != nil {
		msg := fmt.Sprintf(format, v...)
		logger.Printf("[FATAL] %s", msg)
	}
	os.Exit(1)
}

// GetLogPath returns the path to the log file
func GetLogPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".local", "share", "smpe_ls", "smpe_ls.log")
}
