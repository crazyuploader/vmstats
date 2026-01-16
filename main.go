package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	domain := "noble_default"

	// Allow domain to be passed as argument
	if len(os.Args) > 1 {
		domain = os.Args[1]
	}

	p := tea.NewProgram(initialModel(domain), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
