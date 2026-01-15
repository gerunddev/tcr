package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gerund/tcr/output"
	"github.com/gerund/tcr/ui/floating"
	"github.com/gerund/tcr/ui/panels"
	"github.com/gerund/tcr/ui/theme"
	"github.com/gerund/tcr/vcs"
)

// App is the main application model
type App struct {
	vcs        vcs.VCS
	outputPath string
	width      int
	height     int
	ready      bool

	// Panels
	filesPanel *panels.FilesPanel
	diffPanel  *panels.DiffPanel

	// Modal
	feedbackModal *floating.FeedbackModal
	modalOpen     bool

	// Messages
	statusMsg string
}

// NewApp creates a new application
func NewApp(v vcs.VCS, outputPath string) *App {
	filesPanel := panels.NewFilesPanel()
	diffPanel := panels.NewDiffPanel()

	// Both panels are always "focused" visually (yellow border)
	filesPanel.SetFocused(true)
	diffPanel.SetFocused(true)

	return &App{
		vcs:        v,
		outputPath: outputPath,
		filesPanel: filesPanel,
		diffPanel:  diffPanel,
	}
}

func (a *App) Init() tea.Cmd {
	return a.loadFiles
}

func (a *App) loadFiles() tea.Msg {
	files, err := a.vcs.ChangedFiles()
	if err != nil {
		return errMsg{err}
	}
	return filesLoadedMsg{files}
}

type filesLoadedMsg struct {
	files []vcs.FileChange
}

type errMsg struct {
	err error
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
		a.updatePanelSizes()

		if a.feedbackModal != nil {
			a.feedbackModal.SetSize(a.width, a.height)
		}

		return a, nil

	case filesLoadedMsg:
		a.filesPanel.SetFiles(msg.files)
		// Load diff for first file if any
		if len(msg.files) > 0 {
			return a, a.loadDiff(msg.files[0].Path)
		}
		return a, nil

	case panels.FileSelectedMsg:
		return a, a.loadDiff(msg.Path)

	case diffLoadedMsg:
		a.diffPanel.SetDiff(msg.path, msg.content)
		return a, nil

	case floating.FeedbackSavedMsg:
		// Save feedback to file
		err := output.AppendFeedback(a.outputPath, msg.FilePath, msg.LineNumber, msg.Comment)
		if err != nil {
			a.statusMsg = "Error: " + err.Error()
		} else {
			a.statusMsg = "Feedback saved"
		}
		a.closeModal()
		return a, nil

	case floating.FeedbackCancelledMsg:
		a.closeModal()
		return a, nil

	case errMsg:
		a.statusMsg = "Error: " + msg.err.Error()
		return a, nil

	case tea.KeyMsg:
		// Clear status message on any key press
		a.statusMsg = ""

		// Handle modal input first if open
		if a.modalOpen && a.feedbackModal != nil {
			var cmd tea.Cmd
			_, cmd = a.feedbackModal.Update(msg)
			return a, cmd
		}

		// Handle search mode - route all keys to diff panel
		if a.diffPanel.IsSearching() {
			var cmd tea.Cmd
			_, cmd = a.diffPanel.Update(msg)
			return a, cmd
		}

		// Global key handling
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit

		case "enter":
			// Enter on diff panel opens feedback modal
			a.openFeedbackModal()
			return a, nil
		}

		// Route arrow keys to files panel (always)
		switch msg.String() {
		case "up", "down":
			var cmd tea.Cmd
			_, cmd = a.filesPanel.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		// Route emacs bindings and "/" to diff panel (always)
		switch msg.String() {
		case "ctrl+n", "ctrl+p", "ctrl+v", "alt+v", "alt+<", "alt+>", "/":
			var cmd tea.Cmd
			_, cmd = a.diffPanel.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *App) loadDiff(path string) tea.Cmd {
	return func() tea.Msg {
		content, err := a.vcs.Diff(path)
		if err != nil {
			return errMsg{err}
		}
		return diffLoadedMsg{path: path, content: content}
	}
}

type diffLoadedMsg struct {
	path    string
	content string
}

func (a *App) openFeedbackModal() {
	filePath := a.diffPanel.FilePath()
	cursorLine := a.diffPanel.CursorLine()
	lineContent := a.diffPanel.CurrentLineContent()
	diffContent := a.diffPanel.DiffContent()

	if filePath == "" {
		return
	}

	// Calculate actual source line number from diff hunk headers
	actualLineNumber := floating.CalculateLineNumber(diffContent, cursorLine)

	a.feedbackModal = floating.NewFeedbackModal(filePath, actualLineNumber, lineContent)
	a.feedbackModal.SetSize(a.width, a.height)
	a.modalOpen = true
}

func (a *App) closeModal() {
	a.feedbackModal = nil
	a.modalOpen = false
}

func (a *App) updatePanelSizes() {
	if !a.ready {
		return
	}

	// Reserve 1 line for help bar
	availableHeight := a.height - 1

	// Files panel: fixed width on left
	filesWidth := theme.SidebarWidth
	if filesWidth > a.width/3 {
		filesWidth = a.width / 3
	}

	// Diff panel: rest of width
	diffWidth := a.width - filesWidth

	a.filesPanel.SetSize(filesWidth, availableHeight)
	a.diffPanel.SetSize(diffWidth, availableHeight)
}

func (a *App) View() string {
	if !a.ready {
		return "Loading..."
	}

	// Render panels
	filesView := a.filesPanel.View()
	diffView := a.diffPanel.View()

	// Join panels horizontally
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, filesView, diffView)

	// Add help bar
	helpCtx := HelpBarContext{
		ModalOpen:    a.modalOpen,
		SearchActive: a.diffPanel.IsSearching(),
	}
	helpBar := RenderHelpBar(helpCtx, a.width)

	// Combine main view and help bar
	fullView := lipgloss.JoinVertical(lipgloss.Left, mainView, helpBar)

	// Overlay modal if open
	if a.modalOpen && a.feedbackModal != nil {
		return floating.RenderSimpleOverlay(fullView, a.feedbackModal.View(), a.width, a.height)
	}

	// Add status message if any (replaces help bar temporarily)
	if a.statusMsg != "" {
		lines := strings.Split(fullView, "\n")
		if len(lines) > 0 {
			statusStyle := theme.HelpDescStyle.Width(a.width)
			lines[len(lines)-1] = statusStyle.Render(a.statusMsg)
			return strings.Join(lines, "\n")
		}
	}

	return fullView
}
