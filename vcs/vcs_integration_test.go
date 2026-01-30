//go:build integration

package vcs

// Integration tests that require jj and/or git to be installed.
// Run with: go test -tags=integration ./vcs/...
//
// These tests verify real VCS interactions:
// - JJ: ChangedFiles(), Diff(), DiffAll(), resolveBase() with real repositories
// - Git: ChangedFiles(), Diff(), DiffAll() with real repositories
//
// To run locally:
//   1. Ensure jj and git are installed
//   2. Run: go test -tags=integration -v ./vcs/...

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestJJIntegration(t *testing.T) {
	// Skip if jj is not installed
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not installed, skipping integration test")
	}

	// Create a temporary directory for the test repo
	tmpDir, err := os.MkdirTemp("", "tcr-jj-integration-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Initialize jj repo
	cmd := exec.Command("jj", "git", "init")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("jj git init failed: %v\n%s", err, out)
	}

	// Create trunk bookmark on root commit
	cmd = exec.Command("jj", "bookmark", "create", "trunk", "-r", "root()")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("jj bookmark create trunk failed: %v\n%s", err, out)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test detection
	vcs, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if vcs.Name() != "jj" {
		t.Errorf("expected jj, got %s", vcs.Name())
	}

	// Test ChangedFiles
	changes, err := vcs.ChangedFiles()
	if err != nil {
		t.Fatalf("ChangedFiles failed: %v", err)
	}
	if len(changes) == 0 {
		t.Error("expected at least one changed file")
	}
	found := false
	for _, c := range changes {
		if c.Path == "test.txt" {
			found = true
			if c.Status != StatusAdded {
				t.Errorf("expected status Added, got %s", c.Status)
			}
		}
	}
	if !found {
		t.Error("test.txt not found in changed files")
	}

	// Test Diff for specific file
	diff, err := vcs.Diff("test.txt")
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !strings.Contains(diff, "hello world") {
		t.Errorf("diff should contain 'hello world', got: %s", diff)
	}

	// Test DiffAll
	diffAll, err := vcs.DiffAll()
	if err != nil {
		t.Fatalf("DiffAll failed: %v", err)
	}
	if !strings.Contains(diffAll, "hello world") {
		t.Errorf("diffAll should contain 'hello world', got: %s", diffAll)
	}
}

func TestGitIntegration(t *testing.T) {
	// Skip if git is not installed
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed, skipping integration test")
	}

	// Create a temporary directory for the test repo
	tmpDir, err := os.MkdirTemp("", "tcr-git-integration-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	// Configure git for the test
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config email failed: %v\n%s", err, out)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config name failed: %v\n%s", err, out)
	}

	// Create initial commit so we have something to diff against
	readme := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readme, []byte("# Test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test detection (ensure .jj doesn't exist)
	vcs, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if vcs.Name() != "git" {
		t.Errorf("expected git, got %s", vcs.Name())
	}

	// Stage the file for testing staged changes
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}

	// Test ChangedFiles
	changes, err := vcs.ChangedFiles()
	if err != nil {
		t.Fatalf("ChangedFiles failed: %v", err)
	}
	if len(changes) == 0 {
		t.Error("expected at least one changed file")
	}
	found := false
	for _, c := range changes {
		if c.Path == "test.txt" {
			found = true
			if c.Status != StatusAdded {
				t.Errorf("expected status Added, got %s", c.Status)
			}
		}
	}
	if !found {
		t.Error("test.txt not found in changed files")
	}

	// Test Diff for specific file
	diff, err := vcs.Diff("test.txt")
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !strings.Contains(diff, "hello world") {
		t.Errorf("diff should contain 'hello world', got: %s", diff)
	}

	// Test DiffAll
	diffAll, err := vcs.DiffAll()
	if err != nil {
		t.Fatalf("DiffAll failed: %v", err)
	}
	if !strings.Contains(diffAll, "hello world") {
		t.Errorf("diffAll should contain 'hello world', got: %s", diffAll)
	}
}

func TestJJResolveBaseWithNoBookmarks(t *testing.T) {
	// Skip if jj is not installed
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not installed, skipping integration test")
	}

	// Create a temporary directory for the test repo
	tmpDir, err := os.MkdirTemp("", "tcr-jj-nobook-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Initialize jj repo without any bookmarks
	cmd := exec.Command("jj", "git", "init")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("jj git init failed: %v\n%s", err, out)
	}

	jj := &JJ{dir: tmpDir}

	// This should fail or return trunk() as fallback
	_, err = jj.resolveBase()
	// We expect this might fail if no trunk exists, or succeed with trunk
	// The important thing is it doesn't panic
	if err != nil {
		// Verify the error message is helpful
		if !strings.Contains(err.Error(), "bookmark") && !strings.Contains(err.Error(), "trunk") {
			t.Errorf("error message should mention bookmarks or trunk: %v", err)
		}
	}
}
