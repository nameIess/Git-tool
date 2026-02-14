package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CompletePhase struct {
	gitName  string
	gitEmail string
	keyPath  string
	ghUser   string
	ghOK     bool
	logPath  string
	exiting  bool
}

func NewCompletePhase(name, email, keyPath, ghUser string, ghOK bool, logPath string) CompletePhase {
	return CompletePhase{
		gitName:  name,
		gitEmail: email,
		keyPath:  keyPath,
		ghUser:   ghUser,
		ghOK:     ghOK,
		logPath:  logPath,
	}
}

func (c CompletePhase) Init() tea.Cmd { return nil }

func (c CompletePhase) Update(msg tea.Msg) (CompletePhase, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "enter", "q", "esc":
			c.exiting = true
		}
	}
	return c, nil
}

func (c CompletePhase) View() string {
	var b strings.Builder

	banner := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorText).
		Background(ColorSuccess).
		Padding(0, 3).
		Render(" ✓  Setup Complete! ")
	b.WriteString("\n" + banner + "\n\n")

	b.WriteString(SubtitleStyle.Render("Summary"))
	b.WriteString("\n\n")

	b.WriteString(RenderCheckItem(true, "Git Name", c.gitName))
	b.WriteString("\n")
	b.WriteString(RenderCheckItem(true, "Git Email", c.gitEmail))
	b.WriteString("\n")
	b.WriteString(RenderCheckItem(true, "SSH Key", c.keyPath))
	b.WriteString("\n")
	if c.ghOK {
		b.WriteString(RenderCheckItem(true, "GitHub", fmt.Sprintf("Connected as %s", c.ghUser)))
	} else {
		b.WriteString(RenderCheckItem(false, "GitHub", "Not verified"))
	}
	b.WriteString("\n\n")

	if c.logPath != "" {
		b.WriteString(MutedStyle.Render(fmt.Sprintf("  Log file: %s", c.logPath)))
		b.WriteString("\n\n")
	}

	b.WriteString(MutedStyle.Render("  Press Enter or ESC to exit."))
	b.WriteString("\n")

	return b.String()
}

func (c CompletePhase) IsComplete() bool { return c.exiting }
func (c CompletePhase) ShouldExit() bool { return c.exiting }
