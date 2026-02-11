package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"utils/shared"
)

// Handler handles getting avatar selections from all RSVPs
func Handler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get database connection
	db := shared.NewDatabase()
	if !db.IsConfigured() {
		log.Printf("Database not configured")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Return empty array rather than error
		json.NewEncoder(w).Encode(shared.GetAvatarsResponse{
			Success: true,
			Avatars: []shared.GuestAvatar{},
		})
		return
	}

	// Fetch all RSVPs from Supabase using REST API - only verified and attending guests
	apiURL := fmt.Sprintf("%s/rest/v1/rsvps?select=avatar_data&is_attending=eq.true&verified=eq.true", os.Getenv("SUPABASE_URL"))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(shared.GetAvatarsResponse{
			Success: false,
			Avatars: []shared.GuestAvatar{},
		})
		return
	}

	apiKey := os.Getenv("SUPABASE_API_KEY")
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error fetching RSVPs: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(shared.GetAvatarsResponse{
			Success: false,
			Avatars: []shared.GuestAvatar{},
		})
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error from Supabase: status %d", resp.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Return empty array rather than error
		json.NewEncoder(w).Encode(shared.GetAvatarsResponse{
			Success: true,
			Avatars: []shared.GuestAvatar{},
		})
		return
	}

	// Parse RSVP records - use json.RawMessage to handle null/empty avatar_data
	var rsvps []struct {
		AvatarData json.RawMessage `json:"avatar_data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rsvps); err != nil {
		log.Printf("Error decoding RSVPs: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Return empty array rather than error
		json.NewEncoder(w).Encode(shared.GetAvatarsResponse{
			Success: true,
			Avatars: []shared.GuestAvatar{},
		})
		return
	}

	// Collect all avatars from all RSVPs, deduplicating by guest name
	seenGuests := make(map[string]bool)
	avatars := []shared.GuestAvatar{}
	for _, rsvp := range rsvps {
		// Skip null or empty avatar_data
		if rsvp.AvatarData == nil || string(rsvp.AvatarData) == "null" || string(rsvp.AvatarData) == "[]" {
			continue
		}

		// Try to parse avatar_data as array
		var avatarSelections []shared.AvatarSelection
		if err := json.Unmarshal(rsvp.AvatarData, &avatarSelections); err != nil {
			log.Printf("Error parsing avatar_data: %v (data: %s)", err, string(rsvp.AvatarData))
			continue
		}

		for _, avatarSelection := range avatarSelections {
			// Skip if we've already seen this guest (deduplicate by full name)
			if seenGuests[avatarSelection.GuestName] {
				continue
			}
			seenGuests[avatarSelection.GuestName] = true

			// Extract first name (text before first space)
			firstName := avatarSelection.GuestName
			for i, char := range avatarSelection.GuestName {
				if char == ' ' {
					firstName = avatarSelection.GuestName[:i]
					break
				}
			}

			avatars = append(avatars, shared.GuestAvatar{
				Name:    firstName,
				Avatar:  avatarSelection.Avatar,
				Message: avatarSelection.Message,
			})
		}
	}

	log.Printf("Returning %d unique avatars from RSVPs", len(avatars))

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shared.GetAvatarsResponse{
		Success: true,
		Avatars: avatars,
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
