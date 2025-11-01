# IranGate OpenVPN Configuration Backend

A complete Flask-based backend system for managing OpenVPN server and client configurations with JSON-based storage and secure file operations.

## 🚀 Features

- **Complete Backend API** - RESTful API for OpenVPN configuration management
- **JSON-Based Storage** - Structured data storage for easy management
- **Secure File Operations** - Path traversal protection, atomic writes, file locking
- **Dual Configuration Storage** - Server configs saved to both required directories
- **Client Template System** - Automatic client config generation from server settings
- **Real-time Validation** - Comprehensive input validation and error handling
- **Theme Integration** - Full support for IranGate's 5 themes
- **Production Ready** - Systemd service, logging, error handling

## 📁 Project Structure

```
webpanel/backend/
├── app.py                    # Main Flask application
├── requirements.txt          # Python dependencies
├── install.sh               # Installation script
├── test_api.sh              # API testing script
├── API_DOCUMENTATION.md     # Complete API documentation
└── README.md               # This file

webpanel/frontend/
├── ovpn-wizard.html        # Configuration wizard HTML
├── css/wizard.css          # Wizard styling with theme support
└── js/wizard.js           # Frontend JavaScript with API integration
```

## 🛠️ Installation

### Prerequisites
- Ubuntu/Debian Linux
- Python 3.7+
- Root access

### Quick Installation

1. **Run the installation script:**
```bash
sudo bash /root/OV-Panel/irangate/webpanel/backend/install.sh
```

2. **Verify installation:**
```bash
sudo systemctl status irangate-backend.service
```

3. **Test the API:**
```bash
bash /root/OV-Panel/irangate/webpanel/backend/test_api.sh
```

## 🔧 Configuration

### File Locations
- **JSON Database:** `/var/lib/irangate/`
- **Server Configs:** `/etc/openvpn/` and `/etc/openvpn/server/`
- **Client Templates:** `/etc/openvpn/client-templates/`

### Service Management
```bash
# Start service
sudo systemctl start irangate-backend.service

# Stop service
sudo systemctl stop irangate-backend.service

# Restart service
sudo systemctl restart irangate-backend.service

# View logs
sudo journalctl -u irangate-backend.service -f
```

## 📊 JSON Database Structure

### server_configs.json
Stores all server configuration parameters in structured format:
```json
{
  "configs": [
    {
      "id": "uuid-here",
      "name": "myserver",
      "created_at": "2025-01-25T10:30:00Z",
      "file_path_1": "/etc/openvpn/myserver.conf",
      "file_path_2": "/etc/openvpn/server/myserver.conf",
      "parameters": {
        "dev": {"type": "tun", "number": 0},
        "proto": "udp",
        "port": 1194,
        "cipher": "AES-256-GCM",
        "auth": "SHA512",
        "persist_key": true,
        "persist_tun": true,
        "mssfix": {"enabled": true, "value": 1400},
        "keepalive": {"enabled": true, "interval": 10, "timeout": 120},
        "max_clients": {"enabled": false, "value": null},
        "nobind": false,
        "push_dns": {
          "enabled": true,
          "primary": "8.8.8.8",
          "secondary": "8.8.4.4"
        }
      },
      "status": "active"
    }
  ]
}
```

### client_template.json
Stores client template settings with locked parameters:
```json
{
  "server_config_id": "uuid-here",
  "server_config_name": "myserver",
  "updated_at": "2025-01-25T10:30:00Z",
  "filename_settings": {
    "prefix": "Server1",
    "separator": "_"
  },
  "locked_parameters": {
    "dev": {"type": "tun", "number": 0, "locked": true},
    "proto": {"value": "udp", "locked": true},
    "remote": {
      "address": "CONFIGURED_IN_SETTINGS",
      "port": 1194,
      "locked": true,
      "note": "Configure from Settings > Server IP or Domain"
    },
    "cipher": {"value": "AES-256-GCM", "locked": true},
    "auth": {"value": "SHA512", "locked": true},
    "persist_key": {"enabled": true, "locked": true},
    "persist_tun": {"enabled": true, "locked": true},
    "mssfix": {"enabled": true, "value": 1400, "locked": true},
    "nobind": {"enabled": false, "locked": true},
    "dns": {
      "inherited_from_server": true,
      "primary": "8.8.8.8",
      "secondary": "8.8.4.4",
      "locked": true
    }
  },
  "optional_client_params": {
    "resolv_retry_infinite": {
      "enabled": false,
      "description": "Keeps trying to resolve server hostname if DNS fails"
    },
    "ignore_unknown_option_block_outside_dns": {
      "enabled": false,
      "description": "Prevents DNS leaks on Windows"
    }
  }
}
```

## 🔌 API Endpoints

### Server Configuration Management
- `POST /api/server-config/create` - Create new server config
- `GET /api/server-config/list` - List all server configs
- `GET /api/server-config/<name>` - Get specific server config
- `DELETE /api/server-config/<name>` - Delete server config

