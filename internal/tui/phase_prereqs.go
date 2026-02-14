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

// ─── Messages ───────────────────────────────────────────────────────────────

type prereqsResultMsg struct {
	gitOK      bool
	sshOK      bool
	gitVersion string
	sshVersion string
	gitErr     string
	sshErr     string
}

// ─── Model ──────────────────────────────────────────────────────────────────

type PrereqsPhase struct {
	spinner    spinner.Model
	checking   bool
	done       bool
	gitOK      bool
	sshOK      bool
	gitVersion string
	sshVersion string
	gitErr     string
	sshErr     string
	allPassed  bool
	menu       MenuModel // for failure case: retry / exit
}

func NewPrereqsPhase() PrereqsPhase {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle
	return PrereqsPhase{
		spinner:  s,
		checking: true,
	}
}

func (p PrereqsPhase) Init() tea.Cmd {
	return tea.Batch(p.spinner.Tick, checkPrereqs())
}

func checkPrereqs() tea.Cmd {
	return func() tea.Msg {
		// Small delay for UX
		time.Sleep(500 * time.Millisecond)

		result := prereqsResultMsg{}

		// Check Git
		gitResult := exec.Run("git", "--version")
		if gitResult.Success() {
			result.gitOK = true
			result.gitVersion = strings.TrimSpace(gitResult.Stdout)
			logger.Info("Git found: %s", result.gitVersion)
		} else {
			result.gitOK = false
			result.gitErr = "Git is not installed or not in PATH"
			logger.Error("Git not found: %v", gitResult.Err)
		}

		// Check SSH
		sshResult := exec.Run("ssh", "-V")
		// ssh -V outputs to stderr
		combined := strings.TrimSpace(sshResult.Stderr + sshResult.Stdout)
		if combined != "" && (sshResult.ExitCode == 0 || strings.Contains(strings.ToLower(combined), "openssh")) {
			result.sshOK = true
			result.sshVersion = combined
			logger.Info("SSH found: %s", result.sshVersion)
		} else {
			result.sshOK = false
			result.sshErr = "SSH is not installed or not in PATH"
			logger.Error("SSH not found: %v", sshResult.Err)
		}

		return result
	}
}

func (p PrereqsPhase) Update(msg tea.Msg) (PrereqsPhase, tea.Cmd) {
	switch msg := msg.(type) {
	case prereqsResultMsg:
		p.checking = false
		p.done = true
		p.gitOK = msg.gitOK
		p.sshOK = msg.sshOK
		p.gitVersion = msg.gitVersion
		p.sshVersion = msg.sshVersion
		p.gitErr = msg.gitErr
		p.sshErr = msg.sshErr
		p.allPassed = msg.gitOK && msg.sshOK

		if !p.allPassed {
			p.menu = NewMenu([]string{"Retry", "Exit"})
		}
		return p, nil

	case tea.KeyMsg:
		if p.done && !p.allPassed {
			switch msg.String() {
			case "up", "k":
				p.menu.Up()
			case "down", "j":
				p.menu.Down()
			case "enter":
				p.menu.Select()
				if p.menu.SelectedItem() == "Retry" {
					p.checking = true
					p.done = false
					p.menu = MenuModel{}
					return p, tea.Batch(p.spinner.Tick, checkPrereqs())
				}
			}
		}

	case spinner.TickMsg:
		if p.checking {
			var cmd tea.Cmd
			p.spinner, cmd = p.spinner.Update(msg)
			return p, cmd
		}
	}
	return p, nil
}

func (p PrereqsPhase) View() string {
	var b strings.Builder

	b.WriteString(SubtitleStyle.Render("Checking Prerequisites"))
	b.WriteString("\n\n")

	if p.checking {
		b.WriteString(fmt.Sprintf("  %s Checking for Git and SSH...\n", p.spinner.View()))
		return b.String()
	}

	// Git status
	if p.gitOK {
		b.WriteString(RenderCheckItem(true, "Git", p.gitVersion))
	} else {
		b.WriteString(RenderCheckItem(false, "Git", "Not found"))
		b.WriteString("\n")
		b.WriteString(ErrorStyle.Render("    ╰─ "))
		b.WriteString(InfoStyle.Render("Download: https://git-scm.com/downloads"))
	}
	b.WriteString("\n")

	// SSH status
	if p.sshOK {
		b.WriteString(RenderCheckItem(true, "SSH", p.sshVersion))
	} else {
		b.WriteString(RenderCheckItem(false, "SSH", "Not found"))
		b.WriteString("\n")
		b.WriteString(ErrorStyle.Render("    ╰─ "))
		b.WriteString(InfoStyle.Render("Enable OpenSSH in Windows Settings > Apps > Optional Features"))
	}
	b.WriteString("\n\n")

	if p.allPassed {
		b.WriteString(SuccessStyle.Render("  ✓ All prerequisites met! Press Enter to continue."))
		b.WriteString("\n")
	} else {
		b.WriteString(WarningStyle.Render("  ⚠ Some prerequisites are missing."))
		b.WriteString("\n\n")
		b.WriteString(p.menu.View())
	}

	return b.String()
}

// IsComplete returns true if prerequisites passed and user wants to proceed.
func (p PrereqsPhase) IsComplete() bool {
	return p.done && p.allPassed
}

// ShouldExit returns true if user chose Exit.
func (p PrereqsPhase) ShouldExit() bool {
	return p.done && !p.allPassed && p.menu.SelectedItem() == "Exit"
}

// GitVersion returns the detected git version.
func (p PrereqsPhase) GitVersion() string {
	return p.gitVersion
}

// SSHVersion returns the detected SSH version.
func (p PrereqsPhase) SSHVersion() string {
	return p.sshVersion
}
