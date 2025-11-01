package main
import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	Line      int       `json:"line"`
}
type LogQuery struct {
	LogType    string    `json:"log_type"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Level      string    `json:"level"`
	SearchText string    `json:"search_text"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
}
type LogResponse struct {
	Entries []LogEntry `json:"entries"`
	Total   int        `json:"total"`
	HasMore bool       `json:"has_more"`
}
const (
	OpenVPNLogPath  = "/var/log/openvpn.log"
	SystemLogPath   = "/var/log/syslog"
	WebPanelLogPath = "/var/log/irangate-webpanel.log"
	JournalLogPath  = "/var/log/journal"
)
func getOpenVPNLogsHandler(w http.ResponseWriter, r *http.Request) {
	query := parseLogQuery(r)
	logs, total, hasMore, err := readLogs(OpenVPNLogPath, query, "openvpn")
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to read OpenVPN logs: %v", err), http.StatusInternalServerError)
		return
	}
	response := LogResponse{
		Entries: logs,
		Total:   total,
		HasMore: hasMore,
	}
	sendSuccess(w, "OpenVPN logs retrieved successfully", response)
}
func getWebPanelLogsHandler(w http.ResponseWriter, r *http.Request) {
	query := parseLogQuery(r)
	logs, total, hasMore, err := readLogs(WebPanelLogPath, query, "webpanel")
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to read WebPanel logs: %v", err), http.StatusInternalServerError)
		return
	}
	response := LogResponse{
		Entries: logs,
		Total:   total,
		HasMore: hasMore,
	}
	sendSuccess(w, "WebPanel logs retrieved successfully", response)
}
func getSystemLogsHandler(w http.ResponseWriter, r *http.Request) {
	query := parseLogQuery(r)
	logs, total, hasMore, err := readLogs(SystemLogPath, query, "system")
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to read system logs: %v", err), http.StatusInternalServerError)
		return
	}
	response := LogResponse{
		Entries: logs,
		Total:   total,
		HasMore: hasMore,
	}
	sendSuccess(w, "System logs retrieved successfully", response)
}
func searchLogsHandler(w http.ResponseWriter, r *http.Request) {
	var query LogQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		query = parseLogQuery(r)
	}
	if query.SearchText == "" {
		sendError(w, "Search text is required", http.StatusBadRequest)
		return
	}
	var allLogs []LogEntry
	var totalCount int
	logTypes := []string{}
	if query.LogType != "" {
		logTypes = append(logTypes, query.LogType)
	} else {
		logTypes = []string{"openvpn", "webpanel", "system"}
	}
	for _, logType := range logTypes {
		var logPath string
		switch logType {
		case "openvpn":
			logPath = OpenVPNLogPath
		case "webpanel":
			logPath = WebPanelLogPath
		case "system":
			logPath = SystemLogPath
		default:
			continue
		}
		logs, _, _, err := readLogs(logPath, query, logType)
		if err != nil {
			continue
		}
		allLogs = append(allLogs, logs...)
		totalCount += len(logs)
	}
	if query.Limit <= 0 {
		query.Limit = 100
	}
	start := query.Offset
	end := start + query.Limit
	if start >= len(allLogs) {
		allLogs = []LogEntry{}
	} else {
		if end > len(allLogs) {
			end = len(allLogs)
		}
		allLogs = allLogs[start:end]
	}
	response := LogResponse{
		Entries: allLogs,
		Total:   totalCount,
		HasMore: query.Offset+len(allLogs) < totalCount,
	}
	sendSuccess(w, "Log search completed successfully", response)
}
func tailLogsHandler(w http.ResponseWriter, r *http.Request) {
	logType := r.URL.Query().Get("type")
	lines := r.URL.Query().Get("lines")
	if logType == "" {
		logType = "openvpn"
	}
	numLines := 50
	if lines != "" {
		if parsed, err := strconv.Atoi(lines); err == nil && parsed > 0 && parsed <= 1000 {
			numLines = parsed
		}
	}
	var logPath string
	switch logType {
	case "openvpn":
		logPath = OpenVPNLogPath
	case "webpanel":
		logPath = WebPanelLogPath
	case "system":
		logPath = SystemLogPath
	default:
		sendError(w, "Invalid log type", http.StatusBadRequest)
		return
	}
	tailLogs, err := tailLogFile(logPath, numLines, logType)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to tail log file: %v", err), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Log tail retrieved successfully", tailLogs)
}
func streamLogsHandler(w http.ResponseWriter, r *http.Request) {
	logType := r.URL.Query().Get("type")
	if logType == "" {
		logType = "openvpn"
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var logPath string
	switch logType {
	case "openvpn":
		logPath = OpenVPNLogPath
	case "webpanel":
		logPath = WebPanelLogPath
	case "system":
		logPath = SystemLogPath
	default:
		fmt.Fprintf(w, "data: {\"error\": \"Invalid log type\"}\n\n")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "data: {\"message\": \"Log streaming started\"}\n\n")
	flusher.Flush()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			recentLogs, err := tailLogFile(logPath, 5, logType)
			if err == nil && len(recentLogs) > 0 {
				data, _ := json.Marshal(recentLogs)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}
func parseLogQuery(r *http.Request) LogQuery {
	query := LogQuery{
		Limit: 100,
	}
	if logType := r.URL.Query().Get("log_type"); logType != "" {
		query.LogType = logType
	}
	if level := r.URL.Query().Get("level"); level != "" {
		query.Level = level
	}
	if searchText := r.URL.Query().Get("search"); searchText != "" {
		query.SearchText = searchText
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil && parsed > 0 {
			query.Limit = parsed
		}
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		if parsed, err := strconv.Atoi(offset); err == nil && parsed >= 0 {
			query.Offset = parsed
		}
	}
	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if parsed, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = parsed
		}
	}
	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if parsed, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = parsed
		}
	}
	return query
}
func readLogs(logPath string, query LogQuery, source string) ([]LogEntry, int, bool, error) {
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return []LogEntry{}, 0, false, fmt.Errorf("log file not found: %s", logPath)
	}
	file, err := os.Open(logPath)
	if err != nil {
		return nil, 0, false, err
	}
	defer file.Close()
	var entries []LogEntry
	var allEntries []LogEntry
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		entry := parseLogLine(line, source, lineNum)
		if entry == nil {
			continue
		}
		if !matchesFilters(entry, query) {
			continue
		}
		allEntries = append(allEntries, *entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, false, err
	}
	totalCount := len(allEntries)
	start := query.Offset
	end := start + query.Limit
	if start >= len(allEntries) {
		entries = []LogEntry{}
	} else {
		if end > len(allEntries) {
			end = len(allEntries)
		}
		entries = allEntries[start:end]
	}
	hasMore := query.Offset+len(entries) < totalCount
	return entries, totalCount, hasMore, nil
}
func parseLogLine(line, source string, lineNum int) *LogEntry {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+\w+\s+(\w+):\s*(.*)$`),
		regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}|\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s*(.*)$`),
		regexp.MustCompile(`^\s*\{.*\}\s*$`),
	}
	var timestamp time.Time
	var level string
	var message string
	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			if parsedTime, err := time.Parse("2006-01-02T15:04:05-07:00", matches[1]); err == nil {
				timestamp = parsedTime
			} else if parsedTime, err := time.Parse("Jan 02 15:04:05", matches[1]); err == nil {
				year := time.Now().Year()
				timestamp = time.Date(year, parsedTime.Month(), parsedTime.Day(),
					parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), 0, time.Local)
			}
			if len(matches) >= 3 {
				level = extractLogLevel(matches[2])
				message = matches[2]
			} else if len(matches) >= 2 {
				message = matches[1]
				level = "info"
			}
			break
		}
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
		level = "info"
		message = line
	}
	return &LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Message:   message,
		Source:    source,
		Line:      lineNum,
	}
}
func extractLogLevel(text string) string {
	text = strings.ToLower(text)
	if strings.Contains(text, "error") || strings.Contains(text, "err") {
		return "error"
	} else if strings.Contains(text, "warn") {
		return "warn"
	} else if strings.Contains(text, "debug") {
		return "debug"
	} else if strings.Contains(text, "info") {
		return "info"
	}
	return "info"
}
func matchesFilters(entry *LogEntry, query LogQuery) bool {
	if !query.StartTime.IsZero() && entry.Timestamp.Before(query.StartTime) {
		return false
	}
	if !query.EndTime.IsZero() && entry.Timestamp.After(query.EndTime) {
		return false
	}
	if query.Level != "" && entry.Level != query.Level {
		return false
	}
	if query.SearchText != "" {
		searchLower := strings.ToLower(query.SearchText)
		messageLower := strings.ToLower(entry.Message)
		if !strings.Contains(messageLower, searchLower) {
			return false
		}
	}
	return true
}
func tailLogFile(logPath string, numLines int, source string) ([]LogEntry, error) {
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return []LogEntry{}, fmt.Errorf("log file not found: %s", logPath)
	}
	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > numLines {
			lines = lines[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	var entries []LogEntry
	for i, line := range lines {
		entry := parseLogLine(line, source, i+1)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}
	return entries, nil
}