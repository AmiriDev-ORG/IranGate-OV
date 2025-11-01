package menu
import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"github.com/fatih/color"
)
type MenuItem struct {
	ID          int
	Title       string
	Description string
	Handler     func()
	Icon        string
}
type Menu struct {
	Title    string
	Items    []MenuItem
	Selected int
}
func NewMenu(title string) *Menu {
	return &Menu{
		Title:    title,
		Items:    make([]MenuItem, 0),
		Selected: 0,
	}
}
func (m *Menu) AddItem(id int, title, description, icon string, handler func()) {
	item := MenuItem{
		ID:          id,
		Title:       title,
		Description: description,
		Icon:        icon,
		Handler:     handler,
	}
	m.Items = append(m.Items, item)
}
func (m *Menu) Display() {
	ClearScreen()
	DisplayIranGateLogo()
	titleColor := color.New(color.FgCyan, color.Bold)
	titleColor.Printf("\n║  %s  ║\n", m.Title)
	fmt.Println("║                                                                              ║")
	fmt.Println("║  ┌─────────────────────────────────────────────────────────────────────┐     ║")
	for i, item := range m.Items {
		var line string
		if m.Selected == i {
			selectedColor := color.New(color.FgYellow, color.Bold)
			line = selectedColor.Sprintf("║  │  [%d] %s %s", item.ID, item.Icon, item.Title)
		} else {
			line = fmt.Sprintf("║  │  [%d] %s %s", item.ID, item.Icon, item.Title)
		}
		padding := 71 - len(line) + 5
		for j := 0; j < padding; j++ {
			line += " "
		}
		line += "│     ║"
		fmt.Println(line)
	}
	fmt.Println("║  └─────────────────────────────────────────────────────────────────────┘     ║")
	fmt.Println("║                                                                              ║")
	fmt.Println("║  Enter your choice [0-9]: _                                                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
}
func (m *Menu) GetUserInput() int {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter your choice [0-9]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			fmt.Println("Please enter a number.")
			continue
		}
		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}
		for _, item := range m.Items {
			if item.ID == choice {
				return choice
			}
		}
		fmt.Println("Invalid choice. Please try again.")
	}
}
func (m *Menu) Run() {
	for {
		m.Display()
		choice := m.GetUserInput()
		for _, item := range m.Items {
			if item.ID == choice {
				if item.Handler != nil {
					item.Handler()
				}
				return
			}
		}
	}
}
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}
func DisplayIranGateLogo() {
	cyan := color.New(color.FgCyan, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                                              ║")
	cyan.Println("║    ██╗██████╗  █████╗ ███╗   ██╗ ██████╗  █████╗ ████████╗███████╗           ║")
	cyan.Println("║    ██║██╔══██╗██╔══██╗████╗  ██║██╔════╝ ██╔══██╗╚══██╔══╝██╔════╝           ║")
	cyan.Println("║    ██║██████╔╝███████║██╔██╗ ██║██║  ███╗███████║   ██║   █████╗             ║")
	cyan.Println("║    ██║██╔══██╗██╔══██║██║╚██╗██║██║   ██║██╔══██║   ██║   ██╔══╝             ║")
	cyan.Println("║    ██║██║  ██║██║  ██║██║ ╚████║╚██████╔╝██║  ██║   ██║   ███████╗           ║")
	cyan.Println("║    ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝           ║")
	fmt.Println("║                                                                              ║")
	yellow.Println("║    ╔═══════════════════════════════════════════════════════════════════╗     ║")
	yellow.Println("║    ║              🚀 VPN Gateway Management System 🚀                  ║     ║")
	yellow.Println("║    ╚═══════════════════════════════════════════════════════════════════╝     ║")
	fmt.Println("║                                                                              ║")
}
func WaitForEnter() {
	fmt.Print("\nPress Enter to continue...")
	fmt.Scanln()
}
type MainMenu struct {
	*Menu
}
func NewMainMenu() *MainMenu {
	menu := NewMenu("🚀 IRANGATE VPN MANAGER")
	mainMenu := &MainMenu{Menu: menu}
	mainMenu.setupMenuItems()
	return mainMenu
}
func (m *MainMenu) setupMenuItems() {
	m.AddItem(1, "System Status & Dashboard", "View system status and dashboard", "📊", m.handleSystemStatus)
	m.AddItem(2, "Client Management", "Manage VPN clients", "👥", m.handleClientManagement)
	m.AddItem(3, "Server Configuration", "Configure VPN server settings", "⚙️", m.handleServerConfiguration)
	m.AddItem(4, "Monitoring & Statistics", "View monitoring data and statistics", "📈", m.handleMonitoring)
	m.AddItem(5, "AI Management & Auto-Healing", "Manage AI agent and auto-healing", "🤖", m.handleAIManagement)
	m.AddItem(6, "Scheduled Tasks & Cron Jobs", "Manage scheduled tasks", "📅", m.handleScheduledTasks)
	m.AddItem(7, "Server Control", "Start/Stop/Restart server", "🔧", m.handleServerControl)
	m.AddItem(8, "Install Web Panel", "Install and start the web panel", "🌐", m.handleInstallWebPanel)
	m.AddItem(9, "Web Panel Settings", "Manage Web Panel status and settings", "🛠️", m.handleWebPanelSettings)
	m.AddItem(10, "Uninstall Services", "Uninstall OpenVPN and Web Panel services", "🗑️", m.handleUninstallServices)
	m.AddItem(11, "Quick Actions", "Quick access to common actions", "⚡", m.handleQuickActions)
	m.AddItem(0, "Exit", "Exit the application", "🚪", m.handleExit)
}
func (m *MainMenu) handleSystemStatus() {
	fmt.Println("System Status Dashboard")
	fmt.Println("Use 'irangate status' to check server status")
	fmt.Println("Use 'irangate monitor live' for real-time monitoring")
	WaitForEnter()
}
func (m *MainMenu) handleClientManagement() {
	fmt.Println("Client Management")
	fmt.Println("Use 'irangate client list' to view all clients")
	fmt.Println("Use 'irangate client add <name>' to add a new client")
	fmt.Println("Use 'irangate client remove <name>' to remove a client")
	fmt.Println("Use 'irangate client show <name>' for client details")
	WaitForEnter()
}
func (m *MainMenu) handleServerConfiguration() {
	fmt.Println("Server Configuration")
	fmt.Println("Use 'irangate config show' to view current settings")
	fmt.Println("Use 'irangate config backup' to backup configuration")
	fmt.Println("Use 'irangate config restore <file>' to restore from backup")
	WaitForEnter()
}
func (m *MainMenu) handleMonitoring() {
	fmt.Println("Monitoring & Statistics")
	fmt.Println("Use 'irangate monitor live' for real-time monitoring")
	fmt.Println("Use 'irangate status' to check server status")
	WaitForEnter()
}
func (m *MainMenu) handleAIManagement() {
	fmt.Println("AI Management & Auto-Healing")
	fmt.Println("Use 'irangate ai status' to check AI agent status")
	fmt.Println("Use 'irangate ai start' to start AI monitoring")
	fmt.Println("Use 'irangate ai stop' to stop AI monitoring")
	fmt.Println("Use 'irangate ai config show' to view AI settings")
	WaitForEnter()
}
func (m *MainMenu) handleScheduledTasks() {
	fmt.Println("Scheduled Tasks & Cron Jobs")
	fmt.Println("Use 'irangate cron list' to view all scheduled jobs")
	fmt.Println("Use 'irangate cron add <id> <schedule> <action>' to add a job")
	fmt.Println("Use 'irangate cron remove <id>' to remove a job")
	WaitForEnter()
}
func (m *MainMenu) handleServerControl() {
	fmt.Println("Server Control")
	fmt.Println("Use 'irangate start' to start the VPN server")
	fmt.Println("Use 'irangate stop' to stop the VPN server")
	fmt.Println("Use 'irangate restart' to restart the VPN server")
	fmt.Println("Use 'irangate status' to check server status")
	WaitForEnter()
}
func (m *MainMenu) handleInstallWebPanel() {
	fmt.Println("Web Panel Installer")
	fmt.Println("This will install and start the IranGate Web Panel service.")
	fmt.Println("You will be asked for admin username/password, web path, and port.")
	possiblePaths := []string{
		"/root/OV-Panel/irangate/webpanel/install_webpanel.sh",
		"/root/ov/irangate/webpanel/install_webpanel.sh",
		"./webpanel/install_webpanel.sh",
	}
	var scriptPath string
	var found bool
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			scriptPath = path
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("Error: installer not found. Tried:\n")
		for _, path := range possiblePaths {
			fmt.Printf("  - %s\n", path)
		}
		WaitForEnter()
		return
	}
	cmd := exec.Command("bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Printf("Installer failed: %v\n", err)
	} else {
		fmt.Println("Web Panel installation finished.")
	}
	WaitForEnter()
}
func (m *MainMenu) handleWebPanelSettings() {
	for {
		ClearScreen()
		fmt.Println("Web Panel Settings")
		status := webPanelServiceStatus()
		statusColor := color.New(color.FgRed)
		statusText := "stopped"
		if status == "active" {
			statusColor = color.New(color.FgGreen)
			statusText = "running"
		}
		port, base := readWebpanelEnv()
		ip := detectIP()
		url := fmt.Sprintf("http://%s:%s", ip, port)
		if base != "/" {
			url = fmt.Sprintf("http://%s:%s%s", ip, port, base)
		}
		fmt.Printf("\n- Web Panel Status: %s\n", statusColor.Sprint(statusText))
		fmt.Printf("- URL: %s\n\n", url)
		fmt.Println("1) Start WebPanel Service")
		fmt.Println("2) Stop WebPanel Service")
		fmt.Println("3) Change Current WebPath")
		fmt.Println("4) Change Admin Credentials")
		fmt.Println("0) Back")
		fmt.Print("\nChoose an option: ")
		var choice string
		fmt.Scanln(&choice)
		switch strings.TrimSpace(choice) {
		case "1":
			runCmdInteractive("systemctl", "start", "irangate-webpanel")
		case "2":
			runCmdInteractive("systemctl", "stop", "irangate-webpanel")
		case "3":
			changeWebPath()
		case "4":
			changeAdminCredentials()
		case "0":
			return
		default:
			fmt.Println("Invalid choice")
			WaitForEnter()
		}
	}
}
func (m *MainMenu) handleQuickActions() {
	fmt.Println("Quick Actions")
	fmt.Println("Use 'irangate --help' to see all available commands")
	fmt.Println("Use 'irangate status' for quick server status check")
	fmt.Println("Use 'irangate client list' to quickly view all clients")
	fmt.Println("Use 'irangate monitor live' for real-time monitoring")
	WaitForEnter()
}
func runCmdInteractive(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}
func (m *MainMenu) handleUninstallServices() {
	ClearScreen()
	fmt.Println("Uninstall Services")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("⚠️  WARNING: This will uninstall OpenVPN server and Web Panel services.")
	fmt.Println("   Project files will be preserved.")
	fmt.Println()
	fmt.Println("This will:")
	fmt.Println("• Stop and remove Web Panel service")
	fmt.Println("• Stop and remove OpenVPN service")
	fmt.Println("• Uninstall OpenVPN packages")
	fmt.Println("• Remove systemd service files")
	fmt.Println("• Keep all project files intact")
	fmt.Println()
	fmt.Print("Are you sure you want to continue? (yes/no): ")
	var confirmation string
	fmt.Scanln(&confirmation)
	if strings.ToLower(strings.TrimSpace(confirmation)) != "yes" {
		fmt.Println("Uninstall cancelled.")
		WaitForEnter()
		return
	}
	fmt.Println("\nStarting uninstall process...")
	fmt.Println("1. Stopping Web Panel service...")
	runCmdInteractive("systemctl", "stop", "irangate-webpanel")
	runCmdInteractive("systemctl", "disable", "irangate-webpanel")
	fmt.Println("2. Removing Web Panel systemd service...")
	if err := os.Remove("/etc/systemd/system/irangate-webpanel.service"); err != nil {
		fmt.Printf("   Warning: Could not remove systemd file: %v\n", err)
	}
	fmt.Println("3. Removing Web Panel binary...")
	if err := os.Remove("/usr/local/bin/irangate-webpanel"); err != nil {
		fmt.Printf("   Warning: Could not remove binary: %v\n", err)
	}
	fmt.Println("4. Stopping OpenVPN service...")
	runCmdInteractive("systemctl", "stop", "openvpn")
	runCmdInteractive("systemctl", "disable", "openvpn")
	fmt.Println("5. Removing OpenVPN systemd services...")
	runCmdInteractive("rm", "-f", "/etc/systemd/system/openvpn.service")
	runCmdInteractive("rm", "-f", "/etc/systemd/system/openvpn@.service")
	fmt.Println("6. Uninstalling OpenVPN packages...")
	runCmdInteractive("apt-get", "remove", "--purge", "-y", "openvpn", "easy-rsa")
	fmt.Println("7. Removing OpenVPN configuration...")
	runCmdInteractive("rm", "-rf", "/etc/openvpn")
	runCmdInteractive("rm", "-rf", "/var/log/openvpn.log")
	fmt.Println("8. Reloading systemd...")
	runCmdInteractive("systemctl", "daemon-reload")
	fmt.Println("\n✅ Uninstall completed!")
	fmt.Println("📁 Project files preserved in:", "/root/ov/irangate/")
	fmt.Println("🔄 To reinstall, run: irangate install")
	fmt.Println("🌐 To reinstall web panel, use menu option 9")
	WaitForEnter()
}
func readWebpanelEnv() (string, string) {
	port := "8080"
	base := "/"
	data, err := os.ReadFile("/etc/irangate/webpanel.env")
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "WEBPANEL_PORT=") {
				port = strings.TrimPrefix(line, "WEBPANEL_PORT=")
			}
			if strings.HasPrefix(line, "WEBPANEL_BASEPATH=") {
				base = strings.TrimPrefix(line, "WEBPANEL_BASEPATH=")
				if base == "" {
					base = "/"
				}
			}
		}
	}
	return port, base
}
func detectIP() string {
	out, err := exec.Command("bash", "-lc", "hostname -I | awk '{print $1}'").Output()
	if err != nil {
		return "127.0.0.1"
	}
	ip := strings.TrimSpace(string(out))
	if ip == "" {
		return "127.0.0.1"
	}
	return ip
}
func changeWebPath() {
	port, base := readWebpanelEnv()
	fmt.Printf("Current base path: %s\n", base)
	fmt.Print("Enter new web path (no leading slash, empty for root): ")
	var wp string
	fmt.Scanln(&wp)
	wp = strings.TrimSpace(wp)
	newBase := "/"
	if wp != "" {
		newBase = "/" + strings.TrimLeft(wp, "/")
	}
	data, err := os.ReadFile("/etc/irangate/webpanel.env")
	if err != nil {
		fmt.Printf("Failed to read env: %v\n", err)
		WaitForEnter()
		return
	}
	lines := strings.Split(string(data), "\n")
	found := false
	for i := range lines {
		if strings.HasPrefix(lines[i], "WEBPANEL_BASEPATH=") {
			lines[i] = "WEBPANEL_BASEPATH=" + newBase
			found = true
		}
	}
	if !found {
		lines = append(lines, "WEBPANEL_BASEPATH="+newBase)
	}
	hasPort := false
	for i := range lines {
		if strings.HasPrefix(lines[i], "WEBPANEL_PORT=") {
			hasPort = true
		}
	}
	if !hasPort {
		lines = append(lines, "WEBPANEL_PORT="+port)
	}
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile("/etc/irangate/webpanel.env", []byte(newContent), 0600); err != nil {
		fmt.Printf("Failed to write env: %v\n", err)
		WaitForEnter()
		return
	}
	fmt.Println("Base path updated. Restarting service...")
	runCmdInteractive("systemctl", "restart", "irangate-webpanel")
}
func changeAdminCredentials() {
	adminPath := "/opt/irangate/webpanel/admin.json"
	var existing struct {
		Username string `json:"username"`
	}
	if data, err := os.ReadFile(adminPath); err == nil {
		_ = json.Unmarshal(data, &existing)
		if existing.Username != "" {
			fmt.Printf("Current admin: %s\n", existing.Username)
		}
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter new admin username: ")
	uname, _ := reader.ReadString('\n')
	uname = strings.TrimSpace(uname)
	for {
		fmt.Print("Enter new admin password: ")
		pass1, _ := reader.ReadString('\n')
		pass1 = strings.TrimSpace(pass1)
		fmt.Print("Re-enter new admin password: ")
		pass2, _ := reader.ReadString('\n')
		pass2 = strings.TrimSpace(pass2)
		if pass1 == "" {
			fmt.Println("Password cannot be empty")
			continue
		}
		if pass1 != pass2 {
			fmt.Println("Passwords do not match")
			continue
		}
		if err := os.Remove(adminPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Failed to reset admin: %v\n", err)
			return
		}
		data, err := os.ReadFile("/etc/irangate/webpanel.env")
		if err != nil {
			fmt.Printf("Failed to read env: %v\n", err)
			return
		}
		lines := strings.Split(string(data), "\n")
		set := func(key, val string) {
			replaced := false
			for i := range lines {
				if strings.HasPrefix(lines[i], key+"=") {
					lines[i] = key + "=" + val
					replaced = true
				}
			}
			if !replaced {
				lines = append(lines, key+"="+val)
			}
		}
		set("WEBPANEL_ADMIN_USER", uname)
		set("WEBPANEL_ADMIN_PASS", pass1)
		if err := os.WriteFile("/etc/irangate/webpanel.env", []byte(strings.Join(lines, "\n")), 0600); err != nil {
			fmt.Printf("Failed to write env: %v\n", err)
			return
		}
		fmt.Println("Admin credentials updated. Restarting service...")
		runCmdInteractive("systemctl", "restart", "irangate-webpanel")
		break
	}
}
func webPanelServiceStatus() string {
	out, err := exec.Command("systemctl", "is-active", "irangate-webpanel").Output()
	if err != nil {
		return "inactive"
	}
	return strings.TrimSpace(string(out))
}
func (m *MainMenu) handleExit() {
	fmt.Println("Goodbye! 👋")
	os.Exit(0)
}