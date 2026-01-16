package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gerunddev/tcr/ui/theme"
)

// SearchState holds the state for diff search
type SearchState struct {
	active          bool         // Whether search mode is active
	matches         []int        // Line indices that match (0-indexed)
	matchSet        map[int]bool // O(1) lookup for matched lines
	currentMatch    int          // Index into matches slice (-1 if no matches)
	input           textinput.Model
	externalInputView string     // When set, use this for rendering instead of local input
	fzfError        string       // Error message if fzf unavailable
}

// NewSearchState creates a new search state
func NewSearchState() *SearchState {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Prompt = ""
	ti.CharLimit = 100
	ti.Width = 30

	return &SearchState{
		currentMatch: -1,
		input:        ti,
	}
}

// Reset clears the search state
func (s *SearchState) Reset() {
	s.active = false
	s.matches = nil
	s.matchSet = nil
	s.currentMatch = -1
	s.input.SetValue("")
	s.externalInputView = ""
	s.fzfError = ""
}

// Activate enables search mode and focuses input
func (s *SearchState) Activate() {
	s.active = true
	s.input.Focus()
	s.input.SetValue("")
	s.externalInputView = ""
	s.matches = nil
	s.matchSet = nil
	s.currentMatch = -1
	s.fzfError = ""
}

// Deactivate disables search mode
func (s *SearchState) Deactivate() {
	s.active = false
	s.input.Blur()
}

// Query returns the current search query
func (s *SearchState) Query() string {
	return s.input.Value()
}

// HasMatches returns true if there are matches
func (s *SearchState) HasMatches() bool {
	return len(s.matches) > 0
}

// CurrentMatchLine returns the line index of current match, or -1
func (s *SearchState) CurrentMatchLine() int {
	if s.currentMatch >= 0 && s.currentMatch < len(s.matches) {
		return s.matches[s.currentMatch]
	}
	return -1
}

// NextMatch moves to the next match (wrapping)
func (s *SearchState) NextMatch() {
	if len(s.matches) == 0 {
		return
	}
	s.currentMatch = (s.currentMatch + 1) % len(s.matches)
}

// PrevMatch moves to the previous match (wrapping)
func (s *SearchState) PrevMatch() {
	if len(s.matches) == 0 {
		return
	}
	s.currentMatch--
	if s.currentMatch < 0 {
		s.currentMatch = len(s.matches) - 1
	}
}

// MatchStatus returns "1/5" style status, or error/no matches message
func (s *SearchState) MatchStatus() string {
	if s.fzfError != "" {
		return s.fzfError
	}
	if len(s.matches) == 0 {
		return "no matches"
	}
	return fmt.Sprintf("%d/%d", s.currentMatch+1, len(s.matches))
}

// IsLineMatched returns true if the given line index is in matches (O(1) lookup)
func (s *SearchState) IsLineMatched(lineIdx int) bool {
	return s.matchSet[lineIdx]
}

// IsCurrentMatch returns true if lineIdx is the current match
func (s *SearchState) IsCurrentMatch(lineIdx int) bool {
	return s.CurrentMatchLine() == lineIdx
}

// SetWidth updates the input width
func (s *SearchState) SetWidth(w int) {
	s.input.Width = w - 15
	if s.input.Width < 10 {
		s.input.Width = 10
	}
}

// UpdateInput handles textinput updates
func (s *SearchState) UpdateInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return cmd
}

// InputView renders the textinput
func (s *SearchState) InputView() string {
	if s.externalInputView != "" {
		return s.externalInputView
	}
	return s.input.View()
}

// SetExternalInputView sets an external input view for rendering (used in unified search mode)
func (s *SearchState) SetExternalInputView(view string) {
	s.externalInputView = view
}

// DiffPanel shows diff content with a cursor for line selection
type DiffPanel struct {
	BasePanel
	viewport    viewport.Model
	lines       []string     // Raw diff lines
	cursorLine  int          // Current cursor position (0-indexed)
	filePath    string       // Currently displayed file
	ready       bool
	searchState *SearchState // Search state
}

// NewDiffPanel creates a new diff panel
func NewDiffPanel() *DiffPanel {
	return &DiffPanel{
		BasePanel:   NewBasePanel("Diff", "file diff"),
		searchState: NewSearchState(),
	}
}

