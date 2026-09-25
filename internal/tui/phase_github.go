package tui

import (
	"os/exec"
	"runtime"
	"strings"
	"time"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nameIess/git-tool/internal/clipboard"
	"github.com/nameIess/git-tool/internal/logger"
	"github.com/nameIess/git-tool/internal/runner"
	"github.com/nameIess/git-tool/internal/sshkey"
)

type pubKeyReadMsg struct{content string;err error}
type clipboardMsg struct{ok bool;err string}
type sshTestMsg struct{success bool;output string}
type githubStep int
const(ghStepLoading githubStep=iota;ghStepShowKey;ghStepWaitConfirm;ghStepTesting;ghStepVerified)
type GitHubPhase struct{step githubStep;keyPath,pubKey,clipErr,readErr,sshOutput string;clipOK bool;confirm ConfirmModel;openMenu MenuModel;complete bool}

func NewGitHubPhase(keyPath string)GitHubPhase{return GitHubPhase{step:ghStepLoading,keyPath:keyPath}}
func(g GitHubPhase)Init()tea.Cmd{return readPubKey(g.keyPath)}
func readPubKey(path string)tea.Cmd{return func()tea.Msg{content,err:=sshkey.ReadPublicKey(path);if err!=nil{return pubKeyReadMsg{err:err}};logger.Info("Read public key (%d bytes)",len(content));return pubKeyReadMsg{content:content}}}
func copyToClipboard(text string)tea.Cmd{return func()tea.Msg{if err:=clipboard.Copy(text);err!=nil{return clipboardMsg{err:err.Error()}};return clipboardMsg{ok:true}}}
func testGitHubSSH()tea.Cmd{return func()tea.Msg{res:=runner.RunWithTimeout(20*time.Second,"ssh","-T","-o","BatchMode=yes","-o","StrictHostKeyChecking=accept-new","git@github.com");out:=res.CombinedOutput();lower:=strings.ToLower(out);return sshTestMsg{success:strings.Contains(lower,"successfully authenticated")||strings.Contains(lower,"hi "),output:out}}}
func(g GitHubPhase)Update(msg tea.Msg)(GitHubPhase,tea.Cmd){
	switch msg:=msg.(type){
	case pubKeyReadMsg:if msg.err!=nil{g.readErr=msg.err.Error();g.step=ghStepShowKey;return g,nil};g.pubKey=msg.content;g.step=ghStepShowKey;return g,copyToClipboard(g.pubKey)
	case clipboardMsg:g.clipOK=msg.ok;g.clipErr=msg.err;g.openMenu=NewMenu([]string{"Open GitHub in browser","I'll add it manually"})
	case sshTestMsg:g.sshOutput=msg.output;if msg.success{g.step=ghStepVerified;g.complete=true;logger.Info("GitHub SSH authentication verified")}else{g.step=ghStepWaitConfirm;g.confirm=NewConfirm("GitHub SSH authentication is not verified","Retry test","Not yet")}
	case tea.KeyMsg:
		switch g.step{
		case ghStepShowKey:switch msg.String(){case "up","k":g.openMenu.Up();case "down","j":g.openMenu.Down();case "enter":g.openMenu.Select();if g.openMenu.SelectedItem()=="Open GitHub in browser"{openBrowser("https://github.com/settings/ssh/new")};g.step=ghStepWaitConfirm;g.confirm=NewConfirm("I've added the SSH key to GitHub","Test connection","Not yet")}
		case ghStepWaitConfirm:switch msg.String(){case "left","h":g.confirm.Left();case "right","l":g.confirm.Right();case "enter":g.confirm.Select();if g.confirm.ChosenOption()=="Test connection"||g.confirm.ChosenOption()=="Retry test"{g.step=ghStepTesting;return g,testGitHubSSH()}}
		}
	}
	return g,nil
}
func(g GitHubPhase)View()string{var b strings.Builder;b.WriteString(SubtitleStyle.Render("GitHub Integration"));b.WriteString("

");if g.readErr!=""{b.WriteString(ErrorStyle.Render("  ✗ Failed to read public key
"));b.WriteString(MutedStyle.Render("    "+g.readErr+"
"));return b.String()};switch g.step{case ghStepLoading:b.WriteString("  Reading public key...
");case ghStepShowKey,ghStepWaitConfirm:if g.clipOK{b.WriteString(SuccessStyle.Render("  ✓ Public key copied to clipboard!"))}else if g.clipErr!=""{b.WriteString(WarningStyle.Render("  ⚠ Could not copy to clipboard: "+g.clipErr))};b.WriteString("

");keyDisplay:=g.pubKey;if len(keyDisplay)>200{keyDisplay=keyDisplay[:80]+"
  "+keyDisplay[80:160]+"
  "+keyDisplay[160:]};b.WriteString(KeyBoxStyle.Render(MutedStyle.Render(keyDisplay)));b.WriteString("

");b.WriteString(InfoStyle.Render("  Add this SSH key to GitHub:"));b.WriteString("
  ");b.WriteString(AccentStyle.Render("→ https://github.com/settings/ssh/new"));b.WriteString("

");b.WriteString(MutedStyle.Render("  GitHub → Settings → SSH Keys → New SSH Key → Paste key → Add"));b.WriteString("

");if g.step==ghStepShowKey{b.WriteString(g.openMenu.View())}else{b.WriteString(g.confirm.View())};case ghStepTesting:b.WriteString("  Testing ssh -T git@github.com...
");case ghStepVerified:b.WriteString(RenderCheckItem(true,"GitHub SSH","Authenticated successfully"));b.WriteString("

");b.WriteString(SuccessStyle.Render("  ✓ GitHub SSH connection verified! Press Enter to continue."))};if g.sshOutput!=""&&g.step!=ghStepVerified{b.WriteString("
"+WarningStyle.Render("  SSH output: "+g.sshOutput)+"
")};return b.String()}
func(g GitHubPhase)IsComplete()bool{return g.complete}
func(g GitHubPhase)ShouldExit()bool{return false}
func openBrowser(url string){logger.Info("Opening browser: %s",url);if runtime.GOOS=="windows"{if err:=exec.Command("rundll32","url.dll,FileProtocolHandler",url).Start();err!=nil{logger.Error("Failed to open browser: %v",err)}}else{if err:=exec.Command("xdg-open",url).Start();err!=nil{logger.Error("Failed to open browser: %v",err)}}}
