# Static Assets Directory

This directory contains static assets served by the Go web server for the application UI.

## Directory Structure

- `/css`: Contains CSS stylesheets including Tabler UI framework CSS
- `/js`: Contains JavaScript files including Alpine.js
- `/img`: Contains images like logos and icons

## Caching

These static assets are served with a cache duration of 1 day to improve performance.

## Asset Sources

- Tabler UI: https://tabler.io/
- Alpine.js: https://alpinejs.dev/

## Notes

The Go web server serves these files using the following configuration:

```go
// Static file server with caching
staticHandler := http.FileServer(http.Dir("./static"))
http.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Add cache control header
    w.Header().Set("Cache-Control", "public, max-age=86400") // 1 day
    staticHandler.ServeHTTP(w, r)
})))
```
