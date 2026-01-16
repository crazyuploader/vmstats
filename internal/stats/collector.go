package stats

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// StatsCollector defines an interface for collecting VM stats
type StatsCollector interface {
	GetVMStats(domains []string) ([]VMStats, error)
}

// VirshCollector collects stats using the virsh command
type VirshCollector struct{}

// NewVirshCollector creates a new VirshCollector
func NewVirshCollector() *VirshCollector {
	return &VirshCollector{}
}

// GetVMStats parses virsh domstats output
func (c *VirshCollector) GetVMStats(domains []string) ([]VMStats, error) {
	args := []string{"domstats", "--vcpu", "--balloon", "--block", "--interface", "--state"}
	args = append(args, domains...)
	cmd := exec.Command("virsh", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute virsh: %w", err)
	}

	return parseVirshOutput(string(output))
}

func parseVirshOutput(output string) ([]VMStats, error) {
	var allStats []VMStats
	lines := strings.Split(output, "\n")

	var currentStats *VMStats
	currentVCPU := -1
	currentBlock := -1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse domain name - this indicates start of a new VM
		if strings.HasPrefix(line, "Domain:") {
			if currentStats != nil {
				allStats = append(allStats, *currentStats)
			}
			currentStats = &VMStats{}
			currentStats.DomainName = strings.Trim(strings.TrimPrefix(line, "Domain:"), " '")
			currentVCPU = -1
			currentBlock = -1
			continue
		}

		if currentStats == nil {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Parse state
		if strings.HasPrefix(key, "state.") {
			parseState(key, value, currentStats)
		}

		// Parse balloon stats
		if strings.HasPrefix(key, "balloon.") {
			parseBaloonStat(key, value, currentStats)
		}

		// Parse VCPU stats
		if strings.HasPrefix(key, "vcpu.") {
			parseVCPUStat(key, value, currentStats, &currentVCPU)
		}

		// Parse block stats
		if strings.HasPrefix(key, "block.") {
			parseBlockStat(key, value, currentStats, &currentBlock)
		}

		// Parse interface stats
		if strings.HasPrefix(key, "net.") {
			parseInterfaceStat(key, value, currentStats)
		}
	}

	// Append the last one
	if currentStats != nil {
		allStats = append(allStats, *currentStats)
	}

	// Fetch IP addresses for running VMs
	enrichWithIPs(allStats)

	return allStats, nil
}

func parseBaloonStat(key, value string, stats *VMStats) {
	val, _ := strconv.ParseInt(value, 10, 64)
	switch key {
	case "balloon.current":
		stats.BalloonStats.Current = val
	case "balloon.maximum":
		stats.BalloonStats.Maximum = val
	case "balloon.unused":
		stats.BalloonStats.Unused = val
	case "balloon.available":
		stats.BalloonStats.Available = val
	case "balloon.usable":
		stats.BalloonStats.Usable = val
	case "balloon.rss":
		stats.BalloonStats.RSS = val
	}
}

func parseVCPUStat(key, value string, stats *VMStats, currentVCPU *int) {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return
	}

	if parts[1] == "current" || parts[1] == "maximum" {
		return
	}

	vcpuID, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}

	// Ensure we have enough VCPU entries
	for len(stats.VCPUStats) <= vcpuID {
		stats.VCPUStats = append(stats.VCPUStats, VCPUStats{ID: len(stats.VCPUStats)})
	}

	val, _ := strconv.ParseInt(value, 10, 64)

	if len(parts) >= 3 {
		metric := parts[2]
		switch metric {
		case "state":
			stats.VCPUStats[vcpuID].State = int(val)
		case "time":
			stats.VCPUStats[vcpuID].Time = val
		}
	}

	// Parse sum metrics
	if len(parts) >= 4 && parts[3] == "sum" {
		metric := parts[2]
		switch metric {
		case "exits":
			stats.VCPUStats[vcpuID].Exits = val
		case "halt_exits":
			stats.VCPUStats[vcpuID].HaltExits = val
		case "irq_exits":
			stats.VCPUStats[vcpuID].IRQExits = val
		case "io_exits":
			stats.VCPUStats[vcpuID].IOExits = val
		}
	}
}

