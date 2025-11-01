package main
import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"
)
type AsyncOperation struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Result    interface{}            `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
type AsyncOperationManager struct {
	operations map[string]*AsyncOperation
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}
func NewAsyncOperationManager() *AsyncOperationManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &AsyncOperationManager{
		operations: make(map[string]*AsyncOperation),
		ctx:        ctx,
		cancel:     cancel,
	}
}
func (m *AsyncOperationManager) StartAsyncOperation(operationType string, metadata map[string]interface{}, operationFunc func() (interface{}, error)) string {
	operationID := fmt.Sprintf("op_%d", time.Now().UnixNano())
	operation := &AsyncOperation{
		ID:        operationID,
		Type:      operationType,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  metadata,
	}
	m.mutex.Lock()
	m.operations[operationID] = operation
	m.mutex.Unlock()
	go m.executeOperation(operation, operationFunc)
	return operationID
}
func (m *AsyncOperationManager) executeOperation(operation *AsyncOperation, operationFunc func() (interface{}, error)) {
	m.mutex.Lock()
	operation.Status = "running"
	operation.UpdatedAt = time.Now()
	m.mutex.Unlock()
	result, err := operationFunc()
	m.mutex.Lock()
	defer m.mutex.Unlock()
	operation.UpdatedAt = time.Now()
	if err != nil {
		operation.Status = "failed"
		operation.Error = err.Error()
	} else {
		operation.Status = "completed"
		operation.Result = result
	}
}
func (m *AsyncOperationManager) GetOperationStatus(operationID string) (*AsyncOperation, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	operation, exists := m.operations[operationID]
	return operation, exists
}
func (m *AsyncOperationManager) ListOperations() []*AsyncOperation {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	operations := make([]*AsyncOperation, 0, len(m.operations))
	for _, op := range m.operations {
		operations = append(operations, op)
	}
	return operations
}
func (m *AsyncOperationManager) CleanupOldOperations(maxAge time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	cutoff := time.Now().Add(-maxAge)
	for id, operation := range m.operations {
		if operation.UpdatedAt.Before(cutoff) {
			delete(m.operations, id)
		}
	}
}
func (m *AsyncOperationManager) AsyncClientList() string {
	return m.StartAsyncOperation("client_list", nil, func() (interface{}, error) {
		cmd := exec.Command("irangate", "client", "list", "--json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to get clients: %v", err)
		}
		var clients []interface{}
		if err := json.Unmarshal(output, &clients); err != nil {
			return nil, fmt.Errorf("failed to parse clients: %v", err)
		}
		return clients, nil
	})
}
func (m *AsyncOperationManager) AsyncServerStatus() string {
	return m.StartAsyncOperation("server_status", nil, func() (interface{}, error) {
		cmd := exec.Command("irangate", "status")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to get server status: %v", err)
		}
		status := string(output)
		isRunning := contains(status, "running") || contains(status, "active")
		return map[string]interface{}{
			"status":     status,
			"is_running": isRunning,
		}, nil
	})
}
func (m *AsyncOperationManager) AsyncConfigUpdate(config map[string]string) string {
	return m.StartAsyncOperation("config_update", map[string]interface{}{
		"config_keys": len(config),
	}, func() (interface{}, error) {
		results := make(map[string]interface{})
		for key, value := range config {
			cmd := exec.Command("irangate", "config", "set", key, value)
			output, err := cmd.CombinedOutput()
			if err != nil {
				results[key] = map[string]interface{}{
					"success": false,
					"error":   string(output),
				}
			} else {
				results[key] = map[string]interface{}{
					"success": true,
					"output":  string(output),
				}
			}
		}
		return results, nil
	})
}
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr)))
}
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
var asyncManager = NewAsyncOperationManager()
func StartAsyncCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-asyncManager.ctx.Done():
				return
			case <-ticker.C:
				asyncManager.CleanupOldOperations(24 * time.Hour)
			}
		}
	}()
}