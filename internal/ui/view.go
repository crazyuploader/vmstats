package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/crazyuploader/vmstats/internal/stats"
)

// Styles using our color palette
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText).
			Background(ColorPrimary).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary)

	normalStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	mutedStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	errorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	vmListStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1).
			Width(30)

	selectedVMStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	offlineMessageStyle = lipgloss.NewStyle().
				Foreground(ColorTextMuted).
				Italic(true).
				Padding(1, 2)
)

func renderView(m Model) string {
	var sb strings.Builder

	if m.err != nil {
		sb.WriteString(errorStyle.Render(fmt.Sprintf("⚠️  Error: %v\n", m.err)))
		sb.WriteString(mutedStyle.Render("\nPress 'r' to retry, 'q' to quit\n"))
		return sb.String()
	}

	if !m.initialized || len(m.allStats) == 0 {
		sb.WriteString(mutedStyle.Render("⏳ Loading VM statistics...\n"))
		return sb.String()
	}

	currentStats := &m.allStats[m.currentVM]
	stateInfo := GetVMStateInfo(currentStats.State)

	// Main layout: sidebar + content
	sidebar := renderVMList(m)
	content := renderMainContent(m, currentStats, stateInfo)

	// Combine sidebar and content horizontally
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, "  ", content)
	sb.WriteString(mainView)
	sb.WriteString("\n\n")

	// Help and Footer
	helpView := m.help.ShortHelpView(m.keys.ShortHelp())
	if m.showHelp {
		helpView = m.help.FullHelpView(m.keys.FullHelp())
	}
	sb.WriteString(mutedStyle.Render(helpView) + "\n")

	footer := fmt.Sprintf("Last updated: %s", m.lastUpdate.Format("15:04:05"))
	sb.WriteString(mutedStyle.Render(footer))

	return sb.String()
}

func renderVMList(m Model) string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render("📋 VMs") + "\n")

	var vmItems []string
	for i, vm := range m.allStats {
		stateInfo := GetVMStateInfo(vm.State)
		marker := "  "
		style := normalStyle
		if i == m.currentVM {
			marker = "▶ "
			style = selectedVMStyle
		}
		vmItem := fmt.Sprintf("%s%s %s", marker, stateInfo.Icon, vm.DomainName)
		vmItems = append(vmItems, style.Render(vmItem))
	}

	content := strings.Join(vmItems, "\n")
	return vmListStyle.Render(sb.String() + content)
}

func renderMainContent(m Model, currentStats *stats.VMStats, stateInfo VMStateInfo) string {
	var sb strings.Builder

	// Title with state
	titleRaw := fmt.Sprintf(" %s %s • %s ", stateInfo.Icon, stateInfo.Text, currentStats.DomainName)
	title := titleStyle.Render(titleRaw)
	sb.WriteString(title + "\n\n")

	// If VM is shutoff, show message instead of metrics
	if currentStats.State == VMStateShutoff {
		msg := offlineMessageStyle.Render("💤 This VM is currently shut off.\n   Metrics will appear when the VM is running.")
		sb.WriteString(msg)
		return sb.String()
	}

	// Memory section
	sb.WriteString(renderMemory(currentStats))
	sb.WriteString("\n\n")

	// CPU section
	sb.WriteString(renderCPU(currentStats))
	sb.WriteString("\n\n")

	// Disk section
	sb.WriteString(renderDisk(currentStats))

	return sb.String()
}

func renderMemory(vmStats *stats.VMStats) string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("💾 Memory") + "\n")

	totalBytes := vmStats.BalloonStats.Current * 1024
	usedBytes := (vmStats.BalloonStats.Current - vmStats.BalloonStats.Unused) * 1024
	freeBytes := vmStats.BalloonStats.Unused * 1024
	rssBytes := vmStats.BalloonStats.RSS * 1024

	usagePercent := float64(0)
	if totalBytes > 0 {
		usagePercent = float64(usedBytes) / float64(totalBytes) * 100
	}

	memInfo := fmt.Sprintf(
		"Total: %s │ Used: %s │ Free: %s │ RSS: %s\n"+
			"Usage: %s %.1f%%",
		formatBytes(totalBytes),
		formatBytes(usedBytes),
		formatBytes(freeBytes),
		formatBytes(rssBytes),
		renderColorBar(usagePercent, 30),
		usagePercent,
	)

	sb.WriteString(boxStyle.Render(memInfo))
	return sb.String()
}

