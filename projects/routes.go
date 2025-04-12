package projects

import (
	"html/template"
	"net/http"

	"github.com/scriptmaster/openagent/auth"
)

// RegisterProjectRoutes registers all project-related routes
func RegisterProjectRoutes(mux *http.ServeMux, templates *template.Template, userService *auth.UserService) {
	// Project routes
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		// Get user from session
		user, err := userService.GetUserFromSession(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user to request context
		ctx := auth.SetUserContext(r.Context(), user)
		HandleProjects(w, r.WithContext(ctx), templates)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get user from session
		user, err := userService.GetUserFromSession(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user to request context
		ctx := auth.SetUserContext(r.Context(), user)
		HandleIndex(w, r.WithContext(ctx), templates)
	})
}
