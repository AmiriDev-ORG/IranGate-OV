#!/bin/bash
echo "🔄 Updating IranGate Web Panel..."
SOURCE_DIR="/root/ov/irangate/webpanel/frontend"
DEST_DIR="/opt/irangate/webpanel/frontend"
if [ ! -d "$SOURCE_DIR" ]; then
    echo "❌ Source directory not found: $SOURCE_DIR"
    exit 1
fi
if [ ! -d "$DEST_DIR" ]; then
    echo "❌ Destination directory not found: $DEST_DIR"
    echo "Please install the webpanel first"
    exit 1
fi
echo "📁 Copying updated files..."
cp -r "$SOURCE_DIR"/* "$DEST_DIR/"
chown -R root:root "$DEST_DIR"
chmod -R 644 "$DEST_DIR"
find "$DEST_DIR" -type d -exec chmod 755 {} \;
echo "🔄 Restarting webpanel service..."
systemctl restart irangate-webpanel
if systemctl is-active --quiet irangate-webpanel; then
    echo "✅ Webpanel updated and restarted successfully!"
    echo "🌐 Access your webpanel at: http://$(hostname -I | awk '{print $1}'):444/"
else
    echo "❌ Failed to restart webpanel service"
    systemctl status irangate-webpanel --no-pager
    exit 1
fi
echo "🎉 Update complete!"