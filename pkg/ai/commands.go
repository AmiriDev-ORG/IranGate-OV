package ai
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/spf13/cobra"
)
func AddAICommands(rootCmd *cobra.Command) {
	aiCmd := &cobra.Command{
		Use:   "ai",
		Short: "AI-powered system management",
		Long:  "AI-powered monitoring and management of IRANGATE VPN system",
	}
	aiPath := getAIPackagePath()
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show AI agent status",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := NewAIManager(aiPath)
			if err := manager.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load AI config: %v", err)
			}
			status := manager.GetStatus()
			if running, ok := status["running"].(bool); ok {
				if running {
				} else {
				}
			}
			if configLoaded, ok := status["config_loaded"].(bool); ok {
				if configLoaded {
				} else {
				}
			}
			if config, ok := status["config"].(map[string]interface{}); ok {
				_ = config
			}
			return nil
		},
	}
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start AI monitoring agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := NewAIManager(aiPath)
			if err := manager.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load AI config: %v", err)
			}
			if err := manager.StartAI(); err != nil {
				return fmt.Errorf("failed to start AI agent: %v", err)
			}
			return nil
		},
	}
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop AI monitoring agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := NewAIManager(aiPath)
			if err := manager.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load AI config: %v", err)
			}
			if err := manager.StopAI(); err != nil {
				return fmt.Errorf("failed to stop AI agent: %v", err)
			}
			return nil
		},
	}
	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart AI monitoring agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := NewAIManager(aiPath)
			if err := manager.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load AI config: %v", err)
			}
			if err := manager.RestartAI(); err != nil {
				return fmt.Errorf("failed to restart AI agent: %v", err)
			}
			return nil
		},
	}
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage AI configuration",
		Long:  "Configure AI monitoring settings, API keys, and notifications",
	}
	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current AI configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := NewAIManager(aiPath)
			if err := manager.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load AI config: %v", err)
			}
			return nil
		},
	}
	configSetCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set AI configuration values.
Available keys:
  monitoring_interval     - Monitoring interval in seconds (default: 60)
  cpu_threshold          - CPU usage threshold percentage (default: 90)
  ram_threshold          - RAM usage threshold percentage (default: 95)
  disk_threshold         - Disk free space threshold percentage (default: 10)
  max_failed_connections - Max failed connections before alert (default: 100)
  auto_fix_enabled       - Enable automatic fixing (true/false)
  telegram_notifications - Enable Telegram notifications (true/false)
  openai_api_key        - OpenAI API key
  telegram_bot_token    - Telegram bot token
  admin_chat_id         - Telegram admin chat ID`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := NewAIManager(aiPath)
			if err := manager.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load AI config: %v", err)
			}
			key := args[0]
			value := args[1]
			if err := manager.UpdateConfig(key, value); err != nil {
				return fmt.Errorf("failed to update config: %v", err)
			}
			return nil
		},
	}
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install AI dependencies",
		Long:  "Install required Python packages for AI functionality",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := NewAIManager(aiPath)
			if err := manager.InstallDependencies(); err != nil {
				return fmt.Errorf("failed to install dependencies: %v", err)
			}
			return nil
		},
	}
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test AI functionality",
		Long:  "Test AI configuration and connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := NewAIManager(aiPath)
			if err := manager.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load AI config: %v", err)
			}
			if err := manager.checkPythonDependencies(); err != nil {
			} else {
			}
			if manager.Config.OpenAIAPIKey == "" {
			} else {
			}
			if manager.Config.TelegramBotToken == "" {
			} else {
			}
			return nil
		},
	}
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	aiCmd.AddCommand(statusCmd)
	aiCmd.AddCommand(startCmd)
	aiCmd.AddCommand(stopCmd)
	aiCmd.AddCommand(restartCmd)
	aiCmd.AddCommand(configCmd)
	aiCmd.AddCommand(installCmd)
	aiCmd.AddCommand(testCmd)
	rootCmd.AddCommand(aiCmd)
}
func getAIPackagePath() string {
	currentDir, _ := os.Getwd()
	searchPaths := []string{
		filepath.Join(currentDir, "pkg", "ai"),
		filepath.Join(currentDir, "irangate", "pkg", "ai"),
		"/etc/irangate/ai",
		"./pkg/ai",
	}
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return filepath.Join(currentDir, "pkg", "ai")
}
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "(not set)"
	}
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}
var _ = maskAPIKey