### Client Configuration Management
- `POST /api/client-config/create` - Create client config from template
- `GET /api/client-config/download/<filename>` - Download client config

### Template Management
- `GET /api/client-template` - Get current client template

### System
- `GET /api/health` - Health check

## 🔒 Security Features

1. **Path Traversal Protection** - All filenames sanitized
2. **File Locking** - Concurrent access protection
3. **Atomic Writes** - Prevents file corruption
4. **Permission Control** - Secure file permissions (600)
5. **Input Validation** - Comprehensive validation
6. **Error Handling** - Graceful error responses

## 🎨 Frontend Integration

The frontend wizard now communicates with the backend API:

### Key Features
- **Real-time API Communication** - Sends wizard data to backend
- **Loading States** - Professional loading indicators
- **Error Handling** - User-friendly error messages
- **Success Feedback** - Detailed success information
- **Theme Integration** - Full support for all 5 IranGate themes

### Updated JavaScript Methods
- `generateConfigurations()` - Now calls backend API
- `showSuccessMessage()` - Displays backend response
- `showErrorMessage()` - Handles API errors
- `showLoadingState()` - Loading indicators
- `hideLoadingState()` - Clean up loading states

## 🧪 Testing

### Automated Testing
```bash
# Run comprehensive API tests
bash /root/OV-Panel/irangate/webpanel/backend/test_api.sh
```

### Manual Testing
```bash
# Test health endpoint
curl http://localhost:5000/api/health

# Create server config
curl -X POST http://localhost:5000/api/server-config/create \
  -H "Content-Type: application/json" \
  -d '{"name": "test", "deviceType": "tun", "deviceNumber": 0, "protocol": "udp", "port": 1194, "cipher": "AES-256-GCM"}'

# Create client config
curl -X POST http://localhost:5000/api/client-config/create \
  -H "Content-Type: application/json" \
  -d '{"client_name": "testclient"}'
```

## 📋 Example Flow

### 1. Create Server Configuration
```bash
curl -X POST http://localhost:5000/api/server-config/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "myserver",
    "deviceType": "tun",
    "deviceNumber": 0,
    "protocol": "udp",
    "port": 1194,
    "cipher": "AES-256-GCM",
    "keepalive": true,
    "persistKey": true,
    "persistTun": true,
    "mssFix": true,
    "nobind": false,
    "maxClients": false,
    "pushDns": true,
    "primaryDns": "8.8.8.8",
    "secondaryDns": "8.8.4.4",
    "prefixName": "Server1",
    "prefixSeparator": "_"
  }'
```

**Result:**
- Server config saved to `/etc/openvpn/myserver.conf`
- Server config saved to `/etc/openvpn/server/myserver.conf`
- Configuration stored in `server_configs.json`
- Client template created in `client_template.json`

### 2. Create Client Configuration
```bash
curl -X POST http://localhost:5000/api/client-config/create \
  -H "Content-Type: application/json" \
  -d '{"client_name": "john_doe"}'
```

**Result:**
- Client config saved to `/etc/openvpn/client-templates/Server1_john_doe.ovpn`
- All locked parameters inherited from server config
- Filename uses prefix and separator from template

## 🚨 Troubleshooting

### Common Issues

1. **Service won't start:**
```bash
sudo journalctl -u irangate-backend.service -f
```

2. **Permission errors:**
```bash
sudo chown -R root:root /var/lib/irangate
sudo chmod -R 700 /var/lib/irangate
```

3. **API connection issues:**
```bash
curl -v http://localhost:5000/api/health
```

4. **File creation issues:**
```bash
sudo mkdir -p /etc/openvpn/server
sudo mkdir -p /etc/openvpn/client-templates
sudo chmod 700 /etc/openvpn/server
sudo chmod 700 /etc/openvpn/client-templates
```

## 📈 Performance

- **Fast JSON Operations** - Optimized file locking
- **Atomic Writes** - No file corruption
- **Concurrent Access** - Safe multi-user operations
- **Memory Efficient** - Minimal resource usage

## 🔄 Updates

To update the backend:

1. **Stop the service:**
```bash
sudo systemctl stop irangate-backend.service
```

2. **Update files:**
```bash
# Copy new files to both directories
cp -r /root/OV-Panel/irangate/webpanel/backend/* /opt/irangate/webpanel/backend/
```

3. **Restart the service:**
```bash
sudo systemctl start irangate-backend.service
```

## 📞 Support

For issues or questions:
1. Check the logs: `sudo journalctl -u irangate-backend.service -f`
2. Run the test script: `bash test_api.sh`
3. Verify file permissions and directories
4. Check API documentation: `API_DOCUMENTATION.md`

---

**🎉 The IranGate OpenVPN Configuration Backend is now complete and production-ready!**