func renderCPU(vmStats *stats.VMStats) string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("🖥️  CPU") + "\n")

	if len(vmStats.VCPUStats) == 0 {
		sb.WriteString(boxStyle.Render(mutedStyle.Render("No vCPU data available")))
		return sb.String()
	}

	cpuInfo := fmt.Sprintf("vCPUs: %d\n\n", len(vmStats.VCPUStats))

	cpuInfo += fmt.Sprintf("%-5s %-9s %-12s %-10s %-10s\n",
		"ID", "State", "Time", "Exits", "I/O Exits")
	cpuInfo += mutedStyle.Render(strings.Repeat("─", 50)) + "\n"

	maxDisplay := 6
	for i, vcpu := range vmStats.VCPUStats {
		if i >= maxDisplay {
			cpuInfo += mutedStyle.Render(fmt.Sprintf("... and %d more vCPUs\n", len(vmStats.VCPUStats)-maxDisplay))
			break
		}
		stateStr := lipgloss.NewStyle().Foreground(ColorSuccess).Render("running")
		if vcpu.State == 0 {
			stateStr = mutedStyle.Render("offline")
		}
		cpuInfo += fmt.Sprintf("%-5d %-9s %-12s %-10d %-10d\n",
			vcpu.ID,
			stateStr,
			formatDuration(vcpu.Time),
			vcpu.Exits,
			vcpu.IOExits,
		)
	}

	sb.WriteString(boxStyle.Render(cpuInfo))
	return sb.String()
}

func renderDisk(vmStats *stats.VMStats) string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("💿 Disk") + "\n")

	if len(vmStats.BlockStats) == 0 {
		sb.WriteString(boxStyle.Render(mutedStyle.Render("No disk data available")))
		return sb.String()
	}

	var diskInfo string
	for _, disk := range vmStats.BlockStats {
		if disk.Name == "" {
			continue
		}

		usagePercent := float64(0)
		if disk.Capacity > 0 {
			usagePercent = (float64(disk.Allocation) / float64(disk.Capacity)) * 100
		}

		diskInfo += fmt.Sprintf(
			"📀 %s\n"+
				"   Size: %s / %s %s %.1f%%\n"+
				"   I/O:  ⬇ %s (%d ops) │ ⬆ %s (%d ops)\n",
			disk.Name,
			formatBytes(disk.Allocation),
			formatBytes(disk.Capacity),
			renderColorBar(usagePercent, 20),
			usagePercent,
			formatBytes(disk.ReadBytes),
			disk.ReadReqs,
			formatBytes(disk.WriteBytes),
			disk.WriteReqs,
		)
	}

	sb.WriteString(boxStyle.Render(diskInfo))
	return sb.String()
}

// renderColorBar creates a progress bar with color based on thresholds
func renderColorBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	// Determine color based on thresholds
	var color lipgloss.Color
	switch {
	case percent >= ThresholdHigh:
		color = ColorDanger
	case percent >= ThresholdLow:
		color = ColorWarning
	default:
		color = ColorSuccess
	}

	filledStyle := lipgloss.NewStyle().Foreground(color)
	emptyStyle := lipgloss.NewStyle().Foreground(ColorBorder)

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += filledStyle.Render("█")
		} else {
			bar += emptyStyle.Render("░")
		}
	}

	return "[" + bar + "]"
}

// formatBytes converts bytes to human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration converts nanoseconds to human-readable duration
func formatDuration(ns int64) string {
	seconds := ns / 1_000_000_000
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm%ds", minutes, seconds%60)
	}
	hours := minutes / 60
	return fmt.Sprintf("%dh%dm", hours, minutes%60)
}
