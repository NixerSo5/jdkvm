package arch

import (
	"os"
	"strings"
)

func Validate(str string) string {
	if str == "" {
		str = strings.ToLower(os.Getenv("PROCESSOR_ARCHITECTURE"))
	}
	if strings.Contains(str, "arm64") {
		return "arm64"
	}
	if strings.Contains(str, "64") {
		return "64"
	}
	return "32"
}

// Simplified architecture detection - in a real implementation this would be more robust
func Bit(path string) string {
	// For Java, we can check the file extension or use more complex detection
	// This is a simplified version
	if strings.Contains(strings.ToLower(path), "arm64") {
		return "arm64"
	} else if strings.Contains(strings.ToLower(path), "64") {
		return "64"
	} else if strings.Contains(strings.ToLower(path), "32") {
		return "32"
	}
	return "?"
}
