package client
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
	"github.com/amiridev-org/irangate-ov/pkg/database"
	"github.com/amiridev-org/irangate-ov/pkg/utils"
)
type BulkCreateRequest struct {
	Prefix     string                 `json:"prefix"`
	Count      int                    `json:"count"`
	ExpiryDays int                    `json:"expiry_days"`
	Settings   *database.ClientConfig `json:"settings"`
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
type BulkCreateResult struct {
	Successful []string          `json:"successful"`
	Failed     []BulkCreateError `json:"failed"`
}
type BulkCreateError struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}
func (m *Manager) BulkCreate(req BulkCreateRequest) (*BulkCreateResult, error) {
	if req.Count <= 0 {
		return nil, fmt.Errorf("invalid count: must be greater than 0")
	}
	if req.Count > 1000 {
		return nil, fmt.Errorf("count exceeds maximum allowed (1000)")
	}
	result := &BulkCreateResult{
		Successful: make([]string, 0),
		Failed:     make([]BulkCreateError, 0),
	}
	const maxWorkers = 10
	workerCount := min(maxWorkers, req.Count)
	jobs := make(chan int, req.Count)
	results := make(chan struct {
		name string
		err  error
	}, req.Count)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				name := fmt.Sprintf("%s%d", req.Prefix, idx+1)
				err := m.createSingleClient(name, req)
				results <- struct {
					name string
					err  error
				}{name, err}
			}
		}()
	}
	for i := 0; i < req.Count; i++ {
		jobs <- i
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(results)
	}()
	for res := range results {
		if res.err != nil {
			result.Failed = append(result.Failed, BulkCreateError{
				Name:  res.name,
				Error: res.err.Error(),
			})
		} else {
			result.Successful = append(result.Successful, res.name)
		}
	}
	return result, nil
}
func (m *Manager) createSingleClient(name string, req BulkCreateRequest) error {
	client := &database.Client{
		Name:      name,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 1, 0),
		Active:    true,
		Cipher:    "AES-256-GCM",
		Status:    "active",
	}
	if req.ExpiryDays > 0 {
		expiry := time.Now().AddDate(0, 0, req.ExpiryDays)
		client.ExpiryDate = &expiry
	}
	if req.Settings != nil {
		client.Config = req.Settings
	}
	if err := m.generateClientCerts(client); err != nil {
		return fmt.Errorf("certificate generation failed: %v", err)
	}
	clients, err := m.db.GetClients()
	if err != nil {
		return fmt.Errorf("failed to get clients: %v", err)
	}
	clients = append(clients, *client)
	if err := m.db.SaveClients(clients); err != nil {
		return fmt.Errorf("failed to save client: %v", err)
	}
	return nil
}
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	_, err = destFile.ReadFrom(sourceFile)
	return err
}
func (m *Manager) generateClientCerts(client *database.Client) error {
	if err := os.MkdirAll(ClientConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create client config directory: %v", err)
	}
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && ./easyrsa build-client-full %s nopass", EasyRSAPath, client.Name))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate client certificates: %v", err)
	}
	keySource := filepath.Join(EasyRSAPath, "pki/private", client.Name+".key")
	certSource := filepath.Join(EasyRSAPath, "pki/issued", client.Name+".crt")
	keyDest := filepath.Join(ClientConfigDir, client.Name+".key")
	certDest := filepath.Join(ClientConfigDir, client.Name+".crt")
	if err := copyFile(keySource, keyDest); err != nil {
		return fmt.Errorf("failed to copy client key: %v", err)
	}
	if err := copyFile(certSource, certDest); err != nil {
		return fmt.Errorf("failed to copy client certificate: %v", err)
	}
	if err := m.generateClientConfig(client); err != nil {
		return fmt.Errorf("failed to generate client config: %v", err)
	}
	return nil
}
func (m *Manager) generateClientConfig(client *database.Client) error {
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
		settings, err := m.db.GetSettings()
		if err != nil {
			return fmt.Errorf("failed to get server settings: %v", err)
		}
		serverIP = settings.ServerIP
	}
	settings, err := m.db.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to get server settings: %v", err)
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
	caCert, err := m.readCertificateFile(filepath.Join(utils.EasyRSAPath, "pki/ca.crt"))
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %v", err)
	}
	clientCert, err := m.readCertificateFile(filepath.Join(utils.EasyRSAPath, "pki/issued", client.Name+".crt"))
	if err != nil {
		return fmt.Errorf("failed to read client certificate: %v", err)
	}
	clientKey, err := m.readCertificateFile(filepath.Join(utils.EasyRSAPath, "pki/private", client.Name+".key"))
	if err != nil {
		return fmt.Errorf("failed to read client key: %v", err)
	}
	tlsCrypt, err := m.readCertificateFile(filepath.Join(utils.EasyRSAPath, "pki", "tc.key"))
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
		ClientConfig   *database.ClientConfig
	}{
		ServerAddress:  serverIP,
		ServerPort:     settings.ServerPort,
		ServerProtocol: serverProtocol,
		ClientName:     client.Name,
		CACert:         caCert,
		ClientCert:     clientCert,
		ClientKey:      clientKey,
		TLSCrypt:       tlsCrypt,
		ClientConfig:   client.Config,
	}
	tmpl, err := template.New("client.ovpn").Parse(ClientConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse client template: %v", err)
	}
	configPath := filepath.Join(ClientConfigDir, client.Name+".ovpn")
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
func (m *Manager) detectInboundPortAndProtocol() (int, string) {
	inboundDirs := []string{
		"/etc/openvpn",
		"/etc/openvpn/server",
	}
	for _, dir := range inboundDirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".conf") {
				continue
			}
			configPath := filepath.Join(dir, file.Name())
			port, protocol := m.parseInboundConfig(configPath)
			if port > 0 {
				return port, protocol
			}
		}
	}
	return 0, ""
}
func (m *Manager) parseInboundConfig(configPath string) (int, string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return 0, ""
	}
	lines := strings.Split(string(data), "\n")
	var port int
	var protocol string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "port ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if p, err := strconv.Atoi(parts[1]); err == nil {
					port = p
				}
			}
		}
		if strings.HasPrefix(line, "proto ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				protocol = parts[1]
			}
		}
	}
	return port, protocol
}