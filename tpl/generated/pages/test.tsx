import Layout from '../layouts/layout_pages';

export default function Test({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={``} scriptPaths={`/tsx/js/_common.js`}>
<div className="container-xl">
    <div className="row">
        <div className="col-12">
            <div className="page-header">
                <h1 className="page-title">Hyperscript Tests - Debug Test 2!</h1>
                <p className="text-muted">Testing various hyperscript functionality</p>
            </div>
        </div>
    </div>
    <div className="row">
        <div className="col-md-6">
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">Test 1: Basic Click Event</h3>
                </div>
                <div className="card-body">
                    <p>Click the button to test basic hyperscript click handling:</p>
                    <button className="btn" _="on click add .text-success to me then set my.textContent to 'Clicked!'">
                        Click Me
                    </button>
                </div>
            </div>
        </div>
        <div className="col-md-6">
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">Test 2: Show/Hide Toggle</h3>
                </div>
                <div className="card-body">
                    <button className="btn btn-secondary mb-3" _="on click toggle .d-none on #test2-content">
                        Toggle Content
                    </button>
                    <div id="test2-content" className="alert alert-info d-none">
                        This content will be toggled.
                    </div>
                </div>
            </div>
        </div>
    </div>
    <div className="row mt-4">
        <div className="col-md-6">
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">Test 3: Form Input Interaction</h3>
                </div>
                <div className="card-body">
                    <div className="mb-3">
                        <label className="form-label" for="test3-input">Type something:</label>
                        <input type="text" className="form-control" id="test3-input" placeholder="Type here..." _="on input set #test3-output.textContent to me.value"/>
                    </div>
                    <div id="test3-output" className="alert alert-info">
                        Output will appear here...
                    </div>
                </div>
            </div>
        </div>
        <div className="col-md-6">
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">Test 4: Custom Commands</h3>
                </div>
                <div className="card-body">
                    <p>Test our custom hyperscript commands:</p>
                    {/* <button className="btn btn-info" _="on click post me to '/test' then log result">
                        Test Post Command
                    </button> */}
                    <button className="btn btn-warning" _="on click loadingButton">
                        Test Loading Button
                    </button>
                </div>
            </div>
        </div>
    </div>
    <div className="row mt-4">
        {/* <div className="col-md-6">
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">Test 5: CSS Selector with Attributes</h3>
                </div>
                <div className="card-body">
                    <div id="test5-form">
                        <input type="text" id="test5-email" name="test5-email" value="test@example.com" className="form-control mb-2"/>
                        <input type="hidden" id="test5-copy" name="test5-copy" _="on load set my.value to #test5-email.value"/>
                    </div>
                    <div className="mt-2">
                        <small className="text-muted">Hidden field value: <span id="test5-display"></span></small>
                    </div>
                    <button className="btn btn-sm btn-outline-primary" _="on click set #test5-display.textContent to #test5-copy.value">
                        Show Hidden Value
                    </button>
                </div>
            </div>
        </div> */}
        {/* <div className="col-md-6">
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">Test 6: Event Chaining</h3>
                </div>
                <div className="card-body">
                    <button className="btn btn-success" _="on click 
                        add .btn-outline-success to me
                        remove .btn-success from me
                        set my.textContent to 'Processing...'
                        wait 2s
                        set my.textContent to 'Done!'
                        add .btn-success to me
                        remove .btn-outline-success from me">
                        Chain Events
                    </button>
                </div>
            </div>
        </div> */}
    </div>
    {/* <div className="row mt-4">
        <div className="col-12">
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">Test 7: Complex Form Handling</h3>
                </div>
                <div className="card-body">
                    <form id="test7-form">
                        <div className="row">
                            <div className="col-md-6">
                                <div className="mb-3">
                                    <label className="form-label">Name:</label>
                                    <input type="text" name="name" className="form-control" required/>
                                </div>
                            </div>
                            <div className="col-md-6">
                                <div className="mb-3">
                                    <label className="form-label">Email:</label>
                                    <input type="email" name="email" className="form-control" required/>
                                </div>
                            </div>
                        </div>
                        <div className="mb-3">
                            <label className="form-label">Message:</label>
                            <textarea name="message" className="form-control" rows="3"></textarea>
                        </div>
                        <button type="submit" className="btn btn-primary" _="on click 
                            halt the event's default
                            log 'Form submitted'
                            show .alert-success in #test7-form
                            hide .alert-danger in #test7-form">
                            Submit Form
                        </button>
                    </form>
                    <div className="alert alert-success mt-3" style="display: none;">
                        <i className="ti ti-check me-2"></i>Form submitted successfully!
                    </div>
                    <div className="alert alert-danger mt-3" style="display: none;">
                        <i className="ti ti-alert-circle me-2"></i>Form submission failed!
                    </div>
                </div>
            </div>
        </div>
    </div> */}
    <div className="row mt-4">
        <div className="col-12">
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">Console Output</h3>
                </div>
                <div className="card-body">
                    <p className="text-muted">Check the browser console for hyperscript command logs and any errors.</p>
                    <div className="alert alert-info">
                        <i className="ti ti-info-circle me-2"></i>
                        Open Developer Tools (F12) and check the Console tab to see hyperscript activity.
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
        </Layout>
    );
}