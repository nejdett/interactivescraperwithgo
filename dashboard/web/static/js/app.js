// Main Application JavaScript

let currentPage = 1;
let totalPages = 1;
let sortBy = 'published_at';
let sortOrder = 'desc';
let filterCategory = '';

document.addEventListener('DOMContentLoaded', function() {
    if (window.location.pathname === '/dashboard') {
        initializeDashboard();
    }
});

function initializeDashboard() {
    // Load categories for filter
    loadCategories();
    
    // Load content list
    loadContentList();
    
    // Set up event listeners
    setupEventListeners();
}

function setupEventListeners() {
    // Sort controls
    const sortBySelect = document.getElementById('sort-by');
    const sortOrderSelect = document.getElementById('sort-order');
    const filterCategorySelect = document.getElementById('filter-category');
    
    if (sortBySelect) {
        sortBySelect.addEventListener('change', function() {
            sortBy = this.value;
            currentPage = 1;
            loadContentList();
        });
    }
    
    if (sortOrderSelect) {
        sortOrderSelect.addEventListener('change', function() {
            sortOrder = this.value;
            currentPage = 1;
            loadContentList();
        });
    }
    
    if (filterCategorySelect) {
        filterCategorySelect.addEventListener('change', function() {
            filterCategory = this.value;
            currentPage = 1;
            loadContentList();
        });
    }
    
    // Pagination controls
    const prevButton = document.getElementById('prev-page');
    const nextButton = document.getElementById('next-page');
    
    if (prevButton) {
        prevButton.addEventListener('click', function() {
            if (currentPage > 1) {
                currentPage--;
                loadContentList();
            }
        });
    }
    
    if (nextButton) {
        nextButton.addEventListener('click', function() {
            if (currentPage < totalPages) {
                currentPage++;
                loadContentList();
            }
        });
    }
    
    // Table header sorting
    const tableHeaders = document.querySelectorAll('.table th[data-sort]');
    tableHeaders.forEach(header => {
        header.addEventListener('click', function() {
            const newSortBy = this.getAttribute('data-sort');
            
            // Toggle order if clicking same column
            if (sortBy === newSortBy) {
                sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                sortBy = newSortBy;
                sortOrder = 'desc';
            }
            
            // Update select controls
            document.getElementById('sort-by').value = sortBy;
            document.getElementById('sort-order').value = sortOrder;
            
            currentPage = 1;
            loadContentList();
        });
    });
}

async function loadCategories() {
    try {
        const response = await fetch('/api/categories', {
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to load categories');
        }
        
        const categories = await response.json();
        const filterSelect = document.getElementById('filter-category');
        
        if (filterSelect) {
            // Clear existing options except "All Categories"
            filterSelect.innerHTML = '<option value="">Tüm Kategoriler</option>';
            
            // Add category options
            categories.forEach(category => {
                const option = document.createElement('option');
                option.value = category.name;
                option.textContent = category.name;
                filterSelect.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Error loading categories:', error);
    }
}

async function loadContentList() {
    const loading = document.getElementById('loading');
    const tableContainer = document.getElementById('content-table-container');
    
    // Show loading spinner
    if (loading) loading.classList.remove('hidden');
    if (tableContainer) tableContainer.classList.add('hidden');
    
    try {
        // Build query parameters
        const params = new URLSearchParams({
            page: currentPage,
            page_size: 50,
            sort_by: sortBy,
            order: sortOrder
        });
        
        if (filterCategory) {
            params.append('category', filterCategory);
        }
        
        const response = await fetch(`/api/contents?${params.toString()}`, {
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to load content list');
        }
        
        const data = await response.json();
        renderContentTable(data);
        updatePagination(data);
        
        // Hide loading, show table
        if (loading) loading.classList.add('hidden');
        if (tableContainer) tableContainer.classList.remove('hidden');
    } catch (error) {
        console.error('Error loading content list:', error);
        if (loading) loading.classList.add('hidden');
    }
}

function renderContentTable(data) {
    const tbody = document.getElementById('content-table-body');
    if (!tbody) return;
    
    tbody.innerHTML = '';
    
    if (!data.items || data.items.length === 0) {
        tbody.innerHTML = '<tr><td colspan="4" class="text-center text-muted">No content items found</td></tr>';
        return;
    }
    
    data.items.forEach(item => {
        const row = document.createElement('tr');
        row.style.cursor = 'pointer';
        row.addEventListener('click', function() {
            window.location.href = `/detail/${item.id}`;
        });
        
        // Title
        const titleCell = document.createElement('td');
        titleCell.textContent = item.title || 'Untitled';
        row.appendChild(titleCell);
        
        // Source
        const sourceCell = document.createElement('td');
        sourceCell.textContent = item.source_name || '-';
        row.appendChild(sourceCell);
        
        // Date
        const dateCell = document.createElement('td');
        if (item.published_at) {
            const date = new Date(item.published_at);
            dateCell.textContent = date.toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'short',
                day: 'numeric'
            });
        } else {
            dateCell.textContent = '-';
        }
        row.appendChild(dateCell);
        
        // Criticality
        const criticalityCell = document.createElement('td');
        const badge = document.createElement('span');
        badge.className = 'badge';
        badge.textContent = item.criticality_score || 0;
        
        // Set badge color based on criticality
        const score = item.criticality_score || 0;
        if (score >= 9) {
            badge.classList.add('badge-critical');
        } else if (score >= 7) {
            badge.classList.add('badge-high');
        } else if (score >= 4) {
            badge.classList.add('badge-medium');
        } else {
            badge.classList.add('badge-low');
        }
        
        criticalityCell.appendChild(badge);
        row.appendChild(criticalityCell);
        
        tbody.appendChild(row);
    });
}

function updatePagination(data) {
    totalPages = data.total_pages || 1;
    
    // Update page info
    document.getElementById('current-page').textContent = currentPage;
    document.getElementById('total-pages').textContent = totalPages;
    
    // Update button states
    const prevButton = document.getElementById('prev-page');
    const nextButton = document.getElementById('next-page');
    
    if (prevButton) {
        prevButton.disabled = currentPage <= 1;
    }
    
    if (nextButton) {
        nextButton.disabled = currentPage >= totalPages;
    }
}
