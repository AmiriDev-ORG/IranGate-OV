package main
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
)
type QRCodeRequest struct {
	Content string `json:"content"`
	Size    int    `json:"size"`
	Level   string `json:"level"`
}
type QRCodeResponse struct {
	DataURL string `json:"data_url"`
	Content string `json:"content"`
	Size    int    `json:"size"`
}
func generateClientQRCodeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	size := 256
	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if parsedSize := parseInt(sizeParam); parsedSize > 0 {
			size = parsedSize
		}
	}
	level := qrcode.Medium
	if levelParam := r.URL.Query().Get("level"); levelParam != "" {
		switch levelParam {
		case "L":
			level = qrcode.Low
		case "M":
			level = qrcode.Medium
		case "Q":
			level = qrcode.High
		case "H":
			level = qrcode.Highest
		}
	}
	configContent, err := generateClientConfigContent(clientName)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to generate config content: %v", err), http.StatusInternalServerError)
		return
	}
	pngData, err := qrcode.Encode(configContent, level, size)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to generate QR code: %v", err), http.StatusInternalServerError)
		return
	}
	dataURL := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(pngData))
	response := QRCodeResponse{
		DataURL: dataURL,
		Content: configContent,
		Size:    size,
	}
	format := r.URL.Query().Get("format")
	if format == "png" {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s_qrcode.png", clientName))
		w.Write(pngData)
		return
	}
	sendSuccess(w, "QR code generated successfully", response)
}
func generateQRCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req QRCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		sendError(w, "Content is required", http.StatusBadRequest)
		return
	}
	if req.Size == 0 {
		req.Size = 256
	}
	level := qrcode.Medium
	switch req.Level {
	case "L":
		level = qrcode.Low
	case "M":
		level = qrcode.Medium
	case "Q":
		level = qrcode.High
	case "H":
		level = qrcode.Highest
	default:
		level = qrcode.Medium
	}
	pngData, err := qrcode.Encode(req.Content, level, req.Size)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to generate QR code: %v", err), http.StatusInternalServerError)
		return
	}
	dataURL := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(pngData))
	response := QRCodeResponse{
		DataURL: dataURL,
		Content: req.Content,
		Size:    req.Size,
	}
	sendSuccess(w, "QR code generated successfully", response)
}
func generateClientConfigContent(clientName string) (string, error) {
	configPath := filepath.Join("/opt/irangate/clients", clientName+".ovpn")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join("/etc/openvpn/clients", clientName+".ovpn")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return "", fmt.Errorf("config file not found for client: %s", clientName)
		}
	}
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %v", err)
	}
	content := string(configData)
	return content, nil
}
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
func downloadClientQRCodeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientName := vars["name"]
	size := 512
	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if parsedSize := parseInt(sizeParam); parsedSize > 0 {
			size = parsedSize
		}
	}
	level := qrcode.High
	if levelParam := r.URL.Query().Get("level"); levelParam != "" {
		switch levelParam {
		case "L":
			level = qrcode.Low
		case "M":
			level = qrcode.Medium
		case "Q":
			level = qrcode.High
		case "H":
			level = qrcode.Highest
		}
	}
	configContent, err := generateClientConfigContent(clientName)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to generate config content: %v", err), http.StatusInternalServerError)
		return
	}
	pngData, err := qrcode.Encode(configContent, level, size)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to generate QR code: %v", err), http.StatusInternalServerError)
		return
	}
	filename := fmt.Sprintf("%s_qrcode.png", clientName)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pngData)))
	w.Write(pngData)
}
func batchGenerateQRCodeHandler(w http.ResponseWriter, r *http.Request) {
	clients := r.URL.Query()["clients"]
	if len(clients) == 0 {
		sendError(w, "No clients specified", http.StatusBadRequest)
		return
	}
	size := 256
	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if parsedSize := parseInt(sizeParam); parsedSize > 0 {
			size = parsedSize
		}
	}
	level := qrcode.Medium
	if levelParam := r.URL.Query().Get("level"); levelParam != "" {
		switch levelParam {
		case "L":
			level = qrcode.Low
		case "M":
			level = qrcode.Medium
		case "Q":
			level = qrcode.High
		case "H":
			level = qrcode.Highest
		}
	}
	results := make(map[string]QRCodeResponse)
	for _, clientName := range clients {
		configContent, err := generateClientConfigContent(clientName)
		if err != nil {
			continue
		}
		pngData, err := qrcode.Encode(configContent, level, size)
		if err != nil {
			continue
		}
		dataURL := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(pngData))
		results[clientName] = QRCodeResponse{
			DataURL: dataURL,
			Content: configContent,
			Size:    size,
		}
	}
	sendSuccess(w, "Batch QR codes generated successfully", map[string]interface{}{
		"results": results,
		"count":   len(results),
	})
}