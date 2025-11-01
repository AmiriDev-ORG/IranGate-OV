package traffic
import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/amiridev-org/irangate-ov/pkg/utils"
)
type Collector struct {
	db                *database.DB
	statusLogPath     string
	managementEnabled bool
	managementAddr    string
	dataDir           string
	interval          time.Duration
	running           bool
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	lastStats         map[string]*ClientSnapshot
}
type ClientSnapshot struct {
	ClientName     string
	RealAddress    string
	VirtualIP      string
	BytesReceived  uint64
	BytesSent      uint64
	ConnectedSince time.Time
	Timestamp      time.Time
}
type TrafficRecord struct {
	ClientName     string
	Timestamp      time.Time
	BytesReceived  uint64
	BytesSent      uint64
	TotalBytes     uint64
	UploadRate     float64
	DownloadRate   float64
	Duration       time.Duration
	Connected      bool
	RealAddress    string
	VirtualIP      string
}
func NewCollector(db *database.DB, dataDir string) *Collector {
	ctx, cancel := context.WithCancel(context.Background())
	return &Collector{
		db:                db,
		statusLogPath:     "/etc/openvpn/openvpn-status.log",
		managementEnabled: false,
		dataDir:           dataDir,
		interval:          10 * time.Second,
		running:           false,
		ctx:               ctx,
		cancel:            cancel,
		lastStats:         make(map[string]*ClientSnapshot),
	}
}
func (c *Collector) SetStatusLogPath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statusLogPath = path
}
func (c *Collector) SetInterval(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.interval = interval
}
func (c *Collector) Start() error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("collector already running")
	}
	c.running = true
	c.mu.Unlock()
	if err := os.MkdirAll(c.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}
	c.wg.Add(1)
	go c.collectionLoop()
	logger := utils.GetLogger()
	logger.Info("Traffic collector started", "interval", c.interval)
	return nil
}
func (c *Collector) Stop() {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return
	}
	c.running = false
	c.mu.Unlock()
	c.cancel()
	c.wg.Wait()
	logger := utils.GetLogger()
	logger.Info("Traffic collector stopped")
}
func (c *Collector) collectionLoop() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.collect(); err != nil {
				logger := utils.GetLogger()
				logger.Errorf("Traffic collection error: %v", err)
			}
		}
	}
}
func (c *Collector) collect() error {
	c.mu.Lock()
	currentSnapshot := make(map[string]*ClientSnapshot)
	c.mu.Unlock()
	clients, err := c.parseStatusLog()
	if err != nil {
		return fmt.Errorf("failed to parse status log: %v", err)
	}
	now := time.Now()
	for _, client := range clients {
		currentSnapshot[client.ClientName] = &ClientSnapshot{
			ClientName:     client.ClientName,
			RealAddress:    client.RealAddress,
			VirtualIP:      client.VirtualIP,
			BytesReceived:  client.BytesReceived,
			BytesSent:      client.BytesSent,
			ConnectedSince: client.ConnectedSince,
			Timestamp:      now,
		}
	}
	records := c.calculateRecords(currentSnapshot)
	if err := c.storeRecords(records); err != nil {
		return fmt.Errorf("failed to store records: %v", err)
	}
	if err := c.updateClientStats(records); err != nil {
		logger := utils.GetLogger()
		logger.Warnf("Failed to update client stats: %v", err)
	}
	c.mu.Lock()
	c.lastStats = currentSnapshot
	c.mu.Unlock()
	return nil
}
func (c *Collector) parseStatusLog() ([]ClientSnapshot, error) {
	c.mu.RLock()
	path := c.statusLogPath
	c.mu.RUnlock()
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []ClientSnapshot{}, nil
		}
		return nil, err
	}
	defer file.Close()
	var clients []ClientSnapshot
	scanner := bufio.NewScanner(file)
	inClientList := false
	inRoutingTable := false
	routingTable := make(map[string]string)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "OpenVPN CLIENT LIST":
			inClientList = true
			inRoutingTable = false
		case line == "ROUTING TABLE":
			inClientList = false
			inRoutingTable = true
		case line == "GLOBAL STATS" || line == "END":
			inClientList = false
			inRoutingTable = false
		case strings.HasPrefix(line, "Updated,") || strings.HasPrefix(line, "Common Name,"):
		default:
			if inClientList {
				if client := parseClientLine(line); client != nil {
					clients = append(clients, *client)
				}
			} else if inRoutingTable {
				if vpnIP, cn := parseRoutingLine(line); vpnIP != "" && cn != "" {
					routingTable[cn] = vpnIP
				}
			}
		}
	}
	for i := range clients {
		if vpnIP, exists := routingTable[clients[i].ClientName]; exists {
			clients[i].VirtualIP = vpnIP
		}
	}
	return clients, scanner.Err()
}
func parseClientLine(line string) *ClientSnapshot {
	parts := strings.Split(line, ",")
	if len(parts) < 5 {
		parts = strings.Fields(line)
		if len(parts) < 5 {
			return nil
		}
	}
	clientName := strings.TrimSpace(parts[0])
	if clientName == "" {
		return nil
	}
	realAddr := strings.TrimSpace(parts[1])
	if idx := strings.Index(realAddr, ":"); idx != -1 {
		realAddr = realAddr[:idx]
	}
	bytesReceived, _ := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64)
	bytesSent, _ := strconv.ParseUint(strings.TrimSpace(parts[3]), 10, 64)
	connectedSince := time.Time{}
	if len(parts) >= 5 {
		if t, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(parts[4])); err == nil {
			connectedSince = t
		}
	}
	return &ClientSnapshot{
		ClientName:     clientName,
		RealAddress:    realAddr,
		BytesReceived:  bytesReceived,
		BytesSent:      bytesSent,
		ConnectedSince: connectedSince,
		Timestamp:      time.Now(),
	}
}
func parseRoutingLine(line string) (vpnIP, commonName string) {
	parts := strings.Split(line, ",")
	if len(parts) < 2 {
		parts = strings.Fields(line)
		if len(parts) < 2 {
			return "", ""
		}
	}
	vpnIP = strings.TrimSpace(parts[0])
	commonName = strings.TrimSpace(parts[1])
	return vpnIP, commonName
}
func (c *Collector) calculateRecords(current map[string]*ClientSnapshot) []TrafficRecord {
	records := make([]TrafficRecord, 0, len(current))
	c.mu.RLock()
	last := c.lastStats
	c.mu.RUnlock()
	for clientName, currentSnap := range current {
		record := TrafficRecord{
			ClientName:    clientName,
			Timestamp:     currentSnap.Timestamp,
			BytesReceived: currentSnap.BytesReceived,
			BytesSent:     currentSnap.BytesSent,
			TotalBytes:    currentSnap.BytesReceived + currentSnap.BytesSent,
			Connected:     true,
			RealAddress:   currentSnap.RealAddress,
			VirtualIP:     currentSnap.VirtualIP,
		}
		if lastSnap, exists := last[clientName]; exists {
			deltaTime := currentSnap.Timestamp.Sub(lastSnap.Timestamp).Seconds()
			if deltaTime > 0 {
				record.UploadRate = float64(currentSnap.BytesSent-lastSnap.BytesSent) / deltaTime
				record.DownloadRate = float64(currentSnap.BytesReceived-lastSnap.BytesReceived) / deltaTime
			}
			if !currentSnap.ConnectedSince.IsZero() && !lastSnap.ConnectedSince.IsZero() {
				record.Duration = currentSnap.Timestamp.Sub(lastSnap.Timestamp)
			}
		}
		records = append(records, record)
	}
	for clientName, lastSnap := range last {
		if _, exists := current[clientName]; !exists {
			records = append(records, TrafficRecord{
				ClientName:   clientName,
				Timestamp:    time.Now(),
				Connected:    false,
				RealAddress:  lastSnap.RealAddress,
				VirtualIP:    lastSnap.VirtualIP,
				UploadRate:   0,
				DownloadRate: 0,
			})
		}
	}
	return records
}
func (c *Collector) storeRecords(records []TrafficRecord) error {
	clientRecords := make(map[string][]TrafficRecord)
	for _, record := range records {
		clientRecords[record.ClientName] = append(clientRecords[record.ClientName], record)
	}
	for clientName, clientRecs := range clientRecords {
		filePath := c.getRecordFilePath(clientName, time.Now(), "hourly")
		existingRecs, _ := c.readRecordsFromFile(filePath)
		allRecs := append(existingRecs, clientRecs...)
		if err := c.writeRecordsToFile(filePath, allRecs); err != nil {
			return fmt.Errorf("failed to write records for %s: %v", clientName, err)
		}
	}
	return nil
}
func (c *Collector) getRecordFilePath(clientName string, t time.Time, granularity string) string {
	var filename string
	switch granularity {
	case "hourly":
		filename = fmt.Sprintf("%s_%s_hourly.json", clientName, t.Format("2006-01-02_15"))
	case "daily":
		filename = fmt.Sprintf("%s_%s_daily.json", clientName, t.Format("2006-01-02"))
	default:
		filename = fmt.Sprintf("%s_%s.json", clientName, t.Format("20060102150405"))
	}
	return fmt.Sprintf("%s/records/%s/%s", c.dataDir, granularity, filename)
}
func (c *Collector) readRecordsFromFile(path string) ([]TrafficRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []TrafficRecord{}, nil
		}
		return nil, err
	}
	var records []TrafficRecord
	if err := json.Unmarshal(data, &records); err != nil {
		var single TrafficRecord
		if err := json.Unmarshal(data, &single); err != nil {
			return nil, err
		}
		return []TrafficRecord{single}, nil
	}
	return records, nil
}
func (c *Collector) writeRecordsToFile(path string, records []TrafficRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
func (c *Collector) updateClientStats(records []TrafficRecord) error {
	clientTotals := make(map[string]struct {
		received uint64
		sent     uint64
	})
	for _, record := range records {
		totals := clientTotals[record.ClientName]
		totals.received += record.BytesReceived
		totals.sent += record.BytesSent
		clientTotals[record.ClientName] = totals
	}
	for clientName, totals := range clientTotals {
		client, err := c.db.GetClient(clientName)
		if err != nil {
			continue
		}
	client.BytesReceived += totals.received
	client.BytesSent += totals.sent
	client.Download += int64(totals.received)
	client.Upload += int64(totals.sent)
	client.DataUsedMB = int64((client.BytesReceived + client.BytesSent) / 1024 / 1024)
	client.LastUpdate = time.Now()
		if err := c.db.UpdateClient(clientName, client); err != nil {
			return fmt.Errorf("failed to update client %s: %v", clientName, err)
		}
	}
	return nil
}
func (c *Collector) GetCurrentStats() (map[string]*ClientSnapshot, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]*ClientSnapshot)
	for k, v := range c.lastStats {
		snapshot := *v
		result[k] = &snapshot
	}
	return result, nil
}
func (c *Collector) GetClientStats(clientName string) (*ClientSnapshot, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if stats, exists := c.lastStats[clientName]; exists {
		snapshot := *stats
		return &snapshot, nil
	}
	return nil, fmt.Errorf("client not found: %s", clientName)
}