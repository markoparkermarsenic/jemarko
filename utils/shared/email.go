package shared

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/resend/resend-go/v3"
)

// EmailService handles sending emails via Resend
type EmailService struct {
	client    *resend.Client
	fromEmail string
	fromName  string
	isEnabled bool
}

// NewEmailService creates a new email service instance using Resend SDK
func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("FROM_EMAIL")
	fromName := os.Getenv("FROM_NAME")

	if fromName == "" {
		fromName = "Mim & Marko"
	}

	if fromEmail == "" {
		fromEmail = "mim&marko@mimko.love"
	}

	var client *resend.Client
	isEnabled := apiKey != ""

	if isEnabled {
		client = resend.NewClient(apiKey)
	}

	return &EmailService{
		client:    client,
		fromEmail: fromEmail,
		fromName:  fromName,
		isEnabled: isEnabled,
	}
}

// SendEmail sends an email using Resend SDK
func (es *EmailService) SendEmail(to, subject, body string) error {
	// If no API key is configured, log to console instead
	if !es.isEnabled {
		log.Println("⚠️  RESEND_API_KEY not configured - logging email to console instead")
		return es.logEmailToConsole(to, subject, body)
	}

	// Construct the from field
	from := fmt.Sprintf("%s <%s>", es.fromName, es.fromEmail)

	// Create HTML version of the email
	htmlBody := strings.ReplaceAll(body, "\n", "<br>")

	// Prepare email params using Resend SDK
	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Text:    body,
		Html:    fmt.Sprintf("<div style='font-family: sans-serif;'>%s</div>", htmlBody),
	}

	// Send email using Resend SDK
	sent, err := es.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email via Resend: %v", err)
	}

	log.Printf("✓ Email sent successfully to %s (Resend ID: %s)", to, sent.Id)
	return nil
}

// SendTemplateEmail sends an email using a Resend template
func (es *EmailService) SendTemplateEmail(to, templateName string, templateData map[string]interface{}) error {
	// If no API key is configured, log to console instead
	if !es.isEnabled {
		log.Println("⚠️  RESEND_API_KEY not configured - logging email to console instead")
		return es.logTemplateEmailToConsole(to, templateName, templateData)
	}

	// Construct the from field
	from := fmt.Sprintf("%s <%s>", es.fromName, es.fromEmail)

	// Prepare email params using Resend SDK with template
	params := &resend.SendEmailRequest{
		From: from,
		To:   []string{to},
		Template: &resend.EmailTemplate{
			Id:        templateName,
			Variables: templateData,
		},
	}

	// Send email using Resend SDK
	sent, err := es.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send template email via Resend: %v", err)
	}

	log.Printf("✓ Template email sent successfully to %s (Template: %s, Resend ID: %s)", to, templateName, sent.Id)
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

// logTemplateEmailToConsole logs the template email to console (for development/testing)
func (es *EmailService) logTemplateEmailToConsole(to, templateName string, templateData map[string]interface{}) error {
	log.Println("=" + strings.Repeat("=", 70))
	log.Println("📧 TEMPLATE EMAIL (Console Mode - Resend not configured)")
	log.Println("=" + strings.Repeat("=", 70))
	log.Printf("From: %s <%s>", es.fromName, es.fromEmail)
	log.Printf("To: %s", to)
	log.Printf("Template: %s", templateName)
	log.Println(strings.Repeat("-", 72))
	log.Println("Template Data:")
	for key, value := range templateData {
		log.Printf("  %s: %v", key, value)
	}
	log.Println("=" + strings.Repeat("=", 70))
	return nil
}

