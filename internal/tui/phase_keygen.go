package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/nameIess/git-tool/internal/logger"
	"github.com/nameIess/git-tool/internal/platform"
	"github.com/nameIess/git-tool/internal/sshkey"
)

type keygenResultMsg struct{ success bool; keyPath, errMsg string }
type keygenStep int
const(
	kgStepCheckExisting keygenStep=iota
	kgStepExistingFound
	kgStepAskPassphrase
	kgStepInputPassphrase
	kgStepConfirmPassphrase
	kgStepGenerating
	kgStepSuccess
	kgStepFailed
)

type KeygenPhase struct{step keygenStep;spinner spinner.Model;email,keyPath string;overwriteMenu MenuModel;passphraseYN ConfirmModel;passInput,passConfirm textinput.Model;passphrase,errMsg string;failMenu MenuModel;complete bool}
func NewKeygenPhase(email string)KeygenPhase{
	s:=spinner.New();s.Spinner=spinner.Dot;s.Style=SpinnerStyle
	pi:=textinput.New();pi.Placeholder="Enter passphrase (or leave empty)";pi.EchoMode=textinput.EchoPassword;pi.EchoCharacter='•';pi.CharLimit=200;pi.Width=40;pi.PromptStyle=PromptStyle;pi.TextStyle=TextStyle
	pc:=textinput.New();pc.Placeholder="Confirm passphrase";pc.EchoMode=textinput.EchoPassword;pc.EchoCharacter='•';pc.CharLimit=200;pc.Width=40;pc.PromptStyle=PromptStyle;pc.TextStyle=TextStyle
	return KeygenPhase{step:kgStepCheckExisting,spinner:s,email:email,keyPath:platform.DefaultKeyPath(),passphraseYN:NewConfirm("Add a passphrase to your SSH key? (recommended)","Yes","No"),passInput:pi,passConfirm:pc}
}
func(k KeygenPhase)Init()tea.Cmd{return checkExistingKey(k.keyPath)}
func checkExistingKey(path string)tea.Cmd{return func()tea.Msg{time.Sleep(300*time.Millisecond);if sshkey.Exists(path){logger.Warn("Existing SSH key found at %s",path);return keyExistsMsg{path:path}};return keyNotExistsMsg{}}}
type keyExistsMsg struct{path string}
type keyNotExistsMsg struct{}
func configureKey(path string)error{return sshkey.WriteSSHConfig(path)}
func generateSSHKey(path,email,pass string)tea.Cmd{return func()tea.Msg{if err:=sshkey.Generate(path,email,pass);err!=nil{return keygenResultMsg{errMsg:err.Error()}};if err:=configureKey(path);err!=nil{return keygenResultMsg{errMsg:fmt.Sprintf("SSH key generated, but SSH config failed: %v",err)}};return keygenResultMsg{success:true,keyPath:path}}}
func prepareExistingKey(path string)tea.Cmd{return func()tea.Msg{if err:=configureKey(path);err!=nil{return keygenResultMsg{errMsg:fmt.Sprintf("Could not configure SSH to use existing key: %v",err)}};return keygenResultMsg{success:true,keyPath:path}}}
func(k KeygenPhase)Update(msg tea.Msg)(KeygenPhase,tea.Cmd){
	switch msg:=msg.(type){
	case keyExistsMsg:k.step=kgStepExistingFound;k.overwriteMenu=NewMenu([]string{"Overwrite","Use different name","Skip (use existing)"})
	case keyNotExistsMsg:k.step=kgStepAskPassphrase
	case keygenResultMsg:
		if msg.success{k.step=kgStepSuccess;k.keyPath=msg.keyPath;k.complete=true}else{k.step=kgStepFailed;k.errMsg=msg.errMsg;k.failMenu=NewMenu([]string{"Retry","Exit"})};return k,nil
	case spinner.TickMsg:if k.step==kgStepCheckExisting||k.step==kgStepGenerating{var cmd tea.Cmd;k.spinner,cmd=k.spinner.Update(msg);return k,cmd}
	case tea.KeyMsg:
		switch k.step{
		case kgStepExistingFound:
			switch msg.String(){case "up","k":k.overwriteMenu.Up();case "down","j":k.overwriteMenu.Down();case "enter":k.overwriteMenu.Select();switch k.overwriteMenu.SelectedItem(){case "Overwrite":k.step=kgStepAskPassphrase;case "Use different name":k.keyPath=findNextKeyName(k.keyPath);k.step=kgStepAskPassphrase;case "Skip (use existing)":k.step=kgStepGenerating;return k,tea.Batch(k.spinner.Tick,prepareExistingKey(k.keyPath))}}
		case kgStepAskPassphrase:
			switch msg.String(){case "left","h":k.passphraseYN.Left();case "right","l":k.passphraseYN.Right();case "enter":k.passphraseYN.Select();if k.passphraseYN.ChosenOption()=="Yes"{k.step=kgStepInputPassphrase;k.passInput.Focus();return k,textinput.Blink};k.step=kgStepGenerating;return k,tea.Batch(k.spinner.Tick,generateSSHKey(k.keyPath,k.email,""))}
		case kgStepInputPassphrase:
			switch msg.String(){case "enter":k.passphrase=k.passInput.Value();if len(k.passphrase)<5{k.errMsg="Passphrase must be at least 5 characters";return k,nil};k.errMsg="";k.step=kgStepConfirmPassphrase;k.passInput.Blur();k.passConfirm.Focus();return k,textinput.Blink;default:var cmd tea.Cmd;k.passInput,cmd=k.passInput.Update(msg);return k,cmd}
		case kgStepConfirmPassphrase:
			switch msg.String(){case "enter":if k.passConfirm.Value()!=k.passphrase{k.errMsg="Passphrases do not match";return k,nil};k.errMsg="";k.step=kgStepGenerating;k.passConfirm.Blur();return k,tea.Batch(k.spinner.Tick,generateSSHKey(k.keyPath,k.email,k.passphrase));default:var cmd tea.Cmd;k.passConfirm,cmd=k.passConfirm.Update(msg);return k,cmd}
		case kgStepFailed:
			switch msg.String(){case "up","k":k.failMenu.Up();case "down","j":k.failMenu.Down();case "enter":k.failMenu.Select();if k.failMenu.SelectedItem()=="Retry"{k.step=kgStepGenerating;return k,tea.Batch(k.spinner.Tick,generateSSHKey(k.keyPath,k.email,k.passphrase))}}
		}
	}
	return k,nil
}
func(k KeygenPhase)View()string{
	var b strings.Builder;b.WriteString(SubtitleStyle.Render("SSH Key Generation"));b.WriteString("\n\n")
	switch k.step{
	case kgStepCheckExisting:b.WriteString(fmt.Sprintf("  %s Checking for existing SSH keys...",k.spinner.View()))
	case kgStepExistingFound:b.WriteString(WarningStyle.Render("  ⚠ Existing SSH key found:"));b.WriteString("\n");b.WriteString(MutedStyle.Render("    "+k.keyPath));b.WriteString("\n\n");b.WriteString(k.overwriteMenu.View())
	case kgStepAskPassphrase:b.WriteString(InfoStyle.Render("  Key type: ed25519 (modern & secure)"));b.WriteString("\n");b.WriteString(MutedStyle.Render("  Location: "+k.keyPath));b.WriteString("\n\n");b.WriteString(k.passphraseYN.View())
	case kgStepInputPassphrase:b.WriteString(PromptStyle.Render("  Enter passphrase:"));b.WriteString("\n\n  "+k.passInput.View()+"\n");if k.errMsg!=""{b.WriteString(ErrorStyle.Render("  ✗ "+k.errMsg));b.WriteString("\n")}
	case kgStepConfirmPassphrase:b.WriteString(PromptStyle.Render("  Confirm passphrase:"));b.WriteString("\n\n  "+k.passConfirm.View()+"\n");if k.errMsg!=""{b.WriteString(ErrorStyle.Render("  ✗ "+k.errMsg));b.WriteString("\n")}
	case kgStepGenerating:b.WriteString(fmt.Sprintf("  %s Generating SSH key...",k.spinner.View()))
	case kgStepSuccess:b.WriteString(RenderCheckItem(true,"SSH Key",k.keyPath));b.WriteString("\n");b.WriteString(RenderCheckItem(true,"SSH Config","Managed github.com entry"));b.WriteString("\n\n");b.WriteString(SuccessStyle.Render("  ✓ SSH key configured! Press Enter to continue."))
	case kgStepFailed:b.WriteString(ErrorStyle.Render("  ✗ Failed to configure SSH key"));b.WriteString("\n\n");b.WriteString(InnerBoxStyle.Render(k.errMsg));b.WriteString("\n\n");b.WriteString(k.failMenu.View())
	}
	return b.String()
}
func(k KeygenPhase)IsComplete()bool{return k.complete}
func(k KeygenPhase)ShouldExit()bool{return k.step==kgStepFailed&&k.failMenu.SelectedItem()=="Exit"}
func(k KeygenPhase)KeyPath()string{return k.keyPath}
func findNextKeyName(basePath string)string{return sshkey.FindNextKeyName(basePath)}
