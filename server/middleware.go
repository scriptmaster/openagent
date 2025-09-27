package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/scriptmaster/openagent/admin"    // For maintenance redirect
	"github.com/scriptmaster/openagent/auth"     // For GetUserFromSession to pass to config page
	"github.com/scriptmaster/openagent/projects" // Assuming ProjectService is here
	// For Handle404 or similar
)

// Define a context key type for the project
type projectCtxKey struct{}

// HostProjectMiddleware checks the request host against configured projects.
// If a project matches, it adds it to the context.
// If no project matches, it serves the configuration page.
// Specific paths can be exempted.
func HostProjectMiddleware(next http.Handler, projectService projects.ProjectService, userService auth.UserServicer, exemptPaths map[string]bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for maintenance mode first
		if IsMaintenanceMode() && !strings.HasPrefix(r.URL.Path, "/maintenance") {
			// Allow access only to maintenance routes
			admin.HandleMaintenance(w, r, globalTemplates, auth.IsMaintenanceAuthenticated) // Assuming templates is accessible
			return
		}

		// Check if the path is exempt, but skip this check for the root path "/"
		if r.URL.Path != "/" {
			for prefix := range exemptPaths {
				if strings.HasPrefix(r.URL.Path, prefix) {
					next.ServeHTTP(w, r) // Serve the original request directly
					return
				}
			}
		}

		host := r.Host // Default to the Host header from the request (e.g., "localhost:8800", "myproject.com")

		// Check if the request is likely coming from behind a proxy (e.g., Nginx)
		// by looking for the X-Forwarded-For header.
		if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			// If X-Forwarded-For is present, it indicates a proxy.
			// In such cases, the original host requested by the client is typically found in X-Forwarded-Host.
			if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
				host = forwardedHost
			}
			// If X-Forwarded-Host is not present, but X-Forwarded-For is,
			// we stick with the initial r.Host. X-Forwarded-For itself contains client IP, not the host domain.
			// The presence of X-Forwarded-For serves as the check for being behind a proxy.
		}
		// Clean the host? Remove port? Depends on how domains are stored. Assuming stored without port for now.
		host = strings.Split(host, ":")[0]

		// Use projectService if it's not nil
		var project *projects.Project
		var err error
		if projectService != nil {
			project, err = projectService.GetByDomain(host)
			if err != nil && err != projects.ErrProjectNotFound {
				log.Printf("Error fetching project by domain '%s': %v", host, err)
				http.Error(w, fmt.Sprintf("Internal Server Error checking project domain. GetByDomain(\"%v\")", host), http.StatusInternalServerError)
				return
			}
		} else {
			// If projectService is nil (DB issue), treat as project not found
			err = projects.ErrProjectNotFound
			log.Printf("Project service not available, cannot check host: %s, %v", host, err)
		}

		if project != nil {
			// Project found, add it to context using the projects package function
			ctx := projects.SetProjectContext(r.Context(), project)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			// Project not found for this host, serve/redirect to config page
			log.Printf("\t → \t → No project found for host '%s', serving config page.", host)
			// Add user info if logged in, for display on config page
			// Use GetUserFromContext now
			user := auth.GetUserFromContext(r.Context())
			ctx := r.Context()
			if user != nil {
				ctx = auth.SetUserContext(ctx, user) // SetUserContext is likely still in auth package
			}
			// log.Printf("\t → \t → \t → 6.X.HPM: HandleConfigPage:")
			HandleConfigPage(w, r.WithContext(ctx))
		}
	})
}
