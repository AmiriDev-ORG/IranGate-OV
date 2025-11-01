// Advanced Search Modal with Multi-Criteria Filtering

class AdvancedSearch {
    constructor() {
        this.filters = {
            name: '',
            status: [],
            createdAfter: '',
            createdBefore: '',
            trafficMin: 0,
            trafficMax: 0,
            group: '',
            active: null
        };
        this.savedSearches = this.loadSavedSearches();
    }

    showModal() {
        const modal = this.createModal();
        document.body.appendChild(modal);
        
        // Load saved values
        this.loadFilterValues();
    }

    createModal() {
        const modal = document.createElement('div');
        modal.id = 'advancedSearchModal';
        modal.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
        modal.innerHTML = `
            <div class="bg-white rounded-lg p-8 max-w-2xl w-full mx-4 max-h-screen overflow-y-auto">
                <div class="flex justify-between items-center mb-6">
                    <h2 class="text-2xl font-bold text-gray-800">Advanced Search</h2>
                    <button onclick="advancedSearch.closeModal()" class="text-gray-400 hover:text-gray-600">
                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                    </button>
                </div>

                <form id="advancedSearchForm" class="space-y-4">
                    <!-- Name Search -->
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">Client Name</label>
                        <input type="text" id="searchName" placeholder="Search by name..." 
                               class="w-full px-4 py-2 border rounded-lg">
                    </div>

                    <!-- Status Filter -->
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">Status</label>
                        <div class="flex gap-2">
                            <label class="flex items-center">
                                <input type="checkbox" value="active" class="mr-2 status-checkbox">
                                <span class="text-sm">Active</span>
                            </label>
                            <label class="flex items-center">
                                <input type="checkbox" value="inactive" class="mr-2 status-checkbox">
                                <span class="text-sm">Inactive</span>
                            </label>
                            <label class="flex items-center">
                                <input type="checkbox" value="expired" class="mr-2 status-checkbox">
                                <span class="text-sm">Expired</span>
                            </label>
                        </div>
                    </div>

                    <!-- Date Range -->
                    <div class="grid grid-cols-2 gap-4">
                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">Created After</label>
                            <input type="date" id="createdAfter" class="w-full px-4 py-2 border rounded-lg">
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">Created Before</label>
                            <input type="date" id="createdBefore" class="w-full px-4 py-2 border rounded-lg">
                        </div>
                    </div>

                    <!-- Traffic Range -->
                    <div class="grid grid-cols-2 gap-4">
                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">Min Traffic (MB)</label>
                            <input type="number" id="trafficMin" placeholder="0" class="w-full px-4 py-2 border rounded-lg">
                        </div>
                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">Max Traffic (MB)</label>
                            <input type="number" id="trafficMax" placeholder="Unlimited" class="w-full px-4 py-2 border rounded-lg">
                        </div>
                    </div>

                    <!-- Group Filter -->
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">Group</label>
                        <select id="groupFilter" class="w-full px-4 py-2 border rounded-lg">
                            <option value="">All Groups</option>
                            <option value="VIP">VIP</option>
                            <option value="Basic">Basic</option>
                            <option value="Trial">Trial</option>
                        </select>
                    </div>

                    <!-- Saved Searches -->
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">Saved Searches</label>
                        <div class="flex gap-2">
                            <select id="savedSearchSelect" class="flex-1 px-4 py-2 border rounded-lg">
                                <option value="">-- Load Saved Search --</option>
                            </select>
                            <button type="button" onclick="advancedSearch.saveCurrentSearch()" 
                                    class="bg-blue-500 text-white px-4 py-2 rounded-lg text-sm">
                                Save
                            </button>
                        </div>
                    </div>

                    <!-- Action Buttons -->
                    <div class="flex gap-4 pt-4">
                        <button type="button" onclick="advancedSearch.resetFilters()" 
                                class="flex-1 bg-gray-200 hover:bg-gray-300 py-2 rounded-lg">
                            Reset
                        </button>
                        <button type="submit" 
                                class="flex-1 bg-purple-600 hover:bg-purple-700 text-white py-2 rounded-lg">
                            Apply Search
                        </button>
                    </div>
                </form>
            </div>
        `;

        // Add event listener
        modal.querySelector('#advancedSearchForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.applyFilters();
        });

        this.loadSavedSearchesDropdown(modal);

        return modal;
    }

    loadFilterValues() {
        document.getElementById('searchName').value = this.filters.name;
        document.getElementById('createdAfter').value = this.filters.createdAfter;
        document.getElementById('createdBefore').value = this.filters.createdBefore;
        document.getElementById('trafficMin').value = this.filters.trafficMin || '';
        document.getElementById('trafficMax').value = this.filters.trafficMax || '';
        document.getElementById('groupFilter').value = this.filters.group;

        // Status checkboxes
        document.querySelectorAll('.status-checkbox').forEach(cb => {
            cb.checked = this.filters.status.includes(cb.value);
        });
    }

    readFilterValues() {
        this.filters.name = document.getElementById('searchName').value;
        this.filters.createdAfter = document.getElementById('createdAfter').value;
        this.filters.createdBefore = document.getElementById('createdBefore').value;
        this.filters.trafficMin = parseInt(document.getElementById('trafficMin').value) || 0;
        this.filters.trafficMax = parseInt(document.getElementById('trafficMax').value) || 0;
        this.filters.group = document.getElementById('groupFilter').value;

        // Status checkboxes
        this.filters.status = [];
        document.querySelectorAll('.status-checkbox:checked').forEach(cb => {
            this.filters.status.push(cb.value);
        });
    }

    applyFilters() {
        this.readFilterValues();
        this.closeModal();

        // Apply filters to data (callback to parent)
        if (this.onApply) {
            this.onApply(this.filters);
        }

        // Show active filter count
        this.showActiveFilters();
    }

    filterData(data) {
        return data.filter(client => {
            // Name filter
            if (this.filters.name && !client.name.toLowerCase().includes(this.filters.name.toLowerCase())) {
                return false;
            }

            // Status filter
            if (this.filters.status.length > 0) {
                const clientStatus = client.active ? 'active' : 'inactive';
                const isExpired = new Date(client.expires_at) < new Date();
                if (isExpired && !this.filters.status.includes('expired')) return false;
                if (!isExpired && !this.filters.status.includes(clientStatus)) return false;
            }

            // Date range filter
            if (this.filters.createdAfter) {
                const createdDate = new Date(client.created_at);
                if (createdDate < new Date(this.filters.createdAfter)) return false;
            }

            if (this.filters.createdBefore) {
                const createdDate = new Date(client.created_at);
                if (createdDate > new Date(this.filters.createdBefore)) return false;
            }

            // Traffic filter
            const totalTraffic = (client.bytes_sent + client.bytes_received) / (1024 * 1024); // MB
            if (this.filters.trafficMin > 0 && totalTraffic < this.filters.trafficMin) {
                return false;
            }
            if (this.filters.trafficMax > 0 && totalTraffic > this.filters.trafficMax) {
                return false;
            }

            // Group filter
            if (this.filters.group && client.group !== this.filters.group) {
                return false;
            }

            return true;
        });
    }

    resetFilters() {
        this.filters = {
            name: '',
            status: [],
            createdAfter: '',
            createdBefore: '',
            trafficMin: 0,
            trafficMax: 0,
            group: '',
            active: null
        };
        this.loadFilterValues();
    }

    closeModal() {
        const modal = document.getElementById('advancedSearchModal');
        if (modal) {
            modal.remove();
        }
    }

    showActiveFilters() {
        const activeCount = this.getActiveFilterCount();
        const badge = document.getElementById('activeFilterBadge');
        if (badge) {
            if (activeCount > 0) {
                badge.textContent = activeCount;
                badge.classList.remove('hidden');
            } else {
                badge.classList.add('hidden');
            }
        }
    }

    getActiveFilterCount() {
        let count = 0;
        if (this.filters.name) count++;
        if (this.filters.status.length > 0) count++;
        if (this.filters.createdAfter) count++;
        if (this.filters.createdBefore) count++;
        if (this.filters.trafficMin > 0) count++;
        if (this.filters.trafficMax > 0) count++;
        if (this.filters.group) count++;
        return count;
    }

    saveCurrentSearch() {
        const name = prompt('Enter name for this search:');
        if (!name) return;

        const search = {
            name: name,
            filters: { ...this.filters },
            createdAt: new Date().toISOString()
        };

        this.savedSearches.push(search);
        localStorage.setItem('savedSearches', JSON.stringify(this.savedSearches));
        
        notificationManager.success(`Search "${name}" saved!`);
        this.loadSavedSearchesDropdown();
    }

    loadSavedSearches() {
        const saved = localStorage.getItem('savedSearches');
        return saved ? JSON.parse(saved) : [];
    }

    loadSavedSearchesDropdown(modal = document) {
        const select = modal.querySelector('#savedSearchSelect');
        if (!select) return;

        select.innerHTML = '<option value="">-- Load Saved Search --</option>';
        this.savedSearches.forEach((search, index) => {
            const option = document.createElement('option');
            option.value = index;
            option.textContent = search.name;
            select.appendChild(option);
        });

        select.addEventListener('change', (e) => {
            if (e.target.value !== '') {
                this.loadSavedSearch(parseInt(e.target.value));
            }
        });
    }

    loadSavedSearch(index) {
        if (index >= 0 && index < this.savedSearches.length) {
            this.filters = { ...this.savedSearches[index].filters };
            this.loadFilterValues();
            notificationManager.info(`Loaded search: ${this.savedSearches[index].name}`);
        }
    }
}

// Global instance
const advancedSearch = new AdvancedSearch();

