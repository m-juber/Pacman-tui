package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"pacman-tui/internal/tui"
)

func main() {
	p := tea.NewProgram(
		tui.NewModel(),
		tea.WithAltScreen(),       // Use the alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running pacman-tui: %v\n", err)
		os.Exit(1)
	}
}
