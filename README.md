# Flux

A modern, real-time system monitor for the terminal. Built with Go and Bubble Tea for a responsive, beautiful TUI experience.

![License](https://img.shields.io/badge/License-GPL3-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-success)

## Features

### Real-time System Monitoring
- **CPU Metrics** - Usage percentage with core count
- **Memory Tracking** - RAM usage with total/free/used breakdown
- **Swap Management** - Swap memory utilization
- **Disk Usage** - Per-partition storage overview with percentage
- **Network Statistics** - Real-time I/O for active interfaces

### Process Management
- **Top Processes** - Top 15 processes sorted by CPU usage
- **Detailed Metrics** - CPU%, memory (MB), runtime for each process
- **Live Updates** - Automatic refresh every second
- **Smart Filtering** - Shows only processes with meaningful usage

### Beautiful Terminal Interface
- **Color-Coded Status** 
  - Green: 0-50% (healthy)
  - Yellow: 50-75% (caution)
  - Red: 75-100% (critical)
- **Visual Progress Bars** - Easy-to-scan resource utilization
- **Responsive Layout** - Adapts to terminal size
- **Real-time Updates** - 1-second refresh interval

### Performance
- **Lightweight** - ~5MB binary, minimal dependencies
- **Low Overhead** - ~1-2% CPU usage under normal operation
- **Efficient** - Handles 1000+ processes smoothly
- **No External Dependencies** - Pure Go implementation (except gopsutil)

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Linux, macOS, or Windows

### Installation

**Option 1: From Source**
```bash
git clone https://github.com/ReubenPercival/flux.git
cd flux
make install
```

**Option 2: Using Go**
```bash
go install github.com/ReubenPercival/flux@latest
```

**Option 3: Manual Build**
```bash
git clone https://github.com/ReubenPercival/flux.git
cd flux
go mod download
go build -o flux .
./flux
```

### Running

```bash
flux
```

## Controls

| Key | Action |
|-----|--------|
| `q` | Quit application |
| `Ctrl+C` | Quit application |

## Project Structure

```
flux/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── Makefile               # Build automation
├── README.md              # This file
├── LICENSE                # GPL-3.0 License
├── .gitignore             # Git ignore rules
└── internal/
    ├── monitor/           # System monitoring module
    │   └── monitor.go     # gopsutil integration & data collection
    └── ui/                # Terminal UI module
        ├── ui.go          # Bubble Tea UI components
        └── helpers.go     # Utility functions
```

## Development

### Build Commands

```bash
# Build binary
make build

# Build and run
make run

# Install dependencies
make deps

# Clean build artifacts
make clean

# Run tests
make test

# Format code
make fmt

# Lint code (requires golangci-lint)
make lint

# Install globally
make install
```

### Development Setup

```bash
# Clone repository
git clone https://github.com/ReubenPercival/flux.git
cd flux

# Install dependencies
go mod download
go mod tidy

# Run in development
go run main.go
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/charmbracelet/bubbletea` | TUI framework & event loop |
| `github.com/charmbracelet/bubbles` | UI components (spinner, etc.) |
| `github.com/charmbracelet/lipgloss` | Terminal styling & colors |
| `github.com/shirou/gopsutil/v3` | System information gathering |

All dependencies are automatically managed via `go.mod`.

## Features Showcase

### System Overview Panel
```
┌─ FLUX ─────────────────────────────────────────────┐
│ CPU:  ████████░░ 82% (8 cores)                     │
│ MEM:  ██████░░░░ 61% (16GB/26GB)                   │
│ SWAP: ░░░░░░░░░░ 0% (0MB/2GB)                      │
└────────────────────────────────────────────────────┘
```

### Disk Usage Panel
```
┌─ DISK USAGE ───────────────────────────────────────┐
│ /     : ██████████░░ 82% (412GB/500GB)             │
│ /home : ████████░░░░ 68% (680GB/1TB)               │
│ /var  : ███░░░░░░░░░ 27% (27GB/100GB)              │
└────────────────────────────────────────────────────┘
```

### Top Processes Panel
```
┌─ TOP PROCESSES ────────────────────────────────────┐
│ PID      NAME                     CPU%       MEM   │
│ ─────────────────────────────────────────────────  │
│ 2847     Firefox               12.5%      2.4GB   │
│ 1923     Go Build               8.2%      512MB   │
│ 4012     Node                   4.1%      1.1GB   │
└────────────────────────────────────────────────────┘
```

## Roadmap

### In Development
- [ ] Keyboard navigation
- [ ] Process filtering & searching
- [ ] Sort by different metrics (CPU, memory, uptime)
- [ ] Configuration file support

### Planned Features
- [ ] Historical data visualization
- [ ] Custom theme support
- [ ] GPU monitoring (NVIDIA/AMD)
- [ ] Temperature sensors
- [ ] I/O statistics per process
- [ ] Export to CSV/JSON
- [ ] Custom color schemes
- [ ] Lightweight mode (reduced refresh)
- [ ] Process tree view
- [ ] Log file viewer integration

## Troubleshooting

### High CPU Usage
- Flux is designed to be lightweight. If you see high usage, check if you have many processes with high activity.
- Try reducing terminal refresh rate in future config options.

### Missing Process Information
- Some processes may require elevated permissions to access full details
- Run with `sudo` for complete system information (not recommended for security)

### Slow on Startup
- First run gathers all process information
- Subsequent updates are much faster

## Supported Systems

| OS | Status | Notes |
|-----|--------|-------|
| Linux | Fully Supported | All features available |
| macOS | Fully Supported | All features available |
| Windows | Fully Supported | All features available |

## Performance Specifications

- **Binary Size**: ~5MB (compressed)
- **Memory Usage**: ~20-30MB typical
- **CPU Overhead**: 1-2% typical usage
- **Refresh Rate**: 1 second (configurable in future)
- **Max Processes**: 1000+ (tested on production systems)

## Configuration (Planned)

Future versions will support:
```yaml
# ~/.config/flux/config.yml
refresh_rate: 1s
process_limit: 15
show_swap: true
theme: "dark"
```

## Contributing

Contributions are welcome! Here's how to get started:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Guidelines
- Write clean, idiomatic Go code
- Add tests for new features
- Update README for new functionality
- Follow the existing code style
- Keep commits atomic and well-documented

## License

Flux is licensed under the **GNU General Public License v3.0**. See the [LICENSE](LICENSE) file for details.

This ensures the project remains open source and free for everyone to use, modify, and distribute.

## Acknowledgments

- [Charm](https://charm.sh/) - For the amazing Bubble Tea TUI framework
- [gopsutil](https://github.com/shirou/gopsutil) - For system information gathering
- [htop](https://htop.dev/) & [btop++](https://github.com/aristocratos/btop) - Inspiration
- The Go community - For incredible tools and libraries

## Support

Have questions or issues? 

- **Issues** - [GitHub Issues](https://github.com/ReubenPercival/flux/issues)
- **Discussions** - [GitHub Discussions](https://github.com/ReubenPercival/flux/discussions)
- **Email** - Check the repository for contact info

## Learning Resources

Want to understand how Flux works?

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [gopsutil Guide](https://github.com/shirou/gopsutil)
- [Lipgloss Styling](https://github.com/charmbracelet/lipgloss)
- [Go Module Reference](https://golang.org/ref/mod)

## Changelog

### v0.1.0 (Initial Release)
- Core system monitoring (CPU, RAM, Swap)
- Disk usage tracking
- Top processes display
- Network statistics
- Beautiful TUI with color coding
- Real-time updates every second

---

Made with care by [ReubenPercival](https://github.com/ReubenPercival)

If you find Flux useful, please consider giving it a star!