// SendConfirmationEmail sends a confirmation email to the guest using Resend template
func SendConfirmationEmail(req RSVPRequest) error {
	emailService := NewEmailService()

	// Send email using the rsvp-confirm template
	return emailService.SendTemplateEmail(req.Email, "rsvp-confirm", map[string]interface{}{})
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

// SendUnverifiedRSVPNotification sends an email to admin when unverified guest completes RSVP with verification button
func SendUnverifiedRSVPNotification(req RSVPRequest) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		log.Println("⚠️  ADMIN_EMAIL not configured - skipping unverified RSVP notification")
		return
	}

	adminAPIKey := os.Getenv("ADMIN_API_KEY")
	if adminAPIKey == "" {
		log.Println("⚠️  ADMIN_API_KEY not configured - skipping unverified RSVP notification")
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://your-domain.vercel.app" // fallback
	}
	// Ensure BASE_URL has https:// prefix
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	emailService := NewEmailService()
	timestamp := time.Now().Format(time.RFC1123)

	// Create HTML email with verification button
	subject := fmt.Sprintf("🔔 New Unverified RSVP: %s", req.Name)

	attendingStatus := "NOT ATTENDING"
	guestList := ""
	if req.IsAttending {
		attendingStatus = "ATTENDING"
		for i, guest := range req.AttendingGuests {
			guestList += fmt.Sprintf("%d. %s\n", i+1, guest)
		}
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .details { background-color: #fff; padding: 20px; border: 1px solid #dee2e6; border-radius: 5px; margin-bottom: 20px; }
        .detail-row { margin-bottom: 10px; }
        .label { font-weight: bold; color: #495057; }
        .button { display: inline-block; padding: 12px 24px; background-color: #28a745; color: white; text-decoration: none; border-radius: 5px; font-weight: bold; margin: 20px 0; }
        .button:hover { background-color: #218838; }
        .footer { color: #6c757d; font-size: 0.9em; margin-top: 30px; padding-top: 20px; border-top: 1px solid #dee2e6; }
        .warning { background-color: #fff3cd; padding: 15px; border-left: 4px solid #ffc107; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>🔔 New Unverified RSVP Received</h2>
        </div>
        
        <div class="warning">
            <strong>Action Required:</strong> This guest was not found in your verified guest list. Please review and verify if appropriate.
        </div>
        
        <div class="details">
            <div class="detail-row">
                <span class="label">Name:</span> %s
            </div>
            <div class="detail-row">
                <span class="label">Email:</span> %s
            </div>
            <div class="detail-row">
                <span class="label">Status:</span> %s
            </div>
            <div class="detail-row">
                <span class="label">Submitted:</span> %s
            </div>
            %s
            %s
        </div>
        
        <div style="text-align: center;">
            <a href="%s/api/verify-rsvp?email=%s&apiKey=%s" class="button" style="color: white;">
                ✓ Verify This RSVP
            </a>
        </div>
        
        <div class="footer">
            <p><strong>What happens when you verify:</strong></p>
            <ul>
                <li>The guest will be marked as verified in the database</li>
                <li>They will receive a confirmation email</li>
                <li>Their avatar will appear in the plaza</li>
            </ul>
            <p>This is an automated notification from your wedding RSVP system.</p>
        </div>
    </div>
</body>
</html>
`, req.Name, req.Email, attendingStatus, timestamp,
		func() string {
			if req.IsAttending {
				return fmt.Sprintf(`<div class="detail-row">
                <span class="label">Guests:</span><br/>
                <pre style="margin: 5px 0; padding: 10px; background-color: #f8f9fa; border-radius: 3px;">%s</pre>
            </div>`, guestList)
			}
			return ""
		}(),
		func() string {
			if req.Diet != "" {
				return fmt.Sprintf(`<div class="detail-row">
                <span class="label">Dietary Requirements:</span> %s
            </div>`, req.Diet)
			}
			return ""
		}(),
		baseURL, req.Email, adminAPIKey)

	// Plain text version
	textBody := fmt.Sprintf(`
New Unverified RSVP Received

ACTION REQUIRED: This guest was not found in your verified guest list.

Name: %s
Email: %s
Status: %s
Submitted: %s
%s%s

To verify this RSVP, click the link below or copy it into your browser:
%s/api/verify-rsvp?email=%s&apiKey=%s

What happens when you verify:
- The guest will be marked as verified in the database
- They will receive a confirmation email
- Their avatar will appear in the plaza

---
This is an automated notification from your wedding RSVP system.
`, req.Name, req.Email, attendingStatus, timestamp,
		func() string {
			if req.IsAttending {
				return fmt.Sprintf("\nGuests:\n%s\n", guestList)
			}
			return ""
		}(),
		func() string {
			if req.Diet != "" {
				return fmt.Sprintf("Dietary Requirements: %s\n", req.Diet)
			}
			return ""
		}(),
		baseURL, req.Email, adminAPIKey)

	// Send using the raw email method with HTML
	from := fmt.Sprintf("%s <%s>", emailService.fromName, emailService.fromEmail)

	if !emailService.isEnabled {
		log.Println("⚠️  RESEND_API_KEY not configured - logging email to console instead")
		log.Println("=" + strings.Repeat("=", 70))
		log.Println("📧 UNVERIFIED RSVP NOTIFICATION (Console Mode)")
		log.Println("=" + strings.Repeat("=", 70))
		log.Printf("From: %s", from)
		log.Printf("To: %s", adminEmail)
		log.Printf("Subject: %s", subject)
		log.Println(strings.Repeat("-", 72))
		log.Println(textBody)
		log.Println("=" + strings.Repeat("=", 70))
		return
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{adminEmail},
		Subject: subject,
		Text:    textBody,
		Html:    htmlBody,
	}

	sent, err := emailService.client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send unverified RSVP notification: %v", err)
	} else {
		log.Printf("✓ Sent unverified RSVP notification to admin for: %s (Resend ID: %s)", req.Name, sent.Id)
	}
}
