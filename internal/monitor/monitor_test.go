package monitor

import (
	"testing"
	"time"
)

func TestNewMonitor(t *testing.T) {
	m := NewMonitor()
	if m == nil {
		t.Fatal("NewMonitor() returned nil")
	}
	if m.LastUpdate.IsZero() {
		t.Error("LastUpdate should not be zero")
	}
}

func TestCPUStatsDefaults(t *testing.T) {
	var s CPUStats
	if s.UsagePercent != 0 {
		t.Errorf("expected UsagePercent 0, got %f", s.UsagePercent)
	}
	if s.CoreCount != 0 {
		t.Errorf("expected CoreCount 0, got %d", s.CoreCount)
	}
}

func TestMemStatsDefaults(t *testing.T) {
	var s MemStats
	if s.UsagePercent != 0 {
		t.Errorf("expected UsagePercent 0, got %f", s.UsagePercent)
	}
}

func TestSwapStatsDefaults(t *testing.T) {
	var s SwapStats
	if s.UsagePercent != 0 {
		t.Errorf("expected UsagePercent 0, got %f", s.UsagePercent)
	}
}

func TestDiskStatsDefaults(t *testing.T) {
	var s DiskStats
	if s.Path != "" {
		t.Errorf("expected empty Path, got %s", s.Path)
	}
}

func TestProcessInfoDefaults(t *testing.T) {
	var p ProcessInfo
	if p.PID != 0 {
		t.Errorf("expected PID 0, got %d", p.PID)
	}
	if p.Name != "" {
		t.Errorf("expected empty Name, got %s", p.Name)
	}
}

func TestNetworkStatsDefaults(t *testing.T) {
	var n NetworkStats
	if n.Name != "" {
		t.Errorf("expected empty Name, got %s", n.Name)
	}
}

func TestMonitorUpdate(t *testing.T) {
	m := NewMonitor()
	before := m.LastUpdate

	time.Sleep(time.Millisecond)
	err := m.Update()
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	if m.CPU.CoreCount == 0 {
		t.Error("expected non-zero CPU core count")
	}
	if m.Memory.TotalMB == 0 {
		t.Error("expected non-zero memory total")
	}
	if len(m.Processes) == 0 {
		t.Error("expected at least one process")
	}
	if !m.LastUpdate.After(before) {
		t.Error("LastUpdate should have been updated")
	}
}

func TestSortProcessesByCPU(t *testing.T) {
	m := NewMonitor()
	err := m.Update()
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i < len(m.Processes); i++ {
		if m.Processes[i-1].CPUPercent < m.Processes[i].CPUPercent {
			t.Error("processes not sorted by CPU descending")
			break
		}
	}
}

func TestProcessesReturned(t *testing.T) {
	m := NewMonitor()
	if err := m.Update(); err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}
	if len(m.Processes) == 0 {
		t.Error("expected at least one process")
	}
}
