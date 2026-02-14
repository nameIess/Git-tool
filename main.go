//go:generate goversioninfo -icon=icon.ico -manifest="" -o resource_windows.syso

package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/user/git-tool/internal/logger"
	"github.com/user/git-tool/internal/shell"
	"github.com/user/git-tool/internal/tui"
)

func main() {
	// Determine exe directory for log file
	exePath, err := os.Executable()
	if err != nil {
		exePath = "."
	}
	exeDir := filepath.Dir(exePath)

	// Initialize logger
	log, err := logger.Init(exeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not create log file: %v\n", err)
	} else {
		defer log.Close()
	}

	// Log system info
	logger.Info("=== Git & SSH Setup Tool ===")
	logger.Info("OS: %s", shell.WindowsVersion())

	// Detect shell environment
	shellType := shell.Detect()
	logger.Info("Shell: %s", shellType)
	logger.Info("Home: %s", shell.HomeDir())
	logger.Info("SSH Dir: %s", shell.SSHDir())

	// Determine log path for display
	logPath := ""
	if log != nil {
		logPath = log.FilePath()
	}

	// Create and run TUI
	model := tui.NewModel(logPath)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		logger.Error("TUI error: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
