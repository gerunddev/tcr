package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AppendFeedback appends a feedback comment to the output file
// Format:
// @relative/path:line
// comment text here
// that can span multiple lines
//
func AppendFeedback(outputPath, filePath string, line int, comment string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Open file for appending (create if not exists)
	f, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer f.Close()

	// Format the feedback
	// @path:line (or @path if line is 0)
	// comment
	//
	var feedback string
	if line > 0 {
		feedback = fmt.Sprintf("@%s:%d\n%s\n\n", filePath, line, strings.TrimSpace(comment))
	} else {
		feedback = fmt.Sprintf("@%s\n%s\n\n", filePath, strings.TrimSpace(comment))
	}

	if _, err := f.WriteString(feedback); err != nil {
		return fmt.Errorf("failed to write feedback: %w", err)
	}

	return nil
}

// ValidateOutputPath checks if the output path is valid
func ValidateOutputPath(path string) error {
	if path == "" {
		return fmt.Errorf("output path is required")
	}

	// Check for .md extension
	if !strings.HasSuffix(strings.ToLower(path), ".md") {
		return fmt.Errorf("output file must have .md extension")
	}

	// Check if path is writable by checking parent directory
	dir := filepath.Dir(path)
	if dir == "" {
		dir = "."
	}

	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, that's OK we'll create it
			return nil
		}
		return fmt.Errorf("cannot access directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	return nil
}
