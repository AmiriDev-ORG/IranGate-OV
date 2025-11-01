package traffic
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)
type Config struct {
	DataDir           string `json:"data_dir"`
	CollectionInterval int    `json:"collection_interval_seconds"`
	RetentionDays     int    `json:"retention_days"`
	RollupInterval    int    `json:"rollup_interval_hours"`
	StatusLogPath     string `json:"status_log_path"`
	EnableQuotaTracking bool    `json:"enable_quota_tracking"`
	AlertThreshold      float64 `json:"alert_threshold_percent"`
	BatchSize           int `json:"batch_size"`
	MaxConcurrentWrites int `json:"max_concurrent_writes"`
	EnableTelegramAlerts bool `json:"enable_telegram_alerts"`
	EnableEmailAlerts    bool `json:"enable_email_alerts"`
	EnableAnomalyDetection  bool   `json:"enable_anomaly_detection"`
	AnomalyBaselineDays     int    `json:"anomaly_baseline_days"`
	AnomalyThresholdMultiplier float64 `json:"anomaly_threshold_multiplier"`
	CacheEnabled   bool `json:"cache_enabled"`
	CacheTTL       int  `json:"cache_ttl_seconds"`
	CacheMaxSize   int  `json:"cache_max_size_mb"`
	DefaultExportFormat string `json:"default_export_format"`
	MaxExportRecords    int    `json:"max_export_records"`
}
func DefaultConfig() *Config {
	return &Config{
		DataDir:                "/opt/irangate/traffic",
		CollectionInterval:     10,
		RetentionDays:          90,
		RollupInterval:         1,
		StatusLogPath:          "/etc/openvpn/openvpn-status.log",
		EnableQuotaTracking:    true,
		AlertThreshold:         80.0,
		BatchSize:              100,
		MaxConcurrentWrites:    10,
		EnableTelegramAlerts:   false,
		EnableEmailAlerts:      false,
		EnableAnomalyDetection: true,
		AnomalyBaselineDays:    30,
		AnomalyThresholdMultiplier: 3.0,
		CacheEnabled:           true,
		CacheTTL:               300,
		CacheMaxSize:           512,
		DefaultExportFormat:    "json",
		MaxExportRecords:       1000000,
	}
}
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			config := DefaultConfig()
			if err := SaveConfig(path, config); err != nil {
				return nil, fmt.Errorf("failed to create default config: %v", err)
			}
			return config, nil
		}
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &config, nil
}
func SaveConfig(path string, config *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
func (c *Config) Validate() error {
	if c.CollectionInterval < 1 {
		return fmt.Errorf("collection_interval_seconds must be >= 1")
	}
	if c.RetentionDays < 1 {
		return fmt.Errorf("retention_days must be >= 1")
	}
	if c.RollupInterval < 1 {
		return fmt.Errorf("rollup_interval_hours must be >= 1")
	}
	if c.AlertThreshold < 0 || c.AlertThreshold > 100 {
		return fmt.Errorf("alert_threshold_percent must be 0-100")
	}
	if c.BatchSize < 1 {
		return fmt.Errorf("batch_size must be >= 1")
	}
	if c.MaxConcurrentWrites < 1 {
		return fmt.Errorf("max_concurrent_writes must be >= 1")
	}
	return nil
}