package gui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/nameIess/git-tool/internal/agent"
	"github.com/nameIess/git-tool/internal/gitcfg"
	"github.com/nameIess/git-tool/internal/logger"
	"github.com/nameIess/git-tool/internal/platform"
	"github.com/nameIess/git-tool/internal/runner"
	"github.com/nameIess/git-tool/internal/signing"
	"github.com/nameIess/git-tool/internal/sshkey"
	"github.com/nameIess/git-tool/internal/clipboard"
	"github.com/nameIess/git-tool/internal/setup"
)

type Theme string
const (
	ThemeDark Theme = "dark"
	ThemeWindows Theme = "windows"
)

type State struct {
	Step string `json:"step"`
	Theme Theme `json:"theme"`
	GitName string `json:"gitName"`
	GitEmail string `json:"gitEmail"`
	KeyPath string `json:"keyPath"`
	Agent bool `json:"agent"`
	GitHub bool `json:"github"`
	Signing bool `json:"signing"`
	Error string `json:"error,omitempty"`
}

type App struct {
	ctx context.Context
	mu sync.RWMutex
	state State
	settingsPath string
}

func NewApp() *App {
	return &App{state: State{Step: "prerequisites", Theme: ThemeWindows}}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.settingsPath = filepath.Join(platform.ConfigDir(), "settings.json")
	if err := a.loadSettings(); err != nil { logger.Warn("settings load failed: %v", err) }
}

func (a *App) State() State { a.mu.RLock(); defer a.mu.RUnlock(); return a.state }

func (a *App) SetTheme(theme Theme) error {
	if theme != ThemeDark && theme != ThemeWindows { return fmt.Errorf("unsupported theme %q", theme) }
	a.mu.Lock(); a.state.Theme = theme; a.mu.Unlock()
	return a.saveSettings()
}

func (a *App) Theme() Theme { a.mu.RLock(); defer a.mu.RUnlock(); return a.state.Theme }

func (a *App) CheckPrerequisites() error {
	for _, name := range []string{"git", "ssh", "ssh-keygen", "ssh-add"} {
		if _, err := runner.Which(name); err != nil { return err }
	}
	a.setStep("git")
	return nil
}

func (a *App) ConfigureGit(name, email string, global bool) error {
	name, email = strings.TrimSpace(name), strings.TrimSpace(email)
	if name == "" { return errors.New("Git name is required") }
	if !gitcfg.ValidateEmail(email) { return errors.New("a valid Git email is required") }
	if err := gitcfg.Apply(name, email, global); err != nil { return err }
	a.mu.Lock(); a.state.GitName, a.state.GitEmail, a.state.Step = name, email, "ssh"; a.mu.Unlock()
	return nil
}

func (a *App) ExistingGitConfig() (map[string]string, error) {
	name, email := gitcfg.ReadCurrent()
	return map[string]string{"name": name, "email": email}, nil
}

func (a *App) ExistingSSHKey() bool { return sshkey.Exists(platform.DefaultKeyPath()) }

func (a *App) GenerateSSHKey(passphrase string, overwrite bool) (string, error) {
	keyPath := platform.DefaultKeyPath()
	if sshkey.Exists(keyPath) && !overwrite { keyPath = sshkey.FindNextKeyName(keyPath) }
	if err := sshkey.Generate(keyPath, a.state.GitEmail, passphrase); err != nil { return "", err }
	if err := sshkey.WriteSSHConfig(keyPath); err != nil { return "", err }
	a.mu.Lock(); a.state.KeyPath, a.state.Step = keyPath, "agent"; a.mu.Unlock()
	return keyPath, nil
}

func (a *App) UseExistingSSHKey(path string) error {
	if strings.TrimSpace(path) == "" { path = platform.DefaultKeyPath() }
	if !sshkey.Exists(path) { return fmt.Errorf("SSH key not found: %s", path) }
	if err := sshkey.WriteSSHConfig(path); err != nil { return err }
	a.mu.Lock(); a.state.KeyPath, a.state.Step = path, "agent"; a.mu.Unlock()
	return nil
}

func (a *App) SetupSSHAgent() error {
	if a.state.KeyPath == "" { return errors.New("SSH key is not selected") }
	if err := agent.StartAndAddKey(a.state.KeyPath, platform.Current()); err != nil { return err }
	a.mu.Lock(); a.state.Agent, a.state.Step = true, "github"; a.mu.Unlock()
	return nil
}

func (a *App) GitHubPublicKey() (string, error) {
	if a.state.KeyPath == "" { return "", errors.New("SSH key is not selected") }
	key, err := sshkey.ReadPublicKey(a.state.KeyPath)
	if err != nil { return "", fmt.Errorf("read public key: %w", err) }
	_ = clipboard.Copy(key)
	return key, nil
}

func (a *App) OpenGitHubSSHSettings() {
	setup.OpenURL("https://github.com/settings/ssh/new")
}

func (a *App) VerifyGitHubSSH() error {
	res := runner.RunWithTimeout(20_000_000_000, "ssh", "-T", "-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=accept-new", "git@github.com")
	output := strings.ToLower(res.CombinedOutput())
	if !strings.Contains(output, "successfully authenticated") {
		return fmt.Errorf("GitHub SSH authentication failed: %s", strings.TrimSpace(res.CombinedOutput()))
	}
	a.mu.Lock(); a.state.GitHub, a.state.Step = true, "signing"; a.mu.Unlock()
	return nil
}

func (a *App) ConfigureSigning(enabled bool) error {
	if !enabled {
		a.mu.Lock(); a.state.Signing, a.state.Step = false, "complete"; a.mu.Unlock()
		return nil
	}
	if err := signing.Configure(a.state.KeyPath, true); err != nil { return err }
	if _, err := signing.Verify(); err != nil { return err }
	a.mu.Lock(); a.state.Signing, a.state.Step = true, "complete"; a.mu.Unlock()
	return nil
}

func (a *App) Finish() State { a.mu.Lock(); a.state.Step = "complete"; a.mu.Unlock(); return a.State() }

func (a *App) loadSettings() error {
	data, err := os.ReadFile(a.settingsPath)
	if os.IsNotExist(err) { return nil }
	if err != nil { return err }
	var s struct{ Theme Theme `json:"theme"` }
	if err := json.Unmarshal(data, &s); err != nil { return err }
	if s.Theme == ThemeDark || s.Theme == ThemeWindows { a.state.Theme = s.Theme }
	return nil
}

func (a *App) saveSettings() error {
	if a.settingsPath == "" { return nil }
	if err := os.MkdirAll(filepath.Dir(a.settingsPath), 0700); err != nil { return err }
	data, err := json.MarshalIndent(struct{ Theme Theme `json:"theme"` }{a.state.Theme}, "", "  ")
	if err != nil { return err }
	return os.WriteFile(a.settingsPath, data, 0600)
}

func (a *App) setStep(step string) { a.mu.Lock(); a.state.Step = step; a.mu.Unlock() }

func _() { _, _ = runtime.GOOS, setup.Noop; _ = context.Background }
