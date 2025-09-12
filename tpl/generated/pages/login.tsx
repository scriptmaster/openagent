import Layout from '../layouts/layout_pages';

export default function Login({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={`/tsx/css/login.css`} scriptPaths={`/tsx/js/_common.js`}>
<div className="container container-tight py-4">
    <div className="text-center mb-4">
        <p></p>
    </div>
    <div className="card card-md">
        <div className="card-body">
            <h2 className="card-title text-center mb-4">Login</h2>
            
            <form id="emailForm" _="on submit
                        halt the event
                        loadingButton .btn-primary true 'Processing...'
                        post me to '/auth/request-otp'
                        then if it.ok
                            show .alert-success on me
                            show .otp-section on me
                            hide .email-section on me
                        else
                            show .alert-danger on me
                            set .alert-danger .message.textContent to 'Failed to send OTP' on me
                        end
                        loadingButton .btn-primary false 'Send OTP'">
                
                <div className="email-section">
                    <div className="mb-3">
                        <label className="form-label">Email address</label>
                        <div className="input-group input-group-flat">
                            <span className="input-group-text">
                                <i className="ti ti-mail"></i>
                            </span>
                            <input type="email" name="email" className="form-control" placeholder="your@email.com" autocomplete="off" required />
                        </div>
                    </div>
                    
                    <div className="alert alert-danger" style="display: none;">
                        <i className="ti ti-alert-circle me-2"></i>
                        <span className="message"></span>
                    </div>
                    
                    <div className="form-footer">
                        <button type="submit" className="btn btn-primary w-100">
                            <span className="spinner-border spinner-border-sm me-2" role="status" style="display: none;"></span>
                            <span className="btn-text">Send OTP</span>
                        </button>
                    </div>
                </div>
                
            </form>
            
            <div className="alert alert-success" style="display: none;">
                <i className="ti ti-check me-2"></i> A one-time password has been sent to your email.
            </div>
            
            <div className="otp-section" style="display: none;">
                <form id="otpForm" _="on submit
                    halt the event
                    loadingButton .btn-primary true 'Verifying...'
                    post me to '/auth/verify-otp'
                    then if result.ok
                        set .alert-success.textContent to 'Login successful! Redirecting...' on me
                        wait 1s
                        go to /dashboard
                    else
                        show .alert-danger on me
                        set .alert-danger .message.textContent to 'Invalid OTP' on me
                    end
                    loadingButton .btn-primary false 'Verify'">
                    
                    <input type="hidden" name="email" _="on load set my.value to #emailForm input[name='email'].value" />
                    
                    <div className="mb-3">
                        <label className="form-label">One-Time Password</label>
                        <div className="input-group input-group-flat">
                            <span className="input-group-text">
                                <i className="ti ti-key"></i>
                            </span>
                            <input type="text" name="otp" className="form-control" placeholder="Enter your OTP" autocomplete="off" required />
                        </div>
                    </div>
                    
                    <div className="form-footer">
                        <button type="submit" className="btn btn-primary w-100">
                            <span className="spinner-border spinner-border-sm me-2" role="status" style="display: none;"></span>
                            <span className="btn-text">Verify</span>
                        </button>
                    </div>
                </form>
                
                <div className="text-center mt-3">
                    <form id="resendForm" _="on submit
                        halt the event
                        loadingButton .btn-link true 'Sending...'
                        post #emailForm to '/auth/request-otp'
                        then if result.ok
                            set .alert-success.textContent to 'A new OTP has been sent to your email.' on me
                        else
                            show .alert-danger on me
                            set .alert-danger .message.textContent to 'Failed to resend OTP' on me
                        end
                        loadingButton .btn-link false 'Resend OTP'">
                        
                        <button type="submit" className="btn btn-link">
                            <span className="spinner-border spinner-border-sm me-2" role="status" style="display: none;"></span>
                            <i className="ti ti-rotate me-1"></i>
                            <span className="btn-text">Resend OTP</span>
                        </button>
                    </form>
                </div>
            </div>
        </div>
    </div>
    <div className="text-center text-muted mt-3">
        Don't have an account yet? <a href={`mailto:${page.AdminEmail}`} tabIndex="-1">Contact administrator</a>
    </div>
</div>
        </Layout>
    );
}