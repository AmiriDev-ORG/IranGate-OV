#!/usr/bin/env python3
"""
IRANGATE AI Management Agent
Automated system monitoring and management with OpenAI integration
"""

import os
import sys
import json
import time
import psutil
import logging
import subprocess
import platform
from datetime import datetime
from typing import Dict, List, Optional
import requests
from openai import OpenAI

# Configure logging with cross-platform paths
def get_log_path():
    """Get appropriate log path based on platform"""
    if platform.system() == "Windows":
        return os.path.join(os.environ.get('TEMP', '/tmp'), 'irangate_ai_agent.log')
    else:
        return '/var/log/irangate/ai_agent.log'

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(get_log_path()),
        logging.StreamHandler()
    ]
)

logger = logging.getLogger(__name__)


class TelegramBot:
    """Simple Telegram bot for notifications"""
    
    def __init__(self, token: str):
        self.token = token
        self.base_url = f"https://api.telegram.org/bot{token}"
    
    def send_message(self, chat_id: str, text: str) -> bool:
        """Send message to Telegram chat"""
        try:
            url = f"{self.base_url}/sendMessage"
            data = {
                "chat_id": chat_id,
                "text": text,
                "parse_mode": "Markdown"
            }
            
            response = requests.post(url, data=data, timeout=10)
            return response.status_code == 200
            
        except Exception as e:
            logger.error(f"Error sending Telegram message: {e}")
            return False


