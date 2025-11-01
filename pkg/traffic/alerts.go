package traffic
import (
	"fmt"
	"sync"
	"time"
)
type AlertManager struct {
	mu            sync.RWMutex
	enabled       bool
	telegramEnabled bool
	emailEnabled    bool
	threshold      float64
	lastAlerts     map[string]time.Time
	cooldown       time.Duration
	notifyFunc     AlertNotifyFunc
}
type AlertNotifyFunc func(alert *Alert) error
type Alert struct {
	ClientName    string
	AlertType     string
	Message       string
	Severity      string
	Timestamp     time.Time
	Data          map[string]interface{}
}
func NewAlertManager(threshold float64, telegramEnabled, emailEnabled bool) *AlertManager {
	return &AlertManager{
		enabled:         true,
		telegramEnabled: telegramEnabled,
		emailEnabled:    emailEnabled,
		threshold:       threshold,
		lastAlerts:      make(map[string]time.Time),
		cooldown:        1 * time.Hour,
	}
}
func (am *AlertManager) SetNotifyFunc(fn AlertNotifyFunc) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.notifyFunc = fn
}
func (am *AlertManager) CheckQuotaAlerts(usage *QuotaUsage) error {
	if !am.enabled {
		return nil
	}
	if usage.Percentage >= am.threshold {
		alert := &Alert{
			ClientName: usage.ClientName,
			AlertType:  "quota",
			Message:    fmt.Sprintf("Client %s has used %.1f%% of their quota", usage.ClientName, usage.Percentage),
			Severity:   "warning",
			Timestamp:  time.Now(),
			Data: map[string]interface{}{
				"percentage":      usage.Percentage,
				"total_used":      usage.TotalUsed,
				"quota_limit":     usage.QuotaLimit,
				"remaining_bytes": usage.RemainingBytes,
				"reset_date":      usage.ResetDate,
			},
		}
		if usage.Percentage >= 100.0 {
			alert.Severity = "critical"
			alert.Message = fmt.Sprintf("CRITICAL: Client %s has exceeded their quota", usage.ClientName)
		}
		return am.sendAlert(alert)
	}
	return nil
}
func (am *AlertManager) CheckAnomalyAlerts(clientName string, anomalies []string) error {
	if !am.enabled || len(anomalies) == 0 {
		return nil
	}
	alert := &Alert{
		ClientName: clientName,
		AlertType:  "anomaly",
		Message:    fmt.Sprintf("Unusual traffic patterns detected for %s", clientName),
		Severity:   "warning",
		Timestamp:  time.Now(),
		Data: map[string]interface{}{
			"anomalies": anomalies,
		},
	}
	return am.sendAlert(alert)
}
func (am *AlertManager) CheckThresholdAlerts(clientName string, currentUsage uint64, threshold uint64) error {
	if !am.enabled {
		return nil
	}
	if currentUsage >= threshold {
		alert := &Alert{
			ClientName: clientName,
			AlertType:  "threshold",
			Message:    fmt.Sprintf("Client %s has exceeded usage threshold", clientName),
			Severity:   "warning",
			Timestamp:  time.Now(),
			Data: map[string]interface{}{
				"current_usage": currentUsage,
				"threshold":     threshold,
			},
		}
		return am.sendAlert(alert)
	}
	return nil
}
func (am *AlertManager) sendAlert(alert *Alert) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	alertKey := fmt.Sprintf("%s_%s", alert.ClientName, alert.AlertType)
	lastAlert, exists := am.lastAlerts[alertKey]
	if exists && time.Since(lastAlert) < am.cooldown {
		return nil
	}
	am.lastAlerts[alertKey] = time.Now()
	if am.notifyFunc != nil {
		return am.notifyFunc(alert)
	}
	fmt.Printf("[ALERT] %s - %s: %s\n", alert.Severity, alert.ClientName, alert.Message)
	return nil
}
func (am *AlertManager) GetActiveAlerts() []string {
	am.mu.RLock()
	defer am.mu.RUnlock()
	active := make([]string, 0)
	for alertKey, lastAlert := range am.lastAlerts {
		if time.Since(lastAlert) < am.cooldown {
			active = append(active, alertKey)
		}
	}
	return active
}
func (am *AlertManager) ClearAlert(clientName, alertType string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	alertKey := fmt.Sprintf("%s_%s", clientName, alertType)
	delete(am.lastAlerts, alertKey)
}
type MonitorManager struct {
	mu            sync.RWMutex
	analyzer       *Analyzer
	alertManager   *AlertManager
	enabled        bool
	running        bool
	stopChan       chan struct{}
	wg             sync.WaitGroup
	checkInterval  time.Duration
}
func NewMonitorManager(analyzer *Analyzer, alertManager *AlertManager) *MonitorManager {
	return &MonitorManager{
		analyzer:      analyzer,
		alertManager:  alertManager,
		enabled:       true,
		stopChan:      make(chan struct{}),
		checkInterval: 5 * time.Minute,
	}
}
func (mm *MonitorManager) Start() error {
	if mm.running {
		return fmt.Errorf("monitoring already running")
	}
	mm.running = true
	mm.wg.Add(1)
	go mm.monitoringLoop()
	return nil
}
func (mm *MonitorManager) Stop() {
	if !mm.running {
		return
	}
	mm.running = false
	close(mm.stopChan)
	mm.wg.Wait()
}
func (mm *MonitorManager) monitoringLoop() {
	defer mm.wg.Done()
	ticker := time.NewTicker(mm.checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-mm.stopChan:
			return
		case <-ticker.C:
			if err := mm.checkAllClients(); err != nil {
				fmt.Printf("Error checking clients: %v\n", err)
			}
		}
	}
}
func (mm *MonitorManager) checkAllClients() error {
	clients, err := mm.getAllClients()
	if err != nil {
		return err
	}
	for _, clientName := range clients {
		if err := mm.checkClient(clientName); err != nil {
			fmt.Printf("Error checking client %s: %v\n", clientName, err)
		}
	}
	return nil
}
func (mm *MonitorManager) checkClient(clientName string) error {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -1)
	anomalies, err := mm.analyzer.DetectAnomalies(clientName, startTime, endTime)
	if err == nil && len(anomalies) > 0 {
		if err := mm.alertManager.CheckAnomalyAlerts(clientName, anomalies); err != nil {
			fmt.Printf("Error sending anomaly alert: %v\n", err)
		}
	}
	return nil
}
func (mm *MonitorManager) getAllClients() ([]string, error) {
	return []string{}, nil
}
func (mm *MonitorManager) SetCheckInterval(interval time.Duration) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.checkInterval = interval
}
func (mm *MonitorManager) SetEnabled(enabled bool) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.enabled = enabled
}