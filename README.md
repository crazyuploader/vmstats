# vmstats

A beautiful terminal UI for monitoring libvirt/KVM virtual machine statistics in real-time.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue)

## Features

- 📊 **Real-time monitoring** of VM memory, CPU, and disk usage
- 🎨 **Color-coded progress bars** (green/yellow/red based on thresholds)
- 📋 **VM sidebar** showing all VMs at a glance with status icons
- ⚡ **Configurable refresh rate**
- 🔄 **Multi-VM navigation** with keyboard shortcuts
- 💤 **Smart display** - hides irrelevant metrics for offline VMs

## Installation

### From Source

```bash
git clone https://github.com/crazyuploader/vmstats.git
cd vmstats
make build
```

### Requirements

- Go 1.21+
- `virsh` command (libvirt)
- Running libvirt daemon

## Usage

```bash
# Monitor all VMs
./bin/vmstats

# Monitor specific VMs
./bin/vmstats -domains "vm1,vm2"

# Custom refresh rate (5 seconds)
./bin/vmstats -refresh 5

# Enable logging to file (optional)
./bin/vmstats -log vmstats.log

# Show version
./bin/vmstats --version
```

## Keyboard Shortcuts

| Key                     | Action         |
| ----------------------- | -------------- |
| `↓` / `j` / `Tab`       | Next VM        |
| `↑` / `k` / `Shift+Tab` | Previous VM    |
| `r`                     | Manual refresh |
| `?`                     | Toggle help    |
| `q` / `Ctrl+C`          | Quit           |

## Development

```bash
# Run tests
make test

# Build
make build

# Run with development settings
go run ./cmd/vmstats/main.go

# Lint
golangci-lint run
```

## License

MIT License - see [LICENSE](LICENSE) for details.
