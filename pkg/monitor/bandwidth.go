package monitor
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
)
type BandwidthMonitor struct {
	db            *database.DB
	dataFile      string
	interval      time.Duration
	stats         map[string]*ClientStats
	mu            sync.RWMutex
	stopChan      chan struct{}
	retentionDays int
}
type ClientStats struct {
	ClientName     string     `json:"client_name"`
	BytesIn        int64      `json:"bytes_in"`
	BytesOut       int64      `json:"bytes_out"`
	BytesInRate    int64      `json:"bytes_in_rate"`
	BytesOutRate   int64      `json:"bytes_out_rate"`
	LastUpdate     time.Time  `json:"last_update"`
	ConnectedSince *time.Time `json:"connected_since,omitempty"`
	Status         string     `json:"status"`
}
type ConnectionHistory struct {
	ClientName       string     `json:"client_name"`
	SessionID        string     `json:"session_id"`
	StartTime        time.Time  `json:"start_time"`
	EndTime          *time.Time `json:"end_time,omitempty"`
	BytesIn          int64      `json:"bytes_in"`
	BytesOut         int64      `json:"bytes_out"`
	IPAddress        string     `json:"ip_address"`
	DisconnectReason string     `json:"disconnect_reason,omitempty"`
}
func NewBandwidthMonitor(db *database.DB, dataDir string) (*BandwidthMonitor, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}
	return &BandwidthMonitor{
		db:            db,
		dataFile:      filepath.Join(dataDir, "bandwidth_stats.json"),
		interval:      5 * time.Second,
		stats:         make(map[string]*ClientStats),
		stopChan:      make(chan struct{}),
		retentionDays: 30,
	}, nil
}
func (m *BandwidthMonitor) Start() error {
	if err := m.loadStats(); err != nil {
		return err
	}
	ticker := time.NewTicker(m.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.updateStats()
			case <-m.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
	go m.periodicCleanup()
	return nil
}
func (m *BandwidthMonitor) Stop() {
	close(m.stopChan)
	m.saveStats()
}
func (m *BandwidthMonitor) updateStats() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveStats()
}
func (m *BandwidthMonitor) recordDisconnection(clientName string, stat *ClientStats) {
	history := &ConnectionHistory{
		ClientName: clientName,
		SessionID:  fmt.Sprintf("session_%d", time.Now().Unix()),
		StartTime:  *stat.ConnectedSince,
		EndTime:    &stat.LastUpdate,
		BytesIn:    stat.BytesIn,
		BytesOut:   stat.BytesOut,
	}
	m.saveConnectionHistory(history)
}
func (m *BandwidthMonitor) saveConnectionHistory(history *ConnectionHistory) {
	historyFile := filepath.Join(filepath.Dir(m.dataFile),
		fmt.Sprintf("history_%s.json", history.ClientName))
	var histories []*ConnectionHistory
	if data, err := os.ReadFile(historyFile); err == nil {
		json.Unmarshal(data, &histories)
	}
	histories = append(histories, history)
	data, _ := json.MarshalIndent(histories, "", "  ")
	os.WriteFile(historyFile, data, 0644)
}
func (m *BandwidthMonitor) loadStats() error {
	data, err := os.ReadFile(m.dataFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &m.stats)
}
func (m *BandwidthMonitor) saveStats() error {
	data, err := json.MarshalIndent(m.stats, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.dataFile, data, 0644)
}
func (m *BandwidthMonitor) periodicCleanup() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.cleanupOldHistory()
		case <-m.stopChan:
			return
		}
	}
}
func (m *BandwidthMonitor) cleanupOldHistory() {
	cutoff := time.Now().AddDate(0, 0, -m.retentionDays)
	files, err := filepath.Glob(filepath.Join(filepath.Dir(m.dataFile), "history_*.json"))
	if err != nil {
		return
	}
	for _, file := range files {
		var histories []*ConnectionHistory
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &histories); err != nil {
			continue
		}
		filtered := make([]*ConnectionHistory, 0)
		for _, h := range histories {
			if h.StartTime.After(cutoff) {
				filtered = append(filtered, h)
			}
		}
		if len(filtered) > 0 {
			data, _ = json.MarshalIndent(filtered, "", "  ")
			os.WriteFile(file, data, 0644)
		} else {
			os.Remove(file)
		}
	}
}
func (m *BandwidthMonitor) GetClientStats(clientName string) (*ClientStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	stat, exists := m.stats[clientName]
	if !exists {
		return nil, fmt.Errorf("no stats found for client: %s", clientName)
	}
	return stat, nil
}
func (m *BandwidthMonitor) GetConnectionHistory(clientName string) ([]*ConnectionHistory, error) {
	historyFile := filepath.Join(filepath.Dir(m.dataFile),
		fmt.Sprintf("history_%s.json", clientName))
	data, err := os.ReadFile(historyFile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var histories []*ConnectionHistory
	if err := json.Unmarshal(data, &histories); err != nil {
		return nil, err
	}
	return histories, nil
}