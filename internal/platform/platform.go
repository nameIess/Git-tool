package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nameIess/git-tool/internal/logger"
)

type ShellType int
const (PowerShell ShellType=iota; Cmd; GitBash)
func (s ShellType) String() string {switch s{case GitBash:return "Git Bash";case PowerShell:return "PowerShell";case Cmd:return "CMD";default:return "Unknown"}}
var detectedShell ShellType
func Detect() ShellType {if os.Getenv("MSYSTEM")!=""||strings.Contains(strings.ToLower(os.Getenv("TERM_PROGRAM")),"mintty"){detectedShell=GitBash;return GitBash};if os.Getenv("PSModulePath")!=""{detectedShell=PowerShell;return PowerShell};detectedShell=Cmd;return Cmd}
func Current() ShellType {return detectedShell}
func HomeDir() string {if runtime.GOOS=="windows"{if h:=os.Getenv("USERPROFILE");h!=""{return h};if h:=os.Getenv("HOMEDRIVE")+os.Getenv("HOMEPATH");h!=""{return h}};h,_:=os.UserHomeDir();return h}
func SSHDir() string{return filepath.Join(HomeDir(),".ssh")}
func DefaultKeyPath() string{return filepath.Join(SSHDir(),"id_ed25519")}
func ConfigDir() string {if h:=HomeDir();h!=""{if runtime.GOOS=="windows"{if a:=os.Getenv("APPDATA");a!=""{return filepath.Join(a,"GitTool")};return filepath.Join(h,"AppData","Roaming","GitTool")};return filepath.Join(h,".config","git-tool")};return "."}
func EnsureSSHDir()error{if err:=os.MkdirAll(SSHDir(),0700);err!=nil{return err};return nil}
func WindowsVersion()string{return runtime.GOOS+"/"+runtime.GOARCH}
var _ = logger.Info
