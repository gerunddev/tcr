package vcs

// Unit tests for the vcs package.
//
// These tests cover parsing logic and detection without requiring real VCS repositories.
// For integration tests that require jj/git to be installed, see vcs_integration_test.go
// and run with: go test -tags=integration ./vcs/...

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestParseJJSummary(t *testing.T) {
	input := `M src/main.go
A src/new.go
D src/deleted.go
`
	changes, err := parseJJSummary(input)
	if err != nil {
		t.Fatalf("parseJJSummary failed: %v", err)
	}

	expected := []FileChange{
		{Path: "src/main.go", Status: StatusModified},
		{Path: "src/new.go", Status: StatusAdded},
		{Path: "src/deleted.go", Status: StatusDeleted},
	}

	if len(changes) != len(expected) {
		t.Fatalf("Expected %d changes, got %d", len(expected), len(changes))
	}

	for i, c := range changes {
		if c.Path != expected[i].Path {
			t.Errorf("Change %d: expected path %q, got %q", i, expected[i].Path, c.Path)
		}
		if c.Status != expected[i].Status {
			t.Errorf("Change %d: expected status %q, got %q", i, expected[i].Status, c.Status)
		}
	}
}

func TestParseGitNameStatus(t *testing.T) {
	input := `M	src/main.go
A	src/new.go
D	src/deleted.go
`
	changes, err := parseGitNameStatus(input)
	if err != nil {
		t.Fatalf("parseGitNameStatus failed: %v", err)
	}

	expected := []FileChange{
		{Path: "src/main.go", Status: StatusModified},
		{Path: "src/new.go", Status: StatusAdded},
		{Path: "src/deleted.go", Status: StatusDeleted},
	}

	if len(changes) != len(expected) {
		t.Fatalf("Expected %d changes, got %d", len(expected), len(changes))
	}

	for i, c := range changes {
		if c.Path != expected[i].Path {
			t.Errorf("Change %d: expected path %q, got %q", i, expected[i].Path, c.Path)
		}
		if c.Status != expected[i].Status {
			t.Errorf("Change %d: expected status %q, got %q", i, expected[i].Status, c.Status)
		}
	}
}

