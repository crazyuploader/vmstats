package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/crazyuploader/vmstats/internal/stats"
)

func renderMainContent(currentStats *stats.VMStats, stateInfo VMStateInfo, width int, compact bool) string {
	var sb strings.Builder

	spacing := "\n\n"
	if compact {
		spacing = "\n"
	}

	// Title with state
	osType := ""
	if currentStats.OSType != "" {
		osType = fmt.Sprintf("• %s ", currentStats.OSType)
	}
	titleRaw := fmt.Sprintf(" %s %s %s• %s ", stateInfo.Icon, stateInfo.Text, osType, currentStats.DomainName)
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
	sb.WriteString(renderNetwork(currentStats, width, compact))

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

	// Adjust column spacing based on width
	// In compact mode, we hide "I/O Exits" to save width and potential wraps
	if compact {
		cpuInfo += fmt.Sprintf("%-5s %-9s %-8s %-12s %-10s\n",
			"ID", "State", "Usage", "Time", "Exits")
	} else {
		cpuInfo += fmt.Sprintf("%-5s %-9s %-8s %-12s %-10s %-10s\n",
			"ID", "State", "Usage", "Time", "Exits", "I/O Exits")
	}
	cpuInfo += mutedStyle.Render(strings.Repeat("─", innerWidth)) + "\n"

	maxDisplay := 6
	if compact {
		maxDisplay = 4 // Show a bit more in compact but fewer columns
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

		// Colorize usage
		usageStr := fmt.Sprintf("%.1f%%", vcpu.Usage)
		switch {
		case vcpu.Usage >= 90:
			usageStr = errorStyle.Render(usageStr)
		case vcpu.Usage >= 50:
			usageStr = lipgloss.NewStyle().Foreground(ColorWarning).Render(usageStr)
		}

		if compact {
			cpuInfo += fmt.Sprintf("%-5d %-9s %-8s %-12s %-10d\n",
				vcpu.ID,
				stateStr,
				usageStr,
				formatDuration(vcpu.Time),
				vcpu.Exits,
			)
		} else {
			cpuInfo += fmt.Sprintf("%-5d %-9s %-8s %-12s %-10d %-10d\n",
				vcpu.ID,
				stateStr,
				usageStr,
				formatDuration(vcpu.Time),
				vcpu.Exits,
				vcpu.IOExits,
			)
		}
	}

	sb.WriteString(style.Render(cpuInfo))
	return sb.String()
}

func renderDisk(vmStats *stats.VMStats, width, innerWidth int, compact bool) string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("💿 Virtual Disks (Host)") + "\n")

	style := boxStyle.Width(width)
	if compact {
		style = style.Padding(0, 1)
	}

	if len(vmStats.BlockStats) == 0 {
		sb.WriteString(style.Render(mutedStyle.Render("No disk data available")))
		return sb.String()
	}

	var diskInfo string
	// Limit number of disks displayed to prevent overflow
	maxDisks := 4
	if compact {
		maxDisks = 2
	}

	for i, disk := range vmStats.BlockStats {
		if i >= maxDisks {
			diskInfo += mutedStyle.Render(fmt.Sprintf("... and %d more disks\n", len(vmStats.BlockStats)-maxDisks))
			break
		}

		if disk.Name == "" {
			continue
		}

		// Add separator if not the first item
		if i > 0 {
			diskInfo += "\n"
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
				"   Phys: %s / Max: %s %s %.1f%%\n"+
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

func renderNetwork(vmStats *stats.VMStats, width int, compact bool) string {
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
