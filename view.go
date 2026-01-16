package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("220"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)
)

func renderView(m model) string {
	var sb strings.Builder

	// Title
	title := titleStyle.Render(fmt.Sprintf(" VM Stats Monitor - %s ", m.domain))
	sb.WriteString(title + "\n\n")

	if m.err != nil {
		sb.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v\n", m.err)))
		sb.WriteString("\nPress 'q' to quit\n")
		return sb.String()
	}

	if m.stats == nil {
		sb.WriteString("Loading...\n")
		return sb.String()
	}

	// Memory section
	sb.WriteString(renderMemory(m.stats))
	sb.WriteString("\n\n")

	// CPU section
	sb.WriteString(renderCPU(m.stats))
	sb.WriteString("\n\n")

	// Disk section
	sb.WriteString(renderDisk(m.stats))
	sb.WriteString("\n\n")

	// Footer
	footer := fmt.Sprintf("Last updated: %s | Press 'r' to refresh, 'q' to quit",
		m.lastUpdate.Format("15:04:05"))
	sb.WriteString(normalStyle.Render(footer))

	return sb.String()
}

func renderMemory(stats *VMStats) string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("💾 Memory") + "\n")

	usedMB := (stats.BalloonStats.Current - stats.BalloonStats.Unused) / 1024
	totalMB := stats.BalloonStats.Current / 1024
	usagePercent := float64(usedMB) / float64(totalMB) * 100

	memInfo := fmt.Sprintf(
		"Total: %d MB | Used: %d MB | Free: %d MB | RSS: %d MB\n"+
			"Usage: %.1f%% %s",
		totalMB,
		usedMB,
		stats.BalloonStats.Unused/1024,
		stats.BalloonStats.RSS/1024,
		usagePercent,
		renderBar(usagePercent, 40),
	)

	sb.WriteString(boxStyle.Render(memInfo))
	return sb.String()
}

func renderCPU(stats *VMStats) string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("🖥️  CPU") + "\n")

	cpuInfo := fmt.Sprintf("vCPUs: %d\n\n", len(stats.VCPUStats))

	cpuInfo += fmt.Sprintf("%-6s %-10s %-15s %-12s %-12s\n",
		"vCPU", "State", "CPU Time (s)", "Exits", "I/O Exits")
	cpuInfo += strings.Repeat("─", 70) + "\n"

	for i, vcpu := range stats.VCPUStats {
		if i >= 6 { // Limit to first 6 vCPUs for display
			break
		}
		state := "running"
		if vcpu.State == 0 {
			state = "offline"
		}
		cpuInfo += fmt.Sprintf("%-6d %-10s %-15d %-12d %-12d\n",
			vcpu.ID,
			state,
			vcpu.Time/1000000000, // Convert to seconds
			vcpu.Exits,
			vcpu.IOExits,
		)
	}

	sb.WriteString(boxStyle.Render(cpuInfo))
	return sb.String()
}

func renderDisk(stats *VMStats) string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("💿 Disk") + "\n")

	if len(stats.BlockStats) == 0 {
		sb.WriteString(boxStyle.Render("No disk stats available"))
		return sb.String()
	}

	var diskInfo string
	for _, disk := range stats.BlockStats {
		if disk.Name == "" {
			continue
		}

		usedGB := float64(disk.Allocation) / (1024 * 1024 * 1024)
		totalGB := float64(disk.Capacity) / (1024 * 1024 * 1024)
		usagePercent := (float64(disk.Allocation) / float64(disk.Capacity)) * 100

		readMB := float64(disk.ReadBytes) / (1024 * 1024)
		writeMB := float64(disk.WriteBytes) / (1024 * 1024)

		diskInfo += fmt.Sprintf(
			"Device: %s\n"+
				"Size: %.1f GB / %.1f GB (%.1f%%) %s\n"+
				"Read: %.2f MB (%d ops) | Write: %.2f MB (%d ops)\n",
			disk.Name,
			usedGB, totalGB, usagePercent,
			renderBar(usagePercent, 30),
			readMB, disk.ReadReqs,
			writeMB, disk.WriteReqs,
		)
	}

	sb.WriteString(boxStyle.Render(diskInfo))
	return sb.String()
}

func renderBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"

	return bar
}

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
