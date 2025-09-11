export default function Login({page}: {page: Page}) {
    return (
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>{page.AppName} - Login</title>
<link rel="stylesheet" href="/static/css/tabler.min.css" />
<link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
    <link rel="stylesheet" href="/tsx/css/login.css" />
</head>
<body className="border-top-wide border-primary d-flex flex-column">
    <div className="page">
        <div className="container container-tight py-4">
            <div className="text-center mb-4">
                <p></p>
            </div>
            <div className="card card-md" x-data="loginApp()">
                <div className="card-body">
                    {page.Error && (
                    <div className="alert alert-danger" role="alert">
                        <i className="ti ti-alert-circle me-2"></i> {page.Error}
                    </div>
                    )}
                    <div className="step" :className="{ 'active': currentStep === 1 }">
                        <h2 className="card-title text-center mb-4">Login</h2>
                        <form @submit.prevent="currentAction === 'otp' ? requestOtp() : passwordLogin()">
                            <div className="mb-3">
                                <label className="form-label">Email address</label>
                                <div className="input-group input-group-flat">
                                    <span className="input-group-text">
                                        <i className="ti ti-mail"></i>
                                    </span>
                                    <input type="email" className="form-control" x-model="email" placeholder="your@email.com" autocomplete="off" required/>
                                </div>
                            </div>
                            <div className="form-group" x-show="currentAction === 'password'">
                                <label for="password">Password</label>
                                <input type="password" id="password" name="password" x-model="password" :required="currentAction === 'password'"/>
                            </div>
                            <div className="form-message" x-text="message" :className="{ 'error': error, 'success': success }"></div>
                            <div className="form-footer">
                                <button type="submit" className="btn btn-primary w-100" :disabled="loading">
                                    <span x-show="loading" className="spinner-border spinner-border-sm me-2" role="status"></span>
                                    <span x-text="loading ? 'Processing...' : buttonText"></span>
                                </button>
                            </div>
                        </form>
                    </div>
                    <div className="step" :className="{ 'active': currentStep === 2 }">
                        <h2 className="card-title text-center mb-4">Verify OTP</h2>
                        <div className="alert alert-success mb-3">
                            <i className="ti ti-check me-2"></i> A one-time password has been sent to your email.
                        </div>
                        <form @submit.prevent="verifyOtp">
                            <div className="mb-3">
                                <label className="form-label">One-Time Password</label>
                                <div className="input-group input-group-flat">
                                    <span className="input-group-text">
                                        <i className="ti ti-key"></i>
                                    </span>
                                    <input id="otp-step2-input" type="text" className="form-control" x-model="otp" placeholder="Enter your OTP" autocomplete="off" required/>
                                </div>
                            </div>
                            <div className="form-footer">
                                <button type="submit" className="btn btn-primary w-100" :disabled="loading">
                                    <span x-show="loading" className="spinner-border spinner-border-sm me-2" role="status"></span>
                                    <span x-text="loading ? 'Verifying...' : 'Verify'"></span>
                                </button>
                            </div>
                            <div className="text-center mt-3">
                                <button type="button" className="btn btn-link" @click="resendOtp" :disabled="loading">
                                    <span x-show="resendLoading" className="spinner-border spinner-border-sm me-2" role="status"></span>
                                    <i className="ti ti-rotate me-1" x-show="!resendLoading"></i>
                                    <span x-text="resendLoading ? 'Sending...' : 'Resend OTP'"></span>
                                </button>
                            </div>
                        </form>
                    </div>
                    {/* Notification template */}
                    <template x-teleport="body">
                        <div x-show="notification.visible" 
                             x-transition:enter="transition ease-out duration-300"
                             x-transition:enter-start="opacity-0" 
                             x-transition:enter-end="opacity-100"
                             x-transition:leave="transition ease-in duration-300"
                             x-transition:leave-start="opacity-100" 
                             x-transition:leave-end="opacity-0"
                             className="position-fixed bottom-0 end-0 p-3" 
                             style="z-index: 5000">
                            <div :className="`alert alert-${notification.type} alert-dismissible`" role="alert">
                                <div className="d-flex">
                                    <div>
                                        <i :className="`ti ti-${notification.type === 'success' ? 'check' : 'alert-circle'} alert-icon`"></i>
                                    </div>
                                    <div>
                                        <h4 className="alert-title" x-text="notification.title"></h4>
                                        <div className="text-muted" x-text="notification.message"></div>
                                    </div>
                                </div>
                                <a className="btn-close" @click="closeNotification" aria-label="close"></a>
                            </div>
                        </div>
                    </template>
                </div>
            </div>
            
            <div className="text-center text-muted mt-3">
                Don't have an account yet? <a href="mailto:{page.AdminEmail}" tabindex="-1">Contact administrator</a>
            </div>
        </div>
    </div>
    {/* Alpine.js */}
    {/* Tabler Core */}
    
    <div className="text-center text-muted mt-5">
        <p className="text-muted small">Version {page.AppVersion}</p>
    </div>
<script src="/tsx/js/login.js"></script>
    );
}