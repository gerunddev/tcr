package panels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gerund/tcr/ui/theme"
)

// DiffPanel shows diff content with a cursor for line selection
type DiffPanel struct {
	BasePanel
	viewport   viewport.Model
	lines      []string // Raw diff lines
	cursorLine int      // Current cursor position (0-indexed)
	filePath   string   // Currently displayed file
	ready      bool
}

// NewDiffPanel creates a new diff panel
func NewDiffPanel() *DiffPanel {
	return &DiffPanel{
		BasePanel: NewBasePanel("Diff", "file diff"),
	}
}

// SetDiff sets the diff content for a file
func (p *DiffPanel) SetDiff(filePath, content string) {
	p.filePath = filePath
	p.lines = strings.Split(content, "\n")
	p.cursorLine = 0

	// Update title to show file path
	p.SetTitle("Diff: " + filePath)

	if p.ready {
		p.viewport.SetContent(p.renderContent())
		p.viewport.GotoTop()
	}
}

// ClearDiff clears the diff content
func (p *DiffPanel) ClearDiff() {
	p.filePath = ""
	p.lines = nil
	p.cursorLine = 0
	p.SetTitle("Diff")

	if p.ready {
		p.viewport.SetContent("")
		p.viewport.GotoTop()
	}
}

func (p *DiffPanel) Init() tea.Cmd {
	return nil
}

func (p *DiffPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Emacs-style navigation only
		case "ctrl+n": // Cursor down
			p.cursorDown()
		case "ctrl+p": // Cursor up
			p.cursorUp()
		case "ctrl+v": // Page down
			p.pageDown()
		case "alt+v": // Page up
			p.pageUp()
		case "alt+<": // Top
			p.gotoTop()
		case "alt+>": // Bottom
			p.gotoBottom()
		}

		// Update viewport content after cursor moves
		if p.ready {
			p.viewport.SetContent(p.renderContent())
		}
	}

	return p, nil
}

func (p *DiffPanel) cursorUp() {
	if p.cursorLine > 0 {
		p.cursorLine--
		p.ensureCursorVisible()
	}
}

func (p *DiffPanel) cursorDown() {
	if p.cursorLine < len(p.lines)-1 {
		p.cursorLine++
		p.ensureCursorVisible()
	}
}

func (p *DiffPanel) pageUp() {
	pageSize := p.ContentHeight()
	p.cursorLine -= pageSize
	if p.cursorLine < 0 {
		p.cursorLine = 0
	}
	p.ensureCursorVisible()
}

func (p *DiffPanel) pageDown() {
	pageSize := p.ContentHeight()
	p.cursorLine += pageSize
	if p.cursorLine >= len(p.lines) {
		p.cursorLine = len(p.lines) - 1
	}
	if p.cursorLine < 0 {
		p.cursorLine = 0
	}
	p.ensureCursorVisible()
}

func (p *DiffPanel) gotoTop() {
	p.cursorLine = 0
	p.viewport.GotoTop()
}

func (p *DiffPanel) gotoBottom() {
	if len(p.lines) > 0 {
		p.cursorLine = len(p.lines) - 1
	}
	p.viewport.GotoBottom()
}

func (p *DiffPanel) ensureCursorVisible() {
	if p.cursorLine < p.viewport.YOffset {
		p.viewport.SetYOffset(p.cursorLine)
	} else if p.cursorLine >= p.viewport.YOffset+p.viewport.Height {
		p.viewport.SetYOffset(p.cursorLine - p.viewport.Height + 1)
	}
}

func (p *DiffPanel) View() string {
	if !p.ready {
		return p.RenderFrame("Loading...")
	}
	if len(p.lines) == 0 || (len(p.lines) == 1 && p.lines[0] == "") {
		return p.RenderFrame(theme.DimmedStyle.Render("No diff to show"))
	}
	return p.RenderFrame(p.viewport.View())
}

// SetSize initializes or resizes the viewport
func (p *DiffPanel) SetSize(width, height int) {
	p.BasePanel.SetSize(width, height)

	contentWidth := p.ContentWidth()
	contentHeight := p.ContentHeight()

	if !p.ready {
		p.viewport = viewport.New(contentWidth, contentHeight)
		p.viewport.SetContent(p.renderContent())
		p.ready = true
	} else {
		p.viewport.Width = contentWidth
		p.viewport.Height = contentHeight
		p.viewport.SetContent(p.renderContent())
	}
}

func (p *DiffPanel) renderContent() string {
	if len(p.lines) == 0 {
		return ""
	}

	contentWidth := p.ContentWidth()
	var rendered []string

	for i, line := range p.lines {
		styledLine := p.styleDiffLine(line, contentWidth)

		// Apply cursor highlight
		if i == p.cursorLine {
			// Reverse video effect for cursor line
			styledLine = theme.CursorLineStyle.Width(contentWidth).Render(styledLine)
		}

		rendered = append(rendered, styledLine)
	}

	return strings.Join(rendered, "\n")
}

func (p *DiffPanel) styleDiffLine(line string, maxWidth int) string {
	// Truncate if needed
	if lipgloss.Width(line) > maxWidth {
		line = lipgloss.NewStyle().MaxWidth(maxWidth).Render(line)
	}

	// Apply diff-specific styling
	if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
		return theme.DiffAddLine.Render(line)
	}
	if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
		return theme.DiffRemoveLine.Render(line)
	}
	if strings.HasPrefix(line, "@@") {
		return theme.DiffHunkHeader.Render(line)
	}
	if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") ||
		strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
		return theme.DiffHunkHeader.Render(line)
	}

	return theme.DiffContextLine.Render(line)
}

// CursorLine returns the current cursor line number (0-indexed)
func (p *DiffPanel) CursorLine() int {
	return p.cursorLine
}

// FilePath returns the current file path
func (p *DiffPanel) FilePath() string {
	return p.filePath
}

// CurrentLineContent returns the content of the current cursor line
func (p *DiffPanel) CurrentLineContent() string {
	if p.cursorLine >= 0 && p.cursorLine < len(p.lines) {
		return p.lines[p.cursorLine]
	}
	return ""
}

// DiffContent returns the full diff content as a string
func (p *DiffPanel) DiffContent() string {
	return strings.Join(p.lines, "\n")
}

// Ensure DiffPanel implements Panel
var _ Panel = (*DiffPanel)(nil)
