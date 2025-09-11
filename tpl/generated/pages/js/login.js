


        document.addEventListener('alpine:init', () => {
            Alpine.data('loginApp', () => ({
                currentStep: 1,
                email: '',
                otp: '',
                password: '',
                otpSent: false,
                loading: false,
                resendLoading: false,
                message: 'page.Error}' ? 'page.Error}' : '',
                error: 'page.Error}' ? true : false,
                success: false,
                adminEmail: 'page.AdminEmail}',
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
            }));
        });
    
