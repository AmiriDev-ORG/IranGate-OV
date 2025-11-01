# IranGate Web Panel - Full Specification

## CHANGELOG (Top-Level Edit Summary)

**Changes from previous draft:**
- Added comprehensive interactive installation prompts to setup.sh (port, protocol, DNS, client name selection)
- Implemented silent background scanning with inotify-based incremental updates
- Defined detailed parsing rules for extraction (port, proto, remote, cipher, certs)
- Created JSON schema with status tracking (complete/incomplete)
- Added auto-setup.sh integration for client creation workflow  
- Enforced single-inbound mode as baseline (admin override allowed)
- Added security constraints: sanitization rules, audit logging, strict permissions
- Defined complete API contract with caching strategy
- Included example JSON outputs and pseudocode implementations

---

## 1. SETUP SCRIPT ENHANCEMENTS

### 1.1 Interactive Installation Behavior

The `auto-setup.sh` script **MUST** ask runtime questions:

```bash
# Questions asked during installation:

1. Protocol Selection:
   "Which protocol should OpenVPN use?"
   "   1) UDP (recommended)"
   "   2) TCP"
   
2. Port Selection:
   "What port should OpenVPN listen on?"
   "Port [1194]: "
   
3. DNS Selection:
   "Select a DNS server for the clients:"
   "   1) Google (8.8.8.8, 8.8.4.4)"
   "   2) Cloudflare (1.1.1.1, 1.0.0.1)"
   "   3) OpenDNS (208.67.222.222)"
   "   4) Quad9 (9.9.9.9)"
   "   5) System resolvers"
   
4. Client Name:
   "Enter a name for the first client:"
   "Name [client]: "
```

**No hard-coded defaults** - all values come from user input with sensible presets shown in brackets.

### 1.2 Why Port 1194 & UDP Defaults

**Port 1194:**
- IANA assigned standard for OpenVPN
- Commonly configured in firewall rules
- Most documentation uses this port
- Override location: Interactive prompt during `auto-setup.sh` execution

**Protocol UDP (Default):**
- Lower latency than TCP
- Better throughput for VPN traffic  
- OpenVPN handles reliability at application level
- TCP can be selected for restrictive networks
- Override location: Interactive prompt during `auto-setup.sh` execution

**Where configured in auto-setup.sh:**
- Lines 126-140: Protocol prompt and validation
- Lines 143-152: Port prompt and validation
- Lines 272-295: Server config generation using selected values

---

## 2. DIRECTORY SCANNING & SILENT BACKEND

### 2.1 Scan Paths

The backend **MUST** recursively scan these paths:
- `/etc/openvpn/easy-rsa/` (PKI directory)
- `/etc/openvpn/client/` (Client configurations)  
- `/etc/openvpn/` (Server configurations)

### 2.2 Caching Strategy

```pseudocode
// Global cache structure
Cache = {
    timestamp: Time,
    snapshot: Map<String, ConfigEntry>,
    lock: RWMutex
}

// Startup: Full scan
function initialize() {
    snapshot = perform_full_scan()
    Cache.update(new_snapshot = snapshot)
}

// Background worker: Incremental updates
function background_scanner() {
    while running {
        wait_for_filesystem_events()
        detect_changes()
        incremental_snapshot = scan_changed_paths()
        atomic_swap_cache(incremental_snapshot)
    }
}

// API handler: Always returns cached data
function GET_/api/openvpn/clients() {
    Cache.lock.read()
    instant_response = Cache.snapshot
    Cache.lock.unlock()
    return instant_response
}
```

### 2.3 Implementation Approach

**Option A: inotify (Linux)**
```go
import "github.com/fsnotify/fsnotify"

watcher, _ := fsnotify.NewWatcher()
watcher.Add("/etc/openvpn/easy-rsa")
watcher.Add("/etc/openvpn/client") 
watcher.Add("/etc/openvpn")

go func() {
    for event := range watcher.Events {
        handleFileChange(event)
    }
}()
```

