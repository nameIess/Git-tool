package shell

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/user/git-tool/internal/logger"
)

// ShellType identifies the shell environment.
type ShellType int

const (
	PowerShell ShellType = iota
	Cmd
	GitBash
)

func (s ShellType) String() string {
	switch s {
	case GitBash:
		return "Git Bash"
	case PowerShell:
		return "PowerShell"
	case Cmd:
		return "CMD"
	default:
		return "Unknown"
	}
}

var detectedShell ShellType

// Detect determines the current shell environment.
func Detect() ShellType {
	// Check for Git Bash (MSYSTEM is set in Git Bash / MSYS2)
	if msys := os.Getenv("MSYSTEM"); msys != "" {
		detectedShell = GitBash
		logger.Info("Detected shell: Git Bash (MSYSTEM=%s)", msys)
		return GitBash
	}

	// Check for TERM_PROGRAM for mintty (Git Bash terminal)
	if tp := os.Getenv("TERM_PROGRAM"); strings.Contains(strings.ToLower(tp), "mintty") {
		detectedShell = GitBash
		logger.Info("Detected shell: Git Bash (TERM_PROGRAM=%s)", tp)
		return GitBash
	}

	// Check for PowerShell via PSModulePath
	if psPath := os.Getenv("PSModulePath"); psPath != "" {
		detectedShell = PowerShell
		logger.Info("Detected shell: PowerShell")
		return PowerShell
	}

	detectedShell = Cmd
	logger.Info("Detected shell: CMD (fallback)")
	return Cmd
}

// Current returns the last detected shell type.
func Current() ShellType {
	return detectedShell
}

// HomeDir returns the user's home directory.
func HomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("USERPROFILE")
		if home != "" {
			return home
		}
		// Fallback
		home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home != "" {
			return home
		}
	}
	home, _ := os.UserHomeDir()
	return home
}

// SSHDir returns the path to the .ssh directory.
func SSHDir() string {
	return filepath.Join(HomeDir(), ".ssh")
}

// DefaultKeyPath returns the default SSH key path (without extension).
func DefaultKeyPath() string {
	return filepath.Join(SSHDir(), "id_ed25519")
}

// EnsureSSHDir creates the .ssh directory if it doesn't exist with proper permissions.
func EnsureSSHDir() error {
	dir := SSHDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Info("Creating SSH directory: %s", dir)
		return os.MkdirAll(dir, 0700)
	}
	return nil
}

// WindowsVersion returns a string with the OS version info.
func WindowsVersion() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}
