# IranGate Web Panel API Documentation

## Overview

The IranGate Web Panel provides a RESTful API for managing VPN clients, server configuration, and monitoring. All API endpoints are prefixed with `/api/` and require authentication except for the login endpoint.

## Authentication

### Login

**POST** `/api/login`

Authenticate with username and password to receive a JWT token.

**Request Body:**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "username": "admin",
    "expires_at": "2025-10-15T17:50:30.574950892Z"
  }
}
```

### Using the Token

Include the JWT token in the Authorization header for all protected endpoints:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Client Management

### List Clients

**GET** `/api/clients`

Get a list of all VPN clients.

**Response:**
```json
{
  "success": true,
  "message": "Clients retrieved successfully",
  "data": [
    {
      "name": "client1",
      "active": true,
      "created_at": "2025-10-12T21:01:20.61143035Z",
      "expires_at": "2025-11-12T21:01:20.611430472Z",
      "ip": "",
      "last_connection": null,
      "bytes_received": 0,
      "bytes_sent": 0,
      "download": 0,
      "upload": 0,
      "data_used_mb": 0,
      "cipher": "AES-256-GCM",
      "status": ""
    }
  ]
}
```

### Create Client

**POST** `/api/clients`

Create a new VPN client.

**Request Body:**
```json
{
  "name": "newclient",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Client created successfully",
  "data": {
    "name": "newclient",
    "config": "client\nremote server.com 1194\n..."
  }
}
```

### Remove Client

**DELETE** `/api/clients/{name}`

Remove a VPN client and revoke its certificate.

**Response:**
```json
{
  "success": true,
  "message": "Client removed successfully"
}
```

### Get Client Status

**GET** `/api/clients/{name}/status`

Get the current status of a specific client.

**Response:**
```json
{
  "success": true,
  "message": "Client status retrieved",
  "data": {
    "name": "client1",
    "connected": true,
    "ip": "10.8.0.2",
    "bytes_received": 1024000,
    "bytes_sent": 512000,
    "connected_since": "2025-10-14T10:30:00Z"
  }
}
```

### Export Client Config

**GET** `/api/clients/{name}/export`

Download the OpenVPN configuration file for a client.

**Response:** Binary file (application/octet-stream)

### List Orphaned Certificates

**GET** `/api/clients/orphaned`

Get a list of orphaned certificates that can be cleaned up.

**Response:**
```json
{
  "success": true,
  "message": "Orphaned certificates retrieved",
  "data": [
    {
      "name": "oldclient",
      "created_at": "2025-01-01T00:00:00Z",
      "last_used": "2025-02-01T00:00:00Z"
    }
  ]
}
```

### Cleanup Orphaned Certificates

**POST** `/api/clients/cleanup-orphaned`

Remove orphaned certificates from the system.

**Response:**
```json
{
  "success": true,
  "message": "Orphaned certificates cleaned up",
  "data": {
    "removed_count": 3
  }
}
```

## Server Management

### Get Server Status

**GET** `/api/server/status`

Get the current status of the OpenVPN server.

**Response:**
```json
{
  "success": true,
  "message": "Server status retrieved",
  "data": {
    "running": true,
    "uptime": "2d 5h 30m",
    "clients_connected": 5,
    "total_clients": 25,
    "version": "OpenVPN 2.5.1"
  }
}
```

### Restart Server

**POST** `/api/server/restart`

Restart the OpenVPN server.

**Response:**
```json
{
  "success": true,
  "message": "Server restarted successfully"
}
```

### Get Server Configuration

**GET** `/api/config`

Get the current OpenVPN server configuration.

**Response:**
```json
{
  "success": true,
  "message": "Configuration retrieved",
  "data": {
    "port": 1194,
    "protocol": "udp",
    "cipher": "AES-256-GCM",
    "auth": "SHA256",
    "dh_bits": 2048,
    "server_network": "10.8.0.0/24"
  }
}
```

### Update Server Configuration

**PUT** `/api/config`

Update the OpenVPN server configuration.

**Request Body:**
```json
{
  "port": 1194,
  "protocol": "udp",
  "cipher": "AES-256-GCM",
  "auth": "SHA256",
  "dh_bits": 2048,
  "server_network": "10.8.0.0/24"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Configuration updated successfully"
}
```

## Monitoring

### Get Monitoring Statistics

**GET** `/api/monitor/stats`

Get real-time monitoring statistics.

**Response:**
```json
{
  "success": true,
  "message": "Statistics retrieved",
  "data": {
    "cpu_usage": 15.5,
    "memory_usage": 1024,
    "disk_usage": 2048,
    "network_rx": 1024000,
    "network_tx": 512000,
    "uptime": "2d 5h 30m",
    "connected_clients": 5
  }
}
```

## Settings

### Get Server Settings

**GET** `/api/settings/server`

Get server-related settings.

**Response:**
```json
{
  "success": true,
  "message": "Server settings retrieved",
  "data": {
    "auto_start": true,
    "log_level": "info",
    "max_clients": 100,
    "compression": false
  }
}
```

### Update Server Settings

**PUT** `/api/settings/server`

Update server-related settings.

**Request Body:**
```json
{
  "auto_start": true,
  "log_level": "debug",
  "max_clients": 150,
  "compression": true
}
```

**Response:**
```json
{
  "success": true,
  "message": "Server settings updated successfully"
}
```

### Get Admin Settings

**GET** `/api/settings/admin`

Get admin user settings.

**Response:**
```json
{
  "success": true,
  "message": "Admin settings retrieved",
  "data": {
    "username": "admin",
    "email": "admin@irangate.local",
    "created_at": "2025-10-12T23:24:31.669134217Z",
    "last_login": "2025-10-14T17:27:57.537459287Z"
  }
}
```

### Change Admin Password

**PUT** `/api/settings/admin/password`

Change the admin password.

**Request Body:**
```json
{
  "current_password": "oldpassword",
  "new_password": "newpassword"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Password changed successfully"
}
```

## User Management (RBAC)

### List Users

**GET** `/api/users`

Get a list of all users.

**Response:**
```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "username": "admin",
      "email": "admin@irangate.local",
      "role": "admin",
      "created_at": "2025-10-12T23:24:31.669134217Z",
      "last_login": "2025-10-14T17:27:57.537459287Z",
      "active": true
    }
  ]
}
```

### Create User

**POST** `/api/users`

Create a new user.

**Request Body:**
```json
{
  "username": "newuser",
  "email": "user@example.com",
  "password": "password123",
  "role": "user"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "username": "newuser",
    "email": "user@example.com",
    "role": "user"
  }
}
```

### Update User

**PUT** `/api/users/{username}`

Update an existing user.

**Request Body:**
```json
{
  "email": "newemail@example.com",
  "role": "admin",
  "active": true
}
```

**Response:**
```json
{
  "success": true,
  "message": "User updated successfully"
}
```

### Delete User

**DELETE** `/api/users/{username}`

Delete a user.

**Response:**
```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

