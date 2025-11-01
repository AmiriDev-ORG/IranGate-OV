# 🔧 IRANGATE Troubleshooting Guide
## Version 1.1.0

**Updated:** October 12, 2025  
**Platform:** Linux Only

---

## 🆘 Quick Diagnostics

### First Steps for Any Issue

```bash
# 1. Check version
irangate version

# 2. Verify dependencies
irangate check-deps

# 3. Check service status
irangate status

# 4. Test recovery system
irangate test-recovery

# 5. View logs
sudo tail -f /var/log/irangate/irangate.log
```

---

## 🚨 Common Issues & Solutions

### Issue 1: "EasyRSA dependency not satisfied"

**Symptom:**
```
Error: EasyRSA dependency not satisfied: easyrsa command not found
💡 Run 'irangate check-deps' for detailed information
```

**Solution:**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y easy-rsa

# RHEL/CentOS
sudo yum install -y easy-rsa

# Verify installation
irangate check-deps
```

---

### Issue 2: "systemctl command not found"

**Symptom:**
```
Error: systemctl command not found
```

**Cause:** SystemD not available on your system

**Solution:**
```bash
# Check if systemd is installed
systemctl --version

# If not installed, your system may not be compatible
# IRANGATE requires SystemD-based Linux systems
```

---

### Issue 3: Client Creation Fails

**Symptom:**
```
Failed to create client: failed to generate client certificates
```

**Solution:**
```bash
# 1. Verify OpenVPN is installed
irangate check-deps

# 2. Check EasyRSA directory exists
ls -la /etc/openvpn/easy-rsa/

# 3. Initialize PKI if needed
cd /etc/openvpn/easy-rsa
sudo easyrsa init-pki
sudo easyrsa build-ca nopass

# 4. Try creating client again
sudo irangate client add testuser
```

---

### Issue 4: "client already exists"

**Symptom:**
```
Error: client john already exists
```

**Solution:**
```bash
# Check if client file exists
ls -la /opt/irangate/database/clients/john.json

# If it exists and you want to recreate:
irangate client remove john
irangate client add john

# If file doesn't exist but error persists:
# May be certificate conflict
cd /etc/openvpn/easy-rsa
sudo easyrsa revoke john
irangate client add john
```

---

### Issue 5: Service Won't Start

**Symptom:**
```
Failed to start OpenVPN service
```

**Solution:**
```bash
# 1. Check OpenVPN configuration
sudo openvpn --config /etc/openvpn/server/server.conf --verb 3

# 2. Check logs
sudo tail -f /var/log/openvpn/openvpn.log

# 3. Verify certificates exist
ls -la /etc/openvpn/easy-rsa/pki/ca.crt
ls -la /etc/openvpn/easy-rsa/pki/issued/server.crt

# 4. Try recovery
irangate test-recovery

# 5. Restart service
sudo irangate restart
```

---

### Issue 6: Client File Not Found

**Symptom:**
```
Error: client john not found
```

**Solution:**
```bash
# 1. Check if client file exists
ls -la /opt/irangate/database/clients/john.json

# 2. List all clients
irangate client list

# 3. Check if old format exists
cat /opt/irangate/database/users.json.bak

# 4. If migration issue, manually create client
irangate client add john
```

---

### Issue 7: Concurrent Operation Errors

**Symptom:**
Multiple admins getting errors when creating clients simultaneously

**Solution:**
```
# This is FIXED in v1.1.0!
# If you're still seeing issues, verify version:
irangate version

# Should show v1.0.0 or higher
# Race conditions are fixed in v1.1.0
```

---

### Issue 8: Database Corruption

**Symptom:**
```
Error: failed to parse client file: invalid JSON
```

**Solution:**
```bash
# 1. Test recovery
irangate test-recovery

# 2. Restore from backup
irangate config restore /opt/irangate/database/backups/latest_backup.json

# 3. If specific client file corrupted
rm /opt/irangate/database/clients/corrupted_user.json
irangate client add corrupted_user

