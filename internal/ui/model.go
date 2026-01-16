package ui

import (
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/crazyuploader/vmstats/internal/stats"
)

type tickMsg time.Time

type keyMap struct {
	NextVM      key.Binding
	PrevVM      key.Binding
	Refresh     key.Binding
	TogglePause key.Binding
	Quit        key.Binding
	Help        key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PrevVM, k.NextVM, k.Refresh, k.TogglePause, k.Quit, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PrevVM, k.NextVM},
		{k.Refresh, k.TogglePause, k.Quit, k.Help},
	}
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
	refreshRate time.Duration
	width       int
	height      int
	paused      bool
}

func InitialModel(domains []string, collector stats.StatsCollector, refreshRate time.Duration) Model {
	return Model{
		domains:     domains,
		collector:   collector,
		keys:        keys,
		help:        help.New(),
		refreshRate: refreshRate,
	}
}

var keys = keyMap{
	NextVM: key.NewBinding(
		key.WithKeys("down", "j", "tab"),
		key.WithHelp("↓/j", "next"),
	),
	PrevVM: key.NewBinding(
		key.WithKeys("up", "k", "shift+tab"),
		key.WithHelp("↑/k", "prev"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	TogglePause: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pause/resume"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
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
		case key.Matches(msg, m.keys.Refresh):
			return m, fetchStats(m.collector, m.domains)
		case key.Matches(msg, m.keys.TogglePause):
			m.paused = !m.paused
			return m, nil
		}

	case tickMsg:
		var cmd tea.Cmd
		if !m.paused {
			cmd = fetchStats(m.collector, m.domains)
		}
		return m, tea.Batch(
			m.tickCmd(),
			cmd,
		)

	case []stats.VMStats:
		// Calculate CPU usage if we have previous stats
		if len(m.allStats) > 0 {
			calculateCPUUsage(msg, m.allStats)
		}

		sortVMs(msg)
		m.allStats = msg
		m.lastUpdate = time.Now()
		m.err = nil
		m.initialized = true

		// Reset index if out of bounds (e.g. if VM list shrank)
		if m.currentVM >= len(m.allStats) {
			m.currentVM = 0
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	return renderView(m)
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
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

func sortVMs(vms []stats.VMStats) {
	sort.Slice(vms, func(i, j int) bool {
		// Define priority: Running/Idle/Paused are "active" (priority 0)
		// Others like Shutoff are "inactive" (priority 1)
		p1 := getVMPriority(vms[i].State)
		p2 := getVMPriority(vms[j].State)

		if p1 != p2 {
			return p1 < p2
		}

		// Secondary sort by name
		return vms[i].DomainName < vms[j].DomainName
	})
}

func getVMPriority(state int) int {
	switch state {
	case VMStateRunning, VMStateIdle, VMStatePaused:
		return 0
	default:
		return 1
	}
}

func calculateCPUUsage(newStats, oldStats []stats.VMStats) {
	// Create map for fast lookup of old stats
	oldMap := make(map[string]stats.VMStats)
	for _, vm := range oldStats {
		oldMap[vm.DomainName] = vm
	}

	for i := range newStats {
		vm := &newStats[i]
		oldVM, ok := oldMap[vm.DomainName]
		if !ok {
			continue
		}

		// Time passed between updates in nanoseconds
		deltaFuncTime := vm.LastUpdate - oldVM.LastUpdate
		if deltaFuncTime <= 0 {
			continue
		}

		for j := range vm.VCPUStats {
			if j >= len(oldVM.VCPUStats) {
				break
			}

			// CPU time is in nanoseconds
			deltaCPUTime := vm.VCPUStats[j].Time - oldVM.VCPUStats[j].Time
			if deltaCPUTime < 0 {
				continue
			}

			// Usage % = (delta CPU time / delta Wall time) * 100
			usage := (float64(deltaCPUTime) / float64(deltaFuncTime)) * 100.0

			// Cap at 100% per core
			if usage > 100.0 {
				usage = 100.0
			}

			vm.VCPUStats[j].Usage = usage
		}
	}
}
