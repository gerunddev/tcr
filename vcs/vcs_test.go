package vcs

import (
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
	defer os.RemoveAll(tmpDir)

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
	defer os.RemoveAll(tmpDir2)

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
	defer os.RemoveAll(tmpDir)

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
	defer os.RemoveAll(tmpDir)

	_, err = Detect(tmpDir)
	if err == nil {
		t.Error("Expected error when no VCS found")
	}
}