// SetDiff sets the diff content for a file
func (p *DiffPanel) SetDiff(filePath, content string) {
	p.filePath = filePath
	p.lines = strings.Split(content, "\n")
	p.cursorLine = 0

	// Update title to show file path
	p.SetTitle("Diff: " + filePath)

	// Clear search matches (app will re-apply if needed)
	if p.searchState.active {
		p.searchState.matches = nil
		p.searchState.matchSet = nil
		p.searchState.currentMatch = -1
	}

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
	p.searchState.Reset()
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
		// Handle search mode
		if p.searchState.active {
			return p.handleSearchInput(msg)
		}

		// Handle normal mode
		switch msg.String() {
		case "/":
			// Activate search
			p.searchState.Activate()
			p.searchState.SetWidth(p.ContentWidth())
			p.updateViewportSize()
			return p, textinput.Blink

		// Emacs-style navigation
		case "ctrl+n":
			p.cursorDown()
		case "ctrl+p":
			p.cursorUp()
		case "ctrl+v":
			p.pageDown()
		case "alt+v":
			p.pageUp()
		case "alt+<":
			p.gotoTop()
		case "alt+>":
			p.gotoBottom()
		}

		// Update viewport content after cursor moves
		if p.ready {
			p.viewport.SetContent(p.renderContent())
		}
	}

	return p, nil
}

func (p *DiffPanel) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Exit search mode, keep cursor position
		p.searchState.Deactivate()
		p.updateViewportSize()
		return p, nil

	case "enter":
		// Cycle to next match
		if p.searchState.HasMatches() {
			p.searchState.NextMatch()
			p.cursorLine = p.searchState.CurrentMatchLine()
			p.ensureCursorVisible()
			p.viewport.SetContent(p.renderContent())
		}
		return p, nil

	default:
		// Pass to textinput for query editing (app handles actual search)
		cmd := p.searchState.UpdateInput(msg)
		return p, cmd
	}
}

// IsSearching returns true if search mode is active
func (p *DiffPanel) IsSearching() bool {
	return p.searchState.active
}

// ActivateSearch enables search mode (called by App)
func (p *DiffPanel) ActivateSearch() {
	p.searchState.active = true
	p.searchState.input.Focus()
	p.searchState.SetWidth(p.ContentWidth())
	p.updateViewportSize()
}

// DeactivateSearch disables search mode (called by App)
func (p *DiffPanel) DeactivateSearch() {
	p.searchState.Reset()
	p.updateViewportSize()
	if p.ready {
		p.viewport.SetContent(p.renderContent())
	}
}

// SetSearchQuery updates the search query (called by App)
func (p *DiffPanel) SetSearchQuery(query string) {
	p.searchState.input.SetValue(query)
}

// SetSearchMatches sets the search matches directly (called by App to avoid duplicate fzf calls)
func (p *DiffPanel) SetSearchMatches(matches []int) {
	p.searchState.matches = matches

	// Build O(1) lookup map
	p.searchState.matchSet = make(map[int]bool, len(matches))
	for _, m := range matches {
		p.searchState.matchSet[m] = true
	}

	// Reset current match if we have results
	if len(matches) > 0 {
		p.searchState.currentMatch = 0
		// Move cursor to first match
		p.cursorLine = matches[0]
		p.ensureCursorVisible()
	} else {
		p.searchState.currentMatch = -1
	}

	if p.ready {
		p.viewport.SetContent(p.renderContent())
	}
}

// SetSearchInputView sets the external input view for proper cursor rendering
func (p *DiffPanel) SetSearchInputView(view string) {
	p.searchState.SetExternalInputView(view)
}

// CycleNextMatch moves to the next match and returns true if cursor moved
func (p *DiffPanel) CycleNextMatch() bool {
	if !p.searchState.HasMatches() {
		return false
	}
	p.searchState.NextMatch()
	p.cursorLine = p.searchState.CurrentMatchLine()
	p.ensureCursorVisible()
	if p.ready {
		p.viewport.SetContent(p.renderContent())
	}
	return true
}

