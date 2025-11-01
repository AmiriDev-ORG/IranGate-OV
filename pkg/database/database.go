package database
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)
const (
	DefaultDatabasePath = "/opt/irangate/database"
	ClientsDir          = "clients"
	UsersFile           = "users.json"
	SettingsFile        = "settings.json"
	CronFile            = "cron.json"
)
func GetDefaultDatabasePath() string {
	if os.Getenv("OS") == "Windows_NT" || os.Getenv("GOOS") == "windows" {
		return "./database"
	}
	return DefaultDatabasePath
}
type DB struct {
	path  string
	mutex sync.RWMutex
}
type Client struct {
	Name             string     `json:"name"`
	CreatedAt        time.Time  `json:"created_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	ExpiryDate       *time.Time `json:"expiry_date,omitempty"`
	Active           bool       `json:"active"`
	IP               string     `json:"ip"`
	Cipher           string     `json:"cipher"`
	DataUsedMB       int64      `json:"data_used_mb"`
	LastConnection   *time.Time `json:"last_connection"`
	Upload           int64      `json:"upload"`
	Download         int64      `json:"download"`
	LastUpdate       time.Time  `json:"last_update"`
	BytesReceived    uint64     `json:"bytes_received"`
	BytesSent        uint64     `json:"bytes_sent"`
	Status           string     `json:"status"`
	BandwidthQuotaGB int64      `json:"bandwidth_quota_gb,omitempty"`
	QuotaResetDate   time.Time  `json:"quota_reset_date,omitempty"`
	QuotaEnabled     bool       `json:"quota_enabled"`
	TrafficLimitBytes int64     `json:"traffic_limit_bytes,omitempty"`
	Config           *ClientConfig   `json:"config,omitempty"`
	BandwidthUsage   *BandwidthUsage `json:"bandwidth_usage,omitempty"`
}
type ClientConfig struct {
	BandwidthLimit *BandwidthLimit `json:"bandwidth_limit,omitempty"`
	Routes         []string        `json:"routes,omitempty"`
	DNS            []string        `json:"dns,omitempty"`
	PushOptions    []string        `json:"push_options,omitempty"`
}
type BandwidthLimit struct {
	Upload   int64 `json:"upload_kbps"`
	Download int64 `json:"download_kbps"`
}
type BandwidthUsage struct {
	BytesIn    int64         `json:"bytes_in"`
	BytesOut   int64         `json:"bytes_out"`
	LastUpdate time.Time     `json:"last_update"`
	History    []UsageRecord `json:"history"`
}
type UsageRecord struct {
	Timestamp time.Time `json:"timestamp"`
	BytesIn   int64     `json:"bytes_in"`
	BytesOut  int64     `json:"bytes_out"`
}
type Device struct {
	ID         string    `json:"id"`
	ClientName string    `json:"client_name"`
	DeviceName string    `json:"device_name"`
	DeviceType string    `json:"device_type"`
	LastSeen   time.Time `json:"last_seen"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}
