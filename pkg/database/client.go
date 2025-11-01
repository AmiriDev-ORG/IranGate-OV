package database
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)
func (db *DB) GetAllClients() ([]Client, error) {
	return db.GetClients()
}
func (db *DB) SaveClients(clients []Client) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	clientsDir := filepath.Join(db.path, ClientsDir)
	for _, client := range clients {
		clientPath := filepath.Join(clientsDir, client.Name+".json")
		data, err := json.MarshalIndent(client, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal client %s: %v", client.Name, err)
		}
		if err := os.WriteFile(clientPath, data, 0600); err != nil {
			return fmt.Errorf("failed to write client file %s: %v", client.Name, err)
		}
	}
	return nil
}
func (db *DB) GetClient(name string) (*Client, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	clientPath := filepath.Join(db.path, ClientsDir, name+".json")
	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("client %s not found", name)
	}
	data, err := os.ReadFile(clientPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read client file: %v", err)
	}
	var client Client
	if err := json.Unmarshal(data, &client); err != nil {
		return nil, fmt.Errorf("failed to parse client file: %v", err)
	}
	return &client, nil
}
func (db *DB) UpdateClient(name string, client *Client) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	clientPath := filepath.Join(db.path, ClientsDir, name+".json")
	tempPath := clientPath + ".tmp"
	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		return fmt.Errorf("client %s not found", name)
	}
	data, err := json.MarshalIndent(client, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client data: %v", err)
	}
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary file: %v", err)
	}
	if err := os.Rename(tempPath, clientPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to update client file: %v", err)
	}
	return nil
}