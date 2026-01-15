# tcr - Terminal Code Review

A lightweight terminal tool for reviewing diffs and capturing feedback locally. Built for reviewing AI-generated code changes—browse diffs interactively, add comments on specific lines, and export all feedback to a markdown file that can be fed back into AI models for fixes.

## Usage

```bash
tcr <output.md>
```

Run `tcr` with a markdown file path. The tool will detect your VCS (Git or Jujutsu) and display all changed files.

## Navigation

| Key | Action |
|-----|--------|
| `up/down` | Navigate files |
| `ctrl+n/p` | Next/previous diff line |
| `ctrl+v` / `alt+v` | Page down/up |
| `/` | Search in diff |
| `enter` | Add feedback on current line |
| `q` | Quit |

## Adding Feedback

Press `enter` on any diff line to open the feedback modal. Write your comment and press `enter` to save. Comments are appended to your output file in this format:

```markdown
@src/example.go:42
This function should handle the error case

@src/other.go:17
Consider using a constant here
```

## AI Workflow

1. Let an AI agent make changes to your codebase
2. Run `tcr feedback.md` to review the diff
3. Add comments on lines that need fixes
4. Feed `feedback.md` back to the AI to address your feedback