func parseBlockStat(key, value string, stats *VMStats, currentBlock *int) {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return
	}

	if parts[1] == "count" {
		return
	}

	blockID, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}

	// Ensure we have enough block entries
	for len(stats.BlockStats) <= blockID {
		stats.BlockStats = append(stats.BlockStats, BlockStats{})
	}

	if len(parts) >= 3 {
		metric := parts[2]
		switch metric {
		case "name":
			stats.BlockStats[blockID].Name = value
		case "path":
			stats.BlockStats[blockID].Path = value
		case "allocation":
			val, _ := strconv.ParseInt(value, 10, 64)
			stats.BlockStats[blockID].Allocation = val
		case "capacity":
			val, _ := strconv.ParseInt(value, 10, 64)
			stats.BlockStats[blockID].Capacity = val
		case "physical":
			val, _ := strconv.ParseInt(value, 10, 64)
			stats.BlockStats[blockID].Physical = val
		}
	}

	// Parse rd/wr stats
	if len(parts) >= 4 {
		operation := parts[2] // rd or wr
		metric := parts[3]    // reqs or bytes
		val, _ := strconv.ParseInt(value, 10, 64)

		switch operation {
		case "rd":
			switch metric {
			case "reqs":
				stats.BlockStats[blockID].ReadReqs = val
			case "bytes":
				stats.BlockStats[blockID].ReadBytes = val
			}
		case "wr":
			switch metric {
			case "reqs":
				stats.BlockStats[blockID].WriteReqs = val
			case "bytes":
				stats.BlockStats[blockID].WriteBytes = val
			}
		}
	}
}

func parseInterfaceStat(key, value string, stats *VMStats) {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return
	}

	if parts[1] == "count" {
		return
	}

	ifID, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}

	// Ensure we have enough interface entries
	for len(stats.InterfaceStats) <= ifID {
		stats.InterfaceStats = append(stats.InterfaceStats, InterfaceStats{})
	}

	if len(parts) >= 3 {
		metric := parts[2]

		switch metric {
		case "name":
			stats.InterfaceStats[ifID].Name = value
		case "rx":
			if len(parts) >= 4 {
				val, _ := strconv.ParseInt(value, 10, 64)
				switch parts[3] {
				case "bytes":
					stats.InterfaceStats[ifID].RxBytes = val
				case "pkts":
					stats.InterfaceStats[ifID].RxPackets = val
				case "errs":
					stats.InterfaceStats[ifID].RxErrs = val
				case "drop":
					stats.InterfaceStats[ifID].RxDrop = val
				}
			}
		case "tx":
			if len(parts) >= 4 {
				val, _ := strconv.ParseInt(value, 10, 64)
				switch parts[3] {
				case "bytes":
					stats.InterfaceStats[ifID].TxBytes = val
				case "pkts":
					stats.InterfaceStats[ifID].TxPackets = val
				case "errs":
					stats.InterfaceStats[ifID].TxErrs = val
				case "drop":
					stats.InterfaceStats[ifID].TxDrop = val
				}
			}
		}
	}
}
func parseState(key, value string, stats *VMStats) {
	val, _ := strconv.Atoi(value)
	switch key {
	case "state.state":
		stats.State = val
	}
}

func enrichWithIPs(vms []VMStats) {
	for i := range vms {
		// Only check IPs for running VMs (State == 1)
		if vms[i].State != 1 {
			continue
		}

		cmd := exec.Command("virsh", "domifaddr", vms[i].DomainName, "--full", "--source", "lease")
		output, err := cmd.Output()
		if err != nil {
			// Try without source arg if lease fails, or maybe agent
			// For now, just ignore errors as IPs are "nice to have"
			continue
		}

		parseDomIfAddr(string(output), &vms[i])
	}
}

func parseDomIfAddr(output string, vm *VMStats) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Name") || strings.HasPrefix(line, "-") {
			continue
		}

		// Expected format: interface MAC protocol address
		// vnet0 52:54:00:12:34:56 ipv4 192.168.122.238/24
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			ifName := fields[0]
			// protocol := fields[2]
			address := fields[3]

			// Strip CIDR from address if present
			if idx := strings.Index(address, "/"); idx != -1 {
				address = address[:idx]
			}

			// Find matching interface in stats and append IP
			for j := range vm.InterfaceStats {
				if vm.InterfaceStats[j].Name == ifName {
					vm.InterfaceStats[j].IPs = append(vm.InterfaceStats[j].IPs, address)
					break
				}
			}
		}
	}
}
