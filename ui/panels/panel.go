package panels

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gerund/tcr/ui/borders"
)

// Panel defines the interface for all panels
type Panel interface {
	tea.Model
	Title() string
	ShortHelp() string
	SetFocused(bool)
	IsFocused() bool
	SetSize(width, height int)
}

// BasePanel provides common functionality for all panels
type BasePanel struct {
	title     string
	shortHelp string
	focused   bool
	width     int
	height    int
	cursor    int
}

// NewBasePanel creates a new base panel
func NewBasePanel(title, shortHelp string) BasePanel {
	return BasePanel{
		title:     title,
		shortHelp: shortHelp,
	}
}

func (b *BasePanel) Title() string {
	return b.title
}

func (b *BasePanel) SetTitle(title string) {
	b.title = title
}

func (b *BasePanel) ShortHelp() string {
	return b.shortHelp
}

func (b *BasePanel) SetFocused(focused bool) {
	b.focused = focused
}

func (b *BasePanel) IsFocused() bool {
	return b.focused
}

func (b *BasePanel) SetSize(width, height int) {
	b.width = width
	b.height = height
}

func (b *BasePanel) Width() int {
	return b.width
}

func (b *BasePanel) Height() int {
	return b.height
}

func (b *BasePanel) Cursor() int {
	return b.cursor
}

func (b *BasePanel) SetCursor(c int) {
	b.cursor = c
}

// ContentHeight returns the height available for content (minus borders)
func (b *BasePanel) ContentHeight() int {
	return b.height - 2
}

// ContentWidth returns the width available for content (minus borders)
func (b *BasePanel) ContentWidth() int {
	return b.width - 2
}

// RenderFrame renders the panel frame with title embedded in border
func (b *BasePanel) RenderFrame(content string) string {
	return borders.RenderTitledBorder(content, b.title, b.width, b.height, b.focused)
}

// CursorUp moves the cursor up within bounds
func (b *BasePanel) CursorUp(itemCount int) {
	if b.cursor > 0 {
		b.cursor--
	}
}

// CursorDown moves the cursor down within bounds
func (b *BasePanel) CursorDown(itemCount int) {
	if b.cursor < itemCount-1 {
		b.cursor++
	}
}

// CursorHome moves the cursor to the first item
func (b *BasePanel) CursorHome() {
	b.cursor = 0
}

// CursorEnd moves the cursor to the last item
func (b *BasePanel) CursorEnd(itemCount int) {
	if itemCount > 0 {
		b.cursor = itemCount - 1
	}
}

// CursorPageUp moves the cursor up by a page
func (b *BasePanel) CursorPageUp(itemCount, pageSize int) {
	b.cursor -= pageSize
	if b.cursor < 0 {
		b.cursor = 0
	}
}

// CursorPageDown moves the cursor down by a page
func (b *BasePanel) CursorPageDown(itemCount, pageSize int) {
	b.cursor += pageSize
	if b.cursor >= itemCount {
		b.cursor = itemCount - 1
	}
	if b.cursor < 0 {
		b.cursor = 0
	}
}
