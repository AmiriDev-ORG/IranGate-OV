package templates
const DefaultServerConfig = `# OpenVPN Server Configuration
port {{.Port}}
proto {{.Protocol}}
dev tun
ca /etc/openvpn/server/ca.crt
cert /etc/openvpn/server/server.crt
key /etc/openvpn/server/server.key
dh /etc/openvpn/server/dh.pem
tls-crypt /etc/openvpn/server/tc.key
crl-verify /etc/openvpn/server/crl.pem
server {{.Network}} {{.Netmask}}
ifconfig-pool-persist /var/log/openvpn/ipp.txt
push "redirect-gateway def1 bypass-dhcp"
{{if .DNS1}}push "dhcp-option DNS {{.DNS1}}"{{end}}
{{if .DNS2}}push "dhcp-option DNS {{.DNS2}}"{{end}}
keepalive 10 120
cipher AES-256-GCM
auth SHA256
user nobody
group nogroup
persist-key
persist-tun
status /var/log/openvpn/openvpn-status.log
log /var/log/openvpn/openvpn.log
verb 3
# Management interface for monitoring
management localhost {{.ManagementPort}}
# Client certificate settings
duplicate-cn
max-clients {{.MaxClients}}
# Performance settings
sndbuf {{.SendBufferSize}}
rcvbuf {{.ReceiveBufferSize}}
fast-io
`
var DefaultServerSettings = map[string]interface{}{
	"Port":              1194,
	"Protocol":          "udp",
	"Network":           "10.8.0.0",
	"Netmask":           "255.255.255.0",
	"DNS1":              "8.8.8.8",
	"DNS2":              "8.8.4.4",
	"ManagementPort":    7505,
	"MaxClients":        100,
	"SendBufferSize":    393216,
	"ReceiveBufferSize": 393216,
}