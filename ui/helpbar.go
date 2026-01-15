package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gerund/tcr/ui/theme"
)

// HelpHint represents a single hint (key + description)
type HelpHint struct {
	Key  string
	Desc string
}

// Format renders a hint as "key desc" in uniform dim color
func (h HelpHint) Format() string {
	return theme.HelpDescStyle.Render(h.Key + " " + h.Desc)
}

// HelpBarContext captures the current UI state for help bar rendering
type HelpBarContext struct {
	ModalOpen bool // True if feedback modal is open
}

// getHints returns context-specific hints
func getHints(ctx HelpBarContext) []HelpHint {
	if ctx.ModalOpen {
		return []HelpHint{
			{Key: "enter", Desc: "save"},
			{Key: "esc", Desc: "cancel"},
		}
	}

	// Both panels always active with their own keys
	return []HelpHint{
		{Key: "up/dn", Desc: "file nav"},
		{Key: "C-n/C-p", Desc: "diff nav"},
		{Key: "enter", Desc: "feedback"},
		{Key: "q", Desc: "quit"},
	}
}

// formatHints joins hints with double spaces
func formatHints(hints []HelpHint) string {
	if len(hints) == 0 {
		return ""
	}

	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = h.Format()
	}
	return strings.Join(parts, "  ")
}

// RenderHelpBar renders the help bar at the bottom
func RenderHelpBar(ctx HelpBarContext, width int) string {
	hints := getHints(ctx)
	content := formatHints(hints)

	// Center the content
	contentWidth := lipgloss.Width(content)
	if contentWidth >= width {
		return theme.HelpBarStyle.Width(width).Render(content)
	}

	padding := (width - contentWidth) / 2
	paddedContent := strings.Repeat(" ", padding) + content

	return theme.HelpBarStyle.Width(width).Render(paddedContent)
}
