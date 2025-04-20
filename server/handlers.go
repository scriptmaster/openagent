package server

import (
	"log"
	"net/http"
)

// HandleVoicePage serves the voice agent page
func HandleVoicePage(w http.ResponseWriter, r *http.Request) {
	// Ensure templates are initialized (assuming 'templates' is the global var)
	if templates == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		log.Println("Error: HandleVoicePage called before templates were initialized")
		return
	}

	// Execute the voice template
	// You might want to pass data similar to other pages if needed (e.g., AppName, User)
	err := templates.ExecuteTemplate(w, "voice.html", nil) // Passing nil data for now
	if err != nil {
		log.Printf("Error executing voice template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