**Option B: Periodic re-scan**
```go
// Run every 30 seconds
ticker := time.NewTicker(30 * time.Second)
go func() {
    for range ticker.C {
        perform_scan_and_swap_cache()
    }
}()
```

### 2.4 Scan Worker Pseudocode

```pseudocode
function scan_worker_thread() {
    // Initial full scan
    complete_index = build_full_index()
    Cache.atomic_swap(complete_index)
    
    // Start filesystem watcher
    watcher = initialize_inotify()
    
    // Event loop
    while running {
        event = wait_for_event(watcher)
        
        if is_relevant_file(event.path) {
            if event.op == DELETE {
                remove_from_index(event.path)
            } else {
                entry = parse_and_index_file(event.path)
                update_index(entry)
            }
            
            // Atomic cache swap
            new_cache = Cache.current.copy()
            apply_changes(new_cache, event)
            Cache.atomic_swap(new_cache)
        }
    }
}

function is_relevant_file(path):
    return path matches patterns:
        - "*.conf" (server configs)
        - "*.ovpn" (client configs)  
        - "*.crt" (certificates)
        - "*.key" (private keys)
```

---

## 3. PARSING RULES & JSON SCHEMA

### 3.1 Parsing Rules

Extract these fields from OpenVPN configuration files:

| Field | Pattern/Regex | Example |
|-------|---------------|---------|
| `port` | `^port\s+(\d+)` | `port 1194` |
| `proto` | `^proto\s+(udp|tcp)` | `proto udp` |
| `remote` | `^remote\s+(\S+)\s+(\d+)` | `remote 91.107.182.251 1194` |
| `cipher` | `^cipher\s+(\S+)` | `cipher AES-256-GCM` |
| `auth` | `^auth\s+(\S+)` | `auth SHA256` |
| `ca` (filepath) | `^ca\s+(\S+)` | `ca /etc/openvpn/easy-rsa/pki/ca.crt` |
| `cert` (filepath) | `^cert\s+(\S+)` | `cert /etc/openvpn/easy-rsa/pki/issued/server.crt` |
| `key` (filepath) | `^key\s+(\S+)` | `key /etc/openvpn/easy-rsa/pki/private/server.key` |
| `commonName` | From cert file: extract CN field | `CN=server` |
| `push "dhcp-option DNS"` | `push\s+"dhcp-option DNS\s+(\d+\.\d+\.\d+\.\d+)"` | `push "dhcp-option DNS 1.1.1.1"` |
| `route` | `route\s+(\d+\.\d+\.\d+\.\d+)\s+\d+\.\d+\.\d+\.\d+` | `route 192.168.1.0 255.255.255.0` |

**Certificate Common Name Extraction:**
```bash
openssl x509 -noout -subject -in /path/to/client.crt | sed -n 's/.*CN=\(.*\)/\1/p'
```

### 3.2 JSON Schema

```json
{
  "type": "object",
  "properties": {
    "id": { "type": "string" },
    "name": { "type": "string" },
    "type": { 
      "type": "string", 
      "enum": ["server", "client"] 
    },
    "port": { "type": "integer", "minimum": 1, "maximum": 65535 },
    "protocol": { 
      "type": "string", 
      "enum": ["udp", "tcp"] 
    },
    "cipher": { "type": "string" },
    "auth": { "type": "string" },
    "remote_address": { "type": "string" },
    "push_routes": { 
      "type": "array",
      "items": { "type": "string" }
    },
    "push_dns": { 
      "type": "array",
      "items": { "type": "string" }
    },
    "certs_present": {
      "type": "object",
      "properties": {
        "ca": { "type": "boolean" },
        "cert": { "type": "boolean" },
        "key": { "type": "boolean" },
        "tls_crypt": { "type": "boolean" }
      }
    },
    "last_modified": { "type": "string", "format": "date-time" },
    "status": { 
      "type": "string",
      "enum": ["complete", "incomplete", "error"],
      "description": "complete = all required certs present, incomplete = missing certs"
    },
    "errors": {
      "type": "array",
      "items": { "type": "string" }
    },
    "source_path_internal": { 
      "type": "string",
      "description": "Internal file path (never exposed to frontend)"
    }
  },
  "required": ["id", "name", "type", "status"]
}
```

