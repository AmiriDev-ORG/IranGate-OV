package traffic
import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)
type Analyzer struct {
	dataDir string
}
func NewAnalyzer(dataDir string) *Analyzer {
	return &Analyzer{
		dataDir: dataDir,
	}
}
type AggregatedStats struct {
	ClientName     string
	Period         string
	StartTime      time.Time
	EndTime        time.Time
	TotalReceived  uint64
	TotalSent      uint64
	TotalBytes     uint64
	AvgUploadRate  float64
	AvgDownloadRate float64
	PeakUploadRate float64
	PeakDownloadRate float64
	ConnectedTime  float64
	NumSessions    int
}
type TopClient struct {
	ClientName string
	TotalBytes uint64
	Received   uint64
	Sent       uint64
	Period     string
}
type UsagePattern struct {
	Period     string
	DataPoints []UsageDataPoint
}
type UsageDataPoint struct {
	Timestamp   time.Time
	TotalBytes  uint64
	Received    uint64
	Sent        uint64
	UploadRate  float64
	DownloadRate float64
}
type QuotaUsage struct {
	ClientName      string
	TotalUsed       uint64
	QuotaLimit      uint64
	Percentage      float64
	RemainingBytes  uint64
	ResetDate       time.Time
}
func (a *Analyzer) GetClientAggregatedStats(clientName string, startTime, endTime time.Time, granularity string) (*AggregatedStats, error) {
	records, err := a.loadRecordsForPeriod(clientName, startTime, endTime)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return &AggregatedStats{
			ClientName: clientName,
			Period:     granularity,
			StartTime:  startTime,
			EndTime:    endTime,
		}, nil
	}
	stats := &AggregatedStats{
		ClientName: clientName,
		Period:     granularity,
		StartTime:  startTime,
		EndTime:    endTime,
	}
	var totalUploadRate, totalDownloadRate float64
	sessionSet := make(map[time.Time]bool)
	for _, record := range records {
		stats.TotalReceived += record.BytesReceived
		stats.TotalSent += record.BytesSent
		totalUploadRate += record.UploadRate
		totalDownloadRate += record.DownloadRate
		if stats.PeakUploadRate < record.UploadRate {
			stats.PeakUploadRate = record.UploadRate
		}
		if stats.PeakDownloadRate < record.DownloadRate {
			stats.PeakDownloadRate = record.DownloadRate
		}
		if record.Connected {
			stats.ConnectedTime += record.Duration.Seconds()
			sessionHour := record.Timestamp.Truncate(time.Hour)
			if !sessionSet[sessionHour] {
				sessionSet[sessionHour] = true
				stats.NumSessions++
			}
		}
	}
	stats.TotalBytes = stats.TotalReceived + stats.TotalSent
	if len(records) > 0 {
		stats.AvgUploadRate = totalUploadRate / float64(len(records))
		stats.AvgDownloadRate = totalDownloadRate / float64(len(records))
	}
	return stats, nil
}
func (a *Analyzer) GetTopUsers(limit int, startTime, endTime time.Time) ([]TopClient, error) {
	clients, err := a.getAllClients()
	if err != nil {
		return nil, err
	}
	var topClients []TopClient
	for _, clientName := range clients {
		stats, err := a.GetClientAggregatedStats(clientName, startTime, endTime, "custom")
		if err != nil {
			continue
		}
		if stats.TotalBytes > 0 {
			topClients = append(topClients, TopClient{
				ClientName: clientName,
				TotalBytes: stats.TotalBytes,
				Received:   stats.TotalReceived,
				Sent:       stats.TotalSent,
			})
		}
	}
	sort.Slice(topClients, func(i, j int) bool {
		return topClients[i].TotalBytes > topClients[j].TotalBytes
	})
	if limit > 0 && len(topClients) > limit {
		topClients = topClients[:limit]
	}
	return topClients, nil
}
func (a *Analyzer) GetUsagePattern(clientName string, startTime, endTime time.Time, interval time.Duration) (*UsagePattern, error) {
	records, err := a.loadRecordsForPeriod(clientName, startTime, endTime)
	if err != nil {
		return nil, err
	}
	pattern := &UsagePattern{
		Period:     interval.String(),
		DataPoints: make([]UsageDataPoint, 0),
	}
	intervalMap := make(map[time.Time]*UsageDataPoint)
	for _, record := range records {
		intervalStart := record.Timestamp.Truncate(interval)
		point, exists := intervalMap[intervalStart]
		if !exists {
			point = &UsageDataPoint{
				Timestamp: intervalStart,
			}
			intervalMap[intervalStart] = point
		}
		point.TotalBytes += record.TotalBytes
		point.Received += record.BytesReceived
		point.Sent += record.BytesSent
		point.UploadRate = (point.UploadRate + record.UploadRate) / 2
		point.DownloadRate = (point.DownloadRate + record.DownloadRate) / 2
	}
	for _, point := range intervalMap {
		pattern.DataPoints = append(pattern.DataPoints, *point)
	}
	sort.Slice(pattern.DataPoints, func(i, j int) bool {
		return pattern.DataPoints[i].Timestamp.Before(pattern.DataPoints[j].Timestamp)
	})
	return pattern, nil
}
func (a *Analyzer) GetQuotaUsage(clientName string, quotaLimit uint64, resetDate time.Time, periodStart, periodEnd time.Time) (*QuotaUsage, error) {
	stats, err := a.GetClientAggregatedStats(clientName, periodStart, periodEnd, "period")
	if err != nil {
		return nil, err
	}
	percentage := 0.0
	remainingBytes := uint64(0)
	if quotaLimit > 0 {
		percentage = (float64(stats.TotalBytes) / float64(quotaLimit)) * 100.0
		if stats.TotalBytes < quotaLimit {
			remainingBytes = quotaLimit - stats.TotalBytes
		}
	}
	return &QuotaUsage{
		ClientName:     clientName,
		TotalUsed:      stats.TotalBytes,
		QuotaLimit:     quotaLimit,
		Percentage:     percentage,
		RemainingBytes: remainingBytes,
		ResetDate:      resetDate,
	}, nil
}
func (a *Analyzer) GetBandwidthEfficiency(clientName string, startTime, endTime time.Time) (map[string]interface{}, error) {
	stats, err := a.GetClientAggregatedStats(clientName, startTime, endTime, "custom")
	if err != nil {
		return nil, err
	}
	efficiency := make(map[string]interface{})
	efficiency["total_bytes"] = stats.TotalBytes
	efficiency["total_received"] = stats.TotalReceived
	efficiency["total_sent"] = stats.TotalSent
	efficiency["upload_ratio"] = 0.0
	efficiency["download_ratio"] = 0.0
	if stats.TotalBytes > 0 {
		efficiency["upload_ratio"] = float64(stats.TotalSent) / float64(stats.TotalBytes)
		efficiency["download_ratio"] = float64(stats.TotalReceived) / float64(stats.TotalBytes)
	}
	efficiency["avg_upload_rate_mbps"] = (stats.AvgUploadRate * 8) / 1000000
	efficiency["avg_download_rate_mbps"] = (stats.AvgDownloadRate * 8) / 1000000
	efficiency["peak_upload_rate_mbps"] = (stats.PeakUploadRate * 8) / 1000000
	efficiency["peak_download_rate_mbps"] = (stats.PeakDownloadRate * 8) / 1000000
	efficiency["connection_time_hours"] = stats.ConnectedTime / 3600
	efficiency["num_sessions"] = stats.NumSessions
	if stats.NumSessions > 0 {
		efficiency["avg_session_duration_hours"] = stats.ConnectedTime / float64(stats.NumSessions) / 3600
		efficiency["bytes_per_session"] = float64(stats.TotalBytes) / float64(stats.NumSessions)
	}
	return efficiency, nil
}
func (a *Analyzer) DetectAnomalies(clientName string, startTime, endTime time.Time) ([]string, error) {
	var anomalies []string
	stats, err := a.GetClientAggregatedStats(clientName, startTime, endTime, "custom")
	if err != nil {
		return nil, err
	}
	baselineStart := startTime.AddDate(0, 0, -30)
	baselineStats, err := a.GetClientAggregatedStats(clientName, baselineStart, startTime, "custom")
	if err != nil {
		return anomalies, nil
	}
	if baselineStats.ConnectedTime > 0 {
		baselineAvgBytesPerHour := float64(baselineStats.TotalBytes) / (baselineStats.ConnectedTime / 3600)
		currentDuration := endTime.Sub(startTime).Hours()
		if currentDuration > 0 {
			currentAvgBytesPerHour := float64(stats.TotalBytes) / currentDuration
			if currentAvgBytesPerHour > baselineAvgBytesPerHour*3 {
				anomalies = append(anomalies, fmt.Sprintf("Traffic usage is %.1fx higher than historical average", currentAvgBytesPerHour/baselineAvgBytesPerHour))
			}
		}
	}
	if stats.AvgUploadRate > 50*1024*1024 {
		anomalies = append(anomalies, fmt.Sprintf("Sustained high upload rate: %.2f MB/s", stats.AvgUploadRate/1024/1024))
	}
	if stats.ConnectedTime > 24*3600 {
		anomalies = append(anomalies, fmt.Sprintf("Unusually long connection duration: %.1f hours", stats.ConnectedTime/3600))
	}
	return anomalies, nil
}
func (a *Analyzer) loadRecordsForPeriod(clientName string, startTime, endTime time.Time) ([]TrafficRecord, error) {
	var allRecords []TrafficRecord
	for d := startTime; d.Before(endTime) || d.Equal(endTime); d = d.AddDate(0, 0, 1) {
		for hour := 0; hour < 24; hour++ {
			hourTime := time.Date(d.Year(), d.Month(), d.Day(), hour, 0, 0, 0, d.Location())
			if hourTime.Before(startTime) || hourTime.After(endTime) {
				continue
			}
			filePath := getRecordFilePath(a.dataDir, clientName, hourTime, "hourly")
			records, err := readRecordsFromFile(filePath)
			if err == nil {
				for _, record := range records {
					if record.Timestamp.After(startTime) && (record.Timestamp.Before(endTime) || record.Timestamp.Equal(endTime)) {
						allRecords = append(allRecords, record)
					}
				}
			}
		}
	}
	return allRecords, nil
}
func (a *Analyzer) getAllClients() ([]string, error) {
	recordsDir := fmt.Sprintf("%s/records/hourly", a.dataDir)
	entries, err := os.ReadDir(recordsDir)
	if err != nil {
		return nil, err
	}
	clientSet := make(map[string]bool)
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			parts := strings.Split(name, "_")
			if len(parts) >= 1 {
				clientSet[parts[0]] = true
			}
		}
	}
	clients := make([]string, 0, len(clientSet))
	for client := range clientSet {
		clients = append(clients, client)
	}
	return clients, nil
}
func getRecordFilePath(dataDir, clientName string, t time.Time, granularity string) string {
	var filename string
	switch granularity {
	case "hourly":
		filename = fmt.Sprintf("%s_%s_hourly.json", clientName, t.Format("2006-01-02_15"))
	case "daily":
		filename = fmt.Sprintf("%s_%s_daily.json", clientName, t.Format("2006-01-02"))
	default:
		filename = fmt.Sprintf("%s_%s.json", clientName, t.Format("20060102150405"))
	}
	return fmt.Sprintf("%s/records/%s/%s", dataDir, granularity, filename)
}
func readRecordsFromFile(path string) ([]TrafficRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var records []TrafficRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}