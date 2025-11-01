package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	OpenVPNConfigDir = "/etc/openvpn"
)

type OpenVPNServerConfig struct {
	Filename    string            `json:"filename"`
	Name        string            `json:"name"`
	Port        int               `json:"port"`
	Protocol    string            `json:"protocol"`
	Cipher      string            `json:"cipher"`
	Auth        string            `json:"auth"`
	DevType     string            `json:"dev_type"`
	Status      string            `json:"status"`
	LastUpdated time.Time         `json:"last_updated"`
	Config      map[string]string `json:"config"`
	Errors      []string          `json:"errors,omitempty"`
}
type ConfigSyncStatus struct {
	LastSync      time.Time             `json:"last_sync"`
	TotalConfigs  int                   `json:"total_configs"`
	ActiveConfigs int                   `json:"active_configs"`
	ErrorConfigs  int                   `json:"error_configs"`
	Configs       []OpenVPNServerConfig `json:"configs"`
}

var (
	configCache      map[string]OpenVPNServerConfig
	configCacheMutex sync.RWMutex
	lastSyncTime     time.Time
)

func getOpenVPNConfigsHandler(w http.ResponseWriter, r *http.Request) {
	configs, err := discoverOpenVPNConfigs()
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to discover configs: %v", err), http.StatusInternalServerError)
		return
	}
	configCacheMutex.Lock()
	configCache = make(map[string]OpenVPNServerConfig)
	for _, config := range configs {
		configCache[config.Name] = config
	}
	lastSyncTime = time.Now()
	configCacheMutex.Unlock()
	sendSuccess(w, "Server configurations discovered successfully", map[string]interface{}{
		"configs":     configs,
		"total_count": len(configs),
		"last_sync":   lastSyncTime,
	})
}
func syncOpenVPNConfigsHandler(w http.ResponseWriter, r *http.Request) {
	configs, err := discoverOpenVPNConfigs()
	if err != nil {
		sendError(w, fmt.Sprintf("Sync failed: %v", err), http.StatusInternalServerError)
		return
	}
	configCacheMutex.Lock()
	configCache = make(map[string]OpenVPNServerConfig)
	for _, config := range configs {
		configCache[config.Name] = config
	}
	lastSyncTime = time.Now()
	configCacheMutex.Unlock()
	activeCount := 0
	errorCount := 0
	for _, config := range configs {
		switch config.Status {
		case "active":
			activeCount++
		case "error":
			errorCount++
		}
	}
	status := ConfigSyncStatus{
		LastSync:      lastSyncTime,
		TotalConfigs:  len(configs),
		ActiveConfigs: activeCount,
		ErrorConfigs:  errorCount,
		Configs:       configs,
	}
	sendSuccess(w, "Configuration sync completed", status)
}
func getOpenVPNConfigHandler(w http.ResponseWriter, r *http.Request) {
	name := getURLParam(r, "name")
	if name == "" {
		sendError(w, "Config name is required", http.StatusBadRequest)
		return
	}
	config, err := parseOpenVPNConfig(filepath.Join(OpenVPNConfigDir, name+".conf"))
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to parse config: %v", err), http.StatusNotFound)
		return
	}
	sendSuccess(w, "Configuration retrieved successfully", config)
}
func getOpenVPNSyncStatusHandler(w http.ResponseWriter, r *http.Request) {
	configCacheMutex.RLock()
	defer configCacheMutex.RUnlock()
	configs := make([]OpenVPNServerConfig, 0, len(configCache))
	activeCount := 0
	errorCount := 0
	for _, config := range configCache {
		configs = append(configs, config)
		switch config.Status {
		case "active":
			activeCount++
		case "error":
			errorCount++
		}
	}
	status := ConfigSyncStatus{
		LastSync:      lastSyncTime,
		TotalConfigs:  len(configCache),
		ActiveConfigs: activeCount,
		ErrorConfigs:  errorCount,
		Configs:       configs,
	}
	sendSuccess(w, "Sync status retrieved successfully", status)
}
func discoverOpenVPNConfigs() ([]OpenVPNServerConfig, error) {
	var configs []OpenVPNServerConfig
	if _, err := os.Stat(OpenVPNConfigDir); os.IsNotExist(err) {
		return configs, nil
	}
	entries, err := os.ReadDir(OpenVPNConfigDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", OpenVPNConfigDir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
			continue
		}
		configPath := filepath.Join(OpenVPNConfigDir, entry.Name())
		config, err := parseOpenVPNConfig(configPath)
		if err != nil {
			config = OpenVPNServerConfig{
				Filename:    entry.Name(),
				Name:        strings.TrimSuffix(entry.Name(), ".conf"),
				Status:      "error",
				LastUpdated: time.Now(),
				Errors:      []string{err.Error()},
			}
		}
		configs = append(configs, config)
	}
	return configs, nil
}
func parseOpenVPNConfig(filePath string) (OpenVPNServerConfig, error) {
	config := OpenVPNServerConfig{
		Filename:    filepath.Base(filePath),
		Name:        strings.TrimSuffix(filepath.Base(filePath), ".conf"),
		Status:      "active",
		Config:      make(map[string]string),
		LastUpdated: time.Now(),
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("cannot read file: %w", err)
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		directive := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")
		switch directive {
		case "port":
			if port, err := strconv.Atoi(value); err == nil {
				config.Port = port
			}
		case "proto":
			config.Protocol = value
		case "cipher":
			config.Cipher = value
		case "auth":
			config.Auth = value
		case "dev":
			config.DevType = value
		}
		config.Config[directive] = value
	}
	var errors []string
	if config.Port == 0 {
		errors = append(errors, "missing or invalid port directive")
	}
	if config.Protocol == "" {
		errors = append(errors, "missing protocol directive")
	}
	if config.Cipher == "" {
		errors = append(errors, "missing cipher directive")
	}
	if _, hasServer := config.Config["server"]; !hasServer {
		errors = append(errors, "missing server directive (not a server config)")
	}
	if len(errors) > 0 {
		config.Status = "error"
		config.Errors = errors
	}
	if info, err := os.Stat(filePath); err == nil {
		config.LastUpdated = info.ModTime()
	}
	return config, nil
}
func ValidateConfigFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}
	requiredDirectives := []string{"port", "proto", "server", "ca", "cert", "key", "dh"}
	contentStr := string(content)
	for _, directive := range requiredDirectives {
		pattern := fmt.Sprintf(`(?i)^%s\s+`, regexp.QuoteMeta(directive))
		matched, err := regexp.MatchString(pattern, contentStr)
		if err != nil || !matched {
			return fmt.Errorf("missing required directive: %s", directive)
		}
	}
	dangerousDirectives := []string{"script-security", "up", "down", "route-up", "route-pre-down"}
	for _, directive := range dangerousDirectives {
		pattern := fmt.Sprintf(`(?i)^%s\s+`, regexp.QuoteMeta(directive))
		matched, err := regexp.MatchString(pattern, contentStr)
		if err == nil && matched {
			return fmt.Errorf("potentially dangerous directive found: %s", directive)
		}
	}
	return nil
}
func getURLParam(r *http.Request, key string) string {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == key && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

type InboundConfig struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Port           string `json:"port"`
	Protocol       string `json:"protocol"`
	ClientTemplate string `json:"clientTemplate"`
	ServerTemplate string `json:"serverTemplate"`
}
type CreateConfigsRequest struct {
	Inbounds []InboundConfig `json:"inbounds"`
}

