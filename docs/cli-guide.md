# IranGate CLI - راهنمای کامل

## 📋 فهرست مطالب
- [نصب و راه‌اندازی](#نصب-و-راه‌اندازی)
- [دستورات اصلی](#دستورات-اصلی)
- [مدیریت OpenVPN Server](#مدیریت-openvpn-server)
- [مدیریت کلاینت‌ها](#مدیریت-کلاینت‌ها)
- [پیکربندی سیستم](#پیکربندی-سیستم)
- [مانیتورینگ](#مانیتورینگ)
- [اتوماسیون](#اتوماسیون)
- [🤖 AI Management](#-ai-management)
- [مثال‌های کاربردی](#مثال‌های-کاربردی)
- [عیب‌یابی](#عیب‌یابی)

---

## 🚀 نصب و راه‌اندازی

### پیش‌نیازها
```bash
# سیستم عامل: Linux (Ubuntu 18.04+ یا CentOS 7+)
# دسترسی: Root یا sudo privileges
# بسته‌ها: OpenVPN, easy-rsa
```

### نصب IranGate CLI
```bash
# 1. کلون کردن پروژه
git clone https://github.com/amiridev-org/irangate-ov
cd irangate-ov/irangate

# 2. کامپایل کردن
make build

# 3. نصب در سیستم
sudo make install

# 4. تایید نصب
irangate --version
```

---

## 🎯 دستورات اصلی

### ساختار کلی دستورات
```bash
irangate [GLOBAL_OPTIONS] <COMMAND> [SUBCOMMAND] [OPTIONS] [ARGUMENTS]
```

### Global Options
```bash
--config string    # مسیر فایل تنظیمات
--json            # خروجی JSON format
--help, -h        # راهنمای دستور
--version, -v     # نمایش نسخه
```

### فهرست کامل دستورات

#### 🔧 Core Commands (دستورات اصلی)
```bash
irangate install          # نصب OpenVPN server
irangate uninstall        # حذف OpenVPN server
irangate start            # شروع سرویس OpenVPN
irangate stop             # توقف سرویس OpenVPN
irangate restart          # راه‌اندازی مجدد سرویس
irangate status           # نمایش وضعیت سرویس
irangate version          # نمایش نسخه 🆕
irangate check-deps       # بررسی وابستگی‌ها 🆕
irangate test-recovery    # تست بازیابی سیستم 🆕
```

#### 👥 Client Management (مدیریت کلاینت‌ها)
```bash
irangate client add <name>        # اضافه کردن کلاینت جدید
irangate client remove <name>     # حذف کلاینت
irangate client list              # نمایش لیست کلاینت‌ها
```

#### ⚙️ Configuration (پیکربندی)
```bash
irangate config show              # نمایش تنظیمات فعلی
```

#### 📊 Monitoring (مانیتورینگ)
```bash
irangate monitor live             # نمایش آمار زنده
```

#### 🤖 AI Management (مدیریت هوشمند)
```bash
irangate ai status                # وضعیت AI agent
irangate ai start                 # شروع AI monitoring
irangate ai stop                  # توقف AI monitoring
irangate ai restart               # راه‌اندازی مجدد AI
irangate ai config show           # نمایش تنظیمات AI
irangate ai config set <key> <value>  # تنظیم مقدار AI
irangate ai test                  # تست عملکرد AI
irangate ai install               # نصب dependencies AI
```

#### ⏰ Scheduler (زمان‌بندی)
```bash
irangate cron list                # نمایش لیست cron jobs
```

### مثال کلی
```bash
# نمایش راهنما
irangate --help

# خروجی JSON
irangate --json client list

# استفاده از فایل تنظیمات خاص
irangate --config /etc/irangate/custom.yaml status

# نمایش راهنمای دستور خاص
irangate client --help
irangate ai --help
irangate config --help
```

---

## 🔧 مدیریت OpenVPN Server

### نصب و راه‌اندازی اولیه

#### نصب OpenVPN Server
```bash
# نصب کامل OpenVPN با تنظیمات پیش‌فرض
sudo irangate install

# خروجی موفق:
# ✓ OpenVPN server installed successfully
# ✓ PKI initialized
# ✓ Server configuration created
# ✓ Service ready to start
```

#### مدیریت سرویس
```bash
# شروع سرویس
sudo irangate start

# توقف سرویس
sudo irangate stop

# راه‌اندازی مجدد
sudo irangate restart

# بررسی وضعیت
irangate status

# خروجی status:
# OpenVPN service status: running
# Uptime: 2h 15m 30s
# Connected clients: 3
```

#### حذف کامل OpenVPN
```bash
# حذف کامل (مراقب باشید!)
sudo irangate uninstall

# این دستور:
# - سرویس را متوقف می‌کند
# - فایل‌های تنظیمات را حذف می‌کند
# - گواهی‌ها را پاک می‌کند
# - بسته‌های OpenVPN را حذف می‌کند
```

---

## 👥 مدیریت کلاینت‌ها

### اضافه کردن کلاینت جدید

#### اضافه کردن کلاینت ساده
```bash
# اضافه کردن کلاینت جدید
sudo irangate client add john

# خروجی موفق:
# ✓ Client 'john' created successfully
# ✓ Certificates generated
# ✓ Configuration file created
# ✓ Client added to database
```

#### اضافه کردن کلاینت با تنظیمات خاص
```bash
# نام کلاینت باید:
# - حداقل 2 کاراکتر باشد
# - با حروف/اعداد شروع و پایان یابد
# - فقط حروف، اعداد، _ و - داشته باشد

# مثال‌های معتبر:
sudo irangate client add user123
sudo irangate client add john-doe
sudo irangate client add admin_user

# مثال‌های نامعتبر:
sudo irangate client add -user    # با - شروع می‌شود
sudo irangate client add user-    # با - پایان می‌یابد
sudo irangate client add u        # کمتر از 2 کاراکتر
```

### مدیریت کلاینت‌ها

#### نمایش لیست کلاینت‌ها
```bash
# نمایش تمام کلاینت‌ها
irangate client list

# خروجی:
# VPN Clients:
# [
#   {
#     "name": "john",
#     "created_at": "2025-01-15T10:30:00Z",
#     "expires_at": "2025-02-15T10:30:00Z",
#     "active": true,
#     "ip": "10.8.0.2",
#     "cipher": "AES-256-GCM",
#     "data_used_mb": 1024
#   }
# ]
```

#### نمایش جزئیات کلاینت خاص
```bash
# نمایش اطلاعات کلاینت
irangate client show john

# خروجی:
# Client: john
# Created: 2025-01-15 10:30:00
# Expires: 2025-02-15 10:30:00
# Status: Active
# IP: 10.8.0.2
# Data Used: 1.0 GB
```

#### حذف کلاینت
```bash
# حذف کلاینت
sudo irangate client remove john

# خروجی:
# ✓ Client 'john' removed successfully
# ✓ Certificate revoked
# ✓ Configuration files deleted
# ✓ Client removed from database

# این دستور:
# - گواهی کلاینت را لغو می‌کند
# - فایل‌های تنظیمات را حذف می‌کند
# - کلاینت را از دیتابیس پاک می‌کند
```

#### صادرات تنظیمات کلاینت
```bash
# صادرات فایل .ovpn
sudo irangate client export john /home/admin/john.ovpn

# خروجی:
# ✓ Client configuration exported to /home/admin/john.ovpn

# فایل صادر شده شامل:
# - تنظیمات اتصال به سرور
# - گواهی‌های کلاینت
# - کلیدهای رمزنگاری
# - تنظیمات DNS
```

---

## ⚙️ پیکربندی سیستم

### نمایش تنظیمات فعلی
```bash
# نمایش تمام تنظیمات
irangate config show

# خروجی:
# OpenVPN Configuration:
# {
#   "server_ip": "0.0.0.0",
#   "server_port": 1194,
#   "protocol": "udp",
#   "cipher": "AES-256-GCM",
#   "mtu": 1412,
#   "dns": ["1.1.1.1", "8.8.8.8"],
#   "ipv6_enabled": false,
#   "auto_config_mode": "dynamic",
#   "backup_path": "/opt/irangate/database/backups/",
#   "log_level": 3,
#   "version": "1.0.0"
# }
```

### ویرایش تنظیمات
```bash
# باز کردن ویرایشگر برای تنظیمات
sudo irangate config edit

# این دستور:
# - فایل تنظیمات را در ویرایشگر باز می‌کند
# - پس از ذخیره، سرویس را راه‌اندازی مجدد می‌کند
```

### پشتیبان‌گیری و بازگردانی

#### ایجاد پشتیبان
```bash
# ایجاد پشتیبان از تنظیمات
sudo irangate config backup

# خروجی:
# ✓ Settings backed up successfully
# Backup file: /opt/irangate/database/backups/backup_2025-01-15T12:00:00.json

# فایل پشتیبان شامل:
# - تمام تنظیمات سرور
# - لیست کلاینت‌ها
# - تنظیمات cron jobs
```

#### بازگردانی از پشتیبان
```bash
# بازگردانی از فایل پشتیبان
sudo irangate config restore /opt/irangate/database/backups/backup_2025-01-15T12:00:00.json

# خروجی:
# ✓ Settings restored successfully
# ✓ OpenVPN service restarted
```

---

## 📊 مانیتورینگ

### مانیتورینگ زنده
```bash
# نمایش آمار زنده
irangate monitor live

# خروجی (هر 5 ثانیه به‌روزرسانی می‌شود):
# OpenVPN Statistics:
# {
#   "uptime": "2h 15m 30s",
#   "connected_clients": 3,
#   "total_traffic_mb": 2048.5,
#   "cpu_usage": 5.2,
#   "memory_usage_mb": 128.4
# }
```

### نمایش کلاینت‌های متصل
```bash
# لیست کلاینت‌های آنلاین
irangate monitor clients

# خروجی:
# Connected Clients:
# [
#   {
#     "common_name": "john",
#     "real_address": "203.0.113.100:12345",
#     "virtual_address": "10.8.0.2",
#     "bytes_received": 1048576,
#     "bytes_sent": 524288,
#     "connected_since": "2025-01-15T10:30:00Z"
#   }
# ]
```

### نمایش آمار ترافیک
```bash
# آمار ترافیک کلی
irangate monitor traffic

# خروجی:
# Traffic Statistics:
# {
#   "total_bytes_in": 1073741824,
#   "total_bytes_out": 536870912,
#   "total_clients": 3,
#   "average_speed_kbps": 1024,
#   "peak_speed_kbps": 2048
# }
```

---

## 🤖 اتوماسیون

### مدیریت Cron Jobs

#### نمایش لیست Job ها
```bash
# نمایش تمام job های برنامه‌ریزی شده
irangate cron list

# خروجی:
# Scheduled Jobs:
# [
#   {
#     "id": "backup_daily",
#     "schedule": "0 2 * * *",
#     "action": "backup",
#     "params": {
#       "type": "full",
#       "retention_days": 7
#     },
#     "enabled": true,
#     "last_run": "2025-01-15T02:00:00Z"
#   }
# ]
```

#### اضافه کردن Job جدید
```bash
# اضافه کردن پشتیبان‌گیری روزانه
sudo irangate cron add "0 2 * * *" "backup" '{"type":"full","retention_days":7}'

# اضافه کردن مانیتورینگ هر 30 دقیقه
sudo irangate cron add "*/30 * * * *" "monitor" '{"type":"health","threshold":90}'

# اضافه کردن پاکسازی هفتگی
sudo irangate cron add "0 3 * * 0" "cleanup" '{"retention_days":30}'

# فرمت schedule (cron format):
# * * * * *
# │ │ │ │ │
# │ │ │ │ └─── Day of week (0-7, Sunday = 0 or 7)
# │ │ │ └───── Month (1-12)
# │ │ └─────── Day of month (1-31)
# │ └───────── Hour (0-23)
# └─────────── Minute (0-59)
```

#### حذف Job
```bash
# حذف job با ID
sudo irangate cron remove backup_daily

# خروجی:
# ✓ Cron job 'backup_daily' removed successfully
```

#### اجرای دستی Job
```bash
# اجرای فوری job
sudo irangate cron run backup_daily

# خروجی:
# ✓ Job 'backup_daily' started successfully
# ✓ Job completed in 45.2s
```

---

## 🤖 AI Management

### معرفی AI Management
سیستم مدیریت هوشمند IRANGATE که با استفاده از OpenAI GPT مشکلات سیستم را تشخیص داده و به صورت خودکار حل می‌کند.

### نصب و راه‌اندازی اولیه

#### نصب Dependencies
```bash
# نصب Python packages مورد نیاز
irangate ai install

# خروجی موفق:
# ✓ Installing AI dependencies...
# ✓ openai>=1.0.0 installed
# ✓ psutil>=5.9.0 installed  
# ✓ requests>=2.28.0 installed
# ✓ AI dependencies installed successfully
```

#### تنظیم API Key
```bash
# تنظیم OpenAI API key
irangate ai config set openai_api_key "sk-your-openai-api-key-here"

# تنظیم Telegram bot (اختیاری)
irangate ai config set telegram_bot_token "1234567890:ABC-DEF-GHI-JKL-MNO"
irangate ai config set admin_chat_id "123456789"
```

### دستورات AI Management

#### مدیریت وضعیت AI Agent
```bash
# نمایش وضعیت AI agent
irangate ai status

# خروجی:
# 🤖 IRANGATE AI Agent Status
# ============================
# Status: ✅ Running
# Config: ✅ Loaded
# 
# Configuration:
#   Monitoring Interval: 60 seconds
#   Auto-fix Enabled: true
#   Telegram Notifications: true
#   OpenAI Configured: true
#   Telegram Configured: true

# شروع AI monitoring
irangate ai start

# توقف AI monitoring  
irangate ai stop

# راه‌اندازی مجدد AI monitoring
irangate ai restart
```

#### تنظیمات AI
```bash
# نمایش تنظیمات فعلی
irangate ai config show

# خروجی:
# 🤖 AI Configuration
# ===================
# Monitoring Interval: 60 seconds
# CPU Threshold: 90%
# RAM Threshold: 95%
# Disk Threshold: 10% free
# Max Failed Connections: 100
# Auto-fix Enabled: true
# Telegram Notifications: true
# OpenAI API Key: sk-****1234
# Telegram Bot Token: 1234****5678
# Admin Chat ID: 123456789
# Log File: /var/log/irangate/ai_agent.log

# تغییر تنظیمات
irangate ai config set monitoring_interval 120    # هر 2 دقیقه
irangate ai config set cpu_threshold 85          # آستانه CPU
irangate ai config set auto_fix_enabled false    # غیرفعال کردن auto-fix
```

#### تست و عیب‌یابی
```bash
# تست عملکرد AI
irangate ai test

# خروجی:
# 🧪 Testing AI Configuration
# ===========================
# Testing Python dependencies...
# ✅ Python dependencies: OK
# Testing configuration...
# ✅ OpenAI API key: Configured
# ✅ Telegram bot token: Configured
# 
# ✅ AI configuration test completed
```

### قابلیت‌های AI Management

#### 🔍 مانیتورینگ هوشمند
AI agent هر 60 ثانیه (قابل تنظیم) سیستم را بررسی می‌کند:
- **CPU Usage:** استفاده از پردازنده
- **RAM Usage:** استفاده از حافظه
- **Disk Space:** فضای باقی‌مانده دیسک
- **Network Traffic:** ترافیک شبکه
- **OpenVPN Status:** وضعیت سرویس OpenVPN
- **Failed Connections:** تعداد اتصالات ناموفق

#### 🛠️ Auto-Healing (خوددرمانی)
AI به صورت خودکار مشکلات زیر را حل می‌کند:

**1. Restart OpenVPN Service**
```bash
# زمانی که AI تشخیص دهد:
# - سرویس OpenVPN متوقف شده
# - تعداد اتصالات ناموفق بالا باشد
# 
# اقدام AI:
# sudo systemctl restart openvpn
# ✅ OpenVPN service restarted successfully
```

**2. Cleanup Log Files**
```bash
# زمانی که AI تشخیص دهد:
# - فضای دیسک کمتر از 10% باشد
#
# اقدام AI:
# find /var/log -name "*.log" -mtime +7 -delete
# ✅ Log files cleaned up successfully
```

**3. Check Certificates**
```bash
# زمانی که AI تشخیص دهد:
# - خطاهای مربوط به گواهی‌ها
#
# اقدام AI:
# openssl x509 -in /etc/openvpn/*.crt -noout -dates
# ✅ Certificate check completed
```

#### 📱 Telegram Integration
```bash
# تنظیم Telegram bot:
# 1. ایجاد bot با @BotFather
# 2. دریافت token
# 3. دریافت chat ID با @userinfobot
# 4. تنظیم در IRANGATE:

irangate ai config set telegram_bot_token "YOUR_BOT_TOKEN"
irangate ai config set admin_chat_id "YOUR_CHAT_ID"
```

**دستورات Telegram Bot:**
- `/status` - وضعیت سیستم
- `/logs` - لاگ‌های اخیر
- `/fix` - اجرای دستی fix
- `/config` - نمایش تنظیمات AI

**نمونه پیام‌های هشدار:**
```
🤖 IRANGATE AI Agent

🚨 ALERT: High CPU Usage

📊 Status: Critical
🔧 Auto-fix: ✅ Success
📝 Details: CPU usage reached 95%, service restarted

🕐 Time: 2025-01-15T14:30:00Z
```

### تنظیمات پیشرفته

#### Thresholds (آستانه‌ها)
```bash
# تنظیم آستانه‌های مختلف
irangate ai config set cpu_threshold 80        # CPU > 80%
irangate ai config set ram_threshold 90        # RAM > 90%  
irangate ai config set disk_threshold 15       # Disk < 15%
irangate ai config set max_failed_connections 50  # > 50 failed connections
```

#### Monitoring Intervals
```bash
# تنظیم فواصل مانیتورینگ
irangate ai config set monitoring_interval 30    # هر 30 ثانیه
irangate ai config set monitoring_interval 120   # هر 2 دقیقه
irangate ai config set monitoring_interval 300   # هر 5 دقیقه
```

#### Auto-fix Control
```bash
# کنترل auto-fix
irangate ai config set auto_fix_enabled true     # فعال
irangate ai config set auto_fix_enabled false    # غیرفعال

# در صورت غیرفعال بودن، فقط هشدار ارسال می‌شود
```

### لاگ‌ها و نظارت

#### لاگ‌های AI Agent
```bash
# مشاهده لاگ‌های AI
tail -f /var/log/irangate/ai_agent.log

# نمونه خروجی:
# 2025-01-15T14:30:00Z - INFO - Monitoring cycle started
# 2025-01-15T14:30:01Z - INFO - CPU: 45%, RAM: 67%, Disk: 25%
# 2025-01-15T14:30:02Z - INFO - OpenVPN: running, Failed connections: 12
# 2025-01-15T14:30:03Z - INFO - System healthy, no action needed
```

#### نظارت بر عملکرد AI
```bash
# بررسی وضعیت AI
irangate ai status

# تست عملکرد
irangate ai test

# نمایش تنظیمات
irangate ai config show
```

### مثال‌های کاربردی

#### سناریو 1: راه‌اندازی AI برای سرور جدید
```bash
# 1. نصب dependencies
irangate ai install

# 2. تنظیم OpenAI API key
irangate ai config set openai_api_key "sk-your-key"

# 3. تنظیم Telegram (اختیاری)
irangate ai config set telegram_bot_token "your-bot-token"
irangate ai config set admin_chat_id "your-chat-id"

# 4. شروع AI monitoring
irangate ai start

# 5. تست عملکرد
irangate ai test
```

#### سناریو 2: تنظیم AI برای سرور با ترافیک بالا
```bash
# 1. کاهش فاصله مانیتورینگ
irangate ai config set monitoring_interval 30

# 2. تنظیم آستانه‌های حساس‌تر
irangate ai config set cpu_threshold 75
irangate ai config set ram_threshold 85
irangate ai config set max_failed_connections 25

# 3. فعال‌سازی auto-fix
irangate ai config set auto_fix_enabled true

# 4. راه‌اندازی مجدد AI
irangate ai restart
```

#### سناریو 3: عیب‌یابی AI Agent
```bash
# 1. بررسی وضعیت
irangate ai status

# 2. تست dependencies
irangate ai test

# 3. بررسی لاگ‌ها
tail -f /var/log/irangate/ai_agent.log

# 4. راه‌اندازی مجدد در صورت نیاز
irangate ai restart
```

### نکات مهم AI Management

#### امنیت
- API key های OpenAI را محفوظ نگه دارید
- Telegram bot token را در دسترس عموم قرار ندهید
- از environment variables برای اطلاعات حساس استفاده کنید

#### عملکرد
- AI agent منابع کمی مصرف می‌کند
- مانیتورینگ هر 30-300 ثانیه قابل تنظیم است
- Auto-fix فقط برای مشکلات رایج فعال است

#### هزینه‌ها
- OpenAI API بر اساس usage محاسبه می‌شود
- برای سرورهای کوچک هزینه‌ای ندارد
- برای سرورهای بزرگ، usage را نظارت کنید

---

## 💡 مثال‌های کاربردی

### سناریو 1: راه‌اندازی VPN Server جدید

```bash
# 1. نصب IranGate
sudo make install

# 2. نصب OpenVPN
sudo irangate install

# 3. تنظیم IP سرور
sudo irangate config edit
# در ویرایشگر: server_ip را به IP واقعی سرور تغییر دهید

# 4. شروع سرویس
sudo irangate start

# 5. اضافه کردن کلاینت‌های اولیه
sudo irangate client add admin
sudo irangate client add user1
sudo irangate client add user2

# 6. صادرات فایل‌های کلاینت
sudo irangate client export admin /home/admin/admin.ovpn
sudo irangate client export user1 /home/admin/user1.ovpn
sudo irangate client export user2 /home/admin/user2.ovpn

# 7. راه‌اندازی پشتیبان‌گیری خودکار
sudo irangate cron add "0 2 * * *" "backup" '{"type":"full","retention_days":7}'

# 8. راه‌اندازی AI Management (اختیاری)
irangate ai install
irangate ai config set openai_api_key "sk-your-openai-key"
irangate ai start
```

### سناریو 2: مدیریت روزانه کلاینت‌ها

```bash
# 1. بررسی وضعیت کلی
irangate status
irangate monitor live

# 2. بررسی کلاینت‌های متصل
irangate monitor clients

# 3. اضافه کردن کلاینت جدید
sudo irangate client add newuser

# 4. حذف کلاینت قدیمی
sudo irangate client remove olduser

# 5. بررسی آمار ترافیک
irangate monitor traffic

# 6. ایجاد پشتیبان
sudo irangate config backup

# 7. بررسی وضعیت AI (اگر فعال باشد)
irangate ai status
```

### سناریو 3: عیب‌یابی و نگهداری

```bash
# 1. بررسی وضعیت سرویس
irangate status

# 2. راه‌اندازی مجدد در صورت نیاز
sudo irangate restart

# 3. بررسی لاگ‌ها
sudo journalctl -u openvpn@server -f

# 4. بررسی تنظیمات
irangate config show

# 5. بازگردانی از پشتیبان در صورت مشکل
sudo irangate config restore /path/to/backup.json

# 6. حذف و نصب مجدد در صورت مشکل جدی
sudo irangate uninstall
sudo irangate install

# 7. عیب‌یابی AI (اگر فعال باشد)
irangate ai status
irangate ai test
tail -f /var/log/irangate/ai_agent.log
```

### سناریو 4: راه‌اندازی سرور با AI Management

```bash
# 1. نصب و راه‌اندازی اولیه
sudo make install
sudo irangate install
sudo irangate start

# 2. اضافه کردن کلاینت‌های اولیه
sudo irangate client add admin
sudo irangate client add user1

# 3. نصب و تنظیم AI Management
irangate ai install
irangate ai config set openai_api_key "sk-your-openai-api-key"
irangate ai config set telegram_bot_token "your-telegram-bot-token"
irangate ai config set admin_chat_id "your-chat-id"

# 4. تنظیم آستانه‌های حساس‌تر برای سرور پرترافیک
irangate ai config set monitoring_interval 30
irangate ai config set cpu_threshold 75
irangate ai config set ram_threshold 85
irangate ai config set max_failed_connections 25

# 5. شروع AI monitoring
irangate ai start

# 6. تست عملکرد
irangate ai test

# 7. راه‌اندازی cron jobs
sudo irangate cron add "0 2 * * *" "backup" '{"type":"full","retention_days":7}'
```

---

## 🔧 عیب‌یابی

### مشکلات رایج

#### 1. خطای "Permission Denied"
```bash
# مشکل: عدم دسترسی کافی
# حل:
sudo irangate [command]

# یا اضافه کردن کاربر به گروه sudo
sudo usermod -aG sudo $USER
```

#### 2. خطای "OpenVPN not installed"
```bash
# مشکل: OpenVPN نصب نشده
# حل:
sudo irangate install

# یا نصب دستی:
sudo apt install openvpn easy-rsa  # Ubuntu/Debian
sudo yum install openvpn easy-rsa  # CentOS/RHEL
```

#### 3. خطای "Port already in use"
```bash
# مشکل: پورت 1194 در حال استفاده
# حل:
# بررسی پروسه‌های استفاده‌کننده از پورت
sudo netstat -tulpn | grep 1194

# متوقف کردن پروسه‌های متضاد
sudo systemctl stop openvpn
sudo systemctl stop openvpn@server

# سپس:
sudo irangate start
```

#### 4. خطای "Client already exists"
```bash
# مشکل: کلاینت با همین نام وجود دارد
# حل:
# بررسی کلاینت‌های موجود
irangate client list

# حذف کلاینت قدیمی یا استفاده از نام جدید
sudo irangate client remove oldname
sudo irangate client add newname
```

#### 5. خطای "Database locked"
```bash
# مشکل: دیتابیس قفل شده
# حل:
# بررسی پروسه‌های استفاده‌کننده
sudo lsof /opt/irangate/database/

# راه‌اندازی مجدد سرویس
sudo irangate restart
```

#### 6. خطای "AI Agent failed to start"
```bash
# مشکل: AI agent شروع نمی‌شود
# حل:
# بررسی Python dependencies
irangate ai test

# نصب مجدد dependencies
irangate ai install

# بررسی لاگ‌های AI
tail -f /var/log/irangate/ai_agent.log

# بررسی دسترسی‌ها
sudo chmod +x /path/to/ai_agent.py
```

#### 7. خطای "OpenAI API key invalid"
```bash
# مشکل: API key نامعتبر
# حل:
# بررسی API key
irangate ai config show

# تنظیم مجدد API key
irangate ai config set openai_api_key "sk-new-valid-key"

# تست اتصال
irangate ai test
```

#### 8. خطای "Telegram bot not responding"
```bash
# مشکل: Telegram bot پاسخ نمی‌دهد
# حل:
# بررسی bot token
irangate ai config show

# بررسی chat ID
# از @userinfobot برای دریافت chat ID استفاده کنید

# تنظیم مجدد
irangate ai config set telegram_bot_token "new-bot-token"
irangate ai config set admin_chat_id "new-chat-id"
```

### بررسی سلامت سیستم

#### بررسی کامل وضعیت
```bash
# اسکریپت بررسی سلامت
#!/bin/bash
echo "=== IranGate Health Check ==="
echo "1. Service Status:"
irangate status

echo -e "\n2. Connected Clients:"
irangate monitor clients

echo -e "\n3. Configuration:"
irangate config show

echo -e "\n4. Cron Jobs:"
irangate cron list

echo -e "\n5. System Resources:"
irangate monitor live

echo -e "\n6. AI Agent Status:"
irangate ai status

echo -e "\n7. AI Configuration Test:"
irangate ai test
```

#### لاگ‌های مهم
```bash
# لاگ OpenVPN
sudo tail -f /var/log/openvpn/openvpn.log

# لاگ سیستم
sudo journalctl -u openvpn@server -f

# لاگ IranGate
sudo journalctl -u irangate -f

# لاگ AI Agent
sudo tail -f /var/log/irangate/ai_agent.log

# لاگ‌های سیستم
sudo tail -f /var/log/syslog
```

---

## 📚 نکات مهم

### امنیت
- همیشه از sudo برای دستورات مدیریتی استفاده کنید
- فایل‌های .ovpn را در مکان امن نگهداری کنید
- پشتیبان‌گیری منظم انجام دهید
- کلاینت‌های غیرضروری را حذف کنید

### عملکرد
- مانیتورینگ منظم انجام دهید
- ترافیک کلاینت‌ها را بررسی کنید
- منابع سیستم را نظارت کنید
- Cron jobs را بهینه کنید

### نگهداری
- به‌روزرسانی‌های امنیتی را نصب کنید
- لاگ‌ها را مرتب پاک کنید
- فایل‌های قدیمی را حذف کنید
- تنظیمات را دوره‌ای بررسی کنید
- AI agent را نظارت کنید
- هزینه‌های OpenAI API را کنترل کنید

---

## 🆘 پشتیبانی

### منابع مفید
- [مستندات OpenVPN](https://openvpn.net/community-resources/)
- [راهنمای Easy-RSA](https://github.com/OpenVPN/easy-rsa)
- [مستندات systemd](https://systemd.io/)

### گزارش باگ
در صورت بروز مشکل، لطفاً اطلاعات زیر را ارائه دهید:
- نسخه IranGate: `irangate --version`
- سیستم عامل: `uname -a`
- خروجی دستور مشکل‌دار: `irangate [command] --json`
- لاگ‌های مربوطه: `sudo journalctl -u openvpn@server`

---

## 📋 خلاصه کامل دستورات

### 🔧 Core Commands
| دستور | توضیحات |
|-------|---------|
| `irangate install` | نصب OpenVPN server |
| `irangate uninstall` | حذف OpenVPN server |
| `irangate start` | شروع سرویس OpenVPN |
| `irangate stop` | توقف سرویس OpenVPN |
| `irangate restart` | راه‌اندازی مجدد سرویس |
| `irangate status` | نمایش وضعیت سرویس |

### 👥 Client Management
| دستور | توضیحات |
|-------|---------|
| `irangate client add <name>` | اضافه کردن کلاینت جدید |
| `irangate client remove <name>` | حذف کلاینت |
| `irangate client list` | نمایش لیست کلاینت‌ها |

### ⚙️ Configuration
| دستور | توضیحات |
|-------|---------|
| `irangate config show` | نمایش تنظیمات فعلی |

### 📊 Monitoring
| دستور | توضیحات |
|-------|---------|
| `irangate monitor live` | نمایش آمار زنده |

### 🤖 AI Management
| دستور | توضیحات |
|-------|---------|
| `irangate ai status` | وضعیت AI agent |
| `irangate ai start` | شروع AI monitoring |
| `irangate ai stop` | توقف AI monitoring |
| `irangate ai restart` | راه‌اندازی مجدد AI |
| `irangate ai config show` | نمایش تنظیمات AI |
| `irangate ai config set <key> <value>` | تنظیم مقدار AI |
| `irangate ai test` | تست عملکرد AI |
| `irangate ai install` | نصب dependencies AI |

### ⏰ Scheduler
| دستور | توضیحات |
|-------|---------|
| `irangate cron list` | نمایش لیست cron jobs |

### 🌐 Global Options
| گزینه | توضیحات |
|-------|---------|
| `--config <file>` | مسیر فایل تنظیمات |
| `--json` | خروجی JSON format |
| `--help, -h` | راهنمای دستور |
| `--version, -v` | نمایش نسخه |

### 📝 مثال‌های کاربردی
```bash
# نصب و راه‌اندازی اولیه
sudo irangate install
sudo irangate start
sudo irangate client add user1

# راه‌اندازی AI Management
irangate ai install
irangate ai config set openai_api_key "sk-your-key"
irangate ai start

# مانیتورینگ روزانه
irangate status
irangate monitor live
irangate ai status

# مدیریت کلاینت‌ها
irangate client list
sudo irangate client add newuser
sudo irangate client remove olduser

# عیب‌یابی
irangate ai test
tail -f /var/log/irangate/ai_agent.log
sudo irangate restart
```

---

**🎉 موفق باشید در استفاده از IranGate CLI!**
