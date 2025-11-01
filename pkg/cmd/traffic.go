package cmd
import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/amiridev-org/irangate-ov/pkg/traffic"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)
var (
	trafficCollector *traffic.Collector
	trafficAnalyzer  *traffic.Analyzer
	trafficStorage   *traffic.Storage
)
func init() {
	trafficCmd := &cobra.Command{
		Use:   "traffic",
		Short: "Traffic analysis and bandwidth monitoring",
		Long:  "Monitor and analyze traffic data for OpenVPN clients",
	}
	trafficCmd.AddCommand(trafficStartCmd)
	trafficCmd.AddCommand(trafficStopCmd)
	trafficCmd.AddCommand(trafficStatsCmd)
	trafficCmd.AddCommand(trafficHistoryCmd)
	trafficCmd.AddCommand(trafficTopCmd)
	trafficCmd.AddCommand(trafficExportCmd)
	trafficCmd.AddCommand(trafficLiveCmd)
	RootCmd.AddCommand(trafficCmd)
}
var trafficStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start traffic collection",
	Long:  "Start the background traffic collection service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initTraffic(); err != nil {
			fmt.Printf("Error initializing traffic system: %v\n", err)
			os.Exit(1)
		}
		if err := trafficCollector.Start(); err != nil {
			fmt.Printf("Error starting collector: %v\n", err)
			os.Exit(1)
		}
		if err := trafficStorage.Start(); err != nil {
			fmt.Printf("Error starting storage: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(color.GreenString("✅ Traffic collection started"))
		fmt.Println("Data will be collected every 10 seconds")
	},
}
var trafficStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop traffic collection",
	Long:  "Stop the background traffic collection service",
	Run: func(cmd *cobra.Command, args []string) {
		if trafficCollector != nil {
			trafficCollector.Stop()
		}
		if trafficStorage != nil {
			trafficStorage.Stop()
		}
		fmt.Println(color.YellowString("⚠️  Traffic collection stopped"))
	},
}
var trafficStatsCmd = &cobra.Command{
	Use:   "stats [client]",
	Short: "Show traffic statistics",
	Long:  "Display traffic statistics for all clients or a specific client",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initTraffic(); err != nil {
			fmt.Printf("Error initializing traffic system: %v\n", err)
			os.Exit(1)
		}
		clientName := ""
		if len(args) > 0 {
			clientName = args[0]
		}
		if clientName == "" {
			if jsonOutput {
				showAllStatsJSON()
			} else {
				showAllStats()
			}
		} else {
			if jsonOutput {
				showClientStatsJSON(clientName)
			} else {
				showClientStats(clientName)
			}
		}
	},
}
var trafficHistoryCmd = &cobra.Command{
	Use:   "history [client]",
	Short: "Show traffic history",
	Long:  "Display historical traffic data for all clients or a specific client",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initTraffic(); err != nil {
			fmt.Printf("Error initializing traffic system: %v\n", err)
			os.Exit(1)
		}
		days, _ := cmd.Flags().GetInt("days")
		format, _ := cmd.Flags().GetString("format")
		clientName := ""
		if len(args) > 0 {
			clientName = args[0]
		}
		startTime := time.Now().AddDate(0, 0, -days)
		endTime := time.Now()
		if clientName == "" {
			showHistoryAll(clientName, startTime, endTime, format)
		} else {
			showHistorySingle(clientName, startTime, endTime, format)
		}
	},
}
var trafficTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Show top traffic users",
	Long:  "Display top users by traffic usage",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initTraffic(); err != nil {
			fmt.Printf("Error initializing traffic system: %v\n", err)
			os.Exit(1)
		}
		limit, _ := cmd.Flags().GetInt("limit")
		days, _ := cmd.Flags().GetInt("days")
		if limit <= 0 {
			limit = 10
		}
		startTime := time.Now().AddDate(0, 0, -days)
		endTime := time.Now()
		showTopUsers(limit, startTime, endTime)
	},
}
var trafficExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export traffic data",
	Long:  "Export traffic data to various formats",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initTraffic(); err != nil {
			fmt.Printf("Error initializing traffic system: %v\n", err)
			os.Exit(1)
		}
		client, _ := cmd.Flags().GetString("client")
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		days, _ := cmd.Flags().GetInt("days")
		startTime := time.Now().AddDate(0, 0, -days)
		endTime := time.Now()
		exportTraffic(client, format, output, startTime, endTime)
	},
}
var trafficLiveCmd = &cobra.Command{
	Use:   "live",
	Short: "Live traffic monitoring",
	Long:  "Display real-time traffic monitoring in the terminal",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initTraffic(); err != nil {
			fmt.Printf("Error initializing traffic system: %v\n", err)
			os.Exit(1)
		}
		showLiveTraffic()
	},
}
func init() {
	trafficHistoryCmd.Flags().IntP("days", "d", 7, "Number of days to show")
	trafficHistoryCmd.Flags().StringP("format", "f", "table", "Output format: table, json, csv")
	trafficTopCmd.Flags().IntP("limit", "n", 10, "Number of top users to show")
	trafficTopCmd.Flags().IntP("days", "d", 7, "Number of days to analyze")
	trafficExportCmd.Flags().StringP("client", "c", "", "Client name (empty for all)")
	trafficExportCmd.Flags().StringP("format", "f", "json", "Export format: json, csv")
	trafficExportCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	trafficExportCmd.Flags().IntP("days", "d", 30, "Number of days to export")
}
func initTraffic() error {
	if trafficCollector != nil && trafficAnalyzer != nil && trafficStorage != nil {
		return nil
	}
	db, err := database.New("")
	if err != nil {
		return err
	}
	dataDir := "/opt/irangate/traffic"
	trafficCollector = traffic.NewCollector(db, dataDir)
	trafficAnalyzer = traffic.NewAnalyzer(dataDir)
	trafficStorage = traffic.NewStorage(dataDir)
	return nil
}
func showAllStats() {
	fmt.Println(color.CyanString("\n📊 Current Traffic Statistics"))
	fmt.Println("======================================")
	db, err := database.New("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	clients, err := db.GetClients()
	if err != nil {
		fmt.Printf("Error loading clients: %v\n", err)
		return
	}
	for _, client := range clients {
		fmt.Printf("\n%s:\n", color.YellowString(client.Name))
		fmt.Printf("  Downloaded: %s\n", formatBytes(client.Download))
		fmt.Printf("  Uploaded:   %s\n", formatBytes(client.Upload))
		fmt.Printf("  Total:      %s\n", formatBytes(client.Download+client.Upload))
		fmt.Printf("  Data Used:  %.2f MB\n", float64(client.DataUsedMB))
	}
}
func showAllStatsJSON() {
	db, err := database.New("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	clients, err := db.GetClients()
	if err != nil {
		fmt.Printf("Error loading clients: %v\n", err)
		os.Exit(1)
	}
	json.NewEncoder(os.Stdout).Encode(clients)
}
func showClientStats(clientName string) {
	db, err := database.New("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	client, err := db.GetClient(clientName)
	if err != nil {
		fmt.Printf("Client not found: %s\n", clientName)
		return
	}
	fmt.Println(color.CyanString("\n📊 Traffic Statistics for %s", clientName))
	fmt.Println("======================================")
	fmt.Printf("Downloaded: %s\n", formatBytes(client.Download))
	fmt.Printf("Uploaded:   %s\n", formatBytes(client.Upload))
	fmt.Printf("Total:      %s\n", formatBytes(client.Download+client.Upload))
	fmt.Printf("Data Used:  %.2f MB\n", float64(client.DataUsedMB))
	fmt.Printf("Last Update: %s\n", client.LastUpdate.Format(time.RFC3339))
}
func showClientStatsJSON(clientName string) {
	db, err := database.New("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	client, err := db.GetClient(clientName)
	if err != nil {
		fmt.Printf("Client not found: %s\n", clientName)
		os.Exit(1)
	}
	json.NewEncoder(os.Stdout).Encode(client)
}
func showHistoryAll(clientName string, startTime, endTime time.Time, format string) {
	fmt.Printf("History for all clients (Last %d days)\n", int(endTime.Sub(startTime).Hours()/24))
}
func showHistorySingle(clientName string, startTime, endTime time.Time, format string) {
	stats, err := trafficAnalyzer.GetClientAggregatedStats(clientName, startTime, endTime, "custom")
	if err != nil {
		fmt.Printf("Error getting history: %v\n", err)
		return
	}
	if format == "json" {
		json.NewEncoder(os.Stdout).Encode(stats)
	} else {
		fmt.Println(color.CyanString("\n📊 Traffic History for %s", clientName))
		fmt.Println("======================================")
		fmt.Printf("Period: %s to %s\n", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
		fmt.Printf("Total Downloaded: %s\n", formatBytes(uint64(stats.TotalReceived)))
		fmt.Printf("Total Uploaded:   %s\n", formatBytes(uint64(stats.TotalSent)))
		fmt.Printf("Total Traffic:    %s\n", formatBytes(uint64(stats.TotalBytes)))
		fmt.Printf("Avg Upload Rate:  %.2f Mbps\n", (stats.AvgUploadRate*8)/1000000)
		fmt.Printf("Avg Download Rate: %.2f Mbps\n", (stats.AvgDownloadRate*8)/1000000)
		fmt.Printf("Connected Time:   %.1f hours\n", stats.ConnectedTime/3600)
	}
}
func showTopUsers(limit int, startTime, endTime time.Time) {
	topClients, err := trafficAnalyzer.GetTopUsers(limit, startTime, endTime)
	if err != nil {
		fmt.Printf("Error getting top users: %v\n", err)
		return
	}
	fmt.Println(color.CyanString("\n🏆 Top %d Users by Traffic", limit))
	fmt.Println("======================================")
	for i, client := range topClients {
		fmt.Printf("%d. %s: %s\n", i+1, color.YellowString(client.ClientName), formatBytes(client.TotalBytes))
		fmt.Printf("   Downloaded: %s\n", formatBytes(client.Received))
		fmt.Printf("   Uploaded:   %s\n", formatBytes(client.Sent))
	}
}
func exportTraffic(clientName, format, output string, startTime, endTime time.Time) {
	fmt.Printf("Exporting traffic data...\n")
	fmt.Printf("Client: %s\n", clientName)
	fmt.Printf("Format: %s\n", format)
	fmt.Printf("Period: %s to %s\n", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
}
func showLiveTraffic() {
	fmt.Println(color.CyanString("\n📡 Live Traffic Monitoring"))
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println("======================================")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			stats, err := trafficCollector.GetCurrentStats()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Print("\033[H\033[2J")
			fmt.Println(color.CyanString("📡 Live Traffic Monitoring"))
			fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
			fmt.Println("======================================")
			for _, client := range stats {
				fmt.Printf("%s:\n", color.YellowString(client.ClientName))
				fmt.Printf("  Received: %s\n", formatBytes(client.BytesReceived))
				fmt.Printf("  Sent:     %s\n", formatBytes(client.BytesSent))
				fmt.Printf("  Total:    %s\n", formatBytes(client.BytesReceived+client.BytesSent))
				fmt.Println()
			}
		}
	}
}
func formatBytes(bytes interface{}) string {
	const unit = 1024
	var b int64
	switch v := bytes.(type) {
	case int64:
		b = v
	case uint64:
		b = int64(v)
	case int:
		b = int64(v)
	default:
		return "0 B"
	}
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}