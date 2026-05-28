package monitor

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const raplBase = "/sys/devices/virtual/powercap"

type PowerStats struct {
	PackageWatts float64
	CoreWatts    float64
	UncoreWatts  float64
	PackageTDP   float64
	PackagePL1   float64
	PackagePL2   float64
	HasRealPower bool
}

type raplReader struct {
	prevEnergy map[string]uint64
	prevTime   time.Time
}

func newRAPLReader() *raplReader {
	return &raplReader{
		prevEnergy: make(map[string]uint64),
	}
}

func readUint(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
}

type raplZone struct {
	path string
	name string
}

func discoverRAPLZones() []raplZone {
	var zones []raplZone
	filepath.WalkDir(raplBase, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		nameBytes, err := os.ReadFile(filepath.Join(path, "name"))
		if err != nil {
			return nil
		}
		zones = append(zones, raplZone{
			path: path,
			name: strings.TrimSpace(string(nameBytes)),
		})
		return nil
	})
	return zones
}

func pickPackageZone(zones []raplZone) string {
	var pkg, pkgMMIO string
	for _, z := range zones {
		if z.name == "package-0" {
			if strings.Contains(z.path, "mmio") {
				pkgMMIO = z.path
			} else {
				pkg = z.path
			}
		}
	}
	if pkg != "" {
		return pkg
	}
	return pkgMMIO
}

func readPackagePowerLimits(basePath string) (tdp, pl1, pl2 float64) {
	max, err := readUint(filepath.Join(basePath, "constraint_0_max_power_uw"))
	if err == nil {
		tdp = float64(max) / 1e6
	}

	limit, err := readUint(filepath.Join(basePath, "constraint_0_power_limit_uw"))
	if err == nil {
		pl1 = float64(limit) / 1e6
	}

	limit, err = readUint(filepath.Join(basePath, "constraint_1_power_limit_uw"))
	if err == nil {
		pl2 = float64(limit) / 1e6
	}

	return
}

func (r *raplReader) read() PowerStats {
	stats := PowerStats{}
	zones := discoverRAPLZones()
	if len(zones) == 0 {
		return stats
	}

	zoneMap := make(map[string]string)
	for _, z := range zones {
		zoneMap[z.name] = z.path
	}

	pkgPath := pickPackageZone(zones)
	corePath := zoneMap["core"]
	uncorePath := zoneMap["uncore"]

	if pkgPath != "" {
		stats.PackageTDP, stats.PackagePL1, stats.PackagePL2 = readPackagePowerLimits(pkgPath)
	}

	now := time.Now()

	// First call: store baseline energies and return (no delta yet)
	if r.prevTime.IsZero() {
		for _, key := range []string{"package-0", "core", "uncore"} {
			path, ok := zoneMap[key]
			if !ok || path == "" {
				continue
			}
			energy, err := readUint(filepath.Join(path, "energy_uj"))
			if err != nil {
				continue
			}
			r.prevEnergy[key] = energy
		}
		r.prevTime = now
		return stats
	}

	// Subsequent calls: compute delta from previous readings
	hasAny := false
	elapsed := now.Sub(r.prevTime).Seconds()
	if elapsed <= 0 {
		return stats
	}

	for key, rd := range map[string]struct {
		path string
		dest *float64
	}{
		"package-0": {pkgPath, &stats.PackageWatts},
		"core":      {corePath, &stats.CoreWatts},
		"uncore":    {uncorePath, &stats.UncoreWatts},
	} {
		if rd.path == "" {
			continue
		}
		energy, err := readUint(filepath.Join(rd.path, "energy_uj"))
		if err != nil {
			continue
		}

		prev, ok := r.prevEnergy[key]
		if !ok {
			r.prevEnergy[key] = energy
			continue
		}

		maxEnergy, _ := readUint(filepath.Join(rd.path, "max_energy_range_uj"))
		var delta uint64
		if energy >= prev {
			delta = energy - prev
		} else if maxEnergy > 0 {
			delta = (maxEnergy - prev) + energy
		} else {
			delta = energy
		}

		watts := float64(delta) / elapsed / 1e6
		if !math.IsInf(watts, 0) && !math.IsNaN(watts) {
			*rd.dest = math.Round(watts*10) / 10
			hasAny = true
		}

		r.prevEnergy[key] = energy
	}

	if hasAny {
		stats.HasRealPower = true
	}
	r.prevTime = now

	return stats
}
