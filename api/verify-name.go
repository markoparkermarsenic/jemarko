package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"api/shared"
)

// Handler handles name verification requests
func Handler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req shared.RSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(shared.VerifyNameResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate name is not empty
	if strings.TrimSpace(req.Name) == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(shared.VerifyNameResponse{
			Success: false,
			Message: "Name cannot be empty",
		})
		return
	}

	// Load guest list
	db := shared.NewDatabase()
	guestList, err := db.LoadGuests()
	if err != nil {
		log.Printf("Error loading guests: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(shared.VerifyNameResponse{
			Success: false,
			Message: "Server error - please try again",
		})
		return
	}

	// Search for the guest
	foundGuest := shared.FindGuest(req.Name, guestList)

	w.Header().Set("Content-Type", "application/json")

	if foundGuest != nil {
		log.Printf("Guest found: %s (ID: %s)", foundGuest.Name, foundGuest.ID)
		json.NewEncoder(w).Encode(shared.VerifyNameResponse{
			Success: true,
			Message: "Guest found",
		})
	} else {
		log.Printf("Guest not found: %s", req.Name)

		// Send notification to admin about unlisted guest
		ipAddress := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ipAddress = forwarded
		}
		userAgent := r.Header.Get("User-Agent")

		go shared.SendUnlistedGuestNotification(req, ipAddress, userAgent)

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(shared.VerifyNameResponse{
			Success: false,
			Message: "Name not found on the guest list. Please check the spelling or contact us.",
		})
	}
}

// setCORSHeaders sets CORS headers for cross-origin requests
func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
