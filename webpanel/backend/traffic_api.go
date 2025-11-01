package main

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/amiridev-org/irangate-ov/pkg/traffic"
	"github.com/gorilla/mux"
)

func getTrafficStatsHandler(w http.ResponseWriter, r *http.Request) {
	analyzer := getTrafficAnalyzer()
	clients, err := getAllClientNames()
	if err != nil {
		sendError(w, "Failed to get clients", http.StatusInternalServerError)
		return
	}
	stats := make(map[string]interface{})
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -1)
	for _, clientName := range clients {
		clientStats, err := analyzer.GetClientAggregatedStats(clientName, startTime, endTime, "custom")
		if err != nil {
			continue
		}
		stats[clientName] = clientStats
	}
	sendSuccess(w, "Traffic statistics retrieved", stats)
}
func getClientTrafficStatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	if clientName == "" {
		sendError(w, "Client name required", http.StatusBadRequest)
		return
	}
	analyzer := getTrafficAnalyzer()
	days := getIntQueryParam(r, "days", 7)
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)
	stats, err := analyzer.GetClientAggregatedStats(clientName, startTime, endTime, "custom")
	if err != nil {
		sendError(w, "Failed to get traffic stats", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Client traffic statistics retrieved", stats)
}
func getTrafficHistoryHandler(w http.ResponseWriter, r *http.Request) {
	clientName := r.URL.Query().Get("client")
	days := getIntQueryParam(r, "days", 7)
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)
	analyzer := getTrafficAnalyzer()
	if clientName == "" {
		response := make(map[string]interface{})
		clients, _ := getAllClientNames()
		for _, name := range clients {
			pattern, err := analyzer.GetUsagePattern(name, startTime, endTime, time.Hour)
			if err == nil {
				response[name] = pattern
			}
		}
		sendSuccess(w, "Traffic history retrieved", response)
	} else {
		pattern, err := analyzer.GetUsagePattern(clientName, startTime, endTime, time.Hour)
		if err != nil {
			sendError(w, "Failed to get traffic history", http.StatusInternalServerError)
			return
		}
		sendSuccess(w, "Client traffic history retrieved", pattern)
	}
}
func getTopUsersHandler(w http.ResponseWriter, r *http.Request) {
	limit := getIntQueryParam(r, "limit", 10)
	days := getIntQueryParam(r, "days", 7)
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)
	analyzer := getTrafficAnalyzer()
	topUsers, err := analyzer.GetTopUsers(limit, startTime, endTime)
	if err != nil {
		sendError(w, "Failed to get top users", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Top users retrieved", topUsers)
}
func exportTrafficHandler(w http.ResponseWriter, r *http.Request) {
	clientName := r.URL.Query().Get("client")
	format := r.URL.Query().Get("format")
	days := getIntQueryParam(r, "days", 30)
	if format == "" {
		format = "json"
	}
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)
	analyzer := getTrafficAnalyzer()
	switch format {
	case "json":
		exportJSON(w, analyzer, clientName, startTime, endTime)
	case "csv":
		exportCSV(w, analyzer, clientName, startTime, endTime)
	default:
		sendError(w, "Invalid format. Supported: json, csv", http.StatusBadRequest)
	}
}
func getQuotaUsageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	if clientName == "" {
		sendError(w, "Client name required", http.StatusBadRequest)
		return
	}
	analyzer := getTrafficAnalyzer()
	db, err := getDatabase()
	if err != nil {
		sendError(w, "Failed to access database", http.StatusInternalServerError)
		return
	}
	client, err := db.GetClient(clientName)
	if err != nil {
		sendError(w, "Client not found", http.StatusNotFound)
		return
	}
	periodStart := client.QuotaResetDate
	if periodStart.IsZero() {
		periodStart = time.Now().AddDate(0, 0, -30)
	}
	periodEnd := time.Now()
	if !client.QuotaEnabled {
		sendSuccess(w, "Quota not enabled for client", map[string]interface{}{
			"quota_enabled": false,
		})
		return
	}
	quotaUsage, err := analyzer.GetQuotaUsage(
		clientName,
		uint64(client.BandwidthQuotaGB*1024*1024*1024),
		client.QuotaResetDate,
		periodStart,
		periodEnd,
	)
	if err != nil {
		sendError(w, "Failed to calculate quota usage", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Quota usage retrieved", quotaUsage)
}
func getBandwidthEfficiencyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	if clientName == "" {
		sendError(w, "Client name required", http.StatusBadRequest)
		return
	}
	days := getIntQueryParam(r, "days", 7)
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)
	analyzer := getTrafficAnalyzer()
	efficiency, err := analyzer.GetBandwidthEfficiency(clientName, startTime, endTime)
	if err != nil {
		sendError(w, "Failed to get efficiency metrics", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Bandwidth efficiency metrics retrieved", efficiency)
}
func detectAnomaliesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	if clientName == "" {
		sendError(w, "Client name required", http.StatusBadRequest)
		return
	}
	days := getIntQueryParam(r, "days", 7)
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)
	analyzer := getTrafficAnalyzer()
	anomalies, err := analyzer.DetectAnomalies(clientName, startTime, endTime)
	if err != nil {
		sendError(w, "Failed to detect anomalies", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Anomaly detection completed", anomalies)
}
func getRealTimeStatsHandler(w http.ResponseWriter, r *http.Request) {
	collector := getTrafficCollector()
	stats, err := collector.GetCurrentStats()
	if err != nil {
		sendError(w, "Failed to get real-time stats", http.StatusInternalServerError)
		return
	}
	statsArray := make([]interface{}, 0, len(stats))
	for _, stat := range stats {
		statsArray = append(statsArray, stat)
	}
	sendSuccess(w, "Real-time statistics retrieved", statsArray)
}

var trafficCollectorInstance *traffic.Collector
var trafficAnalyzerInstance *traffic.Analyzer

func getTrafficCollector() *traffic.Collector {
	if trafficCollectorInstance == nil {
		db, _ := getDatabase()
		trafficDataDir := os.Getenv("IRANGATE_TRAFFIC_DIR")
		if trafficDataDir == "" {
			if baseDir := os.Getenv("IRANGATE_DATA_DIR"); baseDir != "" {
				trafficDataDir = filepath.Join(baseDir, "traffic")
			} else {
				trafficDataDir = "/opt/irangate/traffic"
			}
		}
		trafficCollectorInstance = traffic.NewCollector(db, trafficDataDir)
	}
	return trafficCollectorInstance
}
func getTrafficAnalyzer() *traffic.Analyzer {
	if trafficAnalyzerInstance == nil {
		trafficDataDir := os.Getenv("IRANGATE_TRAFFIC_DIR")
		if trafficDataDir == "" {
			if baseDir := os.Getenv("IRANGATE_DATA_DIR"); baseDir != "" {
				trafficDataDir = filepath.Join(baseDir, "traffic")
			} else {
				trafficDataDir = "/opt/irangate/traffic"
			}
		}
		trafficAnalyzerInstance = traffic.NewAnalyzer(trafficDataDir)
	}
	return trafficAnalyzerInstance
}
func exportJSON(w http.ResponseWriter, analyzer *traffic.Analyzer, clientName string, startTime, endTime time.Time) {
	w.Header().Set("Content-Type", "application/json")
	if clientName == "" {
		clients, _ := getAllClientNames()
		export := make(map[string]interface{})
		for _, name := range clients {
			stats, err := analyzer.GetClientAggregatedStats(name, startTime, endTime, "custom")
			if err == nil {
				export[name] = stats
			}
		}
		json.NewEncoder(w).Encode(export)
	} else {
		stats, _ := analyzer.GetClientAggregatedStats(clientName, startTime, endTime, "custom")
		json.NewEncoder(w).Encode(stats)
	}
}
func exportCSV(w http.ResponseWriter, analyzer *traffic.Analyzer, clientName string, startTime, endTime time.Time) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=traffic_export.csv")
	writer := csv.NewWriter(w)
	defer writer.Flush()
	writer.Write([]string{"Client Name", "Period", "Total Received (bytes)", "Total Sent (bytes)", "Total Bytes", "Avg Upload Rate (bps)", "Avg Download Rate (bps)", "Connected Time (sec)"})
	if clientName == "" {
		clients, _ := getAllClientNames()
		for _, name := range clients {
			stats, err := analyzer.GetClientAggregatedStats(name, startTime, endTime, "custom")
			if err == nil {
				writer.Write([]string{
					stats.ClientName,
					stats.Period,
					strconv.FormatUint(stats.TotalReceived, 10),
					strconv.FormatUint(stats.TotalSent, 10),
					strconv.FormatUint(stats.TotalBytes, 10),
					strconv.FormatFloat(stats.AvgUploadRate, 'f', 2, 64),
					strconv.FormatFloat(stats.AvgDownloadRate, 'f', 2, 64),
					strconv.FormatFloat(stats.ConnectedTime, 'f', 2, 64),
				})
			}
		}
	} else {
		stats, _ := analyzer.GetClientAggregatedStats(clientName, startTime, endTime, "custom")
		writer.Write([]string{
			stats.ClientName,
			stats.Period,
			strconv.FormatUint(stats.TotalReceived, 10),
			strconv.FormatUint(stats.TotalSent, 10),
			strconv.FormatUint(stats.TotalBytes, 10),
			strconv.FormatFloat(stats.AvgUploadRate, 'f', 2, 64),
			strconv.FormatFloat(stats.AvgDownloadRate, 'f', 2, 64),
			strconv.FormatFloat(stats.ConnectedTime, 'f', 2, 64),
		})
	}
}
func getDatabase() (*database.DB, error) {
	return database.New("")
}
func getAllClientNames() ([]string, error) {
	db, err := getDatabase()
	if err != nil {
		return nil, err
	}
	clients, err := db.GetClients()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(clients))
	for i, client := range clients {
		names[i] = client.Name
	}
	return names, nil
}
func getIntQueryParam(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}
