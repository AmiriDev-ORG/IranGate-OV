package cmd
import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/ai"
	"github.com/amiridev-org/irangate-ov/pkg/client"
	"github.com/amiridev-org/irangate-ov/pkg/config"
	"github.com/amiridev-org/irangate-ov/pkg/core"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/amiridev-org/irangate-ov/pkg/errors"
	"github.com/amiridev-org/irangate-ov/pkg/menu"
	"github.com/amiridev-org/irangate-ov/pkg/monitor"
	"github.com/amiridev-org/irangate-ov/pkg/scheduler"
	"github.com/amiridev-org/irangate-ov/pkg/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)
var (
	RootCmd = &cobra.Command{
		Use:   "irangate",
		Short: "IranGate OV - A secure VPN gateway manager",
		Long: `IranGate OV is a comprehensive VPN gateway management solution
that helps you secure and monitor your network traffic and OpenVPN server.
Provides CLI tools for OpenVPN management.`,
		Run: func(cmd *cobra.Command, args []string) {
			mainMenu := menu.NewMainMenu()
			mainMenu.Run()
		},
	}
	cfgFile    string
	jsonOutput bool
)
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/irangate/config.yaml)")
	RootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information for IRANGATE CLI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("IRANGATE CLI v1.0.0")
			fmt.Println("Build Date:", time.Now().Format("2006-01-02"))
			fmt.Println("Go Version:", runtime.Version())
			fmt.Println("Platform:", runtime.GOOS+"/"+runtime.GOARCH)
		},
	}
	RootCmd.AddCommand(versionCmd)
	depCheckCmd := &cobra.Command{
		Use:   "check-deps",
		Short: "Check system dependencies",
		Long:  "Check if all required system dependencies are installed and available",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔍 Checking IRANGATE system dependencies...")
			fmt.Println("===========================================")
			dm := utils.NewDependencyManager()
			if err := dm.ValidateAll(); err != nil {
				fmt.Printf("\n❌ Dependency check failed:\n%s\n", err.Error())
				os.Exit(1)
			}
			fmt.Println("\n✅ All dependencies are satisfied!")
		},
	}
	RootCmd.AddCommand(depCheckCmd)
	recoveryCmd := &cobra.Command{
		Use:   "test-recovery",
		Short: "Test recovery procedures",
		Long:  "Test the system recovery procedures for different service types",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔧 Testing IRANGATE recovery procedures...")
			fmt.Println("==========================================")
			config := errors.RecoveryConfig{
				HealthCheckInterval: 30 * time.Second,
				RecoveryTimeout:     5 * time.Minute,
				MaxRecoveryAttempts: 3,
				AutoRecover:         false,
			}
			rm := errors.NewRecoveryManager(config)
			fmt.Println("\n📊 Testing database recovery...")
			if err := rm.AttemptRecovery(context.Background(), "database"); err != nil {
				fmt.Printf("❌ Database recovery test failed: %v\n", err)
			} else {
				fmt.Println("✅ Database recovery test passed")
			}
			if runtime.GOOS == "linux" {
				fmt.Println("\n🔒 Testing VPN service recovery...")
				if err := rm.AttemptRecovery(context.Background(), "vpn"); err != nil {
					fmt.Printf("❌ VPN recovery test failed: %v\n", err)
				} else {
					fmt.Println("✅ VPN recovery test passed")
				}
			}
			fmt.Println("\n🌐 Testing API service recovery...")
			if err := rm.AttemptRecovery(context.Background(), "api"); err != nil {
				fmt.Printf("❌ API recovery test failed: %v\n", err)
			} else {
				fmt.Println("✅ API recovery test passed")
			}
			fmt.Println("\n🎯 Recovery testing completed!")
		},
	}
	RootCmd.AddCommand(recoveryCmd)
	db, err := database.New("")
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	addCoreCommands(db)
	addClientCommands(db)
	addConfigCommands(db)
	addMonitorCommands(db)
	addSchedulerCommands(db)
	ai.AddAICommands(RootCmd)
}
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc/irangate")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
func addCoreCommands(db *database.DB) {
	coreService := core.New(db)
	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install OpenVPN server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := coreService.Install(); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess("OpenVPN server installed successfully")
		},
	}
	var uninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall OpenVPN server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := coreService.Uninstall(); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess("OpenVPN server uninstalled successfully")
		},
	}
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start OpenVPN service",
		Run: func(cmd *cobra.Command, args []string) {
			if err := coreService.Start(); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess("OpenVPN service started successfully")
		},
	}
	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop OpenVPN service",
		Run: func(cmd *cobra.Command, args []string) {
			if err := coreService.Stop(); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess("OpenVPN service stopped successfully")
		},
	}
	var restartCmd = &cobra.Command{
		Use:   "restart",
		Short: "Restart OpenVPN service",
		Run: func(cmd *cobra.Command, args []string) {
			if err := coreService.Restart(); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess("OpenVPN service restarted successfully")
		},
	}
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show OpenVPN service status",
		Run: func(cmd *cobra.Command, args []string) {
			status, err := coreService.Status()
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			if jsonOutput {
				utils.PrintJSON(map[string]string{"status": status})
			} else {
				fmt.Printf("OpenVPN service status: %s\n", color.GreenString(status))
			}
		},
	}
	RootCmd.AddCommand(installCmd, uninstallCmd, startCmd, stopCmd, restartCmd, statusCmd)
}
func addClientCommands(db *database.DB) {
	clientManager := client.New(db)
	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Manage OpenVPN clients",
	}
	var addCmd = &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new client",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := clientManager.CreateClient(args[0])
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			if jsonOutput {
				utils.PrintJSON(client)
			} else {
				utils.PrintSuccess(fmt.Sprintf("Client %s created successfully", args[0]))
			}
		},
	}
	var removeCmd = &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a client",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := clientManager.RemoveClient(args[0]); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess(fmt.Sprintf("Client %s removed successfully", args[0]))
		},
	}
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all clients",
		Run: func(cmd *cobra.Command, args []string) {
			clients, err := clientManager.ListClients()
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			if jsonOutput {
				utils.PrintJSON(clients)
			} else {
				utils.PrintClientList(clients)
			}
		},
	}
	var showCmd = &cobra.Command{
		Use:   "show [name]",
		Short: "Show detailed client information",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := clientManager.GetClient(args[0])
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			if jsonOutput {
				utils.PrintJSON(client)
			} else {
				utils.PrintClientDetails(client)
			}
		},
	}
	var exportCmd = &cobra.Command{
		Use:   "export [name] [path]",
		Short: "Export client configuration to file",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			outputPath := ""
			if len(args) > 1 {
				outputPath = args[1]
			}
			config, err := clientManager.ExportClientConfig(args[0], outputPath)
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			if outputPath != "" {
				utils.PrintSuccess(fmt.Sprintf("Client config exported to: %s", outputPath))
			} else {
				fmt.Println(config)
			}
		},
	}
	var revokeCmd = &cobra.Command{
		Use:   "revoke [name]",
		Short: "Revoke a client certificate (force revocation)",
		Long:  "Force revoke a client certificate even if already removed from database. Useful for security cleanup.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := clientManager.ForceRevokeCertificate(args[0]); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess(fmt.Sprintf("Certificate for %s revoked successfully", args[0]))
			utils.PrintInfo("Client can no longer connect even with old .ovpn file")
		},
	}
	var statusCmd = &cobra.Command{
		Use:   "status [name]",
		Short: "Check certificate status in PKI",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			status, err := clientManager.GetCertificateStatus(args[0])
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			if jsonOutput {
				utils.PrintJSON(map[string]string{"client": args[0], "status": status})
			} else {
				fmt.Printf("Certificate Status for %s: %s\n", color.CyanString(args[0]), color.YellowString(status))
			}
		},
	}
	var listOrphanedCmd = &cobra.Command{
		Use:   "list-orphaned",
		Short: "List certificates not in database (security risk)",
		Long:  "Lists valid certificates that exist in PKI but not in database. These can still connect!",
		Run: func(cmd *cobra.Command, args []string) {
			orphaned, err := clientManager.ListOrphanedCertificates()
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			if len(orphaned) == 0 {
				utils.PrintSuccess("No orphaned certificates found - system is secure!")
				return
			}
			if jsonOutput {
				utils.PrintJSON(map[string]interface{}{
					"count":    len(orphaned),
					"orphaned": orphaned,
				})
			} else {
				fmt.Printf("\n⚠️  Found %d orphaned certificate(s) - SECURITY RISK!\n\n", len(orphaned))
				fmt.Println("These certificates are NOT in database but can still connect:")
				for _, name := range orphaned {
					fmt.Printf("  - %s\n", color.RedString(name))
				}
				fmt.Printf("\nRun 'irangate client cleanup-orphaned' to revoke them all.\n\n")
			}
		},
	}
	var cleanupOrphanedCmd = &cobra.Command{
		Use:   "cleanup-orphaned",
		Short: "Revoke all orphaned certificates (CRITICAL for commercial VPN)",
		Long:  "Finds and revokes all valid certificates that are not in database. Essential for security!",
		Run: func(cmd *cobra.Command, args []string) {
			utils.PrintInfo("Scanning for orphaned certificates...")
			revoked, err := clientManager.CleanupOrphanedCertificates()
			if err != nil {
				utils.PrintError(err)
				if len(revoked) > 0 {
					utils.PrintInfo(fmt.Sprintf("Partially successful: %d certificates revoked", len(revoked)))
				}
				os.Exit(1)
			}
			if len(revoked) == 0 {
				utils.PrintSuccess("No orphaned certificates found - system is secure!")
			} else {
				if jsonOutput {
					utils.PrintJSON(map[string]interface{}{
						"revoked_count": len(revoked),
						"revoked":       revoked,
					})
				} else {
					utils.PrintSuccess(fmt.Sprintf("Successfully revoked %d orphaned certificate(s):", len(revoked)))
					for _, name := range revoked {
						fmt.Printf("  ✓ %s\n", name)
					}
					fmt.Println("\n🔒 All orphaned certificates are now blocked!")
				}
			}
		},
	}
	clientCmd.AddCommand(addCmd, removeCmd, listCmd, showCmd, exportCmd, revokeCmd, statusCmd, listOrphanedCmd, cleanupOrphanedCmd)
	RootCmd.AddCommand(clientCmd)
}
func addConfigCommands(db *database.DB) {
	configManager := config.New(db)
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage OpenVPN configuration",
	}
	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := configManager.GetConfig()
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			if jsonOutput {
				utils.PrintJSON(config)
			} else {
				utils.PrintConfig(config)
			}
		},
	}
	var backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Create a backup of current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if err := db.Backup(); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess("Configuration backup created successfully")
		},
	}
	var restoreCmd = &cobra.Command{
		Use:   "restore [backup-file]",
		Short: "Restore configuration from backup file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := db.Restore(args[0]); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess("Configuration restored successfully")
		},
	}
	var setCmd = &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Long:  "Set a specific configuration value. Example: irangate config set server_port 1195",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]
			settings, err := configManager.GetConfig()
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			switch key {
			case "server_port", "port":
				port, err := utils.ParseInt(value)
				if err != nil || port < 1 || port > 65535 {
					utils.PrintError(fmt.Errorf("invalid port number: %s", value))
					os.Exit(1)
				}
				settings.ServerPort = port
			case "protocol":
				if value != "udp" && value != "tcp" {
					utils.PrintError(fmt.Errorf("protocol must be 'udp' or 'tcp'"))
					os.Exit(1)
				}
				settings.Protocol = value
			case "cipher":
				settings.Cipher = value
			case "server_ip":
				settings.ServerIP = value
			default:
				utils.PrintError(fmt.Errorf("unknown configuration key: %s", key))
				os.Exit(1)
			}
			if err := configManager.UpdateConfig(*settings); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess(fmt.Sprintf("Configuration updated: %s = %s", key, value))
			utils.PrintInfo("Restart OpenVPN service for changes to take effect: irangate restart")
		},
	}
	var changePortCmd = &cobra.Command{
		Use:   "change-port [port]",
		Short: "Change OpenVPN server port",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			port, err := utils.ParseInt(args[0])
			if err != nil || port < 1 || port > 65535 {
				utils.PrintError(fmt.Errorf("invalid port number: %s", args[0]))
				os.Exit(1)
			}
			settings, err := configManager.GetConfig()
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			oldPort := settings.ServerPort
			settings.ServerPort = port
			if err := configManager.UpdateConfig(*settings); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess(fmt.Sprintf("Port changed from %d to %d", oldPort, port))
			utils.PrintInfo("Restart OpenVPN service: irangate restart")
		},
	}
	var changeProtocolCmd = &cobra.Command{
		Use:   "change-protocol [tcp|udp]",
		Short: "Change OpenVPN protocol",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			protocol := strings.ToLower(args[0])
			if protocol != "tcp" && protocol != "udp" {
				utils.PrintError(fmt.Errorf("protocol must be 'tcp' or 'udp'"))
				os.Exit(1)
			}
			settings, err := configManager.GetConfig()
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			settings.Protocol = protocol
			if err := configManager.UpdateConfig(*settings); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess(fmt.Sprintf("Protocol changed to %s", protocol))
			utils.PrintInfo("Restart OpenVPN service: irangate restart")
		},
	}
	var editCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit configuration in text editor",
		Run: func(cmd *cobra.Command, args []string) {
			settings, err := configManager.GetConfig()
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintInfo("Current settings:")
			utils.PrintJSON(settings)
			utils.PrintInfo("\nTo edit, use: irangate config set <key> <value>")
			utils.PrintInfo("Or manually edit: /opt/irangate/database/settings.json")
		},
	}
	configCmd.AddCommand(showCmd, backupCmd, restoreCmd, setCmd, changePortCmd, changeProtocolCmd, editCmd)
	RootCmd.AddCommand(configCmd)
}
func addMonitorCommands(db *database.DB) {
	monitorService := monitor.New(db)
	monitorCmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor OpenVPN server",
	}
	var liveCmd = &cobra.Command{
		Use:   "live",
		Short: "Show real-time monitoring",
		Run: func(cmd *cobra.Command, args []string) {
			stats, err := monitorService.GetStats()
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			if jsonOutput {
				utils.PrintJSON(stats)
			} else {
				utils.PrintStats(stats)
			}
		},
	}
	monitorCmd.AddCommand(liveCmd)
	RootCmd.AddCommand(monitorCmd)
}
func addSchedulerCommands(db *database.DB) {
	schedulerService := scheduler.New(db)
	cronCmd := &cobra.Command{
		Use:   "cron",
		Short: "Manage scheduled tasks",
	}
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all scheduled jobs",
		Run: func(cmd *cobra.Command, args []string) {
			jobs, err := schedulerService.ListJobs()
			if err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			if jsonOutput {
				utils.PrintJSON(jobs)
			} else {
				utils.PrintJobs(jobs)
			}
		},
	}
	var addCmd = &cobra.Command{
		Use:   "add [id] [schedule] [action]",
		Short: "Add a new scheduled job",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			job := database.CronJob{
				ID:       args[0],
				Schedule: args[1],
				Action:   args[2],
				Params:   make(map[string]interface{}),
				Enabled:  true,
			}
			if err := schedulerService.AddJob(job); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess(fmt.Sprintf("Scheduled job '%s' added successfully", args[0]))
		},
	}
	var removeCmd = &cobra.Command{
		Use:   "remove [id]",
		Short: "Remove a scheduled job",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := schedulerService.RemoveJob(args[0]); err != nil {
				utils.PrintError(err)
				os.Exit(1)
			}
			utils.PrintSuccess(fmt.Sprintf("Scheduled job '%s' removed successfully", args[0]))
		},
	}
	cronCmd.AddCommand(listCmd, addCmd, removeCmd)
	RootCmd.AddCommand(cronCmd)
}