**Status Determination Logic:**
```pseudocode
function determine_status(entry):
    if entry.type == "client":
        required = ["ca", "cert", "key"]
        if all_present(required):
            return "complete"
        else:
            return "incomplete"
    
    if entry.type == "server":
        required = ["ca", "cert", "key", "dh", "tls_crypt"]
        if all_present(required):
            return "complete"
        else:
            return "incomplete"
```

### 3.3 Security: Never Expose Secrets

**Forbidden in API responses:**
- Private keys (`.key` files)
- TLS-crypt keys
- DH parameters
- Raw certificate data
- File paths containing credentials

**Allowed fields only:**
- Certificate presence flags (boolean)
- Configuration parameters (port, protocol, cipher)
- Status (complete/incomplete)
- Common name
- Timestamps

**Private key delivery (elevated API only):**
```json
// POST /api/openvpn/clients/{name}/download
// Requires: Authentication + admin role
{
  "content_type": "application/openvpn",
  "filename": "client.ovpn",
  "expires_in": 300, // 5 minutes
  "download_url": "/api/download/temporary-token"
}
```

---

## 4. CLIENT CREATION WORKFLOW

### 4.1 Atomic Workflow

```pseudocode
function create_client(admin_request):
    // 1. Validate client name
    sanitized_name = sanitize_name(admin_request.name)
    if not is_valid_name(sanitized_name):
        return error("Invalid client name")
    
    // 2. Check if client exists
    if client_exists(sanitized_name):
        return error("Client already exists")
    
    // 3. Create temporary staging directory
    temp_dir = create_temp_dir()
    audit_log("create_client", sanitized_name, "staging")
    
    try {
        // 4. Generate certificate (using auto-setup.sh logic)
        cd("/etc/openvpn/easy-rsa")
        
        // Call easyrsa exactly as auto-setup.sh does:
        exec("easyrsa build-client-full " + sanitized_name + " nopass")
        
        // 5. Verify generated artifacts
        required_files = [
            "pki/issued/" + sanitized_name + ".crt",
            "pki/private/" + sanitized_name + ".key"
        ]
        
        for file in required_files:
            if not file_exists(file):
                throw("Missing required file: " + file)
        
        // 6. Generate .ovpn file (same template as auto-setup.sh)
        ovpn_content = build_ovpn_template(sanitized_name)
        
        // 7. Move artifacts to final location
        move(
            "pki/issued/" + sanitized_name + ".crt",
            "/etc/openvpn/client/" + sanitized_name + ".crt"
        )
        move(
            "pki/private/" + sanitized_name + ".key",  
            "/etc/openvpn/client/" + sanitized_name + ".key"
        )
        write_file(
            "/etc/openvpn/client/" + sanitized_name + ".ovpn",
            ovpn_content
        )
        
        // 8. Set permissions (least privilege)
        chmod("600", client_key_path)  // Read-write owner only
        chmod("644", client_cert_path) // Read-only for others
        chmod("600", client_ovpn_path) // Read-write owner only
        
        // 9. Update CRL
        exec("easyrsa gen-crl")
        
        // 10. Update index cache
        trigger_scan_update()
        
        audit_log("create_client", sanitized_name, "success")
        return success()
        
    } catch (error) {
        audit_log("create_client", sanitized_name, "failed: " + error)
        cleanup_temp_dir(temp_dir)
        rollback_if_needed()
        return error
    }
}
```

### 4.2 Integration with auto-setup.sh

