package panels

import (
	"testing"
)

func TestSearchState_NewSearchState(t *testing.T) {
	s := NewSearchState()

	if s.active {
		t.Error("new SearchState should not be active")
	}
	if s.currentMatch != -1 {
		t.Errorf("currentMatch should be -1, got %d", s.currentMatch)
	}
	if len(s.matches) != 0 {
		t.Error("matches should be empty")
	}
}

func TestSearchState_Activate(t *testing.T) {
	s := NewSearchState()
	s.Activate()

	if !s.active {
		t.Error("should be active after Activate()")
	}
	if s.Query() != "" {
		t.Error("query should be empty after Activate()")
	}
}

func TestSearchState_Deactivate(t *testing.T) {
	s := NewSearchState()
	s.Activate()
	s.input.SetValue("test")
	s.Deactivate()

	if s.active {
		t.Error("should not be active after Deactivate()")
	}
}

func TestSearchState_NextMatch_Empty(t *testing.T) {
	s := NewSearchState()
	s.NextMatch() // Should not panic

	if s.currentMatch != -1 {
		t.Errorf("currentMatch should remain -1 with no matches")
	}
}

func TestSearchState_NextMatch_Wrapping(t *testing.T) {
	s := NewSearchState()
	s.matches = []int{5, 10, 15}
	s.currentMatch = 0

	s.NextMatch()
	if s.currentMatch != 1 {
		t.Errorf("expected 1, got %d", s.currentMatch)
	}

	s.NextMatch()
	if s.currentMatch != 2 {
		t.Errorf("expected 2, got %d", s.currentMatch)
	}

	s.NextMatch()
	if s.currentMatch != 0 {
		t.Errorf("expected wrap to 0, got %d", s.currentMatch)
	}
}

func TestSearchState_PrevMatch_Empty(t *testing.T) {
	s := NewSearchState()
	s.PrevMatch() // Should not panic

	if s.currentMatch != -1 {
		t.Errorf("currentMatch should remain -1 with no matches")
	}
}

func TestSearchState_PrevMatch_Wrapping(t *testing.T) {
	s := NewSearchState()
	s.matches = []int{5, 10, 15}
	s.currentMatch = 0

	s.PrevMatch()
	if s.currentMatch != 2 {
		t.Errorf("expected wrap to 2, got %d", s.currentMatch)
	}

	s.PrevMatch()
	if s.currentMatch != 1 {
		t.Errorf("expected 1, got %d", s.currentMatch)
	}
}

