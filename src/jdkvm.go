package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"jdkvm/arch"
	"jdkvm/file"
	"jdkvm/java"
	"jdkvm/utility"
	"jdkvm/web"
)

// Replaced at build time
var JdkvmVersion = "0.1.0"

type Environment struct {
	settings        string
	root            string
	symlink         string
	arch            string
	java_mirror     string
	proxy           string
	originalpath    string
	originalversion string
	verifyssl       bool
}

var home = filepath.Join(os.Getenv("JDKVM_HOME"), "settings.txt")
var symlink = filepath.Clean(os.Getenv("JDKVM_SYMLINK"))

var env = &Environment{
	settings:        home,
	root:            "",
	symlink:         symlink,
	arch:            strings.ToLower(os.Getenv("PROCESSOR_ARCHITECTURE")),
	java_mirror:     "",
	proxy:           "none",
	originalpath:    "",
	originalversion: "",
	verifyssl:       true,
}

func main() {
	// Initialize environment
	initializeEnvironment()

	// Check admin privileges for commands that need it
	args := os.Args
	if len(args) > 1 {
		command := args[1]
		// Commands that require admin privileges
		needsAdmin := []string{"use", "u", "install", "i", "uninstall", "rm"}
		for _, cmd := range needsAdmin {
			if command == cmd {
				checkAdminPrivileges()
				break
			}
		}
	}

	// Parse arguments
	detail := ""
	procarch := arch.Validate(env.arch)

	// Capture any additional arguments
	if len(args) > 2 {
		detail = args[2]
	}
	if len(args) > 3 {
		if args[3] == "32" || args[3] == "arm64" || args[3] == "64" {
			procarch = args[3]
		}
	}
	if len(args) < 2 {
		help()
		return
	}

	// Run the appropriate method
	switch args[1] {
	case "i":
		fallthrough
	case "install":
		install(detail, procarch)
	case "rm":
		fallthrough
	case "uninstall":
		uninstall(detail)
	case "u":
		fallthrough
	case "use":
		use(detail, procarch)
	case "ls":
		fallthrough
	case "list":
		list(detail)
	case "current":
		current()
	case "proxy":
		proxy(detail)
	case "version":
		fmt.Println(JdkvmVersion)
	default:
		fmt.Printf(`"%s" is not a valid command.`+"\n", args[1])
		help()
	}
}

// Initialize the environment and set up default values
func initializeEnvironment() {
	// Set default JDKVM_HOME if not set
	if os.Getenv("JDKVM_HOME") == "" {
		defaultHome := filepath.Join(os.Getenv("USERPROFILE"), ".jdkvm")
		os.Setenv("JDKVM_HOME", defaultHome)
		env.root = defaultHome
		fmt.Printf("JDKVM_HOME not set, using default: %s\n", defaultHome)
	} else {
		env.root = os.Getenv("JDKVM_HOME")
	}

	// Set default JDKVM_SYMLINK if not set (though we're not using symlinks anymore)
	if os.Getenv("JDKVM_SYMLINK") == "" {
		defaultSymlink := filepath.Join(os.Getenv("USERPROFILE"), ".jdkvm", "symlink")
		os.Setenv("JDKVM_SYMLINK", defaultSymlink)
		env.symlink = defaultSymlink
	}

	// Create necessary directories
	os.MkdirAll(env.root, os.ModeDir)
	
	// Load version mapping
	if err := web.LoadVersionMapping(); err != nil {
		fmt.Printf("Warning: Could not load version mapping: %v\n", err)
		fmt.Println("You may need to specify full version numbers instead of just major versions.")
	}
	
	// Load configuration from settings.txt
	loadSettings()
	
	// Apply proxy settings
	web.SetProxy(env.proxy, env.verifyssl)
}

// Check if we have admin privileges, and try to elevate if needed
func checkAdminPrivileges() {
	if !utility.IsAdmin() && !utility.IsElevated() {
		fmt.Println("Warning: JDKVM may require administrator privileges for some operations.")
		fmt.Println("If you encounter permission errors, run this command again as Administrator.")
	}
}

