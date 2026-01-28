package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/netdisco-tui/netdisco-tui/internal/api"
	"github.com/netdisco-tui/netdisco-tui/internal/config"
	"github.com/netdisco-tui/netdisco-tui/internal/tui"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := api.New(cfg)

	p := tea.NewProgram(
		tui.NewRootModel(120, 40, client),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
