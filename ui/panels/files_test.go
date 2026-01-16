package panels

import (
	"testing"

	"github.com/gerunddev/tcr/vcs"
)

func TestFilesPanel_SetFiles(t *testing.T) {
	p := NewFilesPanel()
	p.SetSize(30, 10)

	files := []vcs.FileChange{
		{Path: "a.go", Status: vcs.StatusModified},
		{Path: "b.go", Status: vcs.StatusAdded},
		{Path: "c.go", Status: vcs.StatusDeleted},
	}

	p.SetFiles(files)

	if p.Count() != 3 {
		t.Errorf("expected 3 files, got %d", p.Count())
	}
	if p.TotalCount() != 3 {
		t.Errorf("expected 3 total files, got %d", p.TotalCount())
	}
}

func TestFilesPanel_Filtering(t *testing.T) {
	p := NewFilesPanel()
	p.SetSize(30, 10)

	files := []vcs.FileChange{
		{Path: "a.go", Status: vcs.StatusModified},
		{Path: "b.go", Status: vcs.StatusAdded},
		{Path: "c.go", Status: vcs.StatusDeleted},
		{Path: "d.go", Status: vcs.StatusModified},
	}
	p.SetFiles(files)

	// Initially no filter
	if p.IsFiltered() {
		t.Error("should not be filtered initially")
	}
	if p.Count() != 4 {
		t.Errorf("expected 4 visible files, got %d", p.Count())
	}

	// Set filter to show only files at index 0 and 2
	p.SetFilteredIndices([]int{0, 2})

	if !p.IsFiltered() {
		t.Error("should be filtered after SetFilteredIndices")
	}
	if p.Count() != 2 {
		t.Errorf("expected 2 visible files, got %d", p.Count())
	}
	if p.TotalCount() != 4 {
		t.Errorf("total count should still be 4, got %d", p.TotalCount())
	}

	// Clear filter
	p.ClearFilter()

	if p.IsFiltered() {
		t.Error("should not be filtered after ClearFilter")
	}
	if p.Count() != 4 {
		t.Errorf("expected 4 visible files after clear, got %d", p.Count())
	}
}

func TestFilesPanel_FilteredNavigation(t *testing.T) {
	p := NewFilesPanel()
	p.SetSize(30, 10)

	files := []vcs.FileChange{
		{Path: "a.go", Status: vcs.StatusModified}, // index 0
		{Path: "b.go", Status: vcs.StatusAdded},    // index 1
		{Path: "c.go", Status: vcs.StatusDeleted},  // index 2
		{Path: "d.go", Status: vcs.StatusModified}, // index 3
	}
	p.SetFiles(files)

	// Set filter to show only files at index 0 and 2
	p.SetFilteredIndices([]int{0, 2})

	// Cursor should be at file index 0 (first filtered file)
	selected := p.SelectedFile()
	if selected == nil || selected.Path != "a.go" {
		t.Error("cursor should be on a.go")
	}

	// Navigate down should go to c.go (index 2), skipping b.go
	p.cursorDownFiltered()
	selected = p.SelectedFile()
	if selected == nil || selected.Path != "c.go" {
		t.Errorf("expected c.go after down, got %v", selected)
	}

	// Navigate down again should stay at c.go (end of filtered list)
	p.cursorDownFiltered()
	selected = p.SelectedFile()
	if selected == nil || selected.Path != "c.go" {
		t.Errorf("expected c.go at end, got %v", selected)
	}

	// Navigate up should go back to a.go
	p.cursorUpFiltered()
	selected = p.SelectedFile()
	if selected == nil || selected.Path != "a.go" {
		t.Errorf("expected a.go after up, got %v", selected)
	}

	// Navigate up again should stay at a.go (start of filtered list)
	p.cursorUpFiltered()
	selected = p.SelectedFile()
	if selected == nil || selected.Path != "a.go" {
		t.Errorf("expected a.go at start, got %v", selected)
	}
}

