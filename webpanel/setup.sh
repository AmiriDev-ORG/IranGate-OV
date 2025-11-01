#!/bin/bash
echo "=========================================="
echo "  IranGate Web Panel Setup Script"
echo "=========================================="
echo ""
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}This script must be run as root${NC}"
   exit 1
fi
echo -e "${YELLOW}Step 1: Installing Go dependencies...${NC}"
cd /root/ov/irangate/webpanel/backend
go mod download
go mod tidy
echo -e "${GREEN}✓ Dependencies installed${NC}"
echo ""
echo -e "${YELLOW}Step 2: Building web panel...${NC}"
go build -o webpanel api.go websocket.go rbac.go groups.go subscriptions.go
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Web panel built successfully${NC}"
    echo -e "${GREEN}  - API server${NC}"
    echo -e "${GREEN}  - WebSocket real-time updates${NC}"
    echo -e "${GREEN}  - User management (RBAC)${NC}"
    echo -e "${GREEN}  - Client groups${NC}"
    echo -e "${GREEN}  - Subscription tracking${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi
echo ""
echo -e "${YELLOW}Step 3: Creating systemd service...${NC}"
cat > /etc/systemd/system/irangate-webpanel.service <<EOF
[Unit]
Description=IranGate Web Panel
After=network.target
[Service]
Type=simple
User=root
WorkingDirectory=/root/ov/irangate/webpanel/backend
ExecStart=/root/ov/irangate/webpanel/backend/webpanel
Restart=always
RestartSec=5
[Install]
WantedBy=multi-user.target
EOF
echo -e "${GREEN}✓ Systemd service created${NC}"
echo ""
echo -e "${YELLOW}Step 4: Creating web panel directories...${NC}"
mkdir -p /opt/irangate/webpanel/templates
chmod 755 /opt/irangate/webpanel
chmod 755 /opt/irangate/webpanel/templates
echo -e "${GREEN}✓ Directories created${NC}"
echo ""
echo -e "${YELLOW}Step 5: Enabling and starting service...${NC}"
systemctl daemon-reload
systemctl enable irangate-webpanel
systemctl start irangate-webpanel
sleep 2
if systemctl is-active --quiet irangate-webpanel; then
    echo -e "${GREEN}✓ Service started successfully${NC}"
else
    echo -e "${RED}✗ Service failed to start${NC}"
    echo "Check logs: sudo journalctl -u irangate-webpanel -f"
    exit 1
fi
echo ""
echo -e "${YELLOW}Step 6: Disabling firewalls...${NC}"
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
echo -e "${GREEN}✓ All firewalls disabled${NC}"
echo ""
echo "=========================================="
echo -e "${GREEN}  Web Panel Setup Complete! 🎉${NC}"
echo "=========================================="
echo ""
echo "Access Information:"
echo "==================="
echo "URL:      http://$(hostname -I | awk '{print $1}'):8080"
echo "Username: admin"
echo "Password: admin123"
echo ""
echo -e "${YELLOW}⚠️  IMPORTANT: Change the default password immediately!${NC}"
echo ""
echo "Service Management:"
echo "==================="
echo "Start:   sudo systemctl start irangate-webpanel"
echo "Stop:    sudo systemctl stop irangate-webpanel"
echo "Restart: sudo systemctl restart irangate-webpanel"
echo "Status:  sudo systemctl status irangate-webpanel"
echo "Logs:    sudo journalctl -u irangate-webpanel -f"
echo ""
echo "Next Steps:"
echo "==========="
echo "1. Open http://$(hostname -I | awk '{print $1}'):8080 in your browser"
echo "2. Login with admin / admin123"
echo "3. Go to Settings → Change Password"
echo "4. Configure Telegram bot (optional)"
echo "5. Set server IP/domain"
echo ""
echo "Documentation:"
echo "=============="
echo "Full guide: /root/ov/irangate/webpanel/WEBPANEL_IMPLEMENTATION_GUIDE.txt"
echo ""