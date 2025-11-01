package core
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/amiridev-org/irangate-ov/pkg/utils"
)
const (
	StateRunning = "running"
	StateStopped = "stopped"
	StateError   = "error"
)
const (
	OpenVPNPath    = "/etc/openvpn"
	OpenVPNServer  = "/etc/openvpn/server"
	OpenVPNEasyRSA = "/etc/openvpn/easy-rsa"
)
type Service struct {
	db *database.DB
}
func New(db *database.DB) *Service {
	return &Service{db: db}
}
func (s *Service) Install() error {
	logger := utils.GetLogger()
	logger.Info("Starting OpenVPN installation...")
	if s.isInstalled() {
		return fmt.Errorf("OpenVPN is already installed")
	}
	if err := s.installOpenVPN(); err != nil {
		return fmt.Errorf("failed to install OpenVPN: %v", err)
	}
	if err := s.initializePKI(); err != nil {
		return fmt.Errorf("failed to initialize PKI: %v", err)
	}
	if err := s.createServerConfig(); err != nil {
		return fmt.Errorf("failed to create server configuration: %v", err)
	}
	settings, err := s.db.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to get settings: %v", err)
	}
	serverIP := utils.GetServerIP()
	logger.Infof("Detected server IP: %s", serverIP)
	settings.ServerIP = serverIP
	settings.InstallDate = time.Now()
	if err := s.db.UpdateSettings(*settings); err != nil {
		return fmt.Errorf("failed to update settings: %v", err)
	}
	logger.Info("OpenVPN installation completed successfully")
	return nil
}
func (s *Service) Uninstall() error {
	logger := utils.GetLogger()
	logger.Info("Starting OpenVPN uninstallation...")
	if err := s.Stop(); err != nil {
		logger.Warnf("Failed to stop OpenVPN service: %v", err)
	}
	if err := s.uninstallOpenVPN(); err != nil {
		return fmt.Errorf("failed to uninstall OpenVPN: %v", err)
	}
	dirs := []string{OpenVPNPath, OpenVPNServer, OpenVPNEasyRSA}
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			logger.Warnf("Failed to remove directory %s: %v", dir, err)
		}
	}
	logger.Info("OpenVPN uninstallation completed successfully")
	return nil
}
func (s *Service) Start() error {
	logger := utils.GetLogger()
	logger.Info("Starting OpenVPN service...")
	cmd := exec.Command("systemctl", "start", "openvpn@server")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start OpenVPN: %v (%s)", err, output)
	}
	logger.Info("OpenVPN service started successfully")
	return nil
}
func (s *Service) Stop() error {
	logger := utils.GetLogger()
	logger.Info("Stopping OpenVPN service...")
	cmd := exec.Command("systemctl", "stop", "openvpn@server")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop OpenVPN: %v (%s)", err, output)
	}
	logger.Info("OpenVPN service stopped successfully")
	return nil
}
func (s *Service) Restart() error {
	logger := utils.GetLogger()
	logger.Info("Restarting OpenVPN service...")
	cmd := exec.Command("systemctl", "restart", "openvpn@server")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restart OpenVPN: %v (%s)", err, output)
	}
	logger.Info("OpenVPN service restarted successfully")
	return nil
}
func (s *Service) Status() (string, error) {
	serviceNames := []string{"openvpn@server", "openvpn", "openvpn.service"}
	for _, serviceName := range serviceNames {
		cmd := exec.Command("systemctl", "is-active", serviceName)
		output, err := cmd.CombinedOutput()
		status := strings.TrimSpace(string(output))
		if status == "active" {
			return StateRunning, nil
		}
		if err == nil && status == "inactive" {
			return StateStopped, nil
		}
	}
	return StateStopped, nil
}
func (s *Service) isInstalled() bool {
	_, err := exec.LookPath("openvpn")
	return err == nil
}
func (s *Service) installOpenVPN() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		if _, err := os.Stat("/etc/debian_version"); err == nil {
			cmd = exec.Command("apt-get", "install", "-y", "openvpn", "easy-rsa")
		} else if _, err := os.Stat("/etc/redhat-release"); err == nil {
			cmd = exec.Command("yum", "install", "-y", "openvpn", "easy-rsa")
		} else {
			return fmt.Errorf("unsupported Linux distribution")
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("installation failed: %v (%s)", err, output)
	}
	return nil
}
func (s *Service) uninstallOpenVPN() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		if _, err := os.Stat("/etc/debian_version"); err == nil {
			cmd = exec.Command("apt-get", "remove", "-y", "openvpn", "easy-rsa")
		} else if _, err := os.Stat("/etc/redhat-release"); err == nil {
			cmd = exec.Command("yum", "remove", "-y", "openvpn", "easy-rsa")
		} else {
			return fmt.Errorf("unsupported Linux distribution")
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("uninstallation failed: %v (%s)", err, output)
	}
	return nil
}
func (s *Service) initializePKI() error {
	dirs := []string{OpenVPNPath, OpenVPNServer, OpenVPNEasyRSA}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}
	cmds := [][]string{
		{"easyrsa", "init-pki"},
		{"easyrsa", "build-ca", "nopass"},
		{"easyrsa", "gen-dh"},
		{"easyrsa", "build-server-full", "server", "nopass"},
		{"openvpn", "--genkey", "--secret", "tc.key"},
	}
	for _, cmd := range cmds {
		command := exec.Command(cmd[0], cmd[1:]...)
		command.Dir = OpenVPNEasyRSA
		command.Env = append(os.Environ(), "EASYRSA_BATCH=1")
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("PKI initialization failed at step '%s': %v (%s)", strings.Join(cmd, " "), err, output)
		}
	}
	tcKeySource := filepath.Join(OpenVPNEasyRSA, "tc.key")
	tcKeyDest := filepath.Join(OpenVPNEasyRSA, "pki", "tc.key")
	if _, err := os.Stat(tcKeySource); err == nil {
		if err := os.Rename(tcKeySource, tcKeyDest); err != nil {
			return fmt.Errorf("failed to move tc.key: %v", err)
		}
	}
	return nil
}
func (s *Service) createServerConfig() error {
	settings, err := s.db.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to get settings: %v", err)
	}
	config := fmt.Sprintf(`port %d
proto %s
dev tun
ca %s/pki/ca.crt
cert %s/pki/issued/server.crt
key %s/pki/private/server.key
dh %s/pki/dh.pem
tls-crypt %s/pki/tc.key
server 10.8.0.0 255.255.255.0
ifconfig-pool-persist ipp.txt
push "redirect-gateway def1 bypass-dhcp"
%s
keepalive 10 120
cipher %s
auth SHA256
user nobody
group nogroup
persist-key
persist-tun
status openvpn-status.log
verb 3`,
		settings.ServerPort,
		settings.Protocol,
		OpenVPNEasyRSA,
		OpenVPNEasyRSA,
		OpenVPNEasyRSA,
		OpenVPNEasyRSA,
		OpenVPNEasyRSA,
		s.getDNSPushConfig(settings.DNS),
		settings.Cipher,
	)
	configPath := "/etc/openvpn/server.conf"
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write server config: %v", err)
	}
	return nil
}
func (s *Service) getDNSPushConfig(dns []string) string {
	var config strings.Builder
	if len(dns) > 0 {
		for _, server := range dns {
			if server != "" {
				config.WriteString(fmt.Sprintf("push \"dhcp-option DNS %s\"\n", server))
			}
		}
	}
	return config.String()
}