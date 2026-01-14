package utility

import (
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// GetCurrentPath retrieves the current PATH environment variable
func GetCurrentPath() string {
	return os.Getenv("PATH")
}

// AddToPath adds a directory to the system PATH
func AddToPath(dir string) error {
	// Get current PATH
	currentPath := GetCurrentPath()
	
	// Check if directory is already in PATH
	paths := strings.Split(currentPath, ";")
	for _, path := range paths {
		if strings.EqualFold(strings.TrimSpace(path), dir) {
			return nil // Already in PATH
		}
	}

	// Add directory to PATH
	newPath := currentPath + ";" + dir
	
	// Update system environment variable
	k, err := registry.OpenKey(registry.CURRENT_USER, "Environment", registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		// Try with admin privileges
		k, err = registry.OpenKey(registry.LOCAL_MACHINE, "SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment", registry.SET_VALUE|registry.QUERY_VALUE)
		if err != nil {
			return err
		}
	}
	defer k.Close()

	return k.SetStringValue("PATH", newPath)
}

// SetEnvironmentVariable sets a system environment variable
func SetEnvironmentVariable(name, value string) error {
	// Update current process environment
	err := os.Setenv(name, value)
	if err != nil {
		return err
	}

	// Update system environment variable
	k, err := registry.OpenKey(registry.CURRENT_USER, "Environment", registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		// Try with admin privileges
		k, err = registry.OpenKey(registry.LOCAL_MACHINE, "SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment", registry.SET_VALUE|registry.QUERY_VALUE)
		if err != nil {
			return err
		}
	}
	defer k.Close()

	return k.SetStringValue(name, value)
}

// GetEnvironmentVariable gets a system environment variable
func GetEnvironmentVariable(name string) (string, error) {
	// Try current process first
	value := os.Getenv(name)
	if value != "" {
		return value, nil
	}

	// Try registry
	k, err := registry.OpenKey(registry.CURRENT_USER, "Environment", registry.QUERY_VALUE)
	if err != nil {
		k, err = registry.OpenKey(registry.LOCAL_MACHINE, "SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment", registry.QUERY_VALUE)
		if err != nil {
			return "", err
		}
	}
	defer k.Close()

	regValue, _, regErr := k.GetStringValue(name)
	return regValue, regErr
}

// RemoveFromPath removes a directory from the system PATH
func RemoveFromPath(dir string) error {
	// Get current PATH
	currentPath := GetCurrentPath()
	
	// Split into paths
	paths := strings.Split(currentPath, ";")
	
	// Filter out the directory
	newPaths := make([]string, 0)
	for _, path := range paths {
		if !strings.EqualFold(strings.TrimSpace(path), dir) {
			newPaths = append(newPaths, path)
		}
	}

	// Join back into new PATH
	newPath := strings.Join(newPaths, ";")
	
	// Update system environment variable
	k, err := registry.OpenKey(registry.CURRENT_USER, "Environment", registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		// Try with admin privileges
		k, err = registry.OpenKey(registry.LOCAL_MACHINE, "SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment", registry.SET_VALUE|registry.QUERY_VALUE)
		if err != nil {
			return err
		}
	}
	defer k.Close()

	return k.SetStringValue("PATH", newPath)
}
