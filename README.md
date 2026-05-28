# Flux

Real-time system monitor for the terminal. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **CPU**: aggregate + per-core usage bars, rolling 60s sparkline, load averages (1m/5m/15m)
- **Power**: real-time package/core/uncore wattage via Intel RAPL (when run as root), or TDP/PL1/PL2 limits
- **Memory & Swap**: usage bars with numeric totals
- **GPU**: usage bar, VRAM, and temperature for NVIDIA/AMD
- **Disks**: per-mountpoint usage bars
- **Processes**: top 15 by CPU usage, refreshed every second
- **Network**: cumulative bytes sent/received per interface

```bash
git clone https://github.com/ReubenPercival/flux.git && cd flux
make install   # build & install
make run       # build & run
make test      # run tests
```

Real-time power monitoring requires read access to RAPL `energy_uj` files. Run with `sudo` or add a udev rule:

```
SUBSYSTEM=="powercap", KERNEL=="intel-rapl*", ACTION=="add", RUN+="/bin/chmod 644 %S%p/energy_uj"
```

**Built with**: charmbracelet/{bubbletea,bubbles,lipgloss}, shirou/gopsutil.

Linux only. GPL v3.

[Issues](https://github.com/ReubenPercival/flux/issues)
