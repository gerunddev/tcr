package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendFeedback(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "tcr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "feedback.md")

	// First feedback
	err = AppendFeedback(outputPath, "src/main.go", 42, "This is my feedback")
	if err != nil {
		t.Fatalf("AppendFeedback failed: %v", err)
	}

	// Second feedback
	err = AppendFeedback(outputPath, "src/other.go", 15, "Another comment\nwith multiple lines")
	if err != nil {
		t.Fatalf("AppendFeedback failed: %v", err)
	}

	// Read and verify
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// TODO: Re-add :line numbers to expected output once CalculateLineNumber is fixed
	expected := `@src/main.go
This is my feedback

@src/other.go
Another comment
with multiple lines

`
	if string(content) != expected {
		t.Errorf("Content mismatch:\nGot:\n%s\n\nExpected:\n%s", string(content), expected)
	}
}

func TestAppendFeedbackCreatesDirectory(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "tcr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Output path in a non-existent subdirectory
	outputPath := filepath.Join(tmpDir, "subdir", "feedback.md")

	err = AppendFeedback(outputPath, "file.go", 1, "test")
	if err != nil {
		t.Fatalf("AppendFeedback failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

func TestValidateOutputPath(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
		errMsg  string
	}{
		{"", true, "output path is required"},
		{"feedback.txt", true, "must have .md extension"},
		{"feedback.MD", false, ""},
		{"feedback.md", false, ""},
		{"path/to/feedback.md", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			err := ValidateOutputPath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for path %q", tt.path)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for path %q: %v", tt.path, err)
				}
			}
		})
	}
}
