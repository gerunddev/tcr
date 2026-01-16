# Dev Notes: Diff Panel Cursor Improvements

## Implementation Summary

### Change 1: Cursor Color (theme.go)
- Changed `CursorLineStyle` background from `ColorSurface` (`#403E41`) to `#3D3500` (dark golden yellow)
- This provides much better visual contrast against the background (`#2D2A2E`)

### Change 2: Width Consistency (diff.go)
Refactored `renderContent()` with a new approach:

1. **Added `padToWidth()` helper** - Pads any string with spaces to reach target width
2. **Added `truncateLine()` method** - Extracted truncation logic from old `styleDiffLine()`
3. **Added `applyDiffColors()` method** - Applies only foreground colors based on diff line prefix
4. **Removed `styleDiffLine()`** - No longer needed, logic split between new functions

**New rendering flow:**
```
raw line -> truncate -> pad to width -> apply background (cursor/search) -> apply foreground colors
```

The key insight is padding BEFORE styling ensures the background always extends full width, regardless of the original line length.

## Files Modified
- `ui/theme/theme.go` - CursorLineStyle background color
- `ui/panels/diff.go` - renderContent() refactor, new helper functions

## Testing Notes
All existing tests pass. The plan specifies manual visual testing:
- [ ] Open diff with varying line lengths
- [ ] Verify yellow cursor highlight extends full width on all lines
- [ ] Verify add/remove/context lines show cursor properly
- [ ] Test search mode highlighting
- [ ] Test terminal resize behavior

## Areas for Extra Review Attention
- The order of styling (background then foreground) might behave differently with edge cases involving ANSI codes
- If users find `#3D3500` too subtle, alternative `#4D4400` was mentioned in the plan
