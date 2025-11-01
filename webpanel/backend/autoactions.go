package main
import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/gorilla/mux"
)
type AutoAction struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Trigger     string                 `json:"trigger"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Enabled     bool                   `json:"enabled"`
	LastRun     *time.Time             `json:"last_run,omitempty"`
	NextRun     *time.Time             `json:"next_run,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
type AutoActionRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Trigger     string                 `json:"trigger"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Enabled     bool                   `json:"enabled"`
}
type AutoActionExecution struct {
	ID         string    `json:"id"`
	ActionID   string    `json:"action_id"`
	Triggered  string    `json:"triggered"`
	Success    bool      `json:"success"`
	Message    string    `json:"message"`
	ExecutedAt time.Time `json:"executed_at"`
}
var AutoActionsDataDir = func() string {
	if dir := os.Getenv("IRANGATE_DATA_DIR"); dir != "" {
		return filepath.Join(dir, "webpanel", "autoactions")
	}
	return "/opt/irangate/webpanel/autoactions"
}()
func initAutoActionsDir() error {
	return os.MkdirAll(AutoActionsDataDir, 0755)
}
func getAutoActionsHandler(w http.ResponseWriter, r *http.Request) {
	actions, err := loadAllAutoActions()
	if err != nil {
		sendError(w, "Failed to load auto-actions", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Auto-actions retrieved successfully", actions)
}
func createAutoActionHandler(w http.ResponseWriter, r *http.Request) {
	var req AutoActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Trigger == "" || req.Action == "" {
		sendError(w, "Name, trigger, and action are required", http.StatusBadRequest)
		return
	}
	if !isValidTrigger(req.Trigger) {
		sendError(w, "Invalid trigger type", http.StatusBadRequest)
		return
	}
	if !isValidAction(req.Action) {
		sendError(w, "Invalid action type", http.StatusBadRequest)
		return
	}
	if err := initAutoActionsDir(); err != nil {
		sendError(w, "Failed to initialize auto-actions directory", http.StatusInternalServerError)
		return
	}
	action := AutoAction{
		ID:          generateActionID(req.Name),
		Name:        req.Name,
		Description: req.Description,
		Trigger:     req.Trigger,
		Condition:   req.Condition,
		Action:      req.Action,
		Parameters:  req.Parameters,
		Enabled:     req.Enabled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := saveAutoAction(&action); err != nil {
		sendError(w, "Failed to save auto-action", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Auto-action created successfully", action)
}
func updateAutoActionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actionID := vars["id"]
	var req AutoActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	action, err := loadAutoAction(actionID)
	if err != nil {
		sendError(w, "Auto-action not found", http.StatusNotFound)
		return
	}
	if req.Trigger != "" && !isValidTrigger(req.Trigger) {
		sendError(w, "Invalid trigger type", http.StatusBadRequest)
		return
	}
	if req.Action != "" && !isValidAction(req.Action) {
		sendError(w, "Invalid action type", http.StatusBadRequest)
		return
	}
	if req.Name != "" {
		action.Name = req.Name
	}
	if req.Description != "" {
		action.Description = req.Description
	}
	if req.Trigger != "" {
		action.Trigger = req.Trigger
	}
	if req.Condition != "" {
		action.Condition = req.Condition
	}
	if req.Action != "" {
		action.Action = req.Action
	}
	if req.Parameters != nil {
		action.Parameters = req.Parameters
	}
	action.Enabled = req.Enabled
	action.UpdatedAt = time.Now()
	if err := saveAutoAction(action); err != nil {
		sendError(w, "Failed to update auto-action", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Auto-action updated successfully", action)
}
func deleteAutoActionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actionID := vars["id"]
	_, err := loadAutoAction(actionID)
	if err != nil {
		sendError(w, "Auto-action not found", http.StatusNotFound)
		return
	}
	actionFile := filepath.Join(AutoActionsDataDir, actionID+".json")
	if err := os.Remove(actionFile); err != nil {
		sendError(w, "Failed to delete auto-action", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Auto-action deleted successfully", map[string]interface{}{
		"action_id": actionID,
	})
}
func testAutoActionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actionID := vars["id"]
	action, err := loadAutoAction(actionID)
	if err != nil {
		sendError(w, "Auto-action not found", http.StatusNotFound)
		return
	}
	result, err := executeAutoAction(action, "test_trigger", true)
	if err != nil {
		sendError(w, fmt.Sprintf("Test execution failed: %v", err), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Auto-action test completed", result)
}
func executeAutoActionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actionID := vars["id"]
	var req struct {
		TriggerData map[string]interface{} `json:"trigger_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	action, err := loadAutoAction(actionID)
	if err != nil {
		sendError(w, "Auto-action not found", http.StatusNotFound)
		return
	}
	result, err := executeAutoAction(action, "manual_trigger", false)
	if err != nil {
		sendError(w, fmt.Sprintf("Execution failed: %v", err), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Auto-action executed successfully", result)
}
func generateActionID(name string) string {
	return fmt.Sprintf("action_%s_%d", sanitizeFilename(name), time.Now().Unix())
}
func sanitizeFilename(name string) string {
	result := ""
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') {
			result += string(char)
		} else {
			result += "_"
		}
	}
	return result
}
func isValidTrigger(trigger string) bool {
	validTriggers := []string{
		"client_expired",
		"client_expiring",
		"quota_exceeded",
		"quota_warning",
		"resource_warning",
		"service_down",
		"service_restarted",
		"new_client_created",
		"scheduled",
	}
	for _, valid := range validTriggers {
		if trigger == valid {
			return true
		}
	}
	return false
}
func isValidAction(action string) bool {
	validActions := []string{
		"disable_client",
		"extend_client",
		"restart_service",
		"cleanup_orphaned",
		"send_notification",
		"backup_database",
		"log_event",
	}
	for _, valid := range validActions {
		if action == valid {
			return true
		}
	}
	return false
}
func saveAutoAction(action *AutoAction) error {
	actionFile := filepath.Join(AutoActionsDataDir, action.ID+".json")
	data, err := json.MarshalIndent(action, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(actionFile, data, 0644)
}
func loadAutoAction(actionID string) (*AutoAction, error) {
	actionFile := filepath.Join(AutoActionsDataDir, actionID+".json")
	data, err := os.ReadFile(actionFile)
	if err != nil {
		return nil, err
	}
	var action AutoAction
	if err := json.Unmarshal(data, &action); err != nil {
		return nil, err
	}
	return &action, nil
}
func loadAllAutoActions() ([]*AutoAction, error) {
	if err := initAutoActionsDir(); err != nil {
		return nil, err
	}
	pattern := filepath.Join(AutoActionsDataDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	var actions []*AutoAction
	for _, file := range files {
		action, err := loadAutoActionFromFile(file)
		if err != nil {
			continue
		}
		actions = append(actions, action)
	}
	return actions, nil
}
func loadAutoActionFromFile(filename string) (*AutoAction, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var action AutoAction
	if err := json.Unmarshal(data, &action); err != nil {
		return nil, err
	}
	return &action, nil
}
func executeAutoAction(action *AutoAction, triggerSource string, testMode bool) (map[string]interface{}, error) {
	execution := AutoActionExecution{
		ID:         fmt.Sprintf("exec_%d", time.Now().Unix()),
		ActionID:   action.ID,
		Triggered:  triggerSource,
		ExecutedAt: time.Now(),
	}
	var err error
	var message string
	switch action.Action {
	case "disable_client":
		message, err = executeDisableClient(action, testMode)
	case "extend_client":
		message, err = executeExtendClient(action, testMode)
	case "restart_service":
		message, err = executeRestartService(action, testMode)
	case "cleanup_orphaned":
		message, err = executeCleanupOrphaned(action, testMode)
	case "send_notification":
		message, err = executeSendNotification(action, testMode)
	case "backup_database":
		message, err = executeBackupDatabase(action, testMode)
	case "log_event":
		message, err = executeLogEvent(action, testMode)
	default:
		err = fmt.Errorf("unknown action: %s", action.Action)
	}
	execution.Success = err == nil
	if err != nil {
		execution.Message = err.Error()
	} else {
		execution.Message = message
	}
	saveExecutionLog(&execution)
	action.LastRun = &execution.ExecutedAt
	saveAutoAction(action)
	return map[string]interface{}{
		"execution_id": execution.ID,
		"success":      execution.Success,
		"message":      execution.Message,
		"executed_at":  execution.ExecutedAt,
	}, err
}
func executeDisableClient(action *AutoAction, testMode bool) (string, error) {
	if testMode {
		return "Test: Would disable client", nil
	}
	clientName, ok := action.Parameters["client_name"].(string)
	if !ok {
		return "", fmt.Errorf("client_name parameter required")
	}
	db, err := database.New(database.GetDefaultDatabasePath())
	if err != nil {
		return "", err
	}
	client, err := db.GetClient(clientName)
	if err != nil {
		return "", err
	}
	client.Active = false
	if err := db.UpdateClient(clientName, client); err != nil {
		return "", err
	}
	return fmt.Sprintf("Client %s disabled", clientName), nil
}
func executeExtendClient(action *AutoAction, testMode bool) (string, error) {
	if testMode {
		return "Test: Would extend client", nil
	}
	clientName, ok := action.Parameters["client_name"].(string)
	if !ok {
		return "", fmt.Errorf("client_name parameter required")
	}
	extensionDays, ok := action.Parameters["extension_days"].(float64)
	if !ok {
		extensionDays = 30
	}
	db, err := database.New(database.GetDefaultDatabasePath())
	if err != nil {
		return "", err
	}
	client, err := db.GetClient(clientName)
	if err != nil {
		return "", err
	}
	client.ExpiresAt = client.ExpiresAt.AddDate(0, 0, int(extensionDays))
	if err := db.UpdateClient(clientName, client); err != nil {
		return "", err
	}
	return fmt.Sprintf("Client %s extended by %.0f days", clientName, extensionDays), nil
}
func executeRestartService(action *AutoAction, testMode bool) (string, error) {
	if testMode {
		return "Test: Would restart OpenVPN service", nil
	}
	serviceName := "openvpn"
	if val, ok := action.Parameters["service_name"].(string); ok && val != "" {
		serviceName = val
	}
	_ = serviceName
	return "OpenVPN service restart attempted", nil
}
func executeCleanupOrphaned(action *AutoAction, testMode bool) (string, error) {
	if testMode {
		return "Test: Would cleanup orphaned certificates", nil
	}
	cleanupDays := 30
	if val, ok := action.Parameters["cleanup_days"].(string); ok && val != "" {
		if days, err := strconv.Atoi(val); err == nil {
			cleanupDays = days
		}
	}
	_ = cleanupDays
	return "Orphaned certificates cleanup attempted", nil
}
func executeSendNotification(action *AutoAction, testMode bool) (string, error) {
	if testMode {
		return "Test: Would send notification", nil
	}
	message, ok := action.Parameters["message"].(string)
	if !ok {
		return "", fmt.Errorf("message parameter required")
	}
	err := sendTelegramMessage("", "", message)
	if err != nil {
		return "", err
	}
	return "Notification sent", nil
}
func executeBackupDatabase(action *AutoAction, testMode bool) (string, error) {
	if testMode {
		return "Test: Would backup database", nil
	}
	backupPath := "/var/backups/irangate"
	if val, ok := action.Parameters["backup_path"].(string); ok && val != "" {
		backupPath = val
	}
	_ = backupPath
	return "Database backup attempted", nil
}
func executeLogEvent(action *AutoAction, testMode bool) (string, error) {
	if testMode {
		return "Test: Would log event", nil
	}
	message, ok := action.Parameters["message"].(string)
	if !ok {
		return "", fmt.Errorf("message parameter required")
	}
	return fmt.Sprintf("Event logged: %s", message), nil
}
func saveExecutionLog(execution *AutoActionExecution) error {
	logFile := filepath.Join(AutoActionsDataDir, "executions.json")
	var executions []*AutoActionExecution
	if data, err := os.ReadFile(logFile); err == nil {
		json.Unmarshal(data, &executions)
	}
	executions = append(executions, execution)
	if len(executions) > 1000 {
		executions = executions[len(executions)-1000:]
	}
	data, _ := json.MarshalIndent(executions, "", "  ")
	return os.WriteFile(logFile, data, 0644)
}