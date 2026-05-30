package monitor

import (
	"testing"
)

func TestGPUsDetected(t *testing.T) {
	gpus := detectGPUsSysfs()
	t.Logf("detected %d GPU(s)", len(gpus))
	for _, g := range gpus {
		t.Logf("  %s (vendor=%s usage=%.0f%% vram=%d/%dMB temp=%.0f°C)",
			g.Name, g.Vendor, g.Usage, g.MemUsed, g.MemTotal, g.Temp)
	}
}

func TestGPULookup(t *testing.T) {
	tests := []struct {
		vendor, device string
	}{
		{"8086", "46a6"},
		{"1002", "73df"},
		{"10de", "2684"},
		{"10de", "2d98"},
		{"8086", "a7ab"},
	}
	for _, tc := range tests {
		name := pciLookup(tc.vendor, tc.device)
		t.Logf("%s:%s -> %s", tc.vendor, tc.device, name)
		if name == "" {
			t.Errorf("pciLookup(%s,%s) returned empty", tc.vendor, tc.device)
		}
	}
}
