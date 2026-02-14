package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ─── Color Palette ──────────────────────────────────────────────────────────

var (
	ColorPrimary   = lipgloss.Color("#7C3AED") // Vibrant purple
	ColorSecondary = lipgloss.Color("#06B6D4") // Cyan
	ColorSuccess   = lipgloss.Color("#10B981") // Emerald green
	ColorError     = lipgloss.Color("#EF4444") // Red
	ColorWarning   = lipgloss.Color("#F59E0B") // Amber
	ColorInfo       = lipgloss.Color("#3B82F6") // Blue
	ColorMuted     = lipgloss.Color("#6B7280") // Gray
	ColorText      = lipgloss.Color("#F9FAFB") // Almost white
	ColorSubtext   = lipgloss.Color("#9CA3AF") // Light gray
	ColorBg        = lipgloss.Color("#111827") // Dark bg
	ColorBgAlt     = lipgloss.Color("#1F2937") // Slightly lighter bg
	ColorAccent    = lipgloss.Color("#A78BFA") // Light purple
	ColorHighlight = lipgloss.Color("#34D399") // Teal green
)

// ─── Reusable Styles ────────────────────────────────────────────────────────

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText).
			Background(ColorPrimary).
			Padding(0, 2).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent).
			MarginBottom(1)

	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorError)

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWarning)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	PromptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary)

	AccentStyle = lipgloss.NewStyle().
			Foreground(ColorAccent)

	TextStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	InnerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 2)

	KeyBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	HelpBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	UnselectedStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	CheckmarkStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	CrossStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	ProgressDotActive = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	ProgressDotDone = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	ProgressDotPending = lipgloss.NewStyle().
				Foreground(ColorMuted)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Width(14)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true)
)

// ─── Phase Header ───────────────────────────────────────────────────────────

var phaseNames = []string{
	"Prerequisites",
	"Git Config",
	"SSH Key",
	"SSH Agent",
	"GitHub",
	"Connection",
	"Complete",
}

// RenderHeader renders a phase header with progress dots.
func RenderHeader(currentPhase int) string {
	// Progress dots
	dots := ""
	for i := 0; i < len(phaseNames); i++ {
		if i > 0 {
			dots += MutedStyle.Render(" ─ ")
		}
		label := fmt.Sprintf("%d", i+1)
		if i < currentPhase {
			dots += ProgressDotDone.Render("✓")
		} else if i == currentPhase {
			dots += ProgressDotActive.Render("●") + " " + AccentStyle.Render(phaseNames[i])
		} else {
			dots += ProgressDotPending.Render(label)
		}
	}

	title := TitleStyle.Render(fmt.Sprintf(" ⚙  Git & SSH Setup Tool "))
	return title + "\n" + dots + "\n"
}

// RenderHelpBar renders the bottom help bar.
func RenderHelpBar(extra string) string {
	help := "ESC quit"
	if extra != "" {
		help = extra + "  │  " + help
	}
	return HelpBarStyle.Render("  " + help)
}

// RenderCheckItem renders a summary line with check or cross.
func RenderCheckItem(ok bool, label, value string) string {
	icon := CheckmarkStyle.Render("✓")
	if !ok {
		icon = CrossStyle.Render("✗")
	}
	return fmt.Sprintf("  %s %s %s", icon, LabelStyle.Render(label), ValueStyle.Render(value))
}
