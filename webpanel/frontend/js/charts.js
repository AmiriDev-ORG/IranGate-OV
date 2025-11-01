// Chart.js Manager for IranGate Web Panel
// Handles all chart creation and updates

class ChartManager {
    constructor() {
        this.charts = {};
        this.defaultColors = {
            cyan: 'rgba(100, 255, 218, 0.8)',
            blue: 'rgba(0, 212, 255, 0.8)',
            purple: 'rgba(147, 51, 234, 0.8)',
            green: 'rgba(16, 185, 129, 0.8)',
            yellow: 'rgba(245, 158, 11, 0.8)',
            red: 'rgba(239, 68, 68, 0.8)'
        };
        
        // Chart.js default config for dark theme
        Chart.defaults.color = getComputedStyle(document.documentElement).getPropertyValue('--text-secondary') || '#8892b0';
        Chart.defaults.borderColor = getComputedStyle(document.documentElement).getPropertyValue('--border-color') || 'rgba(30, 58, 95, 0.3)';
        Chart.defaults.backgroundColor = getComputedStyle(document.documentElement).getPropertyValue('--card-bg') || 'rgba(46, 80, 144, 0.1)';
    }

    // Traffic Chart (Line Chart)
    createTrafficChart(canvasId, data) {
        const ctx = document.getElementById(canvasId);
        if (!ctx) return;

        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
        }

        this.charts[canvasId] = new Chart(ctx, {
            type: 'line',
            data: {
                labels: data.labels || [],
                datasets: [
                    {
                        label: 'Upload (MB)',
                        data: data.upload || [],
                        borderColor: this.defaultColors.cyan,
                        backgroundColor: getComputedStyle(document.documentElement).getPropertyValue('--accent-color') + '20' || 'rgba(100, 255, 218, 0.1)',
                        tension: 0.4,
                        fill: true
                    },
                    {
                        label: 'Download (MB)',
                        data: data.download || [],
                        borderColor: this.defaultColors.blue,
                        backgroundColor: getComputedStyle(document.documentElement).getPropertyValue('--accent-color') + '20' || 'rgba(0, 212, 255, 0.1)',
                        tension: 0.4,
                        fill: true
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                interaction: {
                    intersect: false,
                    mode: 'index'
                },
                plugins: {
                    legend: {
                        labels: {
                            color: '#e6f1ff',
                            font: { size: 12 }
                        }
                    },
                    tooltip: {
                        backgroundColor: getComputedStyle(document.documentElement).getPropertyValue('--primary-bg') || 'rgba(17, 34, 64, 0.95)',
                        titleColor: '#64ffda',
                        bodyColor: '#e6f1ff',
                        borderColor: '#64ffda',
                        borderWidth: 1
                    }
                },
                scales: {
                    x: {
                        grid: {
                            color: 'rgba(30, 58, 95, 0.3)'
                        },
                        ticks: {
                            color: '#8892b0'
                        }
                    },
                    y: {
                        grid: {
                            color: 'rgba(30, 58, 95, 0.3)'
                        },
                        ticks: {
                            color: '#8892b0',
                            callback: (value) => value + ' MB'
                        }
                    }
                }
            }
        });

        return this.charts[canvasId];
    }

    // Client Growth Chart (Area Chart)
    createClientGrowthChart(canvasId, data) {
        const ctx = document.getElementById(canvasId);
        if (!ctx) return;

        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
        }

        this.charts[canvasId] = new Chart(ctx, {
            type: 'line',
            data: {
                labels: data.labels || [],
                datasets: [{
                    label: 'Total Clients',
                    data: data.values || [],
                    borderColor: this.defaultColors.cyan,
                    backgroundColor: getComputedStyle(document.documentElement).getPropertyValue('--accent-color') + '40' || 'rgba(100, 255, 218, 0.2)',
                    tension: 0.4,
                    fill: true,
                    pointBackgroundColor: this.defaultColors.cyan,
                    pointBorderColor: '#0a192f',
                    pointBorderWidth: 2,
                    pointRadius: 4,
                    pointHoverRadius: 6
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        backgroundColor: getComputedStyle(document.documentElement).getPropertyValue('--primary-bg') || 'rgba(17, 34, 64, 0.95)',
                        titleColor: '#64ffda',
                        bodyColor: '#e6f1ff',
                        borderColor: '#64ffda',
                        borderWidth: 1
                    }
                },
                scales: {
                    x: {
                        grid: {
                            color: 'rgba(30, 58, 95, 0.3)'
                        },
                        ticks: {
                            color: '#8892b0'
                        }
                    },
                    y: {
                        grid: {
                            color: 'rgba(30, 58, 95, 0.3)'
                        },
                        ticks: {
                            color: '#8892b0',
                            stepSize: 1
                        }
                    }
                }
            }
        });

        return this.charts[canvasId];
    }

