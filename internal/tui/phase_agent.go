package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/nameIess/git-tool/internal/agent"
	"github.com/nameIess/git-tool/internal/platform"
	"github.com/nameIess/git-tool/internal/signing"
)

type agentResultMsg struct{agentStarted bool;keyAdded bool;errMsg string}
type agentStep int
const(asStepStarting agentStep=iota;asStepAskSigning;asStepSuccess;asStepFailed)

type AgentPhase struct{step agentStep;spinner spinner.Model;keyPath,errMsg string;failMenu MenuModel;complete bool;shellType platform.ShellType;signingYN ConfirmModel}
func NewAgentPhase(keyPath string)AgentPhase{s:=spinner.New();s.Spinner=spinner.Dot;s.Style=SpinnerStyle;return AgentPhase{step:asStepStarting,spinner:s,keyPath:keyPath,shellType:platform.Current(),signingYN:NewConfirm("Enable SSH commit/tag signing?","Yes","No")}}
func(a AgentPhase)Init()tea.Cmd{return tea.Batch(a.spinner.Tick,startAgentAndAddKey(a.keyPath,a.shellType))}
func startAgentAndAddKey(keyPath string,shellType platform.ShellType)tea.Cmd{return func()tea.Msg{time.Sleep(500*time.Millisecond);if err:=agent.StartAndAddKey(keyPath,shellType);err!=nil{return agentResultMsg{errMsg:err.Error()}};if shellType==platform.GitBash{if err:=agent.WriteBashrcSnippet(keyPath);err!=nil{return agentResultMsg{agentStarted:true,keyAdded:true,errMsg:fmt.Sprintf("SSH agent is ready, but .bashrc update failed: %v",err)}}};return agentResultMsg{agentStarted:true,keyAdded:true}}}
func(a AgentPhase)Update(msg tea.Msg)(AgentPhase,tea.Cmd){
	switch msg:=msg.(type){
	case agentResultMsg:
		if msg.agentStarted&&msg.keyAdded&&msg.errMsg==""{a.step=asStepAskSigning}else{a.step=asStepFailed;a.errMsg=msg.errMsg;if a.errMsg==""{a.errMsg="Required SSH-agent setup could not be verified"};a.failMenu=NewMenu([]string{"Retry","Continue anyway","Exit"})}
	case tea.KeyMsg:
		switch a.step{
		case asStepAskSigning:
			switch msg.String(){case "left","h":a.signingYN.Left();case "right","l":a.signingYN.Right();case "enter":a.signingYN.Select();if a.signingYN.ChosenOption()=="Yes"{if err:=signing.Configure(a.keyPath,true);err!=nil{a.errMsg="SSH agent is ready, but commit signing failed: "+err.Error();a.step=asStepFailed;a.failMenu=NewMenu([]string{"Retry","Continue without signing","Exit"})}else{a.step=asStepSuccess;a.complete=true}}else{a.step=asStepSuccess;a.complete=true}}
		case asStepFailed:
			switch msg.String(){case "up","k":a.failMenu.Up();case "down","j":a.failMenu.Down();case "enter":a.failMenu.Select();switch a.failMenu.SelectedItem(){case "Retry":a.step=asStepStarting;a.failMenu=MenuModel{};return a,tea.Batch(a.spinner.Tick,startAgentAndAddKey(a.keyPath,a.shellType));case "Continue anyway","Continue without signing":a.complete=true;a.step=asStepSuccess}}
		}
	case spinner.TickMsg:if a.step==asStepStarting{var cmd tea.Cmd;a.spinner,cmd=a.spinner.Update(msg);return a,cmd}
	}
	return a,nil
}
func(a AgentPhase)View()string{var b strings.Builder;b.WriteString(SubtitleStyle.Render("SSH Agent Setup"));b.WriteString("

");b.WriteString(MutedStyle.Render(fmt.Sprintf("  Environment: %s",a.shellType)));b.WriteString("

");switch a.step{case asStepStarting:b.WriteString(fmt.Sprintf("  %s Starting ssh-agent and adding key...",a.spinner.View()));case asStepAskSigning:b.WriteString(RenderCheckItem(true,"SSH Agent","Expected key loaded"));b.WriteString("

");b.WriteString(a.signingYN.View());case asStepSuccess:b.WriteString(RenderCheckItem(true,"SSH Agent","Expected key loaded"));b.WriteString("

");b.WriteString(SuccessStyle.Render("  ✓ Required SSH-agent checks passed. Press Enter to continue."));case asStepFailed:b.WriteString(ErrorStyle.Render("  ⚠ Required setup is not fully verified"));b.WriteString("

");b.WriteString(InnerBoxStyle.Render(a.errMsg));b.WriteString("

");b.WriteString(a.failMenu.View())};return b.String()}
func(a AgentPhase)IsComplete()bool{return a.complete}
func(a AgentPhase)ShouldExit()bool{return a.step==asStepFailed&&a.failMenu.SelectedItem()=="Exit"}
