package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ReubenPercival/flux/internal/monitor"
)

var (
	colorsGreen  = lipgloss.Color("#2ecc71")
	colorYellow  = lipgloss.Color("#f39c12")
	colorRed     = lipgloss.Color("#e74c3c")
	colorCyan    = lipgloss.Color("#3498db")
	colorMagenta = lipgloss.Color("#9b59b6")
	colorWhite   = lipgloss.Color("#ecf0f1")
	colorDark    = lipgloss.Color("#2c3e50")

	titleStyle = lipgloss.NewStyle().
		Foreground(colorCyan).
		Bold(true).
		Padding(0, 2)

	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMagenta).
		Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
		Foreground(colorCyan).
		Bold(true)
)

type tickMsg time.Time

// Model represents the TUI state
type Model struct {
	monitor *monitor.Monitor
	spinner spinner.Model
	ticker  *time.Ticker
}

// NewModel creates a new UI model
func NewModel(mon *monitor.Monitor) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorGreen)

	return Model{
		monitor: mon,
		spinner: sp,
		ticker:  time.NewTicker(1 * time.Second),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Update monitor immediately
	m.monitor.Update()
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

// Update handles messages
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

// View renders the UI
func (m Model) View() string {
	output := ""

	// Header
	header := titleStyle.Render("╱═══════ FLUX ════════╲") + "\n"
	output += header

	// CPU, Memory, Swap stats
	stats := m.renderSystemStats()
	output += stats + "\n"

	// Disk usage
	disks := m.renderDisks()
	output += disks + "\n"

	// Processes
	processes := m.renderProcesses()
	output += processes + "\n"

	// Network
	network := m.renderNetwork()
	output += network + "\n"

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(colorDark).
		Render("q to quit • Updated: " + m.monitor.LastUpdate.Format("15:04:05"))
	output += footer

	return output
}

func (m Model) renderSystemStats() string {
	cpuBar := m.renderProgressBar(m.monitor.CPU.UsagePercent, 30)
	memBar := m.renderProgressBar(m.monitor.Memory.UsagePercent, 30)
	swapBar := m.renderProgressBar(m.monitor.Swap.UsagePercent, 30)

	stats := fmt.Sprintf(
		"CPU:  %s %.1f%% (%d cores)\nMEM:  %s %.1f%% (%dMB/%dMB)\nSWAP: %s %.1f%% (%dMB/%dMB)\n",
		cpuBar, m.monitor.CPU.UsagePercent, m.monitor.CPU.CoreCount,
		memBar, m.monitor.Memory.UsagePercent, m.monitor.Memory.UsedMB, m.monitor.Memory.TotalMB,
		swapBar, m.monitor.Swap.UsagePercent, m.monitor.Swap.UsedMB, m.monitor.Swap.TotalMB,
	)

	return panelStyle.Render(stats)
}

func (m Model) renderDisks() string {
	content := headerStyle.Render("📀 DISK USAGE") + "\n\n"

	for _, disk := range m.monitor.Disks {
		bar := m.renderProgressBar(disk.UsagePercent, 25)
		content += fmt.Sprintf("%s: %s %.1f%% (%.1fGB/%.1fGB)\n",
			disk.Path, bar, disk.UsagePercent, disk.UsedGB, disk.TotalGB,
		)
	}

	return panelStyle.Render(content)
}

func (m Model) renderProcesses() string {
	content := headerStyle.Render("⚙️  TOP PROCESSES") + "\n\n"
	content += fmt.Sprintf("%-8s %-20s %8s %10s %8s\n", "PID", "NAME", "CPU%", "MEM(MB)", "TIME")
	content += "─────────────────────────────────────────────\n"

	for _, proc := range m.monitor.Processes {
		if proc.CPUPercent > 0.1 || proc.MemoryMB > 10 {
			runtimeStr := fmt.Sprintf("%dh%dm", proc.RuntimeSecs/3600, (proc.RuntimeSecs%3600)/60)
			content += fmt.Sprintf("%-8d %-20s %7.1f%% %10d %8s\n",
				proc.PID, proc.Name[:minInt(len(proc.Name), 19)], proc.CPUPercent, proc.MemoryMB, runtimeStr,
			)
		}
	}

	return panelStyle.Render(content)
}

func (m Model) renderNetwork() string {
	content := headerStyle.Render("🌐 NETWORK") + "\n\n"

	for _, iface := range m.monitor.Network {
		if iface.Name == "lo" {
			continue // Skip loopback
		}
		content += fmt.Sprintf("%s: ↓ %.2f MB/s | ↑ %.2f MB/s\n",
			iface.Name,
			float64(iface.BytesRecv)/1024/1024,
			float64(iface.BytesSent)/1024/1024,
		)
	}

	return panelStyle.Render(content)
}

func (m Model) renderProgressBar(percent float64, width int) string {
	filled := int(float64(width) * percent / 100)
	empty := width - filled

	var color lipgloss.Color
	if percent < 50 {
		color = colorsGreen
	} else if percent < 75 {
		color = colorYellow
	} else {
		color = colorRed
	}

	bar := "[" +
		lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%s", repeatStr("█", filled))) +
		reptStr("░", empty) +
		"]"

	return bar
}

func repeatStr(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
