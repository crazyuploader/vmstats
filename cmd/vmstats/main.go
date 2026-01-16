package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/crazyuploader/vmstats/internal/stats"
	"github.com/crazyuploader/vmstats/internal/ui"
	"github.com/crazyuploader/vmstats/internal/version"
)

func main() {
	// Parse flags
	domainsFlag := flag.String("domains", "", "Comma-separated list of libvirt domains to monitor (empty for all)")
	logFile := flag.String("log", "", "Log file path (optional)")
	refreshInterval := flag.String("interval", "2s", "Refresh interval (e.g., 500ms, 1s, 2s)")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Println(version.GetFullVersionString())
		return
	}

	// Parse refresh interval
	duration, err := time.ParseDuration(*refreshInterval)
	if err != nil {
		fmt.Printf("Invalid interval format: %v\n", err)
		os.Exit(1)
	}
	if duration < 500*time.Millisecond {
		duration = 500 * time.Millisecond
		fmt.Println("Warning: Interval too low, setting to 500ms")
	}

	// Parse domains
	var domains []string
	if *domainsFlag != "" {
		domains = strings.Split(*domainsFlag, ",")
		for i := range domains {
			domains[i] = strings.TrimSpace(domains[i])
		}
	}

	// Setup logging
	if *logFile != "" {
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
	} else {
		log.SetOutput(io.Discard)
	}

	if len(domains) > 0 {
		log.Printf("Starting vmstats for domains: %v (refresh: %s)", domains, duration)
	} else {
		log.Printf("Starting vmstats for ALL domains (refresh: %s)", duration)
	}

	// Initialize collector
	collector := stats.NewVirshCollector()

	// Initialize Bubble Tea program
	model := ui.InitialModel(domains, collector, duration)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
