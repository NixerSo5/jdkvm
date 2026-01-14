package java

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"jdkvm/file"
)

/**
 * Returns version, architecture
 */
func GetCurrentVersion() (string, string) {
	cmd := exec.Command("java", "-version")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		output := stderr.String()
		// Extract version from output like: java version "1.8.0_301"
		re := regexp.MustCompile(`version "([0-9]+\.[0-9]+\.[0-9]+.*)"`)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			v := matches[1]
			// Extract major version for newer Java versions (Java 9+)
			if strings.HasPrefix(v, "1.") {
				// Old style (Java 8 and below): 1.8.0_301
				v = v[2:]
			}
			// Get architecture
			cmd := exec.Command("java", "-XshowSettings:properties", "-version")
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Run()
			output := stderr.String()
			reArch := regexp.MustCompile(`os\.arch=([a-zA-Z0-9_]+)`)
			archMatches := reArch.FindStringSubmatch(output)
			var bit string
			if len(archMatches) > 1 {
				if archMatches[1] == "amd64" || archMatches[1] == "x86_64" {
					bit = "64"
				} else if archMatches[1] == "arm64" {
					bit = "arm64"
				} else {
					bit = "32"
				}
			} else {
				bit = "Unknown"
			}
			return v, bit
		}
	}
	return "Unknown", ""
}

func IsVersionInstalled(root string, version string, cpu string) bool {
	// First check if exact version exists
	versionDir := filepath.Join(root, "v"+version)
	if file.Exists(versionDir) {
		javaExe := filepath.Join(versionDir, "bin", "java.exe")
		if file.Exists(javaExe) {
			return true
		}
	}

	// If exact version not found, check if it's a major version (like 17, 11, 8)
	// and find any installed version that starts with this major version
	if !strings.Contains(version, ".") {
		files, _ := os.ReadDir(root)
		for _, f := range files {
			if f.IsDir() && strings.HasPrefix(f.Name(), "v"+version+".") {
				versionDir := filepath.Join(root, f.Name())
				javaExe := filepath.Join(versionDir, "bin", "java.exe")
				if file.Exists(javaExe) {
					return true
				}
			}
		}
	}

	return false
}

func GetInstalled(root string) []string {
	list := make([]string, 0)
	files, _ := os.ReadDir(root)

	for i := len(files) - 1; i >= 0; i-- {
		if files[i].IsDir() {
			name := files[i].Name()
			// Check if the directory name starts with "v"
			if strings.HasPrefix(name, "v") {
				// Remove the "v" prefix to get the version string
				versionStr := strings.TrimPrefix(name, "v")
				list = append(list, versionStr)
			}
		}
	}

	// Sort versions in descending order (simple string sort for now)
	sort.Slice(list, func(i, j int) bool {
		return list[i] > list[j]
	})

	return list
}

// Simplified version - in a real implementation this would fetch from remote
func GetAvailable() ([]string, []string) {
	// Return some example versions
	return []string{"17.0.11", "17.0.10", "11.0.23", "11.0.22", "8.0.412", "8.0.402"}, []string{"17.0.11", "11.0.23", "8.0.412"}
}
