package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ReubenPercival/flux/internal/monitor"
)

const fluxTitle = "    ╭──────────────────╮\n    │     FLUX         │\n    ╰──────────────────╯"

var (
	colorCyan     = lipgloss.Color("#7dcfff")
	colorGreen    = lipgloss.Color("#9ece6a")
	colorYellow   = lipgloss.Color("#e0af68")
	colorOrange   = lipgloss.Color("#ff9e64")
	colorRed      = lipgloss.Color("#f7768e")
	colorPurple   = lipgloss.Color("#bb9af7")
	colorTeal     = lipgloss.Color("#2ac3de")
	colorDim      = lipgloss.Color("#565f89")
	colorBorder   = lipgloss.Color("#3b4261")

	appStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
		Foreground(colorCyan).
		Bold(true).
		MarginBottom(1).
		Render

	headerStyle = lipgloss.NewStyle().
		Foreground(colorTeal).
		Bold(true)

	labelStyle = lipgloss.NewStyle().
		Foreground(colorDim).
		Render

	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1)
)

type tickMsg time.Time

// Model represents the UI state
// Note: Model is not shared across goroutines. Bubble Tea runs the Update and View
// functions sequentially on a single goroutine, so the model is thread-safe within
// the context of Bubble Tea's event loop.
type Model struct {
	monitor *monitor.Monitor
	spinner spinner.Model
}

func NewModel(mon *monitor.Monitor) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorTeal)

	return Model{
		monitor: mon,
		spinner: sp,
	}
}

func (m Model) Init() tea.Cmd {
	m.monitor.Update()
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tickMsg:
		m.monitor.Update()
		return m, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	title := titleStyle(fluxTitle)

	stats := m.renderSystemStats()
	disks := m.renderDisks()
	gpus := m.renderGPUs()
	procs := m.renderProcesses()
	netw := m.renderNetwork()
	footer := m.renderFooter()

	var sections []string
	sections = append(sections, stats)
	if gpus != "" {
		sections = append(sections, gpus)
	}
	sections = append(sections, disks)
	sections = append(sections, procs)
	sections = append(sections, netw)

	body := lipgloss.JoinVertical(lipgloss.Top, sections...)

	return appStyle.Render(
		title + "\n\n" +
			body + "\n" +
			footer,
	)
}

func (m Model) renderSystemStats() string {
	cpuBar := m.renderGradientBar(m.monitor.CPU.UsagePercent, 30)
	memBar := m.renderGradientBar(m.monitor.Memory.UsagePercent, 30)
	swapBar := m.renderGradientBar(m.monitor.Swap.UsagePercent, 30)

	cpuPct := m.colorizePercent(m.monitor.CPU.UsagePercent)
	memPct := m.colorizePercent(m.monitor.Memory.UsagePercent)
	swapPct := m.colorizePercent(m.monitor.Swap.UsagePercent)

	stats := fmt.Sprintf(
		"%s %s  %s (%d cores)\n%s %s  %s (%dMB/%dMB)\n%s %s  %s (%dMB/%dMB)",
		labelStyle(" CPU"), cpuBar, cpuPct, m.monitor.CPU.CoreCount,
		labelStyle(" MEM"), memBar, memPct, m.monitor.Memory.UsedMB, m.monitor.Memory.TotalMB,
		labelStyle("SWAP"), swapBar, swapPct, m.monitor.Swap.UsedMB, m.monitor.Swap.TotalMB,
	)

	return panelStyle.Render(stats)
}

func (m Model) renderGPUs() string {
	if len(m.monitor.GPUs) == 0 {
		return ""
	}

	content := headerStyle.Render(" 🎮 GPU") + "\n"

	for _, gpu := range m.monitor.GPUs {
		line := lipgloss.NewStyle().Foreground(colorCyan).Render(gpu.Name)

		if gpu.Usage >= 0 {
			bar := m.renderGradientBar(gpu.Usage, 20)
			pct := m.colorizePercent(gpu.Usage)
			line += fmt.Sprintf("  %s %s", bar, pct)
		}

		if gpu.MemTotal > 0 {
			memStyle := lipgloss.NewStyle().Foreground(colorPurple).Render
			line += fmt.Sprintf("  %s", memStyle(fmt.Sprintf("VRAM %d/%d MB", gpu.MemUsed, gpu.MemTotal)))
		}

		if gpu.Temp >= 0 {
			tmp := m.colorizeTemp(gpu.Temp)
			line += fmt.Sprintf("  %s", tmp)
		}

		content += line + "\n"
	}

	return panelStyle.Render(strings.TrimSuffix(content, "\n"))
}

func (m Model) renderDisks() string {
	content := headerStyle.Render(" 💾 DISKS") + "\n"

	for _, disk := range m.monitor.Disks {
		bar := m.renderGradientBar(disk.UsagePercent, 25)
		pct := m.colorizePercent(disk.UsagePercent)
		path := lipgloss.NewStyle().Foreground(colorCyan).Render(disk.Path)
		usage := fmt.Sprintf("%.1f/%.1f GB", disk.UsedGB, disk.TotalGB)
		content += fmt.Sprintf(" %s %s %s  %s\n", path, bar, pct, labelStyle(usage))
	}

	return panelStyle.Render(strings.TrimSuffix(content, "\n"))
}

