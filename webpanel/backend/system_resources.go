package main
import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)
type SystemResources struct {
	CPU       CPUInfo       `json:"cpu"`
	Memory    MemoryInfo    `json:"memory"`
	Disk      DiskInfo      `json:"disk"`
	Network   NetworkInfo   `json:"network"`
	Processes []ProcessInfo `json:"processes"`
	Timestamp time.Time     `json:"timestamp"`
}
type CPUInfo struct {
	Usage       float64   `json:"usage"`
	Cores       int       `json:"cores"`
	LoadAvg     []float64 `json:"load_avg"`
	Temperature float64   `json:"temperature"`
}
type MemoryInfo struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Usage     float64 `json:"usage"`
	SwapTotal uint64  `json:"swap_total"`
	SwapUsed  uint64  `json:"swap_used"`
}
type DiskInfo struct {
	Total       uint64       `json:"total"`
	Used        uint64       `json:"used"`
	Available   uint64       `json:"available"`
	Usage       float64      `json:"usage"`
	MountPoints []MountPoint `json:"mount_points"`
}
type NetworkInfo struct {
	BytesReceived   uint64             `json:"bytes_received"`
	BytesSent       uint64             `json:"bytes_sent"`
	PacketsReceived uint64             `json:"packets_received"`
	PacketsSent     uint64             `json:"packets_sent"`
	Interfaces      []NetworkInterface `json:"interfaces"`
}
type MountPoint struct {
	Device     string  `json:"device"`
	MountPoint string  `json:"mount_point"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Available  uint64  `json:"available"`
	Usage      float64 `json:"usage"`
}
type NetworkInterface struct {
	Name            string `json:"name"`
	BytesReceived   uint64 `json:"bytes_received"`
	BytesSent       uint64 `json:"bytes_sent"`
	PacketsReceived uint64 `json:"packets_received"`
	PacketsSent     uint64 `json:"packets_sent"`
}
type ProcessInfo struct {
	PID    int     `json:"pid"`
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu"`
	Memory uint64  `json:"memory"`
	Status string  `json:"status"`
}
var systemResourcesCache struct {
	data      *SystemResources
	timestamp time.Time
	mutex     sync.RWMutex
}
const cacheTTL = 2 * time.Second
func getSystemResources() (*SystemResources, error) {
	systemResourcesCache.mutex.RLock()
	if time.Since(systemResourcesCache.timestamp) < cacheTTL && systemResourcesCache.data != nil {
		data := systemResourcesCache.data
		systemResourcesCache.mutex.RUnlock()
		return data, nil
	}
	systemResourcesCache.mutex.RUnlock()
	resources := &SystemResources{
		Timestamp: time.Now(),
	}
	cpuInfo, err := getCPUUsage()
	if err != nil {
		log.Printf("Failed to get CPU usage: %v", err)
	} else {
		resources.CPU = cpuInfo
	}
	memoryInfo, err := getMemoryInfo()
	if err != nil {
		log.Printf("Failed to get memory info: %v", err)
	} else {
		resources.Memory = memoryInfo
	}
	diskInfo, err := getDiskInfo()
	if err != nil {
		log.Printf("Failed to get disk info: %v", err)
	} else {
		resources.Disk = diskInfo
	}
	networkInfo, err := getNetworkInfo()
	if err != nil {
		log.Printf("Failed to get network info: %v", err)
	} else {
		resources.Network = networkInfo
	}
	processes, err := getTopProcesses()
	if err != nil {
		log.Printf("Failed to get top processes: %v", err)
	} else {
		resources.Processes = processes
	}
	systemResourcesCache.mutex.Lock()
	systemResourcesCache.data = resources
	systemResourcesCache.timestamp = time.Now()
	systemResourcesCache.mutex.Unlock()
	return resources, nil
}
func getCPUUsage() (CPUInfo, error) {
	cpuInfo := CPUInfo{}
	file, err := os.Open("/proc/stat")
	if err != nil {
		return cpuInfo, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return cpuInfo, fmt.Errorf("failed to read /proc/stat")
	}
	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) < 8 {
		return cpuInfo, fmt.Errorf("invalid /proc/stat format")
	}
	var user, nice, system, idle, iowait, irq, softirq, steal int64
	fmt.Sscanf(fields[1], "%d", &user)
	fmt.Sscanf(fields[2], "%d", &nice)
	fmt.Sscanf(fields[3], "%d", &system)
	fmt.Sscanf(fields[4], "%d", &idle)
	fmt.Sscanf(fields[5], "%d", &iowait)
	fmt.Sscanf(fields[6], "%d", &irq)
	fmt.Sscanf(fields[7], "%d", &softirq)
	if len(fields) > 8 {
		fmt.Sscanf(fields[8], "%d", &steal)
	}
	total := user + nice + system + idle + iowait + irq + softirq + steal
	idleTotal := idle + iowait
	used := total - idleTotal
	if total > 0 {
		cpuInfo.Usage = float64(used) / float64(total) * 100
	}
	cpuInfo.Cores = getCPUCores()
	cpuInfo.LoadAvg = getLoadAverage()
	cpuInfo.Temperature = getCPUTemperature()
	return cpuInfo, nil
}
func getCPUCores() int {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 1
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	cores := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "processor") {
			cores++
		}
	}
	if cores == 0 {
		return 1
	}
	return cores
}
func getLoadAverage() []float64 {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return []float64{0, 0, 0}
	}
	defer file.Close()
	var load1, load5, load15 float64
	fmt.Fscanf(file, "%f %f %f", &load1, &load5, &load15)
	return []float64{load1, load5, load15}
}
func getCPUTemperature() float64 {
	tempFiles := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/hwmon/hwmon0/temp1_input",
		"/sys/class/hwmon/hwmon1/temp1_input",
	}
	for _, tempFile := range tempFiles {
		if data, err := os.ReadFile(tempFile); err == nil {
			if temp, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err == nil {
				return temp / 1000.0
			}
		}
	}
	return 0
}
func getMemoryInfo() (MemoryInfo, error) {
	memInfo := MemoryInfo{}
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return memInfo, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		value *= 1024
		switch key {
		case "MemTotal":
			memInfo.Total = value
		case "MemAvailable":
			memInfo.Available = value
		case "SwapTotal":
			memInfo.SwapTotal = value
		case "SwapFree":
			memInfo.SwapUsed = memInfo.SwapTotal - value
		}
	}
	memInfo.Used = memInfo.Total - memInfo.Available
	if memInfo.Total > 0 {
		memInfo.Usage = float64(memInfo.Used) / float64(memInfo.Total) * 100
	}
	return memInfo, nil
}
func getDiskInfo() (DiskInfo, error) {
	diskInfo := DiskInfo{}
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return diskInfo, err
	}
	diskInfo.Total = uint64(stat.Blocks) * uint64(stat.Bsize)
	diskInfo.Available = uint64(stat.Bavail) * uint64(stat.Bsize)
	diskInfo.Used = diskInfo.Total - diskInfo.Available
	if diskInfo.Total > 0 {
		diskInfo.Usage = float64(diskInfo.Used) / float64(diskInfo.Total) * 100
	}
	diskInfo.MountPoints = getMountPoints()
	return diskInfo, nil
}
func getMountPoints() []MountPoint {
	var mountPoints []MountPoint
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return mountPoints
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		device := fields[0]
		mountPoint := fields[1]
		fstype := fields[2]
		if strings.HasPrefix(device, "/dev/") && fstype != "tmpfs" && fstype != "devtmpfs" {
			var stat syscall.Statfs_t
			if err := syscall.Statfs(mountPoint, &stat); err == nil {
				total := uint64(stat.Blocks) * uint64(stat.Bsize)
				available := uint64(stat.Bavail) * uint64(stat.Bsize)
				used := total - available
				usage := float64(used) / float64(total) * 100
				mountPoints = append(mountPoints, MountPoint{
					Device:     device,
					MountPoint: mountPoint,
					Total:      total,
					Used:       used,
					Available:  available,
					Usage:      usage,
				})
			}
		}
	}
	return mountPoints
}
func getNetworkInfo() (NetworkInfo, error) {
	networkInfo := NetworkInfo{}
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return networkInfo, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		interfaceName := strings.TrimSpace(parts[0])
		data := strings.Fields(parts[1])
		if len(data) < 16 {
			continue
		}
		var bytesReceived, packetsReceived, bytesSent, packetsSent uint64
		fmt.Sscanf(data[0], "%d", &bytesReceived)
		fmt.Sscanf(data[1], "%d", &packetsReceived)
		fmt.Sscanf(data[8], "%d", &bytesSent)
		fmt.Sscanf(data[9], "%d", &packetsSent)
		if interfaceName == "lo" {
			continue
		}
		networkInfo.BytesReceived += bytesReceived
		networkInfo.PacketsReceived += packetsReceived
		networkInfo.BytesSent += bytesSent
		networkInfo.PacketsSent += packetsSent
		networkInfo.Interfaces = append(networkInfo.Interfaces, NetworkInterface{
			Name:            interfaceName,
			BytesReceived:   bytesReceived,
			BytesSent:       bytesSent,
			PacketsReceived: packetsReceived,
			PacketsSent:     packetsSent,
		})
	}
	return networkInfo, nil
}
func getTopProcesses() ([]ProcessInfo, error) {
	var processes []ProcessInfo
	cmd := exec.Command("ps", "aux", "--sort=-%cpu", "--no-headers")
	output, err := cmd.Output()
	if err != nil {
		return processes, err
	}
	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if count >= 10 {
			break
		}
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}
		var pid int
		var cpu, mem float64
		var status string
		fmt.Sscanf(fields[1], "%d", &pid)
		fmt.Sscanf(fields[2], "%f", &cpu)
		fmt.Sscanf(fields[3], "%f", &mem)
		status = fields[7]
		name := fields[10]
		if len(fields) > 11 {
			name = strings.Join(fields[10:], " ")
		}
		if len(name) > 50 {
			name = name[:47] + "..."
		}
		processes = append(processes, ProcessInfo{
			PID:    pid,
			Name:   name,
			CPU:    cpu,
			Memory: uint64(mem * 1024 * 1024),
			Status: status,
		})
		count++
	}
	return processes, nil
}
func startSystemResourcesBroadcaster() {
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			resources, err := getSystemResources()
			if err == nil {
				BroadcastEvent(EventSystemResourcesUpdate, resources)
			}
		}
	}()
}