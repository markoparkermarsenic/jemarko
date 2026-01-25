package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// EmailService handles sending emails via Resend
type EmailService struct {
	apiKey    string
	fromEmail string
	fromName  string
}

// ResendEmailRequest represents the Resend API request payload
type ResendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text,omitempty"`
	HTML    string   `json:"html,omitempty"`
}

// ResendEmailResponse represents the Resend API response
type ResendEmailResponse struct {
	ID    string `json:"id"`
	Error struct {
		Message string `json:"message"`
		Name    string `json:"name"`
	} `json:"error,omitempty"`
}

// NewEmailService creates a new email service instance using Resend
func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("FROM_EMAIL")
	fromName := os.Getenv("FROM_NAME")

	if fromName == "" {
		fromName = "Jemima & Marko Wedding"
	}

	if fromEmail == "" {
		fromEmail = "wedding@jemarko.com"
	}

	return &EmailService{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// SendEmail sends an email using Resend API
func (es *EmailService) SendEmail(to, subject, body string) error {
	// If no API key is configured, log to console instead
	if es.apiKey == "" {
		log.Println("⚠️  RESEND_API_KEY not configured - logging email to console instead")
		return es.logEmailToConsole(to, subject, body)
	}

	// Construct the from field
	from := fmt.Sprintf("%s <%s>", es.fromName, es.fromEmail)

	// Create HTML version of the email
	htmlBody := strings.ReplaceAll(body, "\n", "<br>")

	// Prepare request payload
	payload := ResendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Text:    body,
		HTML:    fmt.Sprintf("<div style='font-family: sans-serif;'>%s</div>", htmlBody),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %v", err)
	}

	// Create HTTP request to Resend API
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+es.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email via Resend: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	var resendResp ResendEmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&resendResp); err != nil {
		return fmt.Errorf("failed to decode Resend response: %v", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		if resendResp.Error.Message != "" {
			return fmt.Errorf("Resend API error: %s", resendResp.Error.Message)
		}
		return fmt.Errorf("Resend API returned status %d", resp.StatusCode)
	}

	log.Printf("✓ Email sent successfully to %s (Resend ID: %s)", to, resendResp.ID)
	return nil
}

// logEmailToConsole logs the email to console (for development/testing)
func (es *EmailService) logEmailToConsole(to, subject, body string) error {
	log.Println("=" + strings.Repeat("=", 70))
	log.Println("📧 EMAIL (Console Mode - Resend not configured)")
	log.Println("=" + strings.Repeat("=", 70))
	log.Printf("From: %s <%s>", es.fromName, es.fromEmail)
	log.Printf("To: %s", to)
	log.Printf("Subject: %s", subject)
	log.Println(strings.Repeat("-", 72))
	log.Println(body)
	log.Println("=" + strings.Repeat("=", 70))
	return nil
}

// SendConfirmationEmail sends a confirmation email to the guest
func SendConfirmationEmail(req RSVPRequest) error {
	emailService := NewEmailService()

	var subject, body string

	if req.IsAttending {
		subject = "Wedding RSVP Confirmation - We Can't Wait to See You! 🎉"
		attendingList := strings.Join(req.AttendingGuests, ", ")

		dietInfo := ""
		if req.Diet != "" {
			dietInfo = fmt.Sprintf("\n\nDietary Requirements:\n%s", req.Diet)
		}

		body = fmt.Sprintf(`
Dear %s,

Thank you for confirming your attendance at our wedding!

Attending Guests:
%s%s

We're so excited to celebrate with you!

Wedding Details:
Date: [Your Wedding Date]
Time: [Your Wedding Time]
Location: [Your Wedding Venue]
Address: [Venue Address]

If you need to make any changes to your RSVP, please contact us at [Your Contact Email].

With love,
Jemima & Marko

---
This is an automated confirmation email.
`, req.Name, attendingList, dietInfo)
	} else {
		subject = "Wedding RSVP Confirmation - We'll Miss You"

		body = fmt.Sprintf(`
Dear %s,

Thank you for your RSVP response.

We're sorry you won't be able to join us for our wedding celebration. You'll be missed!

If your plans change or you need to update your RSVP, please don't hesitate to contact us at [Your Contact Email].

With love,
Jemima & Marko

---
This is an automated confirmation email.
`, req.Name)
	}

	return emailService.SendEmail(req.Email, subject, body)
}

// SendUnlistedGuestNotification sends an email to admin when unlisted guest tries to RSVP
func SendUnlistedGuestNotification(req RSVPRequest, ipAddress, userAgent string) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		log.Println("⚠️  ADMIN_EMAIL not configured - skipping unlisted guest notification")
		return
	}

	emailService := NewEmailService()

	timestamp := time.Now().Format(time.RFC1123)

	subject := fmt.Sprintf("⚠️  Unlisted Guest Attempt: %s, email: %s", req.Name, req.Email)

	body := fmt.Sprintf(`
Hello,

Someone not on the guest list attempted to RSVP for your wedding.

Name Entered: %s
Time: %s
IP Address: %s
User Agent: %s

This person was not found in your guest list. You may want to:
1. Check if this is a misspelling of an existing guest
2. Add them to the guest list if they should be invited
3. Contact them directly if needed

---
This is an automated notification from your wedding RSVP system.
`, req.Name, timestamp, ipAddress, userAgent)

	if err := emailService.SendEmail(adminEmail, subject, body); err != nil {
		log.Printf("Failed to send unlisted guest notification: %v", err)
	} else {
		log.Printf("✓ Sent unlisted guest notification to admin for: %s", req.Name)
	}
}
