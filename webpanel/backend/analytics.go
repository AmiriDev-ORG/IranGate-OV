package main
import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/gorilla/mux"
)
type TrafficHistoryRequest struct {
	Period string `json:"period"`
	Client string `json:"client"`
}
type TrafficHistoryResponse struct {
	Period     string           `json:"period"`
	StartDate  time.Time        `json:"start_date"`
	EndDate    time.Time        `json:"end_date"`
	Data       []TrafficPoint   `json:"data"`
	TotalBytes int64            `json:"total_bytes"`
	ClientData map[string]int64 `json:"client_data,omitempty"`
}
type TrafficPoint struct {
	Timestamp time.Time `json:"timestamp"`
	BytesIn   int64     `json:"bytes_in"`
	BytesOut  int64     `json:"bytes_out"`
	Total     int64     `json:"total"`
}
type BandwidthUsageResponse struct {
	ClientName     string    `json:"client_name"`
	BytesUsed      int64     `json:"bytes_used"`
	BytesUsedGB    float64   `json:"bytes_used_gb"`
	QuotaLimit     int64     `json:"quota_limit"`
	QuotaLimitGB   float64   `json:"quota_limit_gb"`
	QuotaUsed      float64   `json:"quota_used"`
	QuotaEnabled   bool      `json:"quota_enabled"`
	QuotaResetDate time.Time `json:"quota_reset_date"`
	RemainingGB    float64   `json:"remaining_gb"`
}
type PeakHoursResponse struct {
	HourlyUsage map[int]int64    `json:"hourly_usage"`
	DailyUsage  map[string]int64 `json:"daily_usage"`
	PeakHour    int              `json:"peak_hour"`
	PeakDay     string           `json:"peak_day"`
}
var AnalyticsDataDir = func() string {
	if dir := os.Getenv("IRANGATE_DATA_DIR"); dir != "" {
		return filepath.Join(dir, "webpanel", "analytics")
	}
	return "/opt/irangate/webpanel/analytics"
}()
func trafficHistoryHandler(w http.ResponseWriter, r *http.Request) {
	var req TrafficHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Period = r.URL.Query().Get("period")
		if req.Period == "" {
			req.Period = "7d"
		}
		req.Client = r.URL.Query().Get("client")
	}
	endTime := time.Now()
	var startTime time.Time
	switch req.Period {
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	case "90d":
		startTime = endTime.Add(-90 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-7 * 24 * time.Hour)
	}
	data, err := loadTrafficData(startTime, endTime, req.Client)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to load traffic data: %v", err), http.StatusInternalServerError)
		return
	}
	response := TrafficHistoryResponse{
		Period:    req.Period,
		StartDate: startTime,
		EndDate:   endTime,
		Data:      data,
	}
	totalBytes := int64(0)
	for _, point := range data {
		totalBytes += point.Total
	}
	response.TotalBytes = totalBytes
	if req.Client != "" {
		response.ClientData = make(map[string]int64)
		response.ClientData[req.Client] = totalBytes
	}
	sendSuccess(w, "Traffic history retrieved successfully", response)
}
func bandwidthUsageHandler(w http.ResponseWriter, r *http.Request) {
	db, err := database.New(database.GetDefaultDatabasePath())
	if err != nil {
		sendError(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	clients, err := db.GetAllClients()
	if err != nil {
		sendError(w, "Failed to get clients", http.StatusInternalServerError)
		return
	}
	var usageData []BandwidthUsageResponse
	totalBytesUsed := int64(0)
	for _, client := range clients {
		bytesUsed := int64(client.BytesReceived + client.BytesSent)
		totalBytesUsed += bytesUsed
		quotaLimit := client.BandwidthQuotaGB * 1024 * 1024 * 1024
		quotaUsed := float64(0)
		if quotaLimit > 0 {
			quotaUsed = (float64(bytesUsed) / float64(quotaLimit)) * 100
		}
		usage := BandwidthUsageResponse{
			ClientName:     client.Name,
			BytesUsed:      bytesUsed,
			BytesUsedGB:    float64(bytesUsed) / (1024 * 1024 * 1024),
			QuotaLimit:     quotaLimit,
			QuotaLimitGB:   float64(quotaLimit) / (1024 * 1024 * 1024),
			QuotaUsed:      quotaUsed,
			QuotaEnabled:   client.QuotaEnabled,
			QuotaResetDate: client.QuotaResetDate,
			RemainingGB:    float64(quotaLimit-bytesUsed) / (1024 * 1024 * 1024),
		}
		usageData = append(usageData, usage)
	}
	sendSuccess(w, "Bandwidth usage retrieved successfully", map[string]interface{}{
		"clients":          usageData,
		"total_bytes_used": totalBytesUsed,
		"total_bytes_gb":   float64(totalBytesUsed) / (1024 * 1024 * 1024),
	})
}
func peakHoursHandler(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d"
	}
	endTime := time.Now()
	var startTime time.Time
	switch period {
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-7 * 24 * time.Hour)
	}
	data, err := loadTrafficData(startTime, endTime, "")
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to load traffic data: %v", err), http.StatusInternalServerError)
		return
	}
	hourlyUsage := make(map[int]int64)
	dailyUsage := make(map[string]int64)
	var peakHour int
	var peakDay string
	var maxHourlyBytes int64
	var maxDailyBytes int64
	for _, point := range data {
		hour := point.Timestamp.Hour()
		day := point.Timestamp.Format("2006-01-02")
		hourlyUsage[hour] += point.Total
		dailyUsage[day] += point.Total
		if hourlyUsage[hour] > maxHourlyBytes {
			maxHourlyBytes = hourlyUsage[hour]
			peakHour = hour
		}
		if dailyUsage[day] > maxDailyBytes {
			maxDailyBytes = dailyUsage[day]
			peakDay = day
		}
	}
	response := PeakHoursResponse{
		HourlyUsage: hourlyUsage,
		DailyUsage:  dailyUsage,
		PeakHour:    peakHour,
		PeakDay:     peakDay,
	}
	sendSuccess(w, "Peak hours analysis retrieved successfully", response)
}
func exportReportHandler(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	period := r.URL.Query().Get("period")
	client := r.URL.Query().Get("client")
	if format == "" {
		format = "csv"
	}
	if period == "" {
		period = "30d"
	}
	endTime := time.Now()
	var startTime time.Time
	switch period {
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	case "90d":
		startTime = endTime.Add(-90 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-30 * 24 * time.Hour)
	}
	data, err := loadTrafficData(startTime, endTime, client)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to load traffic data: %v", err), http.StatusInternalServerError)
		return
	}
	filename := fmt.Sprintf("traffic_report_%s_%s.%s",
		startTime.Format("20060102"),
		endTime.Format("20060102"),
		format)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		writeCSVReport(w, data, startTime, endTime, client)
	} else {
		w.Header().Set("Content-Type", "application/json")
		writeJSONReport(w, data, startTime, endTime, client)
	}
}
func setClientQuotaHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	var req struct {
		QuotaGB      int64     `json:"quota_gb"`
		QuotaEnabled bool      `json:"quota_enabled"`
		ResetDate    time.Time `json:"reset_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	db, err := database.New(database.GetDefaultDatabasePath())
	if err != nil {
		sendError(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	client, err := db.GetClient(clientName)
	if err != nil {
		sendError(w, "Client not found", http.StatusNotFound)
		return
	}
	client.BandwidthQuotaGB = req.QuotaGB
	client.QuotaEnabled = req.QuotaEnabled
	client.QuotaResetDate = req.ResetDate
	if err := db.UpdateClient(clientName, client); err != nil {
		sendError(w, fmt.Sprintf("Failed to update client: %v", err), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Client quota updated successfully", map[string]interface{}{
		"client_name":   clientName,
		"quota_gb":      req.QuotaGB,
		"quota_enabled": req.QuotaEnabled,
		"reset_date":    req.ResetDate,
	})
}
func getClientQuotaStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	db, err := database.New(database.GetDefaultDatabasePath())
	if err != nil {
		sendError(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	client, err := db.GetClient(clientName)
	if err != nil {
		sendError(w, "Client not found", http.StatusNotFound)
		return
	}
	bytesUsed := int64(client.BytesReceived + client.BytesSent)
	quotaLimit := client.BandwidthQuotaGB * 1024 * 1024 * 1024
	quotaUsed := float64(0)
	if quotaLimit > 0 {
		quotaUsed = (float64(bytesUsed) / float64(quotaLimit)) * 100
	}
	status := BandwidthUsageResponse{
		ClientName:     client.Name,
		BytesUsed:      bytesUsed,
		BytesUsedGB:    float64(bytesUsed) / (1024 * 1024 * 1024),
		QuotaLimit:     quotaLimit,
		QuotaLimitGB:   float64(quotaLimit) / (1024 * 1024 * 1024),
		QuotaUsed:      quotaUsed,
		QuotaEnabled:   client.QuotaEnabled,
		QuotaResetDate: client.QuotaResetDate,
		RemainingGB:    float64(quotaLimit-bytesUsed) / (1024 * 1024 * 1024),
	}
	sendSuccess(w, "Client quota status retrieved successfully", status)
}
func loadTrafficData(startTime, endTime time.Time, clientFilter string) ([]TrafficPoint, error) {
	db := NewTrafficDatabase()
	if clientFilter == "" {
		return loadAllClientsTrafficData(db, startTime, endTime)
	}
	records, err := db.GetClientTrafficHistory(clientFilter, startTime, endTime)
	if err != nil {
		return nil, err
	}
	var data []TrafficPoint
	for _, record := range records {
		data = append(data, TrafficPoint{
			Timestamp: record.Timestamp,
			BytesIn:   record.BytesReceived,
			BytesOut:  record.BytesSent,
			Total:     record.TotalBytes,
		})
	}
	return data, nil
}
func loadAllClientsTrafficData(db *TrafficDatabase, startTime, endTime time.Time) ([]TrafficPoint, error) {
	clients, err := loadAllClients()
	if err != nil {
		return nil, err
	}
	var allData []TrafficPoint
	for _, client := range clients {
		clientName, ok := client["name"].(string)
		if !ok {
			continue
		}
		records, err := db.GetClientTrafficHistory(clientName, startTime, endTime)
		if err != nil {
			continue
		}
		interval := time.Hour
		if endTime.Sub(startTime) <= 24*time.Hour {
			interval = 15 * time.Minute
		}
		timeMap := make(map[string]*TrafficPoint)
		for _, record := range records {
			roundedTime := record.Timestamp.Truncate(interval)
			key := roundedTime.Format(time.RFC3339)
			if point, exists := timeMap[key]; exists {
				point.BytesIn += record.BytesReceived
				point.BytesOut += record.BytesSent
				point.Total += record.TotalBytes
			} else {
				timeMap[key] = &TrafficPoint{
					Timestamp: roundedTime,
					BytesIn:   record.BytesReceived,
					BytesOut:  record.BytesSent,
					Total:     record.TotalBytes,
				}
			}
		}
		for _, point := range timeMap {
			allData = append(allData, *point)
		}
	}
	return allData, nil
}
func loadAllClients() ([]map[string]interface{}, error) {
	cmd := exec.Command("irangate", "client", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	var clients []map[string]interface{}
	if err := json.Unmarshal(output, &clients); err != nil {
		return nil, err
	}
	return clients, nil
}
func writeCSVReport(w http.ResponseWriter, data []TrafficPoint, startTime, endTime time.Time, client string) {
	filteredData := make([]TrafficPoint, 0)
	for _, point := range data {
		if point.Timestamp.After(startTime) && point.Timestamp.Before(endTime) {
			filteredData = append(filteredData, point)
		}
	}
	writer := csv.NewWriter(w)
	defer writer.Flush()
	writer.Write([]string{"Timestamp", "Bytes In", "Bytes Out", "Total Bytes", "Client"})
	for _, point := range filteredData {
		writer.Write([]string{
			point.Timestamp.Format("2006-01-02 15:04:05"),
			strconv.FormatInt(point.BytesIn, 10),
			strconv.FormatInt(point.BytesOut, 10),
			strconv.FormatInt(point.Total, 10),
			client,
		})
	}
}
func writeJSONReport(w http.ResponseWriter, data []TrafficPoint, startTime, endTime time.Time, client string) {
	report := map[string]interface{}{
		"report_info": map[string]interface{}{
			"start_date":   startTime,
			"end_date":     endTime,
			"client":       client,
			"total_points": len(data),
		},
		"data": data,
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(report)
}