class IRANGATEAIAgent:
    """AI-powered management agent for IRANGATE VPN system"""
    
    def __init__(self, config_file: str = None):
        # Set default config file based on platform
        if config_file is None:
            if platform.system() == "Windows":
                config_file = os.path.join(os.environ.get('APPDATA', ''), 'irangate', 'ai_config.json')
            else:
                config_file = "/etc/irangate/ai_config.json"
        
        self.config_file = config_file
        self.config = self.load_config()
        self.openai_client = None
        self.telegram_bot = None
        self.running = False
        
        # Initialize OpenAI client if API key is provided
        if self.config.get('openai_api_key'):
            try:
                self.openai_client = OpenAI(api_key=self.config['openai_api_key'])
            except Exception as e:
                logger.error(f"Failed to initialize OpenAI client: {e}")
        
        # Initialize Telegram bot if token is provided
        if self.config.get('telegram_bot_token'):
            try:
                self.telegram_bot = TelegramBot(self.config['telegram_bot_token'])
            except Exception as e:
                logger.error(f"Failed to initialize Telegram bot: {e}")
    
    def load_config(self) -> Dict:
        """Load AI configuration from JSON file"""
        try:
            with open(self.config_file, 'r') as f:
                config = json.load(f)
                return self.validate_config(config)
        except FileNotFoundError:
            logger.warning(f"Config file {self.config_file} not found, using defaults")
            return self.get_default_config()
        except json.JSONDecodeError as e:
            logger.error(f"Error parsing config file: {e}")
            return self.get_default_config()
    
    def validate_config(self, config: Dict) -> Dict:
        """Validate configuration parameters"""
        default_config = self.get_default_config()
        
        # Validate required parameters
        for key, default_value in default_config.items():
            if key not in config:
                logger.warning(f"Missing config parameter: {key}, using default: {default_value}")
                config[key] = default_value
        
        # Validate numeric parameters
        numeric_params = ['monitoring_interval', 'cpu_threshold', 'ram_threshold', 'disk_threshold', 'max_failed_connections']
        for param in numeric_params:
            if not isinstance(config.get(param), (int, float)) or config[param] <= 0:
                logger.warning(f"Invalid {param}: {config.get(param)}, using default: {default_config[param]}")
                config[param] = default_config[param]
        
        # Validate boolean parameters
        boolean_params = ['auto_fix_enabled', 'telegram_notifications']
        for param in boolean_params:
            if not isinstance(config.get(param), bool):
                logger.warning(f"Invalid {param}: {config.get(param)}, using default: {default_config[param]}")
                config[param] = default_config[param]
        
        # Validate thresholds
        if config['cpu_threshold'] > 100:
            logger.warning(f"CPU threshold too high: {config['cpu_threshold']}, setting to 90")
            config['cpu_threshold'] = 90
        
        if config['ram_threshold'] > 100:
            logger.warning(f"RAM threshold too high: {config['ram_threshold']}, setting to 95")
            config['ram_threshold'] = 95
        
        return config
    
    def get_default_config(self) -> Dict:
        """Get default configuration"""
        return {
            "monitoring_interval": 60,  # seconds
            "cpu_threshold": 90,        # percentage
            "ram_threshold": 95,        # percentage
            "disk_threshold": 10,       # percentage free
            "max_failed_connections": 100,
            "auto_fix_enabled": True,
            "telegram_notifications": True,
            "openai_api_key": "",
            "telegram_bot_token": "",
            "admin_chat_id": "",
            "log_file": get_log_path()
        }
    
    def get_system_metrics(self) -> Dict:
        """Get current system metrics"""
        try:
            # CPU usage
            cpu_percent = psutil.cpu_percent(interval=1)
            
            # Memory usage
            memory = psutil.virtual_memory()
            
            # Disk usage
            disk = psutil.disk_usage('/')
            
            # Network statistics
            network = psutil.net_io_counters()
            
            # OpenVPN process status
            openvpn_running = self.check_openvpn_status()
            
            # Failed connections count
            failed_connections = self.get_failed_connections_count()
            
            return {
                "timestamp": datetime.now().isoformat(),
                "cpu_percent": cpu_percent,
                "ram_percent": memory.percent,
                "ram_available": memory.available,
                "disk_percent": (disk.used / disk.total) * 100,
                "disk_free": disk.free,
                "disk_total": disk.total,
                "network_bytes_sent": network.bytes_sent,
                "network_bytes_recv": network.bytes_recv,
                "openvpn_running": openvpn_running,
                "failed_connections": failed_connections
            }
        except Exception as e:
            logger.error(f"Error getting system metrics: {e}")
            return {}
    
    def check_openvpn_status(self) -> bool:
        """Check if OpenVPN service is running"""
        try:
            if platform.system() == "Windows":
                # On Windows, check OpenVPN service via sc command
                try:
                    result = subprocess.run(
                        ["sc", "query", "OpenVPNService"],
                        capture_output=True,
                        text=True,
                        timeout=10
                    )
                    return "RUNNING" in result.stdout
                except (subprocess.TimeoutExpired, subprocess.CalledProcessError):
                    # Fallback: check for OpenVPN process
                    for proc in psutil.process_iter(['pid', 'name']):
                        try:
                            if 'openvpn' in proc.info['name'].lower():
                                return True
                        except (psutil.NoSuchProcess, psutil.AccessDenied):
                            continue
                    return False
            else:
                # On Linux, use systemctl
                try:
                    result = subprocess.run(
                        ["systemctl", "is-active", "openvpn"],
                        capture_output=True,
                        text=True,
                        timeout=10
                    )
                    return result.stdout.strip() == "active"
                except (subprocess.TimeoutExpired, subprocess.CalledProcessError):
                    # Fallback: check for openvpn process
                    for proc in psutil.process_iter(['pid', 'name']):
                        try:
                            if 'openvpn' in proc.info['name'].lower():
                                return True
                        except (psutil.NoSuchProcess, psutil.AccessDenied):
                            continue
                    return False
        except Exception as e:
            logger.error(f"Error checking OpenVPN status: {e}")
            return False
    
    def get_failed_connections_count(self) -> int:
        """Get count of failed connections from logs"""
        try:
            # Get appropriate log file path based on platform
            if platform.system() == "Windows":
                # Windows OpenVPN log paths
                possible_paths = [
                    os.path.join(os.environ.get('PROGRAMDATA', ''), 'OpenVPN', 'log', 'openvpn.log'),
                    os.path.join(os.environ.get('APPDATA', ''), 'OpenVPN', 'log', 'openvpn.log'),
                    os.path.join(os.environ.get('TEMP', ''), 'openvpn.log')
                ]
                log_file = None
                for path in possible_paths:
                    if os.path.exists(path):
                        log_file = path
                        break
                if not log_file:
                    return 0
            else:
                log_file = "/var/log/openvpn/openvpn.log"
                if not os.path.exists(log_file):
                    return 0
            
            failed_count = 0
            with open(log_file, 'r', encoding='utf-8', errors='ignore') as f:
                lines = f.readlines()
                # Check last 100 lines for failed connections
                for line in lines[-100:]:
                    if any(error in line for error in ["TLS Error", "Connection reset", "AUTH_FAILED", "TLS handshake failed"]):
                        failed_count += 1
            
            return failed_count
        except Exception as e:
            logger.error(f"Error counting failed connections: {e}")
            return 0
    
    def analyze_with_ai(self, metrics: Dict) -> Dict:
        """Use OpenAI to analyze system metrics and suggest actions"""
        if not self.openai_client:
            return {"action": "none", "confidence": 0, "reason": "OpenAI not configured"}
        
        try:
            prompt = f"""
            Analyze the following IRANGATE VPN server metrics and suggest the best action:
            
            Metrics:
            - CPU Usage: {metrics.get('cpu_percent', 0)}%
            - RAM Usage: {metrics.get('ram_percent', 0)}%
            - Disk Free: {metrics.get('disk_free', 0)} bytes
            - OpenVPN Running: {metrics.get('openvpn_running', False)}
            - Failed Connections: {metrics.get('failed_connections', 0)}
            
            Possible actions:
            1. restart_openvpn - Restart OpenVPN service
            2. cleanup_logs - Clean up log files
            3. check_certificates - Check certificate validity
            4. restart_system - Restart the entire system
            5. none - No action needed
            
            Respond in JSON format with: action, confidence (0-100), reason
            """
            
            response = self.openai_client.chat.completions.create(
                model="gpt-3.5-turbo",
                messages=[{"role": "user", "content": prompt}],
                temperature=0.1
            )
            
            try:
                result = json.loads(response.choices[0].message.content)
                return result
            except json.JSONDecodeError as e:
                logger.error(f"Failed to parse AI response as JSON: {e}")
                logger.error(f"Raw response: {response.choices[0].message.content}")
                return {"action": "none", "confidence": 0, "reason": "AI response parsing failed"}
            
        except Exception as e:
            logger.error(f"Error analyzing with AI: {e}")
            return {"action": "none", "confidence": 0, "reason": f"AI analysis failed: {e}"}
    
    def execute_action(self, action: str, reason: str) -> bool:
        """Execute the suggested action"""
        try:
            if action == "restart_openvpn":
                logger.info("Restarting OpenVPN service...")
                if platform.system() == "Windows":
                    # On Windows, try to restart OpenVPN service
                    try:
                        subprocess.run(["net", "stop", "OpenVPNService"], check=True, timeout=30)
                        subprocess.run(["net", "start", "OpenVPNService"], check=True, timeout=30)
                    except subprocess.CalledProcessError:
                        logger.warning("Failed to restart OpenVPN service via net command")
                        return False
                else:
                    # On Linux, use systemctl
                    subprocess.run(["systemctl", "restart", "openvpn"], check=True)
                self.send_notification(f"✅ OpenVPN service restarted successfully\nReason: {reason}")
                return True
                
            elif action == "cleanup_logs":
                logger.info("Cleaning up log files...")
                if platform.system() == "Windows":
                    # On Windows, clean temp files older than 7 days
                    temp_dir = os.environ.get('TEMP', '/tmp')
                    for root, dirs, files in os.walk(temp_dir):
                        for file in files:
                            if file.endswith('.log'):
                                file_path = os.path.join(root, file)
                                try:
                                    if os.path.getmtime(file_path) < time.time() - (7 * 24 * 60 * 60):
                                        os.remove(file_path)
                                except (OSError, IOError):
                                    pass  # Skip files that can't be deleted
                else:
                    # On Linux, use find command
                    subprocess.run(["find", "/var/log", "-name", "*.log", "-mtime", "+7", "-delete"], check=True)
                self.send_notification(f"✅ Log files cleaned up successfully\nReason: {reason}")
                return True
                
            elif action == "check_certificates":
                logger.info("Checking certificate validity...")
                # Get appropriate certificate directory based on platform
                if platform.system() == "Windows":
                    # Windows OpenVPN certificate paths
                    possible_dirs = [
                        os.path.join(os.environ.get('PROGRAMDATA', ''), 'OpenVPN', 'config'),
                        os.path.join(os.environ.get('APPDATA', ''), 'OpenVPN', 'config'),
                        os.path.join(os.environ.get('PROGRAMFILES', ''), 'OpenVPN', 'config')
                    ]
                    cert_dir = None
                    for directory in possible_dirs:
                        if os.path.exists(directory):
                            cert_dir = directory
                            break
                else:
                    cert_dir = "/etc/openvpn/easy-rsa/pki"
                
                if not cert_dir or not os.path.exists(cert_dir):
                    logger.warning(f"Certificate directory not found: {cert_dir}")
                    return False
                
                checked_certs = 0
                try:
                    for file in os.listdir(cert_dir):
                        if file.endswith('.crt'):
                            # Check if certificate is expired
                            cert_path = os.path.join(cert_dir, file)
                            try:
                                result = subprocess.run(
                                    ["openssl", "x509", "-in", cert_path, "-noout", "-dates"],
                                    capture_output=True, text=True, timeout=10
                                )
                                if result.returncode == 0:
                                    checked_certs += 1
                                    logger.info(f"Certificate {file} checked successfully")
                                else:
                                    logger.warning(f"Failed to check certificate: {file}")
                            except (subprocess.TimeoutExpired, subprocess.CalledProcessError) as e:
                                logger.warning(f"Failed to check certificate {file}: {e}")
                                continue
                except OSError as e:
                    logger.error(f"Error accessing certificate directory: {e}")
                    return False
                
                self.send_notification(f"✅ Certificate check completed - {checked_certs} certificates checked\nReason: {reason}")
                return True
                
            elif action == "restart_system":
                logger.warning("System restart requested - this requires manual confirmation")
                self.send_notification(f"⚠️ System restart requested\nReason: {reason}\nPlease confirm manually!")
                return False
                
            else:
                logger.info(f"No action taken: {action}")
                return True
                
        except subprocess.CalledProcessError as e:
            logger.error(f"Error executing action {action}: {e}")
            self.send_notification(f"❌ Failed to execute action: {action}\nError: {e}")
            return False
        except Exception as e:
            logger.error(f"Unexpected error executing action {action}: {e}")
            return False
    
    def send_notification(self, message: str):
        """Send notification via Telegram"""
        if not self.telegram_bot or not self.config.get('telegram_notifications'):
            logger.info(f"Notification: {message}")
            return
        
        try:
            self.telegram_bot.send_message(
                chat_id=self.config['admin_chat_id'],
                text=f"🤖 IRANGATE AI Agent\n\n{message}"
            )
        except Exception as e:
            logger.error(f"Error sending Telegram notification: {e}")
    
    def monitor_loop(self):
        """Main monitoring loop"""
        logger.info("Starting IRANGATE AI Agent monitoring loop...")
        self.running = True
        
        consecutive_errors = 0
        max_consecutive_errors = 5
        error_backoff = 10  # seconds
        
        while self.running:
            try:
                # Get system metrics
                metrics = self.get_system_metrics()
                
                # Check for critical issues
                critical_issues = self.check_critical_issues(metrics)
                
                if critical_issues:
                    logger.warning(f"Critical issues detected: {critical_issues}")
                    
                    # Use AI to analyze and suggest actions
                    if self.openai_client and self.config.get('auto_fix_enabled'):
                        ai_result = self.analyze_with_ai(metrics)
                        
                        if ai_result.get('confidence', 0) > 70:  # High confidence threshold
                            action = ai_result.get('action', 'none')
                            reason = ai_result.get('reason', 'AI suggested action')
                            
                            logger.info(f"AI suggests: {action} (confidence: {ai_result.get('confidence')}%)")
                            
                            if action != 'none':
                                success = self.execute_action(action, reason)
                                if not success and action == 'restart_system':
                                    # Send alert for manual intervention
                                    self.send_notification(
                                        f"🚨 CRITICAL: Manual intervention required!\n"
                                        f"Issues: {', '.join(critical_issues)}\n"
                                        f"AI suggested: {action}\n"
                                        f"Please check the system immediately!"
                                    )
                        else:
                            # Low confidence - send alert to admin
                            self.send_notification(
                                f"⚠️ System issues detected but AI confidence is low\n"
                                f"Issues: {', '.join(critical_issues)}\n"
                                f"Please investigate manually."
                            )
                    else:
                        # No AI or auto-fix disabled - send alert
                        self.send_notification(
                            f"🚨 System issues detected:\n"
                            f"{', '.join(critical_issues)}\n"
                            f"Please investigate and resolve manually."
                        )
                
                # Reset error counter on successful iteration
                consecutive_errors = 0
                
                # Wait for next monitoring cycle
                time.sleep(self.config.get('monitoring_interval', 60))
                
            except KeyboardInterrupt:
                logger.info("Monitoring loop interrupted by user")
                break
            except Exception as e:
                consecutive_errors += 1
                logger.error(f"Error in monitoring loop (attempt {consecutive_errors}): {e}")
                
                # If too many consecutive errors, stop the loop
                if consecutive_errors >= max_consecutive_errors:
                    logger.error(f"Too many consecutive errors ({consecutive_errors}), stopping monitoring loop")
                    self.send_notification(
                        f"🚨 AI Agent stopped due to {consecutive_errors} consecutive errors\n"
                        f"Last error: {str(e)}\n"
                        f"Please check the system and restart the agent."
                    )
                    break
                
                # Exponential backoff for errors
                wait_time = error_backoff * (2 ** (consecutive_errors - 1))
                logger.info(f"Waiting {wait_time} seconds before retry...")
                time.sleep(wait_time)
        
        logger.info("AI Agent monitoring loop stopped")
    
    def check_critical_issues(self, metrics: Dict) -> List[str]:
        """Check for critical system issues"""
        issues = []
        
        # CPU usage check
        if metrics.get('cpu_percent', 0) > self.config.get('cpu_threshold', 90):
            issues.append(f"High CPU usage: {metrics.get('cpu_percent')}%")
        
        # RAM usage check
        if metrics.get('ram_percent', 0) > self.config.get('ram_threshold', 95):
            issues.append(f"High RAM usage: {metrics.get('ram_percent')}%")
        
        # Disk space check - convert disk_free to percentage for comparison
        disk_free_bytes = metrics.get('disk_free', 0)
        disk_total_bytes = metrics.get('disk_total', 1)  # Avoid division by zero
        if disk_total_bytes > 0:
            disk_free_percent = (disk_free_bytes / disk_total_bytes) * 100
            disk_threshold_percent = self.config.get('disk_threshold', 10)  # 10% free space
            if disk_free_percent < disk_threshold_percent:
                issues.append(f"Low disk space: {disk_free_percent:.1f}% free ({disk_free_bytes} bytes)")
        
        # OpenVPN status check
        if not metrics.get('openvpn_running', False):
            issues.append("OpenVPN service is not running")
        
        # Failed connections check
        if metrics.get('failed_connections', 0) > self.config.get('max_failed_connections', 100):
            issues.append(f"High failed connections: {metrics.get('failed_connections')}")
        
        return issues
    
    def stop(self):
        """Stop the monitoring loop"""
        self.running = False
        logger.info("AI Agent stop requested")


def main():
    """Main entry point"""
    if len(sys.argv) > 1:
        config_file = sys.argv[1]
    else:
        config_file = None  # Will use default based on platform
    
    # Ensure log directory exists
    log_dir = os.path.dirname(get_log_path())
    os.makedirs(log_dir, exist_ok=True)
    
    # Create and run AI agent
    agent = IRANGATEAIAgent(config_file)
    
    try:
        agent.monitor_loop()
    except KeyboardInterrupt:
        logger.info("AI Agent stopped by user")
    except Exception as e:
        logger.error(f"Fatal error in AI Agent: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
