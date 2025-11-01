// API Client for IranGate Web Panel

const API_BASE = window.location.origin + '/api';

class API {
    constructor() {
        this.token = localStorage.getItem('token');
        this.cache = new Map();
        this.cacheTimeout = 30000; // 30 seconds default
    }

    // Check if token is expired
    isTokenExpired() {
        const expiresAt = localStorage.getItem('expires_at');
        if (!expiresAt) {
            return true; // No expiration date means expired
        }
        
        const expirationTime = new Date(expiresAt).getTime();
        const currentTime = Date.now();
        
        // Check if token expires in less than 5 minutes
        return currentTime >= (expirationTime - 5 * 60 * 1000);
    }

    // Refresh token if needed
    async refreshTokenIfNeeded() {
        if (this.isTokenExpired()) {
            console.warn('Token expired or about to expire. Logging out...');
            localStorage.clear();
            window.location.href = 'index.html';
            return false;
        }
        return true;
    }

    getHeaders() {
        const headers = {
            'Content-Type': 'application/json'
        };
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        return headers;
    }

    async request(endpoint, method = 'GET', body = null) {
        // Check token expiration before making request
        if (!await this.refreshTokenIfNeeded()) {
            return { success: false, message: 'Authentication required' };
        }

        const options = {
            method,
            headers: this.getHeaders()
        };

        if (body) {
            options.body = JSON.stringify(body);
        }

        try {
            const response = await fetch(`${API_BASE}${endpoint}`, options);
            
            // Check if response is ok
            if (!response.ok) {
                // Handle HTTP errors
                if (response.status === 401) {
                    // Token expired or invalid
                    localStorage.clear();
                    window.location.href = 'index.html';
                    return { success: false, message: 'Authentication required' };
                }
                
                // Try to parse error message
                let errorData;
                try {
                    errorData = await response.json();
                } catch {
                    errorData = { success: false, message: `HTTP ${response.status}: ${response.statusText}` };
                }
                
                return errorData;
            }
            
            // Parse JSON response
            const data = await response.json();
            
            if (data && data.success === false && response.status === 401) {
                // Token expired or invalid
                localStorage.clear();
                window.location.href = 'index.html';
                return null;
            }

            return data;
        } catch (error) {
            console.error('API Error:', error);
            return { success: false, message: error.message || 'Network error occurred' };
        }
    }

    // Request with caching
    async requestWithCache(endpoint, ttl = this.cacheTimeout) {
        const cacheKey = endpoint;
        const cached = this.cache.get(cacheKey);
        
        if (cached && Date.now() - cached.timestamp < ttl) {
            return cached.data;
        }

        const data = await this.request(endpoint);
        this.cache.set(cacheKey, {
            data,
            timestamp: Date.now()
        });

        return data;
    }

    // Invalidate cache
    invalidateCache(pattern = null) {
        if (!pattern) {
            this.cache.clear();
        } else {
            for (const key of this.cache.keys()) {
                if (key.includes(pattern)) {
                    this.cache.delete(key);
                }
            }
        }
    }

    // Clear all cache
    clearCache() {
        this.cache.clear();
    }

    // Auth
    async login(username, password) {
        return this.request('/login', 'POST', { username, password });
    }

    // Clients
    async getClients() {
        return this.requestWithCache('/clients');
    }

    async createClient(name) {
        const result = await this.request('/clients', 'POST', { name });
        this.invalidateCache('/clients');
        return result;
    }

    async createClientWithInbound(name, email, inbound) {
        const result = await this.request('/clients', 'POST', { 
            name, 
            email, 
            inbound 
        });
        this.invalidateCache('/clients');
        return result;
    }

    async deleteClient(name) {
        const result = await this.request(`/clients/${name}`, 'DELETE');
        this.invalidateCache('/clients');
        return result;
    }

    async getClientStatus(name) {
        return this.request(`/clients/${name}/status`);
    }