func TestDetect(t *testing.T) {
	// Create temp directory with .jj
	tmpDir, err := os.MkdirTemp("", "tcr-test-jj-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	jjDir := filepath.Join(tmpDir, ".jj")
	if err := os.Mkdir(jjDir, 0755); err != nil {
		t.Fatal(err)
	}

	v, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if v.Name() != "jj" {
		t.Errorf("Expected jj, got %s", v.Name())
	}

	// Create temp directory with .git only
	tmpDir2, err := os.MkdirTemp("", "tcr-test-git-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir2) })

	gitDir := filepath.Join(tmpDir2, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	v, err = Detect(tmpDir2)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if v.Name() != "git" {
		t.Errorf("Expected git, got %s", v.Name())
	}
}

func TestDetectPreferJJ(t *testing.T) {
	// Create temp directory with both .jj and .git
	tmpDir, err := os.MkdirTemp("", "tcr-test-both-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	jjDir := filepath.Join(tmpDir, ".jj")
	if err := os.Mkdir(jjDir, 0755); err != nil {
		t.Fatal(err)
	}

	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	v, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if v.Name() != "jj" {
		t.Errorf("Expected jj (should prefer jj), got %s", v.Name())
	}
}

func TestDetectNoVCS(t *testing.T) {
	// Create temp directory without VCS
	tmpDir, err := os.MkdirTemp("", "tcr-test-novcs-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	_, err = Detect(tmpDir)
	if err == nil {
		t.Error("Expected error when no VCS found")
	}
}

func TestJJResolveBaseCaching(t *testing.T) {
	// Test that resolveBase caches the result and only runs once
	// This test doesn't require jj to be installed - it just verifies
	// that the caching mechanism works by calling resolveBase twice
	// on a JJ instance with pre-set values.

	jj := &JJ{
		dir:     "/nonexistent",
		baseRev: "abc123", // Pre-set to simulate successful resolution
	}

	// Mark that resolution has already happened
	jj.baseOnce.Do(func() {})

	// First call should return cached value
	rev1, err1 := jj.resolveBase()
	if err1 != nil {
		t.Errorf("Expected no error, got %v", err1)
	}
	if rev1 != "abc123" {
		t.Errorf("Expected cached rev 'abc123', got %q", rev1)
	}

	// Second call should return same cached value
	rev2, err2 := jj.resolveBase()
	if err2 != nil {
		t.Errorf("Expected no error, got %v", err2)
	}
	if rev2 != rev1 {
		t.Errorf("Expected same cached value, got different: %q vs %q", rev1, rev2)
	}
}

func TestJJResolveBaseCachesError(t *testing.T) {
	// Test that resolveBase also caches errors
	jj := &JJ{
		dir:     "/nonexistent",
		baseErr: fmt.Errorf("cached error"),
	}

	// Mark that resolution has already happened
	jj.baseOnce.Do(func() {})

	// Call should return cached error
	_, err := jj.resolveBase()
	if err == nil {
		t.Error("Expected cached error, got nil")
	}
	if err.Error() != "cached error" {
		t.Errorf("Expected 'cached error', got %q", err.Error())
	}
}

func TestBaseRevsetConstant(t *testing.T) {
	// Verify the revset constant is what we expect
	expected := "coalesce(heads(::@ & bookmarks()), trunk())"
	if baseRevset != expected {
		t.Errorf("baseRevset changed unexpectedly:\n  got:  %q\n  want: %q", baseRevset, expected)
	}
}

func TestParseJJSummaryEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []FileChange
	}{
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   \n   \n   ",
			expected: nil,
		},
		{
			name:     "single file",
			input:    "M foo.go",
			expected: []FileChange{{Path: "foo.go", Status: StatusModified}},
		},
		{
			name:     "renamed file",
			input:    "R old.go -> new.go",
			expected: []FileChange{{Path: "old.go -> new.go", Status: StatusRenamed}},
		},
		{
			name:     "path with spaces",
			input:    "A path with spaces/file.go",
			expected: []FileChange{{Path: "path with spaces/file.go", Status: StatusAdded}},
		},
		{
			name:  "mixed statuses",
			input: "M file1.go\nA file2.go\nD file3.go\nR file4.go",
			expected: []FileChange{
				{Path: "file1.go", Status: StatusModified},
				{Path: "file2.go", Status: StatusAdded},
				{Path: "file3.go", Status: StatusDeleted},
				{Path: "file4.go", Status: StatusRenamed},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJJSummary(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d changes, got %d", len(tt.expected), len(result))
			}
			for i, c := range result {
				if c.Path != tt.expected[i].Path || c.Status != tt.expected[i].Status {
					t.Errorf("change %d: expected %+v, got %+v", i, tt.expected[i], c)
				}
			}
		})
	}
}

func TestParseGitNameStatusEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []FileChange
	}{
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   \n   \n   ",
			expected: nil,
		},
		{
			name:     "single file",
			input:    "M\tfoo.go",
			expected: []FileChange{{Path: "foo.go", Status: StatusModified}},
		},
		{
			name:     "renamed file",
			input:    "R\told.go\tnew.go",
			expected: []FileChange{{Path: "old.go", Status: StatusRenamed}},
		},
		{
			name:  "mixed statuses",
			input: "M\tfile1.go\nA\tfile2.go\nD\tfile3.go\nR\tfile4.go",
			expected: []FileChange{
				{Path: "file1.go", Status: StatusModified},
				{Path: "file2.go", Status: StatusAdded},
				{Path: "file3.go", Status: StatusDeleted},
				{Path: "file4.go", Status: StatusRenamed},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGitNameStatus(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d changes, got %d", len(tt.expected), len(result))
			}
			for i, c := range result {
				if c.Path != tt.expected[i].Path || c.Status != tt.expected[i].Status {
					t.Errorf("change %d: expected %+v, got %+v", i, tt.expected[i], c)
				}
			}
		})
	}
}

func TestJJName(t *testing.T) {
	jj := &JJ{dir: "/tmp"}
	if jj.Name() != "jj" {
		t.Errorf("expected 'jj', got %q", jj.Name())
	}
}

func TestGitName(t *testing.T) {
	git := &Git{dir: "/tmp"}
	if git.Name() != "git" {
		t.Errorf("expected 'git', got %q", git.Name())
	}
}
