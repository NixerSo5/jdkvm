package utility

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"jdkvm/file"
)

var debugMode = false

func EnableDebugLogs() {
	debugMode = true
}

func DisableDebugLogs() {
	debugMode = false
}

func DebugLog(message string) {
	if debugMode {
		writeToLog("DEBUG", message)
	}
}

func DebugLogf(format string, args ...interface{}) {
	if debugMode {
		message := fmt.Sprintf(format, args...)
		writeToLog("DEBUG", message)
	}
}

func InfoLog(message string) {
	writeToLog("INFO", message)
}

func InfoLogf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writeToLog("INFO", message)
}

func ErrorLog(message string) {
	writeToLog("ERROR", message)
}

func ErrorLogf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writeToLog("ERROR", message)
}

func writeToLog(level string, message string) {
	logFile, err := os.OpenFile("jdkvm.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
		return
	}
	defer logFile.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("%s [%s] %s\n", timestamp, level, message)
	logFile.WriteString(logEntry)
}

func CleanVersion(version string) string {
	// Remove any leading 'v' or other non-numeric characters
	version = strings.TrimPrefix(version, "v")
	// Remove any build info or other suffixes
	if idx := strings.Index(version, "-"); idx != -1 {
		version = version[:idx]
	}
	if idx := strings.Index(version, "."); idx == -1 {
		// Single digit version, add .0.0
		version = version + ".0.0"
	} else if strings.Count(version, ".") == 1 {
		// Two digit version, add .0
		version = version + ".0"
	}
	return version
}

func Rename(oldPath, newPath string) error {
	// Try a simple rename first
	err := os.Rename(oldPath, newPath)
	if err == nil {
		return nil
	}

	// If rename fails, try copy and delete
	if file.IsDir(oldPath) {
		err = file.CopyDir(oldPath, newPath)
	} else {
		err = file.CopyFile(oldPath, newPath)
	}
	if err != nil {
		return err
	}

	// Delete the original
	if file.IsDir(oldPath) {
		return os.RemoveAll(oldPath)
	} else {
		return os.Remove(oldPath)
	}
}

func GetExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}