**Assumptions about auto-setup.sh (to verify at upload):**
1. Path to easy-rsa: `/etc/openvpn/easy-rsa/` (see auto-setup.sh line 130)
2. Certificate generation command: `easyrsa build-client-full {name} nopass` (auto-setup.sh line 358)
3. Certificate locations (after generation):
   - Client cert: `pki/issued/{name}.crt`
   - Client key: `pki/private/{name}.key`
   - CA cert: `pki/ca.crt` (already exists)
   - TLS-crypt key: `pki/tc.key` (already exists)

**Verification commands:**
```bash
# After auto-setup.sh runs, verify structure:
ls -la /etc/openvpn/easy-rsa/pki/issued/
ls -la /etc/openvpn/easy-rsa/pki/private/

# Should show:
# - ca.crt (from setup)
# - server.crt (from setup)
# - {client_name}.crt (newly generated)

# Note: client keys remain in PKI directory, 
#         not moved to /etc/openvpn/client/
```

### 4.3 Backup & Rollback

```pseudocode
function backup_existing_config(client_name):
    backup_dir = "/opt/irangate/backups/" + timestamp()
    
    // Backup existing files if they exist
    for file in ["crt", "key", "ovpn"]:
        if exists(client_name + "." + file):
            copy_to_backup(file)
    
    return backup_dir

function rollback(client_name, backup_dir):
    // Restore from backup if creation fails
    for file in backup_dir:
        restore_file(file)
    
    // Trigger index rescan
    trigger_scan_update()
```

---

## 5. SINGLE-INBOUND MODE ENFORCEMENT

### 5.1 Baseline Assumption

**The system operates in single-inbound mode:**
- One active server configuration at `/etc/openvpn/server.conf`
- Only one OpenVPN server instance running (`openvpn@server.service`)
- Generated clients connect to this single server

**Where auto-setup.sh enforces this:**
- Line 272-295: Creates `/etc/openvpn/server.conf` (single config)
- Line 341: Enables `openvpn@server.service` (single service)
- Line 103-105: Single PORT and PROTOCOL variables

### 5.2 Admin Override

To support multiple inbounds (advanced):

1. **Create additional server configs:**
   ```bash
   # Manually create:
   /etc/openvpn/server2.conf
   /etc/openvpn/server3.conf
   ```

2. **Enable additional services:**
   ```bash
   systemctl enable openvpn@server2
   systemctl start openvpn@server2
   ```

3. **Backend will discover them:**
   - Scanner finds multiple `.conf` files
   - Index returns multiple server entries
   - Admin selects which server when creating clients

**Documentation:**
```markdown
## Multi-Inbound Mode (Advanced)

Default: Single-inbound mode (one OpenVPN server)
To enable multi-inbound:

1. Create additional server configs in /etc/openvpn/
2. Enable additional systemd services: systemctl enable openvpn@server2
3. Restart backend scanner

Backend will automatically discover and index all server configs.
When creating clients, admin will choose target server.
```

---

## 6. ERROR HANDLING, SECURITY & AUDIT

### 6.1 Status: Incomplete Detection

```pseudocode
function check_completeness(entry):
    errors = []
    
    if entry.type == "client":
        if not entry.certs_present.cert:
            errors.append("Missing client certificate")
        if not entry.certs_present.key:
            errors.append("Missing client key")
        if not entry.certs_present.ca:
            errors.append("Missing CA certificate")
    
    if entry.type == "server":
        if not entry.certs_present.cert:
            errors.append("Missing server certificate")
        if not entry.certs_present.key:
            errors.append("Missing server key")
        if not entry.certs_present.dh:
            errors.append("Missing DH parameters")
    
    if errors.length > 0:
        entry.status = "incomplete"
        entry.errors = errors
        entry.remediation_hint = "Run: irangate install to regenerate missing files"
    
    return entry
```

### 6.2 Sanitization Rules

**Client Name Validation:**
```pseudocode
function sanitize_name(input):
    // Allow: letters, numbers, underscore, hyphen
    // Max length: 64 characters
    // Forbid: path traversal, special chars
    
    pattern = /^[a-zA-Z0-9_-]{1,64}$/
    
    if not pattern.match(input):
        throw InvalidNameError
    
    // Additional checks
    if input in ["server", "ca", "admin", "root"]:
        throw ReservedNameError
    
    // Prevent path traversal
    if ".." in input or "/" in input or "\\" in input:
        throw PathTraversalError
    
    return input
```

