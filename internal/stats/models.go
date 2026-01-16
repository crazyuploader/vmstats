package stats

// VMStats holds all the statistics for a domain
type VMStats struct {
	DomainName     string
	BalloonStats   BalloonStats
	VCPUStats      []VCPUStats
	BlockStats     []BlockStats
	InterfaceStats []InterfaceStats
	State          int
	StateReason    int
	LastUpdate     int64
}

// BalloonStats holds memory statistics
type BalloonStats struct {
	Current   int64
	Maximum   int64
	Unused    int64
	Available int64
	Usable    int64
	RSS       int64
}

// VCPUStats holds stats for a single virtual CPU
type VCPUStats struct {
	ID        int
	State     int
	Time      int64
	Exits     int64
	HaltExits int64
	IRQExits  int64
	IOExits   int64
}

// BlockStats holds stats for a block device
type BlockStats struct {
	Name       string
	Path       string
	ReadReqs   int64
	ReadBytes  int64
	WriteReqs  int64
	WriteBytes int64
	Allocation int64
	Capacity   int64
	Physical   int64
}

// InterfaceStats holds stats for a network interface
type InterfaceStats struct {
	Name      string
	RxBytes   int64
	RxPackets int64
	TxBytes   int64
	TxPackets int64
	RxDrop    int64
	RxErrs    int64
	TxDrop    int64
	TxErrs    int64
}
