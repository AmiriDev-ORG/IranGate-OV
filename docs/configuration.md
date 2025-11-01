# ⚙️ IRANGATE Configuration Guide
## Version 1.1.0

**Purpose:** Complete guide to configuring IRANGATE OpenVPN Management System  
**Updated:** October 12, 2025

---

## 📂 Configuration Storage

### Database Configuration
IRANGATE uses JSON files for all configuration:

```
/opt/irangate/database/
├── clients/              # Individual client files (v1.1.0)
│   ├── user1.json       # Each client's configuration
│   ├── user2.json
│   └── ...
├── settings.json         # Server configuration
└── cron.json            # Scheduled tasks
```

### OpenVPN Configuration
```
/etc/openvpn/server/
└── server.conf          # OpenVPN server configuration
```

---

## 🔧 Server Configuration

### View Current Configuration

```bash
# Display all settings
irangate config show

# JSON output
irangate config show --json
```

**Example Output:**
```json
{
  "server_ip": "0.0.0.0",
  "server_port": 1194,
  "protocol": "udp",
  "cipher": "AES-256-GCM",
  "mtu": 1412,
  "dns": ["1.1.1.1", "8.8.8.8"],
  "ipv6_enabled": false,
  "auto_config_mode": "dynamic",
  "log_level": 3,
  "version": "1.0.0"
}
```

### Update Configuration

#### Set Specific Values

```bash
# Change server port
irangate config set server_port 1195

# Update protocol
irangate config set protocol tcp

# Set cipher
irangate config set cipher AES-256-CBC

# Configure MTU
irangate config set mtu 1400

# Set DNS servers
irangate config set dns "1.1.1.1,8.8.8.8"

# Enable IPv6
irangate config set ipv6_enabled true
```

#### Bulk Configuration Update

```bash
# Create JSON file with settings
cat > new_config.json <<EOF
{
  "server_port": 1195,
  "protocol": "tcp",
  "cipher": "AES-256-GCM",
  "mtu": 1420,
  "dns": ["1.1.1.1", "1.0.0.1"]
}
EOF

# Apply configuration
irangate config update new_config.json
```

---

## 📝 Configuration Parameters

### Network Settings

| Parameter | Default | Description | Valid Values |
|-----------|---------|-------------|--------------|
| `server_ip` | `0.0.0.0` | Server bind address | IP address or 0.0.0.0 |
| `server_port` | `1194` | OpenVPN port | 1-65535 |
| `protocol` | `udp` | Network protocol | `udp`, `tcp` |

### Security Settings

| Parameter | Default | Description | Valid Values |
|-----------|---------|-------------|--------------|
| `cipher` | `AES-256-GCM` | Encryption cipher | AES-256-GCM, AES-256-CBC |
| `auth` | `SHA256` | HMAC authentication | SHA256, SHA512 |

### Network Optimization

| Parameter | Default | Description | Valid Values |
|-----------|---------|-------------|--------------|
| `mtu` | `1412` | Maximum transmission unit | 1200-1500 |
| `dns` | `["1.1.1.1", "8.8.8.8"]` | DNS servers | Array of IP addresses |
| `ipv6_enabled` | `false` | Enable IPv6 | true, false |

### System Settings

| Parameter | Default | Description | Valid Values |
|-----------|---------|-------------|--------------|
| `log_level` | `3` | Logging verbosity | 0-5 (0=off, 5=debug) |
| `auto_config_mode` | `dynamic` | Config generation | dynamic, static |
| `backup_path` | `/opt/irangate/database/backups` | Backup location | File path |

---

## 👥 Client Configuration

### Individual Client Files (v1.1.0)

Each client has their own JSON file at:
```
/opt/irangate/database/clients/<clientname>.json
```

**Example: `/opt/irangate/database/clients/john.json`**
```json
{
  "name": "john",
  "created_at": "2025-10-12T15:30:00Z",
  "expires_at": "2025-11-12T15:30:00Z",
  "active": true,
  "ip": "10.8.0.2",
  "cipher": "AES-256-GCM",
  "data_used_mb": 156,
  "last_connection": "2025-10-12T18:00:00Z",
  "upload": 1048576,
  "download": 2097152,
  "bytes_received": 2097152,
  "bytes_sent": 1048576,
  "status": "connected",
  "group": "premium"
}
```

### Client Configuration Commands

```bash
# View all clients
irangate client list

# View specific client
irangate client show john

# Update client settings
irangate client update john --group premium

# Check client file directly
cat /opt/irangate/database/clients/john.json
```

---

## 💾 Backup & Restore

### Create Backup

```bash
# Manual backup
irangate config backup

# Backup location
ls -la /opt/irangate/database/backups/
```

### Restore Configuration