    // Connection Status Pie Chart
    createStatusChart(canvasId, data) {
        const ctx = document.getElementById(canvasId);
        if (!ctx) return;

        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
        }

        this.charts[canvasId] = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: data.labels || ['Active', 'Inactive', 'Expired'],
                datasets: [{
                    data: data.values || [0, 0, 0],
                    backgroundColor: [
                        this.defaultColors.green,
                        this.defaultColors.yellow,
                        this.defaultColors.red
                    ],
                    borderColor: '#0a192f',
                    borderWidth: 2
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: {
                            color: '#e6f1ff',
                            padding: 15,
                            font: { size: 12 }
                        }
                    },
                    tooltip: {
                        backgroundColor: getComputedStyle(document.documentElement).getPropertyValue('--primary-bg') || 'rgba(17, 34, 64, 0.95)',
                        titleColor: '#64ffda',
                        bodyColor: '#e6f1ff',
                        borderColor: '#64ffda',
                        borderWidth: 1
                    }
                }
            }
        });

        return this.charts[canvasId];
    }

    // Peak Hours Bar Chart
    createPeakHoursChart(canvasId, data) {
        const ctx = document.getElementById(canvasId);
        if (!ctx) return;

        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
        }

        this.charts[canvasId] = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: data.labels || ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00'],
                datasets: [{
                    label: 'Connections',
                    data: data.values || [],
                    backgroundColor: this.defaultColors.cyan,
                    borderColor: this.defaultColors.cyan,
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        backgroundColor: getComputedStyle(document.documentElement).getPropertyValue('--primary-bg') || 'rgba(17, 34, 64, 0.95)',
                        titleColor: '#64ffda',
                        bodyColor: '#e6f1ff',
                        borderColor: '#64ffda',
                        borderWidth: 1
                    }
                },
                scales: {
                    x: {
                        grid: {
                            color: 'rgba(30, 58, 95, 0.3)'
                        },
                        ticks: {
                            color: '#8892b0'
                        }
                    },
                    y: {
                        grid: {
                            color: 'rgba(30, 58, 95, 0.3)'
                        },
                        ticks: {
                            color: '#8892b0',
                            stepSize: 1
                        }
                    }
                }
            }
        });

        return this.charts[canvasId];
    }

    // Update chart data dynamically
    updateChart(canvasId, newData) {
        const chart = this.charts[canvasId];
        if (!chart) return;

        chart.data.labels = newData.labels;
        chart.data.datasets.forEach((dataset, index) => {
            if (newData.datasets && newData.datasets[index]) {
                dataset.data = newData.datasets[index];
            }
        });

        chart.update('none'); // Update without animation for real-time
    }

    // Destroy all charts
    destroyAll() {
        Object.values(this.charts).forEach(chart => chart.destroy());
        this.charts = {};
    }

    // Generate sample data for demo
    generateSampleTrafficData(days = 7) {
        const labels = [];
        const upload = [];
        const download = [];
        
        for (let i = days - 1; i >= 0; i--) {
            const date = new Date();
            date.setDate(date.getDate() - i);
            labels.push(date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }));
            upload.push(Math.floor(Math.random() * 500) + 100);
            download.push(Math.floor(Math.random() * 1000) + 200);
        }

        return { labels, upload, download };
    }

    generateSampleGrowthData(days = 30) {
        const labels = [];
        const values = [];
        let total = Math.floor(Math.random() * 10) + 1;
        
        for (let i = days - 1; i >= 0; i--) {
            const date = new Date();
            date.setDate(date.getDate() - i);
            labels.push(date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }));
            total += Math.random() > 0.7 ? Math.floor(Math.random() * 3) : 0;
            values.push(total);
        }

        return { labels, values };
    }

    generateSamplePeakHours() {
        const labels = [];
        const values = [];
        
        for (let i = 0; i < 24; i++) {
            labels.push(`${i.toString().padStart(2, '0')}:00`);
            // Simulate peak hours (8-12, 18-23)
            if ((i >= 8 && i <= 12) || (i >= 18 && i <= 23)) {
                values.push(Math.floor(Math.random() * 15) + 10);
            } else {
                values.push(Math.floor(Math.random() * 5));
            }
        }

        return { labels, values };
    }
}

// Global chart manager instance
const chartManager = new ChartManager();

