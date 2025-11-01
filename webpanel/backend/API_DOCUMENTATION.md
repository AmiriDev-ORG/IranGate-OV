# IranGate OpenVPN Configuration Backend API Documentation

## Overview
This Flask-based backend API manages OpenVPN server and client configurations using JSON-based storage. The API provides secure file operations, configuration management, and client template generation.

## Base URL
```
http://localhost:5000
```

## Authentication
Currently no authentication is implemented. In production, add proper authentication middleware.

## API Endpoints

### 1. Health Check
**GET** `/api/health`

Returns the health status of the backend service.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-25T10:30:00Z",
  "directories": {
    "base": "/var/lib/irangate",
    "openvpn": "/etc/openvpn",
    "openvpn_server": "/etc/openvpn/server",
    "client_templates": "/etc/openvpn/client-templates"
  }
}
```

### 2. Server Configuration Management

#### Create Server Configuration
**POST** `/api/server-config/create`

Creates a new OpenVPN server configuration and saves it to both required directories.

**Request Body:**
```json
{
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
  "maxClientsCount": null,
  "pushDns": true,
  "primaryDns": "8.8.8.8",
  "secondaryDns": "8.8.4.4",
  "prefixName": "Server1",
  "prefixSeparator": "_",
  "resolvRetry": false,
  "ignoreUnknownOption": false
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Server configuration created successfully",
  "config_id": "uuid-here",
  "config_name": "myserver",
  "file_paths": [
    "/etc/openvpn/myserver.conf",
    "/etc/openvpn/server/myserver.conf"
  ]
}
```

**Response (Error):**
```json
{
  "error": "Configuration name already exists"
}
```

#### List Server Configurations
**GET** `/api/server-config/list`

Returns a list of all server configurations.

**Response:**
```json
{
  "configs": [
    {
      "id": "uuid-here",
      "name": "myserver",
      "created_at": "2025-01-25T10:30:00Z",
      "status": "active"
    }
  ]
}
```

#### Get Server Configuration Details
**GET** `/api/server-config/<config_name>`

Returns detailed information about a specific server configuration.

**Response:**
```json
{
  "id": "uuid-here",
  "name": "myserver",
  "created_at": "2025-01-25T10:30:00Z",
  "file_path_1": "/etc/openvpn/myserver.conf",
  "file_path_2": "/etc/openvpn/server/myserver.conf",
  "parameters": {
    "dev": {
      "type": "tun",
      "number": 0
    },
    "proto": "udp",
    "port": 1194,
    "cipher": "AES-256-GCM",
    "auth": "SHA512",
    "persist_key": true,
    "persist_tun": true,
    "mssfix": {
      "enabled": true,
      "value": 1400
    },
    "keepalive": {
      "enabled": true,
      "interval": 10,
      "timeout": 120
    },
    "max_clients": {
      "enabled": false,
      "value": null
    },
    "nobind": false,
    "push_dns": {
      "enabled": true,
      "primary": "8.8.8.8",
      "secondary": "8.8.4.4"
    }
  },
  "status": "active"
}
```

#### Delete Server Configuration
**DELETE** `/api/server-config/<config_name>`

Deletes a server configuration and removes associated files.

**Response:**
```json
{
  "success": true,
  "message": "Configuration deleted successfully"
}
```

### 3. Client Configuration Management

#### Create Client Configuration
**POST** `/api/client-config/create`

Creates a client configuration using the current template.

**Request Body:**
```json
{
  "client_name": "john_doe"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Client configuration created successfully",
  "filename": "Server1_john_doe.ovpn",
  "file_path": "/etc/openvpn/client-templates/Server1_john_doe.ovpn",
  "download_url": "/api/client-config/download/Server1_john_doe.ovpn"
}
```

#### Download Client Configuration
**GET** `/api/client-config/download/<filename>`

Downloads a client configuration file.

**Response:** Binary file download

### 4. Client Template Management

#### Get Client Template
**GET** `/api/client-template`

Returns the current client template settings.

**Response:**
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
    "dev": {
      "type": "tun",
      "number": 0,
      "locked": true
    },
    "proto": {
      "value": "udp",
      "locked": true
    },
    "remote": {
      "address": "CONFIGURED_IN_SETTINGS",
      "port": 1194,
      "locked": true,
      "note": "Configure from Settings > Server IP or Domain"
    },
    "cipher": {
      "value": "AES-256-GCM",
      "locked": true
    },
    "auth": {
      "value": "SHA512",
      "locked": true
    },
    "persist_key": {
      "enabled": true,
      "locked": true
    },
    "persist_tun": {
      "enabled": true,
      "locked": true
    },
    "mssfix": {
      "enabled": true,
      "value": 1400,
      "locked": true
    },
    "nobind": {
      "enabled": false,
      "locked": true
    },
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

## Error Handling

All endpoints return appropriate HTTP status codes:

- **200 OK** - Success
- **400 Bad Request** - Invalid input data
- **404 Not Found** - Resource not found
- **500 Internal Server Error** - Server error

Error responses include a descriptive error message:

```json
{
  "error": "Configuration name already exists"
}
```

## Security Features

1. **Path Traversal Protection** - All filenames are sanitized
2. **File Locking** - JSON files use file locking for concurrent access
3. **Atomic Writes** - Configuration files are written atomically
4. **Permission Control** - Files are created with secure permissions (600)
5. **Input Validation** - All inputs are validated before processing

## File Structure

```
/var/lib/irangate/
├── server_configs.json    # Server configuration database
└── client_template.json   # Client template settings

/etc/openvpn/
├── <config_name>.conf     # Server config (copy 1)
└── server/
    └── <config_name>.conf # Server config (copy 2)

/etc/openvpn/client-templates/
└── <prefix><separator><client_name>.ovpn
```

## Installation

1. Run the installation script:
```bash
sudo bash /path/to/install.sh
```

2. The service will be automatically started and enabled.

3. Check service status:
```bash
sudo systemctl status irangate-backend.service
```

## Usage Examples

### Create a Server Configuration
```bash
curl -X POST http://localhost:5000/api/server-config/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "myserver",
    "deviceType": "tun",
    "deviceNumber": 0,
    "protocol": "udp",
    "port": 1194,
    "cipher": "AES-256-GCM"
  }'
```

### Create a Client Configuration
```bash
curl -X POST http://localhost:5000/api/client-config/create \
  -H "Content-Type: application/json" \
  -d '{"client_name": "john_doe"}'
```

### Download Client Configuration
```bash
curl -O http://localhost:5000/api/client-config/download/Server1_john_doe.ovpn
```
