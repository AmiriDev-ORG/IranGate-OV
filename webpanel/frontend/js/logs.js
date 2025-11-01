// Log viewer functionality

let logStream = null;
let autoRefreshInterval = null;
let currentLogType = 'openvpn';

// Initialize log viewer
function initializeLogViewer() {
    const logTypeSelect = document.getElementById('logType');
    const refreshBtn = document.getElementById('refreshLogsBtn');
    const searchBtn = document.getElementById('searchLogsBtn');
    const tailBtn = document.getElementById('tailLogsBtn');
    const streamBtn = document.getElementById('streamLogsBtn');
    const autoRefreshCheckbox = document.getElementById('autoRefreshLogs');
    const clearBtn = document.getElementById('clearLogsBtn');
    const downloadBtn = document.getElementById('downloadLogsBtn');

    // Event listeners
    logTypeSelect.addEventListener('change', function() {
        currentLogType = this.value;
        updateLogStatus('Log type changed to ' + this.value);
    });

    refreshBtn.addEventListener('click', function() {
        loadLogs();
    });

    searchBtn.addEventListener('click', function() {
        searchLogs();
    });

    tailBtn.addEventListener('click', function() {
        tailLogs();
    });

    streamBtn.addEventListener('click', function() {
        toggleLogStream();
    });

    autoRefreshCheckbox.addEventListener('change', function() {
        toggleAutoRefresh(this.checked);
    });

    clearBtn.addEventListener('click', function() {
        clearLogOutput();
    });

    downloadBtn.addEventListener('click', function() {
        downloadLogs();
    });

    // Load initial logs
    tailLogs();
}

// Load logs with search
async function loadLogs() {
    const searchText = document.getElementById('logSearch').value;
    const level = document.getElementById('logLevel').value;
    const limit = document.getElementById('logLines').value;

    try {
        const params = new URLSearchParams({
            limit: limit
        });

        if (searchText) params.append('search', searchText);
        if (level) params.append('level', level);

        const response = await fetch(`/api/logs/${currentLogType}?${params}`, {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });

        const result = await response.json();

        if (result.success) {
            displayLogs(result.data.entries);
        } else {
            showError('Failed to load logs: ' + result.message);
        }
    } catch (error) {
        console.error('Error loading logs:', error);
        showError('Failed to load logs');
    }
}