type Settings struct {
	ServerIP       string     `json:"server_ip"`
	ServerPort     int        `json:"server_port"`
	Protocol       string     `json:"protocol"`
	Cipher         string     `json:"cipher"`
	MTU            int        `json:"mtu"`
	DNS            []string   `json:"dns"`
	IPv6Enabled    bool       `json:"ipv6_enabled"`
	AutoConfigMode string     `json:"auto_config_mode"`
	BackupPath     string     `json:"backup_path"`
	LogLevel       int        `json:"log_level"`
	Version        string     `json:"version"`
	InstallDate    time.Time  `json:"install_date"`
	LastBackup     *time.Time `json:"last_backup"`
}
type CronJob struct {
	ID       string                 `json:"id"`
	Schedule string                 `json:"schedule"`
	Action   string                 `json:"action"`
	Params   map[string]interface{} `json:"params"`
	LastRun  *time.Time             `json:"last_run"`
	Enabled  bool                   `json:"enabled"`
}
func New(path string) (*DB, error) {
	if path == "" {
		path = GetDefaultDatabasePath()
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}
	for _, dir := range []string{"history", "backups", ClientsDir} {
		if err := os.MkdirAll(filepath.Join(path, dir), 0755); err != nil {
			return nil, fmt.Errorf("failed to create %s directory: %v", dir, err)
		}
	}
	db := &DB{path: path}
	if err := db.migrateLegacyUsersFile(); err != nil {
		fmt.Printf("Warning: Failed to migrate legacy users.json: %v\n", err)
	}
	return db, nil
}
func (db *DB) GetPath() string {
	return db.path
}
func (db *DB) GetClients() ([]Client, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	clientsDir := filepath.Join(db.path, ClientsDir)
	entries, err := os.ReadDir(clientsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Client{}, nil
		}
		return nil, fmt.Errorf("failed to read clients directory: %v", err)
	}
	var clients []Client
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		clientPath := filepath.Join(clientsDir, entry.Name())
		data, err := os.ReadFile(clientPath)
		if err != nil {
			continue
		}
		var client Client
		if err := json.Unmarshal(data, &client); err != nil {
			continue
		}
		clients = append(clients, client)
	}
	return clients, nil
}
func (db *DB) AddClient(client Client) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	clientPath := filepath.Join(db.path, ClientsDir, client.Name+".json")
	data, err := json.MarshalIndent(client, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client data: %v", err)
	}
	file, err := os.OpenFile(clientPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("client %s already exists", client.Name)
		}
		return fmt.Errorf("failed to create client file: %v", err)
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		file.Close()
		os.Remove(clientPath)
		return fmt.Errorf("failed to write client data: %v", err)
	}
	return nil
}
func (db *DB) RemoveClient(name string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	clientPath := filepath.Join(db.path, ClientsDir, name+".json")
	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		return fmt.Errorf("client %s not found", name)
	}
	if err := os.Remove(clientPath); err != nil {
		return fmt.Errorf("failed to delete client file: %v", err)
	}
	return nil
}
func (db *DB) GetSettings() (*Settings, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	var settings Settings
	if err := db.readJSON(SettingsFile, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}
func (db *DB) UpdateSettings(settings Settings) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	return db.writeJSON(SettingsFile, settings)
}
func (db *DB) GetCronJobs() ([]CronJob, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	var data struct {
		Jobs []CronJob `json:"jobs"`
	}
	if err := db.readJSON(CronFile, &data); err != nil {
		return nil, err
	}
	return data.Jobs, nil
}
func (db *DB) AddCronJob(job CronJob) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	var data struct {
		Jobs []CronJob `json:"jobs"`
	}
	if err := db.readJSON(CronFile, &data); err != nil {
		return err
	}
	for _, j := range data.Jobs {
		if j.ID == job.ID {
			return fmt.Errorf("job %s already exists", job.ID)
		}
	}
	data.Jobs = append(data.Jobs, job)
	return db.writeJSON(CronFile, data)
}
func (db *DB) RemoveCronJob(id string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	var data struct {
		Jobs []CronJob `json:"jobs"`
	}
	if err := db.readJSON(CronFile, &data); err != nil {
		return err
	}
	for i, job := range data.Jobs {
		if job.ID == id {
			data.Jobs = append(data.Jobs[:i], data.Jobs[i+1:]...)
			return db.writeJSON(CronFile, data)
		}
	}
	return fmt.Errorf("job %s not found", id)
}
func (db *DB) Backup() error {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	backupDir := filepath.Join(db.path, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("backup_%s.json", timestamp))
	backup := make(map[string]interface{})
	for _, file := range []string{UsersFile, SettingsFile, CronFile} {
		var data interface{}
		if err := db.readJSON(file, &data); err != nil {
			return fmt.Errorf("failed to read %s: %v", file, err)
		}
		backup[file] = data
	}
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup data: %v", err)
	}
	if err := os.WriteFile(backupFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %v", err)
	}
	return nil
}
func (db *DB) Restore(backupFile string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	data, err := os.ReadFile(backupFile)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %v", err)
	}
	var backup map[string]interface{}
	if err := json.Unmarshal(data, &backup); err != nil {
		return fmt.Errorf("failed to unmarshal backup data: %v", err)
	}
	for file, content := range backup {
		data, err := json.MarshalIndent(content, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal %s: %v", file, err)
		}
		if err := os.WriteFile(filepath.Join(db.path, file), data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %v", file, err)
		}
	}
	return nil
}
func (db *DB) readJSON(file string, v interface{}) error {
	data, err := os.ReadFile(filepath.Join(db.path, file))
	if err != nil {
		if os.IsNotExist(err) {
			defaultData := initializeEmptyStructure(file)
			if err := db.writeJSON(file, defaultData); err != nil {
				return fmt.Errorf("failed to create default %s: %v", file, err)
			}
			data, err = os.ReadFile(filepath.Join(db.path, file))
			if err != nil {
				return fmt.Errorf("failed to read newly created %s: %v", file, err)
			}
		} else {
			return fmt.Errorf("failed to read %s: %v", file, err)
		}
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %v", file, err)
	}
	return nil
}
func (db *DB) writeJSON(file string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %v", file, err)
	}
	if err := os.WriteFile(filepath.Join(db.path, file), data, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %v", file, err)
	}
	return nil
}
func (db *DB) readJSONUnsafe(file string, v interface{}) error {
	data, err := os.ReadFile(filepath.Join(db.path, file))
	if err != nil {
		if os.IsNotExist(err) {
			defaultData := initializeEmptyStructure(file)
			if err := db.writeJSONUnsafe(file, defaultData); err != nil {
				return fmt.Errorf("failed to create default %s: %v", file, err)
			}
			data, err = os.ReadFile(filepath.Join(db.path, file))
			if err != nil {
				return fmt.Errorf("failed to read newly created %s: %v", file, err)
			}
		} else {
			return fmt.Errorf("failed to read %s: %v", file, err)
		}
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %v", file, err)
	}
	return nil
}
func (db *DB) writeJSONUnsafe(file string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %v", file, err)
	}
	if err := os.WriteFile(filepath.Join(db.path, file), data, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %v", file, err)
	}
	return nil
}
func initializeEmptyStructure(file string) interface{} {
	switch file {
	case UsersFile:
		return struct {
			Clients []Client `json:"clients"`
		}{
			Clients: []Client{},
		}
	case SettingsFile:
		return Settings{
			ServerIP:       "0.0.0.0",
			ServerPort:     1194,
			Protocol:       "udp",
			Cipher:         "AES-256-GCM",
			MTU:            1412,
			DNS:            []string{"1.1.1.1", "8.8.8.8"},
			IPv6Enabled:    false,
			AutoConfigMode: "dynamic",
			BackupPath:     "/opt/irangate/database/backups/",
			LogLevel:       3,
			Version:        "1.0.0",
			InstallDate:    time.Now(),
		}
	case CronFile:
		return struct {
			Jobs []CronJob `json:"jobs"`
		}{
			Jobs: []CronJob{},
		}
	default:
		return struct{}{}
	}
}
func (db *DB) migrateLegacyUsersFile() error {
	legacyPath := filepath.Join(db.path, UsersFile)
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		return nil
	}
	data, err := os.ReadFile(legacyPath)
	if err != nil {
		return fmt.Errorf("failed to read legacy users.json: %v", err)
	}
	var legacyData struct {
		Clients []Client `json:"clients"`
	}
	if err := json.Unmarshal(data, &legacyData); err != nil {
		return fmt.Errorf("failed to parse legacy users.json: %v", err)
	}
	clientsDir := filepath.Join(db.path, ClientsDir)
	for _, client := range legacyData.Clients {
		clientPath := filepath.Join(clientsDir, client.Name+".json")
		if _, err := os.Stat(clientPath); err == nil {
			continue
		}
		clientData, err := json.MarshalIndent(client, "", "  ")
		if err != nil {
			continue
		}
		if err := os.WriteFile(clientPath, clientData, 0600); err != nil {
			continue
		}
	}
	backupPath := filepath.Join(db.path, "users.json.bak")
	if err := os.Rename(legacyPath, backupPath); err != nil {
	}
	return nil
}