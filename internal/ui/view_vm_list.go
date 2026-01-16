package ui

import (
	"fmt"
	"strings"
)

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

	// Resource Summary
	running := 0
	totalCPUs := 0
	totalMem := int64(0)

	for _, vm := range m.allStats {
		if vm.State == VMStateRunning {
			running++
		}
		totalCPUs += len(vm.VCPUStats)
		totalMem += vm.BalloonStats.Current * 1024
	}

	summary := fmt.Sprintf("\n%s\nRunning: %d/%d\nCPUs: %d | Mem: %s",
		mutedStyle.Render(strings.Repeat("─", 20)),
		running, len(m.allStats),
		totalCPUs, formatBytes(totalMem),
	)

	finalContent := sb.String() + content + summary
	return vmListStyle.Height(height).Render(finalContent)
}
