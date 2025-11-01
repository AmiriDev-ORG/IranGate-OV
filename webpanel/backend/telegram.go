package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/amiridev-org/irangate-ov/pkg/database"
)

type TelegramConfig struct {
	BotToken  string     `json:"bot_token"`
	ChatID    string     `json:"chat_id"`
	Enabled   bool       `json:"enabled"`
	TestMode  bool       `json:"test_mode"`
	LastTest  *time.Time `json:"last_test,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}
type TelegramResponse struct {
	OK          bool        `json:"ok"`
	Description string      `json:"description,omitempty"`
	Result      interface{} `json:"result,omitempty"`
}
type NotificationSettings struct {
	ClientExpiring      bool `json:"client_expiring"`
	ClientQuotaExceeded bool `json:"client_quota_exceeded"`
	NewClientCreated    bool `json:"new_client_created"`
	ServerResourceWarn  bool `json:"server_resource_warn"`
	OpenVPNDown         bool `json:"openvpn_down"`
	OpenVPNRestarted    bool `json:"openvpn_restarted"`
	ExpiringDays        int  `json:"expiring_days"`
}

const TelegramConfigFile = "/opt/irangate/webpanel/telegram_config.json"

func telegramTestHandler(w http.ResponseWriter, r *http.Request) {
	config, err := loadTelegramConfig()
	if err != nil {
		sendError(w, "Failed to load Telegram configuration", http.StatusInternalServerError)
		return
	}
	if !config.Enabled {
		sendError(w, "Telegram notifications are disabled", http.StatusBadRequest)
		return
	}
	testMessage := fmt.Sprintf("🧪 *IranGate Test Message*\n\n"+
		"✅ Telegram notifications are working!\n"+
		"📅 Time: %s\n"+
		"🤖 Bot Status: Connected",
		time.Now().Format("2006-01-02 15:04:05"))
	err = sendTelegramMessage(config.BotToken, config.ChatID, testMessage)
	if err != nil {
		sendError(w, fmt.Sprintf("Failed to send test message: %v", err), http.StatusInternalServerError)
		return
	}
	now := time.Now()
	config.LastTest = &now
	config.TestMode = true
	saveTelegramConfig(config)
	sendSuccess(w, "Test message sent successfully", map[string]interface{}{
		"chat_id":   config.ChatID,
		"test_time": config.LastTest,
		"message":   "Test message sent to Telegram",
	})
}
func telegramConfigureHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BotToken string `json:"bot_token"`
		ChatID   string `json:"chat_id"`
		Enabled  bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.BotToken == "" || req.ChatID == "" {
		sendError(w, "Bot token and Chat ID are required", http.StatusBadRequest)
		return
	}
	config, err := loadTelegramConfig()
	if err != nil {
		config = &TelegramConfig{
			CreatedAt: time.Now(),
		}
	}
	config.BotToken = req.BotToken
	config.ChatID = req.ChatID
	config.Enabled = req.Enabled
	config.UpdatedAt = time.Now()
	if config.Enabled {
		testMessage := "🔧 *IranGate Configuration Test*\n\n✅ Configuration updated successfully!"
		err = sendTelegramMessage(config.BotToken, config.ChatID, testMessage)
		if err != nil {
			sendError(w, fmt.Sprintf("Configuration test failed: %v", err), http.StatusBadRequest)
			return
		}
	}
	if err := saveTelegramConfig(config); err != nil {
		sendError(w, "Failed to save configuration", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Telegram configuration updated successfully", map[string]interface{}{
		"enabled":    config.Enabled,
		"chat_id":    config.ChatID,
		"updated_at": config.UpdatedAt,
	})
}
func telegramStatusHandler(w http.ResponseWriter, r *http.Request) {
	config, err := loadTelegramConfig()
	if err != nil {
		sendError(w, "Failed to load Telegram configuration", http.StatusInternalServerError)
		return
	}
	status := map[string]interface{}{
		"enabled":    config.Enabled,
		"configured": config.BotToken != "" && config.ChatID != "",
		"last_test":  config.LastTest,
		"updated_at": config.UpdatedAt,
	}
	if config.Enabled && config.BotToken != "" && config.ChatID != "" {
		botInfo, err := getBotInfo(config.BotToken)
		if err != nil {
			status["connectivity"] = "failed"
			status["error"] = err.Error()
		} else {
			status["connectivity"] = "success"
			status["bot_info"] = botInfo
		}
	} else {
		status["connectivity"] = "not_configured"
	}
	sendSuccess(w, "Telegram status retrieved successfully", status)
}
func telegramSettingsHandler(w http.ResponseWriter, r *http.Request) {
	settingsFile := "/opt/irangate/webpanel/notification_settings.json"
	settings, err := loadNotificationSettings(settingsFile)
	if err != nil {
		settings = &NotificationSettings{
			ClientExpiring:      true,
			ClientQuotaExceeded: true,
			NewClientCreated:    true,
			ServerResourceWarn:  true,
			OpenVPNDown:         true,
			OpenVPNRestarted:    false,
			ExpiringDays:        7,
		}
	}
	sendSuccess(w, "Notification settings retrieved successfully", settings)
}
func updateTelegramSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var settings NotificationSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if settings.ExpiringDays < 1 || settings.ExpiringDays > 30 {
		sendError(w, "Expiring days must be between 1 and 30", http.StatusBadRequest)
		return
	}
	settingsFile := "/opt/irangate/webpanel/notification_settings.json"
	if err := saveNotificationSettings(settingsFile, &settings); err != nil {
		sendError(w, "Failed to save notification settings", http.StatusInternalServerError)
		return
	}
	sendSuccess(w, "Notification settings updated successfully", settings)
}
func loadTelegramConfig() (*TelegramConfig, error) {
	if _, err := os.Stat(TelegramConfigFile); os.IsNotExist(err) {
		return &TelegramConfig{
			Enabled:   false,
			CreatedAt: time.Now(),
		}, nil
	}
	data, err := os.ReadFile(TelegramConfigFile)
	if err != nil {
		return nil, err
	}
	var config TelegramConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
func saveTelegramConfig(config *TelegramConfig) error {
	dir := filepath.Dir(TelegramConfigFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(TelegramConfigFile, data, 0600)
}
func loadNotificationSettings(filename string) (*NotificationSettings, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var settings NotificationSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}
func saveNotificationSettings(filename string, settings *NotificationSettings) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
func sendTelegramMessage(botToken, chatID, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := TelegramMessage{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "Markdown",
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var response TelegramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}
	if !response.OK {
		return fmt.Errorf("telegram API error: %s", response.Description)
	}
	return nil
}
func getBotInfo(botToken string) (interface{}, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", botToken)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var response TelegramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, fmt.Errorf("telegram API error: %s", response.Description)
	}
	return response.Result, nil
}
func SendClientExpiringNotification(client *database.Client, daysLeft int) error {
	config, err := loadTelegramConfig()
	if err != nil || !config.Enabled {
		return nil
	}
	settings, err := loadNotificationSettings("/opt/irangate/webpanel/notification_settings.json")
	if err != nil || !settings.ClientExpiring {
		return nil
	}
	message := fmt.Sprintf("⚠️ *Client Expiring Soon*\n\n"+
		"👤 Client: `%s`\n"+
		"📅 Expires: %s\n"+
		"⏰ Days Left: %d\n"+
		"🔄 Action: Consider extending subscription",
		client.Name,
		client.ExpiresAt.Format("2006-01-02"),
		daysLeft)
	return sendTelegramMessage(config.BotToken, config.ChatID, message)
}
func SendQuotaExceededNotification(client *database.Client) error {
	config, err := loadTelegramConfig()
	if err != nil || !config.Enabled {
		return nil
	}
	settings, err := loadNotificationSettings("/opt/irangate/webpanel/notification_settings.json")
	if err != nil || !settings.ClientQuotaExceeded {
		return nil
	}
	usedGB := float64(client.BytesReceived+client.BytesSent) / (1024 * 1024 * 1024)
	quotaGB := float64(client.BandwidthQuotaGB)
	message := fmt.Sprintf("🚫 *Quota Exceeded*\n\n"+
		"👤 Client: `%s`\n"+
		"📊 Used: %.2f GB\n"+
		"📈 Quota: %.2f GB\n"+
		"🔄 Action: Client may need quota increase",
		client.Name,
		usedGB,
		quotaGB)
	return sendTelegramMessage(config.BotToken, config.ChatID, message)
}
func SendNewClientNotification(client *database.Client) error {
	config, err := loadTelegramConfig()
	if err != nil || !config.Enabled {
		return nil
	}
	settings, err := loadNotificationSettings("/opt/irangate/webpanel/notification_settings.json")
	if err != nil || !settings.NewClientCreated {
		return nil
	}
	message := fmt.Sprintf("✅ *New Client Created*\n\n"+
		"👤 Client: `%s`\n"+
		"📅 Created: %s\n"+
		"📅 Expires: %s\n",
		client.Name,
		client.CreatedAt.Format("2006-01-02 15:04:05"),
		client.ExpiresAt.Format("2006-01-02"))
	return sendTelegramMessage(config.BotToken, config.ChatID, message)
}
func SendServerResourceNotification(resourceType string, usage float64, threshold float64) error {
	config, err := loadTelegramConfig()
	if err != nil || !config.Enabled {
		return nil
	}
	settings, err := loadNotificationSettings("/opt/irangate/webpanel/notification_settings.json")
	if err != nil || !settings.ServerResourceWarn {
		return nil
	}
	message := fmt.Sprintf("⚠️ *Server Resource Warning*\n\n"+
		"📊 Resource: %s\n"+
		"📈 Usage: %.1f%%\n"+
		"🎯 Threshold: %.1f%%\n"+
		"🔄 Action: Monitor server resources",
		resourceType,
		usage,
		threshold)
	return sendTelegramMessage(config.BotToken, config.ChatID, message)
}
func SendOpenVPNStatusNotification(status string, details string) error {
	config, err := loadTelegramConfig()
	if err != nil || !config.Enabled {
		return nil
	}
	settings, err := loadNotificationSettings("/opt/irangate/webpanel/notification_settings.json")
	if err != nil {
		return nil
	}
	var shouldNotify bool
	var emoji string
	switch status {
	case "down":
		shouldNotify = settings.OpenVPNDown
		emoji = "🔴"
	case "restarted":
		shouldNotify = settings.OpenVPNRestarted
		emoji = "🔄"
	default:
		return nil
	}
	if !shouldNotify {
		return nil
	}
	message := fmt.Sprintf("%s *OpenVPN Service %s*\n\n"+
		"📅 Time: %s\n"+
		"📝 Details: %s",
		emoji,
		status,
		time.Now().Format("2006-01-02 15:04:05"),
		details)
	return sendTelegramMessage(config.BotToken, config.ChatID, message)
}
