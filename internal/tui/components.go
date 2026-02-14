package tui

import (
	"fmt"
	"strings"
)

// ─── Menu Component ─────────────────────────────────────────────────────────

// MenuModel is a simple single-select menu with arrow key navigation.
type MenuModel struct {
	Items    []string
	Cursor   int
	Selected int // -1 = nothing selected yet
}

func NewMenu(items []string) MenuModel {
	return MenuModel{
		Items:    items,
		Cursor:   0,
		Selected: -1,
	}
}

func (m *MenuModel) Up() {
	if m.Cursor > 0 {
		m.Cursor--
	}
}

func (m *MenuModel) Down() {
	if m.Cursor < len(m.Items)-1 {
		m.Cursor++
	}
}

func (m *MenuModel) Select() {
	m.Selected = m.Cursor
}

func (m *MenuModel) IsSelected() bool {
	return m.Selected >= 0
}

func (m *MenuModel) SelectedItem() string {
	if m.Selected >= 0 && m.Selected < len(m.Items) {
		return m.Items[m.Selected]
	}
	return ""
}

func (m MenuModel) View() string {
	var b strings.Builder
	for i, item := range m.Items {
		if i == m.Cursor {
			b.WriteString(SelectedStyle.Render(fmt.Sprintf("  ▸ %s", item)))
		} else {
			b.WriteString(UnselectedStyle.Render(fmt.Sprintf("    %s", item)))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// ─── Confirm Component ──────────────────────────────────────────────────────

// ConfirmModel is a yes/no prompt.
type ConfirmModel struct {
	Question string
	Options  []string
	Cursor   int
	Chosen   int // -1 = not chosen
}

func NewConfirm(question string, options ...string) ConfirmModel {
	if len(options) == 0 {
		options = []string{"Yes", "No"}
	}
	return ConfirmModel{
		Question: question,
		Options:  options,
		Cursor:   0,
		Chosen:   -1,
	}
}

func (c *ConfirmModel) Left() {
	if c.Cursor > 0 {
		c.Cursor--
	}
}

func (c *ConfirmModel) Right() {
	if c.Cursor < len(c.Options)-1 {
		c.Cursor++
	}
}

func (c *ConfirmModel) Select() {
	c.Chosen = c.Cursor
}

func (c *ConfirmModel) IsChosen() bool {
	return c.Chosen >= 0
}

func (c *ConfirmModel) ChosenOption() string {
	if c.Chosen >= 0 && c.Chosen < len(c.Options) {
		return c.Options[c.Chosen]
	}
	return ""
}

func (c ConfirmModel) View() string {
	var b strings.Builder
	b.WriteString(PromptStyle.Render(c.Question))
	b.WriteString("\n\n  ")

	for i, opt := range c.Options {
		if i > 0 {
			b.WriteString("    ")
		}
		if i == c.Cursor {
			b.WriteString(SelectedStyle.Render(fmt.Sprintf("[ %s ]", opt)))
		} else {
			b.WriteString(UnselectedStyle.Render(fmt.Sprintf("  %s  ", opt)))
		}
	}

	b.WriteString("\n")
	return b.String()
}

// ─── Status Indicator ───────────────────────────────────────────────────────

// StatusIcon returns a colored status icon.
func StatusIcon(ok bool) string {
	if ok {
		return CheckmarkStyle.Render("✓")
	}
	return CrossStyle.Render("✗")
}

// LoadingDots creates simple animated dots (call with frame count).
func LoadingDots(frame int) string {
	dots := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return SpinnerStyle.Render(dots[frame%len(dots)])
}