func TestSearchState_MatchStatus(t *testing.T) {
	tests := []struct {
		name         string
		matches      []int
		currentMatch int
		want         string
	}{
		{"no matches", nil, -1, "no matches"},
		{"first of three", []int{1, 2, 3}, 0, "1/3"},
		{"second of three", []int{1, 2, 3}, 1, "2/3"},
		{"single match", []int{5}, 0, "1/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSearchState()
			s.matches = tt.matches
			s.currentMatch = tt.currentMatch

			got := s.MatchStatus()
			if got != tt.want {
				t.Errorf("MatchStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSearchState_IsLineMatched(t *testing.T) {
	s := NewSearchState()
	s.matches = []int{5, 10, 15}
	s.matchSet = map[int]bool{5: true, 10: true, 15: true}

	if !s.IsLineMatched(5) {
		t.Error("line 5 should be matched")
	}
	if !s.IsLineMatched(10) {
		t.Error("line 10 should be matched")
	}
	if s.IsLineMatched(7) {
		t.Error("line 7 should not be matched")
	}
}

func TestSearchState_IsCurrentMatch(t *testing.T) {
	s := NewSearchState()
	s.matches = []int{5, 10, 15}
	s.currentMatch = 1 // Line 10

	if s.IsCurrentMatch(5) {
		t.Error("line 5 should not be current match")
	}
	if !s.IsCurrentMatch(10) {
		t.Error("line 10 should be current match")
	}
}

func TestSearchState_CurrentMatchLine(t *testing.T) {
	s := NewSearchState()

	// No matches
	if s.CurrentMatchLine() != -1 {
		t.Errorf("expected -1 with no matches, got %d", s.CurrentMatchLine())
	}

	// With matches
	s.matches = []int{5, 10, 15}
	s.currentMatch = 1
	if s.CurrentMatchLine() != 10 {
		t.Errorf("expected 10, got %d", s.CurrentMatchLine())
	}
}

func TestSearchState_HasMatches(t *testing.T) {
	s := NewSearchState()

	if s.HasMatches() {
		t.Error("should not have matches initially")
	}

	s.matches = []int{5, 10}
	if !s.HasMatches() {
		t.Error("should have matches after setting")
	}
}

func TestSearchState_Reset(t *testing.T) {
	s := NewSearchState()
	s.Activate()
	s.input.SetValue("test")
	s.matches = []int{1, 2, 3}
	s.currentMatch = 1

	s.Reset()

	if s.active {
		t.Error("should not be active after Reset()")
	}
	if s.Query() != "" {
		t.Error("query should be empty after Reset()")
	}
	if len(s.matches) != 0 {
		t.Error("matches should be empty after Reset()")
	}
	if s.currentMatch != -1 {
		t.Error("currentMatch should be -1 after Reset()")
	}
}

func TestSearchState_Query(t *testing.T) {
	s := NewSearchState()

	if s.Query() != "" {
		t.Error("query should be empty initially")
	}

	s.input.SetValue("hello")
	if s.Query() != "hello" {
		t.Errorf("expected 'hello', got %q", s.Query())
	}
}

func TestSearchState_SetWidth(t *testing.T) {
	s := NewSearchState()

	s.SetWidth(100)
	// Width should be 100 - 15 = 85
	if s.input.Width != 85 {
		t.Errorf("expected width 85, got %d", s.input.Width)
	}

	// Test minimum width
	s.SetWidth(10)
	if s.input.Width != 10 {
		t.Errorf("expected minimum width 10, got %d", s.input.Width)
	}
}

func TestDiffPanel_IsSearching(t *testing.T) {
	p := NewDiffPanel()

	if p.IsSearching() {
		t.Error("should not be searching initially")
	}

	p.searchState.Activate()
	if !p.IsSearching() {
		t.Error("should be searching after Activate()")
	}

	p.searchState.Deactivate()
	if p.IsSearching() {
		t.Error("should not be searching after Deactivate()")
	}
}

func TestDiffPanel_SearchPersistsOnSetDiff(t *testing.T) {
	p := NewDiffPanel()
	p.SetSize(80, 24)

	// Activate search and set some state
	p.searchState.Activate()
	p.searchState.matches = []int{1, 2, 3}
	p.searchState.currentMatch = 1

	// Set new diff should keep search active but clear old matches
	// (matches will be recalculated based on new content and current query)
	p.SetDiff("test.go", "line1\nline2\nline3")

	// Search mode should stay active (unified search behavior)
	if !p.searchState.active {
		t.Error("search should stay active after SetDiff")
	}
	// Old matches should be cleared (will be recalculated if query exists)
	if p.searchState.currentMatch != -1 {
		t.Error("currentMatch should be reset to -1 after SetDiff")
	}
}

func TestDiffPanel_SearchResetOnClearDiff(t *testing.T) {
	p := NewDiffPanel()
	p.SetSize(80, 24)
	p.SetDiff("test.go", "line1\nline2\nline3")

	// Activate search and set some state
	p.searchState.Activate()
	p.searchState.matches = []int{1, 2}
	p.searchState.currentMatch = 0

	// Clear diff should reset search
	p.ClearDiff()

	if p.searchState.active {
		t.Error("search should be deactivated after ClearDiff")
	}
	if len(p.searchState.matches) != 0 {
		t.Error("matches should be cleared after ClearDiff")
	}
}

func TestDiffPanel_ActivateDeactivateSearch(t *testing.T) {
	p := NewDiffPanel()
	p.SetSize(80, 24)
	p.SetDiff("test.go", "line1\nline2\nline3")

	// Activate search externally
	p.ActivateSearch()

	if !p.IsSearching() {
		t.Error("should be searching after ActivateSearch()")
	}

	// Deactivate search externally
	p.DeactivateSearch()

	if p.IsSearching() {
		t.Error("should not be searching after DeactivateSearch()")
	}
}

func TestDiffPanel_SetSearchQuery(t *testing.T) {
	p := NewDiffPanel()
	p.SetSize(80, 24)
	p.SetDiff("test.go", "line1 foo\nline2\nline3 foo bar")

	// Activate search, set query, and set matches (simulating app behavior)
	p.ActivateSearch()
	p.SetSearchQuery("foo")
	p.SetSearchMatches([]int{0, 2}) // Lines 0 and 2 contain "foo"

	// Should have matches set
	if !p.searchState.HasMatches() {
		t.Error("should have matches for 'foo'")
	}
	if p.MatchCount() != 2 {
		t.Errorf("expected 2 matches, got %d", p.MatchCount())
	}
}

func TestDiffPanel_CycleNextMatch(t *testing.T) {
	p := NewDiffPanel()
	p.SetSize(80, 24)
	p.SetDiff("test.go", "line1 foo\nline2\nline3 foo")

	// Activate search, set query, and set matches (simulating app behavior)
	p.ActivateSearch()
	p.SetSearchQuery("foo")
	p.SetSearchMatches([]int{0, 2}) // Lines 0 and 2 contain "foo"

	// Initial match should be first one (line 0)
	if p.CurrentMatchIndex() != 1 {
		t.Errorf("expected current match 1, got %d", p.CurrentMatchIndex())
	}
	if p.cursorLine != 0 {
		t.Errorf("cursor should be at line 0, got %d", p.cursorLine)
	}

	// Cycle to next match (line 2)
	p.CycleNextMatch()
	if p.CurrentMatchIndex() != 2 {
		t.Errorf("expected current match 2, got %d", p.CurrentMatchIndex())
	}
	if p.cursorLine != 2 {
		t.Errorf("cursor should be at line 2, got %d", p.cursorLine)
	}

	// Cycle again should wrap to first match
	p.CycleNextMatch()
	if p.CurrentMatchIndex() != 1 {
		t.Errorf("expected current match 1 after wrap, got %d", p.CurrentMatchIndex())
	}
}

func TestDiffPanel_SetSearchMatches(t *testing.T) {
	p := NewDiffPanel()
	p.SetSize(80, 24)

	// Set diff content
	p.SetDiff("test.go", "line1 foo\nline2\nline3 foo bar")

	// Activate search and set matches directly (simulating app behavior)
	p.ActivateSearch()
	p.SetSearchQuery("foo")
	p.SetSearchMatches([]int{0, 2}) // Lines 0 and 2 contain "foo"

	if !p.IsSearching() {
		t.Error("should be searching after ActivateSearch")
	}
	if p.searchState.Query() != "foo" {
		t.Errorf("expected query 'foo', got %q", p.searchState.Query())
	}
	if !p.searchState.HasMatches() {
		t.Error("should have matches")
	}
	if p.MatchCount() != 2 {
		t.Errorf("expected 2 matches, got %d", p.MatchCount())
	}
}

func TestDiffPanel_SetSearchMatches_Empty(t *testing.T) {
	p := NewDiffPanel()
	p.SetSize(80, 24)

	// Set diff content
	p.SetDiff("test.go", "line1\nline2")

	// Activate search with no matches
	p.ActivateSearch()
	p.SetSearchQuery("nonexistent")
	p.SetSearchMatches(nil)

	if p.searchState.HasMatches() {
		t.Error("should not have matches")
	}
}
