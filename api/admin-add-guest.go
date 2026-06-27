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
	"os"
	"strconv"
	"strings"
	"time"
)

// AddGuestRequest is the request body for adding a new guest
type AddGuestRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// AddGuestResponse is the response after adding a guest
type AddGuestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// validateToken checks the HMAC-SHA256 admin token (mirrors admin-login.go)
func validateToken(token string, secret string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	payload, sig := parts[0], parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	if !hmac.Equal([]byte(sig), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		return false
	}
	payloadParts := strings.SplitN(payload, ":", 2)
	if len(payloadParts) != 2 {
		return false
	}
	ts, err := strconv.ParseInt(payloadParts[1], 10, 64)
	if err != nil {
		return false
	}
	return time.Since(time.Unix(ts, 0)) <= 8*time.Hour
}

// Handler handles adding a new guest to the invite list
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

	// Auth
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AddGuestResponse{Success: false, Message: "Server configuration error"})
		return
	}
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" || !validateToken(token, adminPassword) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AddGuestResponse{Success: false, Message: "Unauthorized"})
		return
	}

	// Parse body
	var req AddGuestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AddGuestResponse{Success: false, Message: "Invalid request format"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Address = strings.TrimSpace(req.Address)
	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AddGuestResponse{Success: false, Message: "Name is required"})
		return
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_API_KEY")
	if supabaseURL == "" || supabaseKey == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AddGuestResponse{Success: false, Message: "Database not configured"})
		return
	}

	// Duplicate check (case-insensitive)
	checkURL := fmt.Sprintf("%s/rest/v1/guests?name=ilike.%s&select=id&limit=1", supabaseURL, req.Name)
	checkReq, err := http.NewRequest("GET", checkURL, nil)
	if err == nil {
		checkReq.Header.Set("apikey", supabaseKey)
		checkReq.Header.Set("Authorization", "Bearer "+supabaseKey)
		if resp, err := http.DefaultClient.Do(checkReq); err == nil {
			defer resp.Body.Close()
			var existing []map[string]interface{}
			if json.NewDecoder(resp.Body).Decode(&existing) == nil && len(existing) > 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(AddGuestResponse{
					Success: false,
					Message: fmt.Sprintf("A guest named %q already exists on the invite list", req.Name),
				})
				return
			}
		}
	}

	// Insert guest
	record := map[string]string{"name": req.Name, "address": req.Address}
	jsonData, _ := json.Marshal(record)

	insertReq, err := http.NewRequest("POST", fmt.Sprintf("%s/rest/v1/guests", supabaseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating insert request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AddGuestResponse{Success: false, Message: "Failed to add guest"})
		return
	}
	insertReq.Header.Set("apikey", supabaseKey)
	insertReq.Header.Set("Authorization", "Bearer "+supabaseKey)
	insertReq.Header.Set("Content-Type", "application/json")
	insertReq.Header.Set("Prefer", "return=minimal")

	resp, err := http.DefaultClient.Do(insertReq)
	if err != nil || (resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK) {
		if err != nil {
			log.Printf("Error inserting guest: %v", err)
		} else {
			log.Printf("Supabase returned %d inserting guest", resp.StatusCode)
			resp.Body.Close()
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AddGuestResponse{Success: false, Message: "Failed to add guest"})
		return
	}
	resp.Body.Close()

	log.Printf("✓ Admin added guest: %s (address: %q)", req.Name, req.Address)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AddGuestResponse{
		Success: true,
		Message: fmt.Sprintf("%s has been added to the invite list", req.Name),
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
