package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gerunddev/tcr/output"
	"github.com/gerunddev/tcr/ui/floating"
	"github.com/gerunddev/tcr/ui/panels"
	"github.com/gerunddev/tcr/ui/search"
	"github.com/gerunddev/tcr/ui/theme"
	"github.com/gerunddev/tcr/vcs"
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

	// Search
	searchCtrl *search.Controller
	diffCache  map[string]string // Cache of loaded diffs by file path

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
		searchCtrl: search.NewController(),
		diffCache:  make(map[string]string),
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
		// Cache the diff
		a.diffCache[msg.path] = msg.content

		// Set the diff content
		a.diffPanel.SetDiff(msg.path, msg.content)

		// If search is active, apply search to the new diff
		if a.searchCtrl.IsActive() {
			a.diffPanel.SetSearchQuery(a.searchCtrl.Query())
			a.updateDiffSearchMatches(a.searchCtrl.Query())
			a.diffPanel.SetSearchInputView(a.searchCtrl.InputView())
		}
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

	case diffsPreloadedBatchMsg:
		// Add preloaded diffs to cache
		for _, result := range msg.results {
			a.diffCache[result.path] = result.content
		}
		// Re-run search if active to include newly cached diffs
		if a.searchCtrl.IsActive() && a.searchCtrl.Query() != "" {
			a.runSearch()
		}
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

		// Handle unified search mode at app level
		if a.searchCtrl.IsActive() {
			return a.handleSearchInput(msg)
		}

		// Global key handling
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit

		case "/":
			// Activate unified search
			return a.activateSearch()

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

		// Route emacs bindings to diff panel (always)
		switch msg.String() {
		case "ctrl+n", "ctrl+p", "ctrl+v", "alt+v", "alt+<", "alt+>":
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

// activateSearch starts unified search mode
func (a *App) activateSearch() (tea.Model, tea.Cmd) {
	// Set width for search input
	diffWidth := a.width - theme.SidebarWidth
	if diffWidth < a.width*2/3 {
		diffWidth = a.width * 2 / 3
	}
	a.searchCtrl.SetWidth(diffWidth - 2) // Account for borders

	// Activate search in controller and diff panel
	cmd := a.searchCtrl.Activate()
	a.diffPanel.ActivateSearch()

	// Sync input view for proper cursor rendering
	a.diffPanel.SetSearchInputView(a.searchCtrl.InputView())

	// Start preloading uncached diffs in background
	preloadCmd := a.preloadDiffsAsync()

	return a, tea.Batch(cmd, preloadCmd)
}

// diffPreloadedMsg is sent when a diff is preloaded into cache
type diffPreloadedMsg struct {
	path    string
	content string
}

// preloadDiffsAsync returns a command that loads uncached diffs in background
func (a *App) preloadDiffsAsync() tea.Cmd {
	paths := a.filesPanel.FilePaths()

	// Collect paths that need loading
	var uncachedPaths []string
	for _, path := range paths {
		if _, ok := a.diffCache[path]; !ok {
			uncachedPaths = append(uncachedPaths, path)
		}
	}

	if len(uncachedPaths) == 0 {
		return nil
	}

	// Load all uncached diffs concurrently
	return func() tea.Msg {
		var results []diffPreloadedMsg
		for _, path := range uncachedPaths {
			content, err := a.vcs.Diff(path)
			if err == nil {
				results = append(results, diffPreloadedMsg{path: path, content: content})
			}
		}
		return diffsPreloadedBatchMsg{results: results}
	}
}

// diffsPreloadedBatchMsg is sent when all background diffs are loaded
type diffsPreloadedBatchMsg struct {
	results []diffPreloadedMsg
}

// handleSearchInput processes keys during search mode
func (a *App) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Exit search mode
		a.deactivateSearch()
		return a, nil

	case "enter":
		// Cycle to next match in current diff
		a.diffPanel.CycleNextMatch()
		return a, nil

	case "up":
		// Navigate to previous file in filtered list
		var cmd tea.Cmd
		_, cmd = a.filesPanel.Update(msg)
		return a, cmd

	case "down":
		// Navigate to next file in filtered list
		var cmd tea.Cmd
		_, cmd = a.filesPanel.Update(msg)
		return a, cmd

	default:
		// Pass to search controller for query editing
		oldQuery := a.searchCtrl.Query()
		cmd := a.searchCtrl.UpdateInput(msg)

		// Re-run search if query changed
		if a.searchCtrl.Query() != oldQuery {
			a.runSearch()
		}

		// Always sync the input view (for cursor position)
		a.diffPanel.SetSearchInputView(a.searchCtrl.InputView())

		return a, cmd
	}
}

// runSearch executes search across all files and updates panels
func (a *App) runSearch() {
	query := a.searchCtrl.Query()

	// Get file paths
	paths := a.filesPanel.FilePaths()

	// Run search across all cached diffs
	a.searchCtrl.SearchAllFiles(query, paths, a.diffCache)

	// Update files panel with filtered indices
	filteredIdxs := a.searchCtrl.FilteredIndices()
	if filteredIdxs != nil {
		a.filesPanel.SetFilteredIndices(filteredIdxs)
	} else {
		a.filesPanel.ClearFilter()
	}

	// Update diff panel with current search query and matches
	a.diffPanel.SetSearchQuery(query)
	a.updateDiffSearchMatches(query)
}

// updateDiffSearchMatches runs search on current diff and updates matches
func (a *App) updateDiffSearchMatches(query string) {
	if query == "" {
		a.diffPanel.SetSearchMatches(nil)
		return
	}

	matches, _ := a.searchCtrl.SearchInDiff(query, a.diffPanel.Lines())
	a.diffPanel.SetSearchMatches(matches)
}

// deactivateSearch exits search mode
func (a *App) deactivateSearch() {
	a.searchCtrl.Deactivate()
	a.filesPanel.ClearFilter()
	a.diffPanel.DeactivateSearch()
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
		SearchActive: a.searchCtrl.IsActive(),
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
