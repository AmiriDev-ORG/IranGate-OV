# 🚀 IRANGATE Installation Guide
## Version 1.1.0 - Production Release

**Platform:** Linux Only (Ubuntu 20.04+, Debian 10+, RHEL 8+, CentOS 8+)  
**Requirements:** OpenVPN, EasyRSA, SystemD  
**Installation Time:** 5-10 minutes

---

## 📋 Prerequisites

### System Requirements
- **Operating System:** Linux with SystemD
- **CPU:** 2+ cores recommended
- **RAM:** 4GB+ recommended
- **Disk:** 10GB+ free space
- **Network:** Static IP recommended
- **Access:** Root or sudo privileges

### Software Dependencies
- **OpenVPN** - VPN server software
- **EasyRSA** - PKI certificate management
- **SystemD** - Service management
- **Python 3** - Optional (for AI features)

---

## ⚡ Quick Installation (5 Minutes)

### Step 1: Install Dependencies

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install -y openvpn easy-rsa
```

**RHEL/CentOS:**
```bash
sudo yum install -y epel-release
sudo yum install -y openvpn easy-rsa
```

### Step 2: Install IRANGATE Binary

```bash
# Copy binary to server (from your local machine)
scp irangate user@your-server:/tmp/

# SSH to server
ssh user@your-server

# Install binary
sudo mv /tmp/irangate /usr/local/bin/irangate
sudo chmod +x /usr/local/bin/irangate
```

### Step 3: Verify Installation

```bash
# Check version
irangate version
# Expected: IRANGATE CLI v1.0.0

# Check dependencies
irangate check-deps
# Expected:
# ✅ OpenVPN: Available
# ✅ EasyRSA: Available
# ✅ SystemD: Available
# ⚠️  Python3: (optional)
```

### Step 4: Initialize System

```bash
# Install and configure OpenVPN server
sudo irangate install

# This will:
# - Initialize PKI infrastructure
# - Generate server certificates
# - Create server configuration
# - Set up directory structure
# - Prepare database
```

### Step 5: Start Service

```bash
# Start OpenVPN service
sudo irangate start

# Check status
sudo irangate status

# Expected: Service running
```

### Step 6: Create First Client

```bash
# Add a new VPN client
sudo irangate client add john

# Export configuration
sudo irangate client export john

# Config file location: /etc/openvpn/client/john.ovpn
# Send this file to your user
```

**Done! Your VPN server is now running!** 🎉

---

## 📂 Directory Structure Created

After installation, the following structure is created:

```
/opt/irangate/
└── database/
    ├── clients/              # Individual client JSON files
    │   └── <clientname>.json # One file per client
    ├── settings.json         # Server configuration
    ├── cron.json            # Scheduled tasks
    ├── backups/             # Automatic backups
    └── history/             # Historical data

/etc/openvpn/
├── server/
│   └── server.conf          # OpenVPN server config
├── client/
│   └── <clientname>.ovpn   # Client config files
└── easy-rsa/
    └── pki/                 # PKI certificates
        ├── ca.crt
        ├── issued/          # Client certificates
        └── private/         # Private keys

/var/log/
├── openvpn/
│   └── openvpn-status.log  # OpenVPN status
└── irangate/
    └── irangate.log         # IRANGATE logs
```

---

## 🔧 Building from Source (Optional)

If you want to build from source instead of using the binary:

```bash
# 1. Install Go 1.19+
sudo apt install -y golang-go

# 2. Clone repository
git clone https://github.com/amiridev-org/irangate-ov
cd irangate/irangate

# 3. Build
go build -ldflags="-s -w" -o irangate .

# 4. Install
sudo cp irangate /usr/local/bin/irangate
sudo chmod +x /usr/local/bin/irangate

# 5. Verify
irangate version
```

---

## ⚙️ Configuration

### Check Current Configuration

```bash
# View all settings
irangate config show

# JSON output for automation
irangate config show --json
```

### Modify Configuration

```bash
# Change server port
irangate config set server_port 1195

# Update DNS servers
irangate config set dns "1.1.1.1,8.8.8.8"

# Set MTU
irangate config set mtu 1420

