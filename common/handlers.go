package common

import (
	"net/http"

	"github.com/scriptmaster/openagent/types"
)

// Handle404 handles requests for unconfigured domains or non-existent routes
func Handle404(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface) {
	w.WriteHeader(http.StatusNotFound)
	data := map[string]interface{}{
		"AppName":    "OpenAgent",
		"AppVersion": "1.0.0",
		"Path":       r.URL.Path,
	}
	if err := templates.ExecuteTemplate(w, "404.html", data); err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}
