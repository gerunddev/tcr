package search

import (
	"testing"
)

func TestController_NewController(t *testing.T) {
	c := NewController()

	if c.IsActive() {
		t.Error("new controller should not be active")
	}
	if c.Query() != "" {
		t.Error("new controller should have empty query")
	}
	if c.FilteredIndices() != nil {
		t.Error("new controller should have nil filtered indices")
	}
}

func TestController_Activate(t *testing.T) {
	c := NewController()
	c.Activate()

	if !c.IsActive() {
		t.Error("controller should be active after Activate()")
	}
	if c.Query() != "" {
		t.Error("query should be empty after Activate()")
	}
	if c.HasNoMatches() {
		t.Error("should not report no matches before search runs")
	}
}

func TestController_Deactivate(t *testing.T) {
	c := NewController()
	c.Activate()
	c.Deactivate()

	if c.IsActive() {
		t.Error("controller should not be active after Deactivate()")
	}
	if c.Query() != "" {
		t.Error("query should be empty after Deactivate()")
	}
}

func TestController_SearchAllFiles_EmptyQuery(t *testing.T) {
	c := NewController()
	c.Activate()

	files := []string{"a.go", "b.go"}
	diffs := map[string]string{
		"a.go": "func main() {}",
		"b.go": "func test() {}",
	}

	c.SearchAllFiles("", files, diffs)

	if c.FilteredIndices() != nil {
		t.Error("empty query should result in nil filtered indices")
	}
	if c.HasNoMatches() {
		t.Error("empty query should not report no matches")
	}
}

func TestController_SearchAllFiles_WithMatches(t *testing.T) {
	c := NewController()
	c.Activate()

	files := []string{"a.go", "b.go", "c.go"}
	diffs := map[string]string{
		"a.go": "func main() { foo() }",
		"b.go": "func test() { bar() }",
		"c.go": "func other() { foo() }",
	}

	c.SearchAllFiles("foo", files, diffs)

	indices := c.FilteredIndices()
	if indices == nil {
		t.Fatal("expected non-nil filtered indices")
	}
	if len(indices) != 2 {
		t.Errorf("expected 2 matching files, got %d", len(indices))
	}
	// Should be files at index 0 (a.go) and 2 (c.go)
	if indices[0] != 0 || indices[1] != 2 {
		t.Errorf("expected indices [0, 2], got %v", indices)
	}
	if c.HasNoMatches() {
		t.Error("should not report no matches when matches exist")
	}
}

func TestController_SearchAllFiles_NoMatches(t *testing.T) {
	c := NewController()
	c.Activate()

	files := []string{"a.go", "b.go"}
	diffs := map[string]string{
		"a.go": "func main() {}",
		"b.go": "func test() {}",
	}

	c.SearchAllFiles("nonexistent", files, diffs)

	if c.FilteredIndices() != nil {
		t.Error("expected nil filtered indices when no matches")
	}
	if !c.HasNoMatches() {
		t.Error("should report no matches")
	}
}

func TestController_Status(t *testing.T) {
	c := NewController()

	// Before any search
	if c.Status() != "" {
		t.Errorf("expected empty status, got %q", c.Status())
	}

	c.Activate()
	files := []string{"a.go", "b.go"}
	diffs := map[string]string{
		"a.go": "func main() { foo() }",
		"b.go": "func test() {}",
	}

	// After search with matches (singular)
	c.SearchAllFiles("foo", files, diffs)
	status := c.Status()
	if status != "1 file" {
		t.Errorf("expected '1 file', got %q", status)
	}

	// After search with no matches
	c.SearchAllFiles("nonexistent", files, diffs)
	status = c.Status()
	if status != "no matches" {
		t.Errorf("expected 'no matches', got %q", status)
	}
}

func TestController_SearchInDiff(t *testing.T) {
	c := NewController()

	lines := []string{
		"package main",
		"",
		"func main() {",
		"    foo()",
		"}",
		"",
		"func foo() {",
		"}",
	}

	matches, err := c.SearchInDiff("foo", lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should match lines 3 and 6 (foo() and func foo())
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

func TestController_SearchInDiff_EmptyQuery(t *testing.T) {
	c := NewController()

	lines := []string{"line1", "line2"}
	matches, err := c.SearchInDiff("", lines)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches != nil {
		t.Error("empty query should return nil matches")
	}
}

func TestController_SearchInDiff_EmptyLines(t *testing.T) {
	c := NewController()

	matches, err := c.SearchInDiff("test", []string{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches != nil {
		t.Error("empty lines should return nil matches")
	}
}

func TestController_SetWidth(t *testing.T) {
	c := NewController()

	c.SetWidth(100)
	// Width should be 100 - 15 = 85
	if c.input.Width != 85 {
		t.Errorf("expected input width 85, got %d", c.input.Width)
	}

	// Test minimum width
	c.SetWidth(10)
	if c.input.Width != 10 {
		t.Errorf("expected minimum input width 10, got %d", c.input.Width)
	}
}