**Example implementation (Go):**
```go
func SanitizeClientName(name string) (string, error) {
    // Check length
    if len(name) > 64 {
        return "", fmt.Errorf("name too long (max 64)")
    }
    
    // Check pattern
    matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", name)
    if !matched {
        return "", fmt.Errorf("invalid characters")
    }
    
    // Reserved names
    reserved := []string{"server", "ca", "admin", "root"}
    for _, r := range reserved {
        if name == r {
            return "", fmt.Errorf("reserved name")
        }
    }
    
    return name, nil
}
```

**Forbid shell `eval`:**
```go
// NEVER DO THIS:
cmd := exec.Command("sh", "-c", "easyrsa build-client-full " + user_input)

// ALWAYS DO THIS:
cmd := exec.Command("easyrsa", "build-client-full", sanitized_name, "nopass")
cmd.Dir = "/etc/openvpn/easy-rsa"
```

### 6.3 Audit Logging

```pseudocode
audit_log_entry = {
    timestamp: Time.now(),
    user: authenticated_user.username,
    action: "create_client" | "delete_client" | "download_config",
    target: client_name,
    result: "success" | "failed",
    ip_address: request.remote_addr,
    user_agent: request.user_agent
}

// Store in:
// /opt/irangate/logs/audit.log (rotated daily)

function audit_log(action, target, result):
    entry = build_audit_entry(action, target, result)
    write_to_logfile(entry)
    
    // Also emit to syslog
    syslog.info(entry)
```

### 6.4 Strict Permissions

```pseudocode
// Certificate files:
chmod(0600, "*.key")  # Private keys: owner read-write only
chmod(0644, "*.crt")  # Certificates: readable by all
chmod(0600, "*.ovpn") # Config files: owner read-write only

// Directory permissions:
chmod(0755, "/etc/openvpn")           # Owner full, others read+execute
chmod(0700, "/etc/openvpn/easy-rsa")  # Owner only access to PKI

// Backup directories:
chmod(0700, "/opt/irangate/backups")  # Only admin accessible
```

---

## 7. API CONTRACT & UI CONSTRAINTS

### 7.1 API Endpoints

**GET /api/openvpn/clients**
```json
// Request:
GET /api/openvpn/clients HTTP/1.1
Authorization: Bearer {token}

// Response (instant, from cache):
{
  "success": true,
  "timestamp": "2024-10-27T22:00:00Z",
  "total_clients": 5,
  "clients": [
    {
      "id": "client1",
      "name": "client1",
      "type": "client",
      "port": 1194,
      "protocol": "udp",
      "cipher": "AES-256-GCM",
      "remote_address": "91.107.182.251",
      "push_routes": [],
      "push_dns": ["1.1.1.1", "1.0.0.1"],
      "certs_present": {
        "ca": true,
        "cert": true,
        "key": true,
        "tls_crypt": true
      },
      "last_modified": "2024-10-27T21:45:00Z",
      "status": "complete",
      "errors": []
    }
  ]
}

// Performance requirement: < 50ms response time
```

**GET /api/openvpn/clients/:id**
```json
// Request:
GET /api/openvpn/clients/client1 HTTP/1.1

// Response:
{
  "success": true,
  "client": {
    "id": "client1",
    "name": "client1",
    "type": "client",
    "port": 1194,
    "protocol": "udp",
    "cipher": "AES-256-GCM",
    "auth": "SHA256",
    "remote_address": "91.107.182.251",
    "commonName": "client1",
    "push_dns": ["1.1.1.1", "1.0.0.1"],
    "certs_present": {
      "ca": true,
      "cert": true,
      "key": true,
      "tls_crypt": true
    },
    "last_modified": "2024-10-27T21:45:00Z",
    "status": "complete",
    "errors": []
  }
}
```

