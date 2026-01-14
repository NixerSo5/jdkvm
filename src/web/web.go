package web

import (
	"archive/zip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"jdkvm/file"
)

var client = &http.Client{}
var javaBaseAddress = "https://github.com/adoptium/temurin/releases/download/"

// JavaVersionInfo contains information about a specific Java version
type JavaVersionInfo struct {
	Latest string `json:"latest"`
	URL    string `json:"url"`
	Short  string `json:"short"`
}

// JavaVersionMapping maps major version numbers to version information
var JavaVersionMapping map[string]JavaVersionInfo

// LoadVersionMapping loads the version mapping from the JSON file
func LoadVersionMapping() error {
	// Define all possible paths to check
	pathsToCheck := []string{}
	
	// Try the current directory first
	pathsToCheck = append(pathsToCheck, "version_mapping.json")
	
	// Try web subdirectory
	pathsToCheck = append(pathsToCheck, filepath.Join("web", "version_mapping.json"))
	
	// Try parent directory
	pathsToCheck = append(pathsToCheck, filepath.Join("..", "web", "version_mapping.json"))
	
	// Try relative to executable
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		pathsToCheck = append(pathsToCheck, filepath.Join(exeDir, "version_mapping.json"))
		pathsToCheck = append(pathsToCheck, filepath.Join(exeDir, "web", "version_mapping.json"))
	}
	
	// Try to find the file in any of the paths
	var mappingPath string
	for _, path := range pathsToCheck {
		if Exists(path) {
			mappingPath = path
			fmt.Printf("Found version mapping file at: %s\n", mappingPath)
			break
		}
	}
	
	if mappingPath == "" {
		return fmt.Errorf("could not find version_mapping.json file")
	}

	content, err := os.ReadFile(mappingPath)
	if err != nil {
		return err
	}

	// Initialize the mapping
	JavaVersionMapping = make(map[string]JavaVersionInfo)
	
	return json.Unmarshal(content, &JavaVersionMapping)
}

// Exists checks if a file exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Unzip extracts a ZIP file to the specified destination
func Unzip(src string, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(dest, file.Name)

		// Create directories if they don't exist
		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModeDir)
			continue
		}

		// Create parent directories
		os.MkdirAll(filepath.Dir(filePath), os.ModeDir)

		// Extract file
		srcFile, err := file.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func SetProxy(p string, verifyssl bool) {
	if p != "" && p != "none" {
		proxyUrl, _ := url.Parse(p)
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl), TLSClientConfig: &tls.Config{InsecureSkipVerify: verifyssl}}}
	} else {
		client = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: verifyssl}}}
	}
}

func SetJavaMirror(mirror string) {
	if mirror != "" && mirror != "none" {
		javaBaseAddress = mirror
		if strings.ToLower(javaBaseAddress[0:4]) != "http" {
			javaBaseAddress = "http://" + javaBaseAddress
		}
		if !strings.HasSuffix(javaBaseAddress, "/") {
			javaBaseAddress = javaBaseAddress + "/"
		}
	}
}

func GetFullJavaUrl(path string) string {
	return javaBaseAddress + path
}

func Download(url string, target string, version string) bool {
	output, err := os.Create(target)
	if err != nil {
		fmt.Println("Error while creating", target, "-", err)
		return false
	}
	defer output.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return false
	}

	req.Header.Set("User-Agent", "JDKVM for Windows")

	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return false
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		fmt.Println("Download failed with status code:", response.StatusCode)
		return false
	}

	_, err = io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while writing to file", target, "-", err)
		return false
	}

	return true
}

