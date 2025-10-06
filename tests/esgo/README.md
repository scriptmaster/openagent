# ESBuild + Goja Test Server

This is a test server that demonstrates the new esbuild + goja architecture for server-side rendering of React components.

## Architecture

1. **HTML → TSX**: Convert HTML templates to TSX using our existing logic
2. **TSX → JS**: Transform TSX to JS using esbuild with React bundling
3. **JS Execution**: Execute JS using goja runtime
4. **Server-side Rendering**: Render components on the server

## Features

- ✅ ESBuild integration for TSX → JS transformation
- ✅ Goja runtime for JavaScript execution
- ✅ React.createElement support
- ✅ Server-side rendering
- ✅ IIFE format for component isolation

## Running the Test Server

```bash
cd tests/esgo
go mod tidy
go run main.go
```

Then visit:
- http://localhost:8801/test - Test component rendering
- http://localhost:8801/health - Health check
- http://localhost:8801/test.html - Test page with results

## Benefits over Wax

- ✅ No complex template parsing
- ✅ Standard JavaScript execution
- ✅ Better error handling
- ✅ Easier debugging
- ✅ More maintainable
- ✅ Better performance