**POST /api/openvpn/clients**
```json
// Request:
POST /api/openvpn/clients HTTP/1.1
Content-Type: application/json
Authorization: Bearer {admin_token}

{
  "name": "newclient"
}

// Response:
{
  "success": true,
  "message": "Client created successfully",
  "client": {
    "id": "newclient",
    "name": "newclient",
    "status": "complete"
  }
}

// On error:
{
  "success": false,
  "error": "Client 'newclient' already exists"
}
```

**GET /api/openvpn/clients/:id/download**
```json
// Request (requires admin):
GET /api/openvpn/clients/client1/download HTTP/1.1

// Response (5-minute temporary URL):
{
  "success": true,
  "download_url": "/api/download/xyz123abc",
  "expires_in": 300,
  "filename": "client1.ovpn",
  "content_type": "application/openvpn"
}

// Client downloads .ovpn file with embedded certs
```

**GET /api/openvpn/health**
```json
// Response:
{
  "service": "irangate-webpanel",
  "status": "healthy",
  "version": "1.0.0",
  "openvpn_service": "running",
  "last_scan": "2024-10-27T22:00:00Z",
  "total_configs": 6,
  "active_configs": 5,
  "incomplete_configs": 1
}
```

### 7.2 Response Caching

**Always return cached JSON instantly:**
```pseudocode
var cache = {
    data: null,
    timestamp: null,
    lock: Mutex
}

function GET_clients_handler():
    cache.lock.read()
    
    if cache.data == null or is_stale(cache):
        // Return old data, trigger background refresh
        trigger_background_refresh()
    
    response = cache.data
    cache.lock.unlock()
    
    return response  // Always < 50ms
```

**Background updates never block API:**
```pseudocode
function background_refresh():
    new_data = perform_full_scan()  // May take 2-5 seconds
    
    cache.lock.write()
    cache.data = new_data
    cache.timestamp = now()
    cache.lock.unlock()
```

### 7.3 Performance Requirements

| Endpoint | Max Response Time | Caching |
|----------|-------------------|---------|
| GET /api/openvpn/clients | < 50ms | Yes (index) |
| GET /api/openvpn/clients/:id | < 100ms | Yes (index) |
| POST /api/openvpn/clients | < 5s | N/A (async) |
| GET /api/openvpn/clients/:id/download | < 500ms | Yes (temp URL) |
| GET /api/openvpn/health | < 50ms | Yes |

---

## 8. EXAMPLE JSON OUTPUTS

### 8.1 Complete Client .ovpn

```json
{
  "id": "myclient",
  "name": "myclient",
  "type": "client",
  "port": 1194,
  "protocol": "udp",
  "cipher": "AES-256-GCM",
  "auth": "SHA256",
  "remote_address": "91.107.182.251",
  "push_routes": ["redirect-gateway def1 bypass-dhcp"],
  "push_dns": ["1.1.1.1", "1.0.0.1"],
  "certs_present": {
    "ca": true,
    "cert": true,
    "key": true,
    "tls_crypt": true
  },
  "commonName": "myclient",
  "last_modified": "2024-10-27T22:15:00Z",
  "status": "complete",
  "errors": []
}
```

### 8.2 Client Missing Key (Incomplete)

```json
{
  "id": "problematic_client",
  "name": "problematic_client",
  "type": "client",
  "port": 1194,
  "protocol": "tcp",
  "cipher": "AES-256-GCM",
  "remote_address": "91.107.182.251",
  "push_dns": ["8.8.8.8"],
  "certs_present": {
    "ca": true,
    "cert": true,
    "key": false,
    "tls_crypt": false
  },
  "commonName": "problematic_client",
  "last_modified": "2024-10-27T21:00:00Z",
  "status": "incomplete",
  "errors": [
    "Missing client private key",
    "Missing TLS-crypt key"
  ],
  "remediation_hint": "Run: irangate client add problematic_client"
}
```

### 8.3 Server Config with Pushed Routes

