package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/crazyuploader/vmstats/internal/stats"
)

type tickMsg time.Time

type keyMap struct {
	NextVM key.Binding
	PrevVM key.Binding
	Quit   key.Binding
	Help   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PrevVM, k.NextVM, k.Quit, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PrevVM, k.NextVM},
		{k.Quit, k.Help},
	}
}

var keys = keyMap{
	NextVM: key.NewBinding(
		key.WithKeys("tab", "n", "l", "right"),
		key.WithHelp("tab/n/→", "next VM"),
	),
	PrevVM: key.NewBinding(
		key.WithKeys("shift+tab", "p", "h", "left"),
		key.WithHelp("shift+tab/p/←", "prev VM"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

type Model struct {
	collector   stats.StatsCollector
	allStats    []stats.VMStats
	domains     []string
	currentVM   int
	err         error
	quitting    bool
	lastUpdate  time.Time
	keys        keyMap
	help        help.Model
	showHelp    bool
	initialized bool
}

func InitialModel(domains []string, collector stats.StatsCollector) Model {
	return Model{
		domains:   domains,
		collector: collector,
		keys:      keys,
		help:      help.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		fetchStats(m.collector, m.domains),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.NextVM):
			if len(m.allStats) > 0 {
				m.currentVM = (m.currentVM + 1) % len(m.allStats)
			}
		case key.Matches(msg, m.keys.PrevVM):
			if len(m.allStats) > 0 {
				m.currentVM = (m.currentVM - 1 + len(m.allStats)) % len(m.allStats)
			}
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
		case msg.String() == "r":
			// Manual refresh
			return m, fetchStats(m.collector, m.domains)
		}

	case tickMsg:
		return m, tea.Batch(
			tickCmd(),
			fetchStats(m.collector, m.domains),
		)

	case []stats.VMStats:
		m.allStats = msg
		m.lastUpdate = time.Now()
		m.err = nil
		m.initialized = true

		// Reset index if out of bounds (e.g. if VM list shrank)
		if m.currentVM >= len(m.allStats) {
			m.currentVM = 0
		}
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

func fetchStats(collector stats.StatsCollector, domains []string) tea.Cmd {
	return func() tea.Msg {
		vmStats, err := collector.GetVMStats(domains)
		if err != nil {
			return err
		}
		return vmStats
	}
}
