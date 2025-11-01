# Traffic Analysis System - Quick Start Guide

## 5-Minute Setup

### Step 1: Configure OpenVPN Status Log

Edit OpenVPN server configuration:

```bash
sudo nano /etc/openvpn/server.conf
```

Add these lines:

```
# Traffic monitoring
status /etc/openvpn/openvpn-status.log 10
status-version 3
```

Restart OpenVPN:

```bash
sudo systemctl restart openvpn@server
```

### Step 2: Verify Status Log

```bash
# Check if file is being created
ls -la /etc/openvpn/openvpn-status.log

# View current status
cat /etc/openvpn/openvpn-status.log
```

### Step 3: Start Traffic Collection

```bash
# Start the collector
sudo irangate traffic start

# Verify it's running
ps aux | grep irangate
```

### Step 4: View Statistics

```bash
# Get current statistics
irangate traffic stats

# Get specific client stats
irangate traffic stats clientname

# Get real-time monitoring
irangate traffic live
```

### Step 5: View History

```bash
# Last 7 days
irangate traffic history clientname --days 7

# Top 10 users
irangate traffic top --limit 10 --days 7
```

## Common Usage Examples

### Monitor All Clients

```bash
# Real-time monitoring
irangate traffic live

# Last 24 hours summary
irangate traffic stats
```

### Analyze Specific Client

```bash
# Current usage
irangate traffic stats clientname

# Last 30 days
irangate traffic history clientname --days 30

# Bandwidth efficiency
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/traffic/efficiency/clientname?days=7
```

### Export Reports

```bash
# Export to JSON
irangate traffic export --format json --output report.json --days 30

# Export to CSV
irangate traffic export --format csv --output report.csv --days 7

# Via API
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/traffic/export?format=json&days=30
```

### Find Heavy Users

```bash
# Top 20 users this week
irangate traffic top --limit 20 --days 7

# Via API
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/traffic/top-users?limit=20&days=7
```

### Check for Issues

```bash
# Detect anomalies
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/traffic/anomalies/clientname?days=7

# Check quota usage
curl -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/traffic/quota/clientname
```

## Configuration Examples

### Minimal Configuration

```json
{
  "data_dir": "/var/lib/irangate/traffic",
  "collection_interval_seconds": 10,
  "retention_days": 30
}
```

### High-Performance Configuration

```json
{
  "data_dir": "/var/lib/irangate/traffic",
  "collection_interval_seconds": 30,
  "retention_days": 90,
  "batch_size": 200,
  "max_concurrent_writes": 20,
  "cache_enabled": true,
  "cache_ttl_seconds": 600,
  "enable_anomaly_detection": true
}
```

### Low-Resource Configuration

```json
{
  "data_dir": "/var/lib/irangate/traffic",
  "collection_interval_seconds": 60,
  "retention_days": 7,
  "batch_size": 50,
  "max_concurrent_writes": 5,
  "cache_enabled": false,
  "enable_anomaly_detection": false
}
```

## Troubleshooting Quick Fixes

### Collector Not Starting

```bash
# Check OpenVPN is running
sudo systemctl status openvpn@server

# Check status log exists
ls -la /etc/openvpn/openvpn-status.log

# Restart collector
sudo irangate traffic stop
sudo irangate traffic start
```

### No Data Being Collected

```bash
# Check OpenVPN configuration
grep status /etc/openvpn/server.conf

# Check file permissions
sudo chmod 644 /etc/openvpn/openvpn-status.log

# Manually test parsing
cat /etc/openvpn/openvpn-status.log | grep "CLIENT LIST"
```

### High Disk Usage

```bash
# Check current usage
du -sh /var/lib/irangate/traffic/

# Reduce retention
echo '{"retention_days": 30}' | sudo tee /etc/irangate/traffic_config.json

# Clean old data manually
find /var/lib/irangate/traffic/records -mtime +90 -delete
```

### Slow Queries

```bash
# Enable caching
echo '{"cache_enabled": true}' | sudo tee /etc/irangate/traffic_config.json

# Increase collection interval
echo '{"collection_interval_seconds": 30}' | sudo tee /etc/irangate/traffic_config.json
```

## API Quick Reference

### Get Statistics

```bash
# All clients
GET /api/traffic/stats

# Specific client
GET /api/traffic/stats/{clientname}?days=7

# Real-time
GET /api/traffic/realtime
```

### Historical Data

```bash
# Client history
GET /api/traffic/history?client={name}&days=7

# Top users
GET /api/traffic/top-users?limit=10&days=7

# Export
GET /api/traffic/export?format=json&days=30
```

### Analytics

```bash
# Efficiency
GET /api/traffic/efficiency/{clientname}?days=7

# Anomalies
GET /api/traffic/anomalies/{clientname}?days=7

# Quota
GET /api/traffic/quota/{clientname}
```

## Next Steps

1. **Set Up Alerts**: Configure quota and anomaly alerts
2. **Integration**: Connect with Telegram/Email notifications
3. **Dashboard**: Build custom visualizations
4. **Reports**: Schedule automated reports
5. **Optimization**: Tune for your specific workload

## Need More Help?

- Full Documentation: [TRAFFIC_ANALYSIS_SYSTEM.md](./TRAFFIC_ANALYSIS_SYSTEM.md)
- CLI Reference: `irangate traffic --help`
- API Docs: [API.md](../webpanel/API.md)

## Examples Directory

```bash
# View examples
ls -la /root/OV-Panel/irangate/examples/traffic/

# Sample reports
cat examples/traffic/sample_report.json

# API examples
cat examples/traffic/api_examples.sh
```

