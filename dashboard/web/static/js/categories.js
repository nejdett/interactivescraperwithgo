// Categories Management JavaScript

document.addEventListener('DOMContentLoaded', function() {
    if (window.location.pathname === '/categories') {
        initializeCategoriesPage();
    }
});

function initializeCategoriesPage() {
    loadCategories();
    setupEventListeners();
}

function setupEventListeners() {
    // Create category form
    const createForm = document.getElementById('create-category-form');
    if (createForm) {
        createForm.addEventListener('submit', handleCreateCategory);
    }
    
    // Edit category form
    const editForm = document.getElementById('edit-category-form');
    if (editForm) {
        editForm.addEventListener('submit', handleEditCategory);
    }
    
    // Modal close buttons
    const closeModal = document.getElementById('close-modal');
    const cancelEdit = document.getElementById('cancel-edit');
    
    if (closeModal) {
        closeModal.addEventListener('click', hideEditModal);
    }
    
    if (cancelEdit) {
        cancelEdit.addEventListener('click', hideEditModal);
    }
    
    // Close modal on outside click
    const modal = document.getElementById('edit-modal');
    if (modal) {
        modal.addEventListener('click', function(e) {
            if (e.target === modal) {
                hideEditModal();
            }
        });
    }
}

async function loadCategories() {
    const loading = document.getElementById('loading');
    const tableContainer = document.getElementById('categories-table-container');
    
    if (loading) loading.classList.remove('hidden');
    if (tableContainer) tableContainer.classList.add('hidden');
    
    try {
        const response = await fetch('/api/categories', {
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to load categories');
        }
        
        const categories = await response.json();
        renderCategoriesTable(categories);
        
        if (loading) loading.classList.add('hidden');
        if (tableContainer) tableContainer.classList.remove('hidden');
    } catch (error) {
        console.error('Error loading categories:', error);
        showError('Kategoriler yüklenemedi');
        if (loading) loading.classList.add('hidden');
    }
}

function renderCategoriesTable(categories) {
    const tbody = document.getElementById('categories-table-body');
    if (!tbody) return;
    
    tbody.innerHTML = '';
    
    if (!categories || categories.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="text-center text-muted">No categories found</td></tr>';
        return;
    }
    
    categories.forEach(category => {
        const row = document.createElement('tr');
        
        // Name
        const nameCell = document.createElement('td');
        nameCell.textContent = category.name;
        row.appendChild(nameCell);
        
        // Description
        const descCell = document.createElement('td');
        descCell.textContent = category.description || '-';
        row.appendChild(descCell);
        
        // Default Criticality
        const critCell = document.createElement('td');
        critCell.textContent = category.default_criticality || '-';
        row.appendChild(critCell);
        
        // Color
        const colorCell = document.createElement('td');
        const colorPreview = document.createElement('span');
        colorPreview.className = 'color-preview';
        colorPreview.style.backgroundColor = category.color || '#3498db';
        colorCell.appendChild(colorPreview);
        row.appendChild(colorCell);
        
        // Actions
        const actionsCell = document.createElement('td');
        const actionsDiv = document.createElement('div');
        actionsDiv.className = 'action-buttons';
        
        const editBtn = document.createElement('button');
        editBtn.className = 'btn btn-primary btn-sm';
        editBtn.textContent = 'Edit';
        editBtn.addEventListener('click', () => showEditModal(category));
        
        const deleteBtn = document.createElement('button');
        deleteBtn.className = 'btn btn-danger btn-sm';
        deleteBtn.textContent = 'Delete';
        deleteBtn.addEventListener('click', () => handleDeleteCategory(category.id, category.name));
        
        actionsDiv.appendChild(editBtn);
        actionsDiv.appendChild(deleteBtn);
        actionsCell.appendChild(actionsDiv);
        row.appendChild(actionsCell);
        
        tbody.appendChild(row);
    });
}

async function handleCreateCategory(event) {
    event.preventDefault();
    
    const form = event.target;
    const formData = new FormData(form);
    
    const categoryData = {
        name: formData.get('name'),
        description: formData.get('description'),
        default_criticality: parseInt(formData.get('default_criticality')),
        color: formData.get('color')
    };
    
    try {
        const response = await fetch('/api/categories', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify(categoryData)
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to create category');
        }
        
        showSuccess('Kategori başarıyla oluşturuldu');
        form.reset();
        loadCategories();
    } catch (error) {
        console.error('Error creating category:', error);
        showError(error.message);
    }
}

function showEditModal(category) {
    const modal = document.getElementById('edit-modal');
    
    // Populate form fields
    document.getElementById('edit-category-id').value = category.id;
    document.getElementById('edit-category-name').value = category.name;
    document.getElementById('edit-category-description').value = category.description || '';
    document.getElementById('edit-category-criticality').value = category.default_criticality;
    document.getElementById('edit-category-color').value = category.color || '#3498db';
    
    // Show modal
    if (modal) {
        modal.classList.remove('hidden');
    }
}

function hideEditModal() {
    const modal = document.getElementById('edit-modal');
    if (modal) {
        modal.classList.add('hidden');
    }
}

async function handleEditCategory(event) {
    event.preventDefault();
    
    const categoryId = document.getElementById('edit-category-id').value;
    const form = event.target;
    const formData = new FormData(form);
    
    const categoryData = {
        name: formData.get('name'),
        description: formData.get('description'),
        default_criticality: parseInt(formData.get('default_criticality')),
        color: formData.get('color')
    };
    
    try {
        const response = await fetch(`/api/categories/${categoryId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify(categoryData)
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to update category');
        }
        
        showSuccess('Kategori başarıyla güncellendi');
        hideEditModal();
        loadCategories();
    } catch (error) {
        console.error('Error updating category:', error);
        showError(error.message);
    }
}

async function handleDeleteCategory(categoryId, categoryName) {
    // Show confirmation dialog
    const confirmed = confirm(`"${categoryName}" kategorisini silmek istediğinizden emin misiniz? Bu işlem geri alınamaz.`);
    
    if (!confirmed) {
        return;
    }
    
    try {
        const response = await fetch(`/api/categories/${categoryId}`, {
            method: 'DELETE',
            credentials: 'include'
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to delete category');
        }
        
        showSuccess('Kategori başarıyla silindi');
        loadCategories();
    } catch (error) {
        console.error('Error deleting category:', error);
        showError(error.message);
    }
}

function showSuccess(message) {
    const successElement = document.getElementById('success-message');
    if (successElement) {
        successElement.textContent = message;
        successElement.classList.remove('hidden');
        
        // Hide after 5 seconds
        setTimeout(() => {
            successElement.classList.add('hidden');
        }, 5000);
    }
}

function showError(message) {
    const errorElement = document.getElementById('error-message');
    if (errorElement) {
        errorElement.textContent = message;
        errorElement.classList.remove('hidden');
        
        // Hide after 5 seconds
        setTimeout(() => {
            errorElement.classList.add('hidden');
        }, 5000);
    }
}
