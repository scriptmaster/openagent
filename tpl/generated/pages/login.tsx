import Layout from '../layouts/layout_pages';

export default function Login({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={`/tsx/css/login.css`} scriptPaths={`/tsx/js/_common.js,/tsx/js/login.js`}>
<div className="container container-tight py-4">
    <div className="text-center mb-4">
        <p></p>
    </div>
    <div className="card card-md" data-alpine="loginApp()">
        <div className="card-body">
            {/* Wizard Step Indicator */}
            <div className="wizard-steps mb-4">
                <div className="row">
                    <div className="col-6">
                        <div className="step-indicator" data-className="currentStep >= 1 ? 'active' : ''">
                            <div className="step-number">1</div>
                            <div className="step-label">Email</div>
                        </div>
                    </div>
                    <div className="col-6">
                        <div className="step-indicator" data-className="currentStep >= 2 ? 'active' : ''">
                            <div className="step-number">2</div>
                            <div className="step-label">Verify</div>
                        </div>
                    </div>
                </div>
            </div>
            
            <div className="step" data-show="currentStep === 1">
                <h2 className="card-title text-center mb-4">Login</h2>
                <form data-submit-prevent="currentAction === 'otp' ? requestOtp() : passwordLogin()">
                    <div className="mb-3">
                        <label className="form-label">Email address</label>
                        <div className="input-group input-group-flat">
                            <span className="input-group-text">
                                <i className="ti ti-mail"></i>
                            </span>
                            <input type="email" className="form-control" data-model="email" placeholder="your@email.com" autocomplete="off" required />
                        </div>
                    </div>
                    <div className="form-group" data-show="currentAction === 'password'">
                        <label htmlFor="password">Password</label>
                        <input type="password" id="password" name="password" data-model="password" data-required="currentAction === 'password'" />
                    </div>
                    <div className="form-message" data-text="message" data-className="{ 'error': error, 'success': success }"></div>
                    <div className="form-footer">
                        <button type="submit" className="btn btn-primary w-100" data-disabled="loading">
                            <span data-show="loading" className="spinner-border spinner-border-sm me-2" role="status"></span>
                            <span data-text="loading ? 'Processing...' : buttonText"></span>
                        </button>
                    </div>
                </form>
            </div>
            <div className="step" data-show="currentStep === 2">
                <h2 className="card-title text-center mb-4">Verify OTP</h2>
                <div className="alert alert-success mb-3">
                    <i className="ti ti-check me-2"></i> A one-time password has been sent to your email.
                </div>
                <form data-submit-prevent="verifyOtp">
                    <div className="mb-3">
                        <label className="form-label">One-Time Password</label>
                        <div className="input-group input-group-flat">
                            <span className="input-group-text">
                                <i className="ti ti-key"></i>
                            </span>
                            <input id="otp-step2-input" type="text" className="form-control" data-model="otp" placeholder="Enter your OTP" autocomplete="off" required />
                        </div>
                    </div>
                    <div className="form-footer">
                        <button type="submit" className="btn btn-primary w-100" data-disabled="loading">
                            <span data-show="loading" className="spinner-border spinner-border-sm me-2" role="status"></span>
                            <span data-text="loading ? 'Verifying...' : 'Verify'"></span>
                        </button>
                    </div>
                    <div className="text-center mt-3">
                        <button type="button" className="btn btn-link" data-click="resendOtp" data-disabled="loading">
                            <span data-show="resendLoading" className="spinner-border spinner-border-sm me-2" role="status"></span>
                            <i className="ti ti-rotate me-1" data-show="!resendLoading"></i>
                            <span data-text="resendLoading ? 'Sending...' : 'Resend OTP'"></span>
                        </button>
                    </div>
                </form>
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