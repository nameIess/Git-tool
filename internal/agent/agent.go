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
	if runtime.GOOS == "windows" && shellType != platform.GitBash {
		logger.Info("Starting ssh-agent service (Windows/PowerShell)")
		runner.Run("powershell", "-Command", "Get-Service ssh-agent | Set-Service -StartupType Manual")
		
		startRes := runner.Run("powershell", "-Command", "Start-Service ssh-agent")
		if !startRes.Success() {
			statusRes := runner.Run("powershell", "-Command", "(Get-Service ssh-agent).Status")
			if !strings.EqualFold(strings.TrimSpace(statusRes.Stdout), "Running") {
				return fmt.Errorf("could not start ssh-agent service. Try running as Administrator")
			}
		}
		
		addRes := runner.Run("ssh-add", keyPath)
		if !addRes.Success() {
			return fmt.Errorf("failed to add key to agent: %s", addRes.CombinedOutput())
		}
		return nil
	}
	
	logger.Info("Starting ssh-agent (Git Bash environment)")
	addRes := runner.Run("ssh-add", keyPath)
	if addRes.Success() {
		return nil
	}
	
	agentRes := runner.Run("ssh-agent", "-s")
	if !agentRes.Success() {
		return fmt.Errorf("failed to start ssh-agent: %s", agentRes.CombinedOutput())
	}
	
	addRes = runner.Run("ssh-add", keyPath)
	if !addRes.Success() {
		return fmt.Errorf("failed to add key: %s", addRes.CombinedOutput())
	}
	return nil
}

func EnsureAgentHasKey(keyPath string) error {
	listRes := runner.Run("ssh-add", "-l")
	if strings.Contains(listRes.Stdout, "ED25519") || strings.Contains(listRes.Stdout, "ed25519") {
		return nil
	}
	return StartAndAddKey(keyPath, platform.Current())
}

const bashrcMarker = "# SSH Agent Auto-Start (added by Git-Tool)"

func BashrcSnippetExists() bool {
	bashrcPath := filepath.Join(platform.HomeDir(), ".bashrc")
	content, err := os.ReadFile(bashrcPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), bashrcMarker)
}

func WriteBashrcSnippet(keyPath string) error {
	if BashrcSnippetExists() {
		return nil
	}
	
	bashrcPath := filepath.Join(platform.HomeDir(), ".bashrc")
	keyName := filepath.Base(keyPath)
	
	snippet := fmt.Sprintf("\n%s\nenv=~/.ssh/agent.env\nagent_load_env() { test -f \"$env\" && . \"$env\" >| /dev/null; }\nagent_start() {\n    (umask 077; ssh-agent >| \"$env\")\n    . \"$env\" >| /dev/null\n}\nagent_load_env\nagent_run_state=$(ssh-add -l >| /dev/null 2>&1; echo $?)\nif [ ! \"$SSH_AUTH_SOCK\" ] || [ \"$agent_run_state\" = 2 ]; then\n    agent_start\n    ssh-add ~/.ssh/%s\nelif [ \"$agent_run_state\" = 1 ]; then\n    ssh-add ~/.ssh/%s\nfi\nunset env\n", bashrcMarker, keyName, keyName)

	f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(snippet)
	return err
}
