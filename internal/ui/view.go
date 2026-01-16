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
	// Check for minimum window size
	// Minimum of 30 lines to properly display all content
	if m.width > 0 && m.height > 0 && (m.width < 80 || m.height < 30) {
		return renderTooSmall(m.width, m.height)
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("⚠️  Error: %v\n", m.err)) +
			mutedStyle.Render("\nPress 'r' to retry, 'q' to quit\n")
	}

	if !m.initialized || len(m.allStats) == 0 {
		return mutedStyle.Render("⏳ Loading VM statistics...\n")
	}

	currentStats := &m.allStats[m.currentVM]
	stateInfo := GetVMStateInfo(currentStats.State)

	// Determine layout mode
	compactMode := m.height < 45

	// Reserved lines: Help (1) + Footer (1) + padding (2) = 4
	reservedLines := 4
	contentHeight := m.height - reservedLines
	if contentHeight < 15 {
		contentHeight = 15
	}

	// Calculate content width
	sidebarWidth := 34
	contentWidth := m.width - sidebarWidth - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	// Main layout: sidebar + content
	sidebar := renderVMList(m, contentHeight)
	content := renderMainContent(m, currentStats, stateInfo, contentWidth, compactMode)

	// Combine sidebar and content horizontally, then constrain height
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, "  ", content)

	// Apply MaxHeight to clip content rather than overflow
	mainViewStyled := lipgloss.NewStyle().MaxHeight(contentHeight).Render(mainView)

	// Help and Footer
	var footer strings.Builder
	helpView := m.help.ShortHelpView(m.keys.ShortHelp())
	if m.showHelp {
		helpView = m.help.FullHelpView(m.keys.FullHelp())
	}
	footer.WriteString(mutedStyle.Render(helpView) + "\n")
	footer.WriteString(mutedStyle.Render(fmt.Sprintf("Last updated: %s", m.lastUpdate.Format("15:04:05"))))

	return mainViewStyled + "\n" + footer.String()
}

func renderTooSmall(w, h int) string {
	style := lipgloss.NewStyle().
		Width(w).
		Height(h).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(ColorWarning)

	return style.Render(fmt.Sprintf("Terminal too small!\nNeed at least 80x30\nCurrent: %dx%d", w, h))
}

func renderVMList(m Model, height int) string {
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
	return vmListStyle.Height(height).Render(sb.String() + content)
}

func renderMainContent(m Model, currentStats *stats.VMStats, stateInfo VMStateInfo, width int, compact bool) string {
	var sb strings.Builder

	spacing := "\n\n"
	if compact {
		spacing = "\n"
	}

	// Title with state
	titleRaw := fmt.Sprintf(" %s %s • %s ", stateInfo.Icon, stateInfo.Text, currentStats.DomainName)
	title := titleStyle.Width(width).Render(titleRaw)
	sb.WriteString(title + spacing)

	// If VM is shutoff, show message instead of metrics
	if currentStats.State == VMStateShutoff {
		msg := offlineMessageStyle.Width(width).Render("💤 This VM is currently shut off.\n   Metrics will appear when the VM is running.")
		sb.WriteString(msg)
		return sb.String()
	}

	// Calculate inner width for boxes
	// Box padding (2) + Border (2) = 4 overhead
	innerWidth := width - 4

	// Memory section
	sb.WriteString(renderMemory(currentStats, width, innerWidth, compact))
	sb.WriteString(spacing)

	// CPU section
	sb.WriteString(renderCPU(currentStats, width, innerWidth, compact))
	sb.WriteString(spacing)

	// Disk section
	sb.WriteString(renderDisk(currentStats, width, innerWidth, compact))
	sb.WriteString(spacing)

	// Network section
	sb.WriteString(renderNetwork(currentStats, width, innerWidth, compact))

	return sb.String()
}

