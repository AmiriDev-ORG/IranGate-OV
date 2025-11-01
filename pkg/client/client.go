package client
import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/amiridev-org/irangate-ov/pkg/utils"
)
const (
	ClientConfigDir = "/etc/openvpn/client"
	EasyRSAPath     = "/etc/openvpn/easy-rsa"
)
type Manager struct {
	db *database.DB
}
func New(db *database.DB) *Manager {
	return &Manager{
		db: db,
	}
}
func (m *Manager) GetClient(name string) (*database.Client, error) {
	clients, err := m.db.GetClients()
	if err != nil {
		return nil, fmt.Errorf("failed to get clients: %v", err)
	}
	for _, c := range clients {
		if c.Name == name {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("client %s not found", name)
}
func (m *Manager) CreateClient(name string) (*database.Client, error) {
	logger := utils.GetLogger()
	logger.Infof("Creating new client: %s", name)
	dm := utils.NewDependencyManager()
	if err := dm.CheckDependency("EasyRSA"); err != nil {
		return nil, fmt.Errorf("EasyRSA dependency not satisfied: %v\n💡 Run 'irangate check-deps' for detailed information", err)
	}
	if !utils.IsValidName(name) {
		return nil, fmt.Errorf("invalid client name: %s", name)
	}
	_, err := m.GetClient(name)
	if err == nil {
		return nil, fmt.Errorf("client %s already exists", name)
	}
	client := &database.Client{
		Name:      name,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 1, 0),
		Active:    true,
		IP:        "",
		Cipher:    "AES-256-GCM",
	}
	if err := m.generateClientCertificates(name); err != nil {
		return nil, fmt.Errorf("failed to generate client certificates: %v", err)
	}
	if err := m.generateClientConfigByName(name); err != nil {
		return nil, fmt.Errorf("failed to generate client config: %v", err)
	}
	if err := m.db.AddClient(*client); err != nil {
		return nil, fmt.Errorf("failed to save client: %v", err)
	}
	return client, nil
}
func (m *Manager) ListClients() ([]database.Client, error) {
	clients, err := m.db.GetClients()
	if err != nil {
		return nil, err
	}
	trafficData := m.getTrafficDataFromOpenVPN()
	var clientsToSave []database.Client
	for i := range clients {
		if traffic, exists := trafficData[clients[i].Name]; exists {
			previousReceived := clients[i].BytesReceived
			previousSent := clients[i].BytesSent
			if traffic.BytesReceived < previousReceived || traffic.BytesSent < previousSent {
				clients[i].BytesReceived += traffic.BytesReceived
				clients[i].BytesSent += traffic.BytesSent
			} else {
				if traffic.BytesReceived > previousReceived {
					clients[i].BytesReceived = traffic.BytesReceived
				}
				if traffic.BytesSent > previousSent {
					clients[i].BytesSent = traffic.BytesSent
				}
			}
			clients[i].DataUsedMB = int64((clients[i].BytesReceived + clients[i].BytesSent) / (1024 * 1024))
			clients[i].Upload = int64(clients[i].BytesSent)
			clients[i].Download = int64(clients[i].BytesReceived)
			clients[i].LastUpdate = time.Now()
			if traffic.Connected {
				clients[i].Status = "Connected"
				now := time.Now()
				clients[i].LastConnection = &now
			} else {
				clients[i].Status = "Disconnected"
			}
			if clients[i].BytesReceived != previousReceived || clients[i].BytesSent != previousSent {
				clientsToSave = append(clientsToSave, clients[i])
			}
		}
	}
	for _, client := range clientsToSave {
		if err := m.db.UpdateClient(client.Name, &client); err != nil {
			fmt.Printf("Warning: Failed to update client %s: %v\n", client.Name, err)
		}
	}
	return clients, nil
}
func (m *Manager) RemoveClient(name string) error {
	return m.DeleteClient(name)
}
func (m *Manager) ExportClientConfig(name, outputPath string) (string, error) {
	logger := utils.GetLogger()
	logger.Infof("Exporting configuration for client: %s", name)
	_, err := m.GetClient(name)
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(ClientConfigDir, name+".ovpn")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("config file for client %s does not exist", name)
	}
	if outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}
		data, err := os.ReadFile(configPath)
		if err != nil {
			return "", fmt.Errorf("failed to read config file: %v", err)
		}
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return "", fmt.Errorf("failed to write config file: %v", err)
		}
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %v", err)
	}
	return string(data), nil
}
func (m *Manager) UpdateBandwidth(name string, bytesReceived, bytesSent uint64) error {
	client, err := m.GetClient(name)
	if err != nil {
		return err
	}
	client.BytesReceived += bytesReceived
	client.BytesSent += bytesSent
	clients, err := m.db.GetClients()
	if err != nil {
		return fmt.Errorf("failed to get clients: %v", err)
	}
	for i, c := range clients {
		if c.Name == name {
			clients[i] = *client
			break
		}
	}
	if err := m.db.SaveClients(clients); err != nil {
		return fmt.Errorf("failed to save client: %v", err)
	}
	return nil
}
func (m *Manager) DeleteClient(name string) error {
	logger := utils.GetLogger()
	logger.Infof("Deleting client: %s", name)
	_, err := m.GetClient(name)
	if err != nil {
		return err
	}
	logger.Infof("Revoking certificate for client: %s", name)
	if err := m.revokeClientCertificate(name); err != nil {
		logger.Warnf("Failed to revoke certificate for %s: %v (continuing with deletion)", name, err)
	} else {
		logger.Infof("Certificate successfully revoked for client: %s", name)
	}
	if err := m.db.RemoveClient(name); err != nil {
		return fmt.Errorf("failed to remove client from database: %v", err)
	}
	certPath := filepath.Join(ClientConfigDir, name+".crt")
	keyPath := filepath.Join(ClientConfigDir, name+".key")
	configPath := filepath.Join(ClientConfigDir, name+".ovpn")
	for _, path := range []string{certPath, keyPath, configPath} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			logger.Warnf("Failed to remove %s: %v", path, err)
		}
	}
	jsonPath := filepath.Join("/opt/irangate/database/clients", name+".json")
	if err := os.Remove(jsonPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove client JSON file %s: %v", jsonPath, err)
	}
	logger.Infof("Client %s successfully deleted and certificate revoked", name)
	return nil
}
func (m *Manager) ListExpiredClients() ([]*database.Client, error) {
	clients, err := m.db.GetClients()
	if err != nil {
		return nil, fmt.Errorf("error getting clients: %w", err)
	}
	now := time.Now()
	var expiredClients []*database.Client
	for i := range clients {
		if clients[i].ExpiryDate != nil && clients[i].ExpiryDate.Before(now) {
			client := clients[i]
			expiredClients = append(expiredClients, &client)
		}
	}
	return expiredClients, nil
}
func (m *Manager) RevokeClient(name string) error {
	_, err := m.GetClient(name)
	if err != nil {
		return err
	}
	client, err := m.db.GetClient(name)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}
	client.Status = "revoked"
	client.Active = false
	if err := m.db.UpdateClient(name, client); err != nil {
		return fmt.Errorf("failed to update client status: %v", err)
	}
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && ./easyrsa revoke %s", EasyRSAPath, name))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to revoke certificate: %v", err)
	}
	cmd = exec.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && ./easyrsa gen-crl", EasyRSAPath))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate CRL: %v", err)
	}
	return nil
}
func (m *Manager) GetExpiredClients() ([]*database.Client, error) {
	clients, err := m.db.GetClients()
	if err != nil {
		return nil, fmt.Errorf("failed to get clients: %v", err)
	}
	now := time.Now()
	var expiredClients []*database.Client
	for i := range clients {
		if !clients[i].ExpiresAt.IsZero() && clients[i].ExpiresAt.Before(now) {
			expiredClients = append(expiredClients, &clients[i])
		}
	}
	return expiredClients, nil
}
func (m *Manager) generateClientCertificates(name string) error {
	if _, err := exec.LookPath("easyrsa"); err != nil {
		return fmt.Errorf("easyrsa not found - please install OpenVPN and easy-rsa first: %v", err)
	}
	if _, err := os.Stat(EasyRSAPath); os.IsNotExist(err) {
		return fmt.Errorf("EasyRSA directory not found at %s - please initialize PKI first", EasyRSAPath)
	}
	cmd := exec.Command("easyrsa", "build-client-full", name, "nopass")
	cmd.Dir = EasyRSAPath
	cmd.Env = append(os.Environ(), "EASYRSA_BATCH=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to generate certificates: %v (%s)", err, output)
	}
	certPath := filepath.Join(EasyRSAPath, "pki/issued", name+".crt")
	keyPath := filepath.Join(EasyRSAPath, "pki/private", name+".key")
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return fmt.Errorf("client certificate was not created at %s", certPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("client key was not created at %s", keyPath)
	}
	return nil
}
func (m *Manager) generateClientConfigByName(name string) error {
	settings, err := m.db.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to get server settings: %v", err)
	}
	if settings.ServerIP == "" || settings.ServerIP == "0.0.0.0" {
		detectedIP := utils.GetServerIP()
		if detectedIP != "" && detectedIP != "YOUR_SERVER_IP" {
			settings.ServerIP = detectedIP
			m.db.UpdateSettings(*settings)
		} else {
			settings.ServerIP = "YOUR_SERVER_IP"
		}
	}
	var serverProtocol string
	if settings.Protocol != "" {
		serverProtocol = settings.Protocol
	} else {
		serverProtocol = "udp"
	}
	if settings.ServerPort == 0 {
		settings.ServerPort = 1194
	}
	caCert, err := m.readCertificateFile(filepath.Join(EasyRSAPath, "pki/ca.crt"))
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %v", err)
	}
	clientCert, err := m.readCertificateFile(filepath.Join(EasyRSAPath, "pki/issued", name+".crt"))
	if err != nil {
		return fmt.Errorf("failed to read client certificate: %v", err)
	}
	clientKey, err := m.readCertificateFile(filepath.Join(EasyRSAPath, "pki/private", name+".key"))
	if err != nil {
		return fmt.Errorf("failed to read client key: %v", err)
	}
	var serverIP string
	optConfigPath := "/opt/irangate/server_config.json"
	if _, statErr := os.Stat(optConfigPath); statErr == nil {
		var configErr error
		serverIP, configErr = utils.GetServerIPFromConfig(optConfigPath)
		if configErr != nil || serverIP == "" || serverIP == "YOUR_SERVER_IP" {
			rootConfigPath := "/root/OV-Panel/irangate/server_config.json"
			serverIP, _ = utils.GetServerIPFromConfig(rootConfigPath)
		}
	} else {
		rootConfigPath := "/root/OV-Panel/irangate/server_config.json"
		serverIP, _ = utils.GetServerIPFromConfig(rootConfigPath)
	}
	if serverIP == "" || serverIP == "YOUR_SERVER_IP" {
		serverIP = settings.ServerIP
	}
	tlsCrypt, err := m.readCertificateFile(filepath.Join(EasyRSAPath, "pki", "tc.key"))
	if err != nil {
		return fmt.Errorf("failed to read TLS-crypt key: %v", err)
	}
	data := struct {
		ServerAddress  string
		ServerPort     int
		ServerProtocol string
		ClientName     string
		CACert         string
		ClientCert     string
		ClientKey      string
		TLSCrypt       string
	}{
		ServerAddress:  serverIP,
		ServerPort:     settings.ServerPort,
		ServerProtocol: serverProtocol,
		ClientName:     name,
		CACert:         caCert,
		ClientCert:     clientCert,
		ClientKey:      clientKey,
		TLSCrypt:       tlsCrypt,
	}
	tmpl, err := template.New("client.ovpn").Parse(ClientConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse client template: %v", err)
	}
	configPath := filepath.Join(ClientConfigDir, name+".ovpn")
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create client config file: %v", err)
	}
	defer file.Close()
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to generate client config: %v", err)
	}
	return nil
}
func (m *Manager) readCertificateFile(filePath string) (string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "# Certificate file not found - please generate certificates first", nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
func (m *Manager) revokeClientCertificate(name string) error {
	cmd := exec.Command("easyrsa", "revoke", name)
	cmd.Dir = EasyRSAPath
	cmd.Env = append(os.Environ(), "EASYRSA_BATCH=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to revoke certificate: %v (%s)", err, output)
	}
	cmd = exec.Command("easyrsa", "gen-crl")
	cmd.Dir = EasyRSAPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to generate CRL: %v (%s)", err, output)
	}
	return nil
}
func (m *Manager) ForceRevokeCertificate(name string) error {
	logger := utils.GetLogger()
	logger.Infof("Force revoking certificate: %s", name)
	if err := m.revokeClientCertificate(name); err != nil {
		return fmt.Errorf("failed to revoke certificate: %v", err)
	}
	logger.Infof("Certificate %s successfully revoked", name)
	return nil
}
func (m *Manager) ListOrphanedCertificates() ([]string, error) {
	logger := utils.GetLogger()
	logger.Info("Checking for orphaned certificates...")
	indexPath := filepath.Join(EasyRSAPath, "pki/index.txt")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PKI index: %v", err)
	}
	dbClients, err := m.db.GetClients()
	if err != nil {
		return nil, fmt.Errorf("failed to get database clients: %v", err)
	}
	dbClientMap := make(map[string]bool)
	for _, client := range dbClients {
		dbClientMap[client.Name] = true
	}
	var orphaned []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 6 {
			continue
		}
		status := fields[0]
		cn := fields[5]
		if idx := strings.Index(cn, "CN="); idx >= 0 {
			cn = cn[idx+3:]
		}
		if cn == "server" {
			continue
		}
		if status == "V" && !dbClientMap[cn] {
			orphaned = append(orphaned, cn)
		}
	}
	return orphaned, nil
}
func (m *Manager) CleanupOrphanedCertificates() ([]string, error) {
	logger := utils.GetLogger()
	logger.Info("Starting orphaned certificate cleanup...")
	orphaned, err := m.ListOrphanedCertificates()
	if err != nil {
		return nil, err
	}
	if len(orphaned) == 0 {
		logger.Info("No orphaned certificates found")
		return nil, nil
	}
	logger.Infof("Found %d orphaned certificates to revoke", len(orphaned))
	var revoked []string
	var failed []string
	for _, name := range orphaned {
		logger.Infof("Revoking orphaned certificate: %s", name)
		if err := m.revokeClientCertificate(name); err != nil {
			logger.Errorf("Failed to revoke %s: %v", name, err)
			failed = append(failed, name)
		} else {
			revoked = append(revoked, name)
		}
	}
	logger.Infof("Cleanup complete: %d revoked, %d failed", len(revoked), len(failed))
	if len(failed) > 0 {
		return revoked, fmt.Errorf("failed to revoke some certificates: %v", failed)
	}
	return revoked, nil
}
func (m *Manager) GetCertificateStatus(name string) (string, error) {
	indexPath := filepath.Join(EasyRSAPath, "pki/index.txt")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return "", fmt.Errorf("failed to read PKI index: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CN="+name) {
			fields := strings.Split(line, "\t")
			if len(fields) > 0 {
				switch fields[0] {
				case "V":
					return "Valid", nil
				case "R":
					return "Revoked", nil
				case "E":
					return "Expired", nil
				}
			}
		}
	}
	return "Not Found", nil
}
type TrafficData struct {
	BytesReceived uint64
	BytesSent     uint64
	Connected     bool
	RealAddress   string
	VirtualIP     string
}
func (m *Manager) getTrafficDataFromOpenVPN() map[string]TrafficData {
	trafficData := make(map[string]TrafficData)
	statusFiles := []string{
		"/etc/openvpn/openvpn-status-server.log",
		"/etc/openvpn/openvpn-status-server-tcp-443.log",
	}
	for _, statusFile := range statusFiles {
		file, err := os.Open(statusFile)
		if err != nil {
			continue
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		inClientList := false
		inRoutingTable := false
		routingTable := make(map[string]string)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "OpenVPN CLIENT LIST" {
				inClientList = true
				inRoutingTable = false
				continue
			}
			if line == "ROUTING TABLE" {
				inClientList = false
				inRoutingTable = true
				continue
			}
			if line == "GLOBAL STATS" || line == "END" {
				inClientList = false
				inRoutingTable = false
				continue
			}
			if inClientList && line != "Updated," && !strings.HasPrefix(line, "Common Name,") {
				parts := strings.Split(line, ",")
				if len(parts) >= 5 {
					clientName := strings.TrimSpace(parts[0])
					realAddress := strings.TrimSpace(parts[1])
					var bytesReceived, bytesSent uint64
					if bytes, err := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64); err == nil {
						bytesReceived = bytes
					}
					if bytes, err := strconv.ParseUint(strings.TrimSpace(parts[3]), 10, 64); err == nil {
						bytesSent = bytes
					}
					trafficData[clientName] = TrafficData{
						BytesReceived: bytesReceived,
						BytesSent:     bytesSent,
						Connected:     true,
						RealAddress:   realAddress,
					}
				}
			}
			if inRoutingTable && line != "Virtual Address," && !strings.HasPrefix(line, "Virtual Address,") {
				parts := strings.Split(line, ",")
				if len(parts) >= 2 {
					virtualIP := strings.TrimSpace(parts[0])
					commonName := strings.TrimSpace(parts[1])
					routingTable[commonName] = virtualIP
				}
			}
		}
		for clientName, data := range trafficData {
			if virtualIP, exists := routingTable[clientName]; exists {
				data.VirtualIP = virtualIP
				trafficData[clientName] = data
			}
		}
	}
	return trafficData
}