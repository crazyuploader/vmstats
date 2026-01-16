package version

// These variables are set at build time using -ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// GetVersionString returns a formatted version string
func GetVersionString() string {
	return Version + " (" + GitCommit[:7] + ")"
}

// GetFullVersionString returns detailed version info
func GetFullVersionString() string {
	return "vmstats " + Version + "\nCommit: " + GitCommit + "\nBuilt: " + BuildDate
}
