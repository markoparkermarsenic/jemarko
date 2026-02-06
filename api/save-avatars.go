package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"utils/shared"
)

// Handler handles saving avatar selections for guests
func Handler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req shared.SaveAvatarsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate request
	if req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
			Success: false,
			Message: "Email is required",
		})
		return
	}

	if len(req.Avatars) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
			Success: false,
			Message: "At least one avatar selection is required",
		})
		return
	}

	// Validate each avatar selection
	for _, avatar := range req.Avatars {
		if avatar.GuestName == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
				Success: false,
				Message: "Guest name is required for all avatars",
			})
			return
		}
		if avatar.Avatar == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
				Success: false,
				Message: "Avatar is required for all guests",
			})
			return
		}
	}

	// Convert avatars to JSON for storage
	avatarData, err := json.Marshal(req.Avatars)
	if err != nil {
		log.Printf("Error marshaling avatar data: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
			Success: false,
			Message: "Failed to process avatar data",
		})
		return
	}

	// Update the rsvps table with avatar data using Supabase REST API
	db := shared.NewDatabase()
	if !db.IsConfigured() {
		log.Printf("Database not configured")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
			Success: false,
			Message: "Database connection failed",
		})
		return
	}

	// Update avatar_data field in rsvps table
	updatePayload := map[string]interface{}{
		"avatar_data": json.RawMessage(avatarData),
	}

	updateJSON, err := json.Marshal(updatePayload)
	if err != nil {
		log.Printf("Error marshaling update payload: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
			Success: false,
			Message: "Failed to save avatar selections",
		})
		return
	}

	// Use Supabase REST API to update the most recent RSVP for this email
	apiURL := fmt.Sprintf("%s/rest/v1/rsvps?email=eq.%s&order=submitted_at.desc&limit=1",
		os.Getenv("SUPABASE_URL"),
		url.QueryEscape(req.Email))

	updateReq, err := http.NewRequest("PATCH", apiURL, bytes.NewBuffer(updateJSON))
	if err != nil {
		log.Printf("Error creating update request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
			Success: false,
			Message: "Failed to save avatar selections",
		})
		return
	}

	apiKey := os.Getenv("SUPABASE_API_KEY")
	updateReq.Header.Set("apikey", apiKey)
	updateReq.Header.Set("Authorization", "Bearer "+apiKey)
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Prefer", "return=minimal")

	resp, err := http.DefaultClient.Do(updateReq)
	if err != nil {
		log.Printf("Error updating RSVP: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
			Success: false,
			Message: "Failed to save avatar selections",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		// Read error body for debugging
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		log.Printf("Error updating RSVP: status %d, body: %v, url: %s, payload: %s",
			resp.StatusCode, errBody, apiURL, string(updateJSON))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
			Success: false,
			Message: "Failed to save avatar selections",
		})
		return
	}

	log.Printf("Successfully saved avatar selections for %d guests (email: %s)", len(req.Avatars), req.Email)

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shared.SaveAvatarsResponse{
		Success: true,
		Message: "Avatar selections saved successfully",
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
