package floating

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gerund/tcr/ui/borders"
	"github.com/gerund/tcr/ui/theme"
)

// FeedbackSavedMsg is sent when feedback is saved
type FeedbackSavedMsg struct {
	FilePath   string
	LineNumber int
	Comment    string
}

// FeedbackCancelledMsg is sent when feedback is cancelled
type FeedbackCancelledMsg struct{}

// FeedbackModal is a floating window for entering feedback
type FeedbackModal struct {
	textarea    textarea.Model
	filePath    string
	lineNumber  int
	lineContent string
	width       int
	height      int
	ready       bool
}

// NewFeedbackModal creates a new feedback modal
func NewFeedbackModal(filePath string, lineNumber int, lineContent string) *FeedbackModal {
	ta := textarea.New()
	ta.Placeholder = "Enter your feedback..."
	ta.Focus()
	ta.CharLimit = 0 // No limit
	ta.ShowLineNumbers = false

	return &FeedbackModal{
		textarea:    ta,
		filePath:    filePath,
		lineNumber:  lineNumber,
		lineContent: lineContent,
	}
}

func (m *FeedbackModal) Init() tea.Cmd {
	return textarea.Blink
}

func (m *FeedbackModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Enter saves feedback
			comment := strings.TrimSpace(m.textarea.Value())
			if comment != "" {
				return m, func() tea.Msg {
					return FeedbackSavedMsg{
						FilePath:   m.filePath,
						LineNumber: m.lineNumber,
						Comment:    comment,
					}
				}
			}
			// Empty comment, treat as cancel
			return m, func() tea.Msg {
				return FeedbackCancelledMsg{}
			}
		case "ctrl+j":
			// Ctrl+J inserts newline
			m.textarea.InsertString("\n")
			return m, nil
		case "esc":
			// Escape cancels
			return m, func() tea.Msg {
				return FeedbackCancelledMsg{}
			}
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m *FeedbackModal) View() string {
	if !m.ready {
		return ""
	}

	// Calculate 75% of screen dimensions
	windowWidth := m.width * 75 / 100
	windowHeight := m.height * 75 / 100

	// Minimum sizes
	if windowWidth < 40 {
		windowWidth = 40
	}
	if windowHeight < 10 {
		windowHeight = 10
	}

	// Calculate content area (minus borders)
	contentWidth := windowWidth - 4
	contentHeight := windowHeight - 4

	// Build content
	var lines []string

	// Show context: file:line (lineNumber is already 1-indexed from CalculateLineNumber)
	context := theme.DimmedStyle.Render(fmt.Sprintf("@%s:%d", m.filePath, m.lineNumber))
	lines = append(lines, context)
	lines = append(lines, "")

	// Show the line content being commented on (truncated if needed)
	if m.lineContent != "" {
		linePreview := m.lineContent
		if len(linePreview) > contentWidth-2 {
			linePreview = linePreview[:contentWidth-5] + "..."
		}
		lines = append(lines, theme.DiffContextLine.Render(linePreview))
		lines = append(lines, "")
	}

	// Textarea
	m.textarea.SetWidth(contentWidth)
	m.textarea.SetHeight(contentHeight - len(lines) - 3)
	lines = append(lines, m.textarea.View())

	// Help text at bottom
	lines = append(lines, "")
	lines = append(lines, theme.HelpDescStyle.Render("enter save  C-j newline  esc cancel"))

	content := strings.Join(lines, "\n")

	// Render floating window
	windowContent := borders.RenderFloatingBorder(content, "Feedback", windowWidth, windowHeight)

	// Center the window
	x := (m.width - windowWidth) / 2
	y := (m.height - windowHeight) / 2

	// Add padding to center
	windowLines := strings.Split(windowContent, "\n")
	for i := range windowLines {
		windowLines[i] = strings.Repeat(" ", x) + windowLines[i]
	}

	paddingTop := strings.Repeat("\n", y)
	return paddingTop + strings.Join(windowLines, "\n")
}

// SetSize sets the available screen size
func (m *FeedbackModal) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.ready = true

	// Update textarea size
	windowWidth := width * 75 / 100
	if windowWidth < 40 {
		windowWidth = 40
	}
	m.textarea.SetWidth(windowWidth - 6)
}

// Overlay renders the modal on top of existing content
func (m *FeedbackModal) Overlay(baseContent string) string {
	if !m.ready {
		return baseContent
	}

	modalView := m.View()

	// Split both into lines
	baseLines := strings.Split(baseContent, "\n")
	modalLines := strings.Split(modalView, "\n")

	// Overlay modal on base
	result := make([]string, len(baseLines))
	for i := range baseLines {
		if i < len(modalLines) && strings.TrimSpace(modalLines[i]) != "" {
			result[i] = modalLines[i]
		} else {
			result[i] = baseLines[i]
		}
	}

	return strings.Join(result, "\n")
}

// FilePath returns the file being commented on
func (m *FeedbackModal) FilePath() string {
	return m.filePath
}

// LineNumber returns the line number (0-indexed)
func (m *FeedbackModal) LineNumber() int {
	return m.lineNumber
}

// Value returns the current textarea value
func (m *FeedbackModal) Value() string {
	return m.textarea.Value()
}

// CalculateLineNumber attempts to get actual line number from diff
// This is a heuristic based on diff format
func CalculateLineNumber(diffContent string, cursorLine int) int {
	lines := strings.Split(diffContent, "\n")
	if cursorLine >= len(lines) {
		return cursorLine + 1
	}

	// Find the most recent hunk header before cursor
	currentLine := 0
	for i := 0; i <= cursorLine && i < len(lines); i++ {
		line := lines[i]

		// Parse hunk header: @@ -start,count +start,count @@
		if strings.HasPrefix(line, "@@") {
			// Extract the new file line number
			parts := strings.Split(line, "+")
			if len(parts) >= 2 {
				numPart := strings.Split(parts[1], ",")[0]
				numPart = strings.TrimSpace(numPart)
				var n int
				fmt.Sscanf(numPart, "%d", &n)
				currentLine = n - 1 // Will be incremented for each non-deleted line
			}
		} else if i > 0 {
			// Count lines in new file (context and additions)
			if !strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "\\") {
				currentLine++
			}
		}
	}

	if currentLine == 0 {
		currentLine = cursorLine + 1
	}

	return currentLine
}

// Simple overlay without background dimming
func RenderSimpleOverlay(base, overlay string, width, height int) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// Ensure baseLines has enough lines
	for len(baseLines) < height {
		baseLines = append(baseLines, strings.Repeat(" ", width))
	}

	// Overlay each line
	result := make([]string, len(baseLines))
	for i := range baseLines {
		if i < len(overlayLines) {
			overlayLine := overlayLines[i]
			// Check if overlay line has content
			if lipgloss.Width(strings.TrimSpace(overlayLine)) > 0 {
				result[i] = overlayLine
			} else {
				result[i] = baseLines[i]
			}
		} else {
			result[i] = baseLines[i]
		}
	}

	return strings.Join(result, "\n")
}
