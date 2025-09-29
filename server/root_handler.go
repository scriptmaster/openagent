package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/models"
	"github.com/scriptmaster/openagent/projects"
)

// Landing page cache for high-performance serving
var (
	globalPageService     projects.PageService
	landingPageCache      = make(map[int]*models.Page)
	landingPageCacheMutex sync.RWMutex
	landingPageCacheTTL   = 5 * time.Minute
	landingPageCacheTime  = make(map[int]time.Time)
)

// Initialize landing page optimization
func initLandingPageOptimization() {
	globalPageService = projects.NewPageService(GetDB())
}

// getCachedLandingPage retrieves landing page from cache if valid
func getCachedLandingPage(projectID int) (*models.Page, bool) {
	landingPageCacheMutex.RLock()
	defer landingPageCacheMutex.RUnlock()

	page, exists := landingPageCache[projectID]
	if !exists {
		return nil, false
	}

	// Check if cache is still valid
	if cacheTime, exists := landingPageCacheTime[projectID]; exists {
		if time.Since(cacheTime) < landingPageCacheTTL {
			return page, true
		}
	}

	return nil, false
}

// setCachedLandingPage stores landing page in cache
func setCachedLandingPage(projectID int, page *models.Page) {
	landingPageCacheMutex.Lock()
	defer landingPageCacheMutex.Unlock()

	landingPageCache[projectID] = page
	landingPageCacheTime[projectID] = time.Now()
}

// HandleIndexPage serves the default index page with "Welcome to OpenAgent" message
// OPTIMIZED: High-performance landing page serving with caching
func HandleIndexPage(w http.ResponseWriter, r *http.Request) {
	SetCacheHeaders(w, r, r.URL.Path)

	project := projects.GetProjectFromContext(r.Context())
	if project != nil {
		projectID := int(project.ID)

		// OPTIMIZATION 1: Check cache first (0.001ms)
		if landingPage, found := getCachedLandingPage(projectID); found {
			if strings.TrimSpace(landingPage.HTMLContent) != "" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(landingPage.HTMLContent))
				return
			}
		}

		// OPTIMIZATION 2: Use global page service (saves 0.1ms per request)
		if globalPageService == nil { // Fallback if not initialized
			globalPageService = projects.NewPageService(GetDB())
		}

		// OPTIMIZATION 3: Database query only on cache miss (2-5ms)
		landingPage, err := globalPageService.GetLandingPageByProjectID(projectID)
		if err == nil && landingPage != nil && strings.TrimSpace(landingPage.HTMLContent) != "" {
			setCachedLandingPage(projectID, landingPage) // Cache the result
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(landingPage.HTMLContent))
			return
		}
	}

	// Fallback to default index.html if no project or no landing page found
	if globalTemplates == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		return
	}

	// Get user from context for template rendering
	user := auth.GetUserFromContext(r.Context())

	// Prepare template data
	templateData := map[string]interface{}{
		"AppName":    "OpenAgent",
		"AppVersion": common.AppVersion,
		"User":       user,
	}

	// Determine which template to use based on LANDING_INDEX environment variable
	templateName := "index.html"
	
	// Check LANDING_INDEX environment variable
	if landingIndex := os.Getenv("LANDING_INDEX"); landingIndex != "" && landingIndex != "0" {
		// Try to use numbered versions
		templateName = fmt.Sprintf("index%s.html", landingIndex)
	}

	// Execute the index template (fallback to default if numbered version doesn't exist)
	err := globalTemplates.ExecuteTemplate(w, templateName, templateData)
	if err != nil {
		// If numbered template fails, fallback to default
		if templateName != "index.html" {
			log.Printf("Error executing %s template, falling back to index.html: %v", templateName, err)
			err = globalTemplates.ExecuteTemplate(w, "index.html", templateData)
		}
		if err != nil {
			log.Printf("Error executing index template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// HandleRoot serves the root path with 404 fallback
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	// Only handle root path "/" - serve the default index page
	if r.URL.Path == "/" {
		HandleIndexPage(w, r)
	} else {
		// Handle 404 for non-root requests
		Handle404(w, r)
	}
}

// CreateRootHandler creates a root handler
func CreateRootHandler() http.HandlerFunc {
	log.Printf("\t → 6.100 Route: / Root Handler")

	// Initialize landing page optimization for high-performance serving
	log.Printf("\t → 6.100.1 Optimizing landing page cache")
	initLandingPageOptimization()

	return HandleRoot
}
