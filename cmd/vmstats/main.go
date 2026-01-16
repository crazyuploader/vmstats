package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/crazyuploader/vmstats/internal/stats"
	"github.com/crazyuploader/vmstats/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse flags
	domain := flag.String("domain", "noble_default", "The libvirt domain to monitor")
	logFile := flag.String("log", "vmstats.log", "Log file path")
	flag.Parse()

	// Setup logging
	f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Printf("Starting vmstats for domain: %s", *domain)

	// Initialize collector
	collector := stats.NewVirshCollector()

	// Initialize Bubble Tea program
	p := tea.NewProgram(ui.InitialModel(*domain, collector), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
