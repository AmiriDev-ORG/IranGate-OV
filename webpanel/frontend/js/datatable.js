// Advanced DataTable Component for IranGate
// Features: Sorting, filtering, export, column visibility

class DataTable {
    constructor(containerId, options = {}) {
        this.container = document.getElementById(containerId);
        this.data = [];
        this.filteredData = [];
        this.currentSort = { column: null, direction: 'asc' };
        this.currentPage = 1;
        this.pageSize = options.pageSize || 25;
        this.columns = options.columns || [];
        this.searchTerm = '';
        this.filters = {};
        this.selectedRows = new Set();
    }

    setData(data) {
        this.data = data;
        this.filteredData = [...data];
        this.render();
    }

    // Sort data
    sort(column) {
        if (this.currentSort.column === column) {
            this.currentSort.direction = this.currentSort.direction === 'asc' ? 'desc' : 'asc';
        } else {
            this.currentSort.column = column;
            this.currentSort.direction = 'asc';
        }

        this.filteredData.sort((a, b) => {
            const aVal = a[column];
            const bVal = b[column];
            
            if (aVal === bVal) return 0;
            
            const comparison = aVal < bVal ? -1 : 1;
            return this.currentSort.direction === 'asc' ? comparison : -comparison;
        });

        this.render();
    }

    // Filter data
    applyFilters() {
        this.filteredData = this.data.filter(row => {
            // Search term filter
            if (this.searchTerm) {
                const searchLower = this.searchTerm.toLowerCase();
                const matchesSearch = Object.values(row).some(val => 
                    String(val).toLowerCase().includes(searchLower)
                );
                if (!matchesSearch) return false;
            }

            // Column filters
            for (const [column, value] of Object.entries(this.filters)) {
                if (value && row[column] !== value) {
                    return false;
                }
            }

            return true;
        });

        this.currentPage = 1;
        this.render();
    }

    setFilter(column, value) {
        this.filters[column] = value;
        this.applyFilters();
    }

    setSearch(term) {
        this.searchTerm = term;
        this.applyFilters();
    }

    // Pagination
    get paginatedData() {
        const start = (this.currentPage - 1) * this.pageSize;
        const end = start + this.pageSize;
        return this.filteredData.slice(start, end);
    }

    get totalPages() {
        return Math.ceil(this.filteredData.length / this.pageSize);
    }

    goToPage(page) {
        if (page >= 1 && page <= this.totalPages) {
            this.currentPage = page;
            this.render();
        }
    }

    nextPage() {
        this.goToPage(this.currentPage + 1);
    }

    prevPage() {
        this.goToPage(this.currentPage - 1);
    }

    // Row selection
    toggleRow(rowId) {
        if (this.selectedRows.has(rowId)) {
            this.selectedRows.delete(rowId);
        } else {
            this.selectedRows.add(rowId);
        }
        this.updateCheckboxes();
    }

    selectAll() {
        this.paginatedData.forEach(row => {
            this.selectedRows.add(row.name || row.id);
        });
        this.updateCheckboxes();
    }

    deselectAll() {
        this.selectedRows.clear();
        this.updateCheckboxes();
    }

    updateCheckboxes() {
        document.querySelectorAll('.row-checkbox').forEach(checkbox => {
            const rowId = checkbox.dataset.id;
            checkbox.checked = this.selectedRows.has(rowId);
        });

        const selectAllCheckbox = document.getElementById('selectAllCheckbox');
        if (selectAllCheckbox) {
            const allSelected = this.paginatedData.every(row => 
                this.selectedRows.has(row.name || row.id)
            );
            selectAllCheckbox.checked = allSelected;
        }

        // Update bulk action button
        this.updateBulkActions();
    }

    updateBulkActions() {
        const bulkActionsBtn = document.getElementById('bulkActionsBtn');
        const count = this.selectedRows.size;
        
        if (bulkActionsBtn) {
            if (count > 0) {
                bulkActionsBtn.classList.remove('hidden');
                bulkActionsBtn.querySelector('.count').textContent = count;
            } else {
                bulkActionsBtn.classList.add('hidden');
            }
        }
    }

    // Export functions
    exportCSV() {
        const headers = this.columns.map(col => col.label).join(',');
        const rows = this.filteredData.map(row => 
            this.columns.map(col => {
                const value = row[col.field];
                // Escape commas and quotes
                return typeof value === 'string' && value.includes(',') ? 
                    `"${value.replace(/"/g, '""')}"` : value;
            }).join(',')
        );

        const csv = [headers, ...rows].join('\n');
        this.downloadFile(csv, 'clients.csv', 'text/csv');
    }

    exportJSON() {
        const json = JSON.stringify(this.filteredData, null, 2);
        this.downloadFile(json, 'clients.json', 'application/json');
    }

    downloadFile(content, filename, mimeType) {
        const blob = new Blob([content], { type: mimeType });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        link.click();
        URL.revokeObjectURL(url);
    }

