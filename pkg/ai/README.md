# 🤖 IRANGATE AI Management Package

## Overview

This package provides AI-powered monitoring and management capabilities for the IRANGATE VPN system. It includes intelligent system monitoring, automatic problem detection, and self-healing capabilities using OpenAI's GPT models.

## Files

- **`ai_agent.py`** - Main Python AI agent script
- **`ai_management.go`** - Go wrapper for AI management
- **`commands.go`** - CLI commands for AI management
- **`config.json`** - Default AI configuration
- **`requirements.txt`** - Python dependencies
- **`README.md`** - This documentation

## Quick Start

### 1. Install Dependencies
```bash
irangate ai install
```

### 2. Configure AI
```bash
irangate ai config set openai_api_key "your-api-key"
```

### 3. Start AI Monitoring
```bash
irangate ai start
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `irangate ai status` | Show AI agent status |
| `irangate ai start` | Start AI monitoring |
| `irangate ai stop` | Stop AI monitoring |
| `irangate ai restart` | Restart AI monitoring |
| `irangate ai config show` | Show configuration |
| `irangate ai config set <key> <value>` | Set configuration |
| `irangate ai test` | Test AI functionality |
| `irangate ai install` | Install dependencies |

## Features

- 🔍 **Real-time Monitoring:** CPU, RAM, Disk, Network
- 🛠️ **Auto-Healing:** Automatic problem resolution
- 📱 **Telegram Integration:** Smart notifications
- 🧠 **AI Analysis:** GPT-powered decision making
- ⚙️ **Configurable:** Flexible thresholds and settings

## Requirements

- Python 3.7+
- OpenAI API key
- Telegram bot token (optional)
- System access (sudo privileges for service management)

## Documentation

See `../docs/ai-management.md` for complete documentation.
