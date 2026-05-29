package monitor

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type GPUStats struct {
	Name      string
	Vendor    string
	Usage     float64
	MemTotal  uint64
	MemUsed   uint64
	Temp      float64
}

func detectGPUs() []GPUStats {
	gpus := detectGPUsSysfs()
	if len(gpus) > 0 {
		enrichWithNvidiaSMI(gpus)
	}
	return gpus
}

func detectGPUsSysfs() []GPUStats {
	gpus := []GPUStats{}

	drmDir := "/sys/class/drm"
	entries, err := os.ReadDir(drmDir)
	if err != nil {
		return gpus
	}

	seen := map[string]bool{}

	for _, e := range entries {
		name := e.Name()

		if !strings.HasPrefix(name, "card") {
			continue
		}

		if seen[name] {
			continue
		}
		seen[name] = true

		devPath := filepath.Join(drmDir, name, "device")
		if _, err := os.Stat(devPath); err != nil {
			continue
		}

		vendor := readVendor(devPath)
		if vendor == "" {
			continue
		}

		gpu := GPUStats{
			Name:   readGPUName(devPath, vendor, readDeviceID(devPath)),
			Vendor: vendor,
		}
		gpu.Usage = readGPULoad(devPath)
		gpu.MemTotal, gpu.MemUsed = readGPUMem(devPath)
		gpu.Temp = readGPUTemp(devPath)

		gpus = append(gpus, gpu)
	}

	return gpus
}

func readGPUName(devPath, vendor, device string) string {
	name := pciLookup(vendor, device)
	if name != "" {
		return name
	}
	return fallbackGPUVendorName(vendor, device)
}

func readVendor(devPath string) string {
	v, err := readFile(filepath.Join(devPath, "vendor"))
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(v, "0x")
}

func readDeviceID(devPath string) string {
	d, err := readFile(filepath.Join(devPath, "device"))
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(d, "0x")
}

func readGPULoad(devPath string) float64 {
	load, err := readFile(filepath.Join(devPath, "gpu_busy_percent"))
	if err != nil {
		return -1
	}
	v, err := strconv.ParseFloat(load, 64)
	if err != nil {
		return -1
	}
	return v
}

func readGPUMem(devPath string) (total, used uint64) {
	totalStr, err := readFile(filepath.Join(devPath, "mem_info_vram_total"))
	if err != nil {
		return 0, 0
	}
	usedStr, err := readFile(filepath.Join(devPath, "mem_info_vram_used"))
	if err != nil {
		return 0, 0
	}

	t, err := strconv.ParseUint(totalStr, 10, 64)
	if err != nil {
		return 0, 0
	}

	u, err := strconv.ParseUint(usedStr, 10, 64)
	if err != nil {
		return 0, 0
	}

	// sysfs values are in bytes; convert to mebibytes (MiB)
	return t / 1024 / 1024, u / 1024 / 1024
}

func readGPUTemp(devPath string) float64 {
	hwmonDir := filepath.Join(devPath, "hwmon")
	hwmonEntries, err := os.ReadDir(hwmonDir)
	if err != nil {
		return -1
	}

	for _, h := range hwmonEntries {
		tempPath := filepath.Join(hwmonDir, h.Name(), "temp1_input")
		tempStr, err := readFile(tempPath)
		if err != nil {
			continue
		}
		v, err := strconv.ParseInt(tempStr, 10, 64)
		if err != nil {
			continue
		}
		return float64(v) / 1000.0
	}

	return -1
}

func enrichWithNvidiaSMI(gpus []GPUStats) {
	hasNVIDIA := false
	for _, g := range gpus {
		if g.Vendor == "10de" {
			hasNVIDIA = true
			break
		}
	}
	if !hasNVIDIA {
		return
	}

	data, err := exec.Command("nvidia-smi",
		"--query-gpu=index,name,utilization.gpu,memory.total,memory.used,temperature.gpu",
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		return
	}

	reader := csv.NewReader(strings.NewReader(strings.TrimSpace(string(data))))
	for {
		parts, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("GPU: failed to parse nvidia-smi output: %v", err)
			continue
		}

		if len(parts) < 6 {
			continue
		}

		idx, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil || idx < 0 || idx >= len(gpus) {
			continue
		}

		gpus[idx].Name = strings.TrimSpace(parts[1])

		usage, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		if err == nil {
			gpus[idx].Usage = usage
		}

		// nvidia-smi returns memory in MiB (same unit as sysfs readGPUMem)
		memTotal, err := strconv.ParseUint(strings.TrimSpace(parts[3]), 10, 64)
		if err == nil {
			gpus[idx].MemTotal = memTotal
		}

		memUsed, err := strconv.ParseUint(strings.TrimSpace(parts[4]), 10, 64)
		if err == nil {
			gpus[idx].MemUsed = memUsed
		}

		temp, err := strconv.ParseFloat(strings.TrimSpace(parts[5]), 64)
		if err == nil {
			gpus[idx].Temp = temp
		}
	}
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func fallbackGPUVendorName(vendor, device string) string {
	vendorName := map[string]string{
		"1002": "AMD",
		"8086": "Intel",
		"10de": "NVIDIA",
	}[vendor]
	if vendorName == "" {
		vendorName = "GPU (0x" + vendor + ")"
	}
	return fmt.Sprintf("%s GPU (0x%s)", vendorName, device)
}

