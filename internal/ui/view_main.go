package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	content := renderMainContent(currentStats, stateInfo, contentWidth, compactMode)

	// Combine sidebar and content horizontally, then constrain height
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, "  ", content)

	// Apply MaxHeight to clip content rather than overflow
	mainViewStyled := lipgloss.NewStyle().MaxHeight(contentHeight).Render(mainView)

	// Help and Footer
	var footer strings.Builder
	helpView := m.help.ShortHelpView(m.keys.ShortHelp())
	if m.showHelp {
		helpView = m.help.FullHelpView(m.keys.FullHelp())

		// Add Legend
		legend := fmt.Sprintf("\n\n%s\n"+
			"• Colors: %s <50%%, %s 50-90%%, %s >90%%\n"+
			"• Phys: Physical disk space used on host\n"+
			"• Max: Maximum virtual disk size\n"+
			"• RSS: Resident Set Size (RAM used)",
			headerStyle.Render("Legend"),
			lipgloss.NewStyle().Foreground(ColorSuccess).Render("Green"),
			lipgloss.NewStyle().Foreground(ColorWarning).Render("Yellow"),
			lipgloss.NewStyle().Foreground(ColorDanger).Render("Red"),
		)
		helpView += mutedStyle.Render(legend)
	}
	footer.WriteString(mutedStyle.Render(helpView) + "\n")

	lastUpdated := fmt.Sprintf("Last updated: %s", m.lastUpdate.Format("15:04:05"))
	if m.paused {
		lastUpdated += " " + errorStyle.Render("[PAUSED]")
	}
	footer.WriteString(mutedStyle.Render(lastUpdated))

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
