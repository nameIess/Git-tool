package tui

import("fmt";"strings";tea "github.com/charmbracelet/bubbletea";"github.com/charmbracelet/lipgloss")

type CompletePhase struct{gitName,gitEmail,keyPath,logPath string;exiting bool}
func NewCompletePhase(name,email,keyPath,logPath string)CompletePhase{return CompletePhase{gitName:name,gitEmail:email,keyPath:keyPath,logPath:logPath}}
func(c CompletePhase)Init()tea.Cmd{return nil}
func(c CompletePhase)Update(msg tea.Msg)(CompletePhase,tea.Cmd){if msg,ok:=msg.(tea.KeyMsg);ok&&(msg.String()=="enter"||msg.String()=="q"||msg.String()=="esc"){c.exiting=true};return c,nil}
func(c CompletePhase)View()string{var b strings.Builder;banner:=lipgloss.NewStyle().Bold(true).Foreground(ColorText).Background(ColorSuccess).Padding(0,3).Render(" ✓ Setup Complete ");b.WriteString("\n"+banner+"\n\n");b.WriteString(SubtitleStyle.Render("Verified Summary"));b.WriteString("\n\n");b.WriteString(RenderCheckItem(true,"Git Name",c.gitName)+"\n");b.WriteString(RenderCheckItem(true,"Git Email",c.gitEmail)+"\n");b.WriteString(RenderCheckItem(true,"SSH Key",c.keyPath)+"\n");b.WriteString(RenderCheckItem(true,"SSH Config","github.com managed entry")+"\n");b.WriteString(RenderCheckItem(true,"SSH Agent","Expected key loaded")+"\n");b.WriteString(RenderCheckItem(true,"GitHub SSH","Authenticated successfully")+"\n\n");b.WriteString(MutedStyle.Render("  Commit signing: optional and configured only if enabled.")+"\n\n");if c.logPath!=""{b.WriteString(MutedStyle.Render(fmt.Sprintf("  Log file: %s",c.logPath))+"\n\n")};b.WriteString(MutedStyle.Render("  Press Enter or ESC to exit.")+"\n");return b.String()}
func(c CompletePhase)IsComplete()bool{return c.exiting}
func(c CompletePhase)ShouldExit()bool{return c.exiting}
