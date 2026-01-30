package vcs

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// FileStatus represents the status of a file change
type FileStatus string

const (
	StatusModified FileStatus = "M"
	StatusAdded    FileStatus = "A"
	StatusDeleted  FileStatus = "D"
	StatusRenamed  FileStatus = "R"
)

// FileChange represents a changed file
type FileChange struct {
	Path   string
	Status FileStatus
}

// VCS defines the interface for version control systems
type VCS interface {
	Name() string                        // "jj" or "git"
	ChangedFiles() ([]FileChange, error) // List of changed files
	Diff(path string) (string, error)    // Diff for specific file
	DiffAll() (string, error)            // Full diff
}

// Detect finds the appropriate VCS for the given directory
// Prefers jj over git if both exist
func Detect(dir string) (VCS, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory: %w", err)
	}

	// Check for jj first
	jjDir := filepath.Join(absDir, ".jj")
	if _, err := os.Stat(jjDir); err == nil {
		return &JJ{dir: absDir}, nil
	}

	// Fall back to git
	gitDir := filepath.Join(absDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return &Git{dir: absDir}, nil
	}

	return nil, fmt.Errorf("no VCS found (looking for .jj or .git in %s)", absDir)
}

// JJ implements VCS for jujutsu
type JJ struct {
	dir      string
	baseRev  string    // Cached base revision
	baseErr  error     // Cached error if resolution failed
	baseOnce sync.Once // Ensures base resolution happens only once
}

func (j *JJ) Name() string {
	return "jj"
}

// baseRevset is the revset expression to find the base revision for diffing.
// It finds the nearest bookmark ancestor, or falls back to trunk().
const baseRevset = "coalesce(heads(::@ & bookmarks()), trunk())"

// resolveBase determines the base revision for diffing.
// It returns the commit ID of the nearest bookmark ancestor, or trunk() as fallback.
// The result is cached so only one jj command is executed per session.
func (j *JJ) resolveBase() (string, error) {
	j.baseOnce.Do(func() {
		cmd := exec.Command("jj", "log", "-r", baseRevset, "-T", "commit_id", "--no-graph", "--limit", "1")
		cmd.Dir = j.dir
		output, err := cmd.Output()
		if err != nil {
			// Check if it's an exit error with stderr
			if exitErr, ok := err.(*exec.ExitError); ok {
				stderr := string(exitErr.Stderr)
				j.baseErr = fmt.Errorf("failed to resolve base revision: %s\nHint: Create a bookmark at your branch point, or ensure a 'main', 'master', or 'trunk' bookmark exists", strings.TrimSpace(stderr))
			} else {
				j.baseErr = fmt.Errorf("failed to resolve base revision: %w\nHint: Create a bookmark at your branch point, or ensure a 'main', 'master', or 'trunk' bookmark exists", err)
			}
			return
		}

		commitID := strings.TrimSpace(string(output))
		if commitID == "" {
			j.baseErr = fmt.Errorf("no base revision found: no bookmarks in ancestry and trunk() not found\nHint: Create a bookmark at your branch point, or ensure a 'main', 'master', or 'trunk' bookmark exists")
			return
		}

		j.baseRev = commitID
	})
	return j.baseRev, j.baseErr
}

func (j *JJ) ChangedFiles() ([]FileChange, error) {
	base, err := j.resolveBase()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("jj", "diff", "--from", base, "--to", "@", "--summary")
	cmd.Dir = j.dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("jj diff --summary failed: %w", err)
	}

	return parseJJSummary(string(output))
}

func (j *JJ) Diff(path string) (string, error) {
	base, err := j.resolveBase()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("jj", "diff", "--from", base, "--to", "@", path)
	cmd.Dir = j.dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("jj diff %s failed: %w", path, err)
	}
	return string(output), nil
}

func (j *JJ) DiffAll() (string, error) {
	base, err := j.resolveBase()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("jj", "diff", "--from", base, "--to", "@")
	cmd.Dir = j.dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("jj diff failed: %w", err)
	}
	return string(output), nil
}

