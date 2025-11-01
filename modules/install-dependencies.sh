#!/bin/bash
install_system_packages() {
    apt-get update
    apt-get install -y \
        curl \
        wget \
        git \
        unzip \
        tar \
        jq \
        netcat \
        net-tools \
        dnsutils \
        sudo \
        nano \
        htop \
        cron \
        logrotate
    apt-get install -y \
        build-essential \
        pkg-config \
        cmake \
        autoconf \
        automake \
        libtool
    apt-get install -y \
        python3 \
        python3-pip \
        python3-dev \
        python3-venv \
        python3-setuptools
    apt-get install -y \
        nginx \
        certbot \
        python3-certbot-nginx
    apt-get install -y \
        sqlite3
    apt-get install -y \
        vnstat \
        iftop \
        iotop \
        sysstat
}
install_golang() {
    add-apt-repository ppa:longsleep/golang-backports -y
    apt-get update
    apt-get install -y golang-go
    echo 'export GOPATH=$HOME/go' >> /etc/profile
    echo 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin' >> /etc/profile
    source /etc/profile
}
install_nodejs() {
    curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
    apt-get install -y nodejs
    npm install -g pm2 yarn
}
install_python_packages() {
    pip3 install --upgrade pip
    pip3 install -r requirements.txt
}
install_openvpn_deps() {
    apt-get install -y \
        openvpn \
        easy-rsa \
}
configure_system() {
    timedatectl set-timezone UTC
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
    cat > /etc/logrotate.d/irangate << EOF
/var/log/irangate/*.log {
    weekly
    rotate 4
    compress
    delaycompress
    missingok
    notifempty
    create 0640 root root
}
EOF
}
main() {
    echo "Starting dependencies installation..."
    install_system_packages
    install_golang
    install_nodejs
    install_openvpn_deps
    configure_system
    echo "Dependencies installation completed!"
}
main