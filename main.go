//go:generate goversioninfo -icon=icon.ico -manifest="" -o resource_windows.syso

package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nameIess/git-tool/internal/logger"
	"github.com/nameIess/git-tool/internal/platform"
	"github.com/nameIess/git-tool/internal/tui"
)

func main() {
	// Determine exe directory for log file
	exePath, err := os.Executable()
	if err != nil {
		exePath = "."
	}
	logDir := filepath.Join(filepath.Dir(exePath), "git-setup-log")
	os.MkdirAll(logDir, 0755)

	// Initialize logger
	log, err := logger.Init(logDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not create log file: %v\n", err)
	} else {
		defer log.Close()
	}

	// Log system info
	logger.Info("=== Git & SSH Setup Tool v2.0.0 ===")
	logger.Info("OS: %s", platform.WindowsVersion())

	// Detect shell environment
	shellType := platform.Detect()
	logger.Info("Shell: %s", shellType)
	logger.Info("Home: %s", platform.HomeDir())
	logger.Info("SSH Dir: %s", platform.SSHDir())

	// Determine log path for display
	logPath := ""
	if log != nil {
		logPath = log.FilePath()
	}

	// Create and run TUI
	app := tui.NewApp(logPath)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		logger.Error("TUI error: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
