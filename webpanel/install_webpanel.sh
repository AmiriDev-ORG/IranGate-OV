#!/usr/bin/env bash
set -euo pipefail
VERBOSE=false
DRY_RUN=false
LOG_FILE="/var/log/irangate-webpanel-install.log"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
log() {
  local level="$1"
  shift
  local message="$*"
  local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
  case "$level" in
    "INFO") echo -e "${BLUE}[INFO]${NC} $message" ;;
    "SUCCESS") echo -e "${GREEN}[SUCCESS]${NC} $message" ;;
    "WARNING") echo -e "${YELLOW}[WARNING]${NC} $message" ;;
    "ERROR") echo -e "${RED}[ERROR]${NC} $message" ;;
  esac
  if [ "$DRY_RUN" = false ]; then
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
  fi
}
vlog() {
  if [ "$VERBOSE" = true ]; then
    log "INFO" "$@"
  fi
}
prompt() {
  local var_name="$1"; shift
  local prompt_text="$1"; shift
  local default_value="${1-}"
  local value
  if [ -n "$default_value" ]; then
    read -rp "$prompt_text [$default_value]: " value || true
    value=${value:-$default_value}
  else
    read -rp "$prompt_text: " value || true
  fi
  printf -v "$var_name" '%s' "$value"
}
validate_port() {
  local port="$1"
  if ! [[ "$port" =~ ^[0-9]+$ ]] || [ "$port" -lt 1 ] || [ "$port" -gt 65535 ]; then
    log "ERROR" "Invalid port number: $port. Must be between 1 and 65535."
    return 1
  fi
  return 0
}
validate_webpath() {
  local path="$1"
  if [[ "$path" =~ [^a-zA-Z0-9_-] ]]; then
    log "ERROR" "Invalid web path: $path. Only alphanumeric characters, hyphens, and underscores are allowed."
    return 1
  fi
  return 0
}
validate_username() {
  local username="$1"
  if [ ${#username} -lt 3 ]; then
    log "ERROR" "Username must be at least 3 characters long."
    return 1
  fi
  if [[ "$username" =~ [^a-zA-Z0-9_-] ]]; then
    log "ERROR" "Username contains invalid characters. Only alphanumeric characters, hyphens, and underscores are allowed."
    return 1
  fi
  return 0
}
validate_password() {
  local password="$1"
  if [ ${#password} -lt 6 ]; then
    log "ERROR" "Password must be at least 6 characters long."
    return 1
  fi
  return 0
}
check_port_available() {
  local port="$1"
  if ss -ltn | awk '{print $4}' | grep -q ":$port$"; then
    log "ERROR" "Port $port is already in use. Choose another port."
    return 1
  fi
  return 0
}
check_service_exists() {
  if systemctl list-unit-files | grep -q "irangate-webpanel.service"; then
    log "WARNING" "irangate-webpanel service already exists."
    echo -n "Do you want to reinstall? (yes/no): "
    read -r response
    if [[ "$response" != "yes" ]]; then
      log "INFO" "Installation cancelled."
      exit 0
    fi
  fi
}
require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    log "ERROR" "$1 is required but not installed."
    exit 1
  }
}
while [[ $# -gt 0 ]]; do
  case $1 in
    -v|--verbose)
      VERBOSE=true
      shift
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help)
      echo "Usage: $0 [OPTIONS]"
      echo "Options:"
      echo "  -v, --verbose    Enable verbose output"
      echo "  --dry-run        Show what would be done without making changes"
      echo "  -h, --help       Show this help message"
      exit 0
      ;;
    *)
      log "ERROR" "Unknown option: $1"
      exit 1
      ;;
  esac
done
require_cmd systemctl
require_cmd bash
require_cmd ss
log "INFO" "=== IranGate Web Panel Installer ==="
check_service_exists
log "INFO" "Gathering installation parameters..."
while true; do
  prompt ADMIN_USER "Enter admin username" "admin"
  if validate_username "$ADMIN_USER"; then
    break
  fi
done
while true; do
  prompt ADMIN_PASS "Enter admin password" "admin123"
  if validate_password "$ADMIN_PASS"; then
    prompt ADMIN_PASS2 "Re-enter admin password" "admin123"
    if [ "$ADMIN_PASS" = "$ADMIN_PASS2" ]; then
      break
    else
      log "ERROR" "Passwords do not match. Try again."
    fi
  fi
done
while true; do
  prompt WEB_PATH "Enter web path (no leading slash)" "example"
  WEB_PATH=${WEB_PATH#/}
  if validate_webpath "$WEB_PATH"; then
    break
  fi
done
while true; do
  prompt PORT "Enter port for web panel (1-65535)" ""
  if [[ -z "$PORT" ]]; then
    log "ERROR" "Port is required."
    continue
  fi
  if validate_port "$PORT" && check_port_available "$PORT"; then
    break
  fi
done
if [ -z "$WEB_PATH" ]; then
  BASEPATH="/"
else
  BASEPATH="/${WEB_PATH}"
fi
SERVICE_NAME="irangate-webpanel"
ENV_DIR="/etc/irangate"
ENV_FILE="$ENV_DIR/webpanel.env"
RUN_USER="root"
APP_DIR="/opt/irangate/webpanel"
BACKEND_DIR="$APP_DIR/backend"
STATIC_DIR="$APP_DIR/frontend"
log "INFO" "Installation Summary:"
log "INFO" "  Admin Username: $ADMIN_USER"
log "INFO" "  Web Path: $WEB_PATH"
log "INFO" "  Port: $PORT"
log "INFO" "  Base Path: $BASEPATH"
log "INFO" "  Static Directory: $STATIC_DIR"
if [ "$DRY_RUN" = true ]; then
  log "INFO" "DRY RUN MODE - No changes will be made"
  exit 0
fi
sudo mkdir -p "$(dirname "$LOG_FILE")"
sudo touch "$LOG_FILE"
sudo chown root:root "$LOG_FILE"
log "INFO" "Starting installation..."
log "INFO" "Creating directories..."
vlog "Creating $ENV_DIR, $APP_DIR, $STATIC_DIR"
sudo mkdir -p "$ENV_DIR" "$APP_DIR" "$STATIC_DIR"
log "INFO" "Building backend..."
if [ -f "$(dirname "$0")/backend/api.go" ]; then
  pushd "$(dirname "$0")/backend" >/dev/null
  require_cmd go
  vlog "Building Go application..."
  if go build -o /usr/local/bin/irangate-webpanel; then
    log "SUCCESS" "Backend built successfully"
  else
    log "ERROR" "Failed to build backend"
    exit 1
  fi
  popd >/dev/null
elif command -v irangate-webpanel >/dev/null 2>&1; then
  log "INFO" "Using existing irangate-webpanel binary"
else
  log "ERROR" "Backend source not found and binary not present"
  exit 1
fi
log "INFO" "Copying frontend files..."
if [ -d "$(dirname "$0")/frontend" ]; then
  vlog "Copying frontend files from $(dirname "$0")/frontend to $STATIC_DIR"
  if sudo cp -r "$(dirname "$0")/frontend"/* "$STATIC_DIR/"; then
    log "SUCCESS" "Frontend files copied to $STATIC_DIR"
  else
    log "ERROR" "Failed to copy frontend files"
    exit 1
  fi
else
  log "ERROR" "Frontend directory not found at $(dirname "$0")/frontend"
  exit 1
fi
log "INFO" "Writing environment file at $ENV_FILE"
if sudo tee "$ENV_FILE" >/dev/null <<EOF
WEBPANEL_PORT=$PORT
WEBPANEL_BASEPATH=$BASEPATH
WEBPANEL_STATIC_DIR=$STATIC_DIR
WEBPANEL_ADMIN_USER=$ADMIN_USER
WEBPANEL_ADMIN_PASS=$ADMIN_PASS
EOF
then
  log "SUCCESS" "Environment file created at $ENV_FILE"
else
  log "ERROR" "Failed to create environment file"
  exit 1
fi
sudo chmod 600 "$ENV_FILE"
log "INFO" "Installing systemd service..."
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
if sudo tee "$SERVICE_FILE" >/dev/null <<'UNIT'
[Unit]
Description=IranGate Web Panel
After=network-online.target
Wants=network-online.target
[Service]
Type=simple
EnvironmentFile=/etc/irangate/webpanel.env
ExecStart=/usr/local/bin/irangate-webpanel
User=root
Restart=on-failure
RestartSec=3
[Install]
WantedBy=multi-user.target
UNIT
then
  log "SUCCESS" "Systemd service file created at $SERVICE_FILE"
else
  log "ERROR" "Failed to create systemd service file"
  exit 1
fi
log "INFO" "Reloading systemd and starting service..."
if sudo systemctl daemon-reload; then
  log "SUCCESS" "Systemd daemon reloaded"
else
  log "ERROR" "Failed to reload systemd daemon"
  exit 1
fi
if sudo systemctl enable --now "$SERVICE_NAME"; then
  log "SUCCESS" "Service enabled and started"
else
  log "ERROR" "Failed to enable or start service"
  exit 1
fi
IP=$(hostname -I 2>/dev/null | awk '{print $1}')
IP=${IP:-127.0.0.1}
URL="http://${IP}:${PORT}"
if [ "$BASEPATH" != "/" ]; then
  URL="${URL}${BASEPATH}/"
else
  URL="${URL}/"
fi
log "SUCCESS" "Installation completed successfully!"
echo
log "INFO" "Web Panel URL: $URL"
log "INFO" "Admin Username: $ADMIN_USER"
if [ "$ADMIN_PASS" = "admin123" ]; then
  log "WARNING" "Please change the default password after login!"
fi
log "INFO" "Service Status: $(systemctl is-active $SERVICE_NAME)"
log "INFO" "Installation log: $LOG_FILE"
echo