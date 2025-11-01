// Bulk Operations Manager for Client Management

class BulkOperations {
    constructor() {
        this.selectedClients = new Set();
    }

    toggleClient(clientName) {
        if (this.selectedClients.has(clientName)) {
            this.selectedClients.delete(clientName);
        } else {
            this.selectedClients.add(clientName);
        }
        this.updateUI();
    }

    selectAll(clients) {
        clients.forEach(c => this.selectedClients.add(c.name));
        this.updateUI();
    }

    deselectAll() {
        this.selectedClients.clear();
        this.updateUI();
    }

    updateUI() {
        const count = this.selectedClients.size;
        const bulkBar = document.getElementById('bulkActionBar');
        
        if (bulkBar) {
            if (count > 0) {
                bulkBar.classList.remove('hidden');
                bulkBar.querySelector('#selectedCount').textContent = count;
            } else {
                bulkBar.classList.add('hidden');
            }
        }

        // Update checkboxes
        document.querySelectorAll('.client-checkbox').forEach(cb => {
            cb.checked = this.selectedClients.has(cb.dataset.client);
        });
    }

    async bulkDelete() {
        if (this.selectedClients.size === 0) return;

        const count = this.selectedClients.size;
        if (!confirm(`Remove ${count} client(s)?\n\nThis will revoke certificates and delete all data.`)) {
            return;
        }

        const notifId = notificationManager.withProgress(`Removing ${count} clients...`, 'info');
        let completed = 0;

        for (const clientName of this.selectedClients) {
            try {
                await api.deleteClient(clientName);
                completed++;
                const progress = (completed / count) * 100;
                notificationManager.updateProgress(notifId, progress);
            } catch (error) {
                console.error(`Failed to delete ${clientName}:`, error);
            }
        }

        this.deselectAll();
        notificationManager.success(`Successfully removed ${completed} client(s)`);
        
        // Reload clients
        if (typeof loadClients === 'function') {
            loadClients();
        }
    }

    async bulkExport() {
        if (this.selectedClients.size === 0) return;

        const notifId = notificationManager.withProgress('Preparing export...', 'info');
        const clients = Array.from(this.selectedClients);
        
        // Create a JSON file with all selected clients
        const data = {
            export_date: new Date().toISOString(),
            client_count: clients.length,
            clients: clients
        };

        const json = JSON.stringify(data, null, 2);
        const blob = new Blob([json], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `irangate_clients_${Date.now()}.json`;
        link.click();
        URL.revokeObjectURL(url);

        notificationManager.updateProgress(notifId, 100);
        notificationManager.success(`Exported ${clients.length} clients`);
    }

    getSelected() {
        return Array.from(this.selectedClients);
    }

    getCount() {
        return this.selectedClients.size;
    }
}

// Global instance
const bulkOps = new BulkOperations();

