package templates
const DefaultClientConfig = `# OpenVPN Client Configuration
client
dev tun
proto {{.Protocol}}
remote {{.ServerAddress}} {{.Port}}
resolv-retry infinite
nobind
# Certificates
ca [inline]
cert [inline]
key [inline]
tls-crypt [inline]
# Security
cipher AES-256-GCM
auth SHA256
auth-nocache
verify-x509-name "{{.ServerName}}" name
# Connection settings
persist-key
persist-tun
compress lz4-v2
# Logging
verb 3
mute 20
# Additional settings
{{if .BlockDNSLeaks}}block-outside-dns{{end}}
{{if .RedirectGateway}}redirect-gateway def1{{end}}
# Optional: Bandwidth limits
{{if .UploadLimit}}up-rate {{.UploadLimit}}{{end}}
{{if .DownloadLimit}}down-rate {{.DownloadLimit}}{{end}}
# Optional: Connection timeout
{{if .ConnTimeout}}connect-timeout {{.ConnTimeout}}{{end}}
# Optional: Keep alive settings
{{if .KeepAlive}}keepalive {{.KeepAlive}}{{end}}
# Optional: Additional routes
{{range .AdditionalRoutes}}
route {{.Network}} {{.Netmask}}{{end}}
`
var DefaultClientSettings = map[string]interface{}{
	"Protocol":         "udp",
	"Port":             1194,
	"ServerName":       "IranGate-Server",
	"BlockDNSLeaks":    true,
	"RedirectGateway":  true,
	"ConnTimeout":      60,
	"KeepAlive":        "10 60",
	"UploadLimit":      "",
	"DownloadLimit":    "",
	"AdditionalRoutes": []map[string]string{},
}