function Simple() {
    return (
        React.createElement('div', null, 'Simple Component')
    );
}

///////////////////////////////

// Component prototype methods

    Object.assign({}, Simple.prototype, {
        hey() {
            alert('hey');
        }
    });

