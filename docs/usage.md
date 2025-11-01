# 📖 IRANGATE Usage Guide
## Version 1.1.0 - Complete Usage Examples

**Updated:** October 12, 2025

---

## 🚀 Quick Start

### First Time Setup

```bash
# 1. Check dependencies
irangate check-deps

# 2. Install OpenVPN server
sudo irangate install

# 3. Start service
sudo irangate start

# 4. Create your first client
sudo irangate client add john

# 5. Export configuration
sudo irangate client export john

# 6. Send /etc/openvpn/client/john.ovpn to your user
```

---

## 👥 Client Management

### Creating Clients

```bash
# Add new client (basic)
irangate client add alice

# Add multiple clients
irangate client add bob
irangate client add charlie
irangate client add david

# Verify clients created
irangate client list
```

**What happens when you create a client:**
1. ✅ PKI certificates generated
2. ✅ `.ovpn` config file created at `/etc/openvpn/client/<name>.ovpn`
3. ✅ Client JSON saved at `/opt/irangate/database/clients/<name>.json`
4. ✅ Default expiry set to 1 month

### Viewing Clients

```bash
# List all clients (table format)
irangate client list

# List all clients (JSON format)
irangate client list --json

# Show specific client
irangate client show john

# Find client file
cat /opt/irangate/database/clients/john.json
```

### Exporting Client Configurations

```bash
# Export client config
irangate client export john

# Config location: /etc/openvpn/client/john.ovpn

# Copy to local machine
scp user@server:/etc/openvpn/client/john.ovpn ./john.ovpn

# Send to user via email/secure channel
```

### Removing Clients

```bash
# Remove client
irangate client remove john

# This will:
# - Delete /opt/irangate/database/clients/john.json
# - Revoke certificates
# - Remove .ovpn file
```

### Managing Client Status

```bash
# List all active clients
irangate client list --json | jq '.[] | select(.active == true)'

# List expired clients
irangate client list expired

# Check specific client status
irangate client show john | grep status
```

---

## 🔧 Service Management

### Start/Stop/Restart

```bash
# Start OpenVPN service
sudo irangate start

# Stop OpenVPN service
sudo irangate stop

# Restart OpenVPN service
sudo irangate restart

# Check service status
irangate status
```

### Service Status Information

```bash
# Basic status
irangate status

# Detailed JSON status
irangate status --json
```

**Status Output:**
- Service state (running/stopped)
- Uptime
- Connected clients count
- Server IP and port
- Configuration file location

---

## ⚙️ Configuration Management

### View Configuration

```bash
# Show all settings
irangate config show

# Show as JSON
irangate config show --json

# Export to file
irangate config show --json > current_config.json
```

### Modify Configuration

```bash
# Change server port
irangate config set server_port 1195

# Update DNS servers
irangate config set dns "1.1.1.1,8.8.8.8,208.67.222.222"

# Set MTU
irangate config set mtu 1420

# Change cipher
irangate config set cipher AES-256-CBC

# Enable IPv6
irangate config set ipv6_enabled true

# After changes, restart service
sudo irangate restart
```

### Backup & Restore

```bash
# Create backup
irangate config backup

# List backups
ls -la /opt/irangate/database/backups/

# Restore from backup
irangate config restore /opt/irangate/database/backups/backup_2025-10-12.json
```

---

## 📊 Monitoring & Statistics

### Real-Time Monitoring

```bash
# Live monitoring dashboard
irangate monitor live

# Show connected clients
irangate monitor clients

# View traffic statistics
irangate monitor traffic
```

### Export Statistics

```bash
# Export current stats
irangate monitor export

# Export as JSON
irangate monitor export --json > stats.json

# Parse with jq
irangate monitor export --json | jq '.connected_clients'
```

---

## 📅 Scheduled Tasks (Cron Jobs)

### Add Scheduled Tasks

```bash
# Daily backup at 2 AM
irangate cron add backup_daily "0 2 * * *" backup

# Hourly monitoring
irangate cron add monitor_hourly "0 * * * *" monitor

# Weekly cleanup (Sunday midnight)
irangate cron add cleanup_weekly "0 0 * * 0" cleanup
```

### Manage Cron Jobs

```bash
# List all jobs
irangate cron list

# Run job manually
irangate cron run backup_daily

# Remove job
irangate cron remove backup_daily
```

### Cron Schedule Examples

```
0 2 * * *      # Daily at 2 AM
*/30 * * * *   # Every 30 minutes
0 */6 * * *    # Every 6 hours
0 0 * * 0      # Weekly on Sunday
0 0 1 * *      # Monthly on 1st
```

