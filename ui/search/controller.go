package search

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FileMatch represents a file that matched the search query
type FileMatch struct {
	Path    string
	Matches []int // Line indices that match (0-indexed)
}

// Controller handles unified search across files and diffs
type Controller struct {
	active       bool              // Whether search mode is active
	input        textinput.Model   // Search input
	query        string            // Current search query
	filteredIdxs []int             // Indices of files that match (into original files list)
	noMatches    bool              // True if search ran but found no matches
	fzfError     string            // Error message if fzf unavailable
	inputWidth   int               // Width for the input field
}

// NewController creates a new search controller
func NewController() *Controller {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Prompt = ""
	ti.CharLimit = 100
	ti.Width = 30

	return &Controller{
		input: ti,
	}
}

// IsActive returns true if search mode is active
func (c *Controller) IsActive() bool {
	return c.active
}

// Activate enables search mode
func (c *Controller) Activate() tea.Cmd {
	c.active = true
	c.query = ""
	c.filteredIdxs = nil
	c.noMatches = false
	c.fzfError = ""
	c.input.SetValue("")
	c.input.Focus()
	return textinput.Blink
}

// Deactivate disables search mode
func (c *Controller) Deactivate() {
	c.active = false
	c.query = ""
	c.filteredIdxs = nil
	c.noMatches = false
	c.input.Blur()
	c.input.SetValue("")
}

// Query returns the current search query
func (c *Controller) Query() string {
	return c.query
}

// FilteredIndices returns indices of files that match the search
// Returns nil if no filtering is active (empty query or no matches mode)
func (c *Controller) FilteredIndices() []int {
	return c.filteredIdxs
}

// HasNoMatches returns true if search was performed but found no matches
func (c *Controller) HasNoMatches() bool {
	return c.noMatches
}

// SetWidth sets the width for the search input
func (c *Controller) SetWidth(w int) {
	c.inputWidth = w
	c.input.Width = w - 15
	if c.input.Width < 10 {
		c.input.Width = 10
	}
}

// UpdateInput handles text input updates
func (c *Controller) UpdateInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)
	c.query = c.input.Value()
	return cmd
}

// InputView returns the rendered input field
func (c *Controller) InputView() string {
	return c.input.View()
}

// Status returns the search status string
func (c *Controller) Status() string {
	if c.fzfError != "" {
		return c.fzfError
	}
	if c.noMatches {
		return "no matches"
	}
	if len(c.filteredIdxs) > 0 {
		if len(c.filteredIdxs) == 1 {
			return "1 file"
		}
		return fmt.Sprintf("%d files", len(c.filteredIdxs))
	}
	return ""
}

// SearchAllFiles runs fzf search across all diffs and returns matching file indices
// diffs is a map from file path to diff content
// files is the ordered list of file paths to preserve ordering
func (c *Controller) SearchAllFiles(query string, files []string, diffs map[string]string) {
	c.query = query
	c.fzfError = ""

	if query == "" {
		c.filteredIdxs = nil
		c.noMatches = false
		return
	}

	// Check if fzf is available
	fzfPath, err := exec.LookPath("fzf")
	if err != nil {
		c.fzfError = "fzf not found"
		c.filteredIdxs = nil
		c.noMatches = true
		return
	}

	var matchingIdxs []int

	// Search each file's diff
	for i, filePath := range files {
		diffContent, ok := diffs[filePath]
		if !ok || diffContent == "" {
			continue
		}

		if c.diffContainsMatch(fzfPath, query, diffContent) {
			matchingIdxs = append(matchingIdxs, i)
		}
	}

	if len(matchingIdxs) == 0 {
		c.filteredIdxs = nil
		c.noMatches = true
	} else {
		c.filteredIdxs = matchingIdxs
		c.noMatches = false
	}
}

// diffContainsMatch checks if a diff contains any matches for the query using fzf
func (c *Controller) diffContainsMatch(fzfPath, query, diffContent string) bool {
	cmd := exec.Command(fzfPath, "--filter", query, "--exact")
	cmd.Stdin = strings.NewReader(diffContent)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// fzf returns exit code 1 when no matches, which is fine
	_ = cmd.Run()

	return stdout.Len() > 0
}

// SearchInDiff runs fzf search on specific diff content and returns matching line indices
func (c *Controller) SearchInDiff(query string, lines []string) ([]int, error) {
	if query == "" || len(lines) == 0 {
		return nil, nil
	}

	fzfPath, err := exec.LookPath("fzf")
	if err != nil {
		return nil, fmt.Errorf("fzf not found")
	}

	// Prepare input with line numbers for tracking
	var input strings.Builder
	for i, line := range lines {
		fmt.Fprintf(&input, "%d:%s\n", i, line)
	}

	cmd := exec.Command(fzfPath, "--filter", query, "--exact")
	cmd.Stdin = strings.NewReader(input.String())

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// fzf returns exit code 1 when no matches
	_ = cmd.Run()

	output := stdout.String()
	if output == "" {
		return nil, nil
	}

	var matches []int
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		// Extract line number from "linenum:content"
		idx := strings.Index(line, ":")
		if idx > 0 {
			var lineNum int
			if _, err := fmt.Sscanf(line[:idx], "%d", &lineNum); err == nil {
				matches = append(matches, lineNum)
			}
		}
	}

	// Sort by line number (fzf sorts by score)
	sort.Ints(matches)

	return matches, nil
}
