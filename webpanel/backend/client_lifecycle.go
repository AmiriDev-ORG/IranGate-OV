package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var lifecycleMutex sync.Mutex

type RevocationLog struct {
	ClientName string    `json:"client_name"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
}

var RevocationLogPath = func() string {
	if dir := os.Getenv("IRANGATE_DATA_DIR"); dir != "" {
		return filepath.Join(dir, "logs", "revocations.json")
	}
	return "/opt/irangate/logs/revocations.json"
}()

func logRevocation(clientName, reason string, success bool, errMsg string) {
	logEntry := RevocationLog{
		ClientName: clientName,
		Reason:     reason,
		Timestamp:  time.Now(),
		Success:    success,
		Error:      errMsg,
	}

	var logs []RevocationLog
	if data, err := os.ReadFile(RevocationLogPath); err == nil {
		json.Unmarshal(data, &logs)
	}

	logs = append(logs, logEntry)

	os.MkdirAll(filepath.Dir(RevocationLogPath), 0755)
	if data, err := json.MarshalIndent(logs, "", "  "); err == nil {
		os.WriteFile(RevocationLogPath, data, 0644)
	}

	if success {
		log.Printf("✅ Client %s revoked: %s", clientName, reason)
	} else {
		log.Printf("❌ Failed to revoke client %s: %s - %s", clientName, reason, errMsg)
	}
}

func RevokeClientCertificate(clientName string, reason string) error {
	lifecycleMutex.Lock()
	defer lifecycleMutex.Unlock()

	log.Printf("Revoking certificate for client: %s (reason: %s)", clientName, reason)

	cmd := exec.Command("easyrsa", "revoke", clientName)
	cmd.Dir = "/etc/openvpn/easy-rsa"
	cmd.Env = append(os.Environ(), "EASYRSA_BATCH=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if string(output) != "" && (strings.Contains(string(output), "already revoked") || strings.Contains(string(output), "Not Found")) {
			log.Printf("Certificate for %s may already be revoked or not found", clientName)
		} else {
			errMsg := fmt.Sprintf("easyrsa revoke failed: %v, output: %s", err, string(output))
			logRevocation(clientName, reason, false, errMsg)
			return fmt.Errorf(errMsg)
		}
	}

	genCRLCmd := exec.Command("easyrsa", "gen-crl")
	genCRLCmd.Dir = "/etc/openvpn/easy-rsa"
	genCRLCmd.Env = append(os.Environ(), "EASYRSA_BATCH=1")

	if output, err := genCRLCmd.CombinedOutput(); err != nil {
		errMsg := fmt.Sprintf("easyrsa gen-crl failed: %v, output: %s", err, string(output))
		logRevocation(clientName, reason, false, errMsg)
		return fmt.Errorf(errMsg)
	}

	crlSource := "/etc/openvpn/easy-rsa/pki/crl.pem"
	crlDest := "/etc/openvpn/crl.pem"
	if _, err := exec.Command("cp", crlSource, crlDest).CombinedOutput(); err != nil {
		log.Printf("Warning: Failed to copy CRL to OpenVPN directory: %v", err)
	}

	exec.Command("systemctl", "reload", "openvpn@server").Run()

	logRevocation(clientName, reason, true, "")
	return nil
}

func DeactivateClient(clientName string) error {
	dataDir := os.Getenv("IRANGATE_DATA_DIR")
	if dataDir == "" {
		dataDir = "/opt/irangate"
	}

	clientPath := filepath.Join(dataDir, "database", "clients", clientName+".json")
	
	data, err := os.ReadFile(clientPath)
	if err != nil {
		return fmt.Errorf("failed to read client data: %v", err)
	}

	var clientData map[string]interface{}
	if err := json.Unmarshal(data, &clientData); err != nil {
		return fmt.Errorf("failed to parse client data: %v", err)
	}

	clientData["active"] = false
	clientData["status"] = "Inactive"

	newData, err := json.MarshalIndent(clientData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client data: %v", err)
	}

	if err := os.WriteFile(clientPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to save client data: %v", err)
	}

	return nil
}

func ActivateClient(clientName string) error {
	dataDir := os.Getenv("IRANGATE_DATA_DIR")
	if dataDir == "" {
		dataDir = "/opt/irangate"
	}

	clientPath := filepath.Join(dataDir, "database", "clients", clientName+".json")
	
	data, err := os.ReadFile(clientPath)
	if err != nil {
		return fmt.Errorf("failed to read client data: %v", err)
	}

	var clientData map[string]interface{}
	if err := json.Unmarshal(data, &clientData); err != nil {
		return fmt.Errorf("failed to parse client data: %v", err)
	}

	clientData["active"] = true
	clientData["status"] = "Active"

	newData, err := json.MarshalIndent(clientData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal client data: %v", err)
	}

	if err := os.WriteFile(clientPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to save client data: %v", err)
	}

	return nil
}

func CheckExpiredClients() error {
	log.Println("🔍 Checking for expired clients...")

	dataDir := os.Getenv("IRANGATE_DATA_DIR")
	if dataDir == "" {
		dataDir = "/opt/irangate"
	}

	clientsDir := filepath.Join(dataDir, "database", "clients")
	entries, err := os.ReadDir(clientsDir)
	if err != nil {
		return fmt.Errorf("failed to read clients directory: %v", err)
	}

	now := time.Now()
	expiredCount := 0

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		clientPath := filepath.Join(clientsDir, entry.Name())
		data, err := os.ReadFile(clientPath)
		if err != nil {
			log.Printf("Warning: Failed to read %s: %v", entry.Name(), err)
			continue
		}

		var clientData map[string]interface{}
		if err := json.Unmarshal(data, &clientData); err != nil {
			log.Printf("Warning: Failed to parse %s: %v", entry.Name(), err)
			continue
		}

		active, _ := clientData["active"].(bool)
		if !active {
			continue
		}

		clientName, _ := clientData["name"].(string)
		expiresAtStr, ok := clientData["expires_at"].(string)
		if !ok || expiresAtStr == "" {
			continue
		}

		expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			log.Printf("Warning: Failed to parse expiration date for %s: %v", clientName, err)
			continue
		}

		if now.After(expiresAt) {
			log.Printf("⏰ Client %s has expired (expiry: %s)", clientName, expiresAt.Format("2006-01-02 15:04:05"))
			
			if err := RevokeClientCertificate(clientName, "expired"); err != nil {
				log.Printf("Error revoking expired client %s: %v", clientName, err)
				continue
			}

			if err := DeactivateClient(clientName); err != nil {
				log.Printf("Error deactivating expired client %s: %v", clientName, err)
			}

			expiredCount++
		}
	}

	log.Printf("✅ Expiration check complete. Found %d expired clients", expiredCount)
	return nil
}

func CheckTrafficLimits() error {
	log.Println("🔍 Checking traffic limits...")

	dataDir := os.Getenv("IRANGATE_DATA_DIR")
	if dataDir == "" {
		dataDir = "/opt/irangate"
	}

	clientsDir := filepath.Join(dataDir, "database", "clients")
	entries, err := os.ReadDir(clientsDir)
	if err != nil {
		return fmt.Errorf("failed to read clients directory: %v", err)
	}

	exceededCount := 0

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		clientPath := filepath.Join(clientsDir, entry.Name())
		data, err := os.ReadFile(clientPath)
		if err != nil {
			log.Printf("Warning: Failed to read %s: %v", entry.Name(), err)
			continue
		}

		var clientData map[string]interface{}
		if err := json.Unmarshal(data, &clientData); err != nil {
			log.Printf("Warning: Failed to parse %s: %v", entry.Name(), err)
			continue
		}

		active, _ := clientData["active"].(bool)
		if !active {
			continue
		}

		clientName, _ := clientData["name"].(string)
		
		trafficLimit, ok := clientData["traffic_limit_bytes"].(float64)
		if !ok || trafficLimit <= 0 {
			continue
		}

		bytesReceived, _ := clientData["bytes_received"].(float64)
		bytesSent, _ := clientData["bytes_sent"].(float64)
		totalBytes := int64(bytesReceived + bytesSent)

		if totalBytes >= int64(trafficLimit) {
			log.Printf("📊 Client %s exceeded traffic limit (%d bytes used, %d bytes limit)",
				clientName, totalBytes, int64(trafficLimit))

			if err := RevokeClientCertificate(clientName, "traffic_limit_exceeded"); err != nil {
				log.Printf("Error revoking client %s: %v", clientName, err)
				continue
			}

			if err := DeactivateClient(clientName); err != nil {
				log.Printf("Error deactivating client %s: %v", clientName, err)
			}

			exceededCount++
		}
	}

	log.Printf("✅ Traffic limit check complete. Found %d clients exceeding limits", exceededCount)
	return nil
}

func StartClientLifecycleMonitor(intervalMinutes int) {
	if intervalMinutes <= 0 {
		intervalMinutes = 30
	}

	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	log.Printf("🚀 Client lifecycle monitor started (checking every %d minutes)", intervalMinutes)

	if err := CheckExpiredClients(); err != nil {
		log.Printf("Error in expiration check: %v", err)
	}

	if err := CheckTrafficLimits(); err != nil {
		log.Printf("Error in traffic limit check: %v", err)
	}

	for range ticker.C {
		if err := CheckExpiredClients(); err != nil {
			log.Printf("Error in expiration check: %v", err)
		}

		if err := CheckTrafficLimits(); err != nil {
			log.Printf("Error in traffic limit check: %v", err)
		}
	}
}

