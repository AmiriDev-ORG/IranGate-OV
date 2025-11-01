# Client Lifecycle Management Features

## Overview

This document describes the new client lifecycle management features added to IranGate-OV Web Panel.

## Features

### 1. Custom Client Expiration

**Location:** Web Panel > Clients > Add Client

When creating a new client, you can now specify a custom expiration period:

- **Field:** Expiration (Days)
- **Default:** 30 days
- **Range:** 1+ days
- **Behavior:** After the expiration date, the client will be automatically revoked

**Example:**
```json
{
  "name": "client1",
  "expiration_days": 60,
  "traffic_limit_bytes": 0
}
```

### 2. Traffic Limit per Client

**Location:** Web Panel > Clients > Add Client

Set a traffic limit for each client:

- **Field:** Traffic Limit (Optional)
- **Units:** MB, GB, TB
- **Default:** 0 (unlimited)
- **Behavior:** When the total traffic (upload + download) exceeds the limit, the client is automatically revoked

**Example:**
```json
{
  "name": "client1",
  "expiration_days": 30,
  "traffic_limit_bytes": 10737418240
}
```
*Note: 10737418240 bytes = 10 GB*

### 3. Manual Client Activation/Deactivation

**Location:** Web Panel > Clients > Actions Column

Each client now has activation controls:

- **Active Status:** Green icon with checkmark
- **Inactive Status:** Yellow icon with X

**Actions:**
- Click the icon to toggle between Active/Inactive
- Deactivating a client will:
  - Revoke the certificate
  - Disconnect active sessions
  - Prevent new connections
  - Update the status in the database

- Activating a client will:
  - Mark as active in the database
  - Allow the client to reconnect (if certificate is valid)

### 4. Automatic Lifecycle Management (Cronjobs)

**Location:** Background Service (runs every 30 minutes)

Two cronjobs run automatically:

#### Expiration Check
- Scans all clients every 30 minutes
- Checks if current time > expiration date
- If expired:
  - Revokes the certificate
  - Marks as inactive
  - Logs the action

#### Traffic Limit Check
- Scans all clients every 30 minutes
- Checks if total traffic > traffic limit
- If exceeded:
  - Revokes the certificate
  - Marks as inactive
  - Logs the action

**Log Location:** `/opt/irangate/logs/revocations.json`

**Log Format:**
```json
{
  "client_name": "client1",
  "reason": "expired",
  "timestamp": "2025-10-31T20:00:00Z",
  "success": true,
  "error": ""
}
```

## API Endpoints

### Create Client with Custom Settings
```http
POST /api/clients
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "client1",
  "email": "client1@example.com",
  "expiration_days": 60,
  "traffic_limit_bytes": 10737418240
}
```

### Activate Client
```http
POST /api/clients/{name}/activate
Authorization: Bearer <token>
```

### Deactivate Client
```http
POST /api/clients/{name}/deactivate
Authorization: Bearer <token>
```

## Database Schema Updates

### Client Model
```go
type Client struct {
    Name              string
    CreatedAt         time.Time
    ExpiresAt         time.Time
    Active            bool
    TrafficLimitBytes int64  // NEW FIELD
    BytesReceived     uint64
    BytesSent         uint64
    // ... other fields
}
```

## Configuration

The cronjob interval is set to **30 minutes** by default. This can be modified in the source code:

**File:** `webpanel/backend/api.go`
```go
go StartClientLifecycleMonitor(30) // 30 minutes
```

## Safety Features

1. **Mutex Locking:** All revocation operations use mutex to prevent race conditions
2. **Error Handling:** Failed revocations are logged but don't stop the process for other clients
3. **Graceful Degradation:** If easyrsa fails, the system logs the error and continues
4. **Audit Trail:** All revocation actions are logged with timestamp, reason, and status
5. **Non-Destructive:** Does not modify core traffic collection or OpenVPN systems

## Troubleshooting

### Check Cronjob Status
Check the web panel logs:
```bash
journalctl -u irangate-webpanel -f
```

Look for messages like:
```
🔍 Checking for expired clients...
⏰ Client client1 has expired (expiry: 2025-10-31 20:00:00)
✅ Expiration check complete. Found 1 expired clients
```

### Check Revocation Logs
```bash
cat /opt/irangate/logs/revocations.json | jq
```

### Manually Test Revocation
```bash
# Via CLI
irangate client remove <client_name>

# Via API
curl -X POST http://localhost:8080/api/clients/<name>/deactivate \
  -H "Authorization: Bearer <token>"
```

## Testing Checklist

- [x] Database schema updated with TrafficLimitBytes field
- [x] Create client with custom expiration (Frontend form)
- [x] Create client with traffic limit (Frontend form)
- [x] Activate/Deactivate toggle button (Frontend UI)
- [x] Activation API endpoint
- [x] Deactivation API endpoint
- [x] Expiration cronjob (every 30 minutes)
- [x] Traffic limit cronjob (every 30 minutes)
- [x] Certificate revocation function
- [x] Revocation logging
- [x] Compiled successfully (no errors)

## Future Enhancements

Possible future improvements:
- Configurable cronjob intervals (via web panel settings)
- Email notifications for expiring clients
- Grace period before revocation
- Client suspension (temporary deactivation without revocation)
- Batch operations for activation/deactivation
- Traffic usage history and charts per client

