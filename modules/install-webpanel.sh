#!/bin/bash
WEBPANEL_PORT="$1"
ADMIN_USER="$2"
ADMIN_PASS="$3"
WEBPANEL_DIR="/opt/irangate/webpanel"
install_webpanel() {
    mkdir -p $WEBPANEL_DIR
    cd $WEBPANEL_DIR
    cp -r ../frontend/* $WEBPANEL_DIR/frontend/
    cp -r ../backend/* $WEBPANEL_DIR/backend/
    if [ -f "$WEBPANEL_DIR/frontend/login.html" ]; then
        sed -i '/default credentials/d' "$WEBPANEL_DIR/frontend/login.html"
        sed -i '/Default username/d' "$WEBPANEL_DIR/frontend/login.html"
        sed -i '/Default password/d' "$WEBPANEL_DIR/frontend/login.html"
    fi
    cd $WEBPANEL_DIR/backend
    go build -o webpanel
    mkdir -p $WEBPANEL_DIR/database
    mkdir -p $WEBPANEL_DIR/logs
    mkdir -p $WEBPANEL_DIR/uploads
}
configure_nginx() {
    cat > /etc/nginx/sites-available/irangate << EOF
server {
    listen $WEBPANEL_PORT;
    server_name _;
    root $WEBPANEL_DIR/frontend;
    index index.html;
    location / {
        try_files \$uri \$uri/ /index.html;
    }
    location /api {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }
    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-XSS-Protection "1; mode=block";
    add_header X-Content-Type-Options "nosniff";
    add_header Referrer-Policy "strict-origin-when-cross-origin";
    add_header Content-Security-Policy "default-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' data:;";
    access_log /var/log/nginx/irangate-access.log;
    error_log /var/log/nginx/irangate-error.log;
}
EOF
    ln -sf /etc/nginx/sites-available/irangate /etc/nginx/sites-enabled/
    rm -f /etc/nginx/sites-enabled/default
    nginx -t && systemctl reload nginx
}
create_service() {
    cat > /etc/systemd/system/irangate-webpanel.service << EOF
[Unit]
Description=IranGate Webpanel
After=network.target
[Service]
Type=simple
User=root
WorkingDirectory=$WEBPANEL_DIR/backend
ExecStart=$WEBPANEL_DIR/backend/webpanel
Restart=always
RestartSec=10
Environment=PORT=8000
Environment=ADMIN_USER=$ADMIN_USER
Environment=ADMIN_PASS=$ADMIN_PASS
[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    systemctl enable irangate-webpanel
    systemctl start irangate-webpanel
}
initialize_database() {
    cat > $WEBPANEL_DIR/database/init.sql << EOF
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Create admin user
INSERT INTO users (username, password, role)
VALUES ('$ADMIN_USER', '$ADMIN_PASS', 'admin');
EOF
    sqlite3 $WEBPANEL_DIR/database/irangate.db < $WEBPANEL_DIR/database/init.sql
    chmod 600 $WEBPANEL_DIR/database/irangate.db
}
configure_logging() {
    mkdir -p /var/log/irangate
    cat > /etc/logrotate.d/irangate-webpanel << EOF
/var/log/irangate/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 root root
}
EOF
}
main() {
    echo "Starting webpanel installation..."
    if [ -z "$WEBPANEL_PORT" ] || [ -z "$ADMIN_USER" ] || [ -z "$ADMIN_PASS" ]; then
        echo "Error: Missing required parameters"
        exit 1
    }
    install_webpanel
    configure_nginx
    create_service
    initialize_database
    configure_logging
    echo "Webpanel installation completed!"
    echo "Access the panel at: http://your-server-ip:$WEBPANEL_PORT"
}
main