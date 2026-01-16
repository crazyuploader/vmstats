package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

type model struct {
	stats      *VMStats
	domain     string
	err        error
	quitting   bool
	lastUpdate time.Time
}

func initialModel(domain string) model {
	return model{
		domain: domain,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		fetchStats(m.domain),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "r":
			// Manual refresh
			return m, fetchStats(m.domain)
		}

	case tickMsg:
		return m, tea.Batch(
			tickCmd(),
			fetchStats(m.domain),
		)

	case *VMStats:
		m.stats = msg
		m.lastUpdate = time.Now()
		m.err = nil
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	return renderView(m)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchStats(domain string) tea.Cmd {
	return func() tea.Msg {
		stats, err := getVMStats(domain)
		if err != nil {
			return err
		}
		return stats
	}
}