---

## 🤖 AI Management (Optional)

### Setup AI Agent

```bash
# Install AI dependencies
irangate ai install

# Configure OpenAI API key
irangate ai config set openai_api_key "sk-your-key-here"

# Enable auto-healing
irangate ai config set enable_auto_healing true

# Start AI monitoring
irangate ai start
```

### AI Commands

```bash
# Check AI status
irangate ai status

# Start AI agent
irangate ai start

# Stop AI agent
irangate ai stop

# Restart AI agent
irangate ai restart

# View AI configuration
irangate ai config show

# Test AI functionality
irangate ai test
```

---

## 🔍 Troubleshooting & Recovery

### Check System Health

```bash
# Verify dependencies
irangate check-deps

# Test recovery procedures
irangate test-recovery

# Check service status
irangate status
```

### Common Operations

```bash
# View logs
sudo tail -f /var/log/irangate/irangate.log
sudo tail -f /var/log/openvpn/openvpn.log

# Restart if issues
sudo irangate restart

# Full reinstall (if needed)
sudo irangate uninstall
sudo irangate install
```

---

## 📝 Complete Workflow Example

### Setting Up a Production VPN Server

```bash
# 1. Initial setup
sudo irangate install
sudo irangate start

# 2. Configure server
irangate config set server_port 1194
irangate config set dns "1.1.1.1,8.8.8.8"
irangate config set cipher AES-256-GCM
sudo irangate restart

# 3. Create clients
irangate client add employee1
irangate client add employee2
irangate client add employee3

# 4. Export configs
irangate client export employee1
irangate client export employee2
irangate client export employee3

# 5. Send .ovpn files to users
# Files are in: /etc/openvpn/client/*.ovpn

# 6. Set up monitoring
irangate monitor live

# 7. Schedule backups
irangate cron add backup_daily "0 2 * * *" backup

# 8. Monitor clients
irangate client list
```

### Daily Operations

```bash
# Morning check
irangate status
irangate client list

# Add new client
irangate client add newuser
irangate client export newuser

# Check monitoring
irangate monitor clients

# Manual backup
irangate config backup

# View logs if issues
sudo tail -f /var/log/openvpn/openvpn.log
```

---

## 🎯 Advanced Usage

### Bulk Client Creation

```bash
# Create multiple clients
for user in user1 user2 user3 user4 user5; do
    irangate client add $user
    irangate client export $user
done

# Verify all created
irangate client list | grep -c "name"
```

### Export All Configurations

```bash
# Export all client configs
mkdir -p /tmp/client_configs
for client in $(irangate client list --json | jq -r '.[].name'); do
    cp /etc/openvpn/client/$client.ovpn /tmp/client_configs/
done

# Create archive
tar -czf client_configs.tar.gz /tmp/client_configs/
```

### Database Operations

```bash
# List all client files
ls -la /opt/irangate/database/clients/

# Count total clients
ls /opt/irangate/database/clients/*.json | wc -l

# Search for specific client
find /opt/irangate/database/clients/ -name "john.json"

# View client data
cat /opt/irangate/database/clients/john.json | jq .
```

---

## 📊 JSON Output for Automation

All commands support `--json` flag for automation:

```bash
# Get clients as JSON
irangate client list --json

# Parse with jq
irangate client list --json | jq '.[].name'

# Filter active clients
irangate client list --json | jq '.[] | select(.active == true)'

# Export to file
irangate config show --json > config_export.json
```

---

## 🔄 Migration from v1.0.0

### Automatic Migration

When you first run v1.1.0, it automatically migrates:

```bash
# Just run any command
irangate client list

# Old users.json is migrated to individual files
# Original saved as users.json.bak

# Verify migration
ls -la /opt/irangate/database/clients/        # Individual files
ls -la /opt/irangate/database/users.json.bak  # Backup of old file
```

---

## 🎯 Quick Command Reference

### Most Used Commands

```bash
# Version & Status
irangate version
irangate status
irangate check-deps

# Client Management
irangate client add <name>
irangate client list
irangate client export <name>
irangate client remove <name>

# Service Control
irangate start
irangate stop
irangate restart

# Configuration
irangate config show
irangate config set <key> <value>
irangate config backup

# Monitoring
irangate monitor live
irangate monitor clients
```

---

*Usage Guide for IRANGATE v1.1.0*  
*Production-Ready Release*  
*October 12, 2025*
