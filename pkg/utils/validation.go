package utils
import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)
var (
	validNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$`)
)
func IsValidName(name string) bool {
	return len(name) >= 2 && validNamePattern.MatchString(name)
}
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
func IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}
func IsValidProtocol(protocol string) bool {
	protocol = strings.ToLower(protocol)
	return protocol == "udp" || protocol == "tcp"
}
func ValidateServerSettings(settings map[string]interface{}) error {
	if serverIP, ok := settings["server_ip"].(string); ok {
		if serverIP != "0.0.0.0" && !IsValidIP(serverIP) {
			return fmt.Errorf("invalid server IP: %s", serverIP)
		}
	}
	if serverPort, ok := settings["server_port"].(int); ok {
		if !IsValidPort(serverPort) {
			return fmt.Errorf("invalid server port: %d", serverPort)
		}
	}
	if protocol, ok := settings["protocol"].(string); ok {
		if !IsValidProtocol(protocol) {
			return fmt.Errorf("invalid protocol: %s", protocol)
		}
	}
	return nil
}
func EnsureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}
func EnsureFilePermissions(filePath string, isConfig bool) error {
	if isConfig {
		return os.Chmod(filePath, 0644)
	} else {
		return os.Chmod(filePath, 0600)
	}
}
func ParseInt(s string) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %s", s)
	}
	return val, nil
}