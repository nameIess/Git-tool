package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/nameIess/git-tool/internal/agent"
	"github.com/nameIess/git-tool/internal/logger"
	"github.com/nameIess/git-tool/internal/platform"
	"github.com/nameIess/git-tool/internal/signing"
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
	shellType platform.ShellType
}

func NewAgentPhase(keyPath string) AgentPhase {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return AgentPhase{
		step:      asStepStarting,
		spinner:   s,
		keyPath:   keyPath,
		shellType: platform.Current(),
	}
}

func (a AgentPhase) Init() tea.Cmd {
	return tea.Batch(a.spinner.Tick, startAgentAndAddKey(a.keyPath, a.shellType))
}

func startAgentAndAddKey(keyPath string, shellType platform.ShellType) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(500 * time.Millisecond)

		result := agentResultMsg{agentStarted: false, keyAdded: false}

		err := agent.StartAndAddKey(keyPath, shellType)
		if err != nil {
			result.errMsg = err.Error()
			return result
		}

		result.agentStarted = true
		result.keyAdded = true

		// Configure Git for SSH commit signing
		signing.Configure(keyPath, true)
		
		// Write bashrc snippet for Git Bash
		if err := agent.WriteBashrcSnippet(keyPath); err != nil {
			logger.Warn("Failed to write bashrc snippet: %v", err)
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
