package client
import (
	"fmt"
	"sync"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/amiridev-org/irangate-ov/pkg/utils"
)
type BatchOperation struct {
	ClientNames []string               `json:"client_names"`
	Operation   string                 `json:"operation"`
	Config      *database.ClientConfig `json:"config,omitempty"`
}
type BatchResult struct {
	Successful []string     `json:"successful"`
	Failed     []BatchError `json:"failed"`
}
type BatchError struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}
func (m *Manager) BatchProcess(batch BatchOperation) (*BatchResult, error) {
	result := &BatchResult{
		Successful: make([]string, 0),
		Failed:     make([]BatchError, 0),
	}
	jobs := make(chan string, len(batch.ClientNames))
	results := make(chan struct {
		name string
		err  error
	}, len(batch.ClientNames))
	var wg sync.WaitGroup
	workerCount := min(10, len(batch.ClientNames))
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for name := range jobs {
				var err error
				switch batch.Operation {
				case "update":
					client, getErr := m.db.GetClient(name)
					if getErr != nil {
						err = getErr
					} else {
						if batch.Config != nil {
							client.Config = batch.Config
						}
						err = m.db.UpdateClient(name, client)
					}
				case "revoke":
					err = m.RevokeClient(name)
				case "delete":
					err = m.DeleteClient(name)
				default:
					err = fmt.Errorf("unknown operation: %s", batch.Operation)
				}
				results <- struct {
					name string
					err  error
				}{name, err}
			}
		}()
	}
	for _, name := range batch.ClientNames {
		jobs <- name
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(results)
	}()
	for res := range results {
		if res.err != nil {
			result.Failed = append(result.Failed, BatchError{
				Name:  res.name,
				Error: res.err.Error(),
			})
		} else {
			result.Successful = append(result.Successful, res.name)
		}
	}
	return result, nil
}
type ExpiryChecker struct {
	manager   *Manager
	interval  time.Duration
	stopChan  chan struct{}
	lastCheck time.Time
}
func NewExpiryChecker(manager *Manager, interval time.Duration) *ExpiryChecker {
	return &ExpiryChecker{
		manager:  manager,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}
func (ec *ExpiryChecker) Start() {
	ticker := time.NewTicker(ec.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				ec.checkExpirations()
			case <-ec.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}
func (ec *ExpiryChecker) Stop() {
	close(ec.stopChan)
}
func (ec *ExpiryChecker) checkExpirations() {
	clients, err := ec.manager.ListClients()
	if err != nil {
		utils.GetLogger().Errorf("Failed to list clients for expiry check: %v", err)
		return
	}
	now := time.Now()
	ec.lastCheck = now
	for _, client := range clients {
		if client.ExpiryDate != nil && now.After(*client.ExpiryDate) && client.Status == "active" {
			if err := ec.manager.RevokeClient(client.Name); err != nil {
				utils.GetLogger().Errorf("Failed to revoke expired client %s: %v", client.Name, err)
				continue
			}
			client.Status = "expired"
			if err := ec.manager.db.UpdateClient(client.Name, &client); err != nil {
				utils.GetLogger().Errorf("Failed to update expired client status %s: %v", client.Name, err)
			}
			utils.GetLogger().Infof("Client %s expired and revoked", client.Name)
		}
	}
}