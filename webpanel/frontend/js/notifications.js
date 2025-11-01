// Enhanced Notification System for IranGate Web Panel

class NotificationManager {
    constructor() {
        this.queue = [];
        this.history = [];
        this.maxHistory = 50;
        this.container = null;
        this.init();
    }

    init() {
        // Create notification container
        this.container = document.createElement('div');
        this.container.id = 'notificationContainer';
        this.container.className = 'fixed top-4 right-4 z-50 space-y-2';
        this.container.style.maxWidth = '400px';
        document.body.appendChild(this.container);
    }

    show(message, type = 'success', options = {}) {
        const notification = {
            id: Date.now() + Math.random(),
            message,
            type,
            duration: options.duration || 5000,
            action: options.action,
            actionText: options.actionText || 'Action',
            progress: options.progress !== undefined ? options.progress : false,
            dismissible: options.dismissible !== false
        };

        this.queue.push(notification);
        this.history.unshift({
            ...notification,
            timestamp: new Date()
        });

        if (this.history.length > this.maxHistory) {
            this.history.pop();
        }

        this.render(notification);

        // Auto dismiss
        if (notification.duration > 0 && !notification.progress) {
            setTimeout(() => this.dismiss(notification.id), notification.duration);
        }

        return notification.id;
    }

    render(notification) {
        const toast = document.createElement('div');
        toast.id = `toast-${notification.id}`;
        toast.className = 'notification-toast transform transition-all duration-300 ease-out';
        
        const typeConfig = {
            success: {
                bg: 'bg-green-600',
                bgColor: getComputedStyle(document.documentElement).getPropertyValue('--success') || '#10b981',
                icon: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>'
            },
            error: {
                bg: 'bg-red-600',
                bgColor: getComputedStyle(document.documentElement).getPropertyValue('--error') || '#ef4444',
                icon: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>'
            },
            warning: {
                bg: 'bg-yellow-600',
                bgColor: getComputedStyle(document.documentElement).getPropertyValue('--warning') || '#f59e0b',
                icon: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>'
            },
            info: {
                bg: 'bg-blue-600',
                bgColor: getComputedStyle(document.documentElement).getPropertyValue('--accent-color') || '#3b82f6',
                icon: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>'
            }
        };

        const config = typeConfig[notification.type] || typeConfig.info;

        toast.innerHTML = `
            <div class="${config.bg} text-white rounded-lg shadow-2xl p-4 backdrop-filter backdrop-blur-sm flex items-start gap-3 min-w-[300px] max-w-[400px]" style="background-color: ${config.bgColor} !important;">
                <svg class="w-6 h-6 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    ${config.icon}
                </svg>
                <div class="flex-1">
                    <p class="text-sm font-medium">${notification.message}</p>
                    ${notification.action ? `
                        <button onclick="notificationManager.handleAction('${notification.id}')" 
                                class="mt-2 text-xs font-semibold underline hover:no-underline">
                            ${notification.actionText}
                        </button>
                    ` : ''}
                    ${notification.progress ? `
                        <div class="mt-2 bg-white bg-opacity-20 rounded-full h-1 overflow-hidden">
                            <div id="progress-${notification.id}" class="bg-white h-full transition-all duration-300" style="width: 0%"></div>
                        </div>
                    ` : ''}
                </div>
                ${notification.dismissible ? `
                    <button onclick="notificationManager.dismiss('${notification.id}')" 
                            class="flex-shrink-0 text-white hover:text-gray-200 transition">
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                    </button>
                ` : ''}
            </div>
        `;

        // Add with animation
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(100%)';
        this.container.appendChild(toast);

        // Trigger animation
        setTimeout(() => {
            toast.style.opacity = '1';
            toast.style.transform = 'translateX(0)';
        }, 10);
    }

    dismiss(id) {
        const toast = document.getElementById(`toast-${id}`);
        if (!toast) return;

        toast.style.opacity = '0';
        toast.style.transform = 'translateX(100%)';

        setTimeout(() => {
            toast.remove();
            this.queue = this.queue.filter(n => n.id !== id);
        }, 300);
    }

    dismissAll() {
        this.queue.forEach(n => this.dismiss(n.id));
    }

    handleAction(id) {
        const notification = this.queue.find(n => n.id === id);
        if (notification && notification.action) {
            notification.action();
            this.dismiss(id);
        }
    }

    updateProgress(id, percent) {
        const progressBar = document.getElementById(`progress-${id}`);
        if (progressBar) {
            progressBar.style.width = `${percent}%`;
        }

        if (percent >= 100) {
            setTimeout(() => this.dismiss(id), 1000);
        }
    }

    success(message, options = {}) {
        return this.show(message, 'success', options);
    }

    error(message, options = {}) {
        return this.show(message, 'error', options);
    }

    warning(message, options = {}) {
        return this.show(message, 'warning', options);
    }

    info(message, options = {}) {
        return this.show(message, 'info', options);
    }

    // Show notification with progress bar
    withProgress(message, type = 'info') {
        return this.show(message, type, { progress: true, duration: 0 });
    }

    // Show notification with action button
    withAction(message, actionText, actionCallback, type = 'info') {
        return this.show(message, type, {
            action: actionCallback,
            actionText: actionText,
            duration: 10000 // Longer duration for actions
        });
    }

    // Get notification history
    getHistory() {
        return this.history;
    }

    // Clear history
    clearHistory() {
        this.history = [];
    }
}

// Global notification manager instance
const notificationManager = new NotificationManager();

// Override the simple showToast function with enhanced version
function showToast(message, type = 'success') {
    notificationManager.show(message, type);
}

