// Telegram notification management

let telegramConfig = null;
let notificationSettings = null;
let autoActions = [];

// Load Telegram status
async function loadTelegramStatus() {
    try {
        const response = await fetch('/api/notifications/telegram/status', {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            telegramConfig = result.data;
            updateTelegramStatusDisplay(result.data);
        } else {
            showError('Failed to load Telegram status: ' + result.message);
        }
    } catch (error) {
        console.error('Error loading Telegram status:', error);
        showError('Failed to load Telegram status');
    }
}

// Update Telegram status display
function updateTelegramStatusDisplay(status) {
    const statusElement = document.getElementById('telegramStatus');
    const statusDot = statusElement.querySelector('.w-3');
    const statusText = statusElement.querySelector('.text-sm');
    
    if (status.enabled && status.configured) {
        if (status.connectivity === 'success') {
            statusDot.className = 'inline-block w-3 h-3 bg-green-400 rounded-full mr-2';
            statusText.textContent = 'Connected';
            statusText.className = 'text-sm text-green-600';
        } else {
            statusDot.className = 'inline-block w-3 h-3 bg-yellow-400 rounded-full mr-2';
            statusText.textContent = 'Configured but not connected';
            statusText.className = 'text-sm text-yellow-600';
        }
    } else {
        statusDot.className = 'inline-block w-3 h-3 bg-gray-400 rounded-full mr-2';
        statusText.textContent = 'Not configured';
        statusText.className = 'text-sm text-gray-600';
    }
}

// Save Telegram configuration
async function saveTelegramConfig(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const config = {
        bot_token: formData.get('bot_token'),
        chat_id: formData.get('chat_id'),
        enabled: formData.has('enabled')
    };
    
    try {
        const response = await fetch('/api/notifications/telegram/configure', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${getToken()}`
            },
            body: JSON.stringify(config)
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess('Telegram configuration saved successfully');
            loadTelegramStatus();
        } else {
            showError('Failed to save Telegram configuration: ' + result.message);
        }
    } catch (error) {
        console.error('Error saving Telegram config:', error);
        showError('Failed to save Telegram configuration');
    }
}

// Test Telegram connection
async function testTelegram() {
    try {
        const response = await fetch('/api/notifications/telegram/test', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess('Test message sent successfully!');
        } else {
            showError('Failed to send test message: ' + result.message);
        }
    } catch (error) {
        console.error('Error testing Telegram:', error);
        showError('Failed to send test message');
    }
}

// Load notification settings
async function loadNotificationSettings() {
    try {
        const response = await fetch('/api/notifications/telegram/settings', {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            notificationSettings = result.data;
            updateNotificationSettingsForm(result.data);
        } else {
            console.error('Failed to load notification settings:', result.message);
        }
    } catch (error) {
        console.error('Error loading notification settings:', error);
    }
}

// Update notification settings form
function updateNotificationSettingsForm(settings) {
    document.getElementById('clientExpiring').checked = settings.client_expiring;
    document.getElementById('quotaExceeded').checked = settings.client_quota_exceeded;
    document.getElementById('newClientCreated').checked = settings.new_client_created;
    document.getElementById('resourceWarning').checked = settings.server_resource_warn;
    document.getElementById('openvpnDown').checked = settings.openvpn_down;
    document.getElementById('openvpnRestarted').checked = settings.openvpn_restarted;
    document.getElementById('expiringDays').value = settings.expiring_days || 7;
}

// Save notification settings
async function saveNotificationSettings(event) {
    event.preventDefault();
    
    const settings = {
        client_expiring: document.getElementById('clientExpiring').checked,
        client_quota_exceeded: document.getElementById('quotaExceeded').checked,
        new_client_created: document.getElementById('newClientCreated').checked,
        server_resource_warn: document.getElementById('resourceWarning').checked,
        openvpn_down: document.getElementById('openvpnDown').checked,
        openvpn_restarted: document.getElementById('openvpnRestarted').checked,
        expiring_days: parseInt(document.getElementById('expiringDays').value)
    };
    
    try {
        const response = await fetch('/api/notifications/telegram/settings', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${getToken()}`
            },
            body: JSON.stringify(settings)
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess('Notification settings saved successfully');
        } else {
            showError('Failed to save notification settings: ' + result.message);
        }
    } catch (error) {
        console.error('Error saving notification settings:', error);
        showError('Failed to save notification settings');
    }
}