    async exportClient(name) {
        // Create a blob URL with proper authorization
        const response = await fetch(`${API_BASE}/clients/${name}/export`, {
            method: 'GET',
            headers: this.getHeaders()
        });
        
        if (response.ok) {
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${name}.ovpn`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
        } else {
            const error = await response.json();
            throw new Error(error.message || 'Export failed');
        }
    }

    async getOrphanedClients() {
        return this.request('/clients/orphaned');
    }

    async cleanupOrphaned() {
        return this.request('/clients/cleanup-orphaned', 'POST');
    }

    // Config
    async getConfig() {
        return this.request('/config');
    }

    async updateConfig(config) {
        return this.request('/config', 'PUT', config);
    }

    // Monitoring
    async getStats() {
        return this.request('/monitor/stats');
    }

    async getServerStatus() {
        return this.request('/server/status');
    }

    async restartServer() {
        return this.request('/server/restart', 'POST');
    }

    async startServer() {
        return this.request('/server/start', 'POST');
    }

    async stopServer() {
        return this.request('/server/stop', 'POST');
    }

    // Traffic Analytics
    async getCurrentTraffic() {
        return this.request('/analytics/current-traffic');
    }

    async getClientTraffic(clientName) {
        return this.request(`/analytics/client-traffic/${clientName}`);
    }

    // Settings
    async getAdminSettings() {
        return this.request('/settings/admin');
    }

    async changePassword(oldPassword, newPassword) {
        return this.request('/settings/admin/password', 'PUT', {
            old_password: oldPassword,
            new_password: newPassword
        });
    }

    async getTelegramSettings() {
        return this.request('/settings/telegram');
    }

    async updateTelegramSettings(settings) {
        return this.request('/settings/telegram', 'PUT', settings);
    }

    async getServerSettings() {
        return this.request('/settings/server');
    }

    async updateServerSettings(settings) {
        return this.request('/settings/server', 'PUT', settings);
    }

    // Backup
    async exportBackup() {
        // Create a blob URL with proper authorization
        const response = await fetch(`${API_BASE}/backup/export`, {
            method: 'GET',
            headers: this.getHeaders()
        });
        
        if (response.ok) {
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `irangate-backup-${new Date().toISOString().split('T')[0]}.json`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
        } else {
            const error = await response.json();
            throw new Error(error.message || 'Backup export failed');
        }
    }

    async importBackup(file) {
        const formData = new FormData();
        formData.append('backup', file);

        const token = localStorage.getItem('token');
        const response = await fetch(`${API_BASE}/backup/import`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            body: formData
        });

        return response.json();
    }
}

// Global API instance
const api = new API();

// Helper functions
function checkAuth() {
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = 'index.html';
        return false;
    }
    return true;
}

function logout() {
    localStorage.clear();
    window.location.href = 'index.html';
}

function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    let bgColor = 'bg-green-500';
    let themeColor = '';
    
    switch (type) {
        case 'success':
            bgColor = 'bg-green-500';
            themeColor = getComputedStyle(document.documentElement).getPropertyValue('--success') || '#10b981';
            break;
        case 'error':
            bgColor = 'bg-red-500';
            themeColor = getComputedStyle(document.documentElement).getPropertyValue('--error') || '#ef4444';
            break;
        case 'warning':
            bgColor = 'bg-yellow-500';
            themeColor = getComputedStyle(document.documentElement).getPropertyValue('--warning') || '#f59e0b';
            break;
        case 'info':
            bgColor = 'bg-blue-500';
            themeColor = getComputedStyle(document.documentElement).getPropertyValue('--accent-color') || '#3b82f6';
            break;
        default:
            bgColor = 'bg-green-500';
            themeColor = getComputedStyle(document.documentElement).getPropertyValue('--success') || '#10b981';
    }
    
    toast.className = `fixed top-4 right-4 px-6 py-3 rounded-lg shadow-lg text-white z-50 ${bgColor}`;
    toast.style.backgroundColor = themeColor;
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(() => {
        toast.remove();
    }, 3000);
}

function formatBytes(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

function formatDate(dateString) {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleString();
}

function formatUptime(seconds) {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) return `${days}d ${hours}h`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
}

