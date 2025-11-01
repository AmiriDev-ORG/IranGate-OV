#!/bin/bash
set -e
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'
OPENVPN_DIR="/etc/openvpn"
EASYRSA_DIR="/etc/openvpn/easy-rsa"
CLIENT_DIR="/etc/openvpn/client"
LOG_FILE="/var/log/irangate-install.log"
DEFAULT_PORT=1194
DEFAULT_PROTOCOL="udp"
DEFAULT_DNS1="8.8.8.8"
DEFAULT_DNS2="8.8.4.4"
DEFAULT_CIPHER="AES-256-GCM"
DEFAULT_COUNTRY="IR"
DEFAULT_PROVINCE="Tehran"
DEFAULT_CITY="Tehran"
DEFAULT_ORG="IranGate"
DEFAULT_EMAIL="admin@irangate.local"
DEFAULT_OU="IranGate VPN"
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}
print_header() {
    echo -e "${CYAN}"
    echo "╔════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                    ║"
    echo "║            🚀 IranGate OpenVPN Auto Installer 🚀                  ║"
    echo "║                    Version 1.0.0                                   ║"
    echo "║                                                                    ║"
    echo "╚════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}
print_step() {
    echo -e "\n${BLUE}[STEP]${NC} $1"
    log "STEP: $1"
}
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
    log "SUCCESS: $1"
}
print_error() {
    echo -e "${RED}❌ ERROR: $1${NC}"
    log "ERROR: $1"
}
print_warning() {
    echo -e "${YELLOW}⚠️  WARNING: $1${NC}"
    log "WARNING: $1"
}
print_info() {
    echo -e "${PURPLE}ℹ️  $1${NC}"
}
check_root() {
    print_step "Checking root privileges..."
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root"
        exit 1
    fi
    print_success "Running as root"
}
check_os() {
    print_step "Checking operating system..."
    if [[ ! -f /etc/os-release ]]; then
        print_error "Cannot detect OS"
        exit 1
    fi
    source /etc/os-release
    if [[ "$ID" == "ubuntu" ]] || [[ "$ID" == "debian" ]]; then
        print_success "Detected: $PRETTY_NAME"
        PKG_MANAGER="apt"
    elif [[ "$ID" == "centos" ]] || [[ "$ID" == "rhel" ]] || [[ "$ID" == "fedora" ]]; then
        print_success "Detected: $PRETTY_NAME"
        PKG_MANAGER="yum"
    else
        print_error "Unsupported OS: $ID"
        exit 1
    fi
}
check_internet() {
    print_step "Checking internet connectivity..."
    if ping -c 1 8.8.8.8 &> /dev/null; then
        print_success "Internet connection OK"
    else
        print_error "No internet connection"
        exit 1
    fi
}
check_ports() {
    print_step "Checking if port $DEFAULT_PORT is available..."
    if netstat -tuln 2>/dev/null | grep -q ":$DEFAULT_PORT " || ss -tuln 2>/dev/null | grep -q ":$DEFAULT_PORT "; then
        print_warning "Port $DEFAULT_PORT is already in use"
        read -p "Do you want to continue anyway? (y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        print_success "Port $DEFAULT_PORT is available"
    fi
}
install_dependencies() {
    print_step "Installing required packages..."
    if [[ "$PKG_MANAGER" == "apt" ]]; then
        apt-get update -qq
        apt-get install -y -qq \
            openvpn \
            easy-rsa \
            openssl \
            ca-certificates \
            curl \
            net-tools \
            python3 2>&1 | tee -a "$LOG_FILE"
    elif [[ "$PKG_MANAGER" == "yum" ]]; then
        yum install -y \
            openvpn \
            easy-rsa \
            openssl \
            ca-certificates \
            curl \
            net-tools \
            python3 2>&1 | tee -a "$LOG_FILE"
    fi
    print_success "All packages installed"
}
setup_easyrsa_symlink() {
    print_step "Setting up EasyRSA symlink..."
    EASYRSA_BIN=$(find /usr -name easyrsa 2>/dev/null | head -1)
    if [[ -z "$EASYRSA_BIN" ]]; then
        print_error "EasyRSA binary not found"
        exit 1
    fi
    if [[ ! -L /usr/local/bin/easyrsa ]]; then
        ln -sf "$EASYRSA_BIN" /usr/local/bin/easyrsa
        print_success "EasyRSA symlink created: /usr/local/bin/easyrsa"
    else
        print_success "EasyRSA symlink already exists"
    fi
    if ! command -v easyrsa &> /dev/null; then
        print_error "EasyRSA command not available"
        exit 1
    fi
    print_success "EasyRSA is ready: $(which easyrsa)"
}
create_easyrsa_vars() {
    print_step "Creating EasyRSA vars file..."
    mkdir -p "$EASYRSA_DIR"
    cat > "$EASYRSA_DIR/vars" <<EOF
set_var EASYRSA_REQ_COUNTRY    "$DEFAULT_COUNTRY"
set_var EASYRSA_REQ_PROVINCE   "$DEFAULT_PROVINCE"
set_var EASYRSA_REQ_CITY       "$DEFAULT_CITY"
set_var EASYRSA_REQ_ORG        "$DEFAULT_ORG"
set_var EASYRSA_REQ_EMAIL      "$DEFAULT_EMAIL"
set_var EASYRSA_REQ_OU         "$DEFAULT_OU"
set_var EASYRSA_ALGO           "ec"
set_var EASYRSA_CURVE          "secp384r1"
set_var EASYRSA_DIGEST         "sha256"
set_var EASYRSA_CA_EXPIRE      3650
set_var EASYRSA_CERT_EXPIRE    825
set_var EASYRSA_BATCH          "yes"
EOF
    chmod 644 "$EASYRSA_DIR/vars"
    print_success "EasyRSA vars file created"
}
initialize_pki() {
    print_step "Initializing PKI (Public Key Infrastructure)..."
    cd "$EASYRSA_DIR"
    if [[ -d pki ]]; then
        print_warning "Old PKI found, removing..."
        rm -rf pki
    fi
    EASYRSA_BATCH=1 easyrsa init-pki 2>&1 | tee -a "$LOG_FILE"
    print_success "PKI initialized"
    print_step "Building Certificate Authority (CA)..."
    EASYRSA_BATCH=1 easyrsa build-ca nopass 2>&1 | tee -a "$LOG_FILE"
    print_success "CA certificate created"
    print_step "Generating DH parameters (this may take a while)..."
    EASYRSA_BATCH=1 easyrsa gen-dh 2>&1 | tee -a "$LOG_FILE"
    print_success "DH parameters generated"
    print_step "Building server certificate..."
    EASYRSA_BATCH=1 easyrsa build-server-full server nopass 2>&1 | tee -a "$LOG_FILE"
    print_success "Server certificate created"
    print_step "Generating TLS-crypt key..."
    openvpn --genkey --secret tc.key 2>&1 | tee -a "$LOG_FILE"
    print_success "TLS-crypt key generated"
    cp pki/ca.crt ca.crt
    cp pki/issued/server.crt server.crt
    cp pki/private/server.key server.key
    cp pki/dh.pem dh.pem
    print_success "All certificates and keys are ready"
}
create_server_config() {
    print_step "Creating OpenVPN server configuration..."
    SERVER_IP=$(hostname -I | awk '{print $1}')
    cat > "$OPENVPN_DIR/server.conf" <<EOF
port $DEFAULT_PORT
proto $DEFAULT_PROTOCOL
dev tun
ca $EASYRSA_DIR/ca.crt
cert $EASYRSA_DIR/server.crt
key $EASYRSA_DIR/server.key
dh $EASYRSA_DIR/dh.pem
crl-verify $EASYRSA_DIR/pki/crl.pem
tls-crypt $EASYRSA_DIR/tc.key
server 10.8.0.0 255.255.255.0
ifconfig-pool-persist ipp.txt
push "redirect-gateway def1 bypass-dhcp"
if [ -n "$DEFAULT_DNS1" ] && [ "$DEFAULT_DNS1" != "" ]; then
    echo "push \"dhcp-option DNS $DEFAULT_DNS1\"" >> "$OPENVPN_DIR/server.conf"
fi
if [ -n "$DEFAULT_DNS2" ] && [ "$DEFAULT_DNS2" != "" ]; then
    echo "push \"dhcp-option DNS $DEFAULT_DNS2\"" >> "$OPENVPN_DIR/server.conf"
fi
cipher $DEFAULT_CIPHER
auth SHA256
tls-version-min 1.2
keepalive 10 120
persist-key
persist-tun
user nobody
group nogroup
status /var/log/openvpn-status.log
log-append /var/log/openvpn.log
verb 3
explicit-exit-notify 1
EOF
    chmod 644 "$OPENVPN_DIR/server.conf"
    print_success "Server configuration created"
    print_info "Server IP: $SERVER_IP"
    print_info "Server Port: $DEFAULT_PORT"
}
enable_ip_forwarding() {
    print_step "Enabling IP forwarding..."
    if ! grep -q "net.ipv4.ip_forward=1" /etc/sysctl.conf; then
        echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
    fi
    sysctl -p > /dev/null 2>&1
    print_success "IP forwarding enabled"
}
disable_firewalls() {
    print_step "Disabling all firewalls..."
    systemctl stop ufw 2>/dev/null || true
    systemctl disable ufw 2>/dev/null || true
    systemctl stop fail2ban 2>/dev/null || true
    systemctl disable fail2ban 2>/dev/null || true
    iptables -F
    iptables -X
    iptables -t nat -F
    iptables -t nat -X
    iptables -t mangle -F
    iptables -t mangle -X
    iptables -P INPUT ACCEPT
    iptables -P FORWARD ACCEPT
    iptables -P OUTPUT ACCEPT
    print_success "All firewalls disabled"
}
start_openvpn() {
    print_step "Starting OpenVPN service..."
    systemctl enable openvpn@server 2>&1 | tee -a "$LOG_FILE"
    systemctl start openvpn@server 2>&1 | tee -a "$LOG_FILE"
    sleep 2
    if systemctl is-active --quiet openvpn@server; then
        print_success "OpenVPN service is running"
    else
        print_error "OpenVPN service failed to start"
        systemctl status openvpn@server --no-pager
        exit 1
    fi
}
create_client_directory() {
    print_step "Creating client directory..."
    mkdir -p "$CLIENT_DIR"
    chmod 755 "$CLIENT_DIR"
    print_success "Client directory created: $CLIENT_DIR"
}
update_irangate_settings() {
    print_step "Updating IranGate database settings..."
    IRANGATE_DB="/opt/irangate/database"
    if [[ -d "$IRANGATE_DB" ]]; then
        SERVER_IP=$(hostname -I | awk '{print $1}')
        cat > "$IRANGATE_DB/settings.json" <<EOF
{
    "server_ip": "$SERVER_IP",
    "server_port": $DEFAULT_PORT,
    "protocol": "$DEFAULT_PROTOCOL",
    "cipher": "$DEFAULT_CIPHER",
    "mtu": 1412,
    "dns": [
        "$DEFAULT_DNS1",
        "$DEFAULT_DNS2"
    ],
    "ipv6_enabled": false,
    "auto_config_mode": "dynamic",
    "backup_path": "$IRANGATE_DB/backups/",
    "log_level": 3,
    "version": "1.0.0",
    "install_date": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "last_backup": null
}
EOF
        print_success "IranGate settings updated"
    else
        print_warning "IranGate database directory not found, skipping"
    fi
}
print_summary() {
    SERVER_IP=$(hostname -I | awk '{print $1}')
    echo -e "\n${GREEN}"
    echo "╔════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                    ║"
    echo "║              🎉 Installation Complete! 🎉                         ║"
    echo "║                                                                    ║"
    echo "╚════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo -e "${CYAN}Server Information:${NC}"
    echo "  🌐 Server IP:       $SERVER_IP"
    echo "  🔌 Port:            $DEFAULT_PORT/$DEFAULT_PROTOCOL"
    echo "  🔐 Cipher:          $DEFAULT_CIPHER"
    echo "  🌍 DNS:             $DEFAULT_DNS1, $DEFAULT_DNS2"
    echo ""
    echo -e "${CYAN}Important Files:${NC}"
    echo "  📝 Server Config:   $OPENVPN_DIR/server.conf"
    echo "  🔑 CA Certificate:  $EASYRSA_DIR/ca.crt"
    echo "  🗂️  Client Configs:  $CLIENT_DIR/"
    echo "  📋 Log File:        $LOG_FILE"
    echo ""
    echo -e "${CYAN}Next Steps:${NC}"
    echo "  1. Add a client:    irangate client add <username>"
    echo "  2. Export config:   irangate client export <username>"
    echo "  3. Check status:    irangate status"
    echo "  4. View logs:       tail -f /var/log/openvpn.log"
    echo ""
    echo -e "${GREEN}✅ OpenVPN is now running and ready to accept connections!${NC}\n"
}
main() {
    print_header
    check_root
    check_os
    check_internet
    check_ports
    install_dependencies
    setup_easyrsa_symlink
    create_easyrsa_vars
    initialize_pki
    create_server_config
    enable_ip_forwarding
    disable_firewalls
    start_openvpn
    create_client_directory
    update_irangate_settings
    print_summary
}
cleanup_on_error() {
    print_error "Installation failed! Check log file: $LOG_FILE"
    exit 1
}
trap cleanup_on_error ERR
mkdir -p "$(dirname "$LOG_FILE")"
touch "$LOG_FILE"
main "$@"