// Load auto-actions
async function loadAutoActions() {
    try {
        const response = await fetch('/api/autoactions', {
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            autoActions = result.data;
            displayAutoActions(result.data);
        } else {
            console.error('Failed to load auto-actions:', result.message);
        }
    } catch (error) {
        console.error('Error loading auto-actions:', error);
    }
}

// Display auto-actions
function displayAutoActions(actions) {
    const container = document.getElementById('autoActionsList');
    
    if (actions.length === 0) {
        container.innerHTML = '<p class="text-gray-500 text-center py-4">No auto-actions configured</p>';
        return;
    }
    
    container.innerHTML = actions.map(action => `
        <div class="border border-gray-200 rounded-lg p-4">
            <div class="flex items-center justify-between">
                <div>
                    <h3 class="font-semibold text-gray-800">${action.name}</h3>
                    <p class="text-sm text-gray-600">${action.description || 'No description'}</p>
                    <div class="flex items-center space-x-4 mt-2">
                        <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${action.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}">
                            ${action.enabled ? 'Enabled' : 'Disabled'}
                        </span>
                        <span class="text-xs text-gray-500">Trigger: ${action.trigger}</span>
                        <span class="text-xs text-gray-500">Action: ${action.action}</span>
                    </div>
                </div>
                <div class="flex space-x-2">
                    <button onclick="testAutoAction('${action.id}')" 
                            class="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded text-sm">
                        Test
                    </button>
                    <button onclick="editAutoAction('${action.id}')" 
                            class="bg-yellow-600 hover:bg-yellow-700 text-white px-3 py-1 rounded text-sm">
                        Edit
                    </button>
                    <button onclick="deleteAutoAction('${action.id}')" 
                            class="bg-red-600 hover:bg-red-700 text-white px-3 py-1 rounded text-sm">
                        Delete
                    </button>
                </div>
            </div>
        </div>
    `).join('');
}

// Show auto-action modal
function showAutoActionModal() {
    document.getElementById('autoActionModal').classList.remove('hidden');
}

// Hide auto-action modal
function hideAutoActionModal() {
    document.getElementById('autoActionModal').classList.add('hidden');
    document.getElementById('autoActionForm').reset();
}

// Save auto-action
async function saveAutoAction(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const action = {
        name: formData.get('name'),
        description: formData.get('description'),
        trigger: formData.get('trigger'),
        action: formData.get('action'),
        enabled: formData.has('enabled')
    };
    
    try {
        const response = await fetch('/api/autoactions', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${getToken()}`
            },
            body: JSON.stringify(action)
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess('Auto-action created successfully');
            hideAutoActionModal();
            loadAutoActions();
        } else {
            showError('Failed to create auto-action: ' + result.message);
        }
    } catch (error) {
        console.error('Error creating auto-action:', error);
        showError('Failed to create auto-action');
    }
}

// Test auto-action
async function testAutoAction(actionId) {
    try {
        const response = await fetch(`/api/autoactions/${actionId}/test`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess('Auto-action test completed: ' + result.data.message);
        } else {
            showError('Auto-action test failed: ' + result.message);
        }
    } catch (error) {
        console.error('Error testing auto-action:', error);
        showError('Failed to test auto-action');
    }
}

// Delete auto-action
async function deleteAutoAction(actionId) {
    if (!confirm('Are you sure you want to delete this auto-action?')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/autoactions/${actionId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${getToken()}`
            }
        });
        
        const result = await response.json();
        
        if (result.success) {
            showSuccess('Auto-action deleted successfully');
            loadAutoActions();
        } else {
            showError('Failed to delete auto-action: ' + result.message);
        }
    } catch (error) {
        console.error('Error deleting auto-action:', error);
        showError('Failed to delete auto-action');
    }
}

// Edit auto-action
function editAutoAction(actionId) {
    const action = autoActions.find(a => a.id === actionId);
    if (!action) {
        showError('Auto-action not found');
        return;
    }
    
    // Populate form with existing data
    document.getElementById('actionName').value = action.name;
    document.getElementById('actionDescription').value = action.description || '';
    document.getElementById('actionTrigger').value = action.trigger;
    document.getElementById('actionAction').value = action.action;
    document.getElementById('actionEnabled').checked = action.enabled;
    
    showAutoActionModal();
}

// Utility functions
function getToken() {
    return localStorage.getItem('authToken');
}

function showSuccess(message) {
    // Create success notification
    const notification = document.createElement('div');
    notification.className = 'fixed top-4 right-4 bg-green-500 text-white px-6 py-3 rounded-lg shadow-lg z-50';
    notification.textContent = message;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 3000);
}

function showError(message) {
    // Create error notification
    const notification = document.createElement('div');
    notification.className = 'fixed top-4 right-4 bg-red-500 text-white px-6 py-3 rounded-lg shadow-lg z-50';
    notification.textContent = message;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 5000);
}
