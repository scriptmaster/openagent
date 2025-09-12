
setTimeout(() => {
console.log('🚀 Login app script loading...');
Alpine.data('loginApp', () => {
    console.log('🚀 loginApp function called, creating component...');
    return {
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
            currentAction: 'otp',
            notification: {
                visible: false,
                type: 'success',
                title: '',
                message: '',
                timeout: null
            },
        
        get buttonText() {
            if (this.loading) return 'Processing...';
            if (this.currentAction === 'otp') {
                return this.otpSent ? 'Verify OTP' : 'Send OTP';
            }
            return 'Login with Password';
        },
        
        init() {
            console.log('🚀 loginApp component initialized!');
            console.log('🚀 Current step:', this.currentStep);
            console.log('🚀 Current action:', this.currentAction);
            console.log('🚀 Email:', this.email);
            console.log('🚀 Message:', this.message);
            
            // Debug: Check what elements are found
            const step1Elements = document.querySelectorAll('[data-show="currentStep === 1"]');
            const step2Elements = document.querySelectorAll('[data-show="currentStep === 2"]');
            const stepIndicators = document.querySelectorAll('.step-indicator');
            
            console.log('🔍 Found step 1 elements:', step1Elements.length, step1Elements);
            console.log('🔍 Found step 2 elements:', step2Elements.length, step2Elements);
            console.log('🔍 Found step indicators:', stepIndicators.length, stepIndicators);
            
            // Debug: Check current display states
            step1Elements.forEach((el, index) => {
                console.log(`🔍 Step 1 element ${index} display:`, el.style.display, 'computed:', window.getComputedStyle(el).display);
            });
            step2Elements.forEach((el, index) => {
                console.log(`🔍 Step 2 element ${index} display:`, el.style.display, 'computed:', window.getComputedStyle(el).display);
            });
        },
        
        async requestOtp() {
            if (!this.email) return;
            
            this.loading = true;
            this.message = '';
            this.error = false;
            this.success = false;
            try {
                const response = await fetch('/auth/request-otp', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ email: this.email }),
                });
                
                const data = await response.json();
                if (data.success) {
                    this.otpSent = true;
                    this.success = true;
                    this.currentStep = 2;
                    this.$nextTick(() => {
                        document.getElementById('otp-step2-input')?.focus();
                    });
                } else {
                    this.error = true;
                    this.message = data.message || 'Failed to send OTP. Please try again.';
                    if (data.data && data.data.otp_error) {
                        this.currentAction = 'password';
                    }
                }
            } catch (error) {
                console.error('Error:', error);
                this.error = true;
                this.message = 'An error occurred. Please try again.';
            } finally {
                this.loading = false;
            }
        },
        
        async verifyOtp() {
            if (!this.otp) return;
            
            this.loading = true;
            this.message = '';
            this.error = false;
            this.success = false;
            try {
                const response = await fetch('/auth/verify-otp', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ 
                        email: this.email,
                        otp: this.otp 
                    }),
                });
                
                const data = await response.json();
                if (data.success) {
                    this.success = true;
                    this.message = data.message || 'OTP verified successfully.';
                    if (data.redirect) {
                        window.location.href = data.redirect;
                    } else {
                        window.location.href = '/';
                    }
                } else {
                    this.error = true;
                    this.message = data.message || 'Invalid OTP. Please try again.';
                    this.otp = '';
                }
            } catch (error) {
                console.error('Error:', error);
                this.error = true;
                this.message = 'An error occurred during OTP verification. Please try again.';
            } finally {
                this.loading = false;
            }
        },
        
        async resendOtp() {
            if (!this.email) return;
            
            this.resendLoading = true;
            try {
                const response = await fetch('/auth/request-otp', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ email: this.email }),
                });
                
                const data = await response.json();
                if (data.success) {
                    this.showNotification('success', 'Success', 'A new OTP has been sent to your email.');
                } else {
                    this.showNotification('danger', 'Error', data.message || 'Failed to resend OTP. Please try again.');
                }
            } catch (error) {
                console.error('Error:', error);
                this.showNotification('danger', 'Error', 'An error occurred. Please try again.');
            } finally {
                this.resendLoading = false;
            }
        },
        
        async passwordLogin() {
            if (!this.email || !this.password) return;
            
            this.loading = true;
            this.message = '';
            this.error = false;
            this.success = false;
            try {
                const response = await fetch('/auth/password-login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ 
                        email: this.email,
                        password: this.password 
                    }),
                });
                
                const data = await response.json();
                if (data.success) {
                    this.success = true;
                    this.message = data.message || 'Login successful.';
                    if (data.redirect) {
                        window.location.href = data.redirect;
                    } else {
                        window.location.href = '/';
                    }
                } else {
                    this.error = true;
                    this.message = data.message || 'Invalid email or password. Please try again.';
                    this.password = '';
                }
            } catch (error) {
                console.error('Error:', error);
                this.error = true;
                this.message = 'An error occurred during password login. Please try again.';
            } finally {
                this.loading = false;
            }
        },
        
        showNotification(type, title, message) {
            this.notification.type = type;
            this.notification.title = title;
            this.notification.message = message;
            this.notification.visible = true;
            
            if (this.notification.timeout) {
                clearTimeout(this.notification.timeout);
            }
            
            this.notification.timeout = setTimeout(() => {
                this.closeNotification();
            }, 5000);
        },
        
        closeNotification() {
            this.notification.visible = false;
        }
    };
});
}, 100);

