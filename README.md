# Flux

Real-time system monitor for the terminal. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) — the elegant Go TUI framework from [Charm](https://charm.sh/).

- **Metrics**: CPU, RAM, swap, disk, network I/O
- **Processes**: Top 15 by CPU, live refresh every second
- **Lightweight**: ~5MB binary, 1-2% CPU overhead

```bash
git clone https://github.com/ReubenPercival/flux.git && cd flux
make install   # build & install
make run       # build & run
make test      # run tests
```

**Depends on**: charmbracelet/{bubbletea,bubbles,lipgloss}, shirou/gopsutil.

Linux only. GPL v3.

[Issues](https://github.com/ReubenPercival/flux/issues)
