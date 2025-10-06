/* Finalized For AI */
package main

import (
	template "esgoja/template"
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	engine := template.NewEngine("tpl")

	// Initial load of templates and start a background watcher
	_ = template.LoadAll(engine)
	go template.DetectChanges(engine)

	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Root: render pages/index.html via top-level name
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data := map[string]any{
			"Title":   "OpenAgent esgoja",
			"Message": "Hello from /",
		}
		if err := engine.Render(w, "index.html", data); err != nil {
			log.Printf("render / error: %v", err)
			http.Error(w, "render error", http.StatusInternalServerError)
			return
		}
	})

	// Test route: render pages/test.html via top-level name
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"Title":   "Test Page",
			"Message": "Hello from /test",
		}
		if err := engine.Render(w, "test.html", data); err != nil {
			log.Printf("render /test error: %v", err)
			http.Error(w, "render error", http.StatusInternalServerError)
			return
		}
	})

	// Static files under /static/ served from ./static
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := "8802"
	log.Printf("Starting esgoja minimal server on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
