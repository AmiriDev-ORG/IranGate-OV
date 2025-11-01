package errors
import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"go.uber.org/zap"
)
type RecoveryManager struct {
	services    map[string]*ServiceHealth
	mu          sync.RWMutex
	config      RecoveryConfig
	subscribers []chan RecoveryEvent
}
type ServiceHealth struct {
	Name           string
	Status         ServiceStatus
	LastCheck      time.Time
	FailureCount   int
	RecoveryCount  int
	LastError      error
	LastRecovery   time.Time
	IsRecovering   bool
	CircuitBreaker *CircuitBreaker
}
type ServiceStatus int
const (
	StatusHealthy ServiceStatus = iota
	StatusDegraded
	StatusFailed
	StatusRecovering
)
type RecoveryConfig struct {
	HealthCheckInterval time.Duration
	RecoveryTimeout     time.Duration
	MaxRecoveryAttempts int
	AutoRecover         bool
}
type RecoveryEvent struct {
	Service   string
	Status    ServiceStatus
	Error     error
	Timestamp time.Time
}
func NewRecoveryManager(config RecoveryConfig) *RecoveryManager {
	return &RecoveryManager{
		services:    make(map[string]*ServiceHealth),
		config:      config,
		subscribers: make([]chan RecoveryEvent, 0),
	}
}
func (rm *RecoveryManager) RegisterService(name string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.services[name] = &ServiceHealth{
		Name:           name,
		Status:         StatusHealthy,
		LastCheck:      time.Now(),
		CircuitBreaker: NewCircuitBreaker(3, 1*time.Minute),
	}
}
func (rm *RecoveryManager) ReportFailure(ctx context.Context, serviceName string, err error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	service, exists := rm.services[serviceName]
	if !exists {
		return
	}
	service.FailureCount++
	service.LastError = err
	service.Status = StatusFailed
	rm.notifySubscribers(RecoveryEvent{
		Service:   serviceName,
		Status:    StatusFailed,
		Error:     err,
		Timestamp: time.Now(),
	})
	if rm.config.AutoRecover {
		go rm.attemptRecovery(ctx, serviceName)
	}
}
func (rm *RecoveryManager) attemptRecovery(ctx context.Context, serviceName string) {
	rm.mu.Lock()
	service := rm.services[serviceName]
	if service == nil || service.Status != StatusFailed {
		rm.mu.Unlock()
		return
	}
	service.IsRecovering = true
	service.Status = StatusRecovering
	rm.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, rm.config.RecoveryTimeout)
	defer cancel()
	err := WithRetry(ctx, func(ctx context.Context) error {
		return rm.executeRecoveryProcedure(ctx, serviceName)
	}, DefaultRetryConfig)
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if err != nil {
		service.Status = StatusFailed
		LogError(ctx, fmt.Errorf("recovery failed for service %s: %w", serviceName, err))
	} else {
		service.Status = StatusHealthy
		service.RecoveryCount++
		service.LastRecovery = time.Now()
	}
	service.IsRecovering = false
	rm.notifySubscribers(RecoveryEvent{
		Service:   serviceName,
		Status:    service.Status,
		Error:     err,
		Timestamp: time.Now(),
	})
}
func (rm *RecoveryManager) executeRecoveryProcedure(ctx context.Context, serviceName string) error {
	switch serviceName {
	case "database":
		return rm.recoverDatabase(ctx)
	case "vpn":
		return rm.recoverVPNService(ctx)
	case "api":
		return rm.recoverAPIService(ctx)
	default:
		return fmt.Errorf("no recovery procedure defined for service: %s", serviceName)
	}
}
func (rm *RecoveryManager) recoverDatabase(ctx context.Context) error {
	logger := GetLogger(ctx)
	logger.Info("Starting database recovery procedure")
	dbPath := "/opt/irangate/database"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		logger.Info("Database directory does not exist, creating...")
		if err := os.MkdirAll(dbPath, 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %v", err)
		}
		logger.Info("Database directory created successfully")
	}
	dbFiles := []string{"settings.json", "cron.json"}
	for _, file := range dbFiles {
		filePath := filepath.Join(dbPath, file)
		if err := rm.validateDatabaseFile(filePath); err != nil {
			logger.Warn("Database file is corrupted", zap.String("file", file), zap.Error(err))
			if err := rm.restoreDatabaseFileFromBackup(file); err != nil {
				logger.Error("Failed to restore from backup", zap.String("file", file), zap.Error(err))
				if err := rm.createDefaultDatabaseFile(filePath); err != nil {
					return fmt.Errorf("failed to create default database file %s: %v", file, err)
				}
				logger.Info("Created new default database file", zap.String("file", file))
			} else {
				logger.Info("Successfully restored from backup", zap.String("file", file))
			}
		}
	}
	clientsDir := filepath.Join(dbPath, "clients")
	if _, err := os.Stat(clientsDir); os.IsNotExist(err) {
		logger.Info("Clients directory does not exist, creating...")
		if err := os.MkdirAll(clientsDir, 0755); err != nil {
			return fmt.Errorf("failed to create clients directory: %v", err)
		}
	}
	db, err := database.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	if _, err := db.GetClients(); err != nil {
		return fmt.Errorf("database client operations failed: %v", err)
	}
	if _, err := db.GetSettings(); err != nil {
		return fmt.Errorf("database settings operations failed: %v", err)
	}
	logger.Info("Database recovery completed successfully")
	return nil
}
func (rm *RecoveryManager) validateDatabaseFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}
	return nil
}
func (rm *RecoveryManager) restoreDatabaseFileFromBackup(filename string) error {
	backupDir := "/opt/irangate/database/backups"
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %v", err)
	}
	var latestBackup string
	var latestTime time.Time
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") && entry.Name() != filename {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestBackup = entry.Name()
			}
		}
	}
	if latestBackup == "" {
		return fmt.Errorf("no backup files found")
	}
	backupPath := filepath.Join(backupDir, latestBackup)
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %v", err)
	}
	var backupMap map[string]interface{}
	if err := json.Unmarshal(backupData, &backupMap); err != nil {
		return fmt.Errorf("failed to parse backup file: %v", err)
	}
	fileData, exists := backupMap[filename]
	if !exists {
		return fmt.Errorf("file %s not found in backup", filename)
	}
	restoredData, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal restored data: %v", err)
	}
	targetPath := filepath.Join("/opt/irangate/database", filename)
	if err := os.WriteFile(targetPath, restoredData, 0600); err != nil {
		return fmt.Errorf("failed to write restored file: %v", err)
	}
	return nil
}
func (rm *RecoveryManager) createDefaultDatabaseFile(filePath string) error {
	var defaultData interface{}
	filename := filepath.Base(filePath)
	switch filename {
	case "settings.json":
		defaultData = database.Settings{
			ServerIP:       "0.0.0.0",
			ServerPort:     1194,
			Protocol:       "udp",
			Cipher:         "AES-256-GCM",
			MTU:            1412,
			DNS:            []string{"1.1.1.1", "8.8.8.8"},
			IPv6Enabled:    false,
			AutoConfigMode: "dynamic",
			BackupPath:     "/opt/irangate/database/backups",
			LogLevel:       3,
			Version:        "1.0.0",
			InstallDate:    time.Now(),
		}
	case "cron.json":
		defaultData = struct {
			Jobs []database.CronJob `json:"jobs"`
		}{Jobs: []database.CronJob{}}
	default:
		return fmt.Errorf("unknown database file type: %s", filename)
	}
	data, err := json.MarshalIndent(defaultData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default data: %v", err)
	}
	return os.WriteFile(filePath, data, 0600)
}
func (rm *RecoveryManager) recoverVPNService(ctx context.Context) error {
	logger := GetLogger(ctx)
	logger.Info("Starting VPN service recovery procedure")
	cmd := exec.Command("systemctl", "is-active", "openvpn")
	output, err := cmd.Output()
	if err != nil {
		logger.Warn("Failed to check OpenVPN service status, attempting to start...")
	} else {
		status := strings.TrimSpace(string(output))
		if status == "active" {
			logger.Info("OpenVPN service is already running")
			return nil
		}
	}
	configPath := "/etc/openvpn/server/server.conf"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Warn("OpenVPN server configuration not found, attempting to regenerate...")
		logger.Warn("Server configuration regeneration requires manual intervention")
	}
	pkiPath := "/etc/openvpn/easy-rsa"
	caPath := filepath.Join(pkiPath, "pki", "ca.crt")
	if _, err := os.Stat(caPath); os.IsNotExist(err) {
		logger.Warn("PKI certificates not found, attempting to initialize...")
		if err := rm.initializePKI(); err != nil {
			logger.Error("Failed to initialize PKI", zap.Error(err))
			return fmt.Errorf("failed to initialize PKI: %v", err)
		}
		logger.Info("PKI initialized successfully")
	}
	logger.Info("Starting OpenVPN service...")
	cmd = exec.Command("systemctl", "start", "openvpn")
	if err := cmd.Run(); err != nil {
		logger.Error("Failed to start OpenVPN service", zap.Error(err))
		serviceNames := []string{"openvpn@server", "openvpn-server"}
		for _, serviceName := range serviceNames {
			cmd = exec.Command("systemctl", "start", serviceName)
			if err := cmd.Run(); err == nil {
				logger.Info("Successfully started OpenVPN service", zap.String("service", serviceName))
				return nil
			}
		}
		return fmt.Errorf("failed to start OpenVPN service with any known service name")
	}
	time.Sleep(2 * time.Second)
	cmd = exec.Command("systemctl", "is-active", "openvpn")
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("OpenVPN service failed to start: %v", err)
	}
	status := strings.TrimSpace(string(output))
	if status != "active" {
		return fmt.Errorf("OpenVPN service is not active, status: %s", status)
	}
	logger.Info("VPN service recovery completed successfully")
	return nil
}
func (rm *RecoveryManager) initializePKI() error {
	pkiPath := "/etc/openvpn/easy-rsa"
	cmd := exec.Command("easyrsa", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("easy-rsa not available: %v", err)
	}
	cmd = exec.Command("easyrsa", "init-pki")
	cmd.Dir = pkiPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize PKI: %v", err)
	}
	cmd = exec.Command("easyrsa", "build-ca", "nopass")
	cmd.Dir = pkiPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build CA: %v", err)
	}
	cmd = exec.Command("easyrsa", "build-server-full", "server", "nopass")
	cmd.Dir = pkiPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate server certificate: %v", err)
	}
	cmd = exec.Command("easyrsa", "gen-dh")
	cmd.Dir = pkiPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate DH parameters: %v", err)
	}
	return nil
}
func (rm *RecoveryManager) recoverAPIService(ctx context.Context) error {
	logger := GetLogger(ctx)
	logger.Info("Starting API service recovery procedure")
	binaryPath := os.Args[0]
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("IRANGATE binary not found at %s", binaryPath)
	}
	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("IRANGATE CLI test failed: %v", err)
	}
	if len(output) == 0 {
		return fmt.Errorf("IRANGATE CLI returned empty output")
	}
	db, err := database.New("")
	if err != nil {
		return fmt.Errorf("database connectivity test failed: %v", err)
	}
	if _, err := db.GetClients(); err != nil {
		return fmt.Errorf("database client operations test failed: %v", err)
	}
	logger.Info("API service recovery completed successfully")
	return nil
}
func (rm *RecoveryManager) AttemptRecovery(ctx context.Context, serviceName string) error {
	return rm.executeRecoveryProcedure(ctx, serviceName)
}
func (rm *RecoveryManager) Subscribe() chan RecoveryEvent {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	ch := make(chan RecoveryEvent, 100)
	rm.subscribers = append(rm.subscribers, ch)
	return ch
}
func (rm *RecoveryManager) Unsubscribe(ch chan RecoveryEvent) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	for i, sub := range rm.subscribers {
		if sub == ch {
			rm.subscribers = append(rm.subscribers[:i], rm.subscribers[i+1:]...)
			close(ch)
			return
		}
	}
}
func (rm *RecoveryManager) notifySubscribers(event RecoveryEvent) {
	for _, ch := range rm.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}
func (rm *RecoveryManager) GetServiceHealth(serviceName string) *ServiceHealth {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.services[serviceName]
}
func (rm *RecoveryManager) GetAllServicesHealth() map[string]*ServiceHealth {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	result := make(map[string]*ServiceHealth)
	for name, health := range rm.services {
		result[name] = health
	}
	return result
}