func (m Model) renderProcesses() string {
	content := headerStyle.Render(" ⚙ PROCESSES") + "\n"
	content += lipgloss.NewStyle().Foreground(colorDim).Render(
		fmt.Sprintf(" %-7s %-19s %6s %8s %7s", "PID", "NAME", "CPU%", "MEM", "UPTIME"),
	) + "\n"
	content += lipgloss.NewStyle().Foreground(colorBorder).Render(" " + strings.Repeat("─", 50)) + "\n"

	for i, proc := range m.monitor.Processes {
		if proc.CPUPercent < 0.1 && proc.MemoryMB < 10 {
			continue
		}

		runtime := fmt.Sprintf("%dh%02dm", proc.RuntimeSecs/3600, (proc.RuntimeSecs%3600)/60)
		name := proc.Name
		if len(name) > 19 {
			name = name[:18] + "…"
		}

		cpuColor := m.percentColor(proc.CPUPercent)
		cpuStr := lipgloss.NewStyle().Foreground(cpuColor).Render(fmt.Sprintf("%5.1f", proc.CPUPercent))

		memColor := m.percentColor(proc.MemPercent)
		memStr := lipgloss.NewStyle().Foreground(memColor).Render(fmt.Sprintf("%6d MB", proc.MemoryMB))

		row := fmt.Sprintf(" %-7d %-19s %s %s %7s",
			proc.PID, name, cpuStr, memStr, runtime)

		if i%2 == 1 {
			row = lipgloss.NewStyle().Background(lipgloss.Color("#1f2335")).Render(row)
		}

		content += row + "\n"
	}

	return panelStyle.Render(strings.TrimSuffix(content, "\n"))
}

func (m Model) renderNetwork() string {
	content := headerStyle.Render(" 🌐 NETWORK") + "\n"

	for _, iface := range m.monitor.Network {
		if iface.Name == "lo" {
			continue
		}
		down := fmt.Sprintf("%.1f MB", float64(iface.BytesRecv)/1024/1024)
		up := fmt.Sprintf("%.1f MB", float64(iface.BytesSent)/1024/1024)
		content += fmt.Sprintf(" %s  %s %s  %s %s\n",
			lipgloss.NewStyle().Foreground(colorCyan).Render(iface.Name),
			lipgloss.NewStyle().Foreground(colorGreen).Render("▼"),
			down,
			lipgloss.NewStyle().Foreground(colorRed).Render("▲"),
			up,
		)
	}

	return panelStyle.Render(strings.TrimSuffix(content, "\n"))
}

func (m Model) renderFooter() string {
	timeStr := lipgloss.NewStyle().Foreground(colorTeal).Render(m.monitor.LastUpdate.Format("15:04:05"))
	hint := lipgloss.NewStyle().Foreground(colorDim).Render("q/ctrl+c")
	return fmt.Sprintf("%s  updated %s", hint, timeStr)
}

func (m Model) renderGradientBar(percent float64, width int) string {
	filled := int(float64(width) * percent / 100)
	if filled > width {
		filled = width
	}

	bar := "["

	for i := 0; i < width; i++ {
		if i < filled {
			ratio := float64(i) / float64(width)
			bar += m.barColor(ratio, percent)
		} else {
			bar += lipgloss.NewStyle().Foreground(colorBorder).Render("░")
		}
	}

	bar += "]"
	return bar
}

func (m Model) barColor(ratio, percent float64) string {
	var c lipgloss.Color
	switch {
	case percent < 50:
		if ratio < 0.5 {
			c = lipgloss.Color("#3d59a1")
		} else {
			c = colorGreen
		}
	case percent < 75:
		if ratio < 0.33 {
			c = lipgloss.Color("#3d59a1")
		} else if ratio < 0.66 {
			c = colorYellow
		} else {
			c = colorOrange
		}
	default:
		if ratio < 0.25 {
			c = lipgloss.Color("#3d59a1")
		} else if ratio < 0.5 {
			c = colorOrange
		} else if ratio < 0.75 {
			c = colorRed
		} else {
			c = lipgloss.Color("#db4b4b")
		}
	}
	return lipgloss.NewStyle().Foreground(c).Render("█")
}

func (m Model) colorizePercent(percent float64) string {
	c := m.percentColor(percent)
	return lipgloss.NewStyle().Foreground(c).Render(fmt.Sprintf("%5.1f%%", percent))
}

func (m Model) colorizeTemp(temp float64) string {
	var c lipgloss.Color
	switch {
	case temp < 50:
		c = colorGreen
	case temp < 70:
		c = colorYellow
	case temp < 85:
		c = colorOrange
	default:
		c = colorRed
	}
	return lipgloss.NewStyle().Foreground(c).Render(fmt.Sprintf("%.0f°C", temp))
}

func (m Model) percentColor(percent float64) lipgloss.Color {
	switch {
	case percent < 50:
		return colorGreen
	case percent < 75:
		return colorYellow
	default:
		return colorRed
	}
}