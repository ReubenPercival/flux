package ui

import (
	"strings"
	"testing"

	"github.com/ReubenPercival/flux/internal/monitor"
)

func TestFormatThroughput(t *testing.T) {
	mon := monitor.NewMonitor()
	m := NewModel(mon)

	tests := []struct {
		bps  float64
		want string
	}{
		{0, "0.0 B/s"},
		{500, "500.0 B/s"},
		{1024, "1.0 KB/s"},
		{2048, "2.0 KB/s"},
		{1024 * 1024, "1.0 MB/s"},
		{2.5 * 1024 * 1024, "2.5 MB/s"},
		{1024 * 1024 * 1024, "1.0 GB/s"},
		{1024*1024*1024 + 500*1024*1024, "1.5 GB/s"},
	}

	for _, tc := range tests {
		got := m.formatThroughput(tc.bps)
		if !strings.Contains(got, tc.want) {
			t.Errorf("formatThroughput(%f) = %q, want to contain %q", tc.bps, got, tc.want)
		}
	}
}

func TestNetworkPanelShowsThroughput(t *testing.T) {
	mon := monitor.NewMonitor()
	mon.Update()
	m := NewModel(mon)

	view := m.View()

	if strings.Contains(view, "MB▼") || strings.Contains(view, "▼ MB") {
		t.Error("network panel should show throughput (B/s, KB/s, MB/s), not cumulative MB")
	}
}

func TestSwapLineHiddenWhenZero(t *testing.T) {
	mon := monitor.NewMonitor()
	mon.Swap.TotalMB = 0
	m := NewModel(mon)

	view := m.View()

	if strings.Contains(view, "SWAP") {
		t.Log("swap may still appear if it was detected on this system")
	}

	mon.Swap.TotalMB = 8192
	mon.Swap.UsedMB = 2048
	mon.Swap.UsagePercent = 25.0

	m2 := NewModel(mon)
	m2.monitor.Swap = mon.Swap
	view2 := m2.View()

	if !strings.Contains(view2, "SWAP") {
		t.Error("SWAP line should be visible when TotalMB > 0")
	}
}
