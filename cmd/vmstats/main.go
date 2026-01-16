package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/crazyuploader/vmstats/internal/stats"
	"github.com/crazyuploader/vmstats/internal/ui"
)

func main() {
	// Parse flags
	domainsFlag := flag.String("domains", "", "Comma-separated list of libvirt domains to monitor (empty for all)")
	logFile := flag.String("log", "vmstats.log", "Log file path")
	flag.Parse()

	// Parse domains
	var domains []string
	if *domainsFlag != "" {
		domains = strings.Split(*domainsFlag, ",")
		for i := range domains {
			domains[i] = strings.TrimSpace(domains[i])
		}
	}

	// Setup logging
	f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing log file: %v\n", err)
		}
	}()
	log.SetOutput(f)

	if len(domains) > 0 {
		log.Printf("Starting vmstats for domains: %v", domains)
	} else {
		log.Printf("Starting vmstats for ALL domains")
	}

	// Initialize collector
	collector := stats.NewVirshCollector()

	// Initialize Bubble Tea program
	p := tea.NewProgram(ui.InitialModel(domains, collector), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
