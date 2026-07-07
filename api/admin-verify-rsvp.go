package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"utils/shared"
)

// AdminVerifyRequest is the payload for verifying or rejecting an
// unverified RSVP from the admin panel.
//
// On "verify", the admin may supply corrected canonical names (matched
// against the guests table in the UI) so that the dashboard's
// name-based correlation finds these guests from now on.
type AdminVerifyRequest struct {
	Action          string   `json:"action"` // "verify" | "reject"
	Email           string   `json:"email"`
	Name            string   `json:"name,omitempty"`            // corrected submitter name (optional)
	AttendingGuests []string `json:"attendingGuests,omitempty"` // corrected attending guest names (optional)
}

type AdminVerifyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// avrRSVPRow mirrors the rsvps table columns (snake_case, unlike
// shared.RSVPRequest whose JSON tags are camelCase for the public API).
type avrRSVPRow struct {
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	IsAttending     bool     `json:"is_attending"`
	AttendingGuests []string `json:"attending_guests"`
	Diet            string   `json:"diet"`
}

func avrValidateToken(token, secret string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))
	if !hmac.Equal([]byte(parts[1]), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		return false
	}
	pp := strings.SplitN(parts[0], ":", 2)
	if len(pp) != 2 {
		return false
	}
	ts, err := strconv.ParseInt(pp[1], 10, 64)
	if err != nil {
		return false
	}
	return time.Since(time.Unix(ts, 0)) <= 8*time.Hour
}

func avrJSON(w http.ResponseWriter, code int, success bool, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(AdminVerifyResponse{success, message})
}

func avrCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func Handler(w http.ResponseWriter, r *http.Request) {
	avrCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		avrJSON(w, http.StatusInternalServerError, false, "Server configuration error")
		return
	}
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" || !avrValidateToken(token, adminPassword) {
		avrJSON(w, http.StatusUnauthorized, false, "Unauthorized")
		return
	}

	var req AdminVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		avrJSON(w, http.StatusBadRequest, false, "Invalid request format")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		avrJSON(w, http.StatusBadRequest, false, "email is required")
		return
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_API_KEY")
	if supabaseURL == "" || supabaseKey == "" {
		avrJSON(w, http.StatusInternalServerError, false, "Database not configured")
		return
	}

	switch req.Action {
	case "reject":
		if err := avrDeleteRSVP(supabaseURL, supabaseKey, req.Email); err != nil {
			log.Printf("Error rejecting RSVP %s: %v", req.Email, err)
			avrJSON(w, http.StatusInternalServerError, false, "Failed to reject RSVP")
			return
		}
		log.Printf("Admin rejected unverified RSVP: %s", req.Email)
		avrJSON(w, http.StatusOK, true, "RSVP rejected and removed")

	case "verify":
		rows, err := avrVerifyRSVP(supabaseURL, supabaseKey, req)
		if err != nil {
			log.Printf("Error verifying RSVP %s: %v", req.Email, err)
			avrJSON(w, http.StatusInternalServerError, false, "Failed to verify RSVP")
			return
		}
		if len(rows) == 0 {
			avrJSON(w, http.StatusNotFound, false, "No RSVP found for that email")
			return
		}
		log.Printf("Admin verified RSVP for %s (guests: %v)", req.Email, rows[0].AttendingGuests)

		// Best-effort confirmation email to the now-verified guest
		row := rows[0]
		confirmation := shared.RSVPRequest{
			Name:            row.Name,
			Email:           row.Email,
			IsAttending:     row.IsAttending,
			AttendingGuests: row.AttendingGuests,
			Diet:            row.Diet,
			Verified:        true,
		}
		if err := shared.SendConfirmationEmail(confirmation); err != nil {
			log.Printf("Failed to send confirmation email after admin verification: %v", err)
			avrJSON(w, http.StatusOK, true, "RSVP verified (confirmation email failed to send)")
			return
		}
		avrJSON(w, http.StatusOK, true, fmt.Sprintf("RSVP verified — confirmation sent to %s", row.Email))

	default:
		avrJSON(w, http.StatusBadRequest, false, "action must be \"verify\" or \"reject\"")
	}
}

// avrVerifyRSVP PATCHes the rsvps row(s) for the given email, setting
// verified=true and optionally the corrected name/attending_guests.
// Returns the updated rows.
func avrVerifyRSVP(supabaseURL, apiKey string, req AdminVerifyRequest) ([]avrRSVPRow, error) {
	patch := map[string]interface{}{"verified": true}
	if name := strings.TrimSpace(req.Name); name != "" {
		patch["name"] = name
	}
	if req.AttendingGuests != nil {
		cleaned := make([]string, 0, len(req.AttendingGuests))
		for _, g := range req.AttendingGuests {
			if g = strings.TrimSpace(g); g != "" {
				cleaned = append(cleaned, g)
			}
		}
		patch["attending_guests"] = cleaned
	}

	jsonData, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s/rest/v1/rsvps?email=eq.%s", supabaseURL, url.QueryEscape(req.Email))
	httpReq, err := http.NewRequest("PATCH", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("apikey", apiKey)
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	// return=representation so we know whether any row matched and can
	// send the confirmation email without a second fetch
	httpReq.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase PATCH returned %d", resp.StatusCode)
	}

	var rows []avrRSVPRow
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func avrDeleteRSVP(supabaseURL, apiKey, email string) error {
	apiURL := fmt.Sprintf("%s/rest/v1/rsvps?email=eq.%s", supabaseURL, url.QueryEscape(email))
	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("supabase DELETE returned %d", resp.StatusCode)
	}
	return nil
}
