package tui

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/user/git-tool/internal/exec"
	"github.com/user/git-tool/internal/logger"
	"github.com/user/git-tool/internal/shell"
)

// ─── Messages ───────────────────────────────────────────────────────────────

type agentResultMsg struct {
	agentStarted bool
	keyAdded     bool
	errMsg       string
}

// ─── State ──────────────────────────────────────────────────────────────────

type agentStep int

const (
	asStepStarting agentStep = iota
	asStepSuccess
	asStepFailed
)

// ─── Model ──────────────────────────────────────────────────────────────────

type AgentPhase struct {
	step      agentStep
	spinner   spinner.Model
	keyPath   string
	errMsg    string
	failMenu  MenuModel
	complete  bool
	shellType shell.ShellType
}

func NewAgentPhase(keyPath string) AgentPhase {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return AgentPhase{
		step:      asStepStarting,
		spinner:   s,
		keyPath:   keyPath,
		shellType: shell.Current(),
	}
}

func (a AgentPhase) Init() tea.Cmd {
	return tea.Batch(a.spinner.Tick, startAgentAndAddKey(a.keyPath, a.shellType))
}

func startAgentAndAddKey(keyPath string, shellType shell.ShellType) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(500 * time.Millisecond)

		result := agentResultMsg{agentStarted: false, keyAdded: false}

		if runtime.GOOS == "windows" && shellType != shell.GitBash {
			// PowerShell / CMD: Try to start the ssh-agent service
			logger.Info("Starting ssh-agent service (Windows/PowerShell)")

			// First try to set the service to manual start
			setResult := exec.Run("powershell", "-Command",
				"Get-Service ssh-agent | Set-Service -StartupType Manual")
			if !setResult.Success() {
				logger.Warn("Could not set ssh-agent startup type (may need admin): %s", setResult.CombinedOutput())
				// Continue anyway, it might already be configured
			}

			// Start the service
			startResult := exec.Run("powershell", "-Command",
				"Start-Service ssh-agent")
			if startResult.Success() {
				result.agentStarted = true
				logger.Info("ssh-agent service started")
			} else {
				// Check if it's already running
				statusResult := exec.Run("powershell", "-Command",
					"(Get-Service ssh-agent).Status")
				status := strings.TrimSpace(statusResult.Stdout)
				if strings.EqualFold(status, "Running") {
					result.agentStarted = true
					logger.Info("ssh-agent service already running")
				} else {
					result.errMsg = fmt.Sprintf("Could not start ssh-agent service. Status: %s. You may need to run as Administrator.", status)
					logger.Error("ssh-agent start failed: %s", result.errMsg)
					return result
				}
			}

			// Add key using ssh-add
			addResult := exec.Run("ssh-add", keyPath)
			if addResult.Success() {
				result.keyAdded = true
				logger.Info("SSH key added to agent: %s", keyPath)
			} else {
				result.errMsg = fmt.Sprintf("Failed to add key to agent: %s", addResult.CombinedOutput())
				logger.Error("ssh-add failed: %s", result.errMsg)
			}
		} else {
			// Git Bash: use eval ssh-agent and ssh-add
			logger.Info("Starting ssh-agent (Git Bash environment)")

			// In Git Bash, ssh-agent is usually already available
			// Just try to add the key directly
			result.agentStarted = true

			addResult := exec.Run("ssh-add", keyPath)
			if addResult.Success() {
				result.keyAdded = true
				logger.Info("SSH key added to agent: %s", keyPath)
			} else {
				// Try starting agent first
				agentResult := exec.Run("ssh-agent", "-s")
				if agentResult.Success() {
					logger.Info("ssh-agent started: %s", strings.TrimSpace(agentResult.Stdout))
				}
				// Retry add
				addResult = exec.Run("ssh-add", keyPath)
				if addResult.Success() {
					result.keyAdded = true
					logger.Info("SSH key added to agent (after starting agent)")
				} else {
					result.errMsg = fmt.Sprintf("Failed to add key: %s", addResult.CombinedOutput())
					logger.Error("ssh-add failed: %s", result.errMsg)
				}
			}
		}

		// If key was added, also configure Git for SSH commit signing
		if result.keyAdded {
			configureCommitSigning(keyPath)
		}

		return result
	}
}

