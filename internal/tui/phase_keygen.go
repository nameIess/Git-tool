package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/user/git-tool/internal/exec"
	"github.com/user/git-tool/internal/logger"
	"github.com/user/git-tool/internal/shell"
)

// ─── Messages ───────────────────────────────────────────────────────────────

type keygenResultMsg struct {
	success bool
	keyPath string
	errMsg  string
}

// ─── State ──────────────────────────────────────────────────────────────────

type keygenStep int

const (
	kgStepCheckExisting keygenStep = iota
	kgStepExistingFound
	kgStepAskPassphrase
	kgStepInputPassphrase
	kgStepConfirmPassphrase
	kgStepGenerating
	kgStepSuccess
	kgStepFailed
)

// ─── Model ──────────────────────────────────────────────────────────────────

type KeygenPhase struct {
	step    keygenStep
	spinner spinner.Model
	email   string

	keyPath       string
	existingFound bool

	overwriteMenu MenuModel

	passphraseYN ConfirmModel
	wantPass     bool
	passInput    textinput.Model
	passConfirm  textinput.Model
	passphrase   string

	errMsg   string
	failMenu MenuModel
	complete bool
}

func NewKeygenPhase(email string) KeygenPhase {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	pi := textinput.New()
	pi.Placeholder = "Enter passphrase (or leave empty)"
	pi.EchoMode = textinput.EchoPassword
	pi.EchoCharacter = '•'
	pi.CharLimit = 200
	pi.Width = 40
	pi.PromptStyle = PromptStyle
	pi.TextStyle = TextStyle

	pc := textinput.New()
	pc.Placeholder = "Confirm passphrase"
	pc.EchoMode = textinput.EchoPassword
	pc.EchoCharacter = '•'
	pc.CharLimit = 200
	pc.Width = 40
	pc.PromptStyle = PromptStyle
	pc.TextStyle = TextStyle

	return KeygenPhase{
		step:    kgStepCheckExisting,
		spinner: s,
		email:   email,
		keyPath: shell.DefaultKeyPath(),

		passphraseYN: NewConfirm("Add a passphrase to your SSH key? (recommended)", "Yes", "No"),
		passInput:    pi,
		passConfirm:  pc,
	}
}

func (k KeygenPhase) Init() tea.Cmd {
	return checkExistingKey(k.keyPath)
}

func checkExistingKey(keyPath string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(300 * time.Millisecond)
		if _, err := os.Stat(keyPath); err == nil {
			logger.Warn("Existing SSH key found at %s", keyPath)
			return keyExistsMsg{path: keyPath}
		}
		logger.Info("No existing SSH key at %s", keyPath)
		return keyNotExistsMsg{}
	}
}

type keyExistsMsg struct{ path string }
type keyNotExistsMsg struct{}

func generateSSHKey(keyPath, email, passphrase string) tea.Cmd {
	return func() tea.Msg {
		logger.Info("Generating SSH key: type=ed25519 email=%s path=%s", email, keyPath)

		// Ensure .ssh directory exists
		if err := shell.EnsureSSHDir(); err != nil {
			logger.Error("Failed to create .ssh directory: %v", err)
			return keygenResultMsg{success: false, errMsg: fmt.Sprintf("Failed to create .ssh directory: %v", err)}
		}

		// Build ssh-keygen command
		args := []string{"-t", "ed25519", "-C", email, "-f", keyPath}
		if passphrase == "" {
			args = append(args, "-N", "")
		} else {
			args = append(args, "-N", passphrase)
		}

		result := exec.Run("ssh-keygen", args...)
		if result.Success() {
			logger.Info("SSH key generated successfully at %s", keyPath)
			return keygenResultMsg{success: true, keyPath: keyPath}
		}

		errMsg := strings.TrimSpace(result.CombinedOutput())
		if errMsg == "" {
			errMsg = fmt.Sprintf("ssh-keygen exited with code %d", result.ExitCode)
		}
		logger.Error("SSH key generation failed: %s", errMsg)
		return keygenResultMsg{success: false, errMsg: errMsg}
	}
}

