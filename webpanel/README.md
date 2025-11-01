# IranGate Web Panel

A comprehensive web-based management interface for IranGate VPN server.

## Features

- **Dashboard**: Real-time server status and statistics
- **Client Management**: Add, remove, and manage VPN clients
- **Configuration**: Server settings and OpenVPN configuration
- **Monitoring**: Live connection monitoring and traffic statistics
- **User Management**: Role-based access control (RBAC)
- **Subscriptions**: Manage client subscriptions and expiration
- **Backup/Restore**: Configuration backup and restore
- **Real-time Updates**: WebSocket-based live updates

## Installation

### Method 1: Using IranGate CLI Menu

1. Run IranGate CLI:
   ```bash
   irangate
   ```

2. Select option **9: Install Web Panel**

3. Follow the prompts:
   - Enter admin username (default: admin)
   - Enter admin password (twice for confirmation)
   - Enter web path (default: example)
   - Enter port number (required, no default)

### Method 2: Direct Script Execution

```bash
sudo bash /root/ov/irangate/webpanel/install_webpanel.sh
```

## Configuration

The web panel uses environment variables for configuration, stored in `/etc/irangate/webpanel.env`:

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `WEBPANEL_PORT` | Port for web panel | 8080 | 6652 |
| `WEBPANEL_BASEPATH` | Base path for web panel | / | /example |
| `WEBPANEL_STATIC_DIR` | Static files directory | /opt/irangate/webpanel/frontend | /var/www/panel |
| `WEBPANEL_ADMIN_USER` | Admin username | admin | myadmin |
| `WEBPANEL_ADMIN_PASS` | Admin password | admin123 | mypassword123 |
| `WEBPANEL_ADMIN_EMAIL` | Admin email | admin@irangate.local | admin@example.com |

### Example Configuration

```bash
WEBPANEL_PORT=6652
WEBPANEL_BASEPATH=/example
WEBPANEL_STATIC_DIR=/opt/irangate/webpanel/frontend
WEBPANEL_ADMIN_USER=admin
WEBPANEL_ADMIN_PASS=securepassword123
WEBPANEL_ADMIN_EMAIL=admin@irangate.local
```

## Access

After installation, access the web panel at:

```
http://YOUR_SERVER_IP:PORT/BASEPATH/
```

### Examples

- Default installation: `http://192.168.1.100:8080/`
- Custom port and path: `http://192.168.1.100:6652/example/`

## API Endpoints

### Authentication

- `POST /api/login` - Login with username/password
- `GET /api/settings/admin` - Get admin settings
- `PUT /api/settings/admin/password` - Change password

### Client Management

- `GET /api/clients` - List all clients
- `POST /api/clients` - Create new client
- `DELETE /api/clients/{name}` - Remove client
- `GET /api/clients/{name}/status` - Get client status
- `GET /api/clients/{name}/export` - Export client config
- `GET /api/clients/orphaned` - List orphaned certificates
- `POST /api/clients/cleanup-orphaned` - Cleanup orphaned certificates

### Server Management

- `GET /api/server/status` - Get server status
- `POST /api/server/restart` - Restart server
- `GET /api/config` - Get server configuration
- `PUT /api/config` - Update server configuration

### Monitoring

- `GET /api/monitor/stats` - Get monitoring statistics
- `GET /api/settings/server` - Get server settings
- `PUT /api/settings/server` - Update server settings

### User Management (RBAC)

- `GET /api/users` - List all users
- `POST /api/users` - Create new user
- `PUT /api/users/{username}` - Update user
- `DELETE /api/users/{username}` - Delete user

### Subscriptions

- `GET /api/subscriptions` - List all subscriptions
- `POST /api/subscriptions` - Create new subscription
- `PUT /api/subscriptions/{client}` - Update subscription
- `DELETE /api/subscriptions/{client}` - Delete subscription
- `GET /api/subscriptions/expiring` - Get expiring subscriptions
- `POST /api/subscriptions/{client}/renew` - Renew subscription

### Backup

- `GET /api/backup/export` - Export backup
- `POST /api/backup/import` - Import backup

### WebSocket

- `GET /ws` - WebSocket connection for real-time updates

## Service Management

### Start/Stop Service

