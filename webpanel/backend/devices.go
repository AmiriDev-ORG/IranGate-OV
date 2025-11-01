package main
import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/gorilla/mux"
)
type DeviceRequest struct {
	DeviceName string `json:"device_name"`
	DeviceType string `json:"device_type"`
	UserAgent  string `json:"user_agent"`
}
type DeviceResponse struct {
	ID         string    `json:"id"`
	ClientName string    `json:"client_name"`
	DeviceName string    `json:"device_name"`
	DeviceType string    `json:"device_type"`
	LastSeen   time.Time `json:"last_seen"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}
const DevicesDataDir = "/opt/irangate/webpanel/devices"
func initDevicesDir() error {
	return os.MkdirAll(DevicesDataDir, 0755)
}
func registerDeviceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	var req DeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.DeviceName == "" {
		sendError(w, "Device name is required", http.StatusBadRequest)
		return
	}
	if req.DeviceType == "" {
		req.DeviceType = "unknown"
	}
	if err := initDevicesDir(); err != nil {
		sendError(w, "Failed to initialize devices directory", http.StatusInternalServerError)
		return
	}
	deviceID := generateDeviceID(clientName, req.DeviceName)
	clientIP := getClientIP(r)
	device := database.Device{
		ID:         deviceID,
		ClientName: clientName,
		DeviceName: req.DeviceName,
		DeviceType: req.DeviceType,
		LastSeen:   time.Now(),
		IPAddress:  clientIP,
		UserAgent:  req.UserAgent,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}
	if err := saveDevice(&device); err != nil {
		sendError(w, fmt.Sprintf("Failed to save device: %v", err), http.StatusInternalServerError)
		return
	}
	response := DeviceResponse{
		ID:         device.ID,
		ClientName: device.ClientName,
		DeviceName: device.DeviceName,
		DeviceType: device.DeviceType,
		LastSeen:   device.LastSeen,
		IPAddress:  device.IPAddress,
		UserAgent:  device.UserAgent,
		IsActive:   device.IsActive,
		CreatedAt:  device.CreatedAt,
	}
	sendSuccess(w, "Device registered successfully", response)
}
func getClientDevicesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	devices, err := loadClientDevices(clientName)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to load devices: %v", err), http.StatusInternalServerError)
		return
	}
	var response []DeviceResponse
	for _, device := range devices {
		response = append(response, DeviceResponse{
			ID:         device.ID,
			ClientName: device.ClientName,
			DeviceName: device.DeviceName,
			DeviceType: device.DeviceType,
			LastSeen:   device.LastSeen,
			IPAddress:  device.IPAddress,
			UserAgent:  device.UserAgent,
			IsActive:   device.IsActive,
			CreatedAt:  device.CreatedAt,
		})
	}
	sendSuccess(w, "Client devices retrieved successfully", response)
}
func updateDeviceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	deviceID := vars["id"]
	var req DeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	device, err := loadDevice(clientName, deviceID)
	if err != nil {
		sendError(w, "Device not found", http.StatusNotFound)
		return
	}
	if req.DeviceName != "" {
		device.DeviceName = req.DeviceName
	}
	if req.DeviceType != "" {
		device.DeviceType = req.DeviceType
	}
	if req.UserAgent != "" {
		device.UserAgent = req.UserAgent
	}
	device.LastSeen = time.Now()
	if err := saveDevice(device); err != nil {
		sendError(w, fmt.Sprintf("Failed to update device: %v", err), http.StatusInternalServerError)
		return
	}
	response := DeviceResponse{
		ID:         device.ID,
		ClientName: device.ClientName,
		DeviceName: device.DeviceName,
		DeviceType: device.DeviceType,
		LastSeen:   device.LastSeen,
		IPAddress:  device.IPAddress,
		UserAgent:  device.UserAgent,
		IsActive:   device.IsActive,
		CreatedAt:  device.CreatedAt,
	}
	sendSuccess(w, "Device updated successfully", response)
}
func removeDeviceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	deviceID := vars["id"]
	device, err := loadDevice(clientName, deviceID)
	if err != nil {
		sendError(w, "Device not found", http.StatusNotFound)
		return
	}
	device.IsActive = false
	device.LastSeen = time.Now()
	if err := saveDevice(device); err != nil {
		sendError(w, fmt.Sprintf("Failed to remove device: %v", err), http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Device removed successfully", map[string]interface{}{
		"device_id":   deviceID,
		"client_name": clientName,
	})
}
func getAllDevicesHandler(w http.ResponseWriter, r *http.Request) {
	if err := initDevicesDir(); err != nil {
		sendError(w, "Failed to initialize devices directory", http.StatusInternalServerError)
		return
	}
	pattern := filepath.Join(DevicesDataDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		sendError(w, "Failed to list device files", http.StatusInternalServerError)
		return
	}
	var allDevices []DeviceResponse
	for _, file := range files {
		var device database.Device
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &device); err != nil {
			continue
		}
		if device.IsActive {
			allDevices = append(allDevices, DeviceResponse{
				ID:         device.ID,
				ClientName: device.ClientName,
				DeviceName: device.DeviceName,
				DeviceType: device.DeviceType,
				LastSeen:   device.LastSeen,
				IPAddress:  device.IPAddress,
				UserAgent:  device.UserAgent,
				IsActive:   device.IsActive,
				CreatedAt:  device.CreatedAt,
			})
		}
	}
	sendSuccess(w, "All devices retrieved successfully", allDevices)
}
func cleanupInactiveDevicesHandler(w http.ResponseWriter, r *http.Request) {
	if err := initDevicesDir(); err != nil {
		sendError(w, "Failed to initialize devices directory", http.StatusInternalServerError)
		return
	}
	thresholdDays := 30
	if daysParam := r.URL.Query().Get("days"); daysParam != "" {
		if parsed := parseInt(daysParam); parsed > 0 {
			thresholdDays = parsed
		}
	}
	threshold := time.Now().AddDate(0, 0, -thresholdDays)
	pattern := filepath.Join(DevicesDataDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		sendError(w, "Failed to list device files", http.StatusInternalServerError)
		return
	}
	cleanedCount := 0
	for _, file := range files {
		var device database.Device
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &device); err != nil {
			continue
		}
		if !device.IsActive && device.LastSeen.Before(threshold) {
			os.Remove(file)
			cleanedCount++
		}
	}
	sendSuccess(w, "Device cleanup completed", map[string]interface{}{
		"cleaned_count":  cleanedCount,
		"threshold_days": thresholdDays,
	})
}
func generateDeviceID(clientName, deviceName string) string {
	return fmt.Sprintf("%s_%s_%d", clientName, deviceName, time.Now().Unix())
}
func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
func saveDevice(device *database.Device) error {
	deviceFile := filepath.Join(DevicesDataDir, device.ID+".json")
	data, err := json.MarshalIndent(device, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(deviceFile, data, 0644)
}
func loadDevice(clientName, deviceID string) (*database.Device, error) {
	deviceFile := filepath.Join(DevicesDataDir, deviceID+".json")
	data, err := os.ReadFile(deviceFile)
	if err != nil {
		return nil, err
	}
	var device database.Device
	if err := json.Unmarshal(data, &device); err != nil {
		return nil, err
	}
	if device.ClientName != clientName {
		return nil, fmt.Errorf("device does not belong to client")
	}
	return &device, nil
}
func loadClientDevices(clientName string) ([]*database.Device, error) {
	if err := initDevicesDir(); err != nil {
		return nil, err
	}
	pattern := filepath.Join(DevicesDataDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	var devices []*database.Device
	for _, file := range files {
		var device database.Device
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &device); err != nil {
			continue
		}
		if device.ClientName == clientName {
			devices = append(devices, &device)
		}
	}
	return devices, nil
}