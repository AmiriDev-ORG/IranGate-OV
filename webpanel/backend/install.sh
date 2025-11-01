#!/bin/bash
set -e
echo "🚀 Installing IranGate OpenVPN Configuration Backend..."
if [ "$EUID" -ne 0 ]; then
    echo "❌ Please run as root (use sudo)"
    exit 1
fi
echo "📦 Updating system packages..."
apt update
echo "🐍 Installing Python3 and pip..."
apt install -y python3 python3-pip python3-venv
echo "🔧 Creating virtual environment..."
cd /root/OV-Panel/irangate/webpanel/backend
python3 -m venv venv
source venv/bin/activate
echo "📚 Installing Python dependencies..."
pip install -r requirements.txt
echo "📁 Creating directories..."
mkdir -p /var/lib/irangate
mkdir -p /etc/openvpn/server
mkdir -p /etc/openvpn/client-templates
echo "🔒 Setting permissions..."
chmod 700 /var/lib/irangate
chmod 700 /etc/openvpn
chmod 700 /etc/openvpn/server
chmod 700 /etc/openvpn/client-templates
echo "⚙️ Creating systemd service..."
cat > /etc/systemd/system/irangate-backend.service << EOF
[Unit]
Description=IranGate OpenVPN Configuration Backend
After=network.target
[Service]
Type=simple
User=root
WorkingDirectory=/root/OV-Panel/irangate/webpanel/backend
Environment=PATH=/root/OV-Panel/irangate/webpanel/backend/venv/bin
ExecStart=/root/OV-Panel/irangate/webpanel/backend/venv/bin/python app.py
Restart=always
RestartSec=10
[Install]
WantedBy=multi-user.target
EOF
echo "🔄 Enabling service..."
systemctl daemon-reload
systemctl enable irangate-backend.service
echo "▶️ Starting service..."
systemctl start irangate-backend.service
echo "📊 Checking service status..."
sleep 3
if systemctl is-active --quiet irangate-backend.service; then
    echo "✅ Service is running successfully!"
    echo "🌐 Backend API is available at: http://localhost:5000"
    echo "📋 Health check: http://localhost:5000/api/health"
else
    echo "❌ Service failed to start. Check logs with: journalctl -u irangate-backend.service"
    exit 1
fi
echo ""
echo "🎉 Installation completed successfully!"
echo ""
echo "📋 Next steps:"
echo "1. Update your frontend to use the backend API"
echo "2. Test the API endpoints"
echo "3. Create your first OpenVPN configuration"
echo ""
echo "🔧 Service management:"
echo "  Start:   systemctl start irangate-backend.service"
echo "  Stop:    systemctl stop irangate-backend.service"
echo "  Restart: systemctl restart irangate-backend.service"
echo "  Status:  systemctl status irangate-backend.service"
echo "  Logs:    journalctl -u irangate-backend.service -f"