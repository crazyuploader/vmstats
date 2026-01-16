package stats

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// StatsCollector defines an interface for collecting VM stats
type StatsCollector interface {
	GetVMStats(domain string) (*VMStats, error)
}

// VirshCollector collects stats using the virsh command
type VirshCollector struct{}

// NewVirshCollector creates a new VirshCollector
func NewVirshCollector() *VirshCollector {
	return &VirshCollector{}
}

// GetVMStats parses virsh domstats output
func (c *VirshCollector) GetVMStats(domain string) (*VMStats, error) {
	cmd := exec.Command("virsh", "domstats", "--vcpu", "--balloon", "--block", domain)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute virsh: %w", err)
	}

	return parseVirshOutput(string(output))
}

func parseVirshOutput(output string) (*VMStats, error) {
	stats := &VMStats{}
	lines := strings.Split(output, "\n")

	currentVCPU := -1
	currentBlock := -1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse domain name
		if strings.HasPrefix(line, "Domain:") {
			stats.DomainName = strings.Trim(strings.TrimPrefix(line, "Domain:"), " '")
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Parse balloon stats
		if strings.HasPrefix(key, "balloon.") {
			parseBaloonStat(key, value, stats)
		}

		// Parse VCPU stats
		if strings.HasPrefix(key, "vcpu.") {
			parseVCPUStat(key, value, stats, &currentVCPU)
		}

		// Parse block stats
		if strings.HasPrefix(key, "block.") {
			parseBlockStat(key, value, stats, &currentBlock)
		}
	}

	return stats, nil
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

		if operation == "rd" {
			if metric == "reqs" {
				stats.BlockStats[blockID].ReadReqs = val
			} else if metric == "bytes" {
				stats.BlockStats[blockID].ReadBytes = val
			}
		} else if operation == "wr" {
			if metric == "reqs" {
				stats.BlockStats[blockID].WriteReqs = val
			} else if metric == "bytes" {
				stats.BlockStats[blockID].WriteBytes = val
			}
		}
	}
}
