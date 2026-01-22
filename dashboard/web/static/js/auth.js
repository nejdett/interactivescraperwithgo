// Authentication JavaScript

document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('login-form');
    
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }
});

async function handleLogin(event) {
    event.preventDefault();
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const errorMessage = document.getElementById('error-message');
    const loginText = document.getElementById('login-text');
    const loginSpinner = document.getElementById('login-spinner');
    const submitButton = event.target.querySelector('button[type="submit"]');
    
    // Clear previous errors
    errorMessage.classList.add('hidden');
    errorMessage.textContent = '';
    
    // Show loading state
    submitButton.disabled = true;
    loginText.classList.add('hidden');
    loginSpinner.classList.remove('hidden');
    
    try {
        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username: username,
                password: password
            }),
            credentials: 'include' // Important for cookies
        });
        
        const data = await response.json();
        
        if (response.ok) {
            // Login successful, redirect to dashboard
            window.location.href = '/dashboard';
        } else {
            // Login failed, show error message
            showError(data.message || 'Invalid username or password');
        }
    } catch (error) {
        console.error('Login error:', error);
        showError('Ağ hatası: ' + error.message + '. Sunucunun çalıştığından emin olun.');
    } finally {
        // Reset button state
        submitButton.disabled = false;
        loginText.classList.remove('hidden');
        loginSpinner.classList.add('hidden');
    }
}

function showError(message) {
    const errorMessage = document.getElementById('error-message');
    errorMessage.textContent = message;
    errorMessage.classList.remove('hidden');
}

// Check if user is already logged in (for login page)
async function checkExistingSession() {
    try {
        const response = await fetch('/api/auth/session', {
            credentials: 'include'
        });
        
        if (response.ok) {
            // User is already logged in, redirect to dashboard
            window.location.href = '/dashboard';
        }
    } catch (error) {
        // Session check failed, user needs to login
        console.log('Aktif oturum yok');
    }
}

// Run session check on login page load
if (window.location.pathname === '/login') {
    checkExistingSession();
}
