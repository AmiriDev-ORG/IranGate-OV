# IranGate-OV 🚀

**Professional OpenVPN Gateway Management System with AI Monitoring & Web Panel**

---

## Overview

**IranGate-OV** is a powerful VPN gateway management system built for OpenVPN servers. It offers full CLI control, a modern web panel, and intelligent AI-based monitoring for seamless server management.

* **Version:** 3.0.0
* **Platform:** Linux (Ubuntu 22.04+, Debian 11+, CentOS 9+, Fedora)
* **Maintainer:** AmiriDev_ORG
* **Repository:** [GitHub](https://github.com/AmiriDev-ORG/IranGate-OV)
* **Telegram:** [@IranGate_Official](https://t.me/IranGate_Official)
* **Telegram Group:** [@IranGate_Community](https://t.me/IranGate_Community)
---

## Key Features

* ✅ **CLI Management:** Start/stop OpenVPN, manage clients, monitor traffic
* ✅ **Web Panel:** Real-time dashboard, client management, subscriptions, RBAC
* ✅ **AI Monitoring:** Auto-healing, resource analysis, Telegram alerts
* ✅ **Traffic Analytics:** Live stats, quota tracking, anomaly detection
* ✅ **Backup & Recovery:** Automatic system backups & disaster recovery
* ✅ **Multi-Protocol Support:** UDP & TCP
* ✅ **Production Ready:** Logging, error handling, and system health checks

---

## Installation

```bash
git clone https://github.com/AmiriDev-ORG/IranGate-OV.git
cd IranGate-OV
chmod +x install.sh
sudo ./install.sh
```

**Installation script includes:**

* System dependencies (OpenVPN, Easy-RSA, Go, Python)
* PKI & OpenVPN configuration
* CLI & Web Panel setup
* Firewall (iptables) configuration
* systemd services & auto-start

---

## Project Structure

```
irangate/
├── main.go          # CLI entry point
├── pkg/             # Core Go packages (CLI, AI, client, traffic, backup)
├── webpanel/        # Frontend + Backend
├── modules/         # Installation scripts
├── scripts/         # Helper scripts
├── clients/         # Generated client .ovpn files
├── database/        # JSON/SQLite storage
├── logs/            # Logs
└── templates/       # OpenVPN templates
```

---

## Quick Workflow

### Adding a Client

1. Use CLI: `irangate client add` or Web Panel
2. Generates certificate via EasyRSA
3. Creates `.ovpn` configuration & optional QR code
4. Stores data in `/opt/irangate/clients/` and database

### Traffic Monitoring

* Parses `openvpn-status.log` in real-time
* Aggregates hourly → daily → monthly
* Sends alerts & updates Web Panel via WebSocket

### AI Auto-Healing ( Under Codes - Working for it !)

* Python agent monitors CPU, RAM, Disk, Network
* Analyzes via OpenAI API
* Performs auto-recovery & sends Telegram alerts

---

## Tech Stack

* **Backend:** Go 1.24, Cobra, Viper, Logrus
* **Frontend:** HTML5, CSS3, JS, WebSocket
* **Database:** JSON, SQLite, in-memory cache
* **Infrastructure:** OpenVPN, Easy-RSA, systemd, iptables
* **AI:** Python 3.7+, OpenAI GPT API, Telegram Bot API

---

## Contribution

1. Fork the repository
2. Create a new branch
3. Commit your changes
4. Open a Pull Request

---

## 💰 Donation: [DonateLink](https://donito.me/irangate)
---

## License

MIT License

Copyright (c) 2025 Amirreza Radsar - AmiriDev

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
