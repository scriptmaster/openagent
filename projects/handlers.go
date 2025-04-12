package projects

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/scriptmaster/openagent/auth"
)

// HandleProjectsAPI handles the /api/projects endpoint
func HandleProjectsAPI(w http.ResponseWriter, r *http.Request, store *ProjectStore) {
	if r.Method == http.MethodGet {
		projects := store.ListProjects()
		json.NewEncoder(w).Encode(projects)
		return
	}

	if r.Method == http.MethodPost {
		var project Project
		if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		project.ID = uuid.New().String()
		project.CreatedAt = time.Now()
		project.UpdatedAt = time.Now()

		if err := store.CreateProject(&project); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(project)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleProjectAPI handles the /api/projects/{id} endpoint
func HandleProjectAPI(w http.ResponseWriter, r *http.Request, store *ProjectStore) {
	id := r.URL.Path[len("/api/projects/"):]

	switch r.Method {
	case http.MethodGet:
		project, err := store.GetProject(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(project)

	case http.MethodPut:
		var project Project
		if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		project.ID = id
		project.UpdatedAt = time.Now()

		if err := store.UpdateProject(&project); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(project)

	case http.MethodDelete:
		if err := store.DeleteProject(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleProjects handles the /projects route
func HandleProjects(w http.ResponseWriter, r *http.Request, templates *template.Template) {
	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create a new project store
	store, err := NewProjectStore("data")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get all projects
	projects := store.ListProjects()

	// Prepare template data using ProjectsPageData struct
	data := ProjectsPageData{
		AppName:    "OpenAgent",
		PageTitle:  "Projects",
		User:       *user,
		Projects:   projects,
		AppVersion: "1.0.0", // TODO: Get actual version
	}

	// Execute the template
	if err := templates.ExecuteTemplate(w, "projects.html", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleIndex handles the root route
func HandleIndex(w http.ResponseWriter, r *http.Request, templates *template.Template) {
	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Create a new project store
	store, err := NewProjectStore("data")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get all projects
	projects := store.ListProjects()

	// Prepare template data using ProjectsPageData struct
	data := ProjectsPageData{
		AppName:    "OpenAgent",
		PageTitle:  "Dashboard",
		User:       *user,
		Projects:   projects,
		AppVersion: "1.0.0", // TODO: Get actual version
	}

	// Execute the template
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
