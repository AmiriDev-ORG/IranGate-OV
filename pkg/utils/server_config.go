package utils
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)
const (
	EasyRSAPath = "/etc/openvpn/easy-rsa"
)
type ServerConfig struct {
	ServerIP     string `json:"server_ip"`
	ServerDomain string `json:"server_domain"`
	AutoDetectIP bool   `json:"auto_detect_ip"`
	LastUpdated  string `json:"last_updated"`
	DetectedIPs  struct {
		Primary string `json:"primary"`
		Public  string `json:"public"`
		Local   string `json:"local"`
	} `json:"detected_ips"`
	Settings struct {
		UseDomain      bool `json:"use_domain"`
		FallbackToIP   bool `json:"fallback_to_ip"`
		UpdateInterval int  `json:"update_interval"`
	} `json:"settings"`
}
func GetServerIPFromConfig(configPath string) (string, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := ServerConfig{
			ServerIP:     "YOUR_SERVER_IP",
			ServerDomain: "",
			AutoDetectIP: true,
			LastUpdated:  time.Now().Format(time.RFC3339),
		}
		if err := SaveServerConfig(configPath, &defaultConfig); err != nil {
			return "YOUR_SERVER_IP", err
		}
		return "YOUR_SERVER_IP", nil
	}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "YOUR_SERVER_IP", fmt.Errorf("failed to read config file: %v", err)
	}
	var config ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return "YOUR_SERVER_IP", fmt.Errorf("failed to parse config file: %v", err)
	}
	if config.Settings.UseDomain && config.ServerDomain != "" {
		return config.ServerDomain, nil
	}
	if config.ServerIP != "" && config.ServerIP != "YOUR_SERVER_IP" {
		return config.ServerIP, nil
	}
	if config.AutoDetectIP {
		detectedIP := GetServerIP()
		if detectedIP != "" && detectedIP != "YOUR_SERVER_IP" {
			config.ServerIP = detectedIP
			config.DetectedIPs.Primary = detectedIP
			config.LastUpdated = time.Now().Format(time.RFC3339)
			SaveServerConfig(configPath, &config)
			return detectedIP, nil
		}
	}
	return "YOUR_SERVER_IP", nil
}
func SaveServerConfig(configPath string, config *ServerConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	return nil
}
func UpdateServerIP(configPath, newIP string) error {
	config, err := LoadServerConfig(configPath)
	if err != nil {
		return err
	}
	config.ServerIP = newIP
	config.LastUpdated = time.Now().Format(time.RFC3339)
	return SaveServerConfig(configPath, config)
}
func LoadServerConfig(configPath string) (*ServerConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	var config ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}
	return &config, nil
}