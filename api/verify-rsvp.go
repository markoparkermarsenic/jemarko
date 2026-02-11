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

	// Support both GET (from email link) and POST (from API)
	var email, apiKeyParam string

	if r.Method == http.MethodGet {
		// Parse from query parameters (email link)
		email = r.URL.Query().Get("email")
		apiKeyParam = r.URL.Query().Get("apiKey")
	} else if r.Method == http.MethodPost {
		// Parse from JSON body (API)
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
		email = req.Email
		apiKeyParam = req.APIKey
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify API key
	adminAPIKey := os.Getenv("ADMIN_API_KEY")
	if adminAPIKey == "" {
		log.Printf("ADMIN_API_KEY not configured")
		showErrorPage(w, r, "Server configuration error")
		return
	}

	if apiKeyParam != adminAPIKey {
		log.Printf("Invalid API key provided for verification attempt")
		showErrorPage(w, r, "Invalid API key")
		return
	}

	// Validate email
	if email == "" {
		showErrorPage(w, r, "Email is required")
		return
	}

	// Update RSVP in database to set verified = true
	db := shared.NewDatabase()
	if !db.IsConfigured() {
		log.Printf("Database not configured")
		showErrorPage(w, r, "Database not configured")
		return
	}

	// Update the RSVP using Supabase REST API
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseAPIKey := os.Getenv("SUPABASE_API_KEY")

	updateData := map[string]interface{}{
		"verified": true,
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		log.Printf("Error marshaling update data: %v", err)
		showErrorPage(w, r, "Server error")
		return
	}

	// Update using email filter
	apiURL := fmt.Sprintf("%s/rest/v1/rsvps?email=eq.%s", supabaseURL, email)
	updateReq, err := http.NewRequest("PATCH", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating update request: %v", err)
		showErrorPage(w, r, "Server error")
		return
	}

	updateReq.Header.Set("apikey", supabaseAPIKey)
	updateReq.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Prefer", "return=minimal")

	resp, err := http.DefaultClient.Do(updateReq)
	if err != nil {
		log.Printf("Error executing update request: %v", err)
		showErrorPage(w, r, "Failed to update RSVP")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		log.Printf("Supabase returned status %d when updating RSVP", resp.StatusCode)
		showErrorPage(w, r, "Failed to update RSVP")
		return
	}

	log.Printf("✓ RSVP verified for email: %s", email)

	// Send confirmation email to the now-verified guest
	// Fetch the RSVP to get full details
	fetchURL := fmt.Sprintf("%s/rest/v1/rsvps?email=eq.%s&select=*", supabaseURL, email)
	fetchReq, err := http.NewRequest("GET", fetchURL, nil)
	if err == nil {
		fetchReq.Header.Set("apikey", supabaseAPIKey)
		fetchReq.Header.Set("Authorization", "Bearer "+supabaseAPIKey)

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
					log.Printf("✓ Sent confirmation email to verified guest: %s", email)
				}
			}
		}
	}

	// Return appropriate response based on request method
	if r.Method == http.MethodGet {
		showSuccessPage(w, r, email)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "RSVP verified successfully",
		})
	}
}

// showSuccessPage shows an HTML success page
func showSuccessPage(w http.ResponseWriter, r *http.Request, email string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>RSVP Verified</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
        }
        .container {
            background: white;
            border-radius: 10px;
            padding: 40px;
            max-width: 500px;
            text-align: center;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
        }
        .success-icon {
            font-size: 64px;
            color: #28a745;
            margin-bottom: 20px;
        }
        h1 {
            color: #333;
            margin: 0 0 10px 0;
        }
        p {
            color: #666;
            line-height: 1.6;
            margin: 20px 0;
        }
        .email {
            background: #f8f9fa;
            padding: 10px;
            border-radius: 5px;
            font-family: monospace;
            color: #495057;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">✓</div>
        <h1>RSVP Verified!</h1>
        <p>The RSVP for <strong class="email">%s</strong> has been successfully verified.</p>
        <p>A confirmation email has been sent to the guest.</p>
        <p>You can close this window now.</p>
    </div>
</body>
</html>
`, email)
	w.Write([]byte(html))
}

// showErrorPage shows an HTML error page
func showErrorPage(w http.ResponseWriter, r *http.Request, message string) {
	// If this is an API request (POST), return JSON
	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": message,
		})
		return
	}

	// Otherwise return HTML page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Verification Error</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
        }
        .container {
            background: white;
            border-radius: 10px;
            padding: 40px;
            max-width: 500px;
            text-align: center;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
        }
        .error-icon {
            font-size: 64px;
            color: #dc3545;
            margin-bottom: 20px;
        }
        h1 {
            color: #333;
            margin: 0 0 10px 0;
        }
        p {
            color: #666;
            line-height: 1.6;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="error-icon">✗</div>
        <h1>Verification Failed</h1>
        <p>%s</p>
        <p>Please contact the administrator if you believe this is an error.</p>
    </div>
</body>
</html>
`, message)
	w.Write([]byte(html))
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
