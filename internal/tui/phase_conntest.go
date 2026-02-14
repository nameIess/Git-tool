package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/user/git-tool/internal/exec"
	"github.com/user/git-tool/internal/logger"
)

type connectionTestMsg struct {
	success  bool
	username string
	errMsg   string
	output   string
}

type connTestStep int

const (
	ctStepTesting connTestStep = iota
	ctStepSuccess
	ctStepFailed
)

type ConnectionTestPhase struct {
	step     connTestStep
	spinner  spinner.Model
	username string
	errMsg   string
	output   string
	failMenu MenuModel
	complete bool
}

func NewConnectionTestPhase() ConnectionTestPhase {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle
	return ConnectionTestPhase{step: ctStepTesting, spinner: s}
}

func (c ConnectionTestPhase) Init() tea.Cmd {
	return tea.Batch(c.spinner.Tick, testGitHubConn())
}

func testGitHubConn() tea.Cmd {
	return func() tea.Msg {
		logger.Info("Testing SSH connection to GitHub...")
		time.Sleep(500 * time.Millisecond)
		result := exec.RunWithTimeout(15*time.Second,
			"ssh", "-T", "-o", "StrictHostKeyChecking=accept-new", "git@github.com")
		combined := result.CombinedOutput()
		logger.Info("SSH test output: %s (exit %d)", combined, result.ExitCode)

		if strings.Contains(combined, "successfully authenticated") {
			un := ""
			if idx := strings.Index(combined, "Hi "); idx >= 0 {
				rest := combined[idx+3:]
				if end := strings.Index(rest, "!"); end >= 0 {
					un = rest[:end]
				}
			}
			return connectionTestMsg{success: true, username: un, output: combined}
		}
		errMsg := "Connection test did not return expected response."
		if strings.Contains(combined, "Permission denied") {
			errMsg = "Permission denied. SSH key may not be added to GitHub."
		} else if strings.Contains(combined, "Connection refused") || strings.Contains(combined, "timed out") {
			errMsg = "Could not reach GitHub. Check network/firewall."
		}
		return connectionTestMsg{success: false, errMsg: errMsg, output: combined}
	}
}

func (c ConnectionTestPhase) Update(msg tea.Msg) (ConnectionTestPhase, tea.Cmd) {
	switch msg := msg.(type) {
	case connectionTestMsg:
		if msg.success {
			c.step = ctStepSuccess
			c.username = msg.username
			c.output = msg.output
			c.complete = true
		} else {
			c.step = ctStepFailed
			c.errMsg = msg.errMsg
			c.output = msg.output
			c.failMenu = NewMenu([]string{"Retry test", "Continue anyway", "Exit"})
		}
		return c, nil
	case spinner.TickMsg:
		if c.step == ctStepTesting {
			var cmd tea.Cmd
			c.spinner, cmd = c.spinner.Update(msg)
			return c, cmd
		}
	case tea.KeyMsg:
		if c.step == ctStepFailed {
			switch msg.String() {
			case "up", "k":
				c.failMenu.Up()
			case "down", "j":
				c.failMenu.Down()
			case "enter":
				c.failMenu.Select()
				switch c.failMenu.SelectedItem() {
				case "Retry test":
					c.step = ctStepTesting
					c.failMenu = MenuModel{}
					return c, tea.Batch(c.spinner.Tick, testGitHubConn())
				case "Continue anyway":
					c.complete = true
				}
			}
		}
	}
	return c, nil
}

func (c ConnectionTestPhase) View() string {
	var b strings.Builder
	b.WriteString(SubtitleStyle.Render("Connection Test"))
	b.WriteString("\n\n")
	switch c.step {
	case ctStepTesting:
		b.WriteString(fmt.Sprintf("  %s Testing SSH connection to GitHub...\n", c.spinner.View()))
		b.WriteString(MutedStyle.Render("    ssh -T git@github.com\n"))
	case ctStepSuccess:
		b.WriteString(RenderCheckItem(true, "GitHub", fmt.Sprintf("Authenticated as %s", c.username)))
		b.WriteString("\n\n")
		b.WriteString(InnerBoxStyle.Render(SuccessStyle.Render(c.output)))
		b.WriteString("\n\n")
		b.WriteString(SuccessStyle.Render("  ✓ Connection verified! Press Enter to continue.\n"))
	case ctStepFailed:
		b.WriteString(ErrorStyle.Render("  ✗ Connection test failed\n\n"))
		b.WriteString(WarningStyle.Render("  " + c.errMsg + "\n\n"))
		if c.output != "" {
			b.WriteString(InnerBoxStyle.Render(MutedStyle.Render(c.output)))
			b.WriteString("\n\n")
		}
		b.WriteString(c.failMenu.View())
	}
	return b.String()
}

func (c ConnectionTestPhase) IsComplete() bool { return c.complete }
func (c ConnectionTestPhase) ShouldExit() bool {
	return c.step == ctStepFailed && c.failMenu.SelectedItem() == "Exit"
}
func (c ConnectionTestPhase) Username() string { return c.username }
