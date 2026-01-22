// Content Detail Page JavaScript

document.addEventListener('DOMContentLoaded', function() {
    if (window.location.pathname.startsWith('/detail/')) {
        loadContentDetail();
    }
});

async function loadContentDetail() {
    const loading = document.getElementById('loading');
    const detailContainer = document.getElementById('detail-container');
    const errorContainer = document.getElementById('error-container');
    
    // Get content ID from URL
    const pathParts = window.location.pathname.split('/');
    const contentId = pathParts[pathParts.length - 1];
    
    if (!contentId) {
        showError('Invalid content ID');
        return;
    }
    
    try {
        const response = await fetch(`/api/contents/${contentId}`, {
            credentials: 'include'
        });
        
        if (!response.ok) {
            if (response.status === 404) {
                throw new Error('Content not found');
            }
            throw new Error('Failed to load content');
        }
        
        const content = await response.json();
        renderContentDetail(content);
        
        // Hide loading, show content
        if (loading) loading.classList.add('hidden');
        if (detailContainer) detailContainer.classList.remove('hidden');
    } catch (error) {
        console.error('Error loading content detail:', error);
        showError(error.message);
    }
}

function renderContentDetail(content) {
    // Title
    const titleElement = document.getElementById('content-title');
    if (titleElement) {
        titleElement.textContent = content.title || 'Untitled';
    }
    
    // Criticality badge
    const criticalityElement = document.getElementById('content-criticality');
    if (criticalityElement) {
        const score = content.criticality_score || 0;
        criticalityElement.textContent = `Criticality: ${score}`;
        
        // Set badge color
        if (score >= 9) {
            criticalityElement.className = 'badge badge-critical';
        } else if (score >= 7) {
            criticalityElement.className = 'badge badge-high';
        } else if (score >= 4) {
            criticalityElement.className = 'badge badge-medium';
        } else {
            criticalityElement.className = 'badge badge-low';
        }
    }
    
    // Source
    const sourceElement = document.getElementById('content-source');
    if (sourceElement) {
        sourceElement.textContent = content.source_name || '-';
    }
    
    // Source URL
    const urlElement = document.getElementById('content-url');
    if (urlElement && content.source_url) {
        urlElement.href = content.source_url;
        urlElement.textContent = content.source_url;
    }
    
    // Published date
    const publishedElement = document.getElementById('content-published');
    if (publishedElement && content.published_at) {
        const date = new Date(content.published_at);
        publishedElement.textContent = date.toLocaleString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }
    
    // Collected date
    const collectedElement = document.getElementById('content-collected');
    if (collectedElement && content.collected_at) {
        const date = new Date(content.collected_at);
        collectedElement.textContent = date.toLocaleString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }
    
    // Categories
    const categoriesElement = document.getElementById('content-categories');
    if (categoriesElement) {
        if (content.categories && content.categories.length > 0) {
            categoriesElement.innerHTML = '';
            content.categories.forEach(category => {
                const badge = document.createElement('span');
                badge.className = 'category-badge';
                badge.textContent = category.name;
                badge.style.backgroundColor = category.color || '#3498db';
                categoriesElement.appendChild(badge);
            });
        } else {
            categoriesElement.textContent = 'No categories';
        }
    }
    
    // Content text
    const contentTextElement = document.getElementById('content-text');
    if (contentTextElement) {
        contentTextElement.textContent = content.content || 'No content available';
    }
}

function showError(message) {
    const loading = document.getElementById('loading');
    const detailContainer = document.getElementById('detail-container');
    const errorContainer = document.getElementById('error-container');
    const errorMessage = document.getElementById('error-message');
    
    if (loading) loading.classList.add('hidden');
    if (detailContainer) detailContainer.classList.add('hidden');
    if (errorContainer) errorContainer.classList.remove('hidden');
    if (errorMessage) errorMessage.textContent = message;
}
