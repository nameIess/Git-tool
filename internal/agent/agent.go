package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nameIess/git-tool/internal/logger"
	"github.com/nameIess/git-tool/internal/platform"
	"github.com/nameIess/git-tool/internal/runner"
)

func StartAndAddKey(keyPath string, shellType platform.ShellType) error {
	if runtime.GOOS=="windows" && shellType!=platform.GitBash {
		logger.Info("Starting ssh-agent service (Windows/PowerShell)")
		if res:=runner.Run("powershell","-Command","Get-Service ssh-agent | Set-Service -StartupType Manual");!res.Success(){logger.Warn("Could not set ssh-agent startup type: %s",res.CombinedOutput())}
		startRes:=runner.Run("powershell","-Command","Start-Service ssh-agent")
		if !startRes.Success(){statusRes:=runner.Run("powershell","-Command","(Get-Service ssh-agent).Status");if !strings.EqualFold(strings.TrimSpace(statusRes.Stdout),"Running"){return fmt.Errorf("could not start ssh-agent service. Try running as Administrator")}}
		if addRes:=runner.Run("ssh-add",keyPath);!addRes.Success(){return fmt.Errorf("failed to add key to agent: %s",addRes.CombinedOutput())}
		return VerifyKeyLoaded(keyPath)
	}
	if addRes:=runner.Run("ssh-add",keyPath);addRes.Success(){return VerifyKeyLoaded(keyPath)}
	agentRes:=runner.Run("ssh-agent","-s")
	if !agentRes.Success(){return fmt.Errorf("failed to start ssh-agent: %s",agentRes.CombinedOutput())}
	addRes:=runner.Run("ssh-add",keyPath)
	if !addRes.Success(){return fmt.Errorf("failed to add key: %s",addRes.CombinedOutput())}
	return VerifyKeyLoaded(keyPath)
}

func EnsureAgentHasKey(keyPath string)error{if err:=VerifyKeyLoaded(keyPath);err==nil{return nil};return StartAndAddKey(keyPath,platform.Current())}

func VerifyKeyLoaded(keyPath string)error{
	res:=runner.Run("ssh-keygen","-lf",keyPath+".pub")
	if !res.Success(){return fmt.Errorf("failed to calculate SSH key fingerprint: %s",res.CombinedOutput())}
	fields:=strings.Fields(res.Stdout);if len(fields)<2{return fmt.Errorf("could not parse SSH key fingerprint")}
	expected:=fields[1]
	listRes:=runner.Run("ssh-add","-l")
	if !listRes.Success(){return fmt.Errorf("ssh-agent is not available or has no keys loaded")}
	for _,line:=range strings.Split(listRes.Stdout,"\n"){fields:=strings.Fields(line);if len(fields)>=2&&fields[1]==expected{return nil}}
	return fmt.Errorf("the expected SSH key is not loaded in ssh-agent")
}

const bashrcMarker="# SSH Agent Auto-Start (added by Git-Tool)"
const bashrcEndMarker="# END Git-Tool SSH Agent Auto-Start"

func BashrcSnippetExists()bool{content,err:=os.ReadFile(filepath.Join(platform.HomeDir(),".bashrc"));return err==nil&&strings.Contains(string(content),bashrcMarker)}

func WriteBashrcSnippet(keyPath string)error{
	path:=filepath.Join(platform.HomeDir(),".bashrc")
	data,err:=os.ReadFile(path);if err!=nil&&!os.IsNotExist(err){return err}
	content:=string(data);keyName:=filepath.Base(keyPath)
	snippet:=fmt.Sprintf("%s\n"+
		"env=~/.ssh/agent.env\n"+
		"agent_load_env() { test -f \"$env\" && . \"$env\" >| /dev/null; }\n"+
		"agent_start() {\n"+
		"    (umask 077; ssh-agent >| \"$env\")\n"+
		"    . \"$env\" >| /dev/null\n"+
		"}\n"+
		"agent_load_env\n"+
		"agent_run_state=$(ssh-add -l >| /dev/null 2>&1; echo $?)\n"+
		"if [ ! \"$SSH_AUTH_SOCK\" ] || [ \"$agent_run_state\" = 2 ]; then\n"+
		"    agent_start\n"+
		"    ssh-add ~/.ssh/%s\n"+
		"elif [ \"$agent_run_state\" = 1 ]; then\n"+
		"    ssh-add ~/.ssh/%s\n"+
		"fi\n"+
		"unset env\n"+
		"%s",bashrcMarker,keyName,keyName,bashrcEndMarker)
	if start:=strings.Index(content,bashrcMarker);start>=0{
		end:=strings.Index(content[start:],bashrcEndMarker);if end<0{return fmt.Errorf("found an incomplete Git-Tool .bashrc block; refusing to modify it automatically")}
		end+=start+len(bashrcEndMarker);content=strings.TrimRight(content[:start],"\r\n")+"\n"+snippet+content[end:]
	}else{trimmed:=strings.TrimRight(content,"\r\n");if trimmed==""{content=snippet+"\n"}else{content=trimmed+"\n\n"+snippet+"\n"}}
	return os.WriteFile(path,[]byte(content),0644)
}
