# 🤖 IRANGATE AI Management

## Overview

IRANGATE AI Management is an intelligent monitoring and management system that uses OpenAI's GPT models to automatically monitor, diagnose, and fix common issues with your VPN server.

## Features

### 🔍 **Intelligent Monitoring**
- **Real-time System Monitoring:** CPU, RAM, Disk, Network usage
- **OpenVPN Service Monitoring:** Service status, connection counts, failed connections
- **Smart Analysis:** AI-powered analysis of system metrics and logs

### 🛠️ **Auto-Healing**
- **Automatic Problem Detection:** Identifies issues before they become critical
- **Smart Fixes:** Automatically resolves common problems
- **Escalation:** Notifies admin when manual intervention is required

### 📱 **Telegram Integration**
- **Real-time Notifications:** Instant alerts for critical issues
- **Status Updates:** Regular system status reports
- **Interactive Commands:** Control the system via Telegram bot

## Installation

### 1. Install AI Dependencies

```bash
# Install Python packages
irangate ai install
```

### 2. Configure AI Settings

```bash
# Set OpenAI API key
irangate ai config set openai_api_key "your-openai-api-key"

# Set Telegram bot token (optional)
irangate ai config set telegram_bot_token "your-telegram-bot-token"

# Set admin chat ID (optional)
irangate ai config set admin_chat_id "your-chat-id"
```

### 3. Start AI Monitoring

```bash
# Start AI agent
irangate ai start
```

## Configuration

### Configuration File: `/etc/irangate/ai_config.json`

```json
{
  "monitoring_interval": 60,
  "cpu_threshold": 90,
  "ram_threshold": 95,
  "disk_threshold": 10,
  "max_failed_connections": 100,
  "auto_fix_enabled": true,
  "telegram_notifications": true,
  "openai_api_key": "your-api-key",
  "telegram_bot_token": "your-bot-token",
  "admin_chat_id": "your-chat-id",
  "log_file": "/var/log/irangate/ai_agent.log"
}
```

### Configuration Parameters

| Parameter | Description | Default | Range |
|-----------|-------------|---------|-------|
| `monitoring_interval` | Monitoring interval in seconds | 60 | 30-300 |
| `cpu_threshold` | CPU usage alert threshold (%) | 90 | 50-100 |
| `ram_threshold` | RAM usage alert threshold (%) | 95 | 50-100 |
| `disk_threshold` | Disk free space alert threshold (%) | 10 | 5-50 |
| `max_failed_connections` | Max failed connections before alert | 100 | 10-1000 |
| `auto_fix_enabled` | Enable automatic problem fixing | true | true/false |
| `telegram_notifications` | Enable Telegram notifications | true | true/false |

## CLI Commands

### Status Commands

```bash
# Show AI agent status
irangate ai status

# Show current configuration
irangate ai config show

# Test AI functionality
irangate ai test
```

### Control Commands

```bash
# Start AI monitoring
irangate ai start

# Stop AI monitoring
irangate ai stop

# Restart AI monitoring
irangate ai restart
```

### Configuration Commands

```bash
# Set monitoring interval
irangate ai config set monitoring_interval 120

# Set CPU threshold
irangate ai config set cpu_threshold 85

# Enable/disable auto-fix
irangate ai config set auto_fix_enabled true

# Set OpenAI API key
irangate ai config set openai_api_key "sk-..."

# Set Telegram bot token
irangate ai config set telegram_bot_token "1234567890:ABC..."
```

## AI Capabilities

### 🔧 **Auto-Fix Actions**

The AI can automatically perform these actions:

1. **Restart OpenVPN Service**
   - Trigger: Service crashes or high failed connections
   - Action: `systemctl restart openvpn`

2. **Clean Up Log Files**
   - Trigger: Low disk space
   - Action: Remove old log files (>7 days)

3. **Check Certificates**
   - Trigger: Certificate-related errors
   - Action: Validate certificate expiration dates

4. **System Restart** (with confirmation)
   - Trigger: Critical system issues
   - Action: Requires manual confirmation

### 🧠 **AI Analysis Process**

1. **Data Collection:** Gathers system metrics every monitoring interval
2. **Pattern Recognition:** AI analyzes patterns and trends
3. **Issue Detection:** Identifies potential problems
4. **Confidence Assessment:** AI provides confidence score (0-100%)
5. **Action Decision:** Takes action if confidence > 70%
6. **Escalation:** Notifies admin if manual intervention needed

## Telegram Bot Commands

### Bot Setup

1. Create a bot with [@BotFather](https://t.me/botfather)
2. Get your bot token
3. Get your chat ID using [@userinfobot](https://t.me/userinfobot)
4. Configure in IRANGATE:

```bash
irangate ai config set telegram_bot_token "YOUR_BOT_TOKEN"
irangate ai config set admin_chat_id "YOUR_CHAT_ID"
```

### Bot Commands

- `/status` - Get current system status
- `/logs` - Get recent log entries
- `/fix` - Trigger manual fix attempt
- `/config` - Show current AI configuration

## Monitoring & Logs

### Log Files

- **AI Agent Log:** `/var/log/irangate/ai_agent.log`
- **System Logs:** `/var/log/syslog`
- **OpenVPN Logs:** `/var/log/openvpn/openvpn.log`

### Log Levels

- **INFO:** Normal operations, status updates
- **WARNING:** Issues detected, auto-fix attempts
- **ERROR:** Failed operations, critical issues
- **CRITICAL:** System failures, manual intervention required

## Security Considerations

### API Key Security

- Store OpenAI API key securely
- Use environment variables for sensitive data
- Regularly rotate API keys
- Monitor API usage and costs

### Network Security

- AI agent runs locally on the server
- No external data transmission except to OpenAI/Telegram
- All communications use HTTPS/TLS
- Telegram bot requires authentication

## Troubleshooting

### Common Issues

1. **AI Agent Won't Start**
   ```bash
   # Check Python dependencies
   irangate ai test
   
   # Install missing packages
   irangate ai install
   ```

2. **OpenAI API Errors**
   ```bash
   # Verify API key
   irangate ai config show
   
   # Test API connectivity
   irangate ai test
   ```

3. **Telegram Notifications Not Working**
   ```bash
   # Check bot configuration
   irangate ai config show
   
   # Test bot manually
   # Send /start to your bot
   ```

### Debug Mode

```bash
# Enable verbose logging
tail -f /var/log/irangate/ai_agent.log
```

## Best Practices

### 1. **Gradual Rollout**
- Start with high confidence thresholds (80%+)
- Monitor AI decisions closely
- Adjust thresholds based on performance

### 2. **Backup Strategy**
- Always backup before major changes
- Test AI fixes in staging environment
- Keep manual override capabilities

### 3. **Cost Management**
- Monitor OpenAI API usage
- Set reasonable monitoring intervals
- Use auto-fix sparingly for critical systems

### 4. **Monitoring**
- Regularly check AI agent logs
- Monitor system performance
- Review AI decision accuracy

## Support

For issues and support:

1. Check logs: `/var/log/irangate/ai_agent.log`
2. Run diagnostics: `irangate ai test`
3. Review configuration: `irangate ai config show`
4. Contact support with log files

---

**🤖 AI Management makes your IRANGATE server truly intelligent and self-healing!**
