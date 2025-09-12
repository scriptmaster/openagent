
// Hyperscript login functionality
console.log('🚀 Hyperscript login app loading...');

// Initialize login state
let loginState = {
    currentStep: 1,
    email: '',
    otp: '',
    password: '',
    otpSent: false,
    loading: false,
    resendLoading: false,
    message: '',
    error: false,
    success: false,
    adminEmail: 'admin@example.com',
    currentAction: 'otp'
};

// Helper functions
function showNotification(type, title, message) {
    console.log(`📢 ${type.toUpperCase()}: ${title} - ${message}`);
    // You can add a toast notification here
}

function updateButtonText() {
    if (loginState.loading) return 'Processing...';
    if (loginState.currentAction === 'otp') {
        return loginState.otpSent ? 'Verify OTP' : 'Send OTP';
    }
    return 'Login with Password';
}

function updateStepDisplay() {
    const step1Elements = document.querySelectorAll('[data-step="1"]');
    const step2Elements = document.querySelectorAll('[data-step="2"]');
    
    step1Elements.forEach(el => {
        el.style.display = loginState.currentStep === 1 ? 'block' : 'none';
    });
    
    step2Elements.forEach(el => {
        el.style.display = loginState.currentStep === 2 ? 'block' : 'none';
    });
    
    // Update step indicators
    const indicators = document.querySelectorAll('.step-indicator');
    indicators.forEach((indicator, index) => {
        if (index + 1 <= loginState.currentStep) {
            indicator.classList.add('active');
        } else {
            indicator.classList.remove('active');
        }
    });
}

// Login functions
async function requestOtp() {
    if (!loginState.email) return;
    
    loginState.loading = true;
    loginState.message = '';
    loginState.error = false;
    
    try {
        const response = await fetch('/auth/request-otp', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email: loginState.email })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            loginState.otpSent = true;
            loginState.currentStep = 2;
            showNotification('success', 'OTP Sent', 'Please check your email for the one-time password.');
            updateStepDisplay();
        } else {
            loginState.error = true;
            loginState.message = data.error || 'Failed to send OTP';
        }
    } catch (error) {
        loginState.error = true;
        loginState.message = 'Network error. Please try again.';
    } finally {
        loginState.loading = false;
        updateStepDisplay();
    }
}

async function verifyOtp() {
    if (!loginState.otp) return;
    
    loginState.loading = true;
    loginState.message = '';
    loginState.error = false;
    
    try {
        const response = await fetch('/auth/verify-otp', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                email: loginState.email, 
                otp: loginState.otp 
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            loginState.success = true;
            loginState.message = 'Login successful! Redirecting...';
            showNotification('success', 'Success', 'Login successful! Redirecting...');
            
            // Redirect to dashboard or intended page
            setTimeout(() => {
                window.location.href = '/dashboard';
            }, 1000);
        } else {
            loginState.error = true;
            loginState.message = data.error || 'Invalid OTP';
        }
    } catch (error) {
        loginState.error = true;
        loginState.message = 'Network error. Please try again.';
    } finally {
        loginState.loading = false;
        updateStepDisplay();
    }
}

async function passwordLogin() {
    if (!loginState.email || !loginState.password) return;
    
    loginState.loading = true;
    loginState.message = '';
    loginState.error = false;
    
    try {
        const response = await fetch('/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                email: loginState.email, 
                password: loginState.password 
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            loginState.success = true;
            loginState.message = 'Login successful! Redirecting...';
            showNotification('success', 'Success', 'Login successful! Redirecting...');
            
            // Redirect to dashboard or intended page
            setTimeout(() => {
                window.location.href = '/dashboard';
            }, 1000);
        } else {
            loginState.error = true;
            loginState.message = data.error || 'Invalid credentials';
        }
    } catch (error) {
        loginState.error = true;
        loginState.message = 'Network error. Please try again.';
    } finally {
        loginState.loading = false;
        updateStepDisplay();
    }
}

async function resendOtp() {
    if (!loginState.email) return;
    
    loginState.resendLoading = true;
    
    try {
        const response = await fetch('/auth/request-otp', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email: loginState.email })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showNotification('success', 'OTP Resent', 'A new OTP has been sent to your email.');
        } else {
            showNotification('error', 'Error', data.error || 'Failed to resend OTP');
        }
    } catch (error) {
        showNotification('error', 'Error', 'Network error. Please try again.');
    } finally {
        loginState.resendLoading = false;
        updateStepDisplay();
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    console.log('🚀 Login page initialized with hyperscript');
    updateStepDisplay();
});

