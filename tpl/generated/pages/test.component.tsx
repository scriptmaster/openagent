export default function Test({page}) {
    return (
<main>
<div className="container-xl">
    <div className="row">
        <div className="col-12">
            <div className="page-header">
                <h1 className="page-title">React Tests - Interactive Components! 🚀</h1>
                <p className="text-muted">Testing various React functionality with JSX transpilation</p>
            </div>
        </div>
    </div>
    <div className="row mb-4">
        <div className="col-12">
            <div className="alert alert-info">
                <h4 className="alert-heading">🎉 React Transpilation Success!</h4>
                <p>This page demonstrates our custom JSX-to-React transpilation system. All components below are written in JSX and automatically converted to React.createElement calls!</p>
                <hr/>
                <p className="mb-0">Check the browser console to see the transpiled React code in action.</p>
            </div>
        </div>
    </div>
    <div className="row mb-4">
        <div className="col-12">
            <h2 className="h3 mb-3">Part 1: Basic React Demos</h2>
        </div>
    </div>
    <div className="row mb-4">
        <div className="col-md-6">
            <div className="card">
                <div className="card-header">
                    <h5 className="card-title mb-0">Demo 1: Counter</h5>
                </div>
                <div className="card-body">
                    <p className="card-text">Simple counter with increment/decrement buttons</p>
                    <div id="component-counter"></div>
                </div>
            </div>
        </div>
    </div>
    <div className="row mb-4">
        <div className="col-md-6">
            <div className="card">
                <div className="card-header">
                    <h5 className="card-title mb-0">Demo 2: Toggle</h5>
                </div>
                <div className="card-body">
                    <p className="card-text">Toggle button that shows/hides content</p>
                    <div id="demo2-toggle"></div>
                </div>
            </div>
        </div>
    </div>
</div>
</main>
    );
}