func (k KeygenPhase) Update(msg tea.Msg) (KeygenPhase, tea.Cmd) {
	switch msg := msg.(type) {
	case keyExistsMsg:
		k.existingFound = true
		k.step = kgStepExistingFound
		k.overwriteMenu = NewMenu([]string{"Overwrite", "Use different name", "Skip (use existing)"})
		return k, nil

	case keyNotExistsMsg:
		k.step = kgStepAskPassphrase
		return k, nil

	case keygenResultMsg:
		if msg.success {
			k.step = kgStepSuccess
			k.keyPath = msg.keyPath
		} else {
			k.step = kgStepFailed
			k.errMsg = msg.errMsg
			k.failMenu = NewMenu([]string{"Retry", "Exit"})
		}
		return k, nil

	case spinner.TickMsg:
		if k.step == kgStepCheckExisting || k.step == kgStepGenerating {
			var cmd tea.Cmd
			k.spinner, cmd = k.spinner.Update(msg)
			return k, cmd
		}

	case tea.KeyMsg:
		switch k.step {
		case kgStepExistingFound:
			switch msg.String() {
			case "up", "k":
				k.overwriteMenu.Up()
			case "down", "j":
				k.overwriteMenu.Down()
			case "enter":
				k.overwriteMenu.Select()
				switch k.overwriteMenu.SelectedItem() {
				case "Overwrite":
					k.step = kgStepAskPassphrase
				case "Use different name":
					// Find next available name
					k.keyPath = findNextKeyName(k.keyPath)
					k.step = kgStepAskPassphrase
				case "Skip (use existing)":
					k.step = kgStepSuccess
					k.complete = true
				}
			}

		case kgStepAskPassphrase:
			switch msg.String() {
			case "left", "h":
				k.passphraseYN.Left()
			case "right", "l":
				k.passphraseYN.Right()
			case "enter":
				k.passphraseYN.Select()
				if k.passphraseYN.ChosenOption() == "Yes" {
					k.wantPass = true
					k.step = kgStepInputPassphrase
					k.passInput.Focus()
					return k, textinput.Blink
				}
				// No passphrase, generate immediately
				k.step = kgStepGenerating
				return k, tea.Batch(k.spinner.Tick, generateSSHKey(k.keyPath, k.email, ""))
			}

		case kgStepInputPassphrase:
			switch msg.String() {
			case "enter":
				k.passphrase = k.passInput.Value()
				if len(k.passphrase) < 5 {
					k.errMsg = "Passphrase must be at least 5 characters"
					return k, nil
				}
				k.errMsg = ""
				k.step = kgStepConfirmPassphrase
				k.passInput.Blur()
				k.passConfirm.Focus()
				return k, textinput.Blink
			default:
				var cmd tea.Cmd
				k.passInput, cmd = k.passInput.Update(msg)
				return k, cmd
			}

		case kgStepConfirmPassphrase:
			switch msg.String() {
			case "enter":
				if k.passConfirm.Value() != k.passphrase {
					k.errMsg = "Passphrases do not match"
					return k, nil
				}
				k.errMsg = ""
				k.step = kgStepGenerating
				k.passConfirm.Blur()
				return k, tea.Batch(k.spinner.Tick, generateSSHKey(k.keyPath, k.email, k.passphrase))
			default:
				var cmd tea.Cmd
				k.passConfirm, cmd = k.passConfirm.Update(msg)
				return k, cmd
			}

		case kgStepFailed:
			switch msg.String() {
			case "up", "k":
				k.failMenu.Up()
			case "down", "j":
				k.failMenu.Down()
			case "enter":
				k.failMenu.Select()
				if k.failMenu.SelectedItem() == "Retry" {
					k.step = kgStepGenerating
					return k, tea.Batch(k.spinner.Tick, generateSSHKey(k.keyPath, k.email, k.passphrase))
				}
			}
		}
	}

	return k, nil
}

func (k KeygenPhase) View() string {
	var b strings.Builder

	b.WriteString(SubtitleStyle.Render("SSH Key Generation"))
	b.WriteString("\n\n")

	switch k.step {
	case kgStepCheckExisting:
		b.WriteString(fmt.Sprintf("  %s Checking for existing SSH keys...\n", k.spinner.View()))

	case kgStepExistingFound:
		b.WriteString(WarningStyle.Render("  ⚠ Existing SSH key found:"))
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render(fmt.Sprintf("    %s", k.keyPath)))
		b.WriteString("\n\n")
		b.WriteString(k.overwriteMenu.View())

	case kgStepAskPassphrase:
		b.WriteString(InfoStyle.Render("  Key type: ed25519 (modern & secure)"))
		b.WriteString("\n")
		b.WriteString(MutedStyle.Render(fmt.Sprintf("  Location: %s", k.keyPath)))
		b.WriteString("\n\n")
		b.WriteString(k.passphraseYN.View())

	case kgStepInputPassphrase:
		b.WriteString(PromptStyle.Render("  Enter passphrase:"))
		b.WriteString("\n\n")
		b.WriteString("  " + k.passInput.View())
		b.WriteString("\n")
		if k.errMsg != "" {
			b.WriteString(ErrorStyle.Render("  ✗ " + k.errMsg))
			b.WriteString("\n")
		}

	case kgStepConfirmPassphrase:
		b.WriteString(PromptStyle.Render("  Confirm passphrase:"))
		b.WriteString("\n\n")
		b.WriteString("  " + k.passConfirm.View())
		b.WriteString("\n")
		if k.errMsg != "" {
			b.WriteString(ErrorStyle.Render("  ✗ " + k.errMsg))
			b.WriteString("\n")
		}

	case kgStepGenerating:
		b.WriteString(fmt.Sprintf("  %s Generating SSH key...\n", k.spinner.View()))

	case kgStepSuccess:
		b.WriteString(RenderCheckItem(true, "SSH Key", k.keyPath))
		b.WriteString("\n")
		b.WriteString(RenderCheckItem(true, "Public Key", k.keyPath+".pub"))
		b.WriteString("\n\n")
		b.WriteString(SuccessStyle.Render("  ✓ SSH key generated! Press Enter to continue."))
		b.WriteString("\n")

	case kgStepFailed:
		b.WriteString(ErrorStyle.Render("  ✗ Failed to generate SSH key"))
		b.WriteString("\n\n")
		b.WriteString(InnerBoxStyle.Render(k.errMsg))
		b.WriteString("\n\n")
		b.WriteString(k.failMenu.View())
	}

	return b.String()
}

func (k KeygenPhase) IsComplete() bool {
	return k.step == kgStepSuccess
}

func (k KeygenPhase) ShouldExit() bool {
	return k.step == kgStepFailed && k.failMenu.SelectedItem() == "Exit"
}

func (k KeygenPhase) KeyPath() string {
	return k.keyPath
}

func findNextKeyName(basePath string) string {
	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)
	for i := 2; i <= 99; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d", base, i))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	return basePath + "_new"
}