# Enable IPv6
irangate config set ipv6_enabled true
```

---

## 🔥 Firewall Configuration

All firewalls are disabled for maximum compatibility:

```bash
# All firewalls are automatically disabled during installation
# No firewall configuration needed
```

---

## 🔄 Upgrading from v1.0.0

### Automatic Migration

If you're upgrading from v1.0.0, your data will be automatically migrated:

```bash
# 1. Backup current system
sudo cp -r /opt/irangate/database /opt/irangate/database_backup_v1.0.0

# 2. Stop service
sudo irangate stop

# 3. Replace binary
sudo cp irangate_v1.1.0 /usr/local/bin/irangate

# 4. Start service
sudo irangate start

# 5. Verify migration
ls -la /opt/irangate/database/clients/
# Should show individual .json files

# 6. Check old file backed up
ls -la /opt/irangate/database/users.json.bak
# Original users.json preserved
```

**Migration is automatic and safe!** Your old `users.json` is automatically converted to individual client files.

---

## 🧪 Post-Installation Testing

### Test Core Functionality

```bash
# 1. Check service status
irangate status

# 2. Add test client
irangate client add testuser

# 3. List clients
irangate client list

# 4. Export config
irangate client export testuser

# 5. Verify file created
ls -la /etc/openvpn/client/testuser.ovpn
ls -la /opt/irangate/database/clients/testuser.json

# 6. Clean up
irangate client remove testuser
```

### Test Recovery System

```bash
# Test recovery procedures
irangate test-recovery

# Expected output:
# 🔧 Testing IRANGATE recovery procedures...
# 📊 Testing database recovery...
# ✅ Database recovery test passed
# 🔒 Testing VPN service recovery...
# ✅ VPN recovery test passed
```

---

## 🆘 Troubleshooting

### Issue: "EasyRSA dependency not satisfied"

**Solution:**
```bash
sudo apt install -y easy-rsa
sudo mkdir -p /etc/openvpn/easy-rsa
irangate check-deps
```

### Issue: "systemctl command not found"

**Solution:**
```bash
# Your system must have SystemD
systemctl --version

# If not available, use init.d or service command
```

### Issue: Client creation fails

**Solution:**
```bash
# Check dependencies
irangate check-deps

# Verify OpenVPN is installed
openvpn --version

# Check logs
tail -f /var/log/irangate/irangate.log
```

### Issue: Service won't start

**Solution:**
```bash
# Check service status
irangate status

# Check OpenVPN logs
sudo tail -f /var/log/openvpn/openvpn.log

# Restart service
sudo irangate restart
```

---

## 📊 Verification Checklist

After installation, verify:

- [ ] `irangate version` shows v1.0.0
- [ ] `irangate check-deps` shows all dependencies available
- [ ] `irangate status` shows service running
- [ ] Can create client: `irangate client add test`
- [ ] Client file created: `/opt/irangate/database/clients/test.json`
- [ ] Config file created: `/etc/openvpn/client/test.ovpn`
- [ ] Can list clients: `irangate client list`
- [ ] Can remove client: `irangate client remove test`
- [ ] Service restarts: `irangate restart`

---

## 🎯 Next Steps

After successful installation:

1. **Create your first real client**
   ```bash
   irangate client add your-first-user
   irangate client export your-first-user
   ```

2. **Firewalls disabled**
   ```bash
   # All firewalls are automatically disabled during installation
   ```

3. **Set up automatic backups** (optional)
   ```bash
   irangate cron add backup_daily "0 2 * * *" backup
   ```

4. **Enable monitoring** (optional)
   ```bash
   irangate monitor live
   ```

5. **Read the full CLI guide**
   ```bash
   cat /opt/irangate/docs/cli-guide.md
   ```

---

## 📖 Additional Resources

- **CLI Guide:** `docs/cli-guide.md` - Complete command reference
- **Configuration:** `docs/configuration.md` - Advanced settings
- **Troubleshooting:** `docs/troubleshooting.md` - Common issues
- **AI Features:** `docs/ai-management.md` - AI-powered monitoring

---

*Installation Guide for IRANGATE v1.1.0*  
*Production-Ready Release*  
*October 12, 2025*
