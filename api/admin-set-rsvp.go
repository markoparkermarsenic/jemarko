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
)

const adminOverrideEmailSuffix = "@admin.jemarko.internal"

type SetRSVPRequest struct {
	GuestName string `json:"guestName"`
	Status    string `json:"status"` // "attending", "not_attending", "no_response"
}

type SetRSVPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func validateToken(token, secret string) bool {
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

func overrideEmail(guestName string) string {
	safe := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(guestName), " ", "."))
	return safe + adminOverrideEmailSuffix
}

func Handler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)
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
		jsonErr(w, http.StatusInternalServerError, "Server configuration error")
		return
	}
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" || !validateToken(token, adminPassword) {
		jsonErr(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req SetRSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	req.GuestName = strings.TrimSpace(req.GuestName)
	if req.GuestName == "" {
		jsonErr(w, http.StatusBadRequest, "guestName is required")
		return
	}
	if req.Status != "attending" && req.Status != "not_attending" && req.Status != "no_response" {
		jsonErr(w, http.StatusBadRequest, "invalid status value")
		return
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_API_KEY")
	if supabaseURL == "" || supabaseKey == "" {
		jsonErr(w, http.StatusInternalServerError, "Database not configured")
		return
	}

	email := overrideEmail(req.GuestName)

	if err := deleteOverride(supabaseURL, supabaseKey, email); err != nil {
		log.Printf("Error deleting override for %s: %v", req.GuestName, err)
		jsonErr(w, http.StatusInternalServerError, "Failed to update RSVP status")
		return
	}

	if req.Status == "no_response" {
		log.Printf("Admin reset RSVP to no_response for: %s", req.GuestName)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SetRSVPResponse{true,
			fmt.Sprintf("%s has been reset to no response", req.GuestName)})
		return
	}

	if err := insertOverride(supabaseURL, supabaseKey, req.GuestName, email, req.Status == "attending"); err != nil {
		log.Printf("Error inserting override for %s: %v", req.GuestName, err)
		jsonErr(w, http.StatusInternalServerError, "Failed to update RSVP status")
		return
	}

	statusMsg := "not attending"
	if req.Status == "attending" {
		statusMsg = "attending"
	}
	log.Printf("Admin set RSVP for %s → %s", req.GuestName, req.Status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SetRSVPResponse{true,
		fmt.Sprintf("%s has been marked as %s", req.GuestName, statusMsg)})
}

func insertOverride(supabaseURL, apiKey, guestName, email string, isAttending bool) error {
	attendingGuests := []string{}
	if isAttending {
		attendingGuests = []string{guestName}
	}
	record := map[string]interface{}{
		"name": guestName, "email": email,
		"is_attending": isAttending, "attending_guests": attendingGuests,
		"diet": "", "verified": true,
		"submitted_at": time.Now().UTC().Format(time.RFC3339),
	}
	jsonData, err := json.Marshal(record)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/rest/v1/rsvps", supabaseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("supabase returned %d", resp.StatusCode)
	}
	return nil
}

func deleteOverride(supabaseURL, apiKey, email string) error {
	req, err := http.NewRequest("DELETE",
		fmt.Sprintf("%s/rest/v1/rsvps?email=eq.%s", supabaseURL, url.QueryEscape(email)), nil)
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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("supabase DELETE returned %d", resp.StatusCode)
	}
	return nil
}

func jsonErr(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(SetRSVPResponse{false, message})
}

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
