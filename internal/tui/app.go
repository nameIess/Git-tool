package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	PhasePrereqs   = 0
	PhaseGitConfig = 1
	PhaseKeygen    = 2
	PhaseAgent     = 3
	PhaseGitHub    = 4
	PhaseComplete  = 5
)

type App struct {
	phase    int
	width    int
	height   int
	logPath  string
	quitting bool

	prereqs  PrereqsPhase
	gitconf  GitConfigPhase
	keygen   KeygenPhase
	agent    AgentPhase
	github   GitHubPhase
	complete CompletePhase

	gitName  string
	gitEmail string
	keyPath  string
}

func NewApp(logPath string) App {
	return App{
		phase:   PhasePrereqs,
		prereqs: NewPrereqsPhase(),
		logPath: logPath,
	}
}

func (m App) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("Git Tool"), m.prereqs.Init())
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			return m.handlePhaseTransition(msg)
		}
	}
	return m.updateCurrentPhase(msg)
}

func (m App) handlePhaseTransition(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.phase {
	case PhasePrereqs:
		if m.prereqs.IsComplete() {
			m.phase = PhaseGitConfig
			m.gitconf = NewGitConfigPhase()
			return m, m.gitconf.Init()
		}
	case PhaseGitConfig:
		if m.gitconf.IsComplete() {
			m.gitName = m.gitconf.Name()
			m.gitEmail = m.gitconf.Email()
			m.phase = PhaseKeygen
			m.keygen = NewKeygenPhase(m.gitEmail)
			return m, tea.Batch(m.keygen.spinner.Tick, m.keygen.Init())
		}
	case PhaseKeygen:
		if m.keygen.IsComplete() {
			m.keyPath = m.keygen.KeyPath()
			m.phase = PhaseAgent
			m.agent = NewAgentPhase(m.keyPath)
			return m, m.agent.Init()
		}
	case PhaseAgent:
		if m.agent.IsComplete() {
			m.phase = PhaseGitHub
			m.github = NewGitHubPhase(m.keyPath)
			return m, m.github.Init()
		}
	case PhaseGitHub:
		if m.github.IsComplete() {
			m.phase = PhaseComplete
			m.complete = NewCompletePhase(
				m.gitName, m.gitEmail, m.keyPath, m.logPath,
			)
			return m, m.complete.Init()
		}
	case PhaseComplete:
		if m.complete.IsComplete() {
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m.updateCurrentPhase(msg)
}

func (m App) updateCurrentPhase(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.phase {
	case PhasePrereqs:
		m.prereqs, cmd = m.prereqs.Update(msg)
		if m.prereqs.ShouldExit() {
			m.quitting = true
			return m, tea.Quit
		}
	case PhaseGitConfig:
		m.gitconf, cmd = m.gitconf.Update(msg)
		if m.gitconf.ShouldExit() {
			m.quitting = true
			return m, tea.Quit
		}
	case PhaseKeygen:
		m.keygen, cmd = m.keygen.Update(msg)
		if m.keygen.ShouldExit() {
			m.quitting = true
			return m, tea.Quit
		}
	case PhaseAgent:
		m.agent, cmd = m.agent.Update(msg)
		if m.agent.ShouldExit() {
			m.quitting = true
			return m, tea.Quit
		}
	case PhaseGitHub:
		m.github, cmd = m.github.Update(msg)
	case PhaseComplete:
		m.complete, cmd = m.complete.Update(msg)
		if m.complete.ShouldExit() {
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, cmd
}

func (m App) View() string {
	if m.quitting {
		return "\n  Goodbye! \U0001f44b\n\n"
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(RenderHeader(m.phase))
	b.WriteString("\n")

	switch m.phase {
	case PhasePrereqs:
		b.WriteString(m.prereqs.View())
	case PhaseGitConfig:
		b.WriteString(m.gitconf.View())
	case PhaseKeygen:
		b.WriteString(m.keygen.View())
	case PhaseAgent:
		b.WriteString(m.agent.View())
	case PhaseGitHub:
		b.WriteString(m.github.View())
	case PhaseComplete:
		b.WriteString(m.complete.View())
	}

	help := "↑↓ navigate  │  Enter select"
	if m.phase == PhaseGitConfig || m.phase == PhaseKeygen {
		help = "Tab/Enter next  │  ↑↓ navigate"
	}
	b.WriteString("\n")
	b.WriteString(RenderHelpBar(help))
	b.WriteString("\n")
	return b.String()
}