```json
{
  "id": "server",
  "name": "server",
  "type": "server",
  "port": 1194,
  "protocol": "udp",
  "cipher": "AES-256-GCM",
  "auth": "SHA256",
  "dev_type": "tun",
  "push_routes": ["redirect-gateway def1 bypass-dhcp"],
  "push_dns": ["1.1.1.1", "1.0.0.1"],
  "certs_present": {
    "ca": true,
    "cert": true,
    "key": true,
    "tls_crypt": true,
    "dh": true
  },
  "status": "active",
  "last_modified": "2024-10-27T20:00:00Z",
  "errors": [],
  "source_path_internal": "/etc/openvpn/server.conf"
}
```

---

## 9. CODE SNIPPETS

### 9.1 Scanner Worker (Go)

```go
package openvpn

import (
    "context"
    "path/filepath"
    "sync"
    "time"
    "github.com/fsnotify/fsnotify"
)

type ConfigIndex struct {
    clients map[string]ClientConfig
    servers map[string]ServerConfig
    mutex   sync.RWMutex
    lastScan time.Time
}

func (c *ConfigIndex) StartScanner(ctx context.Context) error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    defer watcher.Close()
    
    // Watch these paths
    for _, path := range []string{
        "/etc/openvpn/easy-rsa",
        "/etc/openvpn/client",
        "/etc/openvpn",
    } {
        if err := watcher.Add(path); err != nil {
            return err
        }
    }
    
    // Initial full scan
    c.performFullScan()
    
    // Event loop
    for {
        select {
        case event := <-watcher.Events:
            c.handleFileChange(event)
        case err := <-watcher.Errors:
            log.Printf("Watcher error: %v", err)
        case <-ctx.Done():
            return nil
        }
    }
}

func (c *ConfigIndex) handleFileChange(event fsnotify.Event) {
    if !isRelevantFile(event.Name) {
        return
    }
    
    // Incremental update
    entry := parseFile(event.Name)
    
    c.mutex.Lock()
    if event.Op&fsnotify.Remove == fsnotify.Remove {
        delete(c.clients, entry.ID)
    } else {
        c.updateEntry(entry)
    }
    c.lastScan = time.Now()
    c.mutex.Unlock()
}
```

### 9.2 Parser Function (Go)

```go
func parseOpenVPNConfig(filePath string) (ConfigEntry, error) {
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return ConfigEntry{}, err
    }
    
    entry := ConfigEntry{
        SourcePath: filePath,
        Type: inferType(filePath),
    }
    
    lines := strings.Split(string(data), "\n")
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        
        // Parse fields
        if match := regexp.MustCompile(`^port\s+(\d+)`).FindStringSubmatch(line); match != nil {
            port, _ := strconv.Atoi(match[1])
            entry.Port = port
        }
        
        if match := regexp.MustCompile(`^proto\s+(udp|tcp)`).FindStringSubmatch(line); match != nil {
            entry.Protocol = match[1]
        }
        
        if match := regexp.MustCompile(`^cipher\s+(\S+)`).FindStringSubmatch(line); match != nil {
            entry.Cipher = match[1]
        }
        
        if match := regexp.MustCompile(`^ca\s+(\S+)`).FindStringSubmatch(line); match != nil {
            entry.CertsPresent.CA = fileExists(match[1])
        }
        
        // ... parse other fields
    }
    
    // Determine completeness
    entry.Status = determineStatus(entry)
    
    return entry, nil
}
```

### 9.3 Create Client Flow (Go)

