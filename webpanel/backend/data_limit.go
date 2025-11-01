package main
import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)
type DataLimitConfig struct {
	ClientName string    `json:"client_name"`
	LimitBytes int64     `json:"limit_bytes"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type ClientDataUsage struct {
	ClientName    string `json:"client_name"`
	BytesSent     int64  `json:"bytes_sent"`
	BytesReceived int64  `json:"bytes_received"`
	TotalBytes    int64  `json:"total_bytes"`
}
var DataLimitConfigPath = func() string {
	if dir := os.Getenv("IRANGATE_DATA_DIR"); dir != "" {
		return filepath.Join(dir, "webpanel", "data_limits.json")
	}
	return "/opt/irangate/webpanel/data_limits.json"
}()
func LoadDataLimits() (map[string]DataLimitConfig, error) {
	limits := make(map[string]DataLimitConfig)
	data, err := os.ReadFile(DataLimitConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return limits, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &limits); err != nil {
		return nil, err
	}
	return limits, nil
}
func SaveDataLimits(limits map[string]DataLimitConfig) error {
	if err := os.MkdirAll(filepath.Dir(DataLimitConfigPath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(limits, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(DataLimitConfigPath, data, 0644)
}
func SetClientDataLimit(clientName string, limitBytes int64, enabled bool) error {
	limits, err := LoadDataLimits()
	if err != nil {
		return err
	}
	config, exists := limits[clientName]
	if !exists {
		config = DataLimitConfig{
			ClientName: clientName,
			CreatedAt:  time.Now(),
		}
	}
	config.LimitBytes = limitBytes
	config.Enabled = enabled
	config.UpdatedAt = time.Now()
	limits[clientName] = config
	return SaveDataLimits(limits)
}
func GetClientDataUsage(clientName string) (*ClientDataUsage, error) {
	cmd := exec.Command("irangate", "client", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get client list: %v", err)
	}
	var clients []map[string]interface{}
	if err := json.Unmarshal(output, &clients); err != nil {
		return nil, fmt.Errorf("failed to parse client list: %v", err)
	}
	for _, client := range clients {
		if name, ok := client["name"].(string); ok && name == clientName {
			usage := &ClientDataUsage{
				ClientName: clientName,
			}
			if sent, ok := client["bytes_sent"].(float64); ok {
				usage.BytesSent = int64(sent)
			}
			if received, ok := client["bytes_received"].(float64); ok {
				usage.BytesReceived = int64(received)
			}
			usage.TotalBytes = usage.BytesSent + usage.BytesReceived
			return usage, nil
		}
	}
	return nil, fmt.Errorf("client not found: %s", clientName)
}
func CheckAndEnforceDataLimits() error {
	log.Println("Starting data limit enforcement check...")
	limits, err := LoadDataLimits()
	if err != nil {
		return fmt.Errorf("failed to load data limits: %v", err)
	}
	deletedClients := []string{}
	for clientName, config := range limits {
		if !config.Enabled || config.LimitBytes <= 0 {
			continue
		}
		usage, err := GetClientDataUsage(clientName)
		if err != nil {
			log.Printf("Warning: Failed to get usage for client %s: %v", clientName, err)
			continue
		}
		if usage.TotalBytes >= config.LimitBytes {
			log.Printf("Client %s exceeded data limit (%d bytes used, %d bytes limit)",
				clientName, usage.TotalBytes, config.LimitBytes)
			if err := deleteClientAndRevoke(clientName); err != nil {
				log.Printf("Error deleting client %s: %v", clientName, err)
				continue
			}
			deletedClients = append(deletedClients, clientName)
			delete(limits, clientName)
			log.Printf("Successfully deleted client %s due to data limit exceeded", clientName)
		}
	}
	if len(deletedClients) > 0 {
		if err := SaveDataLimits(limits); err != nil {
			log.Printf("Warning: Failed to save updated data limits: %v", err)
		}
	}
	log.Printf("Data limit enforcement check completed. Deleted %d clients: %v",
		len(deletedClients), deletedClients)
	return nil
}
func deleteClientAndRevoke(clientName string) error {
	cmd := exec.Command("irangate", "client", "remove", clientName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove client: %v, output: %s", err, string(output))
	}
	dataDir := os.Getenv("IRANGATE_DATA_DIR")
	if dataDir == "" {
		dataDir = "/opt/irangate"
	}
	clientDataPath := filepath.Join(dataDir, "database", "clients", clientName+".json")
	if err := os.Remove(clientDataPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to remove client data file %s: %v", clientDataPath, err)
	}
	return nil
}
func StartDataLimitScanner(intervalMinutes int) {
	if intervalMinutes <= 0 {
		intervalMinutes = 60
	}
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()
	log.Printf("Data limit scanner started (checking every %d minutes)", intervalMinutes)
	if err := CheckAndEnforceDataLimits(); err != nil {
		log.Printf("Error in data limit check: %v", err)
	}
	for range ticker.C {
		if err := CheckAndEnforceDataLimits(); err != nil {
			log.Printf("Error in data limit check: %v", err)
		}
	}
}