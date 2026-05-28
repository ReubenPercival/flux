# Flux

A modern, real-time system monitor for the terminal. Built with GoLang  and Bubble Tea (https://github.com/charmbracelet/bubbletea)   for a responsive, beautiful TUI experience.
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

### Performance
- **Lightweight** - ~5MB binary, minimal dependencies
- **Low Overhead** - ~1-2% CPU usage under normal operation
- **Efficient** - Handles 1000+ processes smoothly
- **No External Dependencies** - Pure Go implementation (except gopsutil)

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Linux

### Installation

**Option 1: From Source**
```bash
git clone https://github.com/ReubenPercival/flux.git
cd flux
make install
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


## Troubleshooting

### High CPU Usage
- Flux is designed to be lightweight. If you see high usage, check if you have many processes with high activity.


### Missing Process Information
- Some processes may require elevated permissions to access full details


### Slow on Startup
- First run gathers all process information
- Subsequent updates are much faster

## Supported Systems

| OS | Status | Notes |
|-----|--------|-------|
| Linux | Fully Supported | All features available |

## Performance Specifications

- **Binary Size**: ~5MB (compressed)
- **Memory Usage**: ~20-30MB typical
- **CPU Overhead**: 1-2% typical usage
- **Refresh Rate**: 1 second (configurable in future)
- **Max Processes**: 1000+ (tested on production systems)


## Contributing

Contributions are welcome! Here's how to get started:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request
OR send it over email. I accept patch over email.

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

Made with hatred by [ReubenPercival](https://github.com/ReubenPercival)

If you find Flux useful, please consider giving it a star!
