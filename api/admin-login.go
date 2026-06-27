package handler

import (
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

const adminUsername = "jemarko"
const tokenValidityHours = 8

// AdminLoginRequest represents a login request
type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AdminLoginResponse represents a login response
type AdminLoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}

// generateToken creates an HMAC-SHA256 token containing username:timestamp
func generateToken(username string, secret string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	payload := fmt.Sprintf("%s:%s", username, timestamp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	// Token format: base64-ish readable string: payload.signature
	return fmt.Sprintf("%s.%s", payload, sig)
}

// ValidateToken checks an HMAC token and returns true if valid and not expired
func ValidateToken(token string, secret string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	payload := parts[0]
	sig := parts[1]

	// Re-compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return false
	}

	// Check expiry — payload is "username:timestamp"
	payloadParts := strings.SplitN(payload, ":", 2)
	if len(payloadParts) != 2 {
		return false
	}
	ts, err := strconv.ParseInt(payloadParts[1], 10, 64)
	if err != nil {
		return false
	}
	issued := time.Unix(ts, 0)
	if time.Since(issued) > time.Duration(tokenValidityHours)*time.Hour {
		return false
	}
	return true
}

// Handler handles admin login requests
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

	var req AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AdminLoginResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Printf("ADMIN_PASSWORD env var not configured")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AdminLoginResponse{
			Success: false,
			Message: "Server configuration error",
		})
		return
	}

	// Validate credentials — username is hardcoded, password from env
	if strings.ToLower(strings.TrimSpace(req.Username)) != adminUsername ||
		req.Password != adminPassword {
		log.Printf("Failed admin login attempt for username: %s", req.Username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AdminLoginResponse{
			Success: false,
			Message: "Invalid username or password",
		})
		return
	}

	token := generateToken(adminUsername, adminPassword)
	log.Printf("Admin login successful for: %s", adminUsername)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AdminLoginResponse{
		Success: true,
		Token:   token,
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
