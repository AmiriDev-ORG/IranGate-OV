package config
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
)
type ExportData struct {
	Version     string          `json:"version"`
	ExportTime  time.Time       `json:"export_time"`
	Settings    SystemSettings  `json:"settings"`
	Clients     []ClientData    `json:"clients"`
	ServerCerts CertificateData `json:"server_certificates"`
}
type SystemSettings struct {
	Server     ServerConfig  `json:"server"`
	API        APIConfig     `json:"api"`
	Monitoring MonitorConfig `json:"monitoring"`
	Backup     BackupConfig  `json:"backup"`
}
type BackupConfig struct {
	Enabled       bool   `json:"enabled"`
	Path          string `json:"path"`
	RetentionDays int    `json:"retention_days"`
}
type BandwidthLimit struct {
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
}
type ServerConfig struct {
	Port       int              `json:"port"`
	Protocol   string           `json:"protocol"`
	Subnet     string           `json:"subnet"`
	DNS        []string         `json:"dns"`
	Routes     []string         `json:"routes"`
	Management ManagementConfig `json:"management"`
}
type APIConfig struct {
	Port    int    `json:"port"`
	SSLCert string `json:"ssl_cert"`
	SSLKey  string `json:"ssl_key"`
}
type MonitorConfig struct {
	StatsInterval   int  `json:"stats_interval"`
	RetentionDays   int  `json:"retention_days"`
	EnableBandwidth bool `json:"enable_bandwidth"`
}
type ManagementConfig struct {
	Enable bool   `json:"enable"`
	Port   int    `json:"port"`
	Host   string `json:"host"`
}
type ClientData struct {
	Name           string          `json:"name"`
	Created        time.Time       `json:"created"`
	CreatedAt      time.Time       `json:"created_at"`
	ExpiryDate     *time.Time      `json:"expiry_date,omitempty"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty"`
	Active         bool            `json:"active"`
	IP             string          `json:"ip"`
	Cipher         string          `json:"cipher"`
	BandwidthLimit *BandwidthLimit `json:"bandwidth_limit,omitempty"`
	Routes         []string        `json:"routes,omitempty"`
}
type CertificateData struct {
	CAExpiry     time.Time `json:"ca_expiry"`
	ServerExpiry time.Time `json:"server_expiry"`
}
func (m *Manager) ExportSettings(filename string) error {
	export := ExportData{
		Version:    "1.0",
		ExportTime: time.Now(),
	}
	settings, err := m.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get system settings: %v", err)
	}
	systemSettings := SystemSettings{
		Server: ServerConfig{
			Port:     settings.ServerPort,
			Protocol: settings.Protocol,
			Subnet:   "10.8.0.0/24",
			DNS:      settings.DNS,
			Routes:   []string{},
		},
	}
	export.Settings = systemSettings
	clients, err := m.db.GetAllClients()
	if err != nil {
		return fmt.Errorf("failed to get clients: %v", err)
	}
	clientData := make([]ClientData, len(clients))
	for i, client := range clients {
		clientData[i] = ClientData{
			Name:      client.Name,
			CreatedAt: client.CreatedAt,
			ExpiresAt: &client.ExpiresAt,
			Active:    client.Active,
			IP:        client.IP,
			Cipher:    client.Cipher,
		}
	}
	export.Clients = clientData
	certs, err := m.getCertificateInfo()
	if err != nil {
		return fmt.Errorf("failed to get certificate info: %v", err)
	}
	export.ServerCerts = certs
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %v", err)
	}
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %v", err)
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %v", err)
	}
	return nil
}
func (m *Manager) ImportSettings(filename string, options ImportOptions) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read import file: %v", err)
	}
	var export ExportData
	if err := json.Unmarshal(data, &export); err != nil {
		return fmt.Errorf("failed to parse import data: %v", err)
	}
	if export.Version != "1.0" {
		return fmt.Errorf("unsupported export version: %s", export.Version)
	}
	defer func() {
	}()
	if options.ImportSettings {
		dbSettings := &database.Settings{
			ServerIP:   export.Settings.Server.Subnet,
			ServerPort: export.Settings.Server.Port,
			Protocol:   export.Settings.Server.Protocol,
			DNS:        export.Settings.Server.DNS,
			Cipher:     "AES-256-GCM",
			MTU:        1412,
		}
		if err := m.UpdateConfig(*dbSettings); err != nil {
			return fmt.Errorf("failed to import settings: %v", err)
		}
	}
	if options.ImportClients {
		if err := m.importClients(export.Clients, options.OverwriteExisting); err != nil {
			return fmt.Errorf("failed to import clients: %v", err)
		}
	}
	return nil
}
type ImportOptions struct {
	ImportSettings    bool `json:"import_settings"`
	ImportClients     bool `json:"import_clients"`
	OverwriteExisting bool `json:"overwrite_existing"`
}
func (m *Manager) importClients(clients []ClientData, overwrite bool) error {
	for _, client := range clients {
		_, err := m.db.GetClient(client.Name)
		exists := err == nil
		if exists && !overwrite {
			continue
		}
		dbClient := &database.Client{
			Name:      client.Name,
			CreatedAt: client.CreatedAt,
			ExpiresAt: *client.ExpiresAt,
			Active:    client.Active,
			IP:        client.IP,
			Cipher:    client.Cipher,
		}
		if err := m.db.UpdateClient(client.Name, dbClient); err != nil {
			return err
		}
	}
	return nil
}
func (m *Manager) getCertificateInfo() (CertificateData, error) {
	return CertificateData{
		CAExpiry:     time.Now().AddDate(10, 0, 0),
		ServerExpiry: time.Now().AddDate(1, 0, 0),
	}, nil
}