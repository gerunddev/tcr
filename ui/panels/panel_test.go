package panels

import "testing"

func TestBasePanel_CursorNavigation(t *testing.T) {
	p := NewBasePanel("Test", "test panel")

	// Initial cursor should be 0
	if p.Cursor() != 0 {
		t.Errorf("Initial cursor should be 0, got %d", p.Cursor())
	}

	itemCount := 10

	// Test cursor down
	p.CursorDown(itemCount)
	if p.Cursor() != 1 {
		t.Errorf("After CursorDown, cursor should be 1, got %d", p.Cursor())
	}

	// Test cursor up
	p.CursorUp(itemCount)
	if p.Cursor() != 0 {
		t.Errorf("After CursorUp, cursor should be 0, got %d", p.Cursor())
	}

	// Test cursor up at boundary (should not go negative)
	p.CursorUp(itemCount)
	if p.Cursor() != 0 {
		t.Errorf("CursorUp at 0 should stay at 0, got %d", p.Cursor())
	}

	// Test cursor end
	p.CursorEnd(itemCount)
	if p.Cursor() != 9 {
		t.Errorf("CursorEnd should go to %d, got %d", itemCount-1, p.Cursor())
	}

	// Test cursor down at boundary (should not exceed itemCount-1)
	p.CursorDown(itemCount)
	if p.Cursor() != 9 {
		t.Errorf("CursorDown at end should stay at %d, got %d", itemCount-1, p.Cursor())
	}

	// Test cursor home
	p.CursorHome()
	if p.Cursor() != 0 {
		t.Errorf("CursorHome should go to 0, got %d", p.Cursor())
	}
}

func TestBasePanel_PageNavigation(t *testing.T) {
	p := NewBasePanel("Test", "test panel")
	itemCount := 100
	pageSize := 10

	// Page down
	p.CursorPageDown(itemCount, pageSize)
	if p.Cursor() != 10 {
		t.Errorf("After first PageDown, cursor should be 10, got %d", p.Cursor())
	}

	// Page down again
	p.CursorPageDown(itemCount, pageSize)
	if p.Cursor() != 20 {
		t.Errorf("After second PageDown, cursor should be 20, got %d", p.Cursor())
	}

	// Page up
	p.CursorPageUp(itemCount, pageSize)
	if p.Cursor() != 10 {
		t.Errorf("After PageUp, cursor should be 10, got %d", p.Cursor())
	}

	// Page up past beginning
	p.CursorPageUp(itemCount, pageSize)
	p.CursorPageUp(itemCount, pageSize)
	if p.Cursor() != 0 {
		t.Errorf("PageUp past beginning should clamp to 0, got %d", p.Cursor())
	}

	// Go near end and page down past
	p.SetCursor(95)
	p.CursorPageDown(itemCount, pageSize)
	if p.Cursor() != 99 {
		t.Errorf("PageDown past end should clamp to %d, got %d", itemCount-1, p.Cursor())
	}
}

func TestBasePanel_SetSize(t *testing.T) {
	p := NewBasePanel("Test", "test panel")
	p.SetSize(80, 24)

	if p.Width() != 80 {
		t.Errorf("Width should be 80, got %d", p.Width())
	}
	if p.Height() != 24 {
		t.Errorf("Height should be 24, got %d", p.Height())
	}
	if p.ContentWidth() != 78 {
		t.Errorf("ContentWidth should be 78 (80-2 for borders), got %d", p.ContentWidth())
	}
	if p.ContentHeight() != 22 {
		t.Errorf("ContentHeight should be 22 (24-2 for borders), got %d", p.ContentHeight())
	}
}

func TestBasePanel_Focus(t *testing.T) {
	p := NewBasePanel("Test", "test panel")

	if p.IsFocused() {
		t.Error("Panel should not be focused initially")
	}

	p.SetFocused(true)
	if !p.IsFocused() {
		t.Error("Panel should be focused after SetFocused(true)")
	}

	p.SetFocused(false)
	if p.IsFocused() {
		t.Error("Panel should not be focused after SetFocused(false)")
	}
}

func TestBasePanel_Title(t *testing.T) {
	p := NewBasePanel("Initial", "help")

	if p.Title() != "Initial" {
		t.Errorf("Title should be 'Initial', got %q", p.Title())
	}

	p.SetTitle("Updated")
	if p.Title() != "Updated" {
		t.Errorf("Title should be 'Updated', got %q", p.Title())
	}
}