func pciLookup(vendor, device string) string {
	key := vendor + ":" + device
	switch key {
	case "1002:1640":
		return "AMD Phoenix (RDNA 3)"
	case "1002:15bf":
		return "AMD Raphael (RDNA 2)"
	case "1002:73df":
		return "AMD Navi 21 (RDNA 2)"
	case "1002:73ff":
		return "AMD Navi 23 (RDNA 2)"
	case "1002:74a1":
		return "AMD Navi 33 (RDNA 3)"
	case "1002:7480":
		return "AMD Navi 31 (RDNA 3)"
	case "1002:67df":
		return "AMD Polaris"
	case "8086:46a6":
		return "Intel Alder Lake-P GT2 (Iris Xe)"
	case "8086:46a1":
		return "Intel Alder Lake-H GT1 (UHD)"
	case "8086:46b3":
		return "Intel Alder Lake GT1 (UHD)"
	case "8086:9a49":
		return "Intel Tiger Lake-LP GT2 (Iris Xe)"
	case "8086:9a78":
		return "Intel Tiger Lake-H GT1 (UHD)"
	case "8086:7d45":
		return "Intel Meteor Lake-P GT2 (Arc)"
	case "8086:7d55":
		return "Intel Meteor Lake-H GT2 (Arc)"
	case "8086:a721":
		return "Intel Raptor Lake-P GT1 (UHD)"
	case "8086:a7a0":
		return "Intel Raptor Lake-P GT2 (Iris Xe)"
	case "8086:a7ab":
		return "Intel Raptor Lake-P GT1 (UHD)"
	case "8086:a7a9":
		return "Intel Raptor Lake-P GT2 (Iris Xe)"
	case "8086:56a0":
		return "Intel Arrow Lake GT1"
	case "8086:56a1":
		return "Intel Arrow Lake GT2"
	case "8086:4e61":
		return "Intel DG1 (Iris Xe MAX)"
	case "8086:4905":
		return "Intel DG2 (Arc A770)"
	case "8086:4906":
		return "Intel DG2 (Arc A750)"
	case "8086:4908":
		return "Intel DG2 (Arc A580)"
	case "8086:56c0":
		return "Intel Battlemage (Arc B580)"
	case "8086:56c1":
		return "Intel Battlemage (Arc B570)"
	case "8086:e2c0":
		return "Intel Panther Lake GT1"
	case "8086:e2c1":
		return "Intel Panther Lake GT2"
	case "10de:2684":
		return "NVIDIA GeForce RTX 4090"
	case "10de:2685":
		return "NVIDIA GeForce RTX 4080 Super"
	case "10de:2783":
		return "NVIDIA GeForce RTX 5090"
	case "10de:27b0":
		return "NVIDIA GeForce RTX 5080"
	case "10de:2d98":
		return "NVIDIA GeForce RTX 5050 Laptop GPU"
	case "10de:2d99":
		return "NVIDIA GeForce RTX 5060 Laptop GPU"
	case "10de:2d9a":
		return "NVIDIA GeForce RTX 5070 Laptop GPU"
	case "10de:24c9":
		return "NVIDIA GeForce RTX 3090"
	case "10de:24dd":
		return "NVIDIA GeForce RTX 3080 Ti"
	case "10de:2206":
		return "NVIDIA GeForce RTX 3080"
	case "10de:2204":
		return "NVIDIA GeForce RTX 3070 Ti"
	case "10de:2482":
		return "NVIDIA GeForce RTX 3070"
	case "10de:2484":
		return "NVIDIA GeForce RTX 3060 Ti"
	case "10de:2503":
		return "NVIDIA GeForce RTX 3060"
	case "10de:28e1":
		return "NVIDIA GeForce RTX 4060"
	case "10de:28e0":
		return "NVIDIA GeForce RTX 4060 Ti"
	case "10de:27a0":
		return "NVIDIA GeForce RTX 4070"
	case "10de:2786":
		return "NVIDIA GeForce RTX 4070 Ti"
	case "10de:2782":
		return "NVIDIA GeForce RTX 4070 Super"
	case "10de:24b0":
		return "NVIDIA GeForce RTX 4080"
	case "10de:1e81":
		return "NVIDIA GeForce RTX 2080 Super"
	case "10de:1e87":
		return "NVIDIA GeForce RTX 2080 Ti"
	case "10de:1e04":
		return "NVIDIA GeForce RTX 2070 Super"
	case "10de:1f02":
		return "NVIDIA GeForce RTX 2060 Super"
	case "10de:1f04":
		return "NVIDIA GeForce RTX 2060"
	case "10de:1b80":
		return "NVIDIA GeForce GTX 1080"
	case "10de:1b81":
		return "NVIDIA GeForce GTX 1070"
	case "10de:1c03":
		return "NVIDIA GeForce GTX 1060"
	case "10de:13c2":
		return "NVIDIA GeForce GTX 980"
	case "10de:13c0":
		return "NVIDIA GeForce GTX 970"
	case "10de:13ba":
		return "NVIDIA GeForce GTX 960"
	case "10de:1d01":
		return "NVIDIA GeForce GTX 1050 Ti"
	case "10de:1c81":
		return "NVIDIA GeForce GTX 1050"
	case "10de:1c82":
		return "NVIDIA GeForce GTX 1650"
	case "10de:1f95":
		return "NVIDIA GeForce GTX 1650 Ti"
	case "10de:1f91":
		return "NVIDIA GeForce GTX 1660"
	case "10de:1f94":
		return "NVIDIA GeForce GTX 1660 Ti"
	case "10de:21c4":
		return "NVIDIA GeForce RTX 3050"
	case "10de:25a0":
		return "NVIDIA GeForce RTX 3050 (GA107)"
	case "10de:1eb1":
		return "NVIDIA Quadro RTX 4000"
	case "10de:1eb5":
		return "NVIDIA Quadro RTX 5000"
	case "10de:1e30":
		return "NVIDIA Quadro RTX 6000/8000"
	case "10de:2331":
		return "NVIDIA RTX 2000 Ada"
	case "10de:27b8":
		return "NVIDIA RTX 5000 Ada"
	default:
		return ""
	}
}
