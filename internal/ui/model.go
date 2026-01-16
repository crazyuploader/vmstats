package ui

import (
	"time"

	"github.com/crazyuploader/vmstats/internal/stats"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

type Model struct {
	collector  stats.StatsCollector
	stats      *stats.VMStats
	domain     string
	err        error
	quitting   bool
	lastUpdate time.Time
}

func InitialModel(domain string, collector stats.StatsCollector) Model {
	return Model{
		domain:    domain,
		collector: collector,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		fetchStats(m.collector, m.domain),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "r":
			// Manual refresh
			return m, fetchStats(m.collector, m.domain)
		}

	case tickMsg:
		return m, tea.Batch(
			tickCmd(),
			fetchStats(m.collector, m.domain),
		)

	case *stats.VMStats:
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

func (m Model) View() string {
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

func fetchStats(collector stats.StatsCollector, domain string) tea.Cmd {
	return func() tea.Msg {
		vmStats, err := collector.GetVMStats(domain)
		if err != nil {
			return err
		}
		return vmStats
	}
}
