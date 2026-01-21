package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gerunddev/tcr/output"
	"github.com/gerunddev/tcr/ui"
	"github.com/gerunddev/tcr/vcs"
)

func main() {
	var outputPath string

	if len(os.Args) < 2 {
		// Generate a random filename in /tmp
		randomBytes := make([]byte, 8)
		if _, err := rand.Read(randomBytes); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating random filename: %v\n", err)
			os.Exit(1)
		}
		outputPath = filepath.Join("/tmp", "tcr-"+hex.EncodeToString(randomBytes)+".md")
		fmt.Fprintf(os.Stderr, "Output file: %s\n", outputPath)
	} else {
		outputPath = os.Args[1]
	}

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
