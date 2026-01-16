package stats

import (
	"testing"
)

func TestParseVirshOutput(t *testing.T) {
	output := `Domain: 'noble_default'
  balloon.current=16777216
  balloon.maximum=16777216
  balloon.unused=15888880
  balloon.available=16777216
  balloon.usable=16301384
  balloon.rss=1128336
  state.state=1
  state.reason=1
  vcpu.current=4
  vcpu.maximum=4
  vcpu.0.state=1
  vcpu.0.time=50350000000
  vcpu.0.wait=0
  vcpu.0.delay=0
  vcpu.1.state=1
  vcpu.1.time=29380000000
  vcpu.1.wait=0
  vcpu.1.delay=0
  block.count=1
  block.0.name=vda
  block.0.path=/var/lib/libvirt/images/noble_default.qcow2
  block.0.rd.reqs=5000
  block.0.rd.bytes=104857600
  block.0.wr.reqs=2000
  block.0.wr.bytes=41943040
  block.0.fl.reqs=0
  block.0.allocation=1428627456
  block.0.capacity=21474836480
  block.0.physical=1429151744
`

	allStats, err := parseVirshOutput(output)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(allStats) != 1 {
		t.Fatalf("Expected 1 stats entry, got %d", len(allStats))
	}
	stats := allStats[0]

	if stats.DomainName != "noble_default" {
		t.Errorf("Expected domain name 'noble_default', got '%s'", stats.DomainName)
	}

	// Verify State
	if stats.State != 1 {
		t.Errorf("Expected state 1, got %d", stats.State)
	}
	if stats.StateReason != 1 {
		t.Errorf("Expected state reason 1, got %d", stats.StateReason)
	}

	// Verify Balloon Stats
	if stats.BalloonStats.Current != 16777216 {
		t.Errorf("Expected balloon current 16777216, got %d", stats.BalloonStats.Current)
	}
	if stats.BalloonStats.Unused != 15888880 {
		t.Errorf("Expected balloon unused 15888880, got %d", stats.BalloonStats.Unused)
	}

	// Verify VCPU Stats
	if len(stats.VCPUStats) < 2 {
		t.Fatalf("Expected at least 2 VCPUs, got %d", len(stats.VCPUStats))
	}
	if stats.VCPUStats[0].State != 1 {
		t.Errorf("Expected VCPU 0 state 1, got %d", stats.VCPUStats[0].State)
	}
	if stats.VCPUStats[0].Time != 50350000000 {
		t.Errorf("Expected VCPU 0 time 50350000000, got %d", stats.VCPUStats[0].Time)
	}

	// Verify Block Stats
	if len(stats.BlockStats) < 1 {
		t.Fatalf("Expected at least 1 Block device, got %d", len(stats.BlockStats))
	}
	if stats.BlockStats[0].Name != "vda" {
		t.Errorf("Expected block 0 name 'vda', got '%s'", stats.BlockStats[0].Name)
	}
	if stats.BlockStats[0].ReadReqs != 5000 {
		t.Errorf("Expected block 0 read reqs 5000, got %d", stats.BlockStats[0].ReadReqs)
	}
}

func TestParseVirshOutputMulti(t *testing.T) {
	output := `Domain: 'vm1'
  balloon.current=1024
Domain: 'vm2'
  balloon.current=2048
`

	allStats, err := parseVirshOutput(output)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(allStats) != 2 {
		t.Fatalf("Expected 2 stats entries, got %d", len(allStats))
	}

	if allStats[0].DomainName != "vm1" {
		t.Errorf("Expected first domain 'vm1', got '%s'", allStats[0].DomainName)
	}
	if allStats[0].BalloonStats.Current != 1024 {
		t.Errorf("Expected first balloon 1024, got %d", allStats[0].BalloonStats.Current)
	}

	if allStats[1].DomainName != "vm2" {
		t.Errorf("Expected second domain 'vm2', got '%s'", allStats[1].DomainName)
	}
	if allStats[1].BalloonStats.Current != 2048 {
		t.Errorf("Expected second balloon 2048, got %d", allStats[1].BalloonStats.Current)
	}
}

func TestParseVirshEmpty(t *testing.T) {
	output := ""
	stats, err := parseVirshOutput(output)
	if err != nil {
		t.Fatalf("Expected no error for empty output, got %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(stats))
	}
}
