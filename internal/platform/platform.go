package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nameIess/git-tool/internal/logger"
)

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

func Detect() ShellType {
	if msys := os.Getenv("MSYSTEM"); msys != "" {
		detectedShell = GitBash
		logger.Info("Detected shell: Git Bash (MSYSTEM=%s)", msys)
		return GitBash
	}
	if tp := os.Getenv("TERM_PROGRAM"); strings.Contains(strings.ToLower(tp), "mintty") {
		detectedShell = GitBash
		logger.Info("Detected shell: Git Bash (TERM_PROGRAM=%s)", tp)
		return GitBash
	}
	if psPath := os.Getenv("PSModulePath"); psPath != "" {
		detectedShell = PowerShell
		logger.Info("Detected shell: PowerShell")
		return PowerShell
	}
	detectedShell = Cmd
	logger.Info("Detected shell: CMD (fallback)")
	return Cmd
}

func Current() ShellType {
	return detectedShell
}

func HomeDir() string {
	if runtime.GOOS == "windows" {
		if home := os.Getenv("USERPROFILE"); home != "" {
			return home
		}
		if home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH"); home != "" {
			return home
		}
	}
	home, _ := os.UserHomeDir()
	return home
}

func SSHDir() string {
	return filepath.Join(HomeDir(), ".ssh")
}

func DefaultKeyPath() string {
	return filepath.Join(SSHDir(), "id_ed25519")
}

func EnsureSSHDir() error {
	dir := SSHDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Info("Creating SSH directory: %s", dir)
		return os.MkdirAll(dir, 0700)
	}
	return nil
}

func WindowsVersion() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}
