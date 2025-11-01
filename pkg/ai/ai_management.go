package ai
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)
type AIManager struct {
	ConfigPath string
	PythonPath string
	IsRunning  bool
	Config     *AIConfig
}
type AIConfig struct {
	MonitoringInterval    int    `json:"monitoring_interval"`
	CPUThreshold          int    `json:"cpu_threshold"`
	RAMThreshold          int    `json:"ram_threshold"`
	DiskThreshold         int    `json:"disk_threshold"`
	MaxFailedConnections  int    `json:"max_failed_connections"`
	AutoFixEnabled        bool   `json:"auto_fix_enabled"`
	TelegramNotifications bool   `json:"telegram_notifications"`
	OpenAIAPIKey          string `json:"openai_api_key"`
	TelegramBotToken      string `json:"telegram_bot_token"`
	AdminChatID           string `json:"admin_chat_id"`
	LogFile               string `json:"log_file"`
	Description           string `json:"description"`
	Version               string `json:"version"`
}
func NewAIManager(configPath string) *AIManager {
	return &AIManager{
		ConfigPath: configPath,
		PythonPath: "python3",
		IsRunning:  false,
	}
}
func (am *AIManager) LoadConfig() error {
	configFile := filepath.Join(am.ConfigPath, "config.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	var config AIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}
	am.Config = &config
	return nil
}
func (am *AIManager) SaveConfig() error {
	if am.Config == nil {
		return fmt.Errorf("no config to save")
	}
	configFile := filepath.Join(am.ConfigPath, "config.json")
	data, err := json.MarshalIndent(am.Config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	return nil
}
func (am *AIManager) StartAI() error {
	if am.IsRunning {
		return fmt.Errorf("AI agent is already running")
	}
	if _, err := exec.LookPath(am.PythonPath); err != nil {
		return fmt.Errorf("Python3 not found: %v", err)
	}
	if err := am.checkPythonDependencies(); err != nil {
		return fmt.Errorf("Python dependencies not satisfied: %v", err)
	}
	scriptPath := filepath.Join(am.ConfigPath, "ai_agent.py")
	configFile := filepath.Join(am.ConfigPath, "config.json")
	cmd := exec.Command(am.PythonPath, scriptPath, configFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start AI agent: %v", err)
	}
	am.IsRunning = true
	time.Sleep(2 * time.Second)
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		am.IsRunning = false
		return fmt.Errorf("AI agent failed to start")
	}
	return nil
}
func (am *AIManager) StopAI() error {
	if !am.IsRunning {
		return fmt.Errorf("AI agent is not running")
	}
	cmd := exec.Command("pkill", "-f", "ai_agent.py")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop AI agent: %v", err)
	}
	am.IsRunning = false
	return nil
}
func (am *AIManager) RestartAI() error {
	if err := am.StopAI(); err != nil {
		return fmt.Errorf("failed to stop AI agent: %v", err)
	}
	time.Sleep(2 * time.Second)
	if err := am.StartAI(); err != nil {
		return fmt.Errorf("failed to start AI agent: %v", err)
	}
	return nil
}
func (am *AIManager) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"running":       am.IsRunning,
		"config_loaded": am.Config != nil,
	}
	if am.Config != nil {
		status["config"] = map[string]interface{}{
			"monitoring_interval":    am.Config.MonitoringInterval,
			"auto_fix_enabled":       am.Config.AutoFixEnabled,
			"telegram_notifications": am.Config.TelegramNotifications,
			"openai_configured":      am.Config.OpenAIAPIKey != "",
			"telegram_configured":    am.Config.TelegramBotToken != "",
		}
	}
	return status
}
func (am *AIManager) UpdateConfig(key, value string) error {
	if am.Config == nil {
		return fmt.Errorf("config not loaded")
	}
	switch key {
	case "monitoring_interval":
		var interval int
		if _, err := fmt.Sscanf(value, "%d", &interval); err != nil {
			return fmt.Errorf("invalid monitoring interval: %v", err)
		}
		am.Config.MonitoringInterval = interval
	case "cpu_threshold":
		var threshold int
		if _, err := fmt.Sscanf(value, "%d", &threshold); err != nil {
			return fmt.Errorf("invalid CPU threshold: %v", err)
		}
		am.Config.CPUThreshold = threshold
	case "ram_threshold":
		var threshold int
		if _, err := fmt.Sscanf(value, "%d", &threshold); err != nil {
			return fmt.Errorf("invalid RAM threshold: %v", err)
		}
		am.Config.RAMThreshold = threshold
	case "disk_threshold":
		var threshold int
		if _, err := fmt.Sscanf(value, "%d", &threshold); err != nil {
			return fmt.Errorf("invalid disk threshold: %v", err)
		}
		am.Config.DiskThreshold = threshold
	case "max_failed_connections":
		var max int
		if _, err := fmt.Sscanf(value, "%d", &max); err != nil {
			return fmt.Errorf("invalid max failed connections: %v", err)
		}
		am.Config.MaxFailedConnections = max
	case "auto_fix_enabled":
		am.Config.AutoFixEnabled = (value == "true" || value == "1")
	case "telegram_notifications":
		am.Config.TelegramNotifications = (value == "true" || value == "1")
	case "openai_api_key":
		am.Config.OpenAIAPIKey = value
	case "telegram_bot_token":
		am.Config.TelegramBotToken = value
	case "admin_chat_id":
		am.Config.AdminChatID = value
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}
	return am.SaveConfig()
}
func (am *AIManager) checkPythonDependencies() error {
	requiredPackages := []string{"openai", "psutil", "requests"}
	for _, pkg := range requiredPackages {
		cmd := exec.Command(am.PythonPath, "-c", fmt.Sprintf("import %s", pkg))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("package %s not installed", pkg)
		}
	}
	return nil
}
func (am *AIManager) InstallDependencies() error {
	requirementsFile := filepath.Join(am.ConfigPath, "requirements.txt")
	cmd := exec.Command(am.PythonPath, "-m", "pip", "install", "-r", requirementsFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install dependencies: %v", err)
	}
	return nil
}