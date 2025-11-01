package monitor
import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/amiridev-org/irangate-ov/pkg/utils"
)
func (m *Monitor) cleanupOldClients() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-ClientDataRetention)
	for commonName, client := range m.clients {
		if client.ConnectedSince.Before(cutoff) {
			delete(m.clients, commonName)
		}
	}
	if len(m.clients) > MaxStoredClients {
		type clientAge struct {
			name string
			age  time.Time
		}
		ages := make([]clientAge, 0, len(m.clients))
		for name, client := range m.clients {
			ages = append(ages, clientAge{name, client.ConnectedSince})
		}
		sort.Slice(ages, func(i, j int) bool {
			return ages[i].age.Before(ages[j].age)
		})
		for i := 0; i < len(ages)-MaxStoredClients; i++ {
			delete(m.clients, ages[i].name)
		}
	}
	m.lastCleanup = now
}
func (m *Monitor) StartMonitoring(interval time.Duration) error {
	logger := utils.GetLogger()
	logger.Info("Starting OpenVPN monitoring...")
	ticker := time.NewTicker(interval)
	cleanupTicker := time.NewTicker(ClientCleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := m.updateStats(); err != nil {
					logger.Errorf("Failed to update stats: %v", err)
				}
			case <-cleanupTicker.C:
				m.cleanupOldClients()
			}
		}
	}()
	return nil
}
const (
	StatusLogPath = "/etc/openvpn/openvpn-status.log"
	LogPath       = "/var/log/openvpn.log"
	MaxStoredClients = 500
	ClientCleanupInterval = 1 * time.Hour
	ClientDataRetention = 24 * time.Hour
)
type ConnectedClient struct {
	CommonName     string    `json:"common_name"`
	RealAddress    string    `json:"real_address"`
	VirtualAddress string    `json:"virtual_address"`
	BytesReceived  uint64    `json:"bytes_received"`
	BytesSent      uint64    `json:"bytes_sent"`
	ConnectedSince time.Time `json:"connected_since"`
}
type ServerStats struct {
	Uptime           time.Duration `json:"uptime"`
	ConnectedClients int           `json:"connected_clients"`
	TotalTrafficMB   float64       `json:"total_traffic_mb"`
	CPUUsage         float64       `json:"cpu_usage"`
	MemoryUsageMB    float64       `json:"memory_usage_mb"`
	LastUpdate       time.Time     `json:"last_update"`
}
type Monitor struct {
	db            *database.DB
	mu            sync.RWMutex
	clients       map[string]*ConnectedClient
	stats         *ServerStats
	lastCleanup   time.Time
	statusLogPath string
}
func New(db *database.DB) *Monitor {
	return &Monitor{
		db:            db,
		clients:       make(map[string]*ConnectedClient),
		stats:         &ServerStats{},
		lastCleanup:   time.Now(),
		statusLogPath: StatusLogPath,
	}
}
func (m *Monitor) GetClients() ([]*ConnectedClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	clients := make([]*ConnectedClient, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	return clients, nil
}
func (m *Monitor) GetStats() (*ServerStats, error) {
	if err := m.updateStats(); err != nil {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return m.stats, nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats, nil
}
func (m *Monitor) GetClientTraffic(commonName string) (*ConnectedClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, ok := m.clients[commonName]
	if !ok {
		return nil, fmt.Errorf("client not found: %s", commonName)
	}
	return client, nil
}
func (m *Monitor) GetLogs(n int) ([]string, error) {
	file, err := os.Open(LogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}
	defer file.Close()
	logs := make([]string, 0, n)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		logs = append(logs, scanner.Text())
		if len(logs) > n {
			logs = logs[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read logs: %v", err)
	}
	return logs, nil
}
func (m *Monitor) UpdateClientStats(commonName string, bytesReceived, bytesSent uint64) error {
	clients, err := m.db.GetClients()
	if err != nil {
		return fmt.Errorf("failed to get clients: %v", err)
	}
	for _, client := range clients {
		if client.Name == commonName {
			client.DataUsedMB = int64((bytesReceived + bytesSent) / 1024 / 1024)
			client.BytesReceived += bytesReceived
			client.BytesSent += bytesSent
			client.LastUpdate = time.Now()
			if err := m.db.UpdateClient(commonName, &client); err != nil {
				return fmt.Errorf("failed to update client in database: %v", err)
			}
			return nil
		}
	}
	return fmt.Errorf("client not found: %s", commonName)
}
func (m *Monitor) ExportStats(w io.Writer) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data := struct {
		Timestamp time.Time          `json:"timestamp"`
		Stats     *ServerStats       `json:"stats"`
		Clients   []*ConnectedClient `json:"clients"`
	}{
		Timestamp: time.Now(),
		Stats:     m.stats,
		Clients:   make([]*ConnectedClient, 0, len(m.clients)),
	}
	for _, client := range m.clients {
		data.Clients = append(data.Clients, client)
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
func (m *Monitor) updateStats() error {
	file, err := os.Open(m.statusLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.mu.Lock()
			defer m.mu.Unlock()
			m.clients = make(map[string]*ConnectedClient)
			m.stats = &ServerStats{
				ConnectedClients: 0,
				TotalTrafficMB:   0,
				LastUpdate:       time.Now(),
			}
			return nil
		}
		return fmt.Errorf("failed to open status log: %v", err)
	}
	defer file.Close()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients = make(map[string]*ConnectedClient)
	scanner := bufio.NewScanner(file)
	section := ""
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "OpenVPN CLIENT LIST"):
			section = "clients"
			scanner.Scan()
		case strings.HasPrefix(line, "ROUTING TABLE"):
			section = "routing"
			scanner.Scan()
		case strings.HasPrefix(line, "GLOBAL STATS"):
			section = "stats"
		case line == "":
			section = ""
		case strings.HasPrefix(line, "Updated,") || strings.HasPrefix(line, "Common Name,") || strings.HasPrefix(line, "Virtual Address,"):
			continue
		default:
			switch section {
			case "clients":
				m.parseClientLine(line)
			case "routing":
				m.parseRoutingLine(line)
			case "stats":
				m.parseStatsLine(line)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read status log: %v", err)
	}
	m.updateServerStats()
	return nil
}
func (m *Monitor) parseClientLine(line string) {
	var fields []string
	if strings.Contains(line, ",") {
		fields = strings.Split(line, ",")
	} else {
		fields = strings.Fields(line)
	}
	if len(fields) < 5 {
		return
	}
	commonName := strings.TrimSpace(fields[0])
	realAddress := strings.TrimSpace(strings.Split(fields[1], ":")[0])
	bytesReceived, _ := strconv.ParseUint(strings.TrimSpace(fields[2]), 10, 64)
	bytesSent, _ := strconv.ParseUint(strings.TrimSpace(fields[3]), 10, 64)
	connectedSince, _ := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(fields[4]))
	m.clients[commonName] = &ConnectedClient{
		CommonName:     commonName,
		RealAddress:    realAddress,
		BytesReceived:  bytesReceived,
		BytesSent:      bytesSent,
		ConnectedSince: connectedSince,
	}
}
func (m *Monitor) parseRoutingLine(line string) {
	var fields []string
	if strings.Contains(line, ",") {
		fields = strings.Split(line, ",")
	} else {
		fields = strings.Fields(line)
	}
	if len(fields) < 3 {
		return
	}
	virtualAddress := strings.TrimSpace(fields[0])
	commonName := strings.TrimSpace(fields[1])
	if client, ok := m.clients[commonName]; ok {
		client.VirtualAddress = virtualAddress
	}
}
func (m *Monitor) parseStatsLine(line string) {
	re := regexp.MustCompile(`^(\w+),(\d+)$`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 3 {
		return
	}
	key := matches[1]
	value, _ := strconv.ParseUint(matches[2], 10, 64)
	switch key {
	case "Max bcast/mcast queue length":
		_ = value
	}
}
func (m *Monitor) updateServerStats() {
	var totalBytes uint64
	for _, client := range m.clients {
		totalBytes += client.BytesReceived + client.BytesSent
	}
	m.stats = &ServerStats{
		ConnectedClients: len(m.clients),
		TotalTrafficMB:   float64(totalBytes) / 1024 / 1024,
		CPUUsage:         getSystemCPUUsage(),
		MemoryUsageMB:    getSystemMemoryUsage(),
	}
	if uptimeStr, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(uptimeStr))
		if uptime, err := strconv.ParseFloat(fields[0], 64); err == nil {
			m.stats.Uptime = time.Duration(uptime * float64(time.Second))
		}
	}
}
func getSystemCPUUsage() float64 {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0.0
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0.0
	}
	fields := strings.Fields(lines[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0.0
	}
	user, _ := strconv.ParseFloat(fields[1], 64)
	system, _ := strconv.ParseFloat(fields[3], 64)
	idle, _ := strconv.ParseFloat(fields[4], 64)
	total := user + system + idle
	if total == 0 {
		return 0.0
	}
	used := user + system
	return (used / total) * 100.0
}
func getSystemMemoryUsage() float64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0.0
	}
	var totalMB, availableMB float64
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, _ := strconv.ParseFloat(fields[1], 64)
		switch fields[0] {
		case "MemTotal:":
			totalMB = value / 1024
		case "MemAvailable:":
			availableMB = value / 1024
		}
	}
	if totalMB == 0 {
		return 0.0
	}
	usedMB := totalMB - availableMB
	return usedMB
}