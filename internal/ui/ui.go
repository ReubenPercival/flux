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
	monitor    *monitor.Monitor
	spinner    spinner.Model
	cpuHistory []float64
	width      int
	height     int
}

// panelContentWidth returns chars available inside a panel's content area.
// appStyle border(1+1)=2 + padding(2+2)=4 = 6 consumed.
// panelStyle border(1+1)=2 + padding(1+1)=2 = 4 consumed.
// Total overhead from both styles = 10.
func (m Model) panelContentWidth() int {
	if m.width <= 0 {
		return 66
	}
	if a := m.width - 10; a > 20 {
		return a
	}
	return 20
}

// mainBarWidth is the number of fillable chars inside bar brackets []
// for CPU / MEM / SWAP lines. The tightest constraint is the MEM/SWAP
// suffix " (99999MB/99999MB)" which adds ~36 chars of fixed overhead.
func (m Model) mainBarWidth() int {
	w := m.panelContentWidth() - 38
	switch {
	case w < 8:
		return 8
	case w > 80:
		return 80
	default:
		return w
	}
}

func (m Model) diskBarWidth() int {
	w := m.panelContentWidth() - 40
	switch {
	case w < 6:
		return 6
	case w > 70:
		return 70
	default:
		return w
	}
}

func (m Model) gpuBarWidth() int {
	// GPU lines vary a lot (names + VRAM + temp), keep bar modest
	w := m.panelContentWidth() * 3 / 10
	switch {
	case w < 6:
		return 6
	case w > 50:
		return 50
	default:
		return w
	}
}

func (m Model) miniBarWidth() int {
	w := m.panelContentWidth() / 10
	switch {
	case w < 3:
		return 3
	case w > 15:
		return 15
	default:
		return w
	}
}

func (m Model) procNameWidth() int {
	w := m.panelContentWidth() - 40
	switch {
	case w < 8:
		return 8
	case w > 35:
		return 35
	default:
		return w
	}
}