func (a AgentPhase) Update(msg tea.Msg) (AgentPhase, tea.Cmd) {
	switch msg := msg.(type) {
	case agentResultMsg:
		if msg.agentStarted && msg.keyAdded {
			a.step = asStepSuccess
			a.complete = true
		} else {
			a.step = asStepFailed
			a.errMsg = msg.errMsg
			if a.errMsg == "" {
				a.errMsg = "Unknown error occurred"
			}
			a.failMenu = NewMenu([]string{"Retry", "Continue anyway", "Exit"})
		}
		return a, nil

	case spinner.TickMsg:
		if a.step == asStepStarting {
			var cmd tea.Cmd
			a.spinner, cmd = a.spinner.Update(msg)
			return a, cmd
		}

	case tea.KeyMsg:
		switch a.step {
		case asStepFailed:
			switch msg.String() {
			case "up", "k":
				a.failMenu.Up()
			case "down", "j":
				a.failMenu.Down()
			case "enter":
				a.failMenu.Select()
				switch a.failMenu.SelectedItem() {
				case "Retry":
					a.step = asStepStarting
					a.failMenu = MenuModel{}
					return a, tea.Batch(a.spinner.Tick, startAgentAndAddKey(a.keyPath, a.shellType))
				case "Continue anyway":
					a.complete = true
					a.step = asStepSuccess
				}
			}
		}
	}

	return a, nil
}

func (a AgentPhase) View() string {
	var b strings.Builder

	b.WriteString(SubtitleStyle.Render("SSH Agent Setup"))
	b.WriteString("\n\n")

	b.WriteString(MutedStyle.Render(fmt.Sprintf("  Environment: %s", a.shellType)))
	b.WriteString("\n\n")

	switch a.step {
	case asStepStarting:
		b.WriteString(fmt.Sprintf("  %s Starting ssh-agent and adding key...\n", a.spinner.View()))

	case asStepSuccess:
		b.WriteString(RenderCheckItem(true, "SSH Agent", "Running"))
		b.WriteString("\n")
		b.WriteString(RenderCheckItem(true, "Key Added", a.keyPath))
		b.WriteString("\n\n")
		b.WriteString(SuccessStyle.Render("  ✓ SSH agent configured! Press Enter to continue."))
		b.WriteString("\n")

	case asStepFailed:
		b.WriteString(ErrorStyle.Render("  ✗ SSH Agent setup encountered an issue"))
		b.WriteString("\n\n")
		b.WriteString(InnerBoxStyle.Render(a.errMsg))
		b.WriteString("\n\n")
		if strings.Contains(a.errMsg, "Administrator") {
			b.WriteString(WarningStyle.Render("  💡 Tip: Try running this tool as Administrator"))
			b.WriteString("\n\n")
		}
		b.WriteString(a.failMenu.View())
	}

	return b.String()
}

func (a AgentPhase) IsComplete() bool {
	return a.complete
}

func (a AgentPhase) ShouldExit() bool {
	return a.step == asStepFailed && a.failMenu.SelectedItem() == "Exit"
}

// configureCommitSigning sets up Git to use SSH for commit signing.
func configureCommitSigning(keyPath string) {
	pubKeyPath := keyPath + ".pub"
	logger.Info("Configuring Git for SSH commit signing (key: %s)", pubKeyPath)

	res := exec.Run("git", "config", "--global", "gpg.format", "ssh")
	if !res.Success() {
		logger.Warn("Failed to set gpg.format: %s", res.CombinedOutput())
	}

	res = exec.Run("git", "config", "--global", "user.signingkey", pubKeyPath)
	if !res.Success() {
		logger.Warn("Failed to set user.signingkey: %s", res.CombinedOutput())
	}

	res = exec.Run("git", "config", "--global", "commit.gpgsign", "true")
	if !res.Success() {
		logger.Warn("Failed to set commit.gpgsign: %s", res.CombinedOutput())
	}

	logger.Info("SSH commit signing configured successfully")
}
