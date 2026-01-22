// Session Management JavaScript

// Check session on protected pages
document.addEventListener('DOMContentLoaded', function() {
    const protectedPaths = ['/dashboard', '/detail', '/categories'];
    const currentPath = window.location.pathname;
    
    // Check if current page is protected
    const isProtected = protectedPaths.some(path => currentPath.startsWith(path));
    
    if (isProtected) {
        validateSession();
        setupLogoutHandler();
        setupCategoryLinkVisibility();
    }
});

async function validateSession() {
    try {
        const response = await fetch('/api/auth/session', {
            credentials: 'include'
        });
        
        if (!response.ok) {
            // Session is invalid, redirect to login
            window.location.href = '/login';
            return;
        }
        
        const sessionData = await response.json();
        updateUserInfo(sessionData);
    } catch (error) {
        console.error('Session validation error:', error);
        // On error, redirect to login
        window.location.href = '/login';
    }
}

function updateUserInfo(sessionData) {
    // Update username display
    const usernameElement = document.getElementById('user-username');
    if (usernameElement) {
        const username = sessionData.user?.username || sessionData.username || 'Kullanıcı';
        usernameElement.textContent = username;
    }
    
    // Update role display
    const roleElement = document.getElementById('user-role');
    if (roleElement) {
        const role = sessionData.user?.role || sessionData.role || '';
        if (role) {
            roleElement.textContent = role.charAt(0).toUpperCase() + role.slice(1);
        }
    }
}

function setupLogoutHandler() {
    const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', handleLogout);
    }
}

async function handleLogout() {
    try {
        const response = await fetch('/api/auth/logout', {
            method: 'POST',
            credentials: 'include'
        });
        
        // Redirect to login regardless of response
        // (session will be cleared on server side)
        window.location.href = '/login';
    } catch (error) {
        console.error('Logout error:', error);
        // Still redirect to login on error
        window.location.href = '/login';
    }
}

async function setupCategoryLinkVisibility() {
    // Check if user has admin role to show/hide category management link
    try {
        const response = await fetch('/api/auth/session', {
            credentials: 'include'
        });
        
        if (response.ok) {
            const sessionData = await response.json();
            const categoryLink = document.getElementById('categories-link');
            
            if (categoryLink) {
                const role = sessionData.user?.role || sessionData.role || '';
                // Show category link only for admin users
                if (role !== 'admin') {
                    categoryLink.parentElement.style.display = 'none';
                } else {
                    categoryLink.parentElement.style.display = 'block';
                }
            }
        }
    } catch (error) {
        console.error('Error checking user role:', error);
        // Hide category link on error
        const categoryLink = document.getElementById('categories-link');
        if (categoryLink) {
            categoryLink.parentElement.style.display = 'none';
        }
    }
}

// Utility function to check if user is authenticated (can be used by other scripts)
async function isAuthenticated() {
    try {
        const response = await fetch('/api/auth/session', {
            credentials: 'include'
        });
        return response.ok;
    } catch (error) {
        return false;
    }
}

// Utility function to get current user info (can be used by other scripts)
async function getCurrentUser() {
    try {
        const response = await fetch('/api/auth/session', {
            credentials: 'include'
        });
        
        if (response.ok) {
            return await response.json();
        }
        return null;
    } catch (error) {
        console.error('Error getting current user:', error);
        return null;
    }
}
