package tui

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/user/git-tool/internal/exec"
	"github.com/user/git-tool/internal/logger"
)

// ─── Messages ───────────────────────────────────────────────────────────────

type gitConfigReadMsg struct {
	name  string
	email string
}

type gitConfigSetMsg struct {
	err error
}

// ─── State ──────────────────────────────────────────────────────────────────

type gitConfigStep int

const (
	gcStepLoading gitConfigStep = iota
	gcStepShowExisting
	gcStepInputName
	gcStepInputEmail
	gcStepChooseScope
	gcStepApplying
	gcStepDone
)

// ─── Model ──────────────────────────────────────────────────────────────────

type GitConfigPhase struct {
	step      gitConfigStep
	existName string
	existEmail string

	nameInput  textinput.Model
	emailInput textinput.Model

	newName  string
	newEmail string

	scopeMenu MenuModel // Global / Local
	isGlobal  bool

	errorMsg string
	complete bool

	menu MenuModel // Change / Keep / Exit (for existing config)
}

func NewGitConfigPhase() GitConfigPhase {
	ni := textinput.New()
	ni.Placeholder = "John Doe"
	ni.CharLimit = 100
	ni.Width = 40
	ni.PromptStyle = PromptStyle
	ni.TextStyle = TextStyle
	ni.Cursor.Style = AccentStyle

	ei := textinput.New()
	ei.Placeholder = "john@example.com"
	ei.CharLimit = 100
	ei.Width = 40
	ei.PromptStyle = PromptStyle
	ei.TextStyle = TextStyle
	ei.Cursor.Style = AccentStyle

	return GitConfigPhase{
		step:       gcStepLoading,
		nameInput:  ni,
		emailInput: ei,
	}
}

func (g GitConfigPhase) Init() tea.Cmd {
	return readGitConfig()
}

func readGitConfig() tea.Cmd {
	return func() tea.Msg {
		nameResult := exec.Run("git", "config", "--global", "user.name")
		emailResult := exec.Run("git", "config", "--global", "user.email")

		name := strings.TrimSpace(nameResult.Stdout)
		email := strings.TrimSpace(emailResult.Stdout)

		logger.Info("Current git config: name=%q email=%q", name, email)
		return gitConfigReadMsg{name: name, email: email}
	}
}

func applyGitConfig(name, email string, global bool) tea.Cmd {
	return func() tea.Msg {
		scope := "--global"
		if !global {
			scope = "--local"
		}

		logger.Info("Setting git config (%s): name=%q email=%q", scope, name, email)

		res := exec.Run("git", "config", scope, "user.name", name)
		if !res.Success() {
			logger.Error("Failed to set user.name: %v", res.Err)
			return gitConfigSetMsg{err: fmt.Errorf("failed to set user.name: %s", res.CombinedOutput())}
		}

		res = exec.Run("git", "config", scope, "user.email", email)
		if !res.Success() {
			logger.Error("Failed to set user.email: %v", res.Err)
			return gitConfigSetMsg{err: fmt.Errorf("failed to set user.email: %s", res.CombinedOutput())}
		}

		return gitConfigSetMsg{err: nil}
	}
}