// ===============================================================
// BEGIN | CLI functions
// ===============================================================
func install(version string, cpuarch string) {
	fmt.Printf("Installing Java version %s (%s-bit)...\n", version, cpuarch)
	
	// Validate version
	if version == "" {
		fmt.Println("Please specify a version to install.")
		return
	}

	// Check if version is already installed
	if java.IsVersionInstalled(env.root, version, cpuarch) {
		fmt.Printf("Java version %s (%s-bit) is already installed.\n", version, cpuarch)
		return
	}

	// Download Java - web.GetJava will handle directory creation with the correct full version
	fmt.Printf("Downloading Java version %s (%s-bit)...\n", version, cpuarch)
	success := web.GetJava(env.root, version, cpuarch)
	if !success {
		fmt.Printf("Failed to download Java version %s (%s-bit).\n", version, cpuarch)
		return
	}

	fmt.Printf("Java version %s (%s-bit) installed successfully.\n", version, cpuarch)
	fmt.Printf("To use this version, type: jdkvm use %s\n", version)
}

func use(version string, cpuarch string) {
	if version == "" {
		fmt.Println("Please specify a version to use.")
		return
	}

	// Determine the actual version to use
	var actualVersion string
	if !strings.Contains(version, ".") {
		// If it's a major version (like 17, 11, 8), find the installed version
		files, _ := os.ReadDir(env.root)
		for _, f := range files {
			if f.IsDir() && strings.HasPrefix(f.Name(), "v"+version+".") {
				// Extract the version number from the directory name
				actualVersion = strings.TrimPrefix(f.Name(), "v")
				break
			}
		}
		
		if actualVersion == "" {
			fmt.Printf("Java version %s (%s-bit) is not installed.\n", version, cpuarch)
			return
		}
	} else {
		// Exact version specified
		actualVersion = version
		if !java.IsVersionInstalled(env.root, actualVersion, cpuarch) {
			fmt.Printf("Java version %s (%s-bit) is not installed.\n", actualVersion, cpuarch)
			return
		}
	}

	// Instead of using symlinks (which require admin rights), we'll directly set JAVA_HOME
	// and add the bin directory to PATH
	installDir := filepath.Join(env.root, "v"+actualVersion)
	if !file.Exists(installDir) {
		fmt.Printf("Java installation directory not found: %s\n", installDir)
		return
	}

	// Set JAVA_HOME environment variable
	fmt.Println("Setting JAVA_HOME environment variable...")
	var err error
	err = utility.SetEnvironmentVariable("JAVA_HOME", installDir)
	if err != nil {
		fmt.Printf("Failed to set JAVA_HOME: %v\n", err)
		fmt.Println("You may need to set it manually or run as Administrator.")
	} else {
		os.Setenv("JAVA_HOME", installDir) // Also set for current process
		fmt.Printf("JAVA_HOME set to: %s\n", installDir)
	}

	// Get the new bin directory
	javaBinDir := filepath.Join(installDir, "bin")
	fmt.Printf("Using Java bin directory: %s\n", javaBinDir)
	
	// Update current process PATH for immediate use
	currentPath := os.Getenv("PATH")
	
	// Remove any existing Java bin directories from PATH
	paths := strings.Split(currentPath, ";")
	newPaths := make([]string, 0)
	for _, path := range paths {
		trimmedPath := strings.TrimSpace(path)
		// Skip any Java bin directories that are from our installation
		if trimmedPath != "" && !strings.Contains(trimmedPath, filepath.Join(env.root, "v")) {
			newPaths = append(newPaths, trimmedPath)
		}
	}
	
	// Add the new bin directory to the beginning of PATH
	newPaths = append([]string{javaBinDir}, newPaths...)
	newPath := strings.Join(newPaths, ";")
	
	// Set the new PATH for current process
	os.Setenv("PATH", newPath)
	
	// Try to set the system PATH (may require admin rights)
	fmt.Printf("Updating PATH environment variable to include %s\n", javaBinDir)
	err = utility.SetEnvironmentVariable("PATH", newPath)
	if err != nil {
		fmt.Printf("Failed to update system PATH: %v\n", err)
		fmt.Println("The PATH has been updated for the current session, but you may need to update it manually for future sessions.")
		fmt.Println("To make it permanent, run this command as Administrator or update the PATH manually.")
	} else {
		fmt.Printf("Successfully updated PATH environment variable.\n")
	}

	fmt.Printf("Now using Java version %s (%s-bit)\n", actualVersion, cpuarch)
	fmt.Println("Note: You may need to restart your command prompt for changes to take effect.")
}

