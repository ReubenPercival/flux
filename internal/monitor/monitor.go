package monitor

import (
	"fmt"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// CPUStats holds CPU information
type CPUStats struct {
	UsagePercent float64
	CoreCount    int32
	PerCPU       []float64
	LoadAvg1     float64
	LoadAvg5     float64
	LoadAvg15    float64
}

// MemStats holds memory information
type MemStats struct {
	TotalMB  uint64
	UsedMB   uint64
	FreeMB   uint64
	UsagePercent float64
}

// SwapStats holds swap information
type SwapStats struct {
	TotalMB  uint64
	UsedMB   uint64
	FreeMB   uint64
	UsagePercent float64
}

// DiskStats holds disk information
type DiskStats struct {
	Path         string
	TotalGB      float64
	UsedGB       float64
	FreeGB       float64
	UsagePercent float64
}

// ProcessInfo holds process information
type ProcessInfo struct {
	PID         int32
	Name        string
	CPUPercent  float64
	MemoryMB    uint64
	MemPercent  float64
	RuntimeSecs int64
}

// NetworkStats holds network interface statistics
type NetworkStats struct {
	Name           string
	BytesSent      uint64
	BytesRecv      uint64
	ThroughputSent float64
	ThroughputRecv float64
}

// Monitor tracks system statistics
type Monitor struct {
	CPU       CPUStats
	Memory    MemStats
	Swap      SwapStats
	Disks     []DiskStats
	GPUs      []GPUStats
	Processes []ProcessInfo
	Network   []NetworkStats
	Power     PowerStats
	LastUpdate time.Time

	raplReader      *raplReader
	lastNetwork     map[string]NetworkStats
	lastNetworkTime time.Time
	lastNvidiaTime  time.Time
}

// NewMonitor creates a new Monitor instance
func NewMonitor() *Monitor {
	return &Monitor{
		LastUpdate:  time.Now(),
		raplReader:  newRAPLReader(),
		lastNetwork: make(map[string]NetworkStats),
	}
}

// Update refreshes all system statistics
func (m *Monitor) Update() error {
	if err := m.updateCPU(); err != nil {
		return fmt.Errorf("cpu error: %w", err)
	}

	if err := m.updateMemory(); err != nil {
		return fmt.Errorf("memory error: %w", err)
	}

	if err := m.updateSwap(); err != nil {
		return fmt.Errorf("swap error: %w", err)
	}

	if err := m.updateDisks(); err != nil {
		return fmt.Errorf("disk error: %w", err)
	}

	if err := m.updateProcesses(); err != nil {
		return fmt.Errorf("process error: %w", err)
	}

	if err := m.updateGPU(); err != nil {
		return fmt.Errorf("gpu error: %w", err)
	}

	if err := m.updateNetwork(); err != nil {
		return fmt.Errorf("network error: %w", err)
	}

	m.updatePower()

	m.LastUpdate = time.Now()
	return nil
}

func (m *Monitor) updateCPU() error {
	percents, err := cpu.Percent(0, true)
	if err != nil {
		return err
	}

	count, err := cpu.Counts(false)
	if err != nil {
		return err
	}

	var total float64
	for _, p := range percents {
		total += p
	}
	if len(percents) == 0 {
		return fmt.Errorf("cpu.Percent returned no data")
	}
	avg := total / float64(len(percents))

	m.CPU.PerCPU = percents
	m.CPU.UsagePercent = avg
	m.CPU.CoreCount = int32(count)

	avgLoad, err := load.Avg()
	if err == nil {
		m.CPU.LoadAvg1 = avgLoad.Load1
		m.CPU.LoadAvg5 = avgLoad.Load5
		m.CPU.LoadAvg15 = avgLoad.Load15
	}

	return nil
}

func (m *Monitor) updateMemory() error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	m.Memory.TotalMB = v.Total / 1024 / 1024
	m.Memory.UsedMB = v.Used / 1024 / 1024
	m.Memory.FreeMB = v.Free / 1024 / 1024
	m.Memory.UsagePercent = v.UsedPercent
	return nil
}

func (m *Monitor) updateSwap() error {
	v, err := mem.SwapMemory()
	if err != nil {
		return err
	}

	m.Swap.TotalMB = v.Total / 1024 / 1024
	m.Swap.UsedMB = v.Used / 1024 / 1024
	m.Swap.FreeMB = v.Free / 1024 / 1024
	m.Swap.UsagePercent = v.UsedPercent
	return nil
}

func (m *Monitor) updateDisks() error {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return err
	}

	m.Disks = make([]DiskStats, 0)
	for _, p := range partitions {
		stat, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		m.Disks = append(m.Disks, DiskStats{
			Path:         p.Mountpoint,
			TotalGB:      float64(stat.Total) / 1024 / 1024 / 1024,
			UsedGB:       float64(stat.Used) / 1024 / 1024 / 1024,
			FreeGB:       float64(stat.Free) / 1024 / 1024 / 1024,
			UsagePercent: stat.UsedPercent,
		})
	}
	return nil
}

func (m *Monitor) updateProcesses() error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}

	m.Processes = make([]ProcessInfo, 0)
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		cpuPercent, err := p.CPUPercent()
		if err != nil {
			cpuPercent = 0
		}

		memInfo, err := p.MemoryInfo()
		if err != nil {
			continue
		}

		memPercent, err := p.MemoryPercent()
		if err != nil {
			memPercent = 0
		}

		createTime, err := p.CreateTime()
		if err != nil {
			createTime = 0
		}
		runtimeSecs := (time.Now().UnixMilli() - createTime) / 1000

		m.Processes = append(m.Processes, ProcessInfo{
			PID:        p.Pid,
			Name:        name,
			CPUPercent:  cpuPercent,
			MemoryMB:    memInfo.RSS / 1024 / 1024,
			MemPercent:  float64(memPercent),
			RuntimeSecs: runtimeSecs,
		})
	}

	// Sort by CPU usage (descending)
	sort.Slice(m.Processes, func(i, j int) bool {
		return m.Processes[i].CPUPercent > m.Processes[j].CPUPercent
	})

	return nil
}

func (m *Monitor) updateGPU() error {
	gpus := detectGPUsSysfs()

	if time.Since(m.lastNvidiaTime) > 5*time.Second {
		enrichWithNvidiaSMI(gpus)
		m.lastNvidiaTime = time.Now()
	}

	m.GPUs = gpus
	return nil
}

func (m *Monitor) updateNetwork() error {
	interfaces, err := net.IOCounters(true)
	if err != nil {
		return err
	}

	now := time.Now()
	elapsed := now.Sub(m.lastNetworkTime).Seconds()
	if m.lastNetworkTime.IsZero() || elapsed <= 0 {
		elapsed = 1
	}

	m.Network = make([]NetworkStats, 0)
	for _, iface := range interfaces {
		ns := NetworkStats{
			Name:      iface.Name,
			BytesSent: iface.BytesSent,
			BytesRecv: iface.BytesRecv,
		}

		if prev, ok := m.lastNetwork[iface.Name]; ok {
			sentDelta := iface.BytesSent - prev.BytesSent
			recvDelta := iface.BytesRecv - prev.BytesRecv
			ns.ThroughputSent = float64(sentDelta) / elapsed
			ns.ThroughputRecv = float64(recvDelta) / elapsed
		}

		m.lastNetwork[iface.Name] = ns
		m.Network = append(m.Network, ns)
	}

	m.lastNetworkTime = now
	return nil
}

func (m *Monitor) updatePower() {
	m.Power = m.raplReader.read()
}