## Groups

### List Groups

**GET** `/api/groups`

Get a list of all groups.

**Response:**
```json
{
  "success": true,
  "message": "Groups retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "premium",
      "description": "Premium users group",
      "max_clients": 10,
      "created_at": "2025-10-12T23:24:31.669134217Z"
    }
  ]
}
```

### Create Group

**POST** `/api/groups`

Create a new group.

**Request Body:**
```json
{
  "name": "vip",
  "description": "VIP users group",
  "max_clients": 5
}
```

**Response:**
```json
{
  "success": true,
  "message": "Group created successfully",
  "data": {
    "id": 2,
    "name": "vip",
    "description": "VIP users group",
    "max_clients": 5
  }
}
```

### Update Group

**PUT** `/api/groups/{id}`

Update an existing group.

**Request Body:**
```json
{
  "name": "vip_updated",
  "description": "Updated VIP users group",
  "max_clients": 10
}
```

**Response:**
```json
{
  "success": true,
  "message": "Group updated successfully"
}
```

### Delete Group

**DELETE** `/api/groups/{id}`

Delete a group.

**Response:**
```json
{
  "success": true,
  "message": "Group deleted successfully"
}
```

## Subscriptions

### List Subscriptions

**GET** `/api/subscriptions`

Get a list of all client subscriptions.

**Response:**
```json
{
  "success": true,
  "message": "Subscriptions retrieved successfully",
  "data": [
    {
      "client": "client1",
      "start_date": "2025-10-01T00:00:00Z",
      "end_date": "2025-11-01T00:00:00Z",
      "active": true,
      "auto_renew": false
    }
  ]
}
```

