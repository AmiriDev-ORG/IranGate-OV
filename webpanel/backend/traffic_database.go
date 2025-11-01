package main
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)
type TrafficDatabase struct {
	DataDir string
}
type ClientTotalTraffic struct {
	ClientName      string    `json:"client_name"`
	TotalReceived   int64     `json:"total_received"`
	TotalSent       int64     `json:"total_sent"`
	LastSessionReceived int64 `json:"last_session_received"`
	LastSessionSent     int64 `json:"last_session_sent"`
	LastUpdate      time.Time `json:"last_update"`
}
type TrafficRecord struct {
	ClientName    string    `json:"client_name"`
	Timestamp     time.Time `json:"timestamp"`
	BytesReceived int64     `json:"bytes_received"`
	BytesSent     int64     `json:"bytes_sent"`
	TotalBytes    int64     `json:"total_bytes"`
	Connected     bool      `json:"connected"`
	RealAddress   string    `json:"real_address,omitempty"`
	VirtualIP     string    `json:"virtual_ip,omitempty"`
}
type DailyTrafficSummary struct {
	Date          string  `json:"date"`
	ClientName    string  `json:"client_name"`
	TotalReceived int64   `json:"total_received"`
	TotalSent     int64   `json:"total_sent"`
	TotalBytes    int64   `json:"total_bytes"`
	PeakHour      int     `json:"peak_hour"`
	PeakBytes     int64   `json:"peak_bytes"`
	AvgHourly     float64 `json:"avg_hourly"`
	ConnectedTime int64   `json:"connected_time_seconds"`
}
type ClientTrafficStats struct {
	ClientName      string    `json:"client_name"`
	CurrentReceived int64     `json:"current_received"`
	CurrentSent     int64     `json:"current_sent"`
	CurrentTotal    int64     `json:"current_total"`
	LastSeen        time.Time `json:"last_seen"`
	Connected       bool      `json:"connected"`
	RealAddress     string    `json:"real_address,omitempty"`
	VirtualIP       string    `json:"virtual_ip,omitempty"`
	DailyUsage      int64     `json:"daily_usage"`
	MonthlyUsage    int64     `json:"monthly_usage"`
	QuotaUsed       float64   `json:"quota_used_percent"`
}
func NewTrafficDatabase() *TrafficDatabase {
	trafficDataDir := os.Getenv("IRANGATE_TRAFFIC_DIR")
	if trafficDataDir == "" {
		if baseDir := os.Getenv("IRANGATE_DATA_DIR"); baseDir != "" {
			trafficDataDir = filepath.Join(baseDir, "traffic")
		} else {
			trafficDataDir = "/opt/irangate/traffic"
		}
	}
	os.MkdirAll(trafficDataDir, 0755)
	return &TrafficDatabase{
		DataDir: trafficDataDir,
	}
}
func (td *TrafficDatabase) StoreTrafficRecord(record TrafficRecord) error {
	dateDir := filepath.Join(td.DataDir, "daily", record.Timestamp.Format("2006-01-02"))
	os.MkdirAll(dateDir, 0755)
	filename := fmt.Sprintf("%s_%s.json",
		record.ClientName,
		record.Timestamp.Format("15-04-05"))
	filepath := filepath.Join(dateDir, filename)
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return err
	}
	td.updateCumulativeTotals(record)
	return nil
}
func (td *TrafficDatabase) updateCumulativeTotals(record TrafficRecord) {
	totalsFile := filepath.Join(td.DataDir, "cumulative_totals.json")
	totalsMap := make(map[string]ClientTotalTraffic)
	if data, err := os.ReadFile(totalsFile); err == nil {
		var totals []ClientTotalTraffic
		if err := json.Unmarshal(data, &totals); err == nil {
			for _, total := range totals {
				totalsMap[total.ClientName] = total
			}
		}
	}
	clientTotal, exists := totalsMap[record.ClientName]
	if !exists {
		clientTotal = ClientTotalTraffic{
			ClientName: record.ClientName,
		}
	}
	if record.BytesReceived < clientTotal.LastSessionReceived ||
	   record.BytesSent < clientTotal.LastSessionSent {
		clientTotal.TotalReceived += clientTotal.LastSessionReceived
		clientTotal.TotalSent += clientTotal.LastSessionSent
		clientTotal.LastSessionReceived = record.BytesReceived
		clientTotal.LastSessionSent = record.BytesSent
	} else {
		if record.BytesReceived > clientTotal.LastSessionReceived {
			clientTotal.LastSessionReceived = record.BytesReceived
		}
		if record.BytesSent > clientTotal.LastSessionSent {
			clientTotal.LastSessionSent = record.BytesSent
		}
	}
	clientTotal.LastUpdate = record.Timestamp
	totalsMap[record.ClientName] = clientTotal
	var totalsList []ClientTotalTraffic
	for _, total := range totalsMap {
		totalsList = append(totalsList, total)
	}
	if data, err := json.MarshalIndent(totalsList, "", "  "); err == nil {
		os.WriteFile(totalsFile, data, 0644)
	}
}
func (td *TrafficDatabase) getCumulativeTotals(clientName string) (int64, int64) {
	totalsFile := filepath.Join(td.DataDir, "cumulative_totals.json")
	if data, err := os.ReadFile(totalsFile); err == nil {
		var totals []ClientTotalTraffic
		if err := json.Unmarshal(data, &totals); err == nil {
			for _, total := range totals {
				if total.ClientName == clientName {
					return total.TotalReceived + total.LastSessionReceived,
					       total.TotalSent + total.LastSessionSent
				}
			}
		}
	}
	return 0, 0
}
func (td *TrafficDatabase) GetClientTrafficHistory(clientName string, startDate, endDate time.Time) ([]TrafficRecord, error) {
	var records []TrafficRecord
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		dateDir := filepath.Join(td.DataDir, "daily", d.Format("2006-01-02"))
		if _, err := os.Stat(dateDir); os.IsNotExist(err) {
			continue
		}
		files, err := filepath.Glob(filepath.Join(dateDir, clientName+"_*.json"))
		if err != nil {
			continue
		}
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			var record TrafficRecord
			if err := json.Unmarshal(data, &record); err != nil {
				continue
			}
			records = append(records, record)
		}
	}
	return records, nil
}
func (td *TrafficDatabase) GetDailySummary(clientName string, date time.Time) (*DailyTrafficSummary, error) {
	records, err := td.GetClientTrafficHistory(clientName, date, date)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return &DailyTrafficSummary{
			Date:       date.Format("2006-01-02"),
			ClientName: clientName,
		}, nil
	}
	var totalReceived, totalSent int64
	hourlyUsage := make(map[int]int64)
	var connectedTime int64
	for _, record := range records {
		totalReceived += record.BytesReceived
		totalSent += record.BytesSent
		hour := record.Timestamp.Hour()
		hourlyUsage[hour] += record.TotalBytes
		if record.Connected {
			connectedTime += 60
		}
	}
	var peakHour int
	var peakBytes int64
	for hour, bytes := range hourlyUsage {
		if bytes > peakBytes {
			peakBytes = bytes
			peakHour = hour
		}
	}
	avgHourly := float64(totalReceived+totalSent) / float64(len(hourlyUsage))
	if len(hourlyUsage) == 0 {
		avgHourly = 0
	}
	return &DailyTrafficSummary{
		Date:          date.Format("2006-01-02"),
		ClientName:    clientName,
		TotalReceived: totalReceived,
		TotalSent:     totalSent,
		TotalBytes:    totalReceived + totalSent,
		PeakHour:      peakHour,
		PeakBytes:     peakBytes,
		AvgHourly:     avgHourly,
		ConnectedTime: connectedTime,
	}, nil
}
func (td *TrafficDatabase) GetCurrentClientStats() ([]ClientTrafficStats, error) {
	statsFile := filepath.Join(td.DataDir, "current_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return []ClientTrafficStats{}, nil
	}
	data, err := os.ReadFile(statsFile)
	if err != nil {
		return nil, err
	}
	var stats []ClientTrafficStats
	err = json.Unmarshal(data, &stats)
	return stats, err
}
func (td *TrafficDatabase) UpdateCurrentStats(stats []ClientTrafficStats) error {
	statsFile := filepath.Join(td.DataDir, "current_stats.json")
	existingStats := make(map[string]ClientTrafficStats)
	if data, err := os.ReadFile(statsFile); err == nil {
		var existing []ClientTrafficStats
		if err := json.Unmarshal(data, &existing); err == nil {
			for _, stat := range existing {
				existingStats[stat.ClientName] = stat
			}
		}
	}
	currentClients := make(map[string]bool)
	for _, stat := range stats {
		currentClients[stat.ClientName] = true
	}
	mergedStats := make([]ClientTrafficStats, 0)
	for _, stat := range stats {
		cumulativeReceived, cumulativeSent := td.getCumulativeTraffic(stat.ClientName)
		stat.CurrentReceived = cumulativeReceived
		stat.CurrentSent = cumulativeSent
		stat.CurrentTotal = cumulativeReceived + cumulativeSent
		if existing, exists := existingStats[stat.ClientName]; exists {
			stat.DailyUsage = existing.DailyUsage
			stat.MonthlyUsage = existing.MonthlyUsage
		} else {
			stat.DailyUsage = td.getDailyTrafficTotal(stat.ClientName, time.Now())
			stat.MonthlyUsage = td.getMonthlyTrafficTotal(stat.ClientName, time.Now())
		}
		mergedStats = append(mergedStats, stat)
	}
	for name, existing := range existingStats {
		if !currentClients[name] {
			cumulativeReceived, cumulativeSent := td.getCumulativeTraffic(name)
			existing.CurrentReceived = cumulativeReceived
			existing.CurrentSent = cumulativeSent
			existing.CurrentTotal = cumulativeReceived + cumulativeSent
			existing.Connected = false
			existing.DailyUsage = td.getDailyTrafficTotal(name, time.Now())
			existing.MonthlyUsage = td.getMonthlyTrafficTotal(name, time.Now())
			mergedStats = append(mergedStats, existing)
		}
	}
	data, err := json.MarshalIndent(mergedStats, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statsFile, data, 0644)
}
func (td *TrafficDatabase) getCumulativeTraffic(clientName string) (int64, int64) {
	return td.getCumulativeTotals(clientName)
}
func (td *TrafficDatabase) getDailyTrafficTotal(clientName string, date time.Time) int64 {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	records, err := td.GetClientTrafficHistory(clientName, startOfDay, endOfDay)
	if err != nil {
		return 0
	}
	var maxReceived, maxSent int64
	for _, record := range records {
		if record.BytesReceived > maxReceived {
			maxReceived = record.BytesReceived
		}
		if record.BytesSent > maxSent {
			maxSent = record.BytesSent
		}
	}
	return maxReceived + maxSent
}
func (td *TrafficDatabase) getMonthlyTrafficTotal(clientName string, date time.Time) int64 {
	startOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	records, err := td.GetClientTrafficHistory(clientName, startOfMonth, endOfMonth)
	if err != nil {
		return 0
	}
	var maxReceived, maxSent int64
	for _, record := range records {
		if record.BytesReceived > maxReceived {
			maxReceived = record.BytesReceived
		}
		if record.BytesSent > maxSent {
			maxSent = record.BytesSent
		}
	}
	return maxReceived + maxSent
}
func (td *TrafficDatabase) CleanupOldData(daysToKeep int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysToKeep)
	dailyDir := filepath.Join(td.DataDir, "daily")
	entries, err := os.ReadDir(dailyDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		date, err := time.Parse("2006-01-02", entry.Name())
		if err != nil {
			continue
		}
		if date.Before(cutoffDate) {
			dirPath := filepath.Join(dailyDir, entry.Name())
			os.RemoveAll(dirPath)
		}
	}
	return nil
}