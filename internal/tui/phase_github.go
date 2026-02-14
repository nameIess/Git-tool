package tui

import (
	"os"
	osExec "os/exec"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/user/git-tool/internal/logger"
)

// ─── Messages ───────────────────────────────────────────────────────────────

type pubKeyReadMsg struct {
	content string
	err     error
}

type clipboardMsg struct {
	ok  bool
	err string
}

// ─── State ──────────────────────────────────────────────────────────────────

type githubStep int

const (
	ghStepLoading githubStep = iota
	ghStepShowKey
	ghStepWaitConfirm
)

// ─── Model ──────────────────────────────────────────────────────────────────

type GitHubPhase struct {
	step      githubStep
	keyPath   string
	pubKey    string
	clipOK    bool
	clipErr   string
	readErr   string
	confirm   ConfirmModel
	openMenu  MenuModel
	complete  bool
	scrollOff int
}

func NewGitHubPhase(keyPath string) GitHubPhase {
	return GitHubPhase{
		step:    ghStepLoading,
		keyPath: keyPath + ".pub",
	}
}

func (g GitHubPhase) Init() tea.Cmd {
	return readPubKey(g.keyPath)
}

func readPubKey(path string) tea.Cmd {
	return func() tea.Msg {
		data, err := os.ReadFile(path)
		if err != nil {
			logger.Error("Failed to read public key: %v", err)
			return pubKeyReadMsg{err: err}
		}
		content := strings.TrimSpace(string(data))
		logger.Info("Read public key from %s (%d bytes)", path, len(content))

		return pubKeyReadMsg{content: content}
	}
}

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(text)
		if err != nil {
			logger.Error("Failed to copy to clipboard: %v", err)
			return clipboardMsg{ok: false, err: err.Error()}
		}
		logger.Info("Public key copied to clipboard")
		return clipboardMsg{ok: true}
	}
}

func (g GitHubPhase) Update(msg tea.Msg) (GitHubPhase, tea.Cmd) {
	switch msg := msg.(type) {
	case pubKeyReadMsg:
		if msg.err != nil {
			g.readErr = msg.err.Error()
			g.step = ghStepShowKey
			return g, nil
		}
		g.pubKey = msg.content
		g.step = ghStepShowKey
		return g, copyToClipboard(g.pubKey)

	case clipboardMsg:
		g.clipOK = msg.ok
		g.clipErr = msg.err
		g.openMenu = NewMenu([]string{"Open GitHub in browser", "I'll add it manually"})
		return g, nil

	case tea.KeyMsg:
		switch g.step {
		case ghStepShowKey:
			switch msg.String() {
			case "up", "k":
				g.openMenu.Up()
			case "down", "j":
				g.openMenu.Down()
			case "enter":
				g.openMenu.Select()
				if g.openMenu.SelectedItem() == "Open GitHub in browser" {
					openBrowser("https://github.com/settings/ssh/new")
				}
				g.step = ghStepWaitConfirm
				g.confirm = NewConfirm("I've added the SSH key to GitHub", "Yes, continue", "Not yet")
			}

		case ghStepWaitConfirm:
			switch msg.String() {
			case "left", "h":
				g.confirm.Left()
			case "right", "l":
				g.confirm.Right()
			case "enter":
				g.confirm.Select()
				if g.confirm.ChosenOption() == "Yes, continue" {
					g.complete = true
					logger.Info("User confirmed key added to GitHub")
				} else {
					// Reset confirm
					g.confirm = NewConfirm("I've added the SSH key to GitHub", "Yes, continue", "Not yet")
				}
			}
		}
	}

	return g, nil
}

func (g GitHubPhase) View() string {
	var b strings.Builder

	b.WriteString(SubtitleStyle.Render("GitHub Integration"))
	b.WriteString("\n\n")

	if g.readErr != "" {
		b.WriteString(ErrorStyle.Render("  ✗ Failed to read public key"))
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render("    " + g.readErr))
		b.WriteString("\n")
		return b.String()
	}

	switch g.step {
	case ghStepLoading:
		b.WriteString("  Reading public key...\n")

	case ghStepShowKey, ghStepWaitConfirm:
		// Show clipboard status
		if g.clipOK {
			b.WriteString(SuccessStyle.Render("  ✓ Public key copied to clipboard!"))
		} else if g.clipErr != "" {
			b.WriteString(WarningStyle.Render("  ⚠ Could not copy to clipboard: " + g.clipErr))
		}
		b.WriteString("\n\n")

		// Show the key in a bordered box
		keyDisplay := g.pubKey
		if len(keyDisplay) > 200 {
			keyDisplay = keyDisplay[:80] + "\n  " + keyDisplay[80:160] + "\n  " + keyDisplay[160:]
		}
		b.WriteString(KeyBoxStyle.Render(MutedStyle.Render(keyDisplay)))
		b.WriteString("\n\n")

		// Instructions
		b.WriteString(InfoStyle.Render("  Add this SSH key to GitHub:"))
		b.WriteString("\n")
		b.WriteString(AccentStyle.Render("  → https://github.com/settings/ssh/new"))
		b.WriteString("\n\n")

		b.WriteString(MutedStyle.Render("  Steps: GitHub → Settings → SSH Keys → New SSH Key →"))
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render("         Paste key → Add SSH Key"))
		b.WriteString("\n\n")

		if g.step == ghStepShowKey {
			b.WriteString(g.openMenu.View())
		} else {
			b.WriteString(g.confirm.View())
		}
	}

	return b.String()
}

func (g GitHubPhase) IsComplete() bool {
	return g.complete
}

func (g GitHubPhase) ShouldExit() bool {
	return false // No exit option here
}

func openBrowser(url string) {
	logger.Info("Opening browser: %s", url)
	if runtime.GOOS == "windows" {
		cmd := osExec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		if err := cmd.Start(); err != nil {
			logger.Error("Failed to open browser: %v", err)
		}
	} else {
		cmd := osExec.Command("xdg-open", url)
		if err := cmd.Start(); err != nil {
			logger.Error("Failed to open browser: %v", err)
		}
	}
}
