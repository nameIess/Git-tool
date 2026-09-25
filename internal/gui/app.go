package gui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nameIess/git-tool/internal/agent"
	"github.com/nameIess/git-tool/internal/clipboard"
	"github.com/nameIess/git-tool/internal/gitcfg"
	"github.com/nameIess/git-tool/internal/logger"
	"github.com/nameIess/git-tool/internal/platform"
	"github.com/nameIess/git-tool/internal/runner"
	"github.com/nameIess/git-tool/internal/setup"
	"github.com/nameIess/git-tool/internal/signing"
	"github.com/nameIess/git-tool/internal/sshkey"
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
}

type App struct {
	ctx context.Context
	mu sync.RWMutex
	state State
	settingsPath string
}

func NewApp() *App { return &App{state: State{Step:"prerequisites", Theme:ThemeWindows}} }

func (a *App) Startup(ctx context.Context) { a.ctx=ctx; a.settingsPath=filepath.Join(platform.ConfigDir(),"settings.json"); if err:=a.loadSettings();err!=nil{logger.Warn("settings load failed: %v",err)} }
func (a *App) State() State { a.mu.RLock(); defer a.mu.RUnlock(); return a.state }
func (a *App) SetTheme(theme Theme) error { if theme!=ThemeDark&&theme!=ThemeWindows{return fmt.Errorf("unsupported theme %q",theme)}; a.mu.Lock();a.state.Theme=theme;a.mu.Unlock();return a.saveSettings() }
func (a *App) Theme() Theme { return a.State().Theme }

func (a *App) CheckPrerequisites() error {
	for _,name:=range []string{"git","ssh","ssh-keygen","ssh-add"}{if _,err:=runner.Which(name);err!=nil{return err}}
	a.setStep("git");return nil
}
func (a *App) ConfigureGit(name,email string,global bool) error {
	name,email=strings.TrimSpace(name),strings.TrimSpace(email);if name==""{return errors.New("Git name is required")};if !gitcfg.ValidateEmail(email){return errors.New("a valid Git email is required")};if err:=gitcfg.Apply(name,email,global);err!=nil{return err};a.mu.Lock();a.state.GitName=name;a.state.GitEmail=email;a.state.Step="ssh";a.mu.Unlock();return nil
}
func (a *App) ExistingGitConfig()(map[string]string,error){n,e:=gitcfg.ReadCurrent();return map[string]string{"name":n,"email":e},nil}
func (a *App) GenerateSSHKey(passphrase string,overwrite bool)(string,error){key:=platform.DefaultKeyPath();if sshkey.Exists(key)&&!overwrite{key=sshkey.FindNextKeyName(key)};if err:=sshkey.Generate(key,a.State().GitEmail,passphrase);err!=nil{return "",err};if err:=sshkey.WriteSSHConfig(key);err!=nil{return "",err};a.mu.Lock();a.state.KeyPath=key;a.state.Step="agent";a.mu.Unlock();return key,nil}
func (a *App) UseExistingSSHKey(path string) error {if strings.TrimSpace(path)==""{path=platform.DefaultKeyPath()};if !sshkey.Exists(path){return fmt.Errorf("SSH key not found: %s",path)};if err:=sshkey.WriteSSHConfig(path);err!=nil{return err};a.mu.Lock();a.state.KeyPath=path;a.state.Step="agent";a.mu.Unlock();return nil}
func (a *App) SetupSSHAgent() error {s:=a.State();if s.KeyPath==""{return errors.New("SSH key is not selected")};if err:=agent.StartAndAddKey(s.KeyPath,platform.Current());err!=nil{return err};a.mu.Lock();a.state.Agent=true;a.state.Step="github";a.mu.Unlock();return nil}
func (a *App) GitHubPublicKey()(string,error){s:=a.State();if s.KeyPath==""{return "",errors.New("SSH key is not selected")};key,err:=sshkey.ReadPublicKey(s.KeyPath);if err!=nil{return "",fmt.Errorf("read public key: %w",err)};_ = clipboard.Copy(key);return key,nil}
func (a *App) OpenGitHubSSHSettings(){setup.OpenURL("https://github.com/settings/ssh/new")}
func (a *App) VerifyGitHubSSH() error {res:=runner.RunWithTimeout(20*time.Second,"ssh","-T","-o","BatchMode=yes","-o","StrictHostKeyChecking=accept-new","git@github.com");out:=strings.ToLower(res.CombinedOutput());if !strings.Contains(out,"successfully authenticated"){return fmt.Errorf("GitHub SSH authentication failed: %s",strings.TrimSpace(res.CombinedOutput()))};a.mu.Lock();a.state.GitHub=true;a.state.Step="signing";a.mu.Unlock();return nil}
func (a *App) ConfigureSigning(enabled bool) error {s:=a.State();if !enabled{a.mu.Lock();a.state.Step="complete";a.state.Signing=false;a.mu.Unlock();return nil};if err:=signing.Configure(s.KeyPath,true);err!=nil{return err};if _,err:=signing.Verify();err!=nil{return err};a.mu.Lock();a.state.Signing=true;a.state.Step="complete";a.mu.Unlock();return nil}
func (a *App) Finish() State {a.mu.Lock();a.state.Step="complete";a.mu.Unlock();return a.State()}
func (a *App) loadSettings()error{data,err:=os.ReadFile(a.settingsPath);if os.IsNotExist(err){return nil};if err!=nil{return err};var s struct{Theme Theme `json:"theme"`};if err:=json.Unmarshal(data,&s);err!=nil{return err};if s.Theme==ThemeDark||s.Theme==ThemeWindows{a.state.Theme=s.Theme};return nil}
func (a *App) saveSettings()error{if a.settingsPath==""{return nil};if err:=os.MkdirAll(filepath.Dir(a.settingsPath),0700);err!=nil{return err};data,err:=json.MarshalIndent(struct{Theme Theme `json:"theme"`}{a.State().Theme},"","  ");if err!=nil{return err};return os.WriteFile(a.settingsPath,data,0600)}
func (a *App) setStep(s string){a.mu.Lock();a.state.Step=s;a.mu.Unlock()}
