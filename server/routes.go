package server

import (
	"net/http"

	"github.com/scriptmaster/openagent/admin"
	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/projects"
)

// RegisterRoutes sets up all the application routes
func RegisterRoutes(router *http.ServeMux, userService auth.UserServicer, salt string) {
	db := GetDB()
	services := GetServices(db)

	// Static file handlers
	router.HandleFunc("/favicon.ico", CreateFaviconHandler())
	router.HandleFunc("/tsx/css/", CreateTSXCSSHandler())
	router.HandleFunc("/tsx/js/", CreateTSXJSHandler())

	fs := http.FileServer(http.Dir("./static"))
	router.Handle("/static/", http.StripPrefix("/static/", CacheMiddleware(fs)))

	// Maintenance and config routes
	router.HandleFunc("/maintenance", HandleMaintenance)
	router.HandleFunc("/config", HandleConfigPage)
	router.HandleFunc("/config/save", CreateConfigSaveHandler(userService, services.ProjectService))

	// Order: auth, projects, admin
	auth.RegisterAuthRoutes(router, globalTemplates, userService)
	projects.RegisterProjectRoutes(router, globalTemplates, userService, services.DB, services.PDBService)
	admin.RegisterAdminRoutes(router, globalTemplates, auth.IsMaintenanceAuthenticated, UpdateDatabaseConfig, InitDB, salt, GetDB, GetAdminStats)

	// Protected routes
	router.Handle("/dashboard", auth.AuthMiddleware(http.HandlerFunc(CreateDashboardHandler(services.ProjectService))))
	router.Handle("/voice", auth.AuthMiddleware(http.HandlerFunc(CreateVoiceHandler())))
	router.Handle("/agent", auth.AuthMiddleware(http.HandlerFunc(CreateAgentHandler())))

	// Public routes
	router.HandleFunc("/version", CreateVersionHandler())
	router.HandleFunc("/test", CreateTestHandler())
	router.HandleFunc("/index1", CreateIndex1Handler())

	// Health check endpoints for Kubernetes
	router.HandleFunc("/livez", HandleLivez)
	router.HandleFunc("/readyz", HandleReadyz)
	router.HandleFunc("/healthz", HandleHealthz) // Legacy endpoint (disabled)

	router.HandleFunc("/", CreateRootHandler())
}
