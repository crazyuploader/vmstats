package ui

import "github.com/charmbracelet/lipgloss"

// VM States from libvirt
const (
	VMStateNoState     = 0
	VMStateRunning     = 1
	VMStateIdle        = 2
	VMStatePaused      = 3
	VMStateShutdown    = 4
	VMStateShutoff     = 5
	VMStateCrashed     = 6
	VMStatePMSuspended = 7
)

// Usage thresholds for color coding
const (
	ThresholdLow  = 70.0 // Green below this
	ThresholdHigh = 90.0 // Yellow below this, Red above
)

// Default refresh rate in seconds
const DefaultRefreshRate = 2

// Colors - Modern, cohesive palette
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#06B6D4") // Cyan

	// Status colors
	ColorSuccess = lipgloss.Color("#10B981") // Green
	ColorWarning = lipgloss.Color("#F59E0B") // Amber
	ColorDanger  = lipgloss.Color("#EF4444") // Red
	ColorInfo    = lipgloss.Color("#3B82F6") // Blue

	// Neutral colors
	ColorText       = lipgloss.Color("#F3F4F6") // Light gray
	ColorTextMuted  = lipgloss.Color("#9CA3AF") // Muted gray
	ColorBorder     = lipgloss.Color("#4B5563") // Dark gray
	ColorBackground = lipgloss.Color("#1F2937") // Dark background
)

// VMStateInfo holds display information for a VM state
type VMStateInfo struct {
	Icon  string
	Text  string
	Color lipgloss.Color
}

// VMStates maps state codes to display info
var VMStates = map[int]VMStateInfo{
	VMStateNoState:     {Icon: "❓", Text: "Unknown", Color: ColorTextMuted},
	VMStateRunning:     {Icon: "🟢", Text: "Running", Color: ColorSuccess},
	VMStateIdle:        {Icon: "🌙", Text: "Idle", Color: ColorInfo},
	VMStatePaused:      {Icon: "⏸️", Text: "Paused", Color: ColorWarning},
	VMStateShutdown:    {Icon: "🔻", Text: "Shutdown", Color: ColorWarning},
	VMStateShutoff:     {Icon: "🔴", Text: "Shutoff", Color: ColorDanger},
	VMStateCrashed:     {Icon: "💥", Text: "Crashed", Color: ColorDanger},
	VMStatePMSuspended: {Icon: "💤", Text: "Suspended", Color: ColorInfo},
}

// GetVMStateInfo returns display info for a given state
func GetVMStateInfo(state int) VMStateInfo {
	if info, ok := VMStates[state]; ok {
		return info
	}
	return VMStates[VMStateNoState]
}