// MatchCount returns the number of matches in current diff
func (p *DiffPanel) MatchCount() int {
	return len(p.searchState.matches)
}

// CurrentMatchIndex returns the current match index (1-based) or 0 if no matches
func (p *DiffPanel) CurrentMatchIndex() int {
	if p.searchState.currentMatch < 0 {
		return 0
	}
	return p.searchState.currentMatch + 1
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

	content := p.viewport.View()

	// Add search bar if active
	if p.searchState.active {
		content = p.renderWithSearchBar(content)
	}

	return p.RenderFrame(content)
}

func (p *DiffPanel) renderWithSearchBar(content string) string {
	contentWidth := p.ContentWidth()

	// Build search bar
	searchBar := p.renderSearchBar(contentWidth)

	// Split content into lines
	lines := strings.Split(content, "\n")

	// Calculate available height for content (minus search bar)
	contentHeight := p.ContentHeight() - 1

	// Ensure we don't exceed height
	if len(lines) > contentHeight {
		lines = lines[:contentHeight]
	}

	// Pad to fill space before search bar
	for len(lines) < contentHeight {
		lines = append(lines, strings.Repeat(" ", contentWidth))
	}

	// Add search bar at bottom
	lines = append(lines, searchBar)

	return strings.Join(lines, "\n")
}

func (p *DiffPanel) renderSearchBar(width int) string {
	// Format: /query                              [1/5]
	prompt := theme.SearchPromptStyle.Render("/")
	query := p.searchState.InputView()
	status := theme.SearchStatusStyle.Render("[" + p.searchState.MatchStatus() + "]")

	// Calculate spacing
	promptWidth := lipgloss.Width(prompt)
	queryWidth := lipgloss.Width(query)
	statusWidth := lipgloss.Width(status)

	spacerWidth := width - promptWidth - queryWidth - statusWidth
	if spacerWidth < 1 {
		spacerWidth = 1
	}
	spacer := strings.Repeat(" ", spacerWidth)

	return theme.SearchBarStyle.Width(width).Render(prompt + query + spacer + status)
}

func (p *DiffPanel) updateViewportSize() {
	contentHeight := p.ContentHeight()
	if p.searchState.active {
		contentHeight-- // Reserve one line for search bar
	}
	p.viewport.Height = contentHeight
	p.viewport.SetContent(p.renderContent())
}

// SetSize initializes or resizes the viewport
func (p *DiffPanel) SetSize(width, height int) {
	p.BasePanel.SetSize(width, height)

	contentWidth := p.ContentWidth()
	contentHeight := p.ContentHeight()

	// Reserve space for search bar when active
	if p.searchState.active {
		contentHeight--
	}

	if !p.ready {
		p.viewport = viewport.New(contentWidth, contentHeight)
		p.viewport.SetContent(p.renderContent())
		p.ready = true
	} else {
		p.viewport.Width = contentWidth
		p.viewport.Height = contentHeight
		p.viewport.SetContent(p.renderContent())
	}

	// Update search input width
	p.searchState.SetWidth(contentWidth)
}

func (p *DiffPanel) renderContent() string {
	if len(p.lines) == 0 {
		return ""
	}

	contentWidth := p.ContentWidth()
	var rendered []string

	for i, line := range p.lines {
		styledLine := p.styleDiffLine(line, contentWidth)

		// Apply search highlighting if search is active
		if p.searchState.active && p.searchState.HasMatches() {
			if p.searchState.IsCurrentMatch(i) {
				// Current match - most prominent
				styledLine = theme.SearchCurrentLineStyle.Width(contentWidth).Render(styledLine)
			} else if p.searchState.IsLineMatched(i) {
				// Other matches - subtle highlight
				styledLine = theme.SearchMatchLineStyle.Width(contentWidth).Render(styledLine)
			} else if i == p.cursorLine {
				// Cursor on non-match line
				styledLine = theme.CursorLineStyle.Width(contentWidth).Render(styledLine)
			}
		} else if i == p.cursorLine {
			// Normal cursor highlight (when not searching)
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

// Lines returns the diff lines for searching
func (p *DiffPanel) Lines() []string {
	return p.lines
}

// Ensure DiffPanel implements Panel
var _ Panel = (*DiffPanel)(nil)
