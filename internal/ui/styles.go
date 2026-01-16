package ui

import "github.com/charmbracelet/lipgloss"

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
