package client
const ClientConfigTemplate = `client
dev tun
remote {{.ServerAddress}} {{.ServerPort}} {{.ServerProtocol}}
resolv-retry infinite
nobind
persist-key
persist-tun
remote-cert-tls server
cipher AES-256-GCM
auth SHA256
verb 3
<ca>
{{.CACert}}
</ca>
<cert>
{{.ClientCert}}
</cert>
<key>
{{.ClientKey}}
</key>
<tls-crypt>
{{.TLSCrypt}}
</tls-crypt>
`