func (g GitConfigPhase) Update(msg tea.Msg) (GitConfigPhase, tea.Cmd) {
	switch msg := msg.(type) {
	case gitConfigReadMsg:
		g.existName = msg.name
		g.existEmail = msg.email
		if msg.name != "" && msg.email != "" {
			g.step = gcStepShowExisting
			g.menu = NewMenu([]string{"Change", "Keep Current", "Exit"})
		} else {
			g.step = gcStepInputName
			g.nameInput.Focus()
			if msg.name != "" {
				g.nameInput.SetValue(msg.name)
			}
			if msg.email != "" {
				g.emailInput.SetValue(msg.email)
			}
			return g, textinput.Blink
		}
		return g, nil

	case gitConfigSetMsg:
		if msg.err != nil {
			g.errorMsg = msg.err.Error()
			g.step = gcStepInputName
		} else {
			g.step = gcStepDone
			g.complete = true
			logger.Info("Git config applied successfully")
		}
		return g, nil

	case tea.KeyMsg:
		switch g.step {
		case gcStepShowExisting:
			switch msg.String() {
			case "up", "k":
				g.menu.Up()
			case "down", "j":
				g.menu.Down()
			case "enter":
				g.menu.Select()
				switch g.menu.SelectedItem() {
				case "Change":
					g.step = gcStepInputName
					g.nameInput.SetValue(g.existName)
					g.emailInput.SetValue(g.existEmail)
					g.nameInput.Focus()
					return g, textinput.Blink
				case "Keep Current":
					g.newName = g.existName
					g.newEmail = g.existEmail
					g.step = gcStepDone
					g.complete = true
					logger.Info("Keeping existing git config")
					return g, nil
				case "Exit":
					// Signal exit
					return g, nil
				}
			}

		case gcStepInputName:
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(g.nameInput.Value())
				if val == "" {
					g.errorMsg = "Name cannot be empty"
					return g, nil
				}
				g.newName = val
				g.errorMsg = ""
				g.step = gcStepInputEmail
				g.nameInput.Blur()
				g.emailInput.Focus()
				if g.existEmail != "" && g.emailInput.Value() == "" {
					g.emailInput.SetValue(g.existEmail)
				}
				return g, textinput.Blink
			default:
				var cmd tea.Cmd
				g.nameInput, cmd = g.nameInput.Update(msg)
				return g, cmd
			}

		case gcStepInputEmail:
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(g.emailInput.Value())
				if val == "" {
					g.errorMsg = "Email cannot be empty"
					return g, nil
				}
				if !isValidEmail(val) {
					g.errorMsg = "Please enter a valid email address"
					return g, nil
				}
				g.newEmail = val
				g.errorMsg = ""
				g.step = gcStepChooseScope
				g.emailInput.Blur()
				g.scopeMenu = NewMenu([]string{"Global (recommended)", "Local"})
				return g, nil
			default:
				var cmd tea.Cmd
				g.emailInput, cmd = g.emailInput.Update(msg)
				return g, cmd
			}

		case gcStepChooseScope:
			switch msg.String() {
			case "up", "k":
				g.scopeMenu.Up()
			case "down", "j":
				g.scopeMenu.Down()
			case "enter":
				g.scopeMenu.Select()
				g.isGlobal = g.scopeMenu.Cursor == 0
				g.step = gcStepApplying
				return g, applyGitConfig(g.newName, g.newEmail, g.isGlobal)
			}
		}
	}

	return g, nil
}

func (g GitConfigPhase) View() string {
	var b strings.Builder

	b.WriteString(SubtitleStyle.Render("Git Configuration"))
	b.WriteString("\n\n")

	switch g.step {
	case gcStepLoading:
		b.WriteString("  Loading current configuration...\n")

	case gcStepShowExisting:
		b.WriteString(InfoStyle.Render("  Current Git configuration:"))
		b.WriteString("\n\n")

		box := InnerBoxStyle.Render(
			fmt.Sprintf("%s  %s\n%s  %s",
				LabelStyle.Render("Name:"),
				ValueStyle.Render(g.existName),
				LabelStyle.Render("Email:"),
				ValueStyle.Render(g.existEmail),
			),
		)
		b.WriteString(box)
		b.WriteString("\n\n")
		b.WriteString(g.menu.View())

	case gcStepInputName:
		b.WriteString(PromptStyle.Render("  Enter your name:"))
		b.WriteString("\n\n")
		b.WriteString("  " + g.nameInput.View())
		b.WriteString("\n")
		if g.errorMsg != "" {
			b.WriteString("\n")
			b.WriteString(ErrorStyle.Render("  ✗ " + g.errorMsg))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render("  Press Enter to continue"))

	case gcStepInputEmail:
		b.WriteString(RenderCheckItem(true, "Name", g.newName))
		b.WriteString("\n\n")
		b.WriteString(PromptStyle.Render("  Enter your email:"))
		b.WriteString("\n\n")
		b.WriteString("  " + g.emailInput.View())
		b.WriteString("\n")
		if g.errorMsg != "" {
			b.WriteString("\n")
			b.WriteString(ErrorStyle.Render("  ✗ " + g.errorMsg))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render("  Press Enter to continue"))

	case gcStepChooseScope:
		b.WriteString(RenderCheckItem(true, "Name", g.newName))
		b.WriteString("\n")
		b.WriteString(RenderCheckItem(true, "Email", g.newEmail))
		b.WriteString("\n\n")
		b.WriteString(PromptStyle.Render("  Set configuration scope:"))
		b.WriteString("\n\n")
		b.WriteString(g.scopeMenu.View())

	case gcStepApplying:
		b.WriteString("  Applying configuration...\n")

	case gcStepDone:
		b.WriteString(RenderCheckItem(true, "Name", g.newName))
		b.WriteString("\n")
		b.WriteString(RenderCheckItem(true, "Email", g.newEmail))
		b.WriteString("\n\n")
		b.WriteString(SuccessStyle.Render("  ✓ Git configuration applied! Press Enter to continue."))
		b.WriteString("\n")
	}

	return b.String()
}

func (g GitConfigPhase) IsComplete() bool {
	return g.complete
}

func (g GitConfigPhase) ShouldExit() bool {
	return g.step == gcStepShowExisting && g.menu.SelectedItem() == "Exit"
}

func (g GitConfigPhase) Name() string  { return g.newName }
func (g GitConfigPhase) Email() string { return g.newEmail }

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
