package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nameIess/git-tool/internal/platform"
	"github.com/nameIess/git-tool/internal/runner"
)

func StartAndAddKey(keyPath string,shell platform.ShellType)error{if runtime.GOOS=="windows"&&shell!=platform.GitBash{runner.Run("powershell","-Command","Set-Service ssh-agent -StartupType Manual");start:=runner.Run("powershell","-Command","Start-Service ssh-agent");if !start.Success(){status:=runner.Run("powershell","-Command","(Get-Service ssh-agent).Status");if !strings.EqualFold(strings.TrimSpace(status.Stdout),"Running"){return fmt.Errorf("could not start ssh-agent service; try running as Administrator")}}};add:=runner.Run("ssh-add",keyPath);if !add.Success(){return fmt.Errorf("failed to add SSH key: %s",add.CombinedOutput())};return VerifyKeyLoaded(keyPath)}
func VerifyKeyLoaded(keyPath string)error{expected:=runner.Run("ssh-keygen","-lf",keyPath+".pub");if !expected.Success(){return fmt.Errorf("cannot calculate SSH key fingerprint: %s",expected.CombinedOutput())};finger:=strings.Fields(expected.Stdout);if len(finger)<2{return fmt.Errorf("invalid SSH fingerprint output")};loaded:=runner.Run("ssh-add","-l");if !loaded.Success(){return fmt.Errorf("cannot list SSH agent keys: %s",loaded.CombinedOutput())};for _,line:=range strings.Split(loaded.Stdout,"\n"){f:=strings.Fields(line);if len(f)>=2&&f[1]==finger[1]{return nil}};return fmt.Errorf("selected SSH key is not loaded in the agent")}
const begin="# BEGIN Git-Tool managed bashrc";const end="# END Git-Tool managed bashrc"
func WriteBashrcSnippet(keyPath string)error{if platform.Current()!=platform.GitBash{return nil};path:=filepath.Join(platform.HomeDir(),".bashrc");old,_:=os.ReadFile(path);block:=fmt.Sprintf("%s\nif ! ssh-add -l >/dev/null 2>&1; then\n  ssh-add %q >/dev/null 2>&1 || true\nfi\n%s\n",begin,keyPath,end);content:=string(old);if i:=strings.Index(content,begin);i>=0{if j:=strings.Index(content[i:],end);j>=0{j=i+j+len(end);content=content[:i]+block+content[j:];return os.WriteFile(path,[]byte(content),0644)}};if content!=""&&!strings.HasSuffix(content,"\n"){content+="\n"};content+=block;return os.WriteFile(path,[]byte(content),0644)}