// parseJJSummary parses output from "jj diff --summary"
// Format: M path/to/file
func parseJJSummary(output string) ([]FileChange, error) {
	var changes []FileChange
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: "M file.txt" or "A file.txt" etc.
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		status := FileStatus(strings.TrimSpace(parts[0]))
		path := strings.TrimSpace(parts[1])

		changes = append(changes, FileChange{
			Path:   path,
			Status: status,
		})
	}

	return changes, nil
}

// Git implements VCS for git
type Git struct {
	dir string
}

func (g *Git) Name() string {
	return "git"
}

func (g *Git) ChangedFiles() ([]FileChange, error) {
	// Get both staged and unstaged changes
	var changes []FileChange

	// Staged changes
	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	cmd.Dir = g.dir
	stagedOutput, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --cached failed: %w", err)
	}
	staged, err := parseGitNameStatus(string(stagedOutput))
	if err != nil {
		return nil, err
	}
	changes = append(changes, staged...)

	// Unstaged changes (only if not already in staged)
	cmd = exec.Command("git", "diff", "--name-status")
	cmd.Dir = g.dir
	unstagedOutput, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}
	unstaged, err := parseGitNameStatus(string(unstagedOutput))
	if err != nil {
		return nil, err
	}

	// Add unstaged files that aren't already in the list
	stagedPaths := make(map[string]bool)
	for _, c := range staged {
		stagedPaths[c.Path] = true
	}
	for _, c := range unstaged {
		if !stagedPaths[c.Path] {
			changes = append(changes, c)
		}
	}

	return changes, nil
}

func (g *Git) Diff(path string) (string, error) {
	var output bytes.Buffer
	var errs []string

	// Get staged diff
	cmd := exec.Command("git", "diff", "--cached", "--", path)
	cmd.Dir = g.dir
	stagedOutput, err := cmd.Output()
	if err != nil {
		errs = append(errs, fmt.Sprintf("staged diff: %v", err))
	}
	output.Write(stagedOutput)

	// Get unstaged diff
	cmd = exec.Command("git", "diff", "--", path)
	cmd.Dir = g.dir
	unstagedOutput, err := cmd.Output()
	if err != nil {
		errs = append(errs, fmt.Sprintf("unstaged diff: %v", err))
	}
	output.Write(unstagedOutput)

	// Only return error if both failed and we got no output
	if len(errs) == 2 && output.Len() == 0 {
		return "", fmt.Errorf("git diff failed: %s", strings.Join(errs, "; "))
	}

	return output.String(), nil
}

func (g *Git) DiffAll() (string, error) {
	var output bytes.Buffer
	var errs []string

	// Get staged diff
	cmd := exec.Command("git", "diff", "--cached")
	cmd.Dir = g.dir
	stagedOutput, err := cmd.Output()
	if err != nil {
		errs = append(errs, fmt.Sprintf("staged diff: %v", err))
	}
	output.Write(stagedOutput)

	// Get unstaged diff
	cmd = exec.Command("git", "diff")
	cmd.Dir = g.dir
	unstagedOutput, err := cmd.Output()
	if err != nil {
		errs = append(errs, fmt.Sprintf("unstaged diff: %v", err))
	}
	output.Write(unstagedOutput)

	// Only return error if both failed and we got no output
	if len(errs) == 2 && output.Len() == 0 {
		return "", fmt.Errorf("git diff failed: %s", strings.Join(errs, "; "))
	}

	return output.String(), nil
}

// parseGitNameStatus parses output from "git diff --name-status"
// Format: M\tpath/to/file
func parseGitNameStatus(output string) ([]FileChange, error) {
	var changes []FileChange
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: "M\tfile.txt" or "A\tfile.txt" etc.
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		status := FileStatus(strings.TrimSpace(parts[0]))
		path := strings.TrimSpace(parts[1])

		changes = append(changes, FileChange{
			Path:   path,
			Status: status,
		})
	}

	return changes, nil
}