// GetJava downloads and installs the specified Java version
func GetJava(root string, v string, a string) bool {
	// Load version mapping if not already loaded
	if JavaVersionMapping == nil {
		if err := LoadVersionMapping(); err != nil {
			fmt.Println("Error loading version mapping:", err)
			return false
		}
	}

	// Check if the version is in our version mapping
	versionInfo, exists := JavaVersionMapping[v]
	if !exists {
		fmt.Printf("Error: Unsupported Java version '%s'. Please use one of the supported versions.\n", v)
		fmt.Println("Supported versions:")
		for version := range JavaVersionMapping {
			fmt.Printf("  - %s\n", version)
		}
		return false
	}

	// Use the full version from the mapping
	fullVersion := versionInfo.Latest
	fmt.Printf("Using Java %s version: %s\n", v, fullVersion)

	// Create version directory
	versionDir := filepath.Join(root, "v"+fullVersion)
	
	// Check if version is already installed (verify directory structure)
	if file.Exists(versionDir) && file.Exists(filepath.Join(versionDir, "bin", "java.exe")) {
		fmt.Printf("Java version %s is already installed.\n", fullVersion)
		return true
	} else if file.Exists(versionDir) {
		// Clean up incomplete installation
		fmt.Printf("Found incomplete installation of Java %s. Cleaning up...\n", fullVersion)
		os.RemoveAll(versionDir)
	}
	
	// Create version directory
	os.MkdirAll(versionDir, os.ModeDir)

	// Download the ZIP file
	tempDir := os.TempDir()
	zipPath := filepath.Join(tempDir, fmt.Sprintf("java-%s.zip", fullVersion))
	
	fmt.Printf("Downloading Java from: %s\n", versionInfo.URL)
	fmt.Printf("Saving to: %s\n", zipPath)
	
	if !Download(versionInfo.URL, zipPath, fullVersion) {
		fmt.Println("Failed to download Java ZIP file.")
		os.Remove(zipPath) // Clean up
		os.RemoveAll(versionDir) // Clean up incomplete directory
		return false
	}

	// Unzip the downloaded file
	fmt.Printf("Unzipping Java %s...\n", fullVersion)
	tempExtractDir := filepath.Join(tempDir, fmt.Sprintf("java-%s-extract", fullVersion))
	os.MkdirAll(tempExtractDir, os.ModeDir)
	
	if err := Unzip(zipPath, tempExtractDir); err != nil {
		fmt.Printf("Failed to unzip Java ZIP file: %v\n", err)
		os.Remove(zipPath) // Clean up
		os.RemoveAll(tempExtractDir) // Clean up
		os.RemoveAll(versionDir) // Clean up incomplete directory
		return false
	}

	// Find the extracted JDK directory (it usually has a name like jdk-17.0.11+9)
	extractedItems, _ := os.ReadDir(tempExtractDir)
	var jdkDir string
	for _, item := range extractedItems {
		if item.IsDir() && strings.HasPrefix(item.Name(), "jdk-") {
			jdkDir = filepath.Join(tempExtractDir, item.Name())
			break
		}
	}

	if jdkDir == "" {
		fmt.Println("Failed to find JDK directory in extracted files.")
		os.Remove(zipPath) // Clean up
		os.RemoveAll(tempExtractDir) // Clean up
		os.RemoveAll(versionDir) // Clean up incomplete directory
		return false
	}

	// Move the JDK contents to our version directory
	jdkContents, _ := os.ReadDir(jdkDir)
	for _, item := range jdkContents {
		srcPath := filepath.Join(jdkDir, item.Name())
		destPath := filepath.Join(versionDir, item.Name())
		
		if err := os.Rename(srcPath, destPath); err != nil {
			fmt.Printf("Failed to move %s to %s: %v\n", item.Name(), versionDir, err)
			os.Remove(zipPath) // Clean up
			os.RemoveAll(tempExtractDir) // Clean up
			os.RemoveAll(versionDir) // Clean up incomplete directory
			return false
		}
	}

	// Clean up temporary files
	os.Remove(zipPath)
	os.RemoveAll(tempExtractDir)

	// Verify the installation
	javaExe := filepath.Join(versionDir, "bin", "java.exe")
	if !file.Exists(javaExe) {
		fmt.Println("Java installation verification failed: java.exe not found.")
		return false
	}

	fmt.Printf("Successfully installed Java %s\n", fullVersion)
	return true
}

// GetAvailableVersions returns the available Java versions from the mapping
func GetAvailableVersions() []string {
	// Load version mapping if not already loaded
	if JavaVersionMapping == nil {
		if err := LoadVersionMapping(); err != nil {
			return []string{}
		}
	}

	// Return only the major versions we support
	versions := make([]string, 0, len(JavaVersionMapping))
	for v := range JavaVersionMapping {
		versions = append(versions, v)
	}
	return versions
}

func GetRemoteTextFile(url string) (string, error) {
	response, httperr := client.Get(url)
	if httperr != nil {
		return "", fmt.Errorf("Could not retrieve %v: %v", url, httperr)
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("Error retrieving \"%s\": HTTP Status %v\n", url, response.StatusCode)
	}

	defer response.Body.Close()

	contents, readerr := io.ReadAll(response.Body)
	if readerr != nil {
		return "", fmt.Errorf("error reading HTTP request body: %v", readerr)
	}

	return string(contents), nil
}

func IsJava64bitAvailable(v string) bool {
	// All modern Java versions (8+) are available in 64-bit
	return true
}

func IsJavaArm64bitAvailable(v string) bool {
	// Java 11+ is available in arm64
	version, err := semver.Make(v)
	if err != nil {
		return false
	}
	corepack, _ := semver.Make("11.0.0")
	return version.GTE(corepack)
}
