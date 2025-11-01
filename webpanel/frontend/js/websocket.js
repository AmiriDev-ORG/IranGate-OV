// WebSocket Client for Real-time Updates

class WebSocketClient {
    constructor() {
        this.ws = null;
        this.reconnectInterval = 5000;
        this.reconnectTimer = null;
        this.eventHandlers = {};
        this.isConnected = false;
        this.autoReconnect = true;
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;

        try {
            this.ws = new WebSocket(wsUrl);

            this.ws.onopen = () => {
                this.isConnected = true;
                this.trigger('connected', {});
                
                // Clear reconnect timer
                if (this.reconnectTimer) {
                    clearTimeout(this.reconnectTimer);
                    this.reconnectTimer = null;
                }

                // Show connection notification
                if (typeof notificationManager !== 'undefined') {
                    notificationManager.success('Real-time updates connected');
                }
            };

            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleMessage(message);
                } catch (error) {
                    console.error('WebSocket message parse error:', error);
                }
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.isConnected = false;
            };

            this.ws.onclose = () => {
                this.isConnected = false;
                this.trigger('disconnected', {});

                // Auto-reconnect
                if (this.autoReconnect) {
                    this.scheduleReconnect();
                }
            };

        } catch (error) {
            console.error('WebSocket connection error:', error);
            this.scheduleReconnect();
        }
    }

    scheduleReconnect() {
        if (this.reconnectTimer) return;

        this.reconnectTimer = setTimeout(() => {
            this.connect();
        }, this.reconnectInterval);
    }

    handleMessage(message) {
        const { type, data, timestamp } = message;

        // Trigger registered handlers
        this.trigger(type, data, timestamp);

        // Handle specific event types
        switch (type) {
            case 'connected':
                break;

            case 'client_created':
                if (typeof notificationManager !== 'undefined') {
                    notificationManager.success(`New client created: ${data.name}`);
                }
                break;

            case 'client_removed':
                if (typeof notificationManager !== 'undefined') {
                    notificationManager.info(`Client removed: ${data.name}`);
                }
                break;

            case 'stats_update':
                // Update dashboard stats in real-time
                this.updateDashboardStats(data);
                break;

            case 'server_restart':
                if (typeof notificationManager !== 'undefined') {
                    notificationManager.warning('OpenVPN server is restarting...');
                }
                break;

            case 'orphaned_found':
                if (typeof notificationManager !== 'undefined') {
                    notificationManager.warning(
                        `Found ${data.count} orphaned certificates!`,
                        {
                            action: () => window.location.href = 'clients.html#orphaned',
                            actionText: 'View Details'
                        }
                    );
                }
                break;

            default:
        }
    }

    updateDashboardStats(stats) {
        // Update dashboard elements if they exist
        const elements = {
            'connectedClients': stats.connected_clients,
            'cpuUsage': stats.cpu_usage?.toFixed(1) + '%',
            'memoryUsage': stats.memory_usage_mb?.toFixed(0) + ' MB'
        };

        for (const [id, value] of Object.entries(elements)) {
            const el = document.getElementById(id);
            if (el && value !== undefined) {
                el.textContent = value;
            }
        }

        // Update progress bars
        if (stats.cpu_usage !== undefined) {
            const cpuBar = document.getElementById('cpuBar');
            if (cpuBar) {
                cpuBar.style.width = stats.cpu_usage + '%';
            }
        }
    }

    // Register event handler
    on(eventType, handler) {
        if (!this.eventHandlers[eventType]) {
            this.eventHandlers[eventType] = [];
        }
        this.eventHandlers[eventType].push(handler);
    }

    // Remove event handler
    off(eventType, handler) {
        if (this.eventHandlers[eventType]) {
            this.eventHandlers[eventType] = this.eventHandlers[eventType]
                .filter(h => h !== handler);
        }
    }

    // Trigger event handlers
    trigger(eventType, data, timestamp) {
        if (this.eventHandlers[eventType]) {
            this.eventHandlers[eventType].forEach(handler => {
                try {
                    handler(data, timestamp);
                } catch (error) {
                    console.error(`Error in ${eventType} handler:`, error);
                }
            });
        }
    }

    // Send message to server
    send(type, data) {
        if (this.isConnected && this.ws) {
            const message = {
                type,
                data,
                timestamp: new Date().toISOString()
            };
            this.ws.send(JSON.stringify(message));
        } else {
            console.warn('WebSocket not connected, cannot send message');
        }
    }

    // Disconnect
    disconnect() {
        this.autoReconnect = false;
        if (this.ws) {
            this.ws.close();
        }
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
        }
    }

    // Get connection status
    isConnectedToServer() {
        return this.isConnected;
    }
}

// Global WebSocket client instance
const wsClient = new WebSocketClient();

// Auto-connect when page loads (if authenticated)
if (typeof checkAuth === 'function') {
    document.addEventListener('DOMContentLoaded', () => {
        if (localStorage.getItem('token')) {
            setTimeout(() => wsClient.connect(), 1000);
        }
    });
}

