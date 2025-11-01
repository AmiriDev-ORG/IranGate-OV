package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amiridev-org/irangate-ov/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var (
	DataBaseDir     = readEnvOrDefault("IRANGATE_DATA_DIR", "/opt/irangate")
	AdminDBPath     = filepath.Join(DataBaseDir, "webpanel", "admin.json")
	TelegramDBPath  = filepath.Join(DataBaseDir, "webpanel", "telegram.json")
	ServerDBPath    = filepath.Join(DataBaseDir, "webpanel", "server.json")
	TemplatesPath   = filepath.Join(DataBaseDir, "webpanel", "templates")
	JWTSecret       = readEnvOrDefault("JWT_SECRET", "your-secret-key-change-this-in-production")
	TokenExpiration = 24 * time.Hour
)

func readEnvOrDefault(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

type Admin struct {
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login"`
}
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResponse struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
}
type ServerConfig struct {
	ServerIP     string `json:"server_ip"`
	ServerDomain string `json:"server_domain"`
	ServerPort   int    `json:"server_port"`
	Protocol     string `json:"protocol"`
}
type Template struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Content     string    `json:"content"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func initDatabases() error {
	os.MkdirAll(filepath.Dir(AdminDBPath), 0755)
	os.MkdirAll(TemplatesPath, 0755)
	if _, err := os.Stat(AdminDBPath); os.IsNotExist(err) {
		initialUser := readEnvOrDefault("WEBPANEL_ADMIN_USER", "admin")
		initialPass := readEnvOrDefault("WEBPANEL_ADMIN_PASS", "admin123")
		initialEmail := readEnvOrDefault("WEBPANEL_ADMIN_EMAIL", "admin@irangate.local")
		passwordHash := hashPassword(initialPass)
		admin := Admin{
			Username:     initialUser,
			PasswordHash: passwordHash,
			Email:        initialEmail,
			CreatedAt:    time.Now(),
		}
		data, err := json.MarshalIndent(admin, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(AdminDBPath, data, 0600); err != nil {
			return err
		}
		log.Printf("⚠️  Initial admin created - Username: %s\n", initialUser)
		if initialPass == "admin123" {
			log.Println("⚠️  CHANGE DEFAULT PASSWORD IMMEDIATELY!")
		}
	}
	if _, err := os.Stat(TelegramDBPath); os.IsNotExist(err) {
		telegram := TelegramConfig{
			Enabled: false,
		}
		data, err := json.MarshalIndent(telegram, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(TelegramDBPath, data, 0600); err != nil {
			return err
		}
	}
	if _, err := os.Stat(ServerDBPath); os.IsNotExist(err) {
		server := ServerConfig{
			ServerPort: 1194,
			Protocol:   "udp",
		}
		data, err := json.MarshalIndent(server, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(ServerDBPath, data, 0600); err != nil {
			return err
		}
	}
	return nil
}
func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Warning: Failed to hash password with bcrypt, falling back to SHA-256: %v", err)
		shaHash := sha256.Sum256([]byte(password))
		return fmt.Sprintf("%x", shaHash)
	}
	return string(hash)
}
func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true
	}
	shaHash := sha256.Sum256([]byte(password))
	shaHashStr := fmt.Sprintf("%x", shaHash)
	return shaHashStr == hash
}
func generateToken(username string) (string, time.Time, error) {
	expirationTime := time.Now().Add(TokenExpiration)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	return tokenString, expirationTime, err
}
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendError(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTSecret), nil
		})
		if err != nil || !token.Valid {
			sendError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
func sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Message: message,
	})
}
func sendSuccess(w http.ResponseWriter, message string, data interface{}) {
	sendJSON(w, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	data, err := os.ReadFile(AdminDBPath)
	if err != nil {
		sendError(w, "Authentication failed", http.StatusInternalServerError)
		return
	}
	var admin Admin
	if err := json.Unmarshal(data, &admin); err != nil {
		sendError(w, "Authentication failed", http.StatusInternalServerError)
		return
	}
	if admin.Username != req.Username || !verifyPassword(req.Password, admin.PasswordHash) {
		sendError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	token, expiresAt, err := generateToken(req.Username)
	if err != nil {
		sendError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	admin.LastLogin = time.Now()
	data, err = json.MarshalIndent(admin, "", "  ")
	if err != nil {
		log.Printf("Warning: Failed to marshal admin data: %v", err)
	}
	if err := os.WriteFile(AdminDBPath, data, 0600); err != nil {
		log.Printf("Warning: Failed to write admin data: %v", err)
	}
	sendSuccess(w, "Login successful", LoginResponse{
		Token:     token,
		Username:  req.Username,
		ExpiresAt: expiresAt,
	})
}

var clientsCache struct {
	data      []interface{}
	timestamp time.Time
	mutex     sync.RWMutex
}

func getClientsHandler(w http.ResponseWriter, r *http.Request) {
	clientsCache.mutex.RLock()
	cacheValid := time.Since(clientsCache.timestamp) < 5*time.Second && clientsCache.data != nil
	clientsCache.mutex.RUnlock()
	if cacheValid {
		clientsCache.mutex.RLock()
		data := clientsCache.data
		clientsCache.mutex.RUnlock()
		sendSuccess(w, "Clients retrieved successfully (cached)", data)
		return
	}
	cmd := exec.Command("irangate", "client", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to get clients: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var clients []interface{}
	if err := json.Unmarshal(output, &clients); err != nil {
		sendError(w, "Failed to parse clients", http.StatusInternalServerError)
		return
	}
	trafficDB := NewTrafficDatabase()
	trafficStats, err := trafficDB.GetCurrentClientStats()
	parser := NewOpenVPNStatusParser()
	currentlyConnected := make(map[string]bool)
	connectedSince := make(map[string]time.Time)
	foundAnyClients := false
	for _, statusFile := range parser.StatusFiles {
		if clients, err := parser.ParseStatusLog(statusFile); err == nil && len(clients) > 0 {
			foundAnyClients = true
			for _, client := range clients {
				currentlyConnected[client.CommonName] = true
				if !client.ConnectedSince.IsZero() {
					connectedSince[client.CommonName] = client.ConnectedSince
				}
			}
			log.Printf("Found %d connected clients from %s", len(clients), statusFile)
		} else if err != nil {
			log.Printf("Error parsing status file %s: %v", statusFile, err)
		}
	}
	if !foundAnyClients {
		log.Printf("Warning: No connected clients found in any status file. Checked: %v", parser.StatusFiles)
	}
	if err == nil {
		trafficMap := make(map[string]ClientTrafficStats)
		for _, stat := range trafficStats {
			trafficMap[stat.ClientName] = stat
		}
		for i, client := range clients {
			if clientMap, ok := client.(map[string]interface{}); ok {
				if name, ok := clientMap["name"].(string); ok {
					isCurrentlyConnected := currentlyConnected[name]
					if name == "sandman" || isCurrentlyConnected {
						log.Printf("DEBUG: Client %s - isCurrentlyConnected=%v, connectedSince=%v",
							name, isCurrentlyConnected, connectedSince[name])
					}
					if stat, exists := trafficMap[name]; exists {
						stat.Connected = isCurrentlyConnected
						if isCurrentlyConnected {
							if connTime, ok := connectedSince[name]; ok && !connTime.IsZero() {
								if connTime.After(stat.LastSeen) {
									stat.LastSeen = connTime
								}
								clientMap["connected_since"] = connTime.Format(time.RFC3339)
							} else {
								clientMap["connected_since"] = time.Now().Format(time.RFC3339)
							}
						}
						clientMap["bytes_received"] = stat.CurrentReceived
						clientMap["bytes_sent"] = stat.CurrentSent
						clientMap["total_bytes"] = stat.CurrentTotal
						clientMap["last_seen"] = stat.LastSeen.Format(time.RFC3339)
						clientMap["last_connection"] = stat.LastSeen.Format(time.RFC3339)
						clientMap["connected"] = isCurrentlyConnected
						clientMap["active"] = isCurrentlyConnected
						clientMap["daily_usage"] = stat.DailyUsage
						clientMap["monthly_usage"] = stat.MonthlyUsage
					} else {
						clientMap["connected"] = isCurrentlyConnected
						clientMap["active"] = isCurrentlyConnected
						if isCurrentlyConnected {
							if connTime, ok := connectedSince[name]; ok && !connTime.IsZero() {
								clientMap["last_seen"] = connTime.Format(time.RFC3339)
								clientMap["last_connection"] = connTime.Format(time.RFC3339)
								clientMap["connected_since"] = connTime.Format(time.RFC3339)
							} else {
								clientMap["last_seen"] = time.Now().Format(time.RFC3339)
								clientMap["last_connection"] = time.Now().Format(time.RFC3339)
								clientMap["connected_since"] = time.Now().Format(time.RFC3339)
							}
						} else {
							now := time.Now()
							weekAgo := now.AddDate(0, 0, -7)
							if history, err := trafficDB.GetClientTrafficHistory(name, weekAgo, now); err == nil && len(history) > 0 {
								var lastRecord TrafficRecord
								for _, record := range history {
									if record.Timestamp.After(lastRecord.Timestamp) {
										lastRecord = record
									}
								}
								if !lastRecord.Timestamp.IsZero() {
									clientMap["last_seen"] = lastRecord.Timestamp.Format(time.RFC3339)
									clientMap["last_connection"] = lastRecord.Timestamp.Format(time.RFC3339)
									cumulativeReceived, cumulativeSent := trafficDB.getCumulativeTraffic(name)
									clientMap["bytes_received"] = cumulativeReceived
									clientMap["bytes_sent"] = cumulativeSent
									clientMap["total_bytes"] = cumulativeReceived + cumulativeSent
								} else {
									clientMap["last_seen"] = ""
									clientMap["last_connection"] = ""
								}
							} else {
								clientMap["last_seen"] = ""
								clientMap["last_connection"] = ""
							}
						}
						if _, exists := clientMap["bytes_received"]; !exists {
							clientMap["bytes_received"] = 0
						}
						if _, exists := clientMap["bytes_sent"]; !exists {
							clientMap["bytes_sent"] = 0
						}
						if _, exists := clientMap["total_bytes"]; !exists {
							clientMap["total_bytes"] = 0
						}
						clientMap["daily_usage"] = int64(0)
						clientMap["monthly_usage"] = int64(0)
					}
					clients[i] = clientMap
				}
			}
		}
	}
	clientsCache.mutex.Lock()
	clientsCache.data = clients
	clientsCache.timestamp = time.Now()
	clientsCache.mutex.Unlock()
	sendSuccess(w, "Clients retrieved successfully", clients)
}
func createClientHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              string `json:"name"`
		Email             string `json:"email,omitempty"`
		ExpirationDays    int    `json:"expiration_days,omitempty"`
		TrafficLimitBytes int64  `json:"traffic_limit_bytes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	cmd := exec.Command("irangate", "client", "add", req.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to create client: "+string(output), http.StatusInternalServerError)
		return
	}
	
	dataDir := os.Getenv("IRANGATE_DATA_DIR")
	if dataDir == "" {
		dataDir = "/opt/irangate"
	}
	
	clientPath := filepath.Join(dataDir, "database", "clients", req.Name+".json")
	
	if req.ExpirationDays > 0 || req.TrafficLimitBytes > 0 {
		data, err := os.ReadFile(clientPath)
		if err == nil {
			var clientData map[string]interface{}
			if err := json.Unmarshal(data, &clientData); err == nil {
				if req.ExpirationDays > 0 {
					expiresAt := time.Now().AddDate(0, 0, req.ExpirationDays)
					clientData["expires_at"] = expiresAt.Format(time.RFC3339)
				}
				
				if req.TrafficLimitBytes > 0 {
					clientData["traffic_limit_bytes"] = req.TrafficLimitBytes
				}
				
				newData, err := json.MarshalIndent(clientData, "", "  ")
				if err == nil {
					os.WriteFile(clientPath, newData, 0644)
				}
			}
		}
	}
	
	clientsCache.mutex.Lock()
	clientsCache.data = nil
	clientsCache.timestamp = time.Time{}
	clientsCache.mutex.Unlock()
	sendSuccess(w, "Client created successfully", map[string]string{"name": req.Name})
}
func deleteClientHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	cmd := exec.Command("irangate", "client", "remove", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to remove client: "+string(output), http.StatusInternalServerError)
		return
	}
	clientsCache.mutex.Lock()
	clientsCache.data = nil
	clientsCache.timestamp = time.Time{}
	clientsCache.mutex.Unlock()
	sendSuccess(w, "Client removed and certificate revoked", map[string]string{"name": name})
}
func activateClientHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if err := ActivateClient(name); err != nil {
		sendError(w, fmt.Sprintf("Failed to activate client: %v", err), http.StatusInternalServerError)
		return
	}
	
	clientsCache.mutex.Lock()
	clientsCache.data = nil
	clientsCache.timestamp = time.Time{}
	clientsCache.mutex.Unlock()
	
	sendSuccess(w, "Client activated successfully", map[string]string{"name": name, "status": "active"})
}
func deactivateClientHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	
	if err := RevokeClientCertificate(name, "manual_deactivation"); err != nil {
		log.Printf("Warning: Failed to revoke certificate for %s: %v", name, err)
	}
	
	if err := DeactivateClient(name); err != nil {
		sendError(w, fmt.Sprintf("Failed to deactivate client: %v", err), http.StatusInternalServerError)
		return
	}
	
	clientsCache.mutex.Lock()
	clientsCache.data = nil
	clientsCache.timestamp = time.Time{}
	clientsCache.mutex.Unlock()
	
	sendSuccess(w, "Client deactivated successfully", map[string]string{"name": name, "status": "inactive"})
}
func getClientStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	cmd := exec.Command("irangate", "client", "status", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to get client status", http.StatusInternalServerError)
		return
	}
	status := strings.TrimSpace(string(output))
	sendSuccess(w, "Client status retrieved", map[string]string{
		"name":   name,
		"status": status,
	})
}
func exportClientHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	cmd := exec.Command("irangate", "client", "export", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to export client config", http.StatusInternalServerError)
		return
	}
	lines := strings.Split(string(output), "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "time=") ||
			strings.HasPrefix(trimmed, "level=") ||
			(strings.Contains(trimmed, "time=") && strings.Contains(trimmed, "level=") && strings.Contains(trimmed, "msg=")) {
			continue
		}
		cleanedLines = append(cleanedLines, line)
	}
	cleanedOutput := strings.Join(cleanedLines, "\n")
	w.Header().Set("Content-Type", "application/x-openvpn-profile")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.ovpn", name))
	w.Write([]byte(cleanedOutput))
}
func getOrphanedClientsHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("irangate", "client", "list-orphaned", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to get orphaned clients", http.StatusInternalServerError)
		return
	}
	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		sendSuccess(w, "No orphaned certificates", map[string]int{"count": 0})
		return
	}
	sendSuccess(w, "Orphaned clients retrieved", result)
}
func cleanupOrphanedHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("irangate", "client", "cleanup-orphaned", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to cleanup orphaned clients", http.StatusInternalServerError)
		return
	}
	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		sendError(w, "Failed to parse cleanup result", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Orphaned clients cleaned up", result)
}
func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("irangate", "config", "show", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to get configuration", http.StatusInternalServerError)
		return
	}
	var config interface{}
	if err := json.Unmarshal(output, &config); err != nil {
		sendError(w, "Failed to parse configuration", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Configuration retrieved", config)
}
func updateConfigHandler(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	for key, value := range req {
		cmd := exec.Command("irangate", "config", "set", key, value)
		if _, err := cmd.CombinedOutput(); err != nil {
			sendError(w, fmt.Sprintf("Failed to update %s", key), http.StatusInternalServerError)
			return
		}
	}
	sendSuccess(w, "Configuration updated successfully", nil)
}
func getStatsHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("irangate", "monitor", "live", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}
	var stats interface{}
	if err := json.Unmarshal(output, &stats); err != nil {
		sendError(w, "Failed to parse statistics", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Statistics retrieved", stats)
}
func getServerStatusHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("irangate", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to get server status", http.StatusInternalServerError)
		return
	}
	status := strings.TrimSpace(string(output))
	isRunning := strings.Contains(status, "running") || strings.Contains(status, "active")
	sendSuccess(w, "Server status retrieved", map[string]interface{}{
		"status":     status,
		"is_running": isRunning,
	})
}
func restartServerHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("irangate", "restart")
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to restart server: "+string(output), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Server restarted successfully", nil)
}
func startServerHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("irangate", "start")
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to start server: "+string(output), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Server started successfully", nil)
}
func stopServerHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("irangate", "stop")
	output, err := cmd.CombinedOutput()
	if err != nil {
		sendError(w, "Failed to stop server: "+string(output), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Server stopped successfully", nil)
}
func getAdminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(AdminDBPath)
	if err != nil {
		sendError(w, "Failed to read admin settings", http.StatusInternalServerError)
		return
	}
	var admin Admin
	if err := json.Unmarshal(data, &admin); err != nil {
		sendError(w, "Failed to parse admin settings", http.StatusInternalServerError)
		return
	}
	admin.PasswordHash = ""
	sendSuccess(w, "Admin settings retrieved", admin)
}
func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	data, err := os.ReadFile(AdminDBPath)
	if err != nil {
		sendError(w, "Failed to read admin data", http.StatusInternalServerError)
		return
	}
	var admin Admin
	if err := json.Unmarshal(data, &admin); err != nil {
		sendError(w, "Failed to parse admin data", http.StatusInternalServerError)
		return
	}
	if !verifyPassword(req.OldPassword, admin.PasswordHash) {
		sendError(w, "Incorrect old password", http.StatusUnauthorized)
		return
	}
	admin.PasswordHash = hashPassword(req.NewPassword)
	data, err = json.MarshalIndent(admin, "", "  ")
	if err != nil {
		sendError(w, "Failed to serialize admin data", http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(AdminDBPath, data, 0600); err != nil {
		sendError(w, "Failed to save admin data", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Password changed successfully", nil)
}
func getTelegramSettingsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(TelegramDBPath)
	if err != nil {
		sendError(w, "Failed to read telegram settings", http.StatusInternalServerError)
		return
	}
	var telegram TelegramConfig
	if err := json.Unmarshal(data, &telegram); err != nil {
		sendError(w, "Failed to parse telegram settings", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Telegram settings retrieved", telegram)
}
func getServerSettingsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(ServerDBPath)
	if err != nil {
		sendError(w, "Failed to read server settings", http.StatusInternalServerError)
		return
	}
	var server ServerConfig
	if err := json.Unmarshal(data, &server); err != nil {
		sendError(w, "Failed to parse server settings", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Server settings retrieved", server)
}
func updateServerSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var server ServerConfig
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	data, err := json.MarshalIndent(server, "", "  ")
	if err != nil {
		sendError(w, "Failed to serialize server settings", http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(ServerDBPath, data, 0600); err != nil {
		sendError(w, "Failed to save server settings", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Server settings updated", server)
}
func exportBackupHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("irangate", "config", "backup")
	if _, err := cmd.CombinedOutput(); err != nil {
		sendError(w, "Failed to create backup", http.StatusInternalServerError)
		return
	}
	backupDir := filepath.Join(DataBaseDir, "database", "backups")
	os.MkdirAll(backupDir, 0755)
	backupDir += "/"
	files, _ := os.ReadDir(backupDir)
	if len(files) == 0 {
		sendError(w, "No backup found", http.StatusNotFound)
		return
	}
	latestFile := files[len(files)-1]
	backupPath := filepath.Join(backupDir, latestFile.Name())
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", latestFile.Name()))
	file, err := os.Open(backupPath)
	if err != nil {
		sendError(w, "Failed to open backup file", http.StatusInternalServerError)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: Failed to close backup file: %v\n", closeErr)
		}
	}()
	io.Copy(w, file)
}
func importBackupHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("backup")
	if err != nil {
		sendError(w, "Failed to read backup file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	backupPath := filepath.Join(DataBaseDir, "database", "backups", handler.Filename)
	dst, err := os.Create(backupPath)
	if err != nil {
		sendError(w, "Failed to save backup file", http.StatusInternalServerError)
		return
	}
	defer func() {
		if closeErr := dst.Close(); closeErr != nil {
			fmt.Printf("Warning: Failed to close backup destination file: %v\n", closeErr)
		}
	}()
	io.Copy(dst, file)
	cmd := exec.Command("irangate", "config", "restore", backupPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		sendError(w, "Failed to restore backup: "+string(output), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Backup restored successfully", nil)
}
func healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now()).String(),
	}
	sendSuccess(w, "Health check passed", health)
}
func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := loadUsers()
	if err != nil {
		sendError(w, "Failed to load users", http.StatusInternalServerError)
		return
	}
	for i := range users {
		users[i].PasswordHash = ""
	}
	sendSuccess(w, "Users retrieved successfully", users)
}
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	role := Role(req.Role)
	if role != RoleAdmin && role != RoleModerator && role != RoleViewer {
		sendError(w, "Invalid role", http.StatusBadRequest)
		return
	}
	createdBy := "admin"
	user, err := createUser(req.Username, req.Password, req.Email, role, createdBy)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	user.PasswordHash = ""
	sendSuccess(w, "User created successfully", user)
}
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := updateUser(username, updates); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "User updated successfully", nil)
}
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	users, _ := loadUsers()
	adminCount := 0
	for _, u := range users {
		if u.Role == RoleAdmin && u.Active {
			adminCount++
		}
	}
	user, _ := getUserByUsername(username)
	if user != nil && user.Role == RoleAdmin && adminCount <= 1 {
		sendError(w, "Cannot delete the last admin user", http.StatusBadRequest)
		return
	}
	if err := deleteUser(username); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "User deleted successfully", nil)
}
func getSubscriptionsHandler(w http.ResponseWriter, _ *http.Request) {
	subs, err := loadSubscriptions()
	if err != nil {
		sendError(w, "Failed to load subscriptions", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Subscriptions retrieved successfully", subs)
}
func createSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	var sub Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := createSubscription(sub); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendSuccess(w, "Subscription created successfully", sub)
}
func updateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["client"]
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := updateSubscription(clientName, updates); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Subscription updated successfully", nil)
}
func deleteSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["client"]
	if err := deleteSubscription(clientName); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Subscription deleted successfully", nil)
}
func getExpiringHandler(w http.ResponseWriter, r *http.Request) {
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		fmt.Sscanf(d, "%d", &days)
	}
	expiring, err := getExpiringSubscriptions(days)
	if err != nil {
		sendError(w, "Failed to get expiring subscriptions", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Expiring subscriptions retrieved", map[string]interface{}{
		"count":         len(expiring),
		"days":          days,
		"subscriptions": expiring,
	})
}
func renewSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["client"]
	var req struct {
		Days int `json:"days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Days = 30
	}
	if err := renewSubscription(clientName, req.Days); err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, fmt.Sprintf("Subscription renewed for %d days", req.Days), nil)
}
func main() {
	if err := initDatabases(); err != nil {
		log.Fatal("Failed to initialize databases:", err)
	}
	go hub.run()
	go startStatsBroadcaster()
	go startSystemResourcesBroadcaster()
	go func() {
		parser := NewOpenVPNStatusParser()
		collectionInterval := 1 * time.Minute
		configPath := "/etc/irangate/traffic_config.json"
		if configData, err := os.ReadFile(configPath); err == nil {
			var config struct {
				CollectionIntervalSeconds int `json:"collection_interval_seconds"`
			}
			if err := json.Unmarshal(configData, &config); err == nil && config.CollectionIntervalSeconds > 0 {
				collectionInterval = time.Duration(config.CollectionIntervalSeconds) * time.Second
				log.Printf("Traffic collection interval set to %v from config", collectionInterval)
			}
		}
		parser.StartTrafficCollection(collectionInterval)
	}()
	go StartDataLimitScanner(30)
	go StartClientLifecycleMonitor(30)
	go StartAsyncCleanup()
	basePathEnv := readEnvOrDefault("WEBPANEL_BASEPATH", "/")
	basePath := strings.TrimSpace(basePathEnv)
	if basePath == "" {
		basePath = "/"
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	if len(basePath) > 1 && strings.HasSuffix(basePath, "/") {
		basePath = strings.TrimRight(basePath, "/")
	}
	portStr := readEnvOrDefault("WEBPANEL_PORT", "8080")
	if _, err := strconv.Atoi(portStr); err != nil {
		log.Fatalf("Invalid WEBPANEL_PORT: %s", portStr)
	}
	staticDir := readEnvOrDefault("WEBPANEL_STATIC_DIR", "/opt/irangate/webpanel/frontend")
	rootRouter := mux.NewRouter()
	var r *mux.Router
	if basePath == "/" {
		r = rootRouter
	} else {
		r = rootRouter.PathPrefix(basePath).Subrouter()
	}
	r.HandleFunc("/ws", wsHandler)
	r.HandleFunc("/api/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/health", healthHandler).Methods("GET")
	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware)
	api.HandleFunc("/clients", getClientsHandler).Methods("GET")
	api.HandleFunc("/clients", createClientHandler).Methods("POST")
	api.HandleFunc("/clients/{name}", deleteClientHandler).Methods("DELETE")
	api.HandleFunc("/clients/{name}/activate", activateClientHandler).Methods("POST")
	api.HandleFunc("/clients/{name}/deactivate", deactivateClientHandler).Methods("POST")
	api.HandleFunc("/clients/{name}/status", getClientStatusHandler).Methods("GET")
	api.HandleFunc("/clients/{name}/export", exportClientHandler).Methods("GET")
	api.HandleFunc("/clients/orphaned", getOrphanedClientsHandler).Methods("GET")
	api.HandleFunc("/clients/cleanup-orphaned", cleanupOrphanedHandler).Methods("POST")
	api.HandleFunc("/clients/bulk-create", bulkCreateClientsHandler).Methods("POST")
	api.HandleFunc("/clients/bulk-delete", bulkDeleteClientsHandler).Methods("POST")
	api.HandleFunc("/clients/bulk-export", bulkExportClientsHandler).Methods("POST")
	api.HandleFunc("/openvpn/configs", getOpenVPNConfigsHandler).Methods("GET")
	api.HandleFunc("/openvpn/sync", syncOpenVPNConfigsHandler).Methods("POST")
	api.HandleFunc("/openvpn/configs/{name}", getOpenVPNConfigHandler).Methods("GET")
	api.HandleFunc("/openvpn/sync/status", getOpenVPNSyncStatusHandler).Methods("GET")
	api.HandleFunc("/openvpn/create-configs", createOpenVPNConfigsHandler).Methods("POST")
	api.HandleFunc("/openvpn/update-config", updateOpenVPNConfigHandler).Methods("PUT")
	api.HandleFunc("/server-checker/ping", initiatePingCheckHandler).Methods("POST")
	api.HandleFunc("/server-checker/status/{request_id}", getPingCheckStatusHandler).Methods("GET")
	api.HandleFunc("/server-checker/history", getPingCheckHistoryHandler).Methods("GET")
	api.HandleFunc("/docs/{name}", getDocumentationHandler).Methods("GET")
	api.HandleFunc("/docs", saveDocumentationHandler).Methods("POST")
	api.HandleFunc("/docs", listDocumentationHandler).Methods("GET")
	api.HandleFunc("/clients/bulk-extend", bulkExtendClientsHandler).Methods("POST")
	api.HandleFunc("/clients/bulk-import", bulkImportClientsHandler).Methods("POST")
	api.HandleFunc("/clients/{name}/qrcode", generateClientQRCodeHandler).Methods("GET")
	api.HandleFunc("/clients/{name}/qrcode-download", downloadClientQRCodeHandler).Methods("GET")
	api.HandleFunc("/qrcode/generate", generateQRCodeHandler).Methods("POST")
	api.HandleFunc("/qrcode/batch-generate", batchGenerateQRCodeHandler).Methods("GET")
	api.HandleFunc("/clients/{name}/device", registerDeviceHandler).Methods("POST")
	api.HandleFunc("/clients/{name}/devices", getClientDevicesHandler).Methods("GET")
	api.HandleFunc("/clients/{name}/device/{id}", updateDeviceHandler).Methods("PUT")
	api.HandleFunc("/clients/{name}/device/{id}", removeDeviceHandler).Methods("DELETE")
	api.HandleFunc("/devices/all", getAllDevicesHandler).Methods("GET")
	api.HandleFunc("/devices/cleanup", cleanupInactiveDevicesHandler).Methods("POST")
	api.HandleFunc("/clients/{name}/quota", setClientQuotaHandler).Methods("POST")
	api.HandleFunc("/clients/{name}/quota-status", getClientQuotaStatusHandler).Methods("GET")
	api.HandleFunc("/clients/{name}/data-limit", setClientDataLimitHandler).Methods("POST")
	api.HandleFunc("/clients/{name}/data-limit", getClientDataLimitHandler).Methods("GET")
	api.HandleFunc("/clients/{name}/data-limit", deleteClientDataLimitHandler).Methods("DELETE")
	api.HandleFunc("/data-limits/check", checkDataLimitsHandler).Methods("POST")
	api.HandleFunc("/config", getConfigHandler).Methods("GET")
	api.HandleFunc("/config", updateConfigHandler).Methods("PUT")
	api.HandleFunc("/monitor/stats", getStatsHandler).Methods("GET")
	api.HandleFunc("/server/status", getServerStatusHandler).Methods("GET")
	api.HandleFunc("/server/restart", restartServerHandler).Methods("POST")
	api.HandleFunc("/server/start", startServerHandler).Methods("POST")
	api.HandleFunc("/server/stop", stopServerHandler).Methods("POST")
	api.HandleFunc("/system/resources", getSystemResourcesHandler).Methods("GET")
	api.HandleFunc("/system/resources/history", getSystemResourcesHistoryHandler).Methods("GET")
	api.HandleFunc("/system/processes", getTopProcessesHandler).Methods("GET")
	api.HandleFunc("/analytics/traffic-history", trafficHistoryHandler).Methods("GET", "POST")
	api.HandleFunc("/analytics/bandwidth-usage", bandwidthUsageHandler).Methods("GET")
	api.HandleFunc("/analytics/peak-hours", peakHoursHandler).Methods("GET")
	api.HandleFunc("/analytics/current-traffic", getCurrentTrafficHandler).Methods("GET")
	api.HandleFunc("/analytics/client-traffic/{name}", getClientTrafficHandler).Methods("GET")
	api.HandleFunc("/analytics/export-report", exportReportHandler).Methods("GET")
	api.HandleFunc("/traffic/stats", getTrafficStatsHandler).Methods("GET")
	api.HandleFunc("/traffic/stats/{name}", getClientTrafficStatsHandler).Methods("GET")
	api.HandleFunc("/traffic/history", getTrafficHistoryHandler).Methods("GET")
	api.HandleFunc("/traffic/top-users", getTopUsersHandler).Methods("GET")
	api.HandleFunc("/traffic/export", exportTrafficHandler).Methods("GET")
	api.HandleFunc("/traffic/quota/{name}", getQuotaUsageHandler).Methods("GET")
	api.HandleFunc("/traffic/efficiency/{name}", getBandwidthEfficiencyHandler).Methods("GET")
	api.HandleFunc("/traffic/anomalies/{name}", detectAnomaliesHandler).Methods("GET")
	api.HandleFunc("/traffic/realtime", getRealTimeStatsHandler).Methods("GET")
	api.HandleFunc("/async/operations", getAsyncOperationsHandler).Methods("GET")
	api.HandleFunc("/async/operations/{id}", getAsyncOperationStatusHandler).Methods("GET")
	api.HandleFunc("/async/clients", startAsyncClientListHandler).Methods("POST")
	api.HandleFunc("/async/server-status", startAsyncServerStatusHandler).Methods("POST")
	api.HandleFunc("/async/config-update", startAsyncConfigUpdateHandler).Methods("POST")
	api.HandleFunc("/settings/admin", getAdminSettingsHandler).Methods("GET")
	api.HandleFunc("/settings/admin/password", changePasswordHandler).Methods("PUT")
	api.HandleFunc("/settings/telegram", getTelegramSettingsHandler).Methods("GET")
	api.HandleFunc("/settings/telegram", updateTelegramSettingsHandler).Methods("PUT")
	api.HandleFunc("/settings/server", getServerSettingsHandler).Methods("GET")
	api.HandleFunc("/settings/server", updateServerSettingsHandler).Methods("PUT")
	api.HandleFunc("/notifications/telegram/test", telegramTestHandler).Methods("POST")
	api.HandleFunc("/notifications/telegram/configure", telegramConfigureHandler).Methods("POST")
	api.HandleFunc("/notifications/telegram/status", telegramStatusHandler).Methods("GET")
	api.HandleFunc("/notifications/telegram/settings", telegramSettingsHandler).Methods("GET")
	api.HandleFunc("/autoactions", getAutoActionsHandler).Methods("GET")
	api.HandleFunc("/autoactions", createAutoActionHandler).Methods("POST")
	api.HandleFunc("/autoactions/{id}", updateAutoActionHandler).Methods("PUT")
	api.HandleFunc("/autoactions/{id}", deleteAutoActionHandler).Methods("DELETE")
	api.HandleFunc("/autoactions/{id}/test", testAutoActionHandler).Methods("POST")
	api.HandleFunc("/autoactions/{id}/execute", executeAutoActionHandler).Methods("POST")
	api.HandleFunc("/backup/schedules", getBackupSchedulesHandler).Methods("GET")
	api.HandleFunc("/backup/schedules", createBackupScheduleHandler).Methods("POST")
	api.HandleFunc("/backup/schedules/{id}", updateBackupScheduleHandler).Methods("PUT")
	api.HandleFunc("/backup/schedules/{id}", deleteBackupScheduleHandler).Methods("DELETE")
	api.HandleFunc("/backup/schedules/{id}/run", runBackupScheduleHandler).Methods("POST")
	api.HandleFunc("/backup/list", listBackupsHandler).Methods("GET")
	api.HandleFunc("/backup/{filename}", deleteBackupHandler).Methods("DELETE")
	api.HandleFunc("/backup/{filename}/download", downloadBackupHandler).Methods("GET")
	api.HandleFunc("/server-config", getServerConfigHandler).Methods("GET")
	api.HandleFunc("/server-config", updateServerConfigHandler).Methods("PUT")
	api.HandleFunc("/server-config/detect-ip", detectServerIPHandler).Methods("POST")
	api.HandleFunc("/logs/openvpn", getOpenVPNLogsHandler).Methods("GET")
	api.HandleFunc("/logs/webpanel", getWebPanelLogsHandler).Methods("GET")
	api.HandleFunc("/logs/system", getSystemLogsHandler).Methods("GET")
	api.HandleFunc("/logs/search", searchLogsHandler).Methods("GET", "POST")
	api.HandleFunc("/logs/tail", tailLogsHandler).Methods("GET")
	api.HandleFunc("/logs/stream", streamLogsHandler).Methods("GET")
	api.HandleFunc("/backup/export", exportBackupHandler).Methods("GET")
	api.HandleFunc("/backup/import", importBackupHandler).Methods("POST")
	api.HandleFunc("/users", getUsersHandler).Methods("GET")
	api.HandleFunc("/users", createUserHandler).Methods("POST")
	api.HandleFunc("/users/{username}", updateUserHandler).Methods("PUT")
	api.HandleFunc("/users/{username}", deleteUserHandler).Methods("DELETE")
	api.HandleFunc("/subscriptions", getSubscriptionsHandler).Methods("GET")
	api.HandleFunc("/subscriptions", createSubscriptionHandler).Methods("POST")
	api.HandleFunc("/subscriptions/{client}", updateSubscriptionHandler).Methods("PUT")
	api.HandleFunc("/subscriptions/{client}", deleteSubscriptionHandler).Methods("DELETE")
	api.HandleFunc("/subscriptions/expiring", getExpiringHandler).Methods("GET")
	api.HandleFunc("/subscriptions/{client}/renew", renewSubscriptionHandler).Methods("POST")
	if basePath == "/" {
		rootRouter.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))
	} else {
		rootRouter.PathPrefix(basePath + "/").Handler(http.StripPrefix(basePath+"/", http.FileServer(http.Dir(staticDir))))
	}
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(rootRouter)
	log.Println("🚀 IranGate Web Panel API Starting...")
	log.Printf("📍 API Server: http://localhost:%s", portStr)
	log.Println("🔐 Default Login: configured via env or admin/admin123")
	if readEnvOrDefault("WEBPANEL_ADMIN_PASS", "admin123") == "admin123" {
		log.Println("⚠️  CHANGE DEFAULT PASSWORD IMMEDIATELY!")
	}
	addr := ":" + portStr
	if err := http.ListenAndServe(addr, corsHandler); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
func getCurrentTrafficHandler(w http.ResponseWriter, r *http.Request) {
	db := NewTrafficDatabase()
	stats, err := db.GetCurrentClientStats()
	if err != nil {
		sendError(w, "Failed to get current traffic stats", http.StatusInternalServerError)
		return
	}
	var totalReceived, totalSent, totalBytes int64
	var connectedClients int
	for _, stat := range stats {
		totalReceived += stat.CurrentReceived
		totalSent += stat.CurrentSent
		totalBytes += stat.CurrentTotal
		if stat.Connected {
			connectedClients++
		}
	}
	response := map[string]interface{}{
		"clients": stats,
		"summary": map[string]interface{}{
			"total_clients":     len(stats),
			"connected_clients": connectedClients,
			"total_received":    totalReceived,
			"total_sent":        totalSent,
			"total_bytes":       totalBytes,
			"total_received_gb": float64(totalReceived) / (1024 * 1024 * 1024),
			"total_sent_gb":     float64(totalSent) / (1024 * 1024 * 1024),
			"total_bytes_gb":    float64(totalBytes) / (1024 * 1024 * 1024),
		},
	}
	sendSuccess(w, "Current traffic data retrieved successfully", response)
}
func getClientTrafficHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	if clientName == "" {
		sendError(w, "Client name is required", http.StatusBadRequest)
		return
	}
	db := NewTrafficDatabase()
	stats, err := db.GetCurrentClientStats()
	if err != nil {
		sendError(w, "Failed to get current traffic stats", http.StatusInternalServerError)
		return
	}
	var clientStats *ClientTrafficStats
	for _, stat := range stats {
		if stat.ClientName == clientName {
			clientStats = &stat
			break
		}
	}
	if clientStats == nil {
		sendError(w, "Client not found", http.StatusNotFound)
		return
	}
	today := time.Now()
	dailySummary, err := db.GetDailySummary(clientName, today)
	if err != nil {
		sendError(w, "Failed to get daily summary", http.StatusInternalServerError)
		return
	}
	weekAgo := today.AddDate(0, 0, -7)
	history, err := db.GetClientTrafficHistory(clientName, weekAgo, today)
	if err != nil {
		sendError(w, "Failed to get traffic history", http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"client_name":    clientName,
		"current_stats":  clientStats,
		"daily_summary":  dailySummary,
		"weekly_history": history,
	}
	sendSuccess(w, "Client traffic data retrieved successfully", response)
}
func setClientDataLimitHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	var req struct {
		LimitGB float64 `json:"limit_gb"`
		Enabled bool    `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	limitBytes := int64(req.LimitGB * 1024 * 1024 * 1024)
	if err := SetClientDataLimit(clientName, limitBytes, req.Enabled); err != nil {
		sendError(w, fmt.Sprintf("Failed to set data limit: %v", err), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Data limit set successfully", map[string]interface{}{
		"client_name": clientName,
		"limit_bytes": limitBytes,
		"limit_gb":    req.LimitGB,
		"enabled":     req.Enabled,
	})
}
func getClientDataLimitHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	limits, err := LoadDataLimits()
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to load data limits: %v", err), http.StatusInternalServerError)
		return
	}
	config, exists := limits[clientName]
	if !exists {
		sendSuccess(w, "No data limit set for this client", map[string]interface{}{
			"client_name": clientName,
			"enabled":     false,
		})
		return
	}
	usage, err := GetClientDataUsage(clientName)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to get data usage: %v", err), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Data limit retrieved successfully", map[string]interface{}{
		"client_name":     clientName,
		"limit_bytes":     config.LimitBytes,
		"limit_gb":        float64(config.LimitBytes) / (1024 * 1024 * 1024),
		"enabled":         config.Enabled,
		"current_bytes":   usage.TotalBytes,
		"current_gb":      float64(usage.TotalBytes) / (1024 * 1024 * 1024),
		"percentage_used": float64(usage.TotalBytes) / float64(config.LimitBytes) * 100,
		"remaining_bytes": config.LimitBytes - usage.TotalBytes,
		"remaining_gb":    float64(config.LimitBytes-usage.TotalBytes) / (1024 * 1024 * 1024),
	})
}
func deleteClientDataLimitHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	limits, err := LoadDataLimits()
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to load data limits: %v", err), http.StatusInternalServerError)
		return
	}
	delete(limits, clientName)
	if err := SaveDataLimits(limits); err != nil {
		sendError(w, fmt.Sprintf("Failed to save data limits: %v", err), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Data limit removed successfully", map[string]string{
		"client_name": clientName,
	})
}
func checkDataLimitsHandler(w http.ResponseWriter, r *http.Request) {
	if err := CheckAndEnforceDataLimits(); err != nil {
		sendError(w, fmt.Sprintf("Failed to check data limits: %v", err), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Data limit check completed successfully", nil)
}
func getAsyncOperationsHandler(w http.ResponseWriter, r *http.Request) {
	operations := asyncManager.ListOperations()
	sendSuccess(w, "Async operations retrieved successfully", operations)
}
func getAsyncOperationStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	operationID := vars["id"]
	operation, exists := asyncManager.GetOperationStatus(operationID)
	if !exists {
		sendError(w, "Operation not found", http.StatusNotFound)
		return
	}
	sendSuccess(w, "Operation status retrieved successfully", operation)
}
func startAsyncClientListHandler(w http.ResponseWriter, r *http.Request) {
	operationID := asyncManager.AsyncClientList()
	sendSuccess(w, "Async client list operation started", map[string]string{
		"operation_id": operationID,
	})
}
func startAsyncServerStatusHandler(w http.ResponseWriter, r *http.Request) {
	operationID := asyncManager.AsyncServerStatus()
	sendSuccess(w, "Async server status operation started", map[string]string{
		"operation_id": operationID,
	})
}
func startAsyncConfigUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var config map[string]string
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	operationID := asyncManager.AsyncConfigUpdate(config)
	sendSuccess(w, "Async config update operation started", map[string]string{
		"operation_id": operationID,
	})
}
func getSystemResourcesHandler(w http.ResponseWriter, r *http.Request) {
	resources, err := getSystemResources()
	if err != nil {
		sendError(w, "Failed to get system resources: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "System resources retrieved successfully", resources)
}
func getSystemResourcesHistoryHandler(w http.ResponseWriter, r *http.Request) {
	history := map[string]interface{}{
		"data":    []interface{}{},
		"message": "Historical data not yet implemented",
	}
	sendSuccess(w, "System resources history retrieved", history)
}
func getTopProcessesHandler(w http.ResponseWriter, r *http.Request) {
	processes, err := getTopProcesses()
	if err != nil {
		sendError(w, "Failed to get top processes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Top processes retrieved successfully", processes)
}
func getServerConfigHandler(w http.ResponseWriter, r *http.Request) {
	serverConfigPath := filepath.Join(DataBaseDir, "server_config.json")
	config, err := utils.LoadServerConfig(serverConfigPath)
	if err != nil {
		sendError(w, "Failed to load server config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Server configuration retrieved successfully", config)
}
func updateServerConfigHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServerIP     string `json:"server_ip"`
		ServerDomain string `json:"server_domain"`
		UseDomain    bool   `json:"use_domain"`
		AutoDetectIP bool   `json:"auto_detect_ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	serverConfigPath := filepath.Join(DataBaseDir, "server_config.json")
	config, err := utils.LoadServerConfig(serverConfigPath)
	if err != nil {
		sendError(w, "Failed to load server config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	config.ServerIP = req.ServerIP
	config.ServerDomain = req.ServerDomain
	config.Settings.UseDomain = req.UseDomain
	config.AutoDetectIP = req.AutoDetectIP
	config.LastUpdated = time.Now().Format(time.RFC3339)
	if err := utils.SaveServerConfig(serverConfigPath, config); err != nil {
		sendError(w, "Failed to save server config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Server configuration updated successfully", config)
}
func detectServerIPHandler(w http.ResponseWriter, r *http.Request) {
	detectedIP := utils.GetServerIP()
	if detectedIP == "" || detectedIP == "YOUR_SERVER_IP" {
		sendError(w, "Failed to detect server IP", http.StatusInternalServerError)
		return
	}
	serverConfigPath := filepath.Join(DataBaseDir, "server_config.json")
	config, err := utils.LoadServerConfig(serverConfigPath)
	if err != nil {
		sendError(w, "Failed to load server config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	config.ServerIP = detectedIP
	config.DetectedIPs.Primary = detectedIP
	config.LastUpdated = time.Now().Format(time.RFC3339)
	if err := utils.SaveServerConfig(serverConfigPath, config); err != nil {
		sendError(w, "Failed to save server config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Server IP detected and updated successfully", map[string]interface{}{
		"detected_ip": detectedIP,
		"config":      config,
	})
}