func createOpenVPNConfigsHandler(w http.ResponseWriter, r *http.Request) {
	sendError(w, "Multi-inbound feature is temporarily disabled. Coming soon!", http.StatusServiceUnavailable)
}
func GenerateOpenVPNConfig(inbound InboundConfig) string {
	port := inbound.Port
	if port == "" {
		port = "443"
	}
	protocol := inbound.Protocol
	if protocol == "" {
		protocol = "tcp"
	}
	config := fmt.Sprintf(`# OpenVPN Server Configuration for %s
port %s
proto %s
dev tun
ca /etc/openvpn/easy-rsa/pki/ca.crt
cert /etc/openvpn/easy-rsa/pki/issued/server.crt
key /etc/openvpn/easy-rsa/pki/private/server.key
dh /etc/openvpn/easy-rsa/pki/dh.pem
server 10.8.0.0 255.255.255.0
push "redirect-gateway def1 bypass-dhcp"
push "dhcp-option DNS 8.8.8.8"
push "dhcp-option DNS 1.1.1.1"
keepalive 10 120
cipher AES-256-GCM
auth SHA256
user nobody
group nogroup
persist-key
persist-tun
status openvpn-status-%s.log
log-append openvpn-%s.log
verb 3
`, inbound.Name, port, protocol, inbound.Name, inbound.Name)
	return config
}
func RestartOpenVPNService() error {
	cmd := exec.Command("systemctl", "restart", "openvpn")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("service", "openvpn", "restart")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restart OpenVPN service: %v", err)
		}
	}
	time.Sleep(2 * time.Second)
	return nil
}
func ManageOpenVPNService(configName, action string) error {
	serviceName := fmt.Sprintf("openvpn@%s.service", configName)
	switch action {
	case "enable":
		cmd := exec.Command("systemctl", "enable", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to enable service %s: %v", serviceName, err)
		}
		return nil
	case "start":
		cmd := exec.Command("systemctl", "start", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start service %s: %v", serviceName, err)
		}
		time.Sleep(2 * time.Second)
		return nil
	case "restart":
		cmd := exec.Command("systemctl", "restart", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restart service %s: %v", serviceName, err)
		}
		time.Sleep(2 * time.Second)
		return nil
	case "stop":
		cmd := exec.Command("systemctl", "stop", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop service %s: %v", serviceName, err)
		}
		return nil
	case "disable":
		cmd := exec.Command("systemctl", "disable", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to disable service %s: %v", serviceName, err)
		}
		return nil
	case "enable-start":
		if err := ManageOpenVPNService(configName, "enable"); err != nil {
			return err
		}
		return ManageOpenVPNService(configName, "start")
	case "stop-disable":
		if err := ManageOpenVPNService(configName, "stop"); err != nil {
			fmt.Printf("Warning: Failed to stop service %s: %v\n", serviceName, err)
		}
		return ManageOpenVPNService(configName, "disable")
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}
func SetupNATRules() error {
	externalIP, err := GetServerExternalIPForNAT()
	if err != nil {
		return fmt.Errorf("failed to get server external IP: %v", err)
	}
	if err := EnableIPForwarding(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %v", err)
	}
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING",
		"-s", "10.8.0.0/24", "!", "-d", "10.8.0.0/24",
		"-j", "SNAT", "--to", externalIP)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add NAT rule: %v", err)
	}
	forwardRules := [][]string{
		{"iptables", "-A", "FORWARD", "-s", "10.8.0.0/24", "-j", "ACCEPT"},
		{"iptables", "-A", "FORWARD", "-d", "10.8.0.0/24", "-j", "ACCEPT"},
		{"iptables", "-A", "FORWARD", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"},
	}
	for _, rule := range forwardRules {
		cmd := exec.Command(rule[0], rule[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add forward rule %v: %v", rule, err)
		}
	}
	return nil
}
func RemoveNATRules() error {
	externalIP, err := GetServerExternalIPForNAT()
	if err != nil {
		return fmt.Errorf("failed to get server external IP: %v", err)
	}
	cmd := exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING",
		"-s", "10.8.0.0/24", "!", "-d", "10.8.0.0/24",
		"-j", "SNAT", "--to", externalIP)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to remove NAT rule: %v\n", err)
	}
	forwardRules := [][]string{
		{"iptables", "-D", "FORWARD", "-s", "10.8.0.0/24", "-j", "ACCEPT"},
		{"iptables", "-D", "FORWARD", "-d", "10.8.0.0/24", "-j", "ACCEPT"},
		{"iptables", "-D", "FORWARD", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"},
	}
	for _, rule := range forwardRules {
		cmd := exec.Command(rule[0], rule[1:]...)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to remove forward rule %v: %v\n", rule, err)
		}
	}
	return nil
}
func GetServerExternalIPForNAT() (string, error) {
	cmd := exec.Command("ip", "route", "get", "8.8.8.8")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "src") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "src" && i+1 < len(parts) {
					return parts[i+1], nil
				}
			}
		}
	}
	return "", fmt.Errorf("could not determine external IP")
}
func EnableIPForwarding() error {
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %v", err)
	}
	cmd = exec.Command("sh", "-c", "echo 'net.ipv4.ip_forward=1' > /etc/sysctl.d/99-openvpn-forward.conf")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to make IP forwarding persistent: %v", err)
	}
	return nil
}

type UpdateConfigRequest struct {
	Inbound InboundConfig `json:"inbound"`
}

func updateOpenVPNConfigHandler(w http.ResponseWriter, r *http.Request) {
	sendError(w, "Multi-inbound feature is temporarily disabled. Coming soon!", http.StatusServiceUnavailable)
}
