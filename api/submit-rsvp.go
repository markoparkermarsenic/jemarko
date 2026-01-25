package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/markoparkermarsenic/jemarko/api/shared"
)

// Handler handles RSVP submission requests
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
		json.NewEncoder(w).Encode(shared.RSVPResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate email
	if !shared.IsValidEmail(req.Email) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(shared.RSVPResponse{
			Success: false,
			Message: "Invalid email address",
		})
		return
	}

	// If attending, validate guest list
	if req.IsAttending {
		// Validate that at least one guest is attending
		if len(req.AttendingGuests) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.RSVPResponse{
				Success: false,
				Message: "At least one guest must be specified when attending",
			})
			return
		}

		// Load guest list to validate attending guests
		db := shared.NewDatabase()
		guestList, err := db.LoadGuests()
		if err != nil {
			log.Printf("Error loading guests: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.RSVPResponse{
				Success: false,
				Message: "Server error - please try again",
			})
			return
		}

		// Validate that all attending guests exist in the guest list
		for _, attendingGuest := range req.AttendingGuests {
			if !shared.IsGuestInList(attendingGuest, guestList) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(shared.RSVPResponse{
					Success: false,
					Message: fmt.Sprintf("Guest '%s' is not on the guest list", attendingGuest),
				})
				return
			}
		}
	}

	// Save RSVP to database
	db := shared.NewDatabase()
	if err := db.SaveRSVP(req); err != nil {
		log.Printf("Failed to save RSVP to database: %v", err)
		// Continue even if database save fails
	}

	// Send confirmation email
	if err := shared.SendConfirmationEmail(req); err != nil {
		log.Printf("Failed to send confirmation email: %v", err)
		// Don't fail the request if email fails - just log it
	}

	if req.IsAttending {
		log.Printf("✓ RSVP completed (ATTENDING): %s (%s) - Guests: %v - Diet: %s",
			req.Name, req.Email, req.AttendingGuests, req.Diet)
	} else {
		log.Printf("✓ RSVP completed (NOT ATTENDING): %s (%s)", req.Name, req.Email)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shared.RSVPResponse{
		Success: true,
		Message: "RSVP submitted successfully",
	})
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
