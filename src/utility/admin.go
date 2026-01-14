package utility

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// Check if the current process has admin privileges
func IsAdmin() bool {
	var sid *windows.SID

	// Create the LocalSystem sid
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	// Check if the current process token is a member of the admin group
	token := windows.Token(0)
	isAdmin, _ := token.IsMember(sid)
	return isAdmin
}

// Check if the current process is elevated
func IsElevated() bool {
	token := windows.Token(0)
	return token.IsElevated()
}

// Exists checks if a file or directory exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Run a command with elevated privileges
func RunElevated(name string, arg ...string) bool {
	// First try to run normally
	cmd := exec.Command(name, arg...)
	err := cmd.Run()
	if err == nil {
		return true
	}

	// If that fails, try with elevation using elevate.cmd (similar to nvm)
	exe, _ := os.Executable()
	elevateCmd := filepath.Join(filepath.Dir(exe), "elevate.cmd")
	
	// If elevate.cmd exists, use it
	if Exists(elevateCmd) {
		cmd := exec.Command(elevateCmd, append([]string{"cmd", "/C", name}, arg...)...)
		err := cmd.Run()
		return err == nil
	}

	// Fallback: Create a simple VBS script to elevate
	vbsPath := filepath.Join(os.TempDir(), "elevate.vbs")
	vbsContent := `
Set objShell = CreateObject("Shell.Application")
objShell.ShellExecute "` + name + `", "` + strings.Join(arg, " ") + `", "", "runas", 1
`
	
	err = os.WriteFile(vbsPath, []byte(vbsContent), 0644)
	if err != nil {
		return false
	}
	defer os.Remove(vbsPath)

	cmd = exec.Command("wscript.exe", vbsPath)
	err = cmd.Run()
	return err == nil
}