func renderMemory(vmStats *stats.VMStats, width, innerWidth int, compact bool) string {
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

	// Dynamic bar width
	barWidth := innerWidth - 60 // Roughly space for text
	if barWidth < 10 {
		barWidth = 10
	}

	memInfo := fmt.Sprintf(
		"Total: %s │ Used: %s │ Free: %s │ RSS: %s\n"+
			"Usage: %s %.1f%%",
		formatBytes(totalBytes),
		formatBytes(usedBytes),
		formatBytes(freeBytes),
		formatBytes(rssBytes),
		renderColorBar(usagePercent, barWidth),
		usagePercent,
	)

	style := boxStyle.Width(width)
	if compact {
		style = style.Padding(0, 1)
	}

	sb.WriteString(style.Render(memInfo))
	return sb.String()
}

func renderCPU(vmStats *stats.VMStats, width, innerWidth int, compact bool) string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("🖥️  CPU") + "\n")

	style := boxStyle.Width(width)
	if compact {
		style = style.Padding(0, 1)
	}

	if len(vmStats.VCPUStats) == 0 {
		sb.WriteString(style.Render(mutedStyle.Render("No vCPU data available")))
		return sb.String()
	}

	cpuInfo := fmt.Sprintf("vCPUs: %d\n\n", len(vmStats.VCPUStats))

	// Adjust column spacing based on width if needed, for now keep fixed
	cpuInfo += fmt.Sprintf("%-5s %-9s %-12s %-10s %-10s\n",
		"ID", "State", "Time", "Exits", "I/O Exits")
	cpuInfo += mutedStyle.Render(strings.Repeat("─", innerWidth)) + "\n"

	maxDisplay := 6
	if compact {
		maxDisplay = 3
	}
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

	sb.WriteString(style.Render(cpuInfo))
	return sb.String()
}

func renderDisk(vmStats *stats.VMStats, width, innerWidth int, compact bool) string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("💿 Disk") + "\n")

	style := boxStyle.Width(width)
	if compact {
		style = style.Padding(0, 1)
	}

	if len(vmStats.BlockStats) == 0 {
		sb.WriteString(style.Render(mutedStyle.Render("No disk data available")))
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

		barWidth := innerWidth - 60
		if barWidth < 10 {
			barWidth = 10
		}

		diskInfo += fmt.Sprintf(
			"📀 %s\n"+
				"   Size: %s / %s %s %.1f%%\n"+
				"   I/O:  ⬇ %s (%d ops) │ ⬆ %s (%d ops)\n",
			disk.Name,
			formatBytes(disk.Allocation),
			formatBytes(disk.Capacity),
			renderColorBar(usagePercent, barWidth),
			usagePercent,
			formatBytes(disk.ReadBytes),
			disk.ReadReqs,
			formatBytes(disk.WriteBytes),
			disk.WriteReqs,
		)
	}

	sb.WriteString(style.Render(diskInfo))
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

func renderNetwork(vmStats *stats.VMStats, width, innerWidth int, compact bool) string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("🌐 Network") + "\n")

	style := boxStyle.Width(width)
	if compact {
		style = style.Padding(0, 1)
	}

	if len(vmStats.InterfaceStats) == 0 {
		sb.WriteString(style.Render(mutedStyle.Render("No network data available")))
		return sb.String()
	}

	var netInfo string
	for _, net := range vmStats.InterfaceStats {
		if net.Name == "" {
			continue
		}

		ipStr := ""
		if len(net.IPs) > 0 {
			ipStr = fmt.Sprintf("   📍 IPs: %s\n", strings.Join(net.IPs, ", "))
		}

		netInfo += fmt.Sprintf(
			"📡 %s\n"+
				"%s"+
				"   ⬇ Rx: %s (%d pkts) │ ❌ %d errs\n"+
				"   ⬆ Tx: %s (%d pkts) │ ❌ %d errs\n",
			net.Name,
			ipStr,
			formatBytes(net.RxBytes),
			net.RxPackets,
			net.RxErrs+net.RxDrop,
			formatBytes(net.TxBytes),
			net.TxPackets,
			net.TxErrs+net.TxDrop,
		)
	}

	sb.WriteString(style.Render(netInfo))
	return sb.String()
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
