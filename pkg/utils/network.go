package utils
import (
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)
func GetServerIP() string {
	if ip := getPublicIP(); ip != "" {
		return ip
	}
	if ip := getLocalIP(); ip != "" {
		return ip
	}
	return "YOUR_SERVER_IP"
}
func getPublicIP() string {
	services := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://checkip.amazonaws.com",
		"https://ifconfig.me/ip",
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	for _, service := range services {
		resp, err := client.Get(service)
		if err != nil {
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		ip := strings.TrimSpace(string(body))
		if isValidIP(ip) {
			return ip
		}
	}
	return ""
}
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return getFirstNonLoopbackIP()
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
func getFirstNonLoopbackIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
func isValidIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.To4() != nil
}