// Search logs
async function searchLogs() {
    const searchText = document.getElementById('logSearch').value;
    const level = document.getElementById('logLevel').value;
    const limit = document.getElementById('logLines').value;

    if (!searchText) {
        showError('Please enter search text');
        return;
    }

    try {
        const searchQuery = {
            log_type: currentLogType,
            search_text: searchText,
            level: level,
            limit: parseInt(limit)
        };

        const response = await fetch('/api/logs/search', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${getToken()}`
            },
            body: JSON.stringify(searchQuery)
        });

        const result = await response.json();

        if (result.success) {
            displayLogs(result.data.entries);
            updateLogStatus(`Found ${result.data.entries.length} matching entries`);
        } else {
            showError('Failed to search logs: ' + result.message);
        }
    } catch (error) {
        console.error('Error searching logs:', error);
        showError('Failed to search logs');
    }
}

// Tail logs (get recent entries)
async function tailLogs() {
    const lines = document.getElementById('logLines').value;

    try {
        const params = new URLSearchParams({
            type: currentLogType,
            lines: lines
        });

        const response = await fetch(`/api/logs/tail?${params}`, {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });

        const result = await response.json();

        if (result.success) {
            displayLogs(result.data);
            updateLogStatus(`Loaded last ${lines} lines`);
        } else {
            showError('Failed to tail logs: ' + result.message);
        }
    } catch (error) {
        console.error('Error tailing logs:', error);
        showError('Failed to tail logs');
    }
}

// Toggle log streaming
function toggleLogStream() {
    const streamBtn = document.getElementById('streamLogsBtn');
    
    if (logStream) {
        // Stop streaming
        logStream.close();
        logStream = null;
        streamBtn.textContent = 'Stream';
        streamBtn.className = 'bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-lg';
        updateLogStatus('Log streaming stopped');
    } else {
        // Start streaming
        startLogStream();
        streamBtn.textContent = 'Stop';
        streamBtn.className = 'bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-lg';
        updateLogStatus('Log streaming started');
    }
}

// Start log streaming
function startLogStream() {
    const params = new URLSearchParams({
        type: currentLogType
    });

    const eventSource = new EventSource(`/api/logs/stream?${params}`);
    
    eventSource.onopen = function() {
    };

    eventSource.onmessage = function(event) {
        try {
            const logs = JSON.parse(event.data);
            if (Array.isArray(logs)) {
                displayLogs(logs, true); // Append to existing logs
            }
        } catch (error) {
            console.error('Error parsing stream data:', error);
        }
    };

    eventSource.onerror = function(error) {
        console.error('Log stream error:', error);
        updateLogStatus('Log stream error');
        eventSource.close();
        logStream = null;
    };

    logStream = eventSource;
}

// Display logs
function displayLogs(logs, append = false) {
    const logOutput = document.getElementById('logOutput');
    
    if (!append) {
        logOutput.innerHTML = '';
    }

    if (!logs || logs.length === 0) {
        logOutput.innerHTML = '<div class="text-gray-500">No log entries found</div>';
        return;
    }

    logs.forEach(log => {
        const logEntry = document.createElement('div');
        logEntry.className = 'mb-1';
        
        const timestamp = new Date(log.timestamp).toLocaleString();
        const levelClass = getLogLevelClass(log.level);
        
        logEntry.innerHTML = `
            <span class="text-gray-400">[${timestamp}]</span>
            <span class="${levelClass}">[${log.level.toUpperCase()}]</span>
            <span class="text-gray-300">${escapeHtml(log.message)}</span>
        `;
        
        logOutput.appendChild(logEntry);
    });

    // Scroll to bottom
    logOutput.scrollTop = logOutput.scrollHeight;
}

// Get log level CSS class
function getLogLevelClass(level) {
    switch (level.toLowerCase()) {
        case 'error':
            return 'text-red-400 font-bold';
        case 'warn':
        case 'warning':
            return 'text-yellow-400 font-medium';
        case 'info':
            return 'text-blue-400';
        case 'debug':
            return 'text-gray-400';
        default:
            return 'text-gray-300';
    }
}

// Toggle auto-refresh
function toggleAutoRefresh(enabled) {
    if (autoRefreshInterval) {
        clearInterval(autoRefreshInterval);
        autoRefreshInterval = null;
    }

    if (enabled) {
        autoRefreshInterval = setInterval(() => {
            tailLogs();
        }, 5000); // Refresh every 5 seconds
        updateLogStatus('Auto-refresh enabled (5s)');
    } else {
        updateLogStatus('Auto-refresh disabled');
    }
}

// Clear log output
function clearLogOutput() {
    const logOutput = document.getElementById('logOutput');
    logOutput.innerHTML = '<div class="text-gray-500">Log output cleared</div>';
    updateLogStatus('Log output cleared');
}

// Download logs
async function downloadLogs() {
    try {
        const params = new URLSearchParams({
            type: currentLogType,
            lines: document.getElementById('logLines').value
        });

        const response = await fetch(`/api/logs/tail?${params}`, {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });

        const result = await response.json();

        if (result.success) {
            // Create downloadable content
            const logs = result.data;
            const content = logs.map(log => {
                const timestamp = new Date(log.timestamp).toISOString();
                return `[${timestamp}] [${log.level.toUpperCase()}] ${log.message}`;
            }).join('\n');

            // Create and download file
            const blob = new Blob([content], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.href = url;
            link.download = `${currentLogType}_logs_${new Date().toISOString().split('T')[0]}.txt`;
            link.style.display = 'none';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            URL.revokeObjectURL(url);

            updateLogStatus('Logs downloaded successfully');
        } else {
            showError('Failed to download logs: ' + result.message);
        }
    } catch (error) {
        console.error('Error downloading logs:', error);
        showError('Failed to download logs');
    }
}

// Update log status
function updateLogStatus(message) {
    const statusElement = document.getElementById('logStatus');
    statusElement.textContent = message;
    
    // Clear status after 3 seconds
    setTimeout(() => {
        statusElement.textContent = 'Ready';
    }, 3000);
}

// Utility functions
function getToken() {
    return localStorage.getItem('authToken');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showError(message) {
    const notification = document.createElement('div');
    notification.className = 'fixed top-4 right-4 bg-red-500 text-white px-6 py-3 rounded-lg shadow-lg z-50';
    notification.textContent = message;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 5000);
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Only initialize if we're on the logs tab
    if (document.getElementById('logsContent')) {
        initializeLogViewer();
    }
});

// Cleanup when page is unloaded
window.addEventListener('beforeunload', function() {
    if (logStream) {
        logStream.close();
    }
    if (autoRefreshInterval) {
        clearInterval(autoRefreshInterval);
    }
});