### Create Subscription

**POST** `/api/subscriptions`

Create a new subscription for a client.

**Request Body:**
```json
{
  "client": "client1",
  "start_date": "2025-10-01T00:00:00Z",
  "end_date": "2025-11-01T00:00:00Z",
  "auto_renew": false
}
```

**Response:**
```json
{
  "success": true,
  "message": "Subscription created successfully"
}
```

### Update Subscription

**PUT** `/api/subscriptions/{client}`

Update an existing subscription.

**Request Body:**
```json
{
  "end_date": "2025-12-01T00:00:00Z",
  "auto_renew": true
}
```

**Response:**
```json
{
  "success": true,
  "message": "Subscription updated successfully"
}
```

### Delete Subscription

**DELETE** `/api/subscriptions/{client}`

Delete a subscription.

**Response:**
```json
{
  "success": true,
  "message": "Subscription deleted successfully"
}
```

### Get Expiring Subscriptions

**GET** `/api/subscriptions/expiring`

Get subscriptions that are expiring soon.

**Query Parameters:**
- `days` (optional): Number of days ahead to check (default: 7)

**Response:**
```json
{
  "success": true,
  "message": "Expiring subscriptions retrieved",
  "data": [
    {
      "client": "client1",
      "end_date": "2025-10-15T00:00:00Z",
      "days_remaining": 3
    }
  ]
}
```

### Renew Subscription

**POST** `/api/subscriptions/{client}/renew`

Renew a subscription for a client.

**Request Body:**
```json
{
  "extension_days": 30
}
```

**Response:**
```json
{
  "success": true,
  "message": "Subscription renewed successfully",
  "data": {
    "new_end_date": "2025-11-15T00:00:00Z"
  }
}
```

## Backup & Restore

### Export Backup

**GET** `/api/backup/export`

Export a backup of all configuration and data.

**Response:** Binary file (application/octet-stream)

### Import Backup

**POST** `/api/backup/import`

Import a backup file.

**Request:** Multipart form with backup file

**Response:**
```json
{
  "success": true,
  "message": "Backup restored successfully"
}
```

## Health Check

### Get Health Status

**GET** `/api/health`

Get the health status of the web panel.

**Response:**
```json
{
  "success": true,
  "message": "Health check passed",
  "data": {
    "status": "healthy",
    "timestamp": "2025-10-14T17:52:55Z",
    "uptime": "265ns",
    "version": "1.0.0"
  }
}
```

## WebSocket

### Real-time Updates

**GET** `/ws`

Establish a WebSocket connection for real-time updates.

**Connection:** WebSocket upgrade request

**Messages:** JSON objects with update information

**Example Message:**
```json
{
  "type": "client_status",
  "data": {
    "client": "client1",
    "connected": true,
    "ip": "10.8.0.2"
  }
}
```

## Error Responses

All endpoints return consistent error responses:

```json
{
  "success": false,
  "message": "Error description",
  "error": "DETAILED_ERROR_CODE"
}
```

**Common HTTP Status Codes:**
- `200` - Success
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

## Rate Limiting

API endpoints are rate-limited to prevent abuse:
- 100 requests per minute per IP
- 10 requests per minute for authentication endpoints

## CORS

The API supports Cross-Origin Resource Sharing (CORS) with the following configuration:
- Allowed Origins: `*` (all origins)
- Allowed Methods: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`
- Allowed Headers: `Content-Type`, `Authorization`

## Examples

### cURL Examples

**Login:**
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  http://localhost:8080/api/login
```

**Get Clients (with token):**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/clients
```

**Create Client:**
```bash
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"name":"newclient"}' \
  http://localhost:8080/api/clients
```

### JavaScript Examples

**Login and get token:**
```javascript
const login = async (username, password) => {
  const response = await fetch('/api/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, password }),
  });
  const data = await response.json();
  return data.data.token;
};

// Usage
const token = await login('admin', 'admin123');
```

**Get clients with authentication:**
```javascript
const getClients = async (token) => {
  const response = await fetch('/api/clients', {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });
  const data = await response.json();
  return data.data;
};

// Usage
const clients = await getClients(token);
```

## Version

- **API Version**: 1.0.0
- **Last Updated**: October 2025
- **Compatibility**: IranGate Web Panel v1.0.0+