func list(listtype string) {
	if listtype == "" {
		listtype = "installed"
	}

	if listtype == "installed" {
		fmt.Println("")
		installed := java.GetInstalled(env.root)
		if len(installed) == 0 {
			fmt.Println("No installations recognized.")
			return
		}

		current, _ := java.GetCurrentVersion()
		for _, version := range installed {
			status := "    "
			if version == current {
				status = "  * "
			}
			fmt.Printf("%s%s\n", status, version)
		}
	} else if listtype == "available" {
		fmt.Println("\nAvailable Java versions:")
		// Get available versions from version mapping
		availableVersions := web.GetAvailableVersions()
		if len(availableVersions) > 0 {
			for _, version := range availableVersions {
				if web.JavaVersionMapping != nil && web.JavaVersionMapping[version].Latest != "" {
					fmt.Printf("%s (latest: %s)\n", version, web.JavaVersionMapping[version].Latest)
				} else {
					fmt.Printf("%s\n", version)
				}
			}
		} else {
			fmt.Println("Could not load available versions. Please check your version_mapping.json file.")
		}
		fmt.Println("\nYou can install any of these versions by typing: jdkvm install <version>")
	fmt.Println("For example: jdkvm install 17")
	} else {
		fmt.Println("\nInvalid list option.\n\nPlease use one of the following\n  - jdkvm list\n  - jdkvm list installed\n  - jdkvm list available")
	}
}

func uninstall(version string) {
	if version == "" {
		fmt.Println("Please specify a version to uninstall.")
		return
	}

	// Check if version is installed
	if !java.IsVersionInstalled(env.root, version, "32") && !java.IsVersionInstalled(env.root, version, "64") {
		fmt.Printf("Java version %s is not installed.\n", version)
		return
	}

	// Remove installation directory
	installDir := filepath.Join(env.root, "v"+version)
	err := os.RemoveAll(installDir)
	if err != nil {
		fmt.Printf("Failed to uninstall Java version %s: %v\n", version, err)
		return
	}

	fmt.Printf("Java version %s uninstalled successfully.\n", version)
}

func current() {
	inuse, arch := java.GetCurrentVersion()
	if inuse == "Unknown" {
		fmt.Println("No current version. Run 'jdkvm use x.x.x' to set a version.")
		return
	}

	fmt.Printf("Java version %s (%s-bit) is currently in use.\n", inuse, arch)
}

// Load settings from configuration file
func loadSettings() {
	if !file.Exists(env.settings) {
		// Create default settings file
		saveSettings()
		return
	}

	content, err := os.ReadFile(env.settings)
	if err != nil {
		fmt.Printf("Warning: Could not read settings file: %v\n", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "proxy":
			env.proxy = value
		case "java_mirror":
			env.java_mirror = value
		case "verifyssl":
			env.verifyssl = value == "true"
		}
	}
}

// Save settings to configuration file
func saveSettings() {
	content := "# JDKVM configuration file\n"
	content += fmt.Sprintf("proxy=%s\n", env.proxy)
	content += fmt.Sprintf("java_mirror=%s\n", env.java_mirror)
	content += fmt.Sprintf("verifyssl=%t\n", env.verifyssl)

	err := os.WriteFile(env.settings, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Warning: Could not save settings file: %v\n", err)
	}
}

// Handle proxy command
func proxy(proxyUrl string) {
	if proxyUrl == "" {
		// Show current proxy settings
		fmt.Printf("Current proxy: %s\n", env.proxy)
		if env.proxy != "none" {
			fmt.Println("To remove proxy, use: jdkvm proxy none")
		}
	} else {
		// Set new proxy settings
		env.proxy = proxyUrl
		
		// Apply proxy settings to HTTP client
		web.SetProxy(proxyUrl, env.verifyssl)
		
		// Save to configuration file
		saveSettings()
		
		if proxyUrl == "none" {
			fmt.Println("Proxy removed.")
		} else {
			fmt.Printf("Proxy set to: %s\n", proxyUrl)
		}
	}
}

func help() {
	fmt.Println("\nUsage: jdkvm [command] [arguments]\n")
	fmt.Println("Commands:")
	fmt.Println("  install, i    Install a specific Java version")
	fmt.Println("  use, u        Switch to a specific Java version")
	fmt.Println("  uninstall, rm Uninstall a specific Java version")
	fmt.Println("  list, ls      List installed or available Java versions")
	fmt.Println("  current       Show current Java version")
	fmt.Println("  proxy         Set or show proxy settings")
	fmt.Println("  version       Show JDKVM version")

	fmt.Println("\nExamples:")
	fmt.Println("  jdkvm install 17")
	fmt.Println("  jdkvm use 17")
	fmt.Println("  jdkvm list installed")
	fmt.Println("  jdkvm proxy http://127.0.0.1:7890")
	fmt.Println("  jdkvm proxy none")
	fmt.Println("  jdkvm uninstall 17")
}
