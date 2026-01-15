package borders

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderTitledBorder(t *testing.T) {
	tests := []struct {
		name    string
		content string
		title   string
		width   int
		height  int
		focused bool
	}{
		{
			name:    "basic border",
			content: "Hello",
			title:   "Test",
			width:   20,
			height:  5,
			focused: false,
		},
		{
			name:    "focused border",
			content: "Content",
			title:   "Title",
			width:   15,
			height:  4,
			focused: true,
		},
		{
			name:    "multi-line content",
			content: "Line 1\nLine 2\nLine 3",
			title:   "Multi",
			width:   20,
			height:  6,
			focused: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderTitledBorder(tt.content, tt.title, tt.width, tt.height, tt.focused)

			lines := strings.Split(result, "\n")

			// Should have correct number of lines
			if len(lines) != tt.height {
				t.Errorf("Expected %d lines, got %d", tt.height, len(lines))
			}

			// Top border should contain title
			if !strings.Contains(lines[0], tt.title) {
				t.Errorf("Top border should contain title %q, got %q", tt.title, lines[0])
			}

			// Top border should start with corner
			if !strings.HasPrefix(lines[0], TopLeft) {
				t.Errorf("Top border should start with %q", TopLeft)
			}

			// Bottom border should have corners
			if !strings.HasPrefix(lines[len(lines)-1], BottomLeft) {
				t.Errorf("Bottom border should start with %q", BottomLeft)
			}
		})
	}
}

func TestRenderTitledBorderMinimumSize(t *testing.T) {
	// Too small dimensions should return content as-is
	result := RenderTitledBorder("test", "title", 2, 1, false)
	if result != "test" {
		t.Errorf("Expected content returned for too-small dimensions")
	}
}

func TestPadOrTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  int // expected width
	}{
		{
			name:  "pad short string",
			input: "hi",
			width: 10,
			want:  10,
		},
		{
			name:  "exact width",
			input: "hello",
			width: 5,
			want:  5,
		},
		{
			name:  "truncate long string",
			input: "this is a very long string",
			width: 10,
			want:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padOrTruncate(tt.input, tt.width)
			gotWidth := lipgloss.Width(result)
			if gotWidth != tt.want {
				t.Errorf("padOrTruncate(%q, %d) width = %d, want %d", tt.input, tt.width, gotWidth, tt.want)
			}
		})
	}
}
