package traffic
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)
type Storage struct {
	dataDir        string
	retentionDays  int
	rollupInterval time.Duration
	running        bool
	stopChan       chan struct{}
}
func NewStorage(dataDir string) *Storage {
	return &Storage{
		dataDir:        dataDir,
		retentionDays:  90,
		rollupInterval: 1 * time.Hour,
		stopChan:       make(chan struct{}),
	}
}
func (s *Storage) SetRetentionDays(days int) {
	s.retentionDays = days
}
func (s *Storage) Start() error {
	if s.running {
		return fmt.Errorf("storage already running")
	}
	s.running = true
	go s.rollupWorker()
	go s.cleanupWorker()
	return nil
}
func (s *Storage) Stop() {
	if !s.running {
		return
	}
	close(s.stopChan)
	s.running = false
}
func (s *Storage) rollupWorker() {
	ticker := time.NewTicker(s.rollupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.performRollup(); err != nil {
				fmt.Printf("Error performing rollup: %v\n", err)
			}
		}
	}
}
func (s *Storage) cleanupWorker() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.cleanupOldData(); err != nil {
				fmt.Printf("Error cleaning up old data: %v\n", err)
			}
		}
	}
}
func (s *Storage) performRollup() error {
	hourlyDir := filepath.Join(s.dataDir, "records", "hourly")
	entries, err := os.ReadDir(hourlyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	currentHour := time.Now().Truncate(time.Hour)
	clientFiles := make(map[string][]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(hourlyDir, entry.Name())
		if timestamp, err := s.extractTimestamp(entry.Name(), "hourly"); err != nil {
			continue
		} else if !timestamp.Before(currentHour) {
			continue
		}
		clientName := s.extractClientName(entry.Name())
		if clientName != "" {
			clientFiles[clientName] = append(clientFiles[clientName], filePath)
		}
	}
	for clientName, files := range clientFiles {
		if err := s.rollupClientData(clientName, files); err != nil {
			fmt.Printf("Error rolling up data for %s: %v\n", clientName, err)
		}
	}
	return nil
}
func (s *Storage) rollupClientData(clientName string, files []string) error {
	dailyGroups := make(map[string][]string)
	for _, file := range files {
		if timestamp, err := s.extractTimestampFromPath(file, "hourly"); err == nil {
			dayKey := timestamp.Format("2006-01-02")
			dailyGroups[dayKey] = append(dailyGroups[dayKey], file)
		}
	}
	for dayKey, dayFiles := range dailyGroups {
		dailyPath := s.getRollupPath(clientName, dayKey, "daily")
		if _, err := os.Stat(dailyPath); err == nil {
			continue
		}
		var allRecords []TrafficRecord
		for _, file := range dayFiles {
			records, err := s.readRecords(file)
			if err != nil {
				continue
			}
			allRecords = append(allRecords, records...)
		}
		if err := s.createDailyRollup(clientName, dayKey, allRecords); err != nil {
			return err
		}
	}
	return nil
}
func (s *Storage) createDailyRollup(clientName, dayKey string, records []TrafficRecord) error {
	if len(records) == 0 {
		return nil
	}
	summary := s.calculateDailySummary(clientName, dayKey, records)
	dailyPath := s.getRollupPath(clientName, dayKey, "daily")
	if err := s.writeDailySummary(dailyPath, summary); err != nil {
		return err
	}
	return nil
}
func (s *Storage) calculateDailySummary(clientName, dayKey string, records []TrafficRecord) *DailySummary {
	summary := &DailySummary{
		ClientName:   clientName,
		Date:         dayKey,
		RecordCount:  len(records),
		HourlyData:   make(map[int]*HourlySummary),
	}
	var totalReceived, totalSent uint64
	var totalUploadRate, totalDownloadRate float64
	for _, record := range records {
		totalReceived += record.BytesReceived
		totalSent += record.BytesSent
		totalUploadRate += record.UploadRate
		totalDownloadRate += record.DownloadRate
		hour := record.Timestamp.Hour()
		if summary.HourlyData[hour] == nil {
			summary.HourlyData[hour] = &HourlySummary{Hour: hour}
		}
		summary.HourlyData[hour].TotalBytes += record.TotalBytes
		summary.HourlyData[hour].Records++
	}
	summary.TotalReceived = totalReceived
	summary.TotalSent = totalSent
	summary.TotalBytes = totalReceived + totalSent
	if len(records) > 0 {
		summary.AvgUploadRate = totalUploadRate / float64(len(records))
		summary.AvgDownloadRate = totalDownloadRate / float64(len(records))
	}
	return summary
}
type DailySummary struct {
	ClientName      string                      `json:"client_name"`
	Date            string                      `json:"date"`
	TotalReceived   uint64                      `json:"total_received"`
	TotalSent       uint64                      `json:"total_sent"`
	TotalBytes      uint64                      `json:"total_bytes"`
	AvgUploadRate   float64                     `json:"avg_upload_rate"`
	AvgDownloadRate float64                     `json:"avg_download_rate"`
	RecordCount     int                         `json:"record_count"`
	HourlyData      map[int]*HourlySummary      `json:"hourly_data"`
}
type HourlySummary struct {
	Hour       int    `json:"hour"`
	TotalBytes uint64 `json:"total_bytes"`
	Records    int    `json:"records"`
}
func (s *Storage) readRecords(filePath string) ([]TrafficRecord, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
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
func (s *Storage) writeDailySummary(filePath string, summary *DailySummary) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tempPath, filePath)
}
func (s *Storage) extractTimestamp(filename, granularity string) (time.Time, error) {
	switch granularity {
	case "hourly":
		parts := strings.Split(filename, "_")
		if len(parts) < 3 {
			return time.Time{}, fmt.Errorf("invalid hourly filename format")
		}
		dateStr := parts[len(parts)-3] + "_" + parts[len(parts)-2]
		return time.Parse("2006-01-02_15", dateStr)
	case "daily":
		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			return time.Time{}, fmt.Errorf("invalid daily filename format")
		}
		dateStr := parts[len(parts)-2]
		return time.Parse("2006-01-02", dateStr)
	default:
		return time.Time{}, fmt.Errorf("unsupported granularity")
	}
}
func (s *Storage) extractTimestampFromPath(filePath, granularity string) (time.Time, error) {
	filename := filepath.Base(filePath)
	return s.extractTimestamp(filename, granularity)
}
func (s *Storage) extractClientName(filename string) string {
	parts := strings.Split(filename, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
func (s *Storage) getRollupPath(clientName, dayKey, granularity string) string {
	return fmt.Sprintf("%s/records/%s/%s_%s_%s.json", s.dataDir, granularity, clientName, dayKey, granularity)
}
func (s *Storage) cleanupOldData() error {
	cutoff := time.Now().AddDate(0, 0, -s.retentionDays)
	if err := s.cleanupGranularity("hourly", cutoff); err != nil {
		return err
	}
	dailyCutoff := time.Now().AddDate(0, 0, -s.retentionDays*2)
	if err := s.cleanupGranularity("daily", dailyCutoff); err != nil {
		return err
	}
	return nil
}
func (s *Storage) cleanupGranularity(granularity string, cutoff time.Time) error {
	dir := filepath.Join(s.dataDir, "records", granularity)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if timestamp, err := s.extractTimestamp(entry.Name(), granularity); err == nil {
			if timestamp.Before(cutoff) {
				filePath := filepath.Join(dir, entry.Name())
				os.Remove(filePath)
			}
		}
	}
	return nil
}
func (s *Storage) GetDailyRollup(clientName, date string) (*DailySummary, error) {
	filePath := s.getRollupPath(clientName, date, "daily")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var summary DailySummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, err
	}
	return &summary, nil
}
func (s *Storage) GetDailyRollupsForPeriod(clientName string, startDate, endDate time.Time) ([]DailySummary, error) {
	var summaries []DailySummary
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if summary, err := s.GetDailyRollup(clientName, dateStr); err == nil {
			summaries = append(summaries, *summary)
		}
	}
	return summaries, nil
}