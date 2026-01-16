package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

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
