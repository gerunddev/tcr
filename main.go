package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gerunddev/tcr/output"
	"github.com/gerunddev/tcr/ui"
	"github.com/gerunddev/tcr/vcs"
)

func main() {
	// Check for positional argument
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: output file is required")
		fmt.Fprintln(os.Stderr, "Usage: tcr <output.md>")
		os.Exit(1)
	}

	outputPath := os.Args[1]

	if err := output.ValidateOutputPath(outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Detect VCS
	v, err := vcs.Detect(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create and run app
	app := ui.NewApp(v, outputPath)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
