package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ReubenPercival/flux/internal/monitor"
	"github.com/ReubenPercival/flux/internal/ui"
)

func main() {
	// Initialize the system monitor
	mon := monitor.NewMonitor()

	// Create the UI model
	model := ui.NewModel(mon)

	// Create and run the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
