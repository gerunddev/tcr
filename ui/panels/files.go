package panels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gerund/tcr/ui/theme"
	"github.com/gerund/tcr/vcs"
)

// FileSelectedMsg is sent when a file is selected
type FileSelectedMsg struct {
	Path string
}

// FilesPanel shows changed files from VCS
type FilesPanel struct {
	BasePanel
	files    []vcs.FileChange
	viewport viewport.Model
	ready    bool
}

// NewFilesPanel creates a new files panel
func NewFilesPanel() *FilesPanel {
	return &FilesPanel{
		BasePanel: NewBasePanel("Files", "changed files"),
	}
}

// SetFiles updates the file list
func (p *FilesPanel) SetFiles(files []vcs.FileChange) {
	p.files = files
	p.cursor = 0
	if p.ready {
		p.viewport.SetContent(p.renderContent())
		p.viewport.GotoTop()
	}
}

func (p *FilesPanel) Init() tea.Cmd {
	return nil
}

func (p *FilesPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	prevCursor := p.cursor

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Arrow keys only for file navigation
		case "up":
			p.CursorUp(len(p.files))
			p.ensureCursorVisible()
		case "down":
			p.CursorDown(len(p.files))
			p.ensureCursorVisible()
		}
	}

	// Update viewport content when cursor changes
	if p.ready {
		p.viewport.SetContent(p.renderContent())
	}

	// Emit selection message if cursor changed
	if p.cursor != prevCursor {
		if file := p.SelectedFile(); file != nil {
			return p, func() tea.Msg {
				return FileSelectedMsg{Path: file.Path}
			}
		}
	}

	return p, nil
}

func (p *FilesPanel) ensureCursorVisible() {
	if p.cursor < p.viewport.YOffset {
		p.viewport.SetYOffset(p.cursor)
	} else if p.cursor >= p.viewport.YOffset+p.viewport.Height {
		p.viewport.SetYOffset(p.cursor - p.viewport.Height + 1)
	}
}

func (p *FilesPanel) View() string {
	if !p.ready {
		return p.RenderFrame("Loading...")
	}
	if len(p.files) == 0 {
		return p.RenderFrame(theme.DimmedStyle.Render("No files changed"))
	}
	return p.RenderFrame(p.viewport.View())
}

// SetSize initializes or resizes the viewport
func (p *FilesPanel) SetSize(width, height int) {
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

func (p *FilesPanel) renderContent() string {
	var lines []string
	contentWidth := p.ContentWidth()

	for i, file := range p.files {
		// Style the status indicator based on file status
		var statusStyle lipgloss.Style
		switch file.Status {
		case vcs.StatusModified:
			statusStyle = theme.ModifiedStyle
		case vcs.StatusAdded:
			statusStyle = theme.AddedStyle
		case vcs.StatusDeleted:
			statusStyle = theme.DeletedStyle
		case vcs.StatusRenamed:
			statusStyle = theme.RenamedStyle
		default:
			statusStyle = theme.NormalItemStyle
		}

		status := statusStyle.Render(string(file.Status))

		// Truncate path if needed
		maxPathLen := contentWidth - 3 // status + space
		path := file.Path
		if len(path) > maxPathLen && maxPathLen > 0 {
			path = truncate(path, maxPathLen)
		}

		if i == p.cursor {
			// Show selected item in yellow
			path = theme.SelectedItemStyle.Render(path)
		} else {
			path = theme.NormalItemStyle.Render(path)
		}

		line := status + " " + path
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// SelectedFile returns the currently selected file
func (p *FilesPanel) SelectedFile() *vcs.FileChange {
	if p.cursor >= 0 && p.cursor < len(p.files) {
		return &p.files[p.cursor]
	}
	return nil
}

// Count returns the number of files
func (p *FilesPanel) Count() int {
	return len(p.files)
}

// truncate shortens a string to the given display width
// Uses lipgloss.Width for proper handling of multi-byte UTF-8 characters
func truncate(s string, maxWidth int) string {
	width := lipgloss.Width(s)
	if width <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return lipgloss.NewStyle().MaxWidth(maxWidth).Render(s)
	}
	// Truncate from the beginning, showing the end of the path
	// This is more useful for file paths where the filename is at the end
	runes := []rune(s)
	for lipgloss.Width(string(runes)) > maxWidth-3 {
		runes = runes[1:]
	}
	return "..." + string(runes)
}

// Ensure FilesPanel implements Panel
var _ Panel = (*FilesPanel)(nil)
