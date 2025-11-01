class SystemResourcesMonitor {
    constructor() {
        this.chart = null;
        this.dataPoints = [];
        this.maxDataPoints = 60; // Keep last 60 data points (2 minutes at 2s intervals)
        this.previousNetwork = null;
        this.isInitialized = false;
    }

    init() {
        if (this.isInitialized) return;
        
        this.setupWebSocket();
        this.setupCharts();
        this.loadInitialData();
        this.isInitialized = true;
    }

    setupWebSocket() {
        if (window.wsClient) {
            window.wsClient.on('system_resources_update', (data) => {
                this.updateResources(data);
            });
        }
    }

    updateResources(resources) {
        // Update CPU
        this.updateCPU(resources.cpu);
        
        // Update Memory
        this.updateMemory(resources.memory);
        
        // Update Disk
        this.updateDisk(resources.disk);
        
        // Update Network
        this.updateNetwork(resources.network);
        
        // Update chart
        this.updateChart(resources);
    }

    updateCPU(cpu) {
        const cpuElement = document.getElementById('cpuUsage');
        const cpuBar = document.getElementById('cpuBar');
        
        if (cpuElement) {
            cpuElement.textContent = `${cpu.usage.toFixed(1)}%`;
        }
        if (cpuBar) {
            cpuBar.style.width = `${cpu.usage}%`;
            cpuBar.className = this.getBarColorClass(cpu.usage, 'blue');
        }
    }

    updateMemory(memory) {
        const memoryElement = document.getElementById('memoryUsage');
        const memoryBar = document.getElementById('memoryBar');
        
        if (memoryElement) {
            const memoryGB = (memory.used / (1024 * 1024 * 1024)).toFixed(1);
            memoryElement.textContent = `${memoryGB} GB`;
        }
        if (memoryBar) {
            memoryBar.style.width = `${memory.usage}%`;
            memoryBar.className = this.getBarColorClass(memory.usage, 'green');
        }
    }

    updateDisk(disk) {
        const diskElement = document.getElementById('diskUsage');
        const diskBar = document.getElementById('diskBar');
        
        if (diskElement) {
            const diskGB = (disk.used / (1024 * 1024 * 1024)).toFixed(1);
            diskElement.textContent = `${diskGB} GB`;
        }
        if (diskBar) {
            diskBar.style.width = `${disk.usage}%`;
            diskBar.className = this.getBarColorClass(disk.usage, 'orange');
        }
    }

    updateNetwork(network) {
        const networkElement = document.getElementById('networkTraffic');
        const networkBar = document.getElementById('networkBar');
        
        if (networkElement && this.previousNetwork) {
            const speed = this.calculateNetworkSpeed(network);
            networkElement.textContent = speed;
        }
        
        if (networkBar) {
            // Calculate network usage percentage (simplified)
            const usage = Math.min(100, (network.bytes_received + network.bytes_sent) / 1024 / 1024); // MB
            networkBar.style.width = `${usage}%`;
            networkBar.className = this.getBarColorClass(usage, 'purple');
        }
        
        this.previousNetwork = network;
    }

    calculateNetworkSpeed(current) {
        if (!this.previousNetwork) return "0 KB/s";
        
        const timeDiff = 2; // 2 seconds
        const bytesDiff = (current.bytes_received + current.bytes_sent) - 
                        (this.previousNetwork.bytes_received + this.previousNetwork.bytes_sent);
        
        const speedKBps = (bytesDiff / 1024) / timeDiff;
        
        if (speedKBps < 1024) {
            return `${speedKBps.toFixed(1)} KB/s`;
        } else {
            return `${(speedKBps / 1024).toFixed(1)} MB/s`;
        }
    }

    getBarColorClass(usage, color) {
        const baseClass = `h-2 rounded-full transition-all duration-300`;
        const colorClass = `bg-${color}-500`;
        
        if (usage > 90) {
            return `${baseClass} ${colorClass} bg-red-500`;
        } else if (usage > 70) {
            return `${baseClass} ${colorClass} bg-yellow-500`;
        } else {
            return `${baseClass} ${colorClass}`;
        }
    }

    updateChart(resources) {
        if (!this.chart) return;
        
        this.dataPoints.push({
            timestamp: new Date(),
            cpu: resources.cpu.usage,
            memory: resources.memory.usage,
            disk: resources.disk.usage
        });
        
        // Keep only last N data points
        if (this.dataPoints.length > this.maxDataPoints) {
            this.dataPoints.shift();
        }
        
        this.chart.update();
    }

    setupCharts() {
        const ctx = document.getElementById('systemResourcesChart');
        if (!ctx) return;
        
        this.chart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [
                    {
                        label: 'CPU %',
                        data: [],
                        borderColor: 'rgb(59, 130, 246)',
                        backgroundColor: 'rgba(59, 130, 246, 0.1)',
                        tension: 0.4,
                        fill: false
                    },
                    {
                        label: 'Memory %',
                        data: [],
                        borderColor: 'rgb(34, 197, 94)',
                        backgroundColor: 'rgba(34, 197, 94, 0.1)',
                        tension: 0.4,
                        fill: false
                    },
                    {
                        label: 'Disk %',
                        data: [],
                        borderColor: 'rgb(249, 115, 22)',
                        backgroundColor: 'rgba(249, 115, 22, 0.1)',
                        tension: 0.4,
                        fill: false
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100,
                        title: {
                            display: true,
                            text: 'Usage %'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    }
                },
                plugins: {
                    legend: {
                        position: 'top',
                    },
                    tooltip: {
                        mode: 'index',
                        intersect: false,
                    }
                },
                interaction: {
                    mode: 'nearest',
                    axis: 'x',
                    intersect: false
                }
            }
        });
    }

    async loadInitialData() {
        try {
            const response = await fetch('/api/system/resources', {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });
            
            if (response.ok) {
                const data = await response.json();
                if (data.success) {
                    this.updateResources(data.data);
                }
            }
        } catch (error) {
            console.error('Failed to load initial system resources:', error);
        }
    }

    // Method to get current system resources data
    getCurrentData() {
        return this.dataPoints[this.dataPoints.length - 1] || null;
    }

    // Method to get historical data
    getHistoricalData() {
        return this.dataPoints;
    }

    // Method to clear historical data
    clearHistoricalData() {
        this.dataPoints = [];
        if (this.chart) {
            this.chart.update();
        }
    }

    // Method to export data
    exportData() {
        const data = {
            timestamp: new Date().toISOString(),
            dataPoints: this.dataPoints,
            summary: this.getCurrentData()
        };
        
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `system-resources-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    if (document.getElementById('systemResourcesChart')) {
        window.systemResourcesMonitor = new SystemResourcesMonitor();
        window.systemResourcesMonitor.init();
    }
});

// Export for global access
window.SystemResourcesMonitor = SystemResourcesMonitor;