# 4. Check file permissions
sudo chmod 600 /opt/irangate/database/clients/*.json
sudo chown root:root /opt/irangate/database/clients/*.json
```

---

## 🔍 Diagnostic Commands

### System Information

```bash
# Check IRANGATE version
irangate version

# Verify all dependencies
irangate check-deps

# Test recovery procedures
irangate test-recovery

# Show service status
irangate status
```

### Database Diagnostics

```bash
# Count total clients
ls /opt/irangate/database/clients/*.json | wc -l

# Check database structure
tree /opt/irangate/database/

# Verify JSON integrity
for file in /opt/irangate/database/clients/*.json; do
    jq . "$file" > /dev/null 2>&1 || echo "Invalid: $file"
done

# Check file permissions
ls -la /opt/irangate/database/clients/
```

### Service Diagnostics

```bash
# Check OpenVPN service
sudo systemctl status openvpn
sudo systemctl status openvpn@server

# Check OpenVPN process
ps aux | grep openvpn

# Check listening ports
sudo netstat -tlnup | grep 1194

# Test OpenVPN config
sudo openvpn --config /etc/openvpn/server/server.conf --verb 3
```

---

## 📊 Log Files

### IRANGATE Logs

```bash
# Application logs
sudo tail -f /var/log/irangate/irangate.log

# With filtering
sudo grep ERROR /var/log/irangate/irangate.log

# Last 100 lines
sudo tail -n 100 /var/log/irangate/irangate.log
```

### OpenVPN Logs

```bash
# OpenVPN status
sudo cat /var/log/openvpn/openvpn-status.log

# OpenVPN server log
sudo tail -f /var/log/openvpn/openvpn.log

# Connection logs
sudo journalctl -u openvpn@server
```

---

## 🔄 Recovery Procedures

### Database Recovery

```bash
# Test database recovery
irangate test-recovery

# Manual recovery steps
cd /opt/irangate/database
ls -la backups/

# Restore from latest backup
irangate config restore backups/backup_latest.json
```

### Service Recovery

```bash
# If service is stuck
sudo irangate stop
sudo killall openvpn
sudo irangate start

# If still issues
sudo irangate uninstall
sudo irangate install
```

### Complete System Reset

```bash
# CAUTION: This deletes all clients!

# 1. Backup first
sudo cp -r /opt/irangate/database /tmp/irangate_backup

# 2. Uninstall
sudo irangate uninstall

# 3. Clean directories
sudo rm -rf /opt/irangate
sudo rm -rf /etc/openvpn

# 4. Reinstall
sudo irangate install

# 5. Restore from backup if needed
sudo cp -r /tmp/irangate_backup /opt/irangate/database
```

---

## 💡 Tips & Best Practices

### Performance Tips

```bash
# Use UDP for better performance
irangate config set protocol udp

# Optimize MTU
irangate config set mtu 1412

# Monitor performance
irangate monitor live
```

### Security Tips

```bash
# Regular backups
irangate cron add backup_daily "0 2 * * *" backup

# Check file permissions
ls -la /opt/irangate/database/clients/

# Review connected clients
irangate client list

# Revoke unused clients
irangate client remove olduser
```

### Maintenance Tips

```bash
# Weekly health check
irangate check-deps
irangate test-recovery
irangate status

# Monthly cleanup
# Remove expired clients
irangate client list expired

# Check disk usage
du -sh /opt/irangate/database/
```

---

## 📞 Getting Help

### Built-in Help

```bash
# General help
irangate --help

# Command-specific help
irangate client --help
irangate config --help
irangate monitor --help

# Subcommand help
irangate client add --help
```

### Diagnostic Output

When reporting issues, include:

```bash
# 1. Version
irangate version

# 2. Dependencies
irangate check-deps

# 3. Status
irangate status

# 4. Client list
irangate client list

# 5. Recent logs
sudo tail -n 50 /var/log/irangate/irangate.log
```

---

## ✅ Health Check Checklist

Run this checklist regularly:

```bash
# System health
[ ] irangate version         # Shows correct version
[ ] irangate check-deps      # All dependencies available
[ ] irangate status          # Service running
[ ] irangate test-recovery   # Recovery tests pass

# Database health
[ ] ls /opt/irangate/database/clients/*.json  # Files readable
[ ] irangate client list     # Returns JSON data
[ ] irangate config show     # Shows configuration

# Service health
[ ] sudo systemctl status openvpn  # Service active
[ ] sudo netstat -tlnup | grep 1194  # Port listening
[ ] irangate monitor clients # Shows connected users
```

---

*Troubleshooting Guide for IRANGATE v1.1.0*  
*October 12, 2025*
