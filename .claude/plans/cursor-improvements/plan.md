# Plan: Diff Panel Cursor Improvements

## Problem Statement

Two issues with the diff panel cursor:
1. **Visibility**: The grey background color (`#403E41`) is barely distinguishable from the panel background (`#2D2A2E`)
2. **Width consistency**: The cursor highlight doesn't always extend across the full panel width

## Root Cause Analysis

### Issue 1: Color
- `ui/theme/theme.go` line 82-83 defines `CursorLineStyle` with `Background(ColorSurface)`
- `ColorSurface` is `#403E41` - only ~16 brightness units different from background `#2D2A2E`

### Issue 2: Width
- In `ui/panels/diff.go` `renderContent()` (lines 519-551):
  1. `styleDiffLine()` applies foreground colors but doesn't pad to full width
  2. `CursorLineStyle.Width(contentWidth).Render(styledLine)` is applied to already-styled string
  3. ANSI codes in the styled string can interfere with lipgloss width calculations

## Proposed Solution

### Change 1: Update cursor color to yellow

**File:** `ui/theme/theme.go`

**Current (lines 81-84):**
```go
CursorLineStyle = lipgloss.NewStyle().
    Background(ColorSurface)
```

**New:**
```go
// Semi-transparent yellow cursor - visible but not overpowering
CursorLineStyle = lipgloss.NewStyle().
    Background(lipgloss.Color("#3D3500"))  // Dark golden yellow
```

Alternative if user prefers more vibrant:
```go
CursorLineStyle = lipgloss.NewStyle().
    Background(lipgloss.Color("#4D4400"))  // Slightly brighter gold
```

### Change 2: Fix width consistency in diff panel

**File:** `ui/panels/diff.go`

**Approach:** Pad the line to full width BEFORE applying any styling, ensuring the background extends fully.

**Current flow:**
```
raw line → styleDiffLine() → CursorLineStyle.Width().Render()
```

**New flow:**
```
raw line → pad to width → apply cursor background (if cursor) → apply foreground colors
```

**Modified `renderContent()` function:**

```go
func (p *DiffPanel) renderContent() string {
    if len(p.lines) == 0 {
        return ""
    }

    contentWidth := p.ContentWidth()
    var rendered []string

    for i, line := range p.lines {
        // Determine cursor/search state for this line
        isCursorLine := (i == p.cursorLine)
        isSearchActive := p.searchState.active && p.searchState.HasMatches()
        isCurrentMatch := isSearchActive && p.searchState.IsCurrentMatch(i)
        isOtherMatch := isSearchActive && p.searchState.IsLineMatched(i) && !isCurrentMatch

        // Step 1: Pad raw line to full content width
        // This ensures background colors extend fully
        padded := padToWidth(line, contentWidth)

        // Step 2: Apply background style based on state
        var withBackground string
        if isCurrentMatch {
            withBackground = theme.SearchCurrentLineStyle.Render(padded)
        } else if isOtherMatch {
            withBackground = theme.SearchMatchLineStyle.Render(padded)
        } else if isCursorLine {
            withBackground = theme.CursorLineStyle.Render(padded)
        } else {
            withBackground = padded
        }

        // Step 3: Apply diff foreground colors
        styledLine := p.applyDiffColors(withBackground, line)

        rendered = append(rendered, styledLine)
    }

    return strings.Join(rendered, "\n")
}

// padToWidth pads a string with spaces to reach the target width
func padToWidth(s string, width int) string {
    currentWidth := lipgloss.Width(s)
    if currentWidth >= width {
        return s
    }
    return s + strings.Repeat(" ", width-currentWidth)
}

// applyDiffColors applies foreground colors based on diff line type
// This replaces styleDiffLine but only handles foreground colors
func (p *DiffPanel) applyDiffColors(styledLine, originalLine string) string {
    // Determine the color based on original line prefix
    var fg lipgloss.Style
    if strings.HasPrefix(originalLine, "+") && !strings.HasPrefix(originalLine, "+++") {
        fg = theme.DiffAddLine
    } else if strings.HasPrefix(originalLine, "-") && !strings.HasPrefix(originalLine, "---") {
        fg = theme.DiffRemoveLine
    } else if strings.HasPrefix(originalLine, "@@") {
        fg = theme.DiffHunkHeader
    } else if strings.HasPrefix(originalLine, "diff ") || strings.HasPrefix(originalLine, "index ") ||
        strings.HasPrefix(originalLine, "---") || strings.HasPrefix(originalLine, "+++") {
        fg = theme.DiffHunkHeader
    } else {
        fg = theme.DiffContextLine
    }

    return fg.Render(styledLine)
}
```

**Also update or remove `styleDiffLine()`** since its logic is now split.

## Files to Modify

| File | Changes |
|------|---------|
| `ui/theme/theme.go` | Change `CursorLineStyle` background from `ColorSurface` to `#3D3500` |
| `ui/panels/diff.go` | Refactor `renderContent()` to pad lines first, then apply background, then foreground |

## Testing

1. Open a diff with varying line lengths (short and long lines)
2. Move cursor up/down - verify yellow highlight extends full width on all lines
3. Verify add lines (green), remove lines (red), and context lines all show cursor properly
4. Test search mode - verify search highlighting still works correctly
5. Resize terminal - verify cursor width adjusts properly

## Risks

- **Low**: Color change is isolated to theme file
- **Medium**: Refactoring `renderContent()` could affect ANSI code handling; test thoroughly

## Alternatives Considered

1. **Use lipgloss.Width() on already-styled lines**: Rejected - doesn't reliably work with ANSI codes
2. **Apply Width() in styleDiffLine()**: Rejected - would require passing cursor state into that function, poor separation
3. **Use different cursor indicator (e.g., prefix character)**: Rejected - user specifically wants background highlight