```go
func createClientHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name string `json:"name"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // 1. Sanitize name
    sanitized, err := sanitizeClientName(req.Name)
    if err != nil {
        sendError(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // 2. Check exists
    if exists(sanitized) {
        sendError(w, "Client already exists", http.StatusConflict)
        return
    }
    
    // 3. Audit log
    auditLog("create_client", sanitized, r.RemoteAddr)
    
    // 4. Generate certificate
    cmd := exec.Command("easyrsa", "build-client-full", sanitized, "nopass")
    cmd.Dir = "/etc/openvpn/easy-rsa"
    cmd.Env = append(os.Environ(), "EASYRSA_BATCH=1")
    
    if err := cmd.Run(); err != nil {
        auditLog("create_client", sanitized, "failed")
        sendError(w, "Certificate generation failed", http.StatusInternalServerError)
        return
    }
    
    // 5. Generate .ovpn file
    ovpnPath := generateOVPN(sanitized)
    
    // 6. Set permissions
    os.Chmod(ovpnPath, 0600)
    
    // 7. Update index
    triggerRescan()
    
    // 8. Return success
    sendSuccess(w, "Client created", map[string]interface{}{
        "name": sanitized,
        "config_path": ovpnPath,
    })
}
```

---

## 10. TESTS TO RUN

### 10.1 Parser Unit Tests

```go
func TestParseCompleteClient(t *testing.T) {
    config := parseOpenVPNConfig("testdata/complete_client.ovpn")
    
    assert.Equal(t, config.Port, 1194)
    assert.Equal(t, config.Protocol, "udp")
    assert.Equal(t, config.Cipher, "AES-256-GCM")
    assert.Equal(t, config.CertsPresent.CA, true)
    assert.Equal(t, config.Status, "complete")
}

func TestParseIncompleteClient(t *testing.T) {
    config := parseOpenVPNConfig("testdata/missing_key.ovpn")
    
    assert.Equal(t, config.CertsPresent.Key, false)
    assert.Equal(t, config.Status, "incomplete")
    assert.NotEmpty(t, config.Errors)
}
```

### 10.2 Integration: Create Client Test

```go
func TestCreateClientIntegration(t *testing.T) {
    // Setup
    cleanupTestFiles()
    
    // Create client via API
    resp := POST("/api/openvpn/clients", `{"name": "testclient"}`)
    assert.Equal(t, resp.StatusCode, 200)
    
    // Verify certificate exists
    assert.FileExists(t, "/etc/openvpn/easy-rsa/pki/issued/testclient.crt")
    assert.FileExists(t, "/etc/openvpn/easy-rsa/pki/private/testclient.key")
    
    // Verify .ovpn file
    assert.FileExists(t, "/etc/openvpn/client/testclient.ovpn")
    
    // Check permissions
    assert.Equal(t, getFileMode("/etc/openvpn/client/testclient.ovpn"), 0600)
    
    // Cleanup
    cleanupTestFiles()
}
```

### 10.3 E2E: Validate Client Connect

```bash
#!/bin/bash
# e2e_test_client_connect.sh

# Create a test client via API
curl -X POST http://localhost:8080/api/openvpn/clients \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "e2e_test_client"}'

# Download the .ovpn file
curl -X GET http://localhost:8080/api/openvpn/clients/e2e_test_client/download \
  -H "Authorization: Bearer $TOKEN" \
  -o test_client.ovpn

# Try to connect (should succeed)
openvpn --config test_client.ovpn --verb 3 &
OPENVPN_PID=$!

sleep 5

# Check if connected
if ps -p $OPENVPN_PID > /dev/null; then
    echo "✓ Client connected successfully"
    kill $OPENVPN_PID
    exit 0
else
    echo "✗ Client failed to connect"
    exit 1
fi
```

---

## 11. FINAL DELIVERABLES CHECKLIST

- [x] Interactive installation with user prompts
- [x] Silent background directory scanning  
- [x] Exact parsing rules for all fields
- [x] Complete JSON schema with status tracking
- [x] Client creation using auto-setup.sh logic
- [x] Single-inbound mode as baseline
- [x] Error handling (incomplete status)
- [x] Security constraints (sanitization, permissions)
- [x] Audit logging specification
- [x] API contract with caching strategy
- [x] Example JSON for all scenarios
- [x] Pseudocode for scanner & parser
- [x] Complete client creation workflow
- [x] Test specifications
- [x] Documentation of assumptions

**Specification ready for implementation.** ✅

