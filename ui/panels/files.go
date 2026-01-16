package panels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gerunddev/tcr/ui/theme"
	"github.com/gerunddev/tcr/vcs"
)

// FileSelectedMsg is sent when a file is selected
type FileSelectedMsg struct {
	Path string
}

// FilesPanel shows changed files from VCS
type FilesPanel struct {
	BasePanel
	files        []vcs.FileChange
	filteredIdxs []int // Indices into files slice, nil means show all
	viewport     viewport.Model
	ready        bool
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
	p.filteredIdxs = nil
	p.cursor = 0
	if p.ready {
		p.viewport.SetContent(p.renderContent())
		p.viewport.GotoTop()
	}
}

// SetFilteredIndices sets which files to show (by index into full files list)
// Pass nil to show all files
func (p *FilesPanel) SetFilteredIndices(indices []int) {
	p.filteredIdxs = indices

	if len(indices) > 0 {
		// If current selection is not in filtered list, move to first filtered file
		found := false
		for _, fileIdx := range indices {
			if fileIdx == p.cursor {
				found = true
				// Keep the cursor at the same file index
				break
			}
		}
		if !found {
			// Move cursor to first filtered file
			p.cursor = indices[0]
		}
	}

	if p.ready {
		p.viewport.SetContent(p.renderContent())
		p.viewport.GotoTop()
	}
}

// ClearFilter removes any active filtering
func (p *FilesPanel) ClearFilter() {
	p.filteredIdxs = nil
	if p.ready {
		p.viewport.SetContent(p.renderContent())
	}
}

// IsFiltered returns true if a filter is active
func (p *FilesPanel) IsFiltered() bool {
	return p.filteredIdxs != nil
}

// displayFiles returns the files to display (filtered or all)
func (p *FilesPanel) displayFiles() []vcs.FileChange {
	if p.filteredIdxs == nil {
		return p.files
	}
	result := make([]vcs.FileChange, 0, len(p.filteredIdxs))
	for _, idx := range p.filteredIdxs {
		if idx >= 0 && idx < len(p.files) {
			result = append(result, p.files[idx])
		}
	}
	return result
}

// displayIndexToFileIndex converts display position to actual file index
func (p *FilesPanel) displayIndexToFileIndex(displayIdx int) int {
	if p.filteredIdxs == nil {
		return displayIdx
	}
	if displayIdx >= 0 && displayIdx < len(p.filteredIdxs) {
		return p.filteredIdxs[displayIdx]
	}
	return -1
}

// fileIndexToDisplayIndex converts actual file index to display position
func (p *FilesPanel) fileIndexToDisplayIndex(fileIdx int) int {
	if p.filteredIdxs == nil {
		return fileIdx
	}
	for i, idx := range p.filteredIdxs {
		if idx == fileIdx {
			return i
		}
	}
	return -1
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
			p.cursorUpFiltered()
			p.ensureCursorVisible()
		case "down":
			p.cursorDownFiltered()
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

// cursorUpFiltered moves cursor up within filtered list (or all files if no filter)
func (p *FilesPanel) cursorUpFiltered() {
	if p.filteredIdxs == nil {
		// No filter, use normal navigation
		p.CursorUp(len(p.files))
		return
	}

	// Find current position in filtered list
	displayIdx := p.fileIndexToDisplayIndex(p.cursor)
	if displayIdx > 0 {
		p.cursor = p.filteredIdxs[displayIdx-1]
	}
}

// cursorDownFiltered moves cursor down within filtered list (or all files if no filter)
func (p *FilesPanel) cursorDownFiltered() {
	if p.filteredIdxs == nil {
		// No filter, use normal navigation
		p.CursorDown(len(p.files))
		return
	}

	// Find current position in filtered list
	displayIdx := p.fileIndexToDisplayIndex(p.cursor)
	if displayIdx >= 0 && displayIdx < len(p.filteredIdxs)-1 {
		p.cursor = p.filteredIdxs[displayIdx+1]
	}
}

func (p *FilesPanel) ensureCursorVisible() {
	// Use display index for viewport positioning
	displayIdx := p.fileIndexToDisplayIndex(p.cursor)
	if displayIdx < 0 {
		displayIdx = 0
	}

	if displayIdx < p.viewport.YOffset {
		p.viewport.SetYOffset(displayIdx)
	} else if displayIdx >= p.viewport.YOffset+p.viewport.Height {
		p.viewport.SetYOffset(displayIdx - p.viewport.Height + 1)
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

	displayFiles := p.displayFiles()
	for displayIdx, file := range displayFiles {
		// Get actual file index for cursor comparison
		fileIdx := p.displayIndexToFileIndex(displayIdx)

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

		if fileIdx == p.cursor {
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

// Count returns the number of visible files (filtered or all)
func (p *FilesPanel) Count() int {
	if p.filteredIdxs != nil {
		return len(p.filteredIdxs)
	}
	return len(p.files)
}

// TotalCount returns the total number of files (ignoring filter)
func (p *FilesPanel) TotalCount() int {
	return len(p.files)
}

// FilePaths returns all file paths (for search)
func (p *FilesPanel) FilePaths() []string {
	paths := make([]string, len(p.files))
	for i, f := range p.files {
		paths[i] = f.Path
	}
	return paths
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
