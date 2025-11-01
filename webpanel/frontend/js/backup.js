// Backup and maintenance management

let backupSchedules = [];
let availableBackups = [];

// Load backup schedules
async function loadBackupSchedules() {
    try {
        const response = await fetch('/api/backup/schedules', {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            backupSchedules = result.data;
            displayBackupSchedules(result.data);
        } else {
            console.error('Failed to load backup schedules:', result.message);
        }
    } catch (error) {
        console.error('Error loading backup schedules:', error);
    }
}

// Display backup schedules
function displayBackupSchedules(schedules) {
    const container = document.getElementById('backupSchedulesList');
    
    if (schedules.length === 0) {
        container.innerHTML = '<p class="text-gray-500 text-center py-4">No backup schedules configured</p>';
        return;
    }
    
    container.innerHTML = schedules.map(schedule => `
        <div class="border border-gray-200 rounded-lg p-4">
            <div class="flex items-center justify-between">
                <div>
                    <h3 class="font-semibold text-gray-800">${schedule.name}</h3>
                    <p class="text-sm text-gray-600">${schedule.description || 'No description'}</p>
                    <div class="flex items-center space-x-4 mt-2">
                        <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${schedule.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}">
                            ${schedule.enabled ? 'Enabled' : 'Disabled'}
                        </span>
                        <span class="text-xs text-gray-500">Schedule: ${schedule.cron_expression}</span>
                        <span class="text-xs text-gray-500">Retention: ${schedule.retention_days} days</span>
                        ${schedule.last_run ? `<span class="text-xs text-gray-500">Last run: ${formatDate(schedule.last_run)}</span>` : ''}
                    </div>
                </div>
                <div class="flex space-x-2">
                    <button onclick="runBackupSchedule('${schedule.id}')" 
                            class="bg-green-600 hover:bg-green-700 text-white px-3 py-1 rounded text-sm">
                        Run Now
                    </button>
                    <button onclick="editBackupSchedule('${schedule.id}')" 
                            class="bg-yellow-600 hover:bg-yellow-700 text-white px-3 py-1 rounded text-sm">
                        Edit
                    </button>
                    <button onclick="deleteBackupSchedule('${schedule.id}')" 
                            class="bg-red-600 hover:bg-red-700 text-white px-3 py-1 rounded text-sm">
                        Delete
                    </button>
                </div>
            </div>
        </div>
    `).join('');
}

// Load available backups
async function loadBackups() {
    try {
        const response = await fetch('/api/backup/list', {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            availableBackups = result.data;
            displayBackups(result.data);
        } else {
            console.error('Failed to load backups:', result.message);
        }
    } catch (error) {
        console.error('Error loading backups:', error);
    }
}

// Display available backups
function displayBackups(backups) {
    const tbody = document.getElementById('backupsTable');
    
    if (backups.length === 0) {
        tbody.innerHTML = '<tr><td colspan="4" class="px-4 py-8 text-center text-gray-500">No backups available</td></tr>';
        return;
    }
    
    tbody.innerHTML = backups.map(backup => `
        <tr>
            <td class="px-4 py-3 text-sm text-gray-800">${backup.filename}</td>
            <td class="px-4 py-3 text-sm text-gray-800">${formatFileSize(backup.size)}</td>
            <td class="px-4 py-3 text-sm text-gray-800">${formatDate(backup.created_at)}</td>
            <td class="px-4 py-3 text-sm text-gray-800">
                <div class="flex space-x-2">
                    <button onclick="downloadBackup('${backup.filename}')" 
                            class="bg-blue-600 hover:bg-blue-700 text-white px-2 py-1 rounded text-xs">
                        Download
                    </button>
                    <button onclick="deleteBackup('${backup.filename}')" 
                            class="bg-red-600 hover:bg-red-700 text-white px-2 py-1 rounded text-xs">
                        Delete
                    </button>
                </div>
            </td>
        </tr>
    `).join('');
}

// Show backup schedule modal
function showBackupScheduleModal() {
    document.getElementById('backupScheduleModal').classList.remove('hidden');
}

// Hide backup schedule modal
function hideBackupScheduleModal() {
    document.getElementById('backupScheduleModal').classList.add('hidden');
    document.getElementById('backupScheduleForm').reset();
}

// Save backup schedule
async function saveBackupSchedule(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const schedule = {
        name: formData.get('name'),
        description: formData.get('description'),
        cron_expression: formData.get('cron_expression'),
        retention_days: parseInt(formData.get('retention_days')),
        backup_path: formData.get('backup_path'),
        enabled: formData.has('enabled'),
        include_paths: [
            '/root/OV-Panel/irangate/database',
            '/etc/openvpn',
            '/root/OV-Panel/irangate/clients'
        ]
    };
    
    try {
        const response = await fetch('/api/backup/schedules', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${getToken()}`
            },
            body: JSON.stringify(schedule)
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess('Backup schedule created successfully');
            hideBackupScheduleModal();
            loadBackupSchedules();
        } else {
            showError('Failed to create backup schedule: ' + result.message);
        }
    } catch (error) {
        console.error('Error creating backup schedule:', error);
        showError('Failed to create backup schedule');
    }
}

// Run backup schedule
async function runBackupSchedule(scheduleId) {
    try {
        const response = await fetch(`/api/backup/schedules/${scheduleId}/run`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess('Backup execution started');
            // Refresh backups after a short delay
            setTimeout(() => {
                loadBackups();
            }, 2000);
        } else {
            showError('Failed to start backup: ' + result.message);
        }
    } catch (error) {
        console.error('Error running backup schedule:', error);
        showError('Failed to start backup');
    }
}

// Delete backup schedule
async function deleteBackupSchedule(scheduleId) {
    if (!confirm('Are you sure you want to delete this backup schedule?')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/backup/schedules/${scheduleId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess('Backup schedule deleted successfully');
            loadBackupSchedules();
        } else {
            showError('Failed to delete backup schedule: ' + result.message);
        }
    } catch (error) {
        console.error('Error deleting backup schedule:', error);
        showError('Failed to delete backup schedule');
    }
}

// Download backup
function downloadBackup(filename) {
    const url = `/api/backup/${filename}/download`;
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    link.style.display = 'none';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}

// Delete backup
async function deleteBackup(filename) {
    if (!confirm(`Are you sure you want to delete backup "${filename}"?`)) {
        return;
    }
    
    try {
        const response = await fetch(`/api/backup/${filename}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess('Backup deleted successfully');
            loadBackups();
        } else {
            showError('Failed to delete backup: ' + result.message);
        }
    } catch (error) {
        console.error('Error deleting backup:', error);
        showError('Failed to delete backup');
    }
}

// Load disk usage
async function loadDiskUsage() {
    try {
        const response = await fetch('/api/maintenance/disk-usage', {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            displayDiskUsage(result.data);
        } else {
            console.error('Failed to load disk usage:', result.message);
        }
    } catch (error) {
        console.error('Error loading disk usage:', error);
        // Display mock data for now
        displayDiskUsage({
            total: 100,
            used: 45,
            free: 55,
            breakdown: [
                { path: '/root/OV-Panel/irangate', size: 2048, percentage: 20 },
                { path: '/var/log', size: 1024, percentage: 10 },
                { path: '/etc/openvpn', size: 512, percentage: 5 }
            ]
        });
    }
}

// Display disk usage
function displayDiskUsage(data) {
    const container = document.getElementById('diskUsageInfo');
    
    const totalGB = data.total || 100;
    const usedGB = data.used || 45;
    const freeGB = data.free || 55;
    const usedPercentage = (usedGB / totalGB) * 100;
    
    container.innerHTML = `
        <div class="bg-gray-50 rounded-lg p-4">
            <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium text-gray-700">Disk Usage</span>
                <span class="text-sm text-gray-600">${usedGB}GB / ${totalGB}GB</span>
            </div>
            <div class="w-full bg-gray-200 rounded-full h-2 mb-4">
                <div class="bg-blue-500 h-2 rounded-full" style="width: ${usedPercentage}%"></div>
            </div>
            <div class="grid grid-cols-2 gap-4 text-sm">
                <div>
                    <span class="text-gray-600">Used:</span>
                    <span class="font-medium">${usedGB}GB</span>
                </div>
                <div>
                    <span class="text-gray-600">Free:</span>
                    <span class="font-medium">${freeGB}GB</span>
                </div>
            </div>
        </div>
        
        <div class="space-y-2">
            <h3 class="text-sm font-medium text-gray-700">Directory Breakdown</h3>
            ${data.breakdown ? data.breakdown.map(item => `
                <div class="flex items-center justify-between text-sm">
                    <span class="text-gray-600">${item.path}</span>
                    <span class="font-medium">${item.size}MB (${item.percentage}%)</span>
                </div>
            `).join('') : '<p class="text-gray-500 text-sm">No breakdown available</p>'}
        </div>
    `;
}

// Cleanup logs
async function cleanupLogs() {
    const days = document.getElementById('logCleanupDays').value;
    
    if (!confirm(`Remove log files older than ${days} days?`)) {
        return;
    }
    
    try {
        const response = await fetch('/api/maintenance/cleanup-logs', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${getToken()}`
            },
            body: JSON.stringify({ days: parseInt(days) })
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess(`Log cleanup completed: ${result.data.cleaned_count} files removed`);
        } else {
            showError('Failed to cleanup logs: ' + result.message);
        }
    } catch (error) {
        console.error('Error cleaning up logs:', error);
        showError('Failed to cleanup logs');
    }
}

// Cleanup history
async function cleanupHistory() {
    const days = document.getElementById('historyCleanupDays').value;
    
    if (!confirm(`Remove connection history older than ${days} days?`)) {
        return;
    }
    
    try {
        const response = await fetch('/api/maintenance/cleanup-history', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${getToken()}`
            },
            body: JSON.stringify({ days: parseInt(days) })
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess(`History cleanup completed: ${result.data.cleaned_count} records removed`);
        } else {
            showError('Failed to cleanup history: ' + result.message);
        }
    } catch (error) {
        console.error('Error cleaning up history:', error);
        showError('Failed to cleanup history');
    }
}

// Optimize database
async function optimizeDatabase() {
    if (!confirm('Optimize database files? This may take a few moments.')) {
        return;
    }
    
    try {
        const response = await fetch('/api/maintenance/optimize-db', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess('Database optimization completed');
        } else {
            showError('Failed to optimize database: ' + result.message);
        }
    } catch (error) {
        console.error('Error optimizing database:', error);
        showError('Failed to optimize database');
    }
}

// Cleanup orphaned certificates
async function cleanupOrphaned() {
    if (!confirm('Remove orphaned certificates? This action cannot be undone.')) {
        return;
    }
    
    try {
        const response = await fetch('/api/clients/cleanup-orphaned', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess(`Orphaned certificates cleanup completed: ${result.data.removed_count} certificates removed`);
        } else {
            showError('Failed to cleanup orphaned certificates: ' + result.message);
        }
    } catch (error) {
        console.error('Error cleaning up orphaned certificates:', error);
        showError('Failed to cleanup orphaned certificates');
    }
}

// Utility functions
function getToken() {
    return localStorage.getItem('authToken');
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString();
}

function showSuccess(message) {
    const notification = document.createElement('div');
    notification.className = 'fixed top-4 right-4 bg-green-500 text-white px-6 py-3 rounded-lg shadow-lg z-50';
    notification.textContent = message;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 3000);
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
