# Flux

A modern, real-time system monitor for the terminal. Written in Go with a beautiful TUI built with Bubble Tea.

## Features

- 🚀 **Real-time System Monitoring**
  - CPU usage with core count
  - Memory (RAM) usage with detailed stats
  - Swap usage tracking
  - Disk I/O and usage per partition
  - Network interface statistics

- 📊 **Process Monitoring**
  - Top 15 processes by CPU usage
  - Per-process CPU and memory metrics
  - Runtime tracking
  - Automatic sorting

- 🎨 **Modern TUI Design**
  - Color-coded status indicators
  - Clean, responsive layout
  - Real-time updates every second
  - Progress bars with health coloring

## Requirements

- Go 1.21 or higher
- Linux, macOS, or Windows (gopsutil supports all)

## Installation

### From Source

```bash
git clone https://github.com/ReubenPercival/flux.git
cd flux
make install
```

Or manually:

```bash
go build -o flux .
./flux
```

## Usage

```bash
flux
```

### Controls

- `q` or `Ctrl+C` - Quit the application

## Build & Development

### Commands

```bash
make build     # Build the binary
make run       # Build and run
make deps      # Install dependencies
make clean     # Clean build artifacts
make test      # Run tests
make install   # Install globally
make fmt       # Format code
make lint      # Lint code
```

## Architecture

```
flux/
├── main.go                  # Entry point
├── go.mod                   # Go module definition
├── Makefile                 # Build commands
├── README.md               # This file
└── internal/
    ├── monitor/            # System monitoring logic
    │   └── monitor.go      # gopsutil integration
    └── ui/                 # TUI components
        ├── ui.go           # Bubble Tea UI
        └── helpers.go      # UI utilities
```

## Dependencies

- `github.com/charmbracelet/bubbles` - TUI components (spinner)
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/shirou/gopsutil/v3` - System information gathering

## Performance

- Lightweight (~5MB binary)
- Low CPU footprint (~1-2% typical usage)
- Efficient polling interval (1 second)
- Handles 1000+ processes smoothly

## Roadmap

- [ ] Keyboard navigation and filtering
- [ ] Process filtering and searching
- [ ] Historical data visualization
- [ ] Configuration file support
- [ ] Theme customization
- [ ] GPU monitoring
- [ ] Temperature sensors
- [ ] Export to CSV/JSON

## License

GPL-3.0 - See LICENSE for details

## Contributing

Contributions are welcome! Feel free to open issues and pull requests.

---

Made with ❤️ by [ReubenPercival](https://github.com/ReubenPercival)
