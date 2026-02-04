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
		fromName = "Jemima & Marko Wedding"
	}

	if fromEmail == "" {
		fromEmail = "wedding@jemarko.com"
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
