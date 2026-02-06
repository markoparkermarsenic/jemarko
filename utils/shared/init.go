package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// InitializeDatabase checks if the guests table exists, creates it if needed, and imports CSV data
func (db *Database) InitializeDatabase() error {
	if !db.IsConfigured() {
		return fmt.Errorf("database not configured")
	}

	log.Println("🔧 Initializing database...")
	log.Printf("📡 Connecting to Supabase at: %s", db.url)

	// Test connection first
	if err := db.testConnection(); err != nil {
		return fmt.Errorf("database connection failed: %v", err)
	}

	// Check if guests table exists
	exists, err := db.tableExists("guests")
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %v", err)
	}

	if !exists {
		return fmt.Errorf("guests table doesn't exist - please run migrations: supabase db reset")
	}

	log.Println("📥 Loading guest list from CSV...")

	// Try multiple possible paths for the CSV file
	csvPaths := []string{
		"api/invite_list.csv", // From project root (vercel dev)
		"invite_list.csv",     // From api directory (direct execution)
	}

	var guests []Guest
	loaded := false

	for _, path := range csvPaths {
		guests, err = LoadGuestsFromCSV(path)
		if err == nil {
			loaded = true
			log.Printf("✓ Loaded %d guests from %s", len(guests), path)
			break
		}
	}

	if !loaded {
		return fmt.Errorf("could not load invite_list.csv: %v", err)
	}

	if err := db.ImportGuestsToDatabase(guests); err != nil {
		return fmt.Errorf("failed to import guests: %v", err)
	}

	log.Println("✅ Database initialized successfully!")
	return nil
}

// testConnection tests if we can connect to the database
func (db *Database) testConnection() error {
	url := fmt.Sprintf("%s/rest/v1/", db.url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", db.apiKey)
	req.Header.Set("Authorization", "Bearer "+db.apiKey)

	resp, err := db.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error (status %d)", resp.StatusCode)
	}

	log.Printf("✓ Connected to Supabase successfully (status %d)", resp.StatusCode)
	return nil
}

// tableExists checks if a table exists in the database
func (db *Database) tableExists(tableName string) (bool, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?limit=0", db.url, tableName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("apikey", db.apiKey)
	req.Header.Set("Authorization", "Bearer "+db.apiKey)

	resp, err := db.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// 404 means table doesn't exist
	// 200 means table exists
	return resp.StatusCode == http.StatusOK, nil
}

// isTableEmpty checks if a table has any rows
func (db *Database) isTableEmpty(tableName string) (bool, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?select=id&limit=1", db.url, tableName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("apikey", db.apiKey)
	req.Header.Set("Authorization", "Bearer "+db.apiKey)

	resp, err := db.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var records []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return false, err
	}

	return len(records) == 0, nil
}
