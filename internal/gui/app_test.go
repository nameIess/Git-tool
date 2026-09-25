package gui

import (
	"path/filepath"
	"testing"
)

func TestThemePersistence(t *testing.T) {
	a := NewApp()
	a.settingsPath = filepath.Join(t.TempDir(), "settings.json")
	if err := a.SetTheme(ThemeDark); err != nil { t.Fatal(err) }
	if got := a.Theme(); got != ThemeDark { t.Fatalf("theme=%q", got) }
	if err := a.loadSettings(); err != nil { t.Fatal(err) }
	if got := a.Theme(); got != ThemeDark { t.Fatalf("persisted theme=%q", got) }
}

func TestThemeValidation(t *testing.T) {
	a := NewApp()
	if err := a.SetTheme(Theme("unknown")); err == nil { t.Fatal("expected invalid theme error") }
}
