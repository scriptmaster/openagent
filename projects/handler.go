package projects

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
)

type ProjectHandler struct {
	service ProjectService
}

func NewProjectHandler(service ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// GetURLParam extracts a URL parameter from the request
func GetURLParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

func (h *ProjectHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil || !user.IsAdmin {
		common.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	projects, err := h.service.List()
	if err != nil {
		common.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	common.JSONResponse(w, projects)
}

func (h *ProjectHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil || !user.IsAdmin {
		common.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var project Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		common.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	project.CreatedBy = int64(user.ID)
	id, err := h.service.Create(&project)
	if err != nil {
		common.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	project.ID = id
	common.JSONResponse(w, project)
}

func (h *ProjectHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil || !user.IsAdmin {
		common.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := GetURLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.JSONError(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var project Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		common.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	project.ID = id
	if err := h.service.Update(&project); err != nil {
		common.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	common.JSONResponse(w, project)
}

func (h *ProjectHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil || !user.IsAdmin {
		common.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := GetURLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.JSONError(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		common.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	common.JSONResponse(w, map[string]string{"message": "Project deleted successfully"})
}

func (h *ProjectHandler) HandleGetByDomain(w http.ResponseWriter, r *http.Request) {
	domain := GetURLParam(r, "domain")
	if domain == "" {
		common.JSONError(w, "Domain parameter is required", http.StatusBadRequest)
		return
	}

	project, err := h.service.GetByDomain(domain)
	if err != nil {
		common.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	common.JSONResponse(w, project)
}