func (m Model) procSepWidth() int {
	w := m.panelContentWidth() - 1
	if w < 10 {
		return 10
	}
	return w
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tickMsg:
		m.monitor.Update()
		m.cpuHistory = append(m.cpuHistory, m.monitor.CPU.UsagePercent)
		if len(m.cpuHistory) > 60 {
			m.cpuHistory = m.cpuHistory[1:]
		}
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

	// Simple line count estimation for height-aware truncation
	estLines := func(s string) int {
		return strings.Count(s, "\n") + 1
	}

	titleHeight := estLines(fluxTitle)
	footerHeight := estLines(footer)
	// appStyle: top+bottom border (2) + top+bottom padding (2) = 4 vertical overhead
	// panelStyle per section: top+bottom border (2) + top+bottom padding (0, but lipgloss adds ~1)
	overhead := 4 + 2*len(sections)

	totalEst := titleHeight + overhead + footerHeight
	for _, s := range sections {
		totalEst += estLines(s)
	}

	if m.height > 0 && totalEst > m.height {
		// Remove sections from the end until it fits
		for m.height > 0 && totalEst > m.height && len(sections) > 0 {
			last := sections[len(sections)-1]
			totalEst -= estLines(last) + 2 // panel overhead
			sections = sections[:len(sections)-1]
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Top, sections...)

	return appStyle.Render(
		title + "\n\n" +
			body + "\n" +
			footer,
	)
}

func (m Model) renderSystemStats() string {
	bw := m.mainBarWidth()
	cpuBar := m.renderGradientBar(m.monitor.CPU.UsagePercent, bw)
	cpuPct := m.colorizePercent(m.monitor.CPU.UsagePercent)
	cpuLine := fmt.Sprintf("%s %s  %s (%d cores)",
		labelStyle(" CPU"), cpuBar, cpuPct, m.monitor.CPU.CoreCount)

	var cpuExtras []string
	if len(m.cpuHistory) > 0 {
		spark := lipgloss.NewStyle().Foreground(colorTeal).Render(
			m.renderSparkline(m.cpuHistory, bw))
		loadStr := m.colorizeLoad(m.monitor.CPU.LoadAvg1, m.monitor.CPU.LoadAvg5, m.monitor.CPU.LoadAvg15, m.monitor.CPU.CoreCount)
		cpuExtras = append(cpuExtras, fmt.Sprintf("      %s  %s", spark, loadStr))
	}
	if p := m.renderPower(); p != "" {
		cpuExtras = append(cpuExtras, p)
	}
	if len(m.monitor.CPU.PerCPU) > 0 {
		cpuExtras = append(cpuExtras, m.renderPerCoreCPU())
	}

	bw = m.mainBarWidth()
	memBar := m.renderGradientBar(m.monitor.Memory.UsagePercent, bw)
	memPct := m.colorizePercent(m.monitor.Memory.UsagePercent)
	memLine := fmt.Sprintf("%s %s  %s (%dMB/%dMB)",
		labelStyle(" MEM"), memBar, memPct, m.monitor.Memory.UsedMB, m.monitor.Memory.TotalMB)

	parts := []string{cpuLine}
	parts = append(parts, cpuExtras...)
	parts = append(parts, memLine)

	if m.monitor.Swap.TotalMB > 0 {
		swapBar := m.renderGradientBar(m.monitor.Swap.UsagePercent, bw)
		swapPct := m.colorizePercent(m.monitor.Swap.UsagePercent)
		swapLine := fmt.Sprintf("%s %s  %s (%dMB/%dMB)",
			labelStyle("SWAP"), swapBar, swapPct, m.monitor.Swap.UsedMB, m.monitor.Swap.TotalMB)
		parts = append(parts, swapLine)
	}

	return panelStyle.Render(strings.Join(parts, "\n"))
}

func (m Model) renderGPUs() string {
	if len(m.monitor.GPUs) == 0 {
		return ""
	}

	content := headerStyle.Render(" GPU") + "\n"

	for _, gpu := range m.monitor.GPUs {
		line := lipgloss.NewStyle().Foreground(colorCyan).Render(gpu.Name)

		if gpu.Usage >= 0 {
			bar := m.renderGradientBar(gpu.Usage, m.gpuBarWidth())
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
	content := headerStyle.Render(" DISKS") + "\n"

	for _, disk := range m.monitor.Disks {
		bar := m.renderGradientBar(disk.UsagePercent, m.diskBarWidth())
		pct := m.colorizePercent(disk.UsagePercent)
		path := lipgloss.NewStyle().Foreground(colorCyan).Render(disk.Path)
		usage := fmt.Sprintf("%.1f/%.1f GB", disk.UsedGB, disk.TotalGB)
		content += fmt.Sprintf(" %s %s %s  %s\n", path, bar, pct, labelStyle(usage))
	}

	return panelStyle.Render(strings.TrimSuffix(content, "\n"))
}

func (m Model) renderProcesses() string {
	content := headerStyle.Render(" PROCESSES") + "\n"
	nw := m.procNameWidth()
	content += lipgloss.NewStyle().Foreground(colorDim).Render(
		fmt.Sprintf(" %-7s %-*s %6s %8s %7s", "PID", nw, "NAME", "CPU%", "MEM", "UPTIME"),
	) + "\n"
	content += lipgloss.NewStyle().Foreground(colorBorder).Render(" " + strings.Repeat("─", m.procSepWidth())) + "\n"

	for i, proc := range m.monitor.Processes {
		if proc.CPUPercent < 0.1 && proc.MemoryMB < 10 {
			continue
		}

		runtime := fmt.Sprintf("%dh%02dm", proc.RuntimeSecs/3600, (proc.RuntimeSecs%3600)/60)
		name := proc.Name
		if len(name) > nw {
			name = name[:nw-1] + "…"
		}

		cpuColor := m.percentColor(proc.CPUPercent)
		cpuStr := lipgloss.NewStyle().Foreground(cpuColor).Render(fmt.Sprintf("%5.1f", proc.CPUPercent))

		memColor := m.percentColor(proc.MemPercent)
		memStr := lipgloss.NewStyle().Foreground(memColor).Render(fmt.Sprintf("%6d MB", proc.MemoryMB))

		row := fmt.Sprintf(" %-7d %-*s %s %s %7s", proc.PID, nw, name, cpuStr, memStr, runtime)

		if i%2 == 1 {
			row = lipgloss.NewStyle().Background(lipgloss.Color("#1f2335")).Render(row)
		}

		content += row + "\n"
	}

	return panelStyle.Render(strings.TrimSuffix(content, "\n"))
}

func (m Model) renderNetwork() string {
	content := headerStyle.Render(" NETWORK") + "\n"

	for _, iface := range m.monitor.Network {
		if iface.Name == "lo" {
			continue
		}
		down := m.formatThroughput(iface.ThroughputRecv)
		up := m.formatThroughput(iface.ThroughputSent)
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

func (m Model) renderSparkline(data []float64, width int) string {
	if len(data) == 0 {
		return strings.Repeat(" ", width)
	}

	if len(data) > width {
		data = data[len(data)-width:]
	}

	chars := []rune("▁▂▃▄▅▆▇█")
	var sb strings.Builder
	for _, val := range data {
		idx := int(val / 100.0 * 7)
		if idx < 0 {
			idx = 0
		}
		if idx > 7 {
			idx = 7
		}
		sb.WriteRune(chars[idx])
	}

	return sb.String()
}

func (m Model) renderMiniBar(percent float64, width int) string {
	filled := int(float64(width) * percent / 100)
	if filled > width {
		filled = width
	}

	c := m.percentColor(percent)
	filledChar := lipgloss.NewStyle().Foreground(c).Render("█")
	emptyChar := lipgloss.NewStyle().Foreground(colorBorder).Render("░")

	var sb strings.Builder
	for i := 0; i < width; i++ {
		if i < filled {
			sb.WriteString(filledChar)
		} else {
			sb.WriteString(emptyChar)
		}
	}

	return sb.String()
}

func (m Model) renderPerCoreCPU() string {
	cores := m.monitor.CPU.PerCPU
	if len(cores) == 0 {
		return ""
	}

	barWidth := m.miniBarWidth()
	var rows []string

	for i := 0; i < len(cores); i += 2 {
		var row string

		c := cores[i]
		bar := m.renderMiniBar(c, barWidth)
		pct := lipgloss.NewStyle().Foreground(m.percentColor(c)).Render(fmt.Sprintf("%5.1f%%", c))
		row += fmt.Sprintf(" C%d %s %s", i, bar, pct)

		if i+1 < len(cores) {
			c = cores[i+1]
			bar = m.renderMiniBar(c, barWidth)
			pct = lipgloss.NewStyle().Foreground(m.percentColor(c)).Render(fmt.Sprintf("%5.1f%%", c))
			row += fmt.Sprintf("  C%d %s %s", i+1, bar, pct)
		}

		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

func (m Model) colorizeLoad(load1, load5, load15 float64, cores int32) string {
	loadColor := func(l float64) lipgloss.Color {
		switch {
		case l < float64(cores)*0.7:
			return colorGreen
		case l < float64(cores)*1.5:
			return colorYellow
		default:
			return colorRed
		}
	}

	l1 := lipgloss.NewStyle().Foreground(loadColor(load1)).Render(fmt.Sprintf("%.2f", load1))
	l5 := lipgloss.NewStyle().Foreground(loadColor(load5)).Render(fmt.Sprintf("%.2f", load5))
	l15 := lipgloss.NewStyle().Foreground(loadColor(load15)).Render(fmt.Sprintf("%.2f", load15))

	return fmt.Sprintf("Load: %s %s %s", l1, l5, l15)
}

func (m Model) renderPower() string {
	p := m.monitor.Power
	if p.PackageTDP == 0 && p.PackageWatts == 0 && !p.HasRealPower {
		return ""
	}

	wattStyle := lipgloss.NewStyle().Foreground(colorYellow)
	label := lipgloss.NewStyle().Foreground(colorDim).Render

	if p.HasRealPower {
		var parts []string

		type zoneEntry struct {
			label string
			value float64
			color lipgloss.Color
		}
		zones := []zoneEntry{
			{"PKG", p.PackageWatts, colorYellow},
			{"COR", p.CoreWatts, colorCyan},
			{"UNC", p.UncoreWatts, colorPurple},
			{"DRAM", p.DramWatts, colorGreen},
			{"PSys", p.PsysWatts, colorTeal},
			{"GPU", p.GpuWatts, colorOrange},
		}

		for _, z := range zones {
			if z.value > 0 {
				style := lipgloss.NewStyle().Foreground(z.color)
				parts = append(parts, z.label+" "+style.Render(fmt.Sprintf("%.1fW", z.value)))
			}
		}

		if p.TotalWatts > 0 {
			totalStyle := lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
			parts = append(parts, "TOT "+totalStyle.Render(fmt.Sprintf("%.1fW", p.TotalWatts)))
		}

		if p.PackageTDP > 0 {
			parts = append(parts, label(fmt.Sprintf("(TDP %.0fW)", p.PackageTDP)))
		}

		if len(parts) > 0 {
			return "      " + strings.Join(parts, "  ")
		}
		return ""
	}

	var parts []string
	if p.PackageTDP > 0 {
		parts = append(parts, label("TDP")+" "+wattStyle.Render(fmt.Sprintf("%.0fW", p.PackageTDP)))
	}
	if p.PackagePL1 > 0 && p.PackagePL1 != p.PackageTDP {
		parts = append(parts, label("PL1")+" "+wattStyle.Render(fmt.Sprintf("%.0fW", p.PackagePL1)))
	}
	if p.PackagePL2 > 0 {
		parts = append(parts, label("PL2")+" "+wattStyle.Render(fmt.Sprintf("%.0fW", p.PackagePL2)))
	}
	if len(parts) == 0 {
		return ""
	}
	return "      " + strings.Join(parts, "  ")
}

func (m Model) formatThroughput(bps float64) string {
	var val float64
	var unit string
	switch {
	case bps >= 1024*1024*1024:
		val = bps / 1024 / 1024 / 1024
		unit = "GB/s"
	case bps >= 1024*1024:
		val = bps / 1024 / 1024
		unit = "MB/s"
	case bps >= 1024:
		val = bps / 1024
		unit = "KB/s"
	default:
		val = bps
		unit = "B/s"
	}
	return lipgloss.NewStyle().Foreground(colorDim).Render(fmt.Sprintf("%6.1f %s", val, unit))
}