```bash
# Start web panel
sudo systemctl start irangate-webpanel

# Stop web panel
sudo systemctl stop irangate-webpanel

# Restart web panel
sudo systemctl restart irangate-webpanel

# Check status
sudo systemctl status irangate-webpanel

# Enable auto-start
sudo systemctl enable irangate-webpanel

# Disable auto-start
sudo systemctl disable irangate-webpanel
```

### View Logs

```bash
# View recent logs
journalctl -u irangate-webpanel -n 100

# Follow logs in real-time
journalctl -u irangate-webpanel -f

# View logs since specific time
journalctl -u irangate-webpanel --since "1 hour ago"
```

## Configuration Management

### Change Web Path

Using IranGate CLI:
1. Run `irangate`
2. Select **10: Web Panel Settings**
3. Select **3: Change Current WebPath**

Or manually edit `/etc/irangate/webpanel.env`:
```bash
sudo nano /etc/irangate/webpanel.env
# Change WEBPANEL_BASEPATH=/newpath
sudo systemctl restart irangate-webpanel
```

### Change Port

Edit configuration file:
```bash
sudo nano /etc/irangate/webpanel.env
# Change WEBPANEL_PORT=8080
sudo systemctl restart irangate-webpanel
```

### Change Admin Credentials

Using IranGate CLI:
1. Run `irangate`
2. Select **10: Web Panel Settings**
3. Select **4: Change Admin Credentials**

## Security

### Authentication

- Uses JWT tokens for API authentication
- Tokens expire after 24 hours by default
- Password hashing using SHA-256
- Session management with secure cookies

### CORS Configuration

- Currently allows all origins (`*`)
- Should be restricted in production environments
- Configure allowed origins in backend code

### HTTPS Support

Currently HTTP only. For production:
1. Configure reverse proxy (nginx/apache)
2. Use SSL certificates
3. Force HTTPS redirects

## Troubleshooting

### Common Issues

1. **White page/blank screen**
   - Check if frontend files are copied to `/opt/irangate/webpanel/frontend/`
   - Verify service is running: `systemctl status irangate-webpanel`
   - Check logs: `journalctl -u irangate-webpanel`

2. **Port already in use**
   - Check what's using the port: `netstat -tulpn | grep :PORT`
   - Kill process or use different port

3. **Service won't start**
   - Check logs: `journalctl -u irangate-webpanel -f`
   - Verify environment file: `cat /etc/irangate/webpanel.env`
   - Check binary exists: `ls -la /usr/local/bin/irangate-webpanel`

4. **API calls failing**
   - Check CORS headers
   - Verify JWT token is valid
   - Check API endpoint URLs

### Debug Mode

Enable verbose logging by editing the service file:
```bash
sudo systemctl edit irangate-webpanel
```

Add:
```ini
[Service]
Environment=DEBUG=true
Environment=LOG_LEVEL=debug
```

Then restart:
```bash
sudo systemctl daemon-reload
sudo systemctl restart irangate-webpanel
```

## File Structure

```
/opt/irangate/webpanel/
├── frontend/           # Static web files
│   ├── index.html     # Main dashboard
│   ├── clients.html   # Client management
│   ├── settings.html  # Settings page
│   ├── css/          # Stylesheets
│   └── js/           # JavaScript files
├── admin.json         # Admin user data
├── telegram.json      # Telegram bot config
├── server.json        # Server configuration
└── templates/         # Email templates

/etc/irangate/
└── webpanel.env       # Environment configuration

/usr/local/bin/
└── irangate-webpanel  # Web panel binary

/etc/systemd/system/
└── irangate-webpanel.service  # Systemd service file
```

## Development

### Building from Source

```bash
cd /root/ov/irangate/webpanel/backend
go build -o /usr/local/bin/irangate-webpanel
```

### Frontend Development

Frontend files are located in `/root/ov/irangate/webpanel/frontend/`. After making changes:
1. Copy files to service directory:
   ```bash
   sudo cp -r /root/ov/irangate/webpanel/frontend/* /opt/irangate/webpanel/frontend/
   ```
2. Restart service:
   ```bash
   sudo systemctl restart irangate-webpanel
   ```

## Support

For issues and support:
1. Check this documentation
2. Review troubleshooting section
3. Check logs with `journalctl -u irangate-webpanel`
4. Refer to main IranGate documentation in `/root/ov/irangate/docs/`

## Version

- **Current Version**: 1.0.0
- **Last Updated**: October 2025
- **Compatibility**: Ubuntu 20.04, 22.04, 24.04
