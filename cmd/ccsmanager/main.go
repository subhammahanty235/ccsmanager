/*
ccmanager - A terminal UI for managing Claude Code sessions.

Scans ~/.claude/projects for session history files and provides a three-panel
interface to browse, view, and resume conversations with the Claude CLI..
*/
package main

import (
	"fmt"
	"os"

	"github.com/subhammahanty235/ccsmanager/internal/app"
	"github.com/subhammahanty235/ccsmanager/internal/claude"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if !claude.CLIAvailable() {
		fmt.Fprintln(os.Stderr, "Warning: 'claude' CLI not found. Resume functionality requires it.")
		fmt.Fprintln(os.Stderr, "Install from: https://claude.ai/code")
		fmt.Fprintln(os.Stderr)
	}

	p := tea.NewProgram(
		app.NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
