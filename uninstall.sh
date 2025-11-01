#!/bin/bash
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'
print_message() {
    echo -e "${2}${1}${NC}"
}
if [[ $EUID -ne 0 ]]; then
    print_message "This script must be run as root" "$RED"
    exit 1
fi
print_message "WARNING: This will completely remove IranGate and all its components!" "$RED"
print_message "This includes:" "$YELLOW"
echo "- OpenVPN server and all client configurations"
echo "- Webpanel and all its data"
echo "- All certificates and keys"
echo "- All logs and database files"
echo "- Related system configurations"
echo ""
read -p "Are you sure you want to continue? (yes/no): " confirm
if [[ "$confirm" != "yes" ]]; then
    print_message "Uninstallation cancelled" "$YELLOW"
    exit 1
fi
print_message "Stopping services..." "$YELLOW"
systemctl stop openvpn@server
systemctl stop irangate-webpanel
systemctl disable openvpn@server
systemctl disable irangate-webpanel
print_message "Removing OpenVPN configurations and certificates..." "$YELLOW"
rm -rf /etc/openvpn/*
rm -rf /usr/share/easy-rsa/*
print_message "Removing IranGate files..." "$YELLOW"
rm -rf /opt/irangate
rm -rf /var/log/irangate
rm -rf /var/log/openvpn
print_message "Removing IranGate CLI and Menu..." "$YELLOW"
rm -f /usr/local/bin/irangate
rm -f /usr/bin/irangate
rm -rf ~/.irangate
rm -rf /root/.irangate
rm -rf /etc/irangate
rm -rf /root/ov/irangate/pkg/menu
rm -rf /opt/irangate/menu
rm -rf /opt/irangate/templates
rm -rf /etc/openvpn/ccd/*
rm -rf /opt/irangate/webpanel/templates
rm -rf /opt/irangate/database/*
rm -f /opt/irangate/webpanel/database/irangate.db
rm -f /opt/irangate/webpanel/database/init.sql
rm -f /etc/irangate/webpanel.env
rm -f /opt/irangate/webpanel/telegram_config.json
rm -f /opt/irangate/webpanel/notification_settings.json
rm -rf /opt/irangate/database/clients/*
rm -rf /etc/openvpn/clients/*
rm -rf /opt/irangate/database/backups/*
rm -f /opt/irangate/database/cron.json
print_message "Removing service files..." "$YELLOW"
rm -f /etc/systemd/system/irangate-webpanel.service
systemctl daemon-reload
print_message "Removing nginx configuration..." "$YELLOW"
rm -f /etc/nginx/sites-available/irangate
rm -f /etc/nginx/sites-enabled/irangate
systemctl reload nginx
print_message "Removing logrotate configuration..." "$YELLOW"
rm -f /etc/logrotate.d/irangate
rm -f /etc/logrotate.d/irangate-webpanel
print_message "Removing IP forwarding configuration..." "$YELLOW"
rm -f /etc/sysctl.d/99-openvpn.conf
sysctl -p
print_message "Disabling all firewalls..." "$YELLOW"
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
print_message "Do you want to remove installed packages (OpenVPN, nginx, etc.)? (yes/no): " "$YELLOW"
read -r remove_packages
if [[ "$remove_packages" == "yes" ]]; then
    print_message "Removing installed packages..." "$YELLOW"
    systemctl stop nginx
    systemctl stop fail2ban 2>/dev/null || true
    apt-get remove --purge -y openvpn easy-rsa nginx fail2ban sqlite3
    apt-get autoremove -y
    rm -rf /etc/nginx
    rm -rf /var/log/nginx
    rm -rf /var/www/html
    rm -rf /var/lib/openvpn
    rm -rf /usr/share/easy-rsa
fi
if [ -d "/root/ov/irangate" ]; then
    print_message "Do you want to remove the project source code in /root/ov/irangate? (yes/no): " "$YELLOW"
    read -r remove_source
    if [[ "$remove_source" == "yes" ]]; then
        rm -rf /root/ov/irangate
    fi
fi
print_message "IranGate has been completely removed from your system!" "$GREEN"
print_message "Note: If you want to remove installed packages (OpenVPN, etc.), you can run:" "$YELLOW"
echo "apt-get remove --purge -y openvpn easy-rsa"
echo "apt-get autoremove -y"