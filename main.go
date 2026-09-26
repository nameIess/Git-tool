//go:generate goversioninfo -icon=icon.ico -manifest="" -o resource_windows.syso

package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	windows "github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/nameIess/git-tool/internal/gui"
	"github.com/nameIess/git-tool/internal/logger"
)

var (
	//go:embed all:frontend/dist
	assets embed.FS
)

func main() {
	exePath, err := os.Executable()
	if err != nil { exePath = "." }
	logDir := filepath.Join(filepath.Dir(exePath), "git-setup-log")
	if err := os.MkdirAll(logDir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot create log directory: %v\n", err)
	}
	log, err := logger.Init(logDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot initialize logger: %v\n", err)
	} else {
		defer log.Close()
	}

	app := gui.NewApp()
	err = wails.Run(&options.App{
		Title: "Git Tool",
		Width: 1180,
		Height: 760,
		MinWidth: 980,
		MinHeight: 640,
		Assets: assets,
		BackgroundColour: &options.RGBA{R: 13, G: 17, B: 23, A: 1},
		OnStartup: app.Startup,
		Bind: []interface{}{app},
		Windows: &windows.Options{Theme: windows.SystemDefault},
	})
	if err != nil {
		logger.Error("GUI exited with error: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