```bash
# List available backups
ls -la /opt/irangate/database/backups/

# Restore from backup
irangate config restore /opt/irangate/database/backups/backup_2025-10-12.json
```

### Automated Backups

```bash
# Schedule daily backup at 2 AM
irangate cron add backup_daily "0 2 * * *" backup

# Verify scheduled job
irangate cron list
```

---

## 📅 Scheduled Tasks (Cron Jobs)

### Configuration File

Location: `/opt/irangate/database/cron.json`

**Example:**
```json
{
  "jobs": [
    {
      "id": "backup_daily",
      "schedule": "0 2 * * *",
      "action": "backup",
      "params": {
        "type": "full",
        "retention_days": 7
      },
      "enabled": true
    }
  ]
}
```

### Manage Cron Jobs

```bash
# Add backup job
irangate cron add backup_daily "0 2 * * *" backup

# Add monitoring job
irangate cron add monitor_hourly "0 * * * *" monitor

# List all jobs
irangate cron list

# Remove job
irangate cron remove backup_daily

# Run job manually
irangate cron run backup_daily
```

---

## 🤖 AI Configuration (Optional)

### AI Config File

Location: `irangate/pkg/ai/config.json`

**Example:**
```json
{
  "openai_api_key": "sk-your-api-key-here",
  "model": "gpt-4",
  "check_interval": 300,
  "enable_auto_healing": true,
  "telegram_bot_token": "your-bot-token",
  "telegram_notifications": true,
  "admin_chat_id": "your-chat-id",
  "thresholds": {
    "cpu_percent": 80,
    "ram_percent": 85,
    "disk_percent": 90
  }
}
```

### Configure AI

```bash
# Set OpenAI API key
irangate ai config set openai_api_key "sk-your-key"

# Enable auto-healing
irangate ai config set enable_auto_healing true

# Set Telegram notifications
irangate ai config set telegram_notifications true
irangate ai config set telegram_bot_token "your-token"

# View AI config
irangate ai config show
```

---

## 📊 Monitoring Configuration

### Enable Monitoring

```bash
# Start real-time monitoring
irangate monitor live

# Export statistics
irangate monitor export --json > stats.json
```

### Log Configuration

```bash
# Set log level (0=off, 5=debug)
irangate config set log_level 3

# View logs
tail -f /var/log/irangate/irangate.log
tail -f /var/log/openvpn/openvpn.log
```

---

## 🔒 Security Configuration

### File Permissions

Automatic security permissions:
- Client JSON files: `0600` (owner only)
- Settings files: `0600`
- Certificates: `0600`
- Directories: `0755`

### Certificate Management

```bash
# Certificates location
/etc/openvpn/easy-rsa/pki/

# Regenerate certificates (if needed)
cd /etc/openvpn/easy-rsa
sudo easyrsa build-client-full newclient nopass
```

---

## 🔄 Migration from v1.0.0

### What Changes Automatically

✅ **Automatic Migration:**
- Old `users.json` → Individual `clients/<name>.json` files
- Backup created as `users.json.bak`
- All client data preserved
- No manual intervention required

### Verify Migration

```bash
# Check new structure
ls -la /opt/irangate/database/clients/

# Count clients in new format
ls /opt/irangate/database/clients/*.json | wc -l

# Verify old file backed up
ls -la /opt/irangate/database/users.json.bak
```

---

## 📋 Configuration Best Practices

### Production Settings

```bash
# Secure cipher
irangate config set cipher AES-256-GCM

# Reliable DNS
irangate config set dns "1.1.1.1,8.8.8.8"

# Optimal MTU
irangate config set mtu 1412

# Enable logging
irangate config set log_level 3

# Regular backups
irangate cron add backup_daily "0 2 * * *" backup
```

### Performance Tuning

```bash
# For high-load servers
irangate config set mtu 1500
irangate config set protocol udp

# For stability over speed
irangate config set protocol tcp
irangate config set mtu 1400
```

---

## 🎯 Quick Reference

### Essential Configuration Commands

```bash
# View config
irangate config show

# Set value
irangate config set <key> <value>

# Backup config
irangate config backup

# Restore config
irangate config restore <backup-file>

# Check dependencies
irangate check-deps

# Test recovery
irangate test-recovery
```

### Configuration Files

| File | Purpose | Location |
|------|---------|----------|
| Settings | Server config | `/opt/irangate/database/settings.json` |
| Clients | Client data | `/opt/irangate/database/clients/*.json` |
| Cron | Scheduled tasks | `/opt/irangate/database/cron.json` |
| Server | OpenVPN config | `/etc/openvpn/server/server.conf` |

---

*Configuration Guide for IRANGATE v1.1.0*  
*October 12, 2025*
