package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Database handles Supabase operations
type Database struct {
	url    string
	apiKey string
	client *http.Client
}

// GuestRecord represents a guest in the database
type GuestRecord struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at,omitempty"`
	Dietary   string `json:"dietary,omitempty"`
}

// RSVPRecord represents an RSVP submission in the database
type RSVPRecord struct {
	ID              string   `json:"id,omitempty"`
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	IsAttending     bool     `json:"is_attending"`
	AttendingGuests []string `json:"attending_guests,omitempty"`
	Diet            string   `json:"diet,omitempty"`
	SubmittedAt     string   `json:"submitted_at,omitempty"`
}

var (
	guestListCache []Guest
	guestListMutex sync.RWMutex
	guestListOnce  sync.Once
)

// NewDatabase creates a new database connection
func NewDatabase() *Database {
	url := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_API_KEY")

	return &Database{
		url:    url,
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// IsConfigured checks if the database is properly configured
func (db *Database) IsConfigured() bool {
	return db.url != "" && db.apiKey != ""
}

// LoadGuests loads the guest list from Supabase with caching
func (db *Database) LoadGuests() ([]Guest, error) {
	// Try to use cache first
	guestListMutex.RLock()
	if len(guestListCache) > 0 {
		guests := guestListCache
		guestListMutex.RUnlock()
		return guests, nil
	}
	guestListMutex.RUnlock()

	// Load from database
	if !db.IsConfigured() {
		log.Println("⚠️  Supabase not configured - using in-memory guest list")
		guests := getDefaultGuestList()

		// Cache the result
		guestListMutex.Lock()
		guestListCache = guests
		guestListMutex.Unlock()

		return guests, nil
	}

	url := fmt.Sprintf("%s/rest/v1/guests?select=*", db.url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", db.apiKey)
	req.Header.Set("Authorization", "Bearer "+db.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := db.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch guests: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Supabase returned status %d", resp.StatusCode)
	}

	var records []GuestRecord
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Convert to Guest format
	guests := make([]Guest, len(records))
	for i, record := range records {
		guests[i] = Guest{
			ID:   record.ID,
			Name: record.Name,
		}
	}

	// Cache the result
	guestListMutex.Lock()
	guestListCache = guests
	guestListMutex.Unlock()

	log.Printf("✓ Loaded %d guests from Supabase", len(guests))
	return guests, nil
}

// SaveRSVP saves an RSVP submission to Supabase
func (db *Database) SaveRSVP(rsvp RSVPRequest) error {
	if !db.IsConfigured() {
		log.Printf("⚠️  Supabase not configured - RSVP logged to console only")
		log.Printf("RSVP: %s (%s) - Attending: %v - Guests: %v - Diet: %s",
			rsvp.Name, rsvp.Email, rsvp.IsAttending, rsvp.AttendingGuests, rsvp.Diet)
		return nil
	}

	record := RSVPRecord{
		Name:            rsvp.Name,
		Email:           rsvp.Email,
		IsAttending:     rsvp.IsAttending,
		AttendingGuests: rsvp.AttendingGuests,
		Diet:            rsvp.Diet,
		SubmittedAt:     time.Now().UTC().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal RSVP: %v", err)
	}

	url := fmt.Sprintf("%s/rest/v1/rsvps", db.url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", db.apiKey)
	req.Header.Set("Authorization", "Bearer "+db.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := db.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to save RSVP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Supabase returned status %d", resp.StatusCode)
	}

	log.Printf("✓ RSVP saved to Supabase for %s", rsvp.Email)
	return nil
}

// getDefaultGuestList returns a default guest list for development
func getDefaultGuestList() []Guest {
	return []Guest{
		{ID: "1", Name: "John Smith"},
		{ID: "2", Name: "Jane Smith"},
		{ID: "3", Name: "Bob Johnson"},
		{ID: "4", Name: "Alice Williams"},
		{ID: "5", Name: "Tom Williams"},
	}
}
