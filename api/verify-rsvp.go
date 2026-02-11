package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"utils/shared"
)

// VerifyRSVPRequest represents the verification request
type VerifyRSVPRequest struct {
	Email  string `json:"email"`
	APIKey string `json:"apiKey"`
}

// Handler handles RSVP verification requests (admin only)
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

	var req VerifyRSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request format",
		})
		return
	}

	// Verify API key
	adminAPIKey := os.Getenv("ADMIN_API_KEY")
	if adminAPIKey == "" {
		log.Printf("ADMIN_API_KEY not configured")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Server configuration error",
		})
		return
	}

	if req.APIKey != adminAPIKey {
		log.Printf("Invalid API key provided for verification attempt")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid API key",
		})
		return
	}

	// Validate email
	if req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Email is required",
		})
		return
	}

	// Update RSVP in database to set verified = true
	db := shared.NewDatabase()
	if !db.IsConfigured() {
		log.Printf("Database not configured")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Database not configured",
		})
		return
	}

	// Update the RSVP using Supabase REST API
	supabaseURL := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_API_KEY")

	updateData := map[string]interface{}{
		"verified": true,
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		log.Printf("Error marshaling update data: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Server error",
		})
		return
	}

	// Update using email filter
	apiURL := fmt.Sprintf("%s/rest/v1/rsvps?email=eq.%s", supabaseURL, req.Email)
	updateReq, err := http.NewRequest("PATCH", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating update request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Server error",
		})
		return
	}

	updateReq.Header.Set("apikey", apiKey)
	updateReq.Header.Set("Authorization", "Bearer "+apiKey)
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Prefer", "return=minimal")

	resp, err := http.DefaultClient.Do(updateReq)
	if err != nil {
		log.Printf("Error executing update request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to update RSVP",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		log.Printf("Supabase returned status %d when updating RSVP", resp.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to update RSVP",
		})
		return
	}

	log.Printf("✓ RSVP verified for email: %s", req.Email)

	// Send confirmation email to the now-verified guest
	// Fetch the RSVP to get full details
	fetchURL := fmt.Sprintf("%s/rest/v1/rsvps?email=eq.%s&select=*", supabaseURL, req.Email)
	fetchReq, err := http.NewRequest("GET", fetchURL, nil)
	if err == nil {
		fetchReq.Header.Set("apikey", apiKey)
		fetchReq.Header.Set("Authorization", "Bearer "+apiKey)

		if fetchResp, err := http.DefaultClient.Do(fetchReq); err == nil {
			defer fetchResp.Body.Close()

			var rsvps []shared.RSVPRequest
			if err := json.NewDecoder(fetchResp.Body).Decode(&rsvps); err == nil && len(rsvps) > 0 {
				rsvp := rsvps[0]
				rsvp.Verified = true

				// Send confirmation email
				if err := shared.SendConfirmationEmail(rsvp); err != nil {
					log.Printf("Failed to send confirmation email after verification: %v", err)
				} else {
					log.Printf("✓ Sent confirmation email to verified guest: %s", req.Email)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "RSVP verified successfully",
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