func TestFilesPanel_FilterPreservesCursor(t *testing.T) {
	p := NewFilesPanel()
	p.SetSize(30, 10)

	files := []vcs.FileChange{
		{Path: "a.go", Status: vcs.StatusModified}, // index 0
		{Path: "b.go", Status: vcs.StatusAdded},    // index 1
		{Path: "c.go", Status: vcs.StatusDeleted},  // index 2
	}
	p.SetFiles(files)

	// Move cursor to b.go (index 1)
	p.cursor = 1

	// Set filter that includes current file
	p.SetFilteredIndices([]int{1, 2})

	selected := p.SelectedFile()
	if selected == nil || selected.Path != "b.go" {
		t.Error("cursor should stay on b.go when filter includes it")
	}

	// Now set filter that excludes current file
	p.SetFilteredIndices([]int{0, 2})

	selected = p.SelectedFile()
	if selected == nil || selected.Path != "a.go" {
		t.Errorf("cursor should move to first filtered file, got %v", selected)
	}
}

func TestFilesPanel_FilePaths(t *testing.T) {
	p := NewFilesPanel()
	p.SetSize(30, 10)

	files := []vcs.FileChange{
		{Path: "a.go", Status: vcs.StatusModified},
		{Path: "b.go", Status: vcs.StatusAdded},
		{Path: "c.go", Status: vcs.StatusDeleted},
	}
	p.SetFiles(files)

	paths := p.FilePaths()

	if len(paths) != 3 {
		t.Errorf("expected 3 paths, got %d", len(paths))
	}
	if paths[0] != "a.go" || paths[1] != "b.go" || paths[2] != "c.go" {
		t.Errorf("unexpected paths: %v", paths)
	}
}

func TestFilesPanel_IndexConversion(t *testing.T) {
	p := NewFilesPanel()
	p.SetSize(30, 10)

	files := []vcs.FileChange{
		{Path: "a.go", Status: vcs.StatusModified}, // index 0
		{Path: "b.go", Status: vcs.StatusAdded},    // index 1
		{Path: "c.go", Status: vcs.StatusDeleted},  // index 2
		{Path: "d.go", Status: vcs.StatusModified}, // index 3
	}
	p.SetFiles(files)

	// Without filter, display index equals file index
	if p.displayIndexToFileIndex(1) != 1 {
		t.Error("without filter, display index should equal file index")
	}
	if p.fileIndexToDisplayIndex(2) != 2 {
		t.Error("without filter, file index should equal display index")
	}

	// With filter [0, 2, 3]
	p.SetFilteredIndices([]int{0, 2, 3})

	// Display index 0 -> file index 0
	// Display index 1 -> file index 2
	// Display index 2 -> file index 3
	if p.displayIndexToFileIndex(0) != 0 {
		t.Errorf("expected file index 0, got %d", p.displayIndexToFileIndex(0))
	}
	if p.displayIndexToFileIndex(1) != 2 {
		t.Errorf("expected file index 2, got %d", p.displayIndexToFileIndex(1))
	}
	if p.displayIndexToFileIndex(2) != 3 {
		t.Errorf("expected file index 3, got %d", p.displayIndexToFileIndex(2))
	}

	// File index 0 -> display index 0
	// File index 2 -> display index 1
	// File index 1 -> display index -1 (not in filter)
	if p.fileIndexToDisplayIndex(0) != 0 {
		t.Errorf("expected display index 0, got %d", p.fileIndexToDisplayIndex(0))
	}
	if p.fileIndexToDisplayIndex(2) != 1 {
		t.Errorf("expected display index 1, got %d", p.fileIndexToDisplayIndex(2))
	}
	if p.fileIndexToDisplayIndex(1) != -1 {
		t.Errorf("expected -1 for file not in filter, got %d", p.fileIndexToDisplayIndex(1))
	}
}
