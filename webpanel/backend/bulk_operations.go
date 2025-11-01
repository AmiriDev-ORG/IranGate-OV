package main
import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
)
type BulkCreateRequest struct {
	Clients []BulkClientData `json:"clients"`
	Format  string           `json:"format"`
}
type BulkClientData struct {
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	QuotaGB   int64     `json:"quota_gb,omitempty"`
	Notes     string    `json:"notes,omitempty"`
}
type BulkDeleteRequest struct {
	ClientNames []string `json:"client_names"`
}
type BulkExportRequest struct {
	ClientNames []string `json:"client_names"`
	Format      string   `json:"format"`
}
type BulkExtendRequest struct {
	ClientNames     []string `json:"client_names"`
	ExtensionDays   int      `json:"extension_days"`
	ExtensionMonths int      `json:"extension_months,omitempty"`
}
type BulkOperationResponse struct {
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	Results      []BulkOperationResult  `json:"results"`
	Summary      map[string]interface{} `json:"summary"`
}
type BulkOperationResult struct {
	ClientName string `json:"client_name"`
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}
func bulkCreateClientsHandler(w http.ResponseWriter, r *http.Request) {
	var req BulkCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Clients) == 0 {
		sendError(w, "No clients provided", http.StatusBadRequest)
		return
	}
	if len(req.Clients) > 100 {
		sendError(w, "Maximum 100 clients allowed per bulk operation", http.StatusBadRequest)
		return
	}
	db, err := database.New(database.GetDefaultDatabasePath())
	if err != nil {
		sendError(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	var results []BulkOperationResult
	successCount := 0
	failureCount := 0
	for _, clientData := range req.Clients {
		result := BulkOperationResult{
			ClientName: clientData.Name,
		}
		if clientData.Name == "" {
			result.Success = false
			result.Error = "Client name is required"
			failureCount++
		} else {
			_, err := db.GetClient(clientData.Name)
			if err == nil {
				result.Success = false
				result.Error = "Client already exists"
				failureCount++
			} else {
				client := &database.Client{
					Name:             clientData.Name,
					CreatedAt:        time.Now(),
					ExpiresAt:        clientData.ExpiresAt,
					Active:           true,
					BandwidthQuotaGB: clientData.QuotaGB,
					QuotaEnabled:     clientData.QuotaGB > 0,
					Status:           "active",
					LastUpdate:       time.Now(),
				}
				if client.ExpiresAt.IsZero() {
					client.ExpiresAt = time.Now().AddDate(1, 0, 0)
				}
				if client.QuotaEnabled {
					client.QuotaResetDate = time.Now().AddDate(0, 1, 0)
				}
				err = db.UpdateClient(client.Name, client)
				if err != nil {
					result.Success = false
					result.Error = fmt.Sprintf("Failed to create client: %v", err)
					failureCount++
				} else {
					result.Success = true
					result.Message = "Client created successfully"
					successCount++
				}
			}
		}
		results = append(results, result)
	}
	response := BulkOperationResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
		Summary: map[string]interface{}{
			"total_requested": len(req.Clients),
			"success_rate":    float64(successCount) / float64(len(req.Clients)) * 100,
		},
	}
	sendSuccess(w, "Bulk client creation completed", response)
}
func bulkDeleteClientsHandler(w http.ResponseWriter, r *http.Request) {
	var req BulkDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.ClientNames) == 0 {
		sendError(w, "No client names provided", http.StatusBadRequest)
		return
	}
	if len(req.ClientNames) > 100 {
		sendError(w, "Maximum 100 clients allowed per bulk operation", http.StatusBadRequest)
		return
	}
	db, err := database.New(database.GetDefaultDatabasePath())
	if err != nil {
		sendError(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	var results []BulkOperationResult
	successCount := 0
	failureCount := 0
	for _, clientName := range req.ClientNames {
		result := BulkOperationResult{
			ClientName: clientName,
		}
		_, err := db.GetClient(clientName)
		if err != nil {
			result.Success = false
			result.Error = "Client not found"
			failureCount++
		} else {
			err = db.RemoveClient(clientName)
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("Failed to delete client: %v", err)
				failureCount++
			} else {
				result.Success = true
				result.Message = "Client deleted successfully"
				successCount++
			}
		}
		results = append(results, result)
	}
	response := BulkOperationResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
		Summary: map[string]interface{}{
			"total_requested": len(req.ClientNames),
			"success_rate":    float64(successCount) / float64(len(req.ClientNames)) * 100,
		},
	}
	sendSuccess(w, "Bulk client deletion completed", response)
}
func bulkExportClientsHandler(w http.ResponseWriter, r *http.Request) {
	var req BulkExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.ClientNames) == 0 {
		sendError(w, "No client names provided", http.StatusBadRequest)
		return
	}
	if len(req.ClientNames) > 100 {
		sendError(w, "Maximum 100 clients allowed per bulk operation", http.StatusBadRequest)
		return
	}
	if req.Format == "" {
		req.Format = "zip"
	}
	db, err := database.New(database.GetDefaultDatabasePath())
	if err != nil {
		sendError(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	timestamp := time.Now().Format("20060102_150405")
	var filename string
	var contentType string
	if req.Format == "zip" {
		filename = fmt.Sprintf("clients_export_%s.zip", timestamp)
		contentType = "application/zip"
	} else {
		filename = fmt.Sprintf("clients_export_%s.csv", timestamp)
		contentType = "text/csv"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if req.Format == "zip" {
		zipWriter := zip.NewWriter(w)
		defer zipWriter.Close()
		successCount := 0
		for _, clientName := range req.ClientNames {
			client, err := db.GetClient(clientName)
			if err != nil {
				continue
			}
			configContent, err := getClientConfigContent(clientName)
			if err == nil {
				configWriter, err := zipWriter.Create(clientName + ".ovpn")
				if err == nil {
					_, err = configWriter.Write([]byte(configContent))
					if err == nil {
						successCount++
					}
				}
			}
			clientInfo, _ := json.MarshalIndent(client, "", "  ")
			infoWriter, err := zipWriter.Create(clientName + "_info.json")
			if err == nil {
				_, _ = infoWriter.Write(clientInfo)
			}
		}
		summary := map[string]interface{}{
			"export_date":    time.Now(),
			"total_clients":  len(req.ClientNames),
			"exported_count": successCount,
			"format":         "zip",
		}
		summaryData, _ := json.MarshalIndent(summary, "", "  ")
		summaryWriter, _ := zipWriter.Create("export_summary.json")
		_, _ = summaryWriter.Write(summaryData)
	} else {
		csvWriter := csv.NewWriter(w)
		defer csvWriter.Flush()
		_ = csvWriter.Write([]string{
			"Name", "Created", "Expires", "Active",
			"Quota_GB", "Quota_Enabled", "Status", "Last_Update",
		})
		successCount := 0
		for _, clientName := range req.ClientNames {
			client, err := db.GetClient(clientName)
			if err != nil {
				continue
			}
			_ = csvWriter.Write([]string{
				client.Name,
				client.CreatedAt.Format("2006-01-02 15:04:05"),
				client.ExpiresAt.Format("2006-01-02 15:04:05"),
				strconv.FormatBool(client.Active),
				strconv.FormatInt(client.BandwidthQuotaGB, 10),
				strconv.FormatBool(client.QuotaEnabled),
				client.Status,
				client.LastUpdate.Format("2006-01-02 15:04:05"),
			})
			successCount++
		}
	}
}
func bulkExtendClientsHandler(w http.ResponseWriter, r *http.Request) {
	var req BulkExtendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.ClientNames) == 0 {
		sendError(w, "No client names provided", http.StatusBadRequest)
		return
	}
	if len(req.ClientNames) > 100 {
		sendError(w, "Maximum 100 clients allowed per bulk operation", http.StatusBadRequest)
		return
	}
	if req.ExtensionDays <= 0 && req.ExtensionMonths <= 0 {
		sendError(w, "Extension days or months must be greater than 0", http.StatusBadRequest)
		return
	}
	db, err := database.New(database.GetDefaultDatabasePath())
	if err != nil {
		sendError(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	var results []BulkOperationResult
	successCount := 0
	failureCount := 0
	for _, clientName := range req.ClientNames {
		result := BulkOperationResult{
			ClientName: clientName,
		}
		client, err := db.GetClient(clientName)
		if err != nil {
			result.Success = false
			result.Error = "Client not found"
			failureCount++
		} else {
			newExpiry := client.ExpiresAt
			if req.ExtensionMonths > 0 {
				newExpiry = newExpiry.AddDate(0, req.ExtensionMonths, 0)
			}
			if req.ExtensionDays > 0 {
				newExpiry = newExpiry.AddDate(0, 0, req.ExtensionDays)
			}
			client.ExpiresAt = newExpiry
			client.LastUpdate = time.Now()
			err = db.UpdateClient(clientName, client)
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("Failed to extend client: %v", err)
				failureCount++
			} else {
				result.Success = true
				result.Message = fmt.Sprintf("Client extended until %s",
					newExpiry.Format("2006-01-02"))
				successCount++
			}
		}
		results = append(results, result)
	}
	response := BulkOperationResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
		Summary: map[string]interface{}{
			"total_requested":  len(req.ClientNames),
			"extension_days":   req.ExtensionDays,
			"extension_months": req.ExtensionMonths,
			"success_rate":     float64(successCount) / float64(len(req.ClientNames)) * 100,
		},
	}
	sendSuccess(w, "Bulk client extension completed", response)
}
func bulkImportClientsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		sendError(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		sendError(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()
	format := r.FormValue("format")
	if format == "" {
		if strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
			format = "csv"
		} else {
			format = "json"
		}
	}
	var clients []BulkClientData
	var results []BulkOperationResult
	successCount := 0
	failureCount := 0
	if format == "csv" {
		content, err := io.ReadAll(file)
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusBadRequest)
			return
		}
		clients, err = parseCSVFromString(string(content))
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to parse CSV: %v", err), http.StatusBadRequest)
			return
		}
	} else {
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&clients); err != nil {
			sendError(w, fmt.Sprintf("Failed to parse JSON: %v", err), http.StatusBadRequest)
			return
		}
	}
	if len(clients) == 0 {
		sendError(w, "No client data found in file", http.StatusBadRequest)
		return
	}
	if len(clients) > 100 {
		sendError(w, "Maximum 100 clients allowed per import", http.StatusBadRequest)
		return
	}
	db, err := database.New(database.GetDefaultDatabasePath())
	if err != nil {
		sendError(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	for _, clientData := range clients {
		result := BulkOperationResult{
			ClientName: clientData.Name,
		}
		if clientData.Name == "" {
			result.Success = false
			result.Error = "Client name is required"
			failureCount++
		} else {
			_, err := db.GetClient(clientData.Name)
			if err == nil {
				result.Success = false
				result.Error = "Client already exists"
				failureCount++
			} else {
				client := &database.Client{
					Name:             clientData.Name,
					CreatedAt:        time.Now(),
					ExpiresAt:        clientData.ExpiresAt,
					Active:           true,
					BandwidthQuotaGB: clientData.QuotaGB,
					QuotaEnabled:     clientData.QuotaGB > 0,
					Status:           "active",
					LastUpdate:       time.Now(),
				}
				if client.ExpiresAt.IsZero() {
					client.ExpiresAt = time.Now().AddDate(1, 0, 0)
				}
				err = db.UpdateClient(client.Name, client)
				if err != nil {
					result.Success = false
					result.Error = fmt.Sprintf("Failed to create client: %v", err)
					failureCount++
				} else {
					result.Success = true
					result.Message = "Client imported successfully"
					successCount++
				}
			}
		}
		results = append(results, result)
	}
	response := BulkOperationResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
		Summary: map[string]interface{}{
			"total_imported": len(clients),
			"success_rate":   float64(successCount) / float64(len(clients)) * 100,
			"format":         format,
		},
	}
	sendSuccess(w, "Bulk client import completed", response)
}
func parseCSVFromString(content string) ([]BulkClientData, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must have at least a header row and one data row")
	}
	var clients []BulkClientData
	headers := records[0]
	nameIndex := -1
	expiresIndex := -1
	quotaIndex := -1
	for i, header := range headers {
		switch strings.ToLower(header) {
		case "name", "client_name":
			nameIndex = i
		case "expires", "expires_at", "expiry":
			expiresIndex = i
		case "quota", "quota_gb":
			quotaIndex = i
		}
	}
	if nameIndex == -1 {
		return nil, fmt.Errorf("CSV must have a 'name' or 'client_name' column")
	}
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) <= nameIndex {
			continue
		}
		client := BulkClientData{
			Name: strings.TrimSpace(record[nameIndex]),
		}
		if expiresIndex >= 0 && len(record) > expiresIndex && record[expiresIndex] != "" {
			if parsedTime, err := time.Parse("2006-01-02", record[expiresIndex]); err == nil {
				client.ExpiresAt = parsedTime
			} else if parsedTime, err := time.Parse("2006-01-02 15:04:05", record[expiresIndex]); err == nil {
				client.ExpiresAt = parsedTime
			}
		}
		if quotaIndex >= 0 && len(record) > quotaIndex && record[quotaIndex] != "" {
			if quota, err := strconv.ParseInt(record[quotaIndex], 10, 64); err == nil {
				client.QuotaGB = quota
			}
		}
		if client.Name != "" {
			clients = append(clients, client)
		}
	}
	return clients, nil
}
func getClientConfigContent(clientName string) (string, error) {
	configPaths := []string{
		filepath.Join("/opt/irangate/clients", clientName+".ovpn"),
		filepath.Join("/etc/openvpn/clients", clientName+".ovpn"),
	}
	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("config file not found for client: %s", clientName)
}