    // Render table
    render() {
        if (!this.container) return;

        const data = this.paginatedData;
        
        // Build table HTML
        let html = `
            <div class="overflow-x-auto">
                <table class="w-full">
                    <thead class="bg-gray-50">
                        <tr>
                            <th class="px-6 py-3 text-left">
                                <input type="checkbox" id="selectAllCheckbox" 
                                       onchange="dataTable.selectAll()" 
                                       class="rounded border-gray-600 bg-navy-800">
                            </th>
        `;

        // Column headers with sort
        this.columns.forEach(col => {
            const isSorted = this.currentSort.column === col.field;
            const icon = isSorted ? 
                (this.currentSort.direction === 'asc' ? '↑' : '↓') : '↕';
            
            html += `
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:text-cyan-400"
                    onclick="dataTable.sort('${col.field}')">
                    ${col.label} <span class="ml-1">${icon}</span>
                </th>
            `;
        });

        html += `
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                                Actions
                            </th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-200">
        `;

        // Data rows
        if (data.length === 0) {
            html += `
                <tr>
                    <td colspan="${this.columns.length + 2}" class="px-6 py-8 text-center text-gray-500">
                        No data found
                    </td>
                </tr>
            `;
        } else {
            data.forEach(row => {
                const rowId = row.name || row.id;
                html += `
                    <tr class="hover:bg-gray-50">
                        <td class="px-6 py-4">
                            <input type="checkbox" class="row-checkbox rounded border-gray-600 bg-navy-800" 
                                   data-id="${rowId}"
                                   onchange="dataTable.toggleRow('${rowId}')">
                        </td>
                `;

                this.columns.forEach(col => {
                    let value = row[col.field];
                    
                    // Custom formatters
                    if (col.formatter) {
                        value = col.formatter(value, row);
                    } else if (col.field.includes('date') || col.field.includes('Date')) {
                        value = formatDate(value);
                    }

                    html += `<td class="px-6 py-4 whitespace-nowrap text-sm">${value}</td>`;
                });

                // Actions column
                html += `
                        <td class="px-6 py-4 whitespace-nowrap text-sm font-medium space-x-2">
                            ${col.actions ? col.actions(row) : ''}
                        </td>
                    </tr>
                `;
            });
        }

        html += `
                    </tbody>
                </table>
            </div>
        `;

        // Pagination
        if (this.totalPages > 1) {
            html += this.renderPagination();
        }

        this.container.innerHTML = html;
        this.updateCheckboxes();
    }

    renderPagination() {
        const pages = [];
        const maxVisible = 5;
        let startPage = Math.max(1, this.currentPage - 2);
        let endPage = Math.min(this.totalPages, startPage + maxVisible - 1);

        if (endPage - startPage < maxVisible - 1) {
            startPage = Math.max(1, endPage - maxVisible + 1);
        }

        return `
            <div class="flex items-center justify-between px-6 py-4 border-t border-gray-600">
                <div class="text-sm text-gray-500">
                    Showing ${(this.currentPage - 1) * this.pageSize + 1} to 
                    ${Math.min(this.currentPage * this.pageSize, this.filteredData.length)} of 
                    ${this.filteredData.length} results
                </div>
                <div class="flex items-center gap-2">
                    <button onclick="dataTable.prevPage()" 
                            ${this.currentPage === 1 ? 'disabled' : ''}
                            class="px-3 py-1 rounded bg-navy-800 hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed">
                        Previous
                    </button>
                    
                    ${Array.from({length: endPage - startPage + 1}, (_, i) => startPage + i).map(page => `
                        <button onclick="dataTable.goToPage(${page})"
                                class="px-3 py-1 rounded ${page === this.currentPage ? 
                                    'bg-cyan-600 text-white' : 
                                    'bg-navy-800 hover:bg-blue-600'}">
                            ${page}
                        </button>
                    `).join('')}
                    
                    <button onclick="dataTable.nextPage()" 
                            ${this.currentPage === this.totalPages ? 'disabled' : ''}
                            class="px-3 py-1 rounded bg-navy-800 hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed">
                        Next
                    </button>
                    
                    <select onchange="dataTable.changePageSize(this.value)" 
                            class="px-2 py-1 rounded bg-navy-800 border border-gray-600 text-sm">
                        <option value="10" ${this.pageSize === 10 ? 'selected' : ''}>10</option>
                        <option value="25" ${this.pageSize === 25 ? 'selected' : ''}>25</option>
                        <option value="50" ${this.pageSize === 50 ? 'selected' : ''}>50</option>
                        <option value="100" ${this.pageSize === 100 ? 'selected' : ''}>100</option>
                    </select>
                </div>
            </div>
        `;
    }

    changePageSize(size) {
        this.pageSize = parseInt(size);
        this.currentPage = 1;
        this.render();
    }

    // Get selected rows data
    getSelected() {
        return this.data.filter(row => 
            this.selectedRows.has(row.name || row.id)
        );
    }
}

// Global instance (will be initialized per page)